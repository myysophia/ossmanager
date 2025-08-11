package function

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	ossService "github.com/myysophia/ossmanager/internal/oss"
	"go.uber.org/zap"
)

// OSSEvent 阿里云OSS事件结构
type OSSEvent struct {
	Events []struct {
		EventName    string `json:"eventName"`
		EventSource  string `json:"eventSource"`
		EventTime    string `json:"eventTime"`
		EventVersion string `json:"eventVersion"`
		OSS          struct {
			Bucket struct {
				Arn  string `json:"arn"`
				Name string `json:"name"`
			} `json:"bucket"`
			Object struct {
				Key       string `json:"key"`
				Size      int64  `json:"size"`
				ETag      string `json:"eTag"`
				Type      string `json:"type"`
				URL       string `json:"url"`
				FileACL   string `json:"fileACL"`
				ObjectACL string `json:"objectACL"`
			} `json:"object"`
		} `json:"oss"`
	} `json:"events"`
}

// MD5CalculationRequest 是手动触发MD5计算的请求结构
type MD5CalculationRequest struct {
	BucketName string `json:"bucket_name"`
	ObjectKey  string `json:"object_key"`
	FileID     uint   `json:"file_id"`
}

// MD5Calculator MD5计算器
type MD5Calculator struct {
	storageFactory *ossService.DefaultStorageFactory
	calculateChan  chan *models.OSSFile
	workers        int
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
}

// NewMD5Calculator 创建MD5计算器
func NewMD5Calculator(storageFactory *ossService.DefaultStorageFactory, workers int) *MD5Calculator {
	if workers <= 0 {
		workers = 3 // 默认3个工作协程
	}
	ctx, cancel := context.WithCancel(context.Background())
	calculator := &MD5Calculator{
		storageFactory: storageFactory,
		calculateChan:  make(chan *models.OSSFile, 100),
		workers:        workers,
		ctx:            ctx,
		cancel:         cancel,
	}
	calculator.Start()
	return calculator
}

// Start 启动MD5计算器
func (c *MD5Calculator) Start() {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(i)
	}
	logger.Info("MD5计算器已启动", zap.Int("workers", c.workers))
}

// Stop 停止MD5计算器
func (c *MD5Calculator) Stop() {
	c.cancel()
	close(c.calculateChan)
	c.wg.Wait()
	logger.Info("MD5计算器已停止")
}

// TriggerCalculation 触发MD5计算
func (c *MD5Calculator) TriggerCalculation(file *models.OSSFile) error {
	// 检查文件是否已有MD5
	if file.MD5 != "" {
		return nil
	}

	// 更新文件状态为计算中
	fileUpdate := &models.OSSFile{
		MD5Status: models.MD5StatusCalculating,
	}
	if err := db.GetDB().Model(&models.OSSFile{}).Where("id = ?", file.ID).Updates(fileUpdate).Error; err != nil {
		logger.Error("更新文件MD5状态失败", zap.Uint("id", file.ID), zap.Error(err))
		return err
	}

	// 将文件放入计算队列
	select {
	case c.calculateChan <- file:
		return nil
	case <-time.After(5 * time.Second):
		// 超时处理
		fileUpdate := &models.OSSFile{
			MD5Status: models.MD5StatusFailed,
		}
		db.GetDB().Model(&models.OSSFile{}).Where("id = ?", file.ID).Updates(fileUpdate)
		return errors.New("MD5计算队列已满，请稍后重试")
	}
}

// worker MD5计算工作协程
func (c *MD5Calculator) worker(id int) {
	defer c.wg.Done()
	logger.Info("MD5计算工作协程启动", zap.Int("worker_id", id))

	for {
		select {
		case <-c.ctx.Done():
			logger.Info("MD5计算工作协程停止", zap.Int("worker_id", id))
			return
		case file, ok := <-c.calculateChan:
			if !ok {
				logger.Info("MD5计算工作协程停止", zap.Int("worker_id", id))
				return
			}

			// 计算文件MD5
			c.calculateFileMD5(file)
		}
	}
}

// calculateFileMD5 计算文件MD5
func (c *MD5Calculator) calculateFileMD5(file *models.OSSFile) {
	logger.Info("开始计算文件MD5", zap.Uint("id", file.ID), zap.String("object_key", file.ObjectKey))

	// 更新状态为计算中
	updateStatus := func(status string, md5 string) {
		fileUpdate := &models.OSSFile{
			MD5Status: status,
			MD5:       md5,
		}
		if err := db.GetDB().Model(&models.OSSFile{}).Where("id = ?", file.ID).Updates(fileUpdate).Error; err != nil {
			logger.Error("更新文件MD5状态失败", zap.Uint("id", file.ID), zap.Error(err))
		}
	}

	// 获取配置
	storage, err := c.storageFactory.GetStorageService(file.StorageType)
	if err != nil {
		logger.Error("获取存储提供商失败", zap.String("storage_type", file.StorageType), zap.Error(err))
		updateStatus(models.MD5StatusFailed, "")
		return
	}

	// 下载文件并计算MD5
	reader, err := storage.GetObject(file.ObjectKey)
	if err != nil {
		logger.Error("下载文件失败", zap.String("object_key", file.ObjectKey), zap.Error(err))
		updateStatus(models.MD5StatusFailed, "")
		return
	}
	defer reader.Close()

	// 计算MD5
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		logger.Error("计算MD5失败", zap.String("object_key", file.ObjectKey), zap.Error(err))
		updateStatus(models.MD5StatusFailed, "")
		return
	}

	// 转换为十六进制字符串
	md5Str := hex.EncodeToString(hash.Sum(nil))
	logger.Info("文件MD5计算完成",
		zap.Uint("id", file.ID),
		zap.String("object_key", file.ObjectKey),
		zap.String("md5", md5Str))

	// 更新MD5
	updateStatus(models.MD5StatusCompleted, md5Str)
}

