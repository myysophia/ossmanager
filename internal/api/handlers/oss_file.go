package handlers

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/myysophia/ossmanager-backend/internal/auth"
	"github.com/myysophia/ossmanager-backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/config"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/oss"
	"github.com/myysophia/ossmanager-backend/internal/upload"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ProgressReader 用于追踪读取进度的Reader
type ProgressReader struct {
	reader   io.Reader
	total    int64
	read     int64
	callback func(read, total int64)
}

// NewProgressReader 创建一个新的进度Reader
func NewProgressReader(reader io.Reader, total int64, callback func(read, total int64)) *ProgressReader {
	return &ProgressReader{
		reader:   reader,
		total:    total,
		read:     0,
		callback: callback,
	}
}

// Read 实现io.Reader接口
func (pr *ProgressReader) Read(p []byte) (n int, err error) {
	n, err = pr.reader.Read(p)
	pr.read += int64(n)
	if pr.callback != nil {
		pr.callback(pr.read, pr.total)
	}
	return n, err
}

type OSSFileHandler struct {
	*BaseHandler
	storageFactory oss.StorageFactory
	DB             *gorm.DB
}

func NewOSSFileHandler(storageFactory oss.StorageFactory, db *gorm.DB) *OSSFileHandler {
	return &OSSFileHandler{
		BaseHandler:    NewBaseHandler(),
		storageFactory: storageFactory,
		DB:             db,
	}
}

// Upload 上传文件 - 智能选择上传方式
func (h *OSSFileHandler) Upload(c *gin.Context) {
	// 检查Content-Type以确定使用哪种上传方式
	contentType := c.GetHeader("Content-Type")

	// 获取文件大小阈值配置（默认100MB）
	chunkThreshold := int64(100 * 1024 * 1024) // 100MB
	if thresholdStr := c.GetHeader("X-Chunk-Threshold"); thresholdStr != "" {
		if threshold, err := strconv.ParseInt(thresholdStr, 10, 64); err == nil {
			chunkThreshold = threshold
		}
	}

	// 如果是multipart/form-data，使用表单上传方式
	if strings.Contains(contentType, "multipart/form-data") {
		h.uploadFormFileWithChunking(c, chunkThreshold)
		return
	}

	// 否则使用流式上传方式
	h.uploadStreamWithChunking(c, chunkThreshold)
}

// uploadFormFileWithChunking 表单文件上传（智能选择分片）
func (h *OSSFileHandler) uploadFormFileWithChunking(c *gin.Context, chunkThreshold int64) {
	// 获取用户ID
	userID := c.GetUint("userID")

	// 获取用户指定的 bucket 信息
	regionCode := c.GetHeader("region_code")
	bucketName := c.GetHeader("bucket_name")

	if regionCode == "" || bucketName == "" {
		h.Error(c, utils.CodeInvalidParams, "请指定 region_code 和 bucket_name")
		return
	}

	// 获取存储配置
	var config models.OSSConfig
	if err := h.DB.Where("is_default = ?", true).First(&config).Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取默认存储配置失败")
		return
	}

	// 检查用户是否有权限访问该桶
	if !auth.CheckBucketAccess(h.DB, userID, regionCode, bucketName) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		h.Error(c, utils.CodeInvalidParams, "获取文件失败")
		return
	}

	// 获取存储服务
	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	// 检查是否强制覆盖
	forceOverwrite := c.GetHeader("X-Force-Overwrite") == "true"

	// 声明 objectKey 变量
	var objectKey string

	// 获取自定义路径
	customPath := c.GetHeader("X-Custom-Path")
	if customPath != "" {
		// 清理和验证自定义路径
		customPath = strings.Trim(customPath, "/")
		// 验证路径中不包含危险字符
		if strings.Contains(customPath, "..") || strings.ContainsAny(customPath, "\\<>:\"|?*") {
			h.Error(c, utils.CodeInvalidParams, "自定义路径包含非法字符")
			return
		}
		// 使用用户自定义路径
		if customPath == "" {
			// 自定义路径为空，直接上传到根目录
			objectKey = file.Filename
		} else {
			objectKey = customPath + "/" + file.Filename
		}
	} else {
		// 没有提供自定义路径，使用固定路径生成方式
		username, _ := c.Get("username")
		objectKey = utils.GenerateFixedObjectKey(username.(string), file.Filename)
	}

	// 如果不是强制覆盖，检查文件是否已存在（基于完整路径）
	if !forceOverwrite {
		var existingFile models.OSSFile
		err := h.DB.Where("object_key = ? AND bucket = ? AND status = ?",
			objectKey, bucketName, "ACTIVE").First(&existingFile).Error

		if err == nil {
			// 文件已存在，返回错误提示用户确认
			h.Error(c, utils.CodeFileExists, "在相同路径下文件已存在，请确认是否要覆盖")
			return
		} else if err != gorm.ErrRecordNotFound {
			// 数据库查询错误
			h.Error(c, utils.CodeServerError, "检查文件是否存在失败")
			return
		}
	}

	// 如果客户端提供了上传任务ID，则使用该ID；否则生成新的
	taskID := c.GetHeader("Upload-Task-ID")
	if taskID == "" {
		taskID = c.Query("task_id")
	}
	if taskID == "" {
		taskID = uuid.NewString()
	}

	src, err := file.Open()
	if err != nil {
		h.Error(c, utils.CodeServerError, "打开文件失败")
		return
	}
	defer src.Close()

	// 根据文件大小选择上传方式
	if file.Size <= chunkThreshold {
		// 简单上传
		logger.Info("使用简单上传", zap.Int64("file_size", file.Size), zap.Int64("threshold", chunkThreshold))
		upload.DefaultManager.Start(taskID, file.Size)

		uploadURL, err := storage.UploadToBucketWithProgress(src, objectKey, regionCode, bucketName, func(consumed, total int64) {
			if total == 0 {
				total = file.Size
			}
			upload.DefaultManager.Update(taskID, consumed)
		})
		if err != nil {
			h.Error(c, utils.CodeServerError, "上传文件失败")
			upload.DefaultManager.Finish(taskID)
			return
		}
		upload.DefaultManager.Finish(taskID)

		// 保存文件记录并返回
		h.saveFileRecord(c, config, objectKey, file.Filename, file.Size, bucketName, uploadURL)
	} else {
		// 分片上传
		logger.Info("使用分片上传", zap.Int64("file_size", file.Size), zap.Int64("threshold", chunkThreshold))
		uploadURL, err := h.uploadFileWithChunks(c, storage, src, objectKey, regionCode, bucketName, file.Size, taskID, file.Filename)
		if err != nil {
			h.Error(c, utils.CodeServerError, err.Error())
			upload.DefaultManager.Finish(taskID)
			return
		}

		// 保存文件记录并返回
		h.saveFileRecord(c, config, objectKey, file.Filename, file.Size, bucketName, uploadURL)
	}
}

