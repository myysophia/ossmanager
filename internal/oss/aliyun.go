package oss

import (
	"fmt"
	"io"
	"net/url"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
)

// AliyunOSSService 阿里云OSS存储服务
type AliyunOSSService struct {
	client     *oss.Client
	bucket     *oss.Bucket
	config     *config.AliyunOSSConfig
	bucketName string
	uploadDir  string
}

// NewAliyunOSSService 创建阿里云OSS存储服务
func NewAliyunOSSService(cfg *config.AliyunOSSConfig) (*AliyunOSSService, error) {
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("初始化阿里云OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("获取阿里云OSS Bucket失败: %w", err)
	}

	service := &AliyunOSSService{
		client:     client,
		bucket:     bucket,
		config:     cfg,
		bucketName: cfg.Bucket,
		uploadDir:  cfg.UploadDir,
	}

	return service, nil
}

// GetName 获取存储服务名称
func (s *AliyunOSSService) GetName() string {
	return "阿里云OSS"
}

// GetType 获取存储服务类型
func (s *AliyunOSSService) GetType() string {
	return StorageTypeAliyunOSS
}

// getObjectKey 获取对象键
func (s *AliyunOSSService) getObjectKey(filename string) string {
	return path.Join(s.uploadDir, filename)
}

// getEndpoint 根据配置和region获取正确的endpoint
func (s *AliyunOSSService) getEndpoint(regionCode string) string {
	// 如果开启了传输加速，使用加速域名
	if s.config.TransferAccelerate.Enabled {
		switch s.config.TransferAccelerate.Type {
		case "overseas":
			return "https://oss-accelerate-overseas.aliyuncs.com"
		case "global":
			fallthrough
		default:
			return "https://oss-accelerate.aliyuncs.com"
		}
	}

	// 如果有指定region，使用region特定的endpoint
	if regionCode != "" {
		return fmt.Sprintf("https://oss-%s.aliyuncs.com", regionCode)
	}

	// 否则使用配置中的默认endpoint
	return s.config.Endpoint
}