// CalculateMD5Sync 同步计算OSS文件的MD5
func (c *MD5Calculator) CalculateMD5Sync(file *models.OSSFile) error {
	logger.Info("开始同步计算文件MD5", zap.Uint("file_id", file.ID))

	// 获取存储提供商
	storage, err := c.storageFactory.GetStorageService(file.StorageType)
	if err != nil {
		return fmt.Errorf("获取存储提供商失败: %w", err)
	}

	// 下载文件并计算MD5
	reader, err := storage.GetObject(file.ObjectKey)
	if err != nil {
		return fmt.Errorf("获取文件内容失败: %w", err)
	}
	defer reader.Close()

	// 计算MD5
	hash := md5.New()
	if _, err := io.Copy(hash, reader); err != nil {
		return fmt.Errorf("计算MD5时发生错误: %w", err)
	}

	md5Value := hex.EncodeToString(hash.Sum(nil))
	logger.Info("文件MD5计算完成",
		zap.Uint("file_id", file.ID),
		zap.String("md5", md5Value))

	// 更新数据库中的MD5值
	return c.UpdateFileMD5(file.ID, md5Value)
}

// UpdateFileMD5 更新数据库中文件的MD5值
func (c *MD5Calculator) UpdateFileMD5(fileID uint, md5Value string) error {
	result := db.GetDB().Model(&models.OSSFile{}).Where("id = ?", fileID).Update("md5", md5Value)
	if result.Error != nil {
		return fmt.Errorf("更新文件MD5值失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到ID为%d的文件", fileID)
	}
	return nil
}

// CalculateOSSFileMD5 是阿里云函数计算的入口函数，用于计算OSS文件的MD5值
func CalculateOSSFileMD5(ctx context.Context, event OSSEvent) (string, error) {
	// 打印事件信息
	logger.Info("接收到OSS事件", zap.Any("event", event))

	// 遍历事件
	for _, e := range event.Events {
		bucketName := e.OSS.Bucket.Name
		objectKey := e.OSS.Object.Key

		// 计算文件MD5
		md5Value, err := calculateMD5FromOSS(ctx, bucketName, objectKey)
		if err != nil {
			logger.Error("计算文件MD5失败",
				zap.String("bucket", bucketName),
				zap.String("objectKey", objectKey),
				zap.Error(err))
			return "", err
		}

		// 更新数据库中的MD5值
		err = updateFileMD5InDB(ctx, bucketName, objectKey, md5Value)
		if err != nil {
			logger.Error("更新数据库MD5值失败",
				zap.String("bucket", bucketName),
				zap.String("objectKey", objectKey),
				zap.String("md5", md5Value),
				zap.Error(err))
			return "", err
		}

		logger.Info("成功计算并更新文件MD5",
			zap.String("bucket", bucketName),
			zap.String("objectKey", objectKey),
			zap.String("md5", md5Value))
	}

	return "MD5计算完成", nil
}

// calculateMD5FromOSS 从OSS读取文件并计算MD5值
func calculateMD5FromOSS(ctx context.Context, bucketName, objectKey string) (string, error) {
	// 获取OSS配置
	cfg := config.GetConfig().OSS.AliyunOSS

	// 创建OSS客户端
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("创建阿里云OSS客户端失败: %w", err)
	}

	// 获取存储空间
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("获取存储空间失败: %w", err)
	}

	// 获取文件流
	body, err := bucket.GetObject(objectKey)
	if err != nil {
		return "", fmt.Errorf("获取文件流失败: %w", err)
	}
	defer body.Close()

	// 创建MD5哈希器
	h := md5.New()

	// 使用io.Copy进行流式计算MD5，避免一次性将整个文件加载到内存
	if _, err := io.Copy(h, body); err != nil {
		return "", fmt.Errorf("计算MD5失败: %w", err)
	}

	// 获取MD5值的十六进制表示
	md5Value := hex.EncodeToString(h.Sum(nil))
	return md5Value, nil
}

// updateFileMD5InDB 更新数据库中文件的MD5值
func updateFileMD5InDB(ctx context.Context, bucketName, objectKey, md5Value string) error {
	// 获取数据库连接
	database := db.GetDB()
	if database == nil {
		return fmt.Errorf("数据库连接未初始化")
	}

	// 更新文件记录的MD5值
	result := database.Model(&models.OSSFile{}).
		Where("bucket = ? AND object_key = ?", bucketName, objectKey).
		Update("md5", md5Value)

	if result.Error != nil {
		return fmt.Errorf("更新文件MD5值失败: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("未找到文件记录，bucket=%s, objectKey=%s", bucketName, objectKey)
	}

	return nil
}

// HandleManualRequest 处理手动触发的MD5计算请求
func (c *MD5Calculator) HandleManualRequest(ctx context.Context, req MD5CalculationRequest) error {
	// 从数据库获取完整的文件信息
	var file models.OSSFile
	if err := db.GetDB().First(&file, req.FileID).Error; err != nil {
		return fmt.Errorf("获取文件信息失败: %w", err)
	}

	return c.CalculateMD5Sync(&file)
}

// 注册函数计算处理函数
// 注意：此功能已被Serverless计算平台替代，不再使用RegisterHandler
// func init() {
// 	fc.RegisterHandler(CalculateOSSFileMD5)
// }