// uploadStreamWithChunking 流式文件上传（智能选择分片）
func (h *OSSFileHandler) uploadStreamWithChunking(c *gin.Context, chunkThreshold int64) {
	// 获取用户ID
	userID := c.GetUint("userID")

	// 获取用户指定的 bucket 信息
	regionCode := c.GetHeader("region_code")
	bucketName := c.GetHeader("bucket_name")

	// 获取文件元数据（从请求头中获取）
	originalFilename := c.GetHeader("X-File-Name")
	contentLengthStr := c.GetHeader("Content-Length")

	if regionCode == "" || bucketName == "" {
		h.Error(c, utils.CodeInvalidParams, "请指定 region_code 和 bucket_name")
		return
	}

	if originalFilename == "" {
		h.Error(c, utils.CodeInvalidParams, "请提供文件名（X-File-Name header）")
		return
	}

	contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
	if err != nil || contentLength <= 0 {
		h.Error(c, utils.CodeInvalidParams, "请提供有效的文件大小（Content-Length header）")
		return
	}

	// 获取存储配置
	var config models.OSSConfig
	if err := h.DB.Where("is_default = ?", true).First(&config).Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取默认存储配置失败")
		return
	}

	// 检查用户是否有权限访问该桶
	if !auth.CheckBucketAccess(h.DB, userID, regionCode, bucketName) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	// 获取存储服务
	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	// 检查是否强制覆盖
	forceOverwrite := c.GetHeader("X-Force-Overwrite") == "true"

	// 声明 objectKey 变量
	var objectKey string

	// 获取自定义路径
	customPath := c.GetHeader("X-Custom-Path")
	if customPath != "" {
		// 清理和验证自定义路径
		customPath = strings.Trim(customPath, "/")
		// 验证路径中不包含危险字符
		if strings.Contains(customPath, "..") || strings.ContainsAny(customPath, "\\<>:\"|?*") {
			h.Error(c, utils.CodeInvalidParams, "自定义路径包含非法字符")
			return
		}
		// 使用用户自定义路径
		if customPath == "" {
			// 自定义路径为空，直接上传到根目录
			objectKey = originalFilename
		} else {
			objectKey = customPath + "/" + originalFilename
		}
	} else {
		// 没有提供自定义路径，使用固定路径生成方式
		username, _ := c.Get("username")
		objectKey = utils.GenerateFixedObjectKey(username.(string), originalFilename)
	}

	// 如果不是强制覆盖，检查文件是否已存在（基于完整路径）
	if !forceOverwrite {
		var existingFile models.OSSFile
		err := h.DB.Where("object_key = ? AND bucket = ? AND status = ?",
			objectKey, bucketName, "ACTIVE").First(&existingFile).Error

		if err == nil {
			// 文件已存在，返回错误提示用户确认
			h.Error(c, utils.CodeFileExists, "在相同路径下文件已存在，请确认是否要覆盖")
			return
		} else if err != gorm.ErrRecordNotFound {
			// 数据库查询错误
			h.Error(c, utils.CodeServerError, "检查文件是否存在失败")
			return
		}
	}

	// 获取任务ID
	taskID := c.GetHeader("Upload-Task-ID")
	if taskID == "" {
		taskID = c.Query("task_id")
	}
	if taskID == "" {
		taskID = uuid.NewString()
	}

	// 根据文件大小选择上传方式
	if contentLength <= chunkThreshold {
		// 简单上传
		logger.Info("使用简单上传", zap.Int64("content_length", contentLength), zap.Int64("threshold", chunkThreshold))
		upload.DefaultManager.Start(taskID, contentLength)

		uploadURL, err := storage.UploadToBucketWithProgress(c.Request.Body, objectKey, regionCode, bucketName, func(consumed, total int64) {
			if total == 0 {
				total = contentLength
			}
			upload.DefaultManager.Update(taskID, consumed)
		})
		if err != nil {
			h.Error(c, utils.CodeServerError, "上传文件失败")
			upload.DefaultManager.Finish(taskID)
			return
		}
		upload.DefaultManager.Finish(taskID)

		// 保存文件记录并返回
		h.saveFileRecord(c, config, objectKey, originalFilename, contentLength, bucketName, uploadURL)
	} else {
		// 分片上传
		logger.Info("使用分片上传", zap.Int64("content_length", contentLength), zap.Int64("threshold", chunkThreshold))
		uploadURL, err := h.uploadFileWithChunks(c, storage, c.Request.Body, objectKey, regionCode, bucketName, contentLength, taskID, originalFilename)
		if err != nil {
			h.Error(c, utils.CodeServerError, err.Error())
			upload.DefaultManager.Finish(taskID)
			return
		}

		// 保存文件记录并返回
		h.saveFileRecord(c, config, objectKey, originalFilename, contentLength, bucketName, uploadURL)
	}
}

