package oss

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/myysophia/ossmanager-backend/internal/config"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"go.uber.org/zap"
)

// AWSS3Service AWS S3存储服务
type AWSS3Service struct {
	client     *s3.Client
	config     *config.AWSS3Config
	bucketName string
	uploadDir  string
}

// NewAWSS3Service 创建AWS S3存储服务
func NewAWSS3Service(cfg *config.AWSS3Config) (*AWSS3Service, error) {
	// 创建AWS凭证
	creds := credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"",
	)

	// 创建AWS配置
	awsCfg, err := awsconfig.LoadDefaultConfig(
		context.TODO(),
		awsconfig.WithRegion(cfg.Region),
		awsconfig.WithCredentialsProvider(creds),
	)
	if err != nil {
		logger.Error("创建AWS配置失败", zap.Error(err))
		return nil, fmt.Errorf("创建AWS配置失败: %w", err)
	}

	// 创建S3客户端
	client := s3.NewFromConfig(awsCfg)

	return &AWSS3Service{
		client:     client,
		config:     cfg,
		bucketName: cfg.Bucket,
		uploadDir:  cfg.UploadDir,
	}, nil
}

// GetName 获取存储服务名称
func (s *AWSS3Service) GetName() string {
	return "AWS S3"
}

// GetType 获取存储服务类型
func (s *AWSS3Service) GetType() string {
	return StorageTypeAWSS3
}

// getObjectKey 获取对象键
func (s *AWSS3Service) getObjectKey(filename string) string {
	return path.Join(s.uploadDir, filename)
}

// Upload 上传文件
func (s *AWSS3Service) Upload(file io.Reader, objectKey string) (string, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 上传文件
	_, err := s.client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullObjectKey),
		Body:   file,
	})
	if err != nil {
		logger.Error("AWS S3上传文件失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", fmt.Errorf("上传文件到AWS S3失败: %w", err)
	}

	// 生成预签名URL
	presignClient := s3.NewPresignClient(s.client)
	presignResult, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullObjectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.config.GetOSSURLExpiration()
	})
	if err != nil {
		logger.Error("生成AWS S3下载URL失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", fmt.Errorf("生成AWS S3下载URL失败: %w", err)
	}

	return presignResult.URL, nil
}

// UploadToBucketWithProgress 上传文件到指定的存储桶并回调上传进度
func (s *AWSS3Service) UploadToBucketWithProgress(file io.Reader, objectKey string, regionCode string, bucketName string, progressCallback func(consumedBytes, totalBytes int64)) (string, error) {
	// AWS S3 默认使用配置中的桶，暂不支持回调进度，直接调用 Upload
	return s.Upload(file, objectKey)
}

// InitMultipartUpload 初始化分片上传
func (s *AWSS3Service) InitMultipartUpload(filename string) (string, []string, error) {
	objectKey := s.getObjectKey(filename)

	// 初始化分片上传
	result, err := s.client.CreateMultipartUpload(context.Background(), &s3.CreateMultipartUploadInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		logger.Error("初始化AWS S3分片上传失败", zap.String("filename", filename), zap.Error(err))
		return "", nil, fmt.Errorf("初始化AWS S3分片上传失败: %w", err)
	}

	// 这里返回uploadID，前端需要保存此ID用于后续的分片上传和完成操作
	// 真实场景中，我们还需要根据文件大小计算分片数量，并为每个分片生成上传URL
	// 这里仅作为示例，实际上应该由前端计算分片并请求签名URL
	return *result.UploadId, nil, nil
}

// CompleteMultipartUpload 完成分片上传
func (s *AWSS3Service) CompleteMultipartUpload(uploadID string, parts []Part, objectKey string) (string, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 将我们的Part结构转换为AWS SDK的Part结构
	awsParts := make([]types.CompletedPart, len(parts))
	for i, part := range parts {
		awsParts[i] = types.CompletedPart{
			ETag:       aws.String(part.ETag),
			PartNumber: aws.Int32(int32(part.PartNumber)),
		}
	}

	// 完成分片上传
	_, err := s.client.CompleteMultipartUpload(context.Background(), &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.bucketName),
		Key:      aws.String(fullObjectKey),
		UploadId: aws.String(uploadID),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: awsParts,
		},
	})
	if err != nil {
		logger.Error("完成AWS S3分片上传失败",
			zap.String("objectKey", fullObjectKey),
			zap.String("uploadID", uploadID),
			zap.Error(err))
		return "", fmt.Errorf("完成AWS S3分片上传失败: %w", err)
	}

	// 生成预签名URL
	presignClient := s3.NewPresignClient(s.client)
	presignResult, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullObjectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = s.config.GetOSSURLExpiration()
	})
	if err != nil {
		logger.Error("生成AWS S3下载URL失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", fmt.Errorf("生成AWS S3下载URL失败: %w", err)
	}

	return presignResult.URL, nil
}