// Upload 上传文件
func (s *AliyunOSSService) Upload(file io.Reader, objectKey string) (string, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 设置Content-Disposition为attachment，强制下载而不是预览
	options := []oss.Option{
		oss.ContentDisposition("attachment"),
	}

	// 上传文件
	err := s.bucket.PutObject(fullObjectKey, file, options...)
	if err != nil {
		logger.Error("上传文件失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", fmt.Errorf("上传文件失败: %w", err)
	}

	// 生成下载链接
	url, err := s.bucket.SignURL(fullObjectKey, oss.HTTPGet, 24*3600)
	if err != nil {
		logger.Error("生成下载链接失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", fmt.Errorf("生成下载链接失败: %w", err)
	}

	return url, nil
}

// InitMultipartUpload 初始化分片上传
func (s *AliyunOSSService) InitMultipartUpload(filename string) (string, []string, error) {
	objectKey := s.getObjectKey(filename)
	// 初始化分片上传，设置Content-Disposition为attachment
	imur, err := s.bucket.InitiateMultipartUpload(objectKey, oss.ContentDisposition("attachment"))
	if err != nil {
		logger.Error("初始化阿里云OSS分片上传失败", zap.String("filename", filename), zap.Error(err))
		return "", nil, fmt.Errorf("初始化阿里云OSS分片上传失败: %w", err)
	}

	// 这里返回uploadID，前端需要保存此ID用于后续的分片上传和完成操作
	// 真实场景中，我们还需要根据文件大小计算分片数量，并为每个分片生成上传URL
	// 这里仅作为示例，实际上应该由前端计算分片并请求签名URL
	return imur.UploadID, nil, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *AliyunOSSService) CompleteMultipartUpload(objectKey string, uploadID string, parts []Part) (string, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 将我们的Part结构转换为阿里云SDK的Part结构
	ossParts := make([]oss.UploadPart, len(parts))
	for i, part := range parts {
		ossParts[i] = oss.UploadPart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}

	// 完成分片上传
	_, err := s.bucket.CompleteMultipartUpload(oss.InitiateMultipartUploadResult{
		Key:      fullObjectKey,
		UploadID: uploadID,
	}, ossParts)

	if err != nil {
		logger.Error("完成阿里云OSS分片上传失败",
			zap.String("objectKey", fullObjectKey),
			zap.String("uploadID", uploadID),
			zap.Error(err))
		return "", fmt.Errorf("完成阿里云OSS分片上传失败: %w", err)
	}

	// 生成下载URL
	signedURL, err := s.bucket.SignURL(fullObjectKey, oss.HTTPGet, int64(s.config.GetOSSURLExpiration().Seconds()))
	if err != nil {
		logger.Error("生成阿里云OSS下载URL失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", fmt.Errorf("生成阿里云OSS下载URL失败: %w", err)
	}

	return signedURL, nil
}

// AbortMultipartUpload 取消分片上传
func (s *AliyunOSSService) AbortMultipartUpload(uploadID string, objectKey string) error {
	fullObjectKey := s.getObjectKey(objectKey)

	// 取消分片上传
	err := s.bucket.AbortMultipartUpload(oss.InitiateMultipartUploadResult{
		Key:      fullObjectKey,
		UploadID: uploadID,
	})

	if err != nil {
		logger.Error("取消阿里云OSS分片上传失败",
			zap.String("objectKey", fullObjectKey),
			zap.String("uploadID", uploadID),
			zap.Error(err))
		return fmt.Errorf("取消阿里云OSS分片上传失败: %w", err)
	}

	return nil
}

// AbortMultipartUploadToBucket 取消指定存储桶的分片上传
func (s *AliyunOSSService) AbortMultipartUploadToBucket(uploadID string, objectKey string, regionCode string, bucketName string) error {
	logger.Info("开始取消分片上传",
		zap.String("uploadID", uploadID),
		zap.String("objectKey", objectKey),
		zap.String("regionCode", regionCode),
		zap.String("bucketName", bucketName))

	// 获取正确的endpoint
	endpoint := s.getEndpoint(regionCode)

	// 创建临时的OSS客户端
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		logger.Error("创建OSS客户端失败（用于取消分片上传）",
			zap.String("endpoint", endpoint),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		// 即使取消失败也不返回错误，避免阻塞主流程
		logger.Warn("取消分片上传失败，但继续处理以避免阻塞")
		return nil
	}

	// 获取指定的存储桶
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		logger.Error("获取存储桶失败（用于取消分片上传）",
			zap.String("bucketName", bucketName),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		// 即使取消失败也不返回错误，避免阻塞主流程
		logger.Warn("取消分片上传失败，但继续处理以避免阻塞")
		return nil
	}

	// 取消分片上传
	err = bucket.AbortMultipartUpload(oss.InitiateMultipartUploadResult{
		Key:      objectKey,
		UploadID: uploadID,
	})

	if err != nil {
		logger.Error("取消阿里云OSS分片上传失败",
			zap.String("objectKey", objectKey),
			zap.String("uploadID", uploadID),
			zap.String("regionCode", regionCode),
			zap.String("bucketName", bucketName),
			zap.Error(err))
		// 即使取消失败也不返回错误，避免阻塞主流程
		logger.Warn("取消分片上传失败，但继续处理以避免阻塞")
		return nil
	}

	logger.Info("分片上传取消成功",
		zap.String("uploadID", uploadID),
		zap.String("objectKey", objectKey))

	return nil
}

// ListUploadedPartsToBucket 获取已上传的分片列表
func (s *AliyunOSSService) ListUploadedPartsToBucket(objectKey string, uploadID string, regionCode string, bucketName string) ([]Part, error) {
	// 获取正确的endpoint（考虑传输加速）
	endpoint := s.getEndpoint(regionCode)

	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		return nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return nil, fmt.Errorf("获取存储桶失败: %w", err)
	}

	var uploadedParts []Part
	marker := 0
	for {
		result, err := bucket.ListUploadedParts(oss.InitiateMultipartUploadResult{
			Key:      objectKey,
			UploadID: uploadID,
		}, oss.PartNumberMarker(marker))
		if err != nil {
			return nil, fmt.Errorf("获取已上传分片失败: %w", err)
		}

		for _, part := range result.UploadedParts {
			uploadedParts = append(uploadedParts, Part{
				PartNumber: part.PartNumber,
				ETag:       strings.Trim(part.ETag, "\""),
			})
		}

		if !result.IsTruncated {
			break
		}
		next, _ := strconv.Atoi(result.NextPartNumberMarker)
		marker = next
	}

	sort.Slice(uploadedParts, func(i, j int) bool { return uploadedParts[i].PartNumber < uploadedParts[j].PartNumber })
	return uploadedParts, nil
}

// GeneratePartUploadURL 生成单个分片上传的预签名URL
func (s *AliyunOSSService) GeneratePartUploadURL(objectKey string, uploadID string, partNumber int, regionCode string, bucketName string) (string, error) {
	endpoint := s.getEndpoint(regionCode)
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		return "", fmt.Errorf("创建OSS客户端失败: %w", err)
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("获取存储桶失败: %w", err)
	}

	options := []oss.Option{
		oss.AddParam("uploadId", uploadID),
		oss.AddParam("partNumber", strconv.Itoa(partNumber)),
		oss.ContentType("application/octet-stream"),
	}

	url, err := bucket.SignURL(objectKey, oss.HTTPPut, 3600, options...)
	if err != nil {
		return "", fmt.Errorf("生成分片上传URL失败: %w", err)
	}

	return url, nil
}

// GenerateDownloadURL 生成下载URL
func (s *AliyunOSSService) GenerateDownloadURL(objectKey string, expiration time.Duration) (string, time.Time, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 设置过期时间
	expires := time.Now().Add(expiration)
	expiresSeconds := int64(expiration.Seconds())

	// 生成签名URL
	signedURL, err := s.bucket.SignURL(fullObjectKey, oss.HTTPGet, expiresSeconds)
	if err != nil {
		logger.Error("生成阿里云OSS下载URL失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", time.Time{}, fmt.Errorf("生成阿里云OSS下载URL失败: %w", err)
	}

	return signedURL, expires, nil
}

// DeleteObject 删除对象
func (s *AliyunOSSService) DeleteObject(objectKey string) error {
	fullObjectKey := s.getObjectKey(objectKey)

	// 删除对象
	err := s.bucket.DeleteObject(fullObjectKey)
	if err != nil {
		logger.Error("删除阿里云OSS对象失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return fmt.Errorf("删除阿里云OSS对象失败: %w", err)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (s *AliyunOSSService) GetObjectInfo(objectKey string) (int64, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 获取对象元数据
	props, err := s.bucket.GetObjectDetailedMeta(fullObjectKey)
	if err != nil {
		logger.Error("获取阿里云OSS对象信息失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return 0, fmt.Errorf("获取阿里云OSS对象信息失败: %w", err)
	}

	// 获取文件大小
	contentLength := props.Get("Content-Length")
	if contentLength == "" {
		return 0, fmt.Errorf("获取阿里云OSS对象大小失败: Content-Length为空")
	}

	var size int64
	_, err = fmt.Sscanf(contentLength, "%d", &size)
	if err != nil {
		return 0, fmt.Errorf("解析阿里云OSS对象大小失败: %w", err)
	}

	return size, nil
}

// GetBucketName 获取存储桶名称
func (s *AliyunOSSService) GetBucketName() string {
	return s.bucketName
}

// GetObject 获取对象内容
func (s *AliyunOSSService) GetObject(objectKey string) (io.ReadCloser, error) {
	fullObjectKey := s.getObjectKey(objectKey)
	body, err := s.bucket.GetObject(fullObjectKey)
	if err != nil {
		logger.Error("获取阿里云OSS对象失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return nil, fmt.Errorf("获取阿里云OSS对象失败: %w", err)
	}
	return body, nil
}

// TriggerMD5Calculation 触发计算MD5值
func (s *AliyunOSSService) TriggerMD5Calculation(objectKey string, fileID uint) error {
	logger.Info("触发阿里云OSS对象MD5计算",
		zap.String("objectKey", objectKey),
		zap.Uint("fileID", fileID),
		zap.String("bucket", s.bucketName))

	// 此方法会在后台使用函数计算服务异步计算MD5，
	// 具体实现将在函数计算服务模块中处理
	if s.config.FunctionCompute.Enabled {
		// 后续会集成函数计算客户端来触发MD5计算
		// 目前的实现只是记录日志，表明功能已被触发
		logger.Info("阿里云OSS函数计算将异步计算MD5值")
		return nil
	} else {
		logger.Warn("阿里云OSS函数计算未启用，无法异步计算MD5值")
		return fmt.Errorf("阿里云OSS函数计算未启用，无法异步计算MD5值")
	}
}

// UploadToBucket 上传文件到指定的存储桶
func (s *AliyunOSSService) UploadToBucket(file io.Reader, objectKey string, regionCode string, bucketName string) (string, error) {
	logger.Info("开始上传文件到指定的存储桶",
		zap.String("objectKey", objectKey),
		zap.String("regionCode", regionCode),
		zap.String("bucketName", bucketName))

	// 验证输入参数
	if file == nil {
		logger.Error("上传文件失败：文件流为空")
		return "", fmt.Errorf("文件流不能为空")
	}
	if objectKey == "" {
		logger.Error("上传文件失败：对象键为空")
		return "", fmt.Errorf("对象键不能为空")
	}
	if regionCode == "" {
		logger.Error("上传文件失败：区域代码为空")
		return "", fmt.Errorf("区域代码不能为空")
	}
	if bucketName == "" {
		logger.Error("上传文件失败：存储桶名称为空")
		return "", fmt.Errorf("存储桶名称不能为空")
	}

	// 获取正确的endpoint（考虑传输加速）
	endpoint := s.getEndpoint(regionCode)

	// 安全地显示AccessKeyID，保护密钥安全
	maskedAccessKeyID := s.config.AccessKeyID
	if len(maskedAccessKeyID) > 8 {
		maskedAccessKeyID = maskedAccessKeyID[:8] + "***"
	} else if len(maskedAccessKeyID) > 3 {
		maskedAccessKeyID = maskedAccessKeyID[:3] + "***"
	} else {
		maskedAccessKeyID = "***"
	}

	logger.Info("创建OSS客户端",
		zap.String("endpoint", endpoint),
		zap.String("regionCode", regionCode),
		zap.Bool("transferAccelerate", s.config.TransferAccelerate.Enabled),
		zap.String("accelerateType", s.config.TransferAccelerate.Type),
		zap.String("accessKeyID", maskedAccessKeyID))

	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		logger.Error("创建OSS客户端失败",
			zap.String("endpoint", endpoint),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return "", fmt.Errorf("创建OSS客户端失败: %w", err)
	}
	logger.Info("OSS客户端创建成功")

	// 获取指定的存储桶
	logger.Info("获取存储桶", zap.String("bucketName", bucketName))
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		logger.Error("获取存储桶失败",
			zap.String("bucketName", bucketName),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return "", fmt.Errorf("获取存储桶失败: %w", err)
	}
	logger.Info("存储桶获取成功", zap.String("bucketName", bucketName))

	// 上传文件
	logger.Info("开始上传文件",
		zap.String("objectKey", objectKey),
		zap.String("bucketName", bucketName))

	// 设置Content-Disposition为attachment，强制下载而不是预览
	options := []oss.Option{
		oss.ContentDisposition("attachment"),
	}

	err = bucket.PutObject(objectKey, file, options...)
	if err != nil {
		logger.Error("上传文件失败",
			zap.String("objectKey", objectKey),
			zap.String("bucketName", bucketName),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return "", fmt.Errorf("上传文件失败: %w", err)
	}
	logger.Info("文件上传成功", zap.String("objectKey", objectKey))

	// 生成文件访问URL，过期时间设置为24小时
	logger.Info("生成文件访问URL",
		zap.String("objectKey", objectKey),
		zap.Int64("expireSeconds", 24*3600))

	url, err := bucket.SignURL(objectKey, oss.HTTPGet, 24*3600)
	if err != nil {
		logger.Error("生成文件URL失败",
			zap.String("objectKey", objectKey),
			zap.String("bucketName", bucketName),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return "", fmt.Errorf("生成文件URL失败: %w", err)
	}

	// 安全地显示URL，避免日志过长
	displayURL := url
	if len(url) > 100 {
		displayURL = url[:100] + "..."
	}

	logger.Info("文件上传完成",
		zap.String("objectKey", objectKey),
		zap.String("bucketName", bucketName),
		zap.String("regionCode", regionCode),
		zap.String("url", displayURL))

	return url, nil
}

// UploadToBucketWithProgress 上传文件到指定的存储桶并回调上传进度
func (s *AliyunOSSService) UploadToBucketWithProgress(file io.Reader, objectKey string, regionCode string, bucketName string, progressCallback func(consumedBytes, totalBytes int64)) (string, error) {
	listener := &progressListener{callback: progressCallback}

	endpoint := s.getEndpoint(regionCode)
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		return "", err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", err
	}
	options := []oss.Option{
		oss.Progress(listener),
		oss.ContentDisposition("attachment"),
	}
	if err := bucket.PutObject(objectKey, file, options...); err != nil {
		return "", err
	}
	url, err := bucket.SignURL(objectKey, oss.HTTPGet, 24*3600)
	if err != nil {
		return "", err
	}
	return url, nil
}

type progressListener struct {
	callback func(consumedBytes, totalBytes int64)
}

func (pl *progressListener) ProgressChanged(event *oss.ProgressEvent) {
	if pl.callback != nil {
		pl.callback(event.ConsumedBytes, event.TotalBytes)
	}
}

// InitMultipartUploadToBucket 初始化分片上传到指定的存储桶
func (s *AliyunOSSService) InitMultipartUploadToBucket(objectKey string, regionCode string, bucketName string) (string, []string, error) {
	logger.Info("初始化分片上传到指定的存储桶",
		zap.String("objectKey", objectKey),
		zap.String("regionCode", regionCode),
		zap.String("bucketName", bucketName))

	// 获取正确的endpoint（考虑传输加速）
	endpoint := s.getEndpoint(regionCode)
	logger.Info("创建分片上传OSS客户端",
		zap.String("endpoint", endpoint),
		zap.Bool("transferAccelerate", s.config.TransferAccelerate.Enabled))

	// 创建指定地域的客户端
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		logger.Error("创建分片上传OSS客户端失败", zap.Error(err))
		return "", nil, fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取指定的存储桶
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", nil, fmt.Errorf("获取存储桶失败: %w", err)
	}

	// 初始化分片上传，设置Content-Disposition为attachment
	result, err := bucket.InitiateMultipartUpload(objectKey, oss.ContentDisposition("attachment"))
	if err != nil {
		return "", nil, fmt.Errorf("初始化分片上传失败: %w", err)
	}

	// 生成分片上传URL
	urls := make([]string, 0)
	for i := 1; i <= 100; i++ { // 假设最多100个分片
		options := []oss.Option{
			oss.AddParam("uploadId", result.UploadID),
			oss.AddParam("partNumber", strconv.Itoa(i)),
			// The Content-Type header must be included in the
			// signature, otherwise OSS will report
			// "SignatureDoesNotMatch" when the client sets this
			// header during upload.
			oss.ContentType("application/octet-stream"),
		}
		url, err := bucket.SignURL(objectKey, oss.HTTPPut, 3600, options...)
		if err != nil {
			return "", nil, fmt.Errorf("生成分片上传URL失败: %w", err)
		}
		urls = append(urls, url)
	}

	return result.UploadID, urls, nil
}

// CompleteMultipartUploadToBucket 完成分片上传到指定的存储桶
func (s *AliyunOSSService) CompleteMultipartUploadToBucket(objectKey string, uploadID string, parts []Part, regionCode string, bucketName string) (string, error) {
	logger.Info("完成分片上传到指定的存储桶",
		zap.String("objectKey", objectKey),
		zap.String("uploadID", uploadID),
		zap.String("regionCode", regionCode),
		zap.String("bucketName", bucketName),
		zap.Int("partsCount", len(parts)))

	// 获取正确的endpoint（考虑传输加速）
	endpoint := s.getEndpoint(regionCode)
	logger.Info("创建完成分片上传OSS客户端",
		zap.String("endpoint", endpoint),
		zap.Bool("transferAccelerate", s.config.TransferAccelerate.Enabled))

	// 创建指定地域的客户端
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		logger.Error("创建完成分片上传OSS客户端失败", zap.Error(err))
		return "", fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取指定的存储桶
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", fmt.Errorf("获取存储桶失败: %w", err)
	}

	// 转换parts为阿里云OSS的Part类型
	ossParts := make([]oss.UploadPart, len(parts))
	for i, part := range parts {
		ossParts[i] = oss.UploadPart{
			PartNumber: part.PartNumber,
			ETag:       part.ETag,
		}
	}

	// 完成分片上传
	_, err = bucket.CompleteMultipartUpload(oss.InitiateMultipartUploadResult{
		Key:      objectKey,
		UploadID: uploadID,
	}, ossParts)
	if err != nil {
		return "", fmt.Errorf("完成分片上传失败: %w", err)
	}

	// 生成文件访问URL
	url, err := bucket.SignURL(objectKey, oss.HTTPGet, 24*3600)
	if err != nil {
		return "", fmt.Errorf("生成文件URL失败: %w", err)
	}

	return url, nil
}

// GetDownloadURL 获取文件下载URL
func (s *AliyunOSSService) GetDownloadURL(objectKey string, expires time.Duration) (string, error) {
	// 生成文件访问URL
	url, err := s.bucket.SignURL(objectKey, oss.HTTPGet, int64(expires.Seconds()))
	if err != nil {
		return "", fmt.Errorf("生成文件URL失败: %w", err)
	}

	return url, nil
}

// GenerateDownloadURLWithBucket 生成指定 bucket 的下载链接
func (s *AliyunOSSService) GenerateDownloadURLWithBucket(objectKey string, downloadURL string, expiration time.Duration) (string, time.Time, error) {
	// downloadURL="https://iotdb-backup.oss-cn-hangzhou.aliyuncs.com/20250605%2F175218_13a29526-454f-40ca-bafd-357db0907690.pdf?Expires=1749203558&OSSAccessKeyId=LTAIW2v6S2BYAhZV&Signature=NCmqjIZ5GWrNbPuQrHVqwvvmF6c%3D"
	parsedURL, err := url.Parse(downloadURL)
	if err != nil {
		panic(err)
	}
	hostParts := strings.Split(parsedURL.Host, ".")
	if len(hostParts) < 4 {
		panic("URL host format unexpected")
	}
	bucketName := hostParts[0]
	regionName := hostParts[1]
	//regi := strings.Split(regionName, "-")
	//regionCode := regi[1] + "-" + regi[2]
	endpoint := fmt.Sprintf("https://%s.aliyuncs.com", regionName)
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		return "", time.Time{}, err
	}
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		return "", time.Time{}, err
	}
	expires := time.Now().Add(expiration)
	signedURL, err := bucket.SignURL(objectKey, oss.HTTPGet, int64(expiration.Seconds()))
	if err != nil {
		return "", time.Time{}, err
	}
	return signedURL, expires, nil
}

// DeleteObjectFromBucket 删除指定存储桶中的文件
func (s *AliyunOSSService) DeleteObjectFromBucket(objectKey string, regionCode string, bucketName string) error {
	logger.Info("开始删除指定存储桶中的文件",
		zap.String("objectKey", objectKey),
		zap.String("regionCode", regionCode),
		zap.String("bucketName", bucketName))

	// 验证输入参数
	if objectKey == "" {
		logger.Error("删除文件失败：对象键为空")
		return fmt.Errorf("对象键不能为空")
	}
	if regionCode == "" {
		logger.Error("删除文件失败：区域代码为空")
		return fmt.Errorf("区域代码不能为空")
	}
	if bucketName == "" {
		logger.Error("删除文件失败：存储桶名称为空")
		return fmt.Errorf("存储桶名称不能为空")
	}

	// 获取正确的endpoint（考虑传输加速）
	endpoint := s.getEndpoint(regionCode)

	logger.Info("创建OSS客户端用于删除文件",
		zap.String("endpoint", endpoint),
		zap.String("regionCode", regionCode),
		zap.Bool("transferAccelerate", s.config.TransferAccelerate.Enabled))

	// 创建指定地域的客户端
	client, err := oss.New(endpoint, s.config.AccessKeyID, s.config.AccessKeySecret)
	if err != nil {
		logger.Error("创建OSS客户端失败",
			zap.String("endpoint", endpoint),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return fmt.Errorf("创建OSS客户端失败: %w", err)
	}

	// 获取指定的存储桶
	logger.Info("获取存储桶用于删除文件", zap.String("bucketName", bucketName))
	bucket, err := client.Bucket(bucketName)
	if err != nil {
		logger.Error("获取存储桶失败",
			zap.String("bucketName", bucketName),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return fmt.Errorf("获取存储桶失败: %w", err)
	}

	// 删除文件
	logger.Info("开始删除文件",
		zap.String("objectKey", objectKey),
		zap.String("bucketName", bucketName))

	err = bucket.DeleteObject(objectKey)
	if err != nil {
		logger.Error("删除文件失败",
			zap.String("objectKey", objectKey),
			zap.String("bucketName", bucketName),
			zap.String("regionCode", regionCode),
			zap.Error(err))
		return fmt.Errorf("删除文件失败: %w", err)
	}

	logger.Info("文件删除成功",
		zap.String("objectKey", objectKey),
		zap.String("bucketName", bucketName),
		zap.String("regionCode", regionCode))

	return nil
}