// uploadFileWithChunks 分片上传文件
func (h *OSSFileHandler) uploadFileWithChunks(c *gin.Context, storage oss.StorageService, reader io.Reader, objectKey, regionCode, bucketName string, totalSize int64, taskID, originalFilename string) (string, error) {
	// 默认分片大小：10MB
	chunkSize := int64(10 * 1024 * 1024)
	if chunkSizeStr := c.GetHeader("X-Chunk-Size"); chunkSizeStr != "" {
		if size, err := strconv.ParseInt(chunkSizeStr, 10, 64); err == nil && size > 0 {
			chunkSize = size
		}
	}

	// 并发量，默认为配置值或1
	concurrency := 1
	cfg := config.GetConfig()
	if cfg != nil && cfg.App.ChunkConcurrency > 0 {
		concurrency = cfg.App.ChunkConcurrency
	}
	if concStr := c.GetHeader("X-Chunk-Concurrency"); concStr != "" {
		if cc, err := strconv.Atoi(concStr); err == nil && cc > 0 {
			concurrency = cc
		}
	}

	// 计算总分片数
	totalChunks := int((totalSize + chunkSize - 1) / chunkSize)

	resumeUploadID := c.GetHeader("X-Upload-Id")
	if resumeUploadID == "" {
		resumeUploadID = c.Query("upload_id")
	}
	if resumeUploadID != "" {
		objectKey = c.GetHeader("X-Object-Key")
		if objectKey == "" {
			objectKey = c.Query("object_key")
		}
	}

	var uploadID string
	logger.Debug("Initializing multipart upload", zap.String("objectKey", objectKey), zap.String("regionCode", regionCode), zap.String("bucketName", bucketName))
	var err error
	if resumeUploadID == "" {
		uploadID, _, err = storage.InitMultipartUploadToBucket(objectKey, regionCode, bucketName)
		if err != nil {
			return "", fmt.Errorf("初始化分片上传失败: %v", err)
		}
	} else {
		uploadID = resumeUploadID
	}

	logger.Info("开始分片上传",
		zap.String("task_id", taskID),
		zap.String("object_key", objectKey),
		zap.Int64("total_size", totalSize),
		zap.Int64("chunk_size", chunkSize),
		zap.Int("total_chunks", totalChunks),
	)

	// 开始分片上传进度追踪
	upload.DefaultManager.StartWithChunks(taskID, totalSize, true, totalChunks)

	var parts []oss.Part
	var uploadedBytes int64
	partNumber := 1
	// 包装请求体以在读取过程中实时更新上传进度
	progressReader := upload.NewReader(taskID, reader)
	// 创建带缓冲的reader，并设置合理的缓冲区大小
	//bufferedReader := bufio.NewReaderSize(reader, int(chunkSize))
	bufferedReader := bufio.NewReaderSize(progressReader, int(chunkSize))

	if resumeUploadID != "" {
		existing, err := storage.ListUploadedPartsToBucket(objectKey, uploadID, regionCode, bucketName)
		if err == nil && len(existing) > 0 {
			logger.Info("继续未完成的分片上传", zap.Int("existing_parts", len(existing)))
			for _, p := range existing {
				if p.PartNumber != partNumber {
					break
				}
				parts = append(parts, p)
				size := chunkSize
				if p.PartNumber == totalChunks {
					size = totalSize - int64(totalChunks-1)*chunkSize
				}
				if _, err := io.CopyN(io.Discard, bufferedReader, size); err != nil {
					return "", fmt.Errorf("跳过已上传分片失败: %v", err)
				}
				uploadedBytes += size
				partNumber++
			}
		}
	}

	// 读取分片超时时间，可通过头部 X-Chunk-Read-Timeout 调整，默认 5 分钟
	readTimeout := 5 * time.Minute
	if timeoutStr := c.GetHeader("X-Chunk-Read-Timeout"); timeoutStr != "" {
		if t, err := strconv.Atoi(timeoutStr); err == nil && t > 0 {
			readTimeout = time.Duration(t) * time.Second
		}
	}

	maxRetries := 10

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var mu sync.Mutex
	errCh := make(chan error, 1)

	for uploadedBytes < totalSize && partNumber <= totalChunks {
		if len(errCh) > 0 {
			break
		}
		// 计算当前分片大小
		currentChunkSize := chunkSize
		if uploadedBytes+chunkSize > totalSize {
			currentChunkSize = totalSize - uploadedBytes
		}

		logger.Debug("准备上传分片",
			zap.Int("part_number", partNumber),
			zap.Int64("chunk_size", currentChunkSize),
			zap.Int64("uploaded_bytes", uploadedBytes),
		)

		// 读取分片数据，带重试机制
		var chunkData []byte
		var readErr error

		readStart := time.Now()
		for retry := 0; retry < maxRetries; retry++ {
			if retry > 0 {
				logger.Warn("重试读取分片数据",
					zap.Int("part_number", partNumber),
					zap.Int("retry", retry),
				)
				time.Sleep(time.Duration(retry) * time.Second) // 递增延迟
			}

			chunkData = make([]byte, currentChunkSize)

			// 使用通道和协程实现超时控制
			done := make(chan error, 1)
			go func() {
				_, err := io.ReadFull(bufferedReader, chunkData)
				done <- err
			}()

			select {
			case readErr = <-done:
				// 读取完成
				break
			case <-time.After(readTimeout):
				readErr = fmt.Errorf("读取分片数据超时")
				logger.Warn("读取分片数据超时",
					zap.Int("part_number", partNumber),
					zap.Duration("timeout", readTimeout),
				)
				continue // 重试
			}

			if readErr == nil || readErr == io.EOF || readErr == io.ErrUnexpectedEOF {
				break // 成功或预期的EOF
			}
		}

		if readErr != nil && readErr != io.EOF && readErr != io.ErrUnexpectedEOF {
			upload.DefaultManager.Fail(taskID, "读取分片数据失败")
			return "", fmt.Errorf("读取分片数据失败: %v", readErr)
		}

		if len(chunkData) == 0 {
			break
		}

		logger.Debug("读取分片完成",
			zap.Int("part_number", partNumber),
			zap.Duration("elapsed", time.Since(readStart)),
		)
		// 上传分片
		curPart := partNumber
		dataCopy := make([]byte, len(chunkData))
		copy(dataCopy, chunkData)
		uploadedBytes += int64(len(chunkData))
		partNumber++

		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			urlStart := time.Now()

			uploadURL, err := storage.GeneratePartUploadURL(objectKey, uploadID, curPart, regionCode, bucketName)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("获取分片 %d 上传URL失败: %v", curPart, err):
				default:
				}
				return
			}
			logger.Debug("生成上传URL完成",
				zap.Int("part_number", curPart),
				zap.Duration("elapsed", time.Since(urlStart)),
			)
			uploadStart := time.Now()
			etag, err := h.uploadChunk(storage, uploadURL, dataCopy, curPart)
			if err != nil {
				select {
				case errCh <- fmt.Errorf("上传分片 %d 失败: %v", curPart, err):
				default:
				}
				return
			}
			logger.Debug("上传分片完成",
				zap.Int("part_number", curPart),
				zap.Duration("elapsed", time.Since(uploadStart)),
			)
			mu.Lock()
			parts = append(parts, oss.Part{PartNumber: curPart, ETag: etag})
			mu.Unlock()
			upload.DefaultManager.UpdateChunk(taskID, curPart, true)
			logger.Debug("分片上传成功",
				zap.Int("part_number", curPart),
				zap.String("etag", etag),
			)
		}()
	}

	wg.Wait()
	if len(errCh) > 0 {
		h.safeAbortMultipartUpload(storage, uploadID, objectKey, regionCode, bucketName)
		upload.DefaultManager.Fail(taskID, (<-errCh).Error())
		return "", fmt.Errorf("%v", <-errCh)
	}

	sort.Slice(parts, func(i, j int) bool { return parts[i].PartNumber < parts[j].PartNumber })

	logger.Info("所有分片上传完成，开始合并",
		zap.String("upload_id", uploadID),
		zap.Int("total_parts", len(parts)),
	)

	// 完成分片上传
	uploadURL, err := storage.CompleteMultipartUploadToBucket(objectKey, uploadID, parts, regionCode, bucketName)
	if err != nil {
		// 完成失败，中止分片上传（使用正确的方法）
		h.safeAbortMultipartUpload(storage, uploadID, objectKey, regionCode, bucketName)
		upload.DefaultManager.Fail(taskID, "完成分片上传失败")
		return "", fmt.Errorf("完成分片上传失败: %v", err)
	}

	// 完成进度追踪
	upload.DefaultManager.Finish(taskID)

	logger.Info("分片上传完全成功",
		zap.String("task_id", taskID),
		zap.String("upload_url", uploadURL),
	)

	return uploadURL, nil
}