// AbortMultipartUpload 取消分片上传
func (s *AWSS3Service) AbortMultipartUpload(uploadID string, objectKey string) error {
	fullObjectKey := s.getObjectKey(objectKey)

	// 取消分片上传
	_, err := s.client.AbortMultipartUpload(context.Background(), &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.bucketName),
		Key:      aws.String(fullObjectKey),
		UploadId: aws.String(uploadID),
	})
	if err != nil {
		logger.Error("取消AWS S3分片上传失败",
			zap.String("objectKey", fullObjectKey),
			zap.String("uploadID", uploadID),
			zap.Error(err))
		return fmt.Errorf("取消AWS S3分片上传失败: %w", err)
	}

	return nil
}

// GenerateDownloadURL 生成下载URL
func (s *AWSS3Service) GenerateDownloadURL(objectKey string, expiration time.Duration) (string, time.Time, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	// 设置过期时间
	expires := time.Now().Add(expiration)

	// 生成预签名URL
	presignClient := s3.NewPresignClient(s.client)
	presignResult, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullObjectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = expiration
	})
	if err != nil {
		logger.Error("生成AWS S3下载URL失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return "", time.Time{}, fmt.Errorf("生成AWS S3下载URL失败: %w", err)
	}

	return presignResult.URL, expires, nil
}

// DeleteObject 删除对象
func (s *AWSS3Service) DeleteObject(objectKey string) error {
	fullObjectKey := s.getObjectKey(objectKey)

	// 删除对象
	_, err := s.client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullObjectKey),
	})
	if err != nil {
		logger.Error("删除AWS S3对象失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return fmt.Errorf("删除AWS S3对象失败: %w", err)
	}

	return nil
}

// GetObjectInfo 获取对象信息
func (s *AWSS3Service) GetObjectInfo(objectKey string) (int64, error) {
	objectKey = s.getObjectKey(objectKey)
	// 获取对象信息
	result, err := s.client.HeadObject(context.TODO(), &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		logger.Error("获取对象信息失败", zap.String("bucket", s.bucketName), zap.String("key", objectKey), zap.Error(err))
		return 0, fmt.Errorf("获取对象信息失败: %w", err)
	}

	// 返回对象大小
	if result.ContentLength != nil {
		return *result.ContentLength, nil
	}
	return 0, nil
}

// GetBucketName 获取存储桶名称
func (s *AWSS3Service) GetBucketName() string {
	return s.bucketName
}

// GetObject 获取对象内容
func (s *AWSS3Service) GetObject(objectKey string) (io.ReadCloser, error) {
	fullObjectKey := s.getObjectKey(objectKey)

	ctx := context.Background()
	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(fullObjectKey),
	})

	if err != nil {
		logger.Error("获取AWS S3对象失败", zap.String("objectKey", fullObjectKey), zap.Error(err))
		return nil, fmt.Errorf("获取AWS S3对象失败: %w", err)
	}

	return resp.Body, nil
}

// TriggerMD5Calculation 触发计算MD5值
func (s *AWSS3Service) TriggerMD5Calculation(objectKey string, fileID uint) error {
	logger.Info("触发AWS S3对象MD5计算",
		zap.String("objectKey", objectKey),
		zap.Uint("fileID", fileID),
		zap.String("bucket", s.bucketName))

	// AWS S3通常使用Lambda函数处理这类异步计算
	// 此处只是记录日志，实际项目中需要集成AWS Lambda
	logger.Warn("AWS S3暂不支持异步MD5计算，需要集成AWS Lambda实现")
	return fmt.Errorf("AWS S3暂不支持异步MD5计算，需要集成AWS Lambda实现")
}