// safeAbortMultipartUpload 安全地中止分片上传，不会因为错误而阻塞主流程
func (h *OSSFileHandler) safeAbortMultipartUpload(storage oss.StorageService, uploadID, objectKey, regionCode, bucketName string) {
	// 使用类型断言检查是否为阿里云存储服务
	if aliyunStorage, ok := storage.(*oss.AliyunOSSService); ok {
		// 使用新的AbortMultipartUploadToBucket方法
		err := aliyunStorage.AbortMultipartUploadToBucket(uploadID, objectKey, regionCode, bucketName)
		if err != nil {
			logger.Warn("中止分片上传失败，但继续处理",
				zap.String("upload_id", uploadID),
				zap.String("object_key", objectKey),
				zap.Error(err),
			)
		}
	} else {
		// 其他存储服务使用原来的方法
		err := storage.AbortMultipartUpload(uploadID, objectKey)
		if err != nil {
			logger.Warn("中止分片上传失败，但继续处理",
				zap.String("upload_id", uploadID),
				zap.String("object_key", objectKey),
				zap.Error(err),
			)
		}
	}
}

// uploadChunkGeneric 通用分片上传方法（当预签名URL不可用时）
func (h *OSSFileHandler) uploadChunkGeneric(storage oss.StorageService, data []byte, partNumber int, uploadID, objectKey string) (string, error) {
	// 这里需要实现针对具体存储服务的分片上传逻辑
	// 由于每个云服务商的API不同，这里只是一个占位符
	// 实际实现需要调用storage service的具体分片上传方法
	return "", fmt.Errorf("通用分片上传方法未实现")
}

// uploadChunk 上传单个分片
func (h *OSSFileHandler) uploadChunk(storage oss.StorageService, uploadURL string, data []byte, partNumber int) (string, error) {
	// 这里需要根据具体的存储服务实现分片上传
	// 由于不同的云服务商有不同的分片上传API，这里提供一个通用的HTTP PUT方法

	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			logger.Warn("重试上传分片",
				zap.Int("part_number", partNumber),
				zap.Int("retry", attempt),
			)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(data))
		if err != nil {
			return "", err
		}

		req.Header.Set("Content-Type", "application/octet-stream")
		req.Header.Set("Content-Length", strconv.Itoa(len(data)))

		client := &http.Client{Timeout: 30 * time.Second}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("上传分片失败，状态码: %d, 响应: %s", resp.StatusCode, string(body))
			continue
		}

		// 获取ETag
		etag := resp.Header.Get("ETag")
		if etag == "" {
			lastErr = fmt.Errorf("无法获取分片ETag")
			continue
		}
		// 移除ETag中的引号
		etag = strings.Trim(etag, "\"")

		return etag, nil
	}

	return "", lastErr
}

// saveFileRecord 保存文件记录
func (h *OSSFileHandler) saveFileRecord(c *gin.Context, config models.OSSConfig, objectKey, originalFilename string, fileSize int64, bucketName, uploadURL string) {
	// 从配置中获取过期时间，如果未配置则默认为24小时
	expireTime := config.URLExpireTime
	if expireTime <= 0 {
		expireTime = 24 * 3600 // 默认24小时
	}
	expiresAt := time.Now().Add(time.Duration(expireTime) * time.Second)

	// 开始数据库事务，确保原子性
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 首先将相同object_key的旧记录标记为REPLACED
	if err := tx.Model(&models.OSSFile{}).Where(
		"object_key = ? AND bucket = ? AND status = ?",
		objectKey, bucketName, "ACTIVE",
	).Update("status", "REPLACED").Error; err != nil {
		logger.Warn("标记旧文件记录失败", 
			zap.String("object_key", objectKey),
			zap.Error(err),
		)
		tx.Rollback()
		h.Error(c, utils.CodeServerError, "更新旧文件记录失败")
		return
	}

	// 2. 创建新的文件记录
	ossFile := models.OSSFile{
		ConfigID:         config.ID,
		Filename:         objectKey,
		OriginalFilename: originalFilename,
		FileSize:         fileSize,
		StorageType:      config.StorageType,
		Bucket:           bucketName,
		ObjectKey:        objectKey,
		DownloadURL:      uploadURL,
		UploaderID:       utils.GetUserID(c),
		UploadIP:         c.ClientIP(),
		ExpiresAt:        expiresAt,
		Status:           "ACTIVE",
	}

	if err := tx.Create(&ossFile).Error; err != nil {
		tx.Rollback()
		h.Error(c, utils.CodeServerError, "保存文件记录失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		h.Error(c, utils.CodeServerError, "提交事务失败")
		return
	}

	logger.Info("文件记录保存成功",
		zap.String("object_key", objectKey),
		zap.Uint("file_id", ossFile.ID),
		zap.String("status", "ACTIVE"),
	)

	h.Success(c, ossFile)
}

// saveFileRecordForMultipart 分片上传专用文件记录保存函数
func (h *OSSFileHandler) saveFileRecordForMultipart(c *gin.Context, config models.OSSConfig, objectKey, originalFilename string, fileSize int64, bucketName, uploadURL string) {
	// 从配置中获取过期时间，如果未配置则默认为24小时
	expireTime := config.URLExpireTime
	if expireTime <= 0 {
		expireTime = 24 * 3600 // 默认24小时
	}
	expiresAt := time.Now().Add(time.Duration(expireTime) * time.Second)

	// 开始数据库事务，确保原子性
	tx := h.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. 首先将相同object_key的旧记录标记为REPLACED
	if err := tx.Model(&models.OSSFile{}).Where(
		"object_key = ? AND bucket = ? AND status = ?",
		objectKey, bucketName, "ACTIVE",
	).Update("status", "REPLACED").Error; err != nil {
		logger.Warn("标记旧文件记录失败",
			zap.String("object_key", objectKey),
			zap.Error(err),
		)
		tx.Rollback()
		h.Error(c, utils.CodeServerError, "更新旧文件记录失败")
		return
	}

	// 2. 创建新的文件记录
	ossFile := models.OSSFile{
		ConfigID:         config.ID,
		Filename:         objectKey,
		OriginalFilename: originalFilename,
		FileSize:         fileSize,
		StorageType:      config.StorageType,
		Bucket:           bucketName,
		ObjectKey:        objectKey,
		DownloadURL:      uploadURL,
		UploaderID:       utils.GetUserID(c),
		UploadIP:         c.ClientIP(),
		ExpiresAt:        expiresAt,
		Status:           "ACTIVE",
	}

	if err := tx.Create(&ossFile).Error; err != nil {
		tx.Rollback()
		h.Error(c, utils.CodeServerError, "保存文件记录失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		h.Error(c, utils.CodeServerError, "提交事务失败")
		return
	}

	logger.Info("分片上传文件记录保存成功",
		zap.String("task_id", c.GetHeader("X-Task-ID")),
		zap.Uint("file_id", ossFile.ID),
		zap.String("object_key", objectKey),
		zap.String("status", "ACTIVE"),
		zap.String("download_url", uploadURL),
	)

	h.Success(c, ossFile)
}

// InitMultipartUpload 初始化分片上传
func (h *OSSFileHandler) InitMultipartUpload(c *gin.Context) {
	var req struct {
		RegionCode string `json:"region_code" binding:"required"`
		BucketName string `json:"bucket_name" binding:"required"`
		FileName   string `json:"file_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Error(c, utils.CodeInvalidParams, "参数错误")
		return
	}

	// 获取存储配置
	var config models.OSSConfig
	if err := h.DB.Where("is_default = ?", true).First(&config).Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取默认存储配置失败")
		return
	}

	// 检查用户是否有权限访问该桶
	if !auth.CheckBucketAccess(h.DB, c.GetUint("userID"), req.RegionCode, req.BucketName) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	ext := filepath.Ext(req.FileName)
	username, _ := c.Get("username")
	objectKey := utils.GenerateObjectKey(username.(string), ext)

	uploadID, urls, err := storage.InitMultipartUploadToBucket(objectKey, req.RegionCode, req.BucketName)
	if err != nil {
		h.Error(c, utils.CodeServerError, "初始化分片上传失败")
		return
	}

	h.Success(c, gin.H{
		"upload_id":  uploadID,
		"object_key": objectKey,
		"urls":       urls,
	})
}

// CompleteMultipartUpload 完成分片上传
func (h *OSSFileHandler) CompleteMultipartUpload(c *gin.Context) {
	var req struct {
		RegionCode       string   `json:"region_code" binding:"required"`
		BucketName       string   `json:"bucket_name" binding:"required"`
		ObjectKey        string   `json:"object_key" binding:"required"`
		UploadID         string   `json:"upload_id" binding:"required"`
		Parts            []string `json:"parts" binding:"required"`
		OriginalFilename string   `json:"original_filename"`
		FileSize         int64    `json:"file_size"`
		TaskID           string   `json:"task_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Error(c, utils.CodeInvalidParams, "参数错误")
		return
	}

	// 获取存储配置
	var config models.OSSConfig
	if err := h.DB.Where("is_default = ?", true).First(&config).Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取默认存储配置失败")
		return
	}

	// 检查用户是否有权限访问该桶
	if !auth.CheckBucketAccess(h.DB, c.GetUint("userID"), req.RegionCode, req.BucketName) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	// 转换parts为oss.Part类型
	ossParts := make([]oss.Part, len(req.Parts))
	for i, part := range req.Parts {
		ossParts[i] = oss.Part{
			PartNumber: i + 1,
			ETag:       part,
		}
	}

	logger.Info("开始完成分片上传",
		zap.String("upload_id", req.UploadID),
		zap.String("object_key", req.ObjectKey),
		zap.Int("parts_count", len(ossParts)),
		zap.String("task_id", req.TaskID),
	)

	// 完成分片上传
	url, err := storage.CompleteMultipartUploadToBucket(req.ObjectKey, req.UploadID, ossParts, req.RegionCode, req.BucketName)
	if err != nil {
		if req.TaskID != "" {
			upload.DefaultManager.Fail(req.TaskID, "完成分片上传失败")
		}
		h.Error(c, utils.CodeServerError, "完成分片上传失败")
		return
	}

	// 设置默认值
	originalFilename := req.OriginalFilename
	if originalFilename == "" {
		originalFilename = req.ObjectKey
	}

	// 从配置中获取过期时间，如果未配置则默认为24小时
	expireTime := config.URLExpireTime
	if expireTime <= 0 {
		expireTime = 24 * 3600 // 默认24小时
	}

	// 使用改进的文件记录保存逻辑
	h.saveFileRecordForMultipart(c, config, req.ObjectKey, originalFilename, req.FileSize, req.BucketName, url)

	// 完成进度追踪
	if req.TaskID != "" {
		upload.DefaultManager.Finish(req.TaskID)
	}
}

// AbortMultipartUpload 取消分片上传
func (h *OSSFileHandler) AbortMultipartUpload(c *gin.Context) {
	var req struct {
		ConfigID  string `json:"config_id" binding:"required"`
		ObjectKey string `json:"object_key" binding:"required"`
		UploadID  string `json:"upload_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.Error(c, utils.CodeInvalidParams, "参数错误")
		return
	}

	var config models.OSSConfig
	if err := h.DB.First(&config, req.ConfigID).Error; err != nil {
		h.Error(c, utils.CodeConfigNotFound, "存储配置不存在")
		return
	}

	// 检查用户是否有权限访问该桶
	if !auth.CheckBucketAccess(h.DB, c.GetUint("userID"), config.Region, config.Bucket) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	if err := storage.AbortMultipartUpload(req.UploadID, req.ObjectKey); err != nil {
		h.Error(c, utils.CodeServerError, "取消分片上传失败")
		return
	}

	h.Success(c, nil)
}

// ListUploadedParts 获取已上传的分片编号
func (h *OSSFileHandler) ListUploadedParts(c *gin.Context) {
	regionCode := c.Query("region_code")
	bucketName := c.Query("bucket_name")
	objectKey := c.Query("object_key")
	uploadID := c.Query("upload_id")

	if regionCode == "" || bucketName == "" || objectKey == "" || uploadID == "" {
		h.Error(c, utils.CodeInvalidParams, "参数错误")
		return
	}

	// 获取存储配置
	var config models.OSSConfig
	if err := h.DB.Where("is_default = ?", true).First(&config).Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取默认存储配置失败")
		return
	}

	// 权限检查
	if !auth.CheckBucketAccess(h.DB, c.GetUint("userID"), regionCode, bucketName) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	uploadedParts, err := storage.ListUploadedPartsToBucket(objectKey, uploadID, regionCode, bucketName)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取已上传分片失败")
		return
	}

	partNumbers := make([]int, len(uploadedParts))
	for i, p := range uploadedParts {
		partNumbers[i] = p.PartNumber
	}

	h.Success(c, gin.H{"parts": partNumbers})
}

// List 获取文件列表，相同文件名只获取最新一个
func (h *OSSFileHandler) List(c *gin.Context) {
	// 获取用户ID
	userID := c.GetUint("userID")

	// 获取用户可访问的桶列表
	buckets, err := auth.GetUserAccessibleBuckets(h.DB, userID, "")
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取可访问桶列表失败")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	configID := c.Query("config_id")
	// 首先，获取去重后的所有文件名
	var uniqueFileNames []string
	query := h.DB.Model(&models.OSSFile{}).Select("DISTINCT original_filename").Where("bucket IN ?", buckets)
	if configID != "" {
		query = query.Where("config_id = ?", configID)
	}

	if err := query.Pluck("original_filename", &uniqueFileNames).Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取唯一文件名失败")
		return
	}

	total := int64(len(uniqueFileNames))

	// 对于分页的处理
	startIdx := (page - 1) * pageSize
	endIdx := startIdx + pageSize
	if startIdx >= len(uniqueFileNames) {
		h.Success(c, gin.H{
			"total": total,
			"items": []models.OSSFile{},
		})
		return
	}
	if endIdx > len(uniqueFileNames) {
		endIdx = len(uniqueFileNames)
	}

	// 获取当前页的文件名
	pageFileNames := uniqueFileNames[startIdx:endIdx]

	// 对于每个文件名，获取最新的记录
	var files []models.OSSFile
	for _, fileName := range pageFileNames {
		var latest models.OSSFile
		subQuery := h.DB.Model(&models.OSSFile{}).Where("original_filename = ? AND bucket IN ?", fileName, buckets)
		if configID != "" {
			subQuery = subQuery.Where("config_id = ?", configID)
		}

		if err := subQuery.Order("created_at DESC").First(&latest).Error; err != nil {
			// 如果查询出错，跳过这个文件名
			continue
		}

		files = append(files, latest)
	}

	h.Success(c, gin.H{
		"total": total,
		"items": files,
	})
}

// getRegionByBucket 通过存储桶名称获取区域代码
func (h *OSSFileHandler) getRegionByBucket(bucketName string) (string, error) {
	var mapping models.RegionBucketMapping
	err := h.DB.Where("bucket_name = ?", bucketName).First(&mapping).Error
	if err != nil {
		return "", fmt.Errorf("未找到存储桶 %s 对应的区域信息: %w", bucketName, err)
	}
	return mapping.RegionCode, nil
}

// Delete 删除文件
func (h *OSSFileHandler) Delete(c *gin.Context) {
	// 获取用户ID
	userID := c.GetUint("userID")

	// 获取文件信息
	var file models.OSSFile
	if err := h.DB.First(&file, c.Param("id")).Error; err != nil {
		h.Error(c, utils.CodeFileNotFound, "文件不存在")
		return
	}

	// 通过存储桶名称获取区域信息
	regionCode, err := h.getRegionByBucket(file.Bucket)
	if err != nil {
		logger.Error("获取存储桶区域信息失败",
			zap.String("bucket", file.Bucket),
			zap.Error(err))
		h.Error(c, utils.CodeServerError, "获取存储桶区域信息失败")
		return
	}

	// 获取配置信息
	var config models.OSSConfig
	if err := h.DB.First(&config, file.ConfigID).Error; err != nil {
		h.Error(c, utils.CodeConfigNotFound, "存储配置不存在")
		return
	}

	// 检查用户是否有权限访问该桶（使用获取到的区域和存储桶）
	if !auth.CheckBucketAccess(h.DB, userID, regionCode, file.Bucket) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	// 使用获取到的区域和存储桶信息删除文件
	if err := storage.DeleteObjectFromBucket(file.ObjectKey, regionCode, file.Bucket); err != nil {
		logger.Error("删除文件失败",
			zap.String("objectKey", file.ObjectKey),
			zap.String("region", regionCode),
			zap.String("bucket", file.Bucket),
			zap.Error(err))
		h.Error(c, utils.CodeServerError, "删除文件失败")
		return
	}

	if err := h.DB.Delete(&file).Error; err != nil {
		h.Error(c, utils.CodeServerError, "删除文件记录失败")
		return
	}

	logger.Info("文件删除成功",
		zap.Uint("fileID", file.ID),
		zap.String("objectKey", file.ObjectKey),
		zap.String("region", regionCode),
		zap.String("bucket", file.Bucket))

	h.Success(c, nil)
}

// CheckDuplicateFile 检查重复文件
func (h *OSSFileHandler) CheckDuplicateFile(c *gin.Context) {
	// 获取用户ID
	userID := c.GetUint("userID")
	username, _ := c.Get("username")

	// 获取查询参数
	originalFilename := c.Query("filename")
	regionCode := c.Query("region_code")
	bucketName := c.Query("bucket_name")

	if originalFilename == "" {
		h.Error(c, utils.CodeInvalidParams, "文件名不能为空")
		return
	}

	if regionCode == "" || bucketName == "" {
		h.Error(c, utils.CodeInvalidParams, "请指定 region_code 和 bucket_name")
		return
	}

	// 检查用户是否有权限访问该桶
	if !auth.CheckBucketAccess(h.DB, userID, regionCode, bucketName) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	// 获取自定义路径并生成对象键
	customPath := c.Query("custom_path")
	var objectKey string
	if customPath != "" {
		// 清理和验证自定义路径
		customPath = strings.Trim(customPath, "/")
		// 验证路径中不包含危险字符
		if strings.Contains(customPath, "..") || strings.ContainsAny(customPath, "\\<>:\"|?*") {
			h.Error(c, utils.CodeInvalidParams, "自定义路径包含非法字符")
			return
		}
		// 使用用户自定义路径
		if customPath == "" {
			// 自定义路径为空，直接上传到根目录
			objectKey = originalFilename
		} else {
			objectKey = customPath + "/" + originalFilename
		}
	} else {
		// 没有提供自定义路径，使用固定路径生成方式。但为了兼容性，也支持检查绝对路径
		objectKey = utils.GenerateFixedObjectKey(username.(string), originalFilename)
	}

	// 查询数据库中是否存在相同对象键（完整路径）的文件
	var existingFile models.OSSFile
	err := h.DB.Where("object_key = ? AND bucket = ? AND status = ?",
		objectKey, bucketName, "ACTIVE").First(&existingFile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 文件不存在，可以上传
			h.Success(c, gin.H{
				"exists":     false,
				"object_key": objectKey,
				"message":    "文件不存在，可以上传",
			})
			return
		}
		// 数据库查询错误
		h.Error(c, utils.CodeServerError, "查询文件失败")
		return
	}

	// 文件已存在，提取路径信息
	existingPath := ""
	if strings.Contains(existingFile.ObjectKey, "/") {
		// 提取路径部分（去掉文件名）
		parts := strings.Split(existingFile.ObjectKey, "/")
		if len(parts) > 1 {
			existingPath = strings.Join(parts[:len(parts)-1], "/")
		}
	}

	h.Success(c, gin.H{
		"exists":     true,
		"object_key": objectKey,
		"existing_file": gin.H{
			"id":                existingFile.ID,
			"filename":          existingFile.Filename,
			"original_filename": existingFile.OriginalFilename,
			"file_size":         existingFile.FileSize,
			"created_at":        existingFile.CreatedAt,
			"object_key":        existingFile.ObjectKey,
			"path":              existingPath,
		},
		"message": "在相同路径下发现同名文件，是否要覆盖？",
	})
}

// GetDownloadURL 获取文件下载链接
func (h *OSSFileHandler) GetDownloadURL(c *gin.Context) {
	fileID := c.Param("id")

	var file models.OSSFile
	if err := h.DB.First(&file, fileID).Error; err != nil {
		h.Error(c, utils.CodeFileNotFound, "文件不存在")
		return
	}

	// 获取过期时间参数，支持以下选项：
	// 1, 2, 3, 6, 12, 24, 48 小时，0 表示永不过期
	expireHoursStr := c.Query("expire_hours")
	var expireDuration time.Duration
	var neverExpires bool = false

	if expireHoursStr != "" {
		if expireHours, err := strconv.Atoi(expireHoursStr); err == nil {
			switch expireHours {
			case 0:
				// 0 表示永不过期
				neverExpires = true
				expireDuration = 0 // 这个值不会被使用
			case 1, 2, 3, 6, 12, 24, 48:
				// 允许的小时数
				expireDuration = time.Duration(expireHours) * time.Hour
			default:
				// 不在允许范围内，使用默认值1小时
				expireDuration = 1 * time.Hour
			}
		} else {
			// 解析失败时使用默认值
			expireDuration = 1 * time.Hour
		}
	} else {
		// 未指定时使用默认值
		expireDuration = 1 * time.Hour
	}

	// 通过存储桶名称获取区域信息
	regionCode, err := h.getRegionByBucket(file.Bucket)
	if err != nil {
		logger.Error("获取存储桶区域信息失败",
			zap.String("bucket", file.Bucket),
			zap.Error(err))
		h.Error(c, utils.CodeServerError, "获取存储桶区域信息失败")
		return
	}

	// 获取配置信息
	var config models.OSSConfig
	if err := h.DB.First(&config, file.ConfigID).Error; err != nil {
		h.Error(c, utils.CodeConfigNotFound, "存储配置不存在")
		return
	}

	// 检查用户是否有权限访问该桶（使用获取到的区域和存储桶）
	if !auth.CheckBucketAccess(h.DB, c.GetUint("userID"), regionCode, file.Bucket) {
		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
		return
	}

	storage, err := h.storageFactory.GetStorageService(config.StorageType)
	if err != nil {
		h.Error(c, utils.CodeServerError, "获取存储服务失败")
		return
	}

	// 动态生成下载链接
	var downloadURL string
	var expires time.Time

	if neverExpires {
		// 永不过期：返回文件的原始下载链接（如果是公共访问的桶）或者使用一个很长的过期时间
		if aliyunStorage, ok := storage.(*oss.AliyunOSSService); ok {
			// 对于阿里云OSS，使用最大允许的过期时间（7天）作为近似永不过期
			// 实际应用中可能需要定期刷新链接
			downloadURL, expires, err = aliyunStorage.GenerateDownloadURLWithBucket(file.ObjectKey, file.DownloadURL, 7*24*time.Hour)
			if err != nil {
				h.Error(c, utils.CodeServerError, "生成下载链接失败")
				return
			}
			// 设置一个特殊的过期时间表示永不过期
			expires = time.Time{} // 零值表示永不过期
		} else {
			downloadURL, expires, err = storage.GenerateDownloadURL(file.ObjectKey, 7*24*time.Hour)
			if err != nil {
				h.Error(c, utils.CodeServerError, "生成下载链接失败")
				return
			}
			expires = time.Time{} // 零值表示永不过期
		}
	} else {
		// 使用指定的过期时间
		if aliyunStorage, ok := storage.(*oss.AliyunOSSService); ok {
			downloadURL, expires, err = aliyunStorage.GenerateDownloadURLWithBucket(file.ObjectKey, file.DownloadURL, expireDuration)
			if err != nil {
				h.Error(c, utils.CodeServerError, "生成下载链接失败")
				return
			}
		} else {
			downloadURL, expires, err = storage.GenerateDownloadURL(file.ObjectKey, expireDuration)
			if err != nil {
				h.Error(c, utils.CodeServerError, "生成下载链接失败")
				return
			}
		}
	}

	logger.Info("生成文件下载链接",
		zap.String("fileID", fileID),
		zap.String("objectKey", file.ObjectKey),
		zap.Bool("neverExpires", neverExpires),
		zap.Duration("expireDuration", expireDuration),
		zap.Time("expires", expires))

	response := gin.H{
		"download_url": downloadURL,
		"never_expires": neverExpires,
	}
	if !neverExpires {
		response["expires"] = expires
		response["expire_hours"] = int(expireDuration.Hours())
	}
	h.Success(c, response)
}

// GetByOriginalFilename 根据原始文件名获取文件详情
//func (h *OSSFileHandler) GetByOriginalFilename(c *gin.Context) {
//	filename := c.Query("filename")
//	if filename == "" {
//		h.Error(c, utils.CodeInvalidParams, "文件名不能为空")
//		return
//	}
//
//	var ossFile models.OSSFile
//	if err := h.DB.Where("original_filename = ? AND status = ?", filename, "ACTIVE").First(&ossFile).Error; err != nil {
//		h.Error(c, utils.CodeNotFound, "文件不存在")
//		return
//	}
//
//	// 获取配置信息以获取Region
//	var config models.OSSConfig
//	if err := h.DB.First(&config, ossFile.ConfigID).Error; err != nil {
//		h.Error(c, utils.CodeConfigNotFound, "存储配置不存在")
//		return
//	}
//
//	// 检查用户是否有权限访问该桶
//	if !auth.CheckBucketAccess(h.DB, c.GetUint("userID"), config.Region, ossFile.Bucket) {
//		h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
//		return
//	}
//
//	storage, err := h.storageFactory.GetStorageService(ossFile.StorageType)
//	if err != nil {
//		h.Error(c, utils.CodeServerError, "获取存储服务失败")
//		return
//	}
//
//	downloadURL, expires, err := storage.GenerateDownloadURL(ossFile.ObjectKey, 24*time.Hour)
//	if err != nil {
//		h.Error(c, utils.CodeServerError, "生成下载链接失败")
//		return
//	}
//
//	// 更新下载URL和过期时间
//	ossFile.DownloadURL = downloadURL
//	ossFile.ExpiresAt = expires
//	if err := h.DB.Save(&ossFile).Error; err != nil {
//		logger.Error("更新文件下载URL失败", zap.Error(err))
//	}
//
//	h.Success(c, ossFile)
//}
