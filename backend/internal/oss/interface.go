package oss

import (
	"io"
	"time"
)

// 存储类型枚举
const (
	StorageTypeAliyunOSS = "ALIYUN_OSS"
	StorageTypeAWSS3     = "AWS_S3"
	StorageTypeR2        = "CLOUDFLARE_R2"
)

// Part 分片信息
type Part struct {
	PartNumber int    `json:"part_number"`
	ETag       string `json:"etag"`
}

// StorageService 存储服务接口
type StorageService interface {
	// GetName 获取存储服务名称
	GetName() string

	// GetType 获取存储服务类型
	GetType() string

	// GetBucketName 获取存储桶名称
	GetBucketName() string

	// Upload 上传文件到默认存储桶
	Upload(file io.Reader, objectKey string) (string, error)

	// UploadToBucket 上传文件到指定的存储桶
	UploadToBucket(file io.Reader, objectKey string, regionCode string, bucketName string) (string, error)

	// UploadToBucketWithProgress 上传文件到指定的存储桶并回调上传进度
	// progressCallback: 回调函数，参数为已上传字节数和总字节数
	UploadToBucketWithProgress(file io.Reader, objectKey string, regionCode string, bucketName string, progressCallback func(consumedBytes, totalBytes int64)) (string, error)

	// InitMultipartUpload 初始化分片上传
	InitMultipartUpload(objectKey string) (string, []string, error)

	// InitMultipartUploadToBucket 初始化分片上传到指定的存储桶
	InitMultipartUploadToBucket(objectKey string, regionCode string, bucketName string) (string, []string, error)

	// CompleteMultipartUpload 完成分片上传
	CompleteMultipartUpload(objectKey string, uploadID string, parts []Part) (string, error)

	// CompleteMultipartUploadToBucket 完成分片上传到指定的存储桶
	CompleteMultipartUploadToBucket(objectKey string, uploadID string, parts []Part, regionCode string, bucketName string) (string, error)

	// AbortMultipartUpload 取消分片上传
	AbortMultipartUpload(objectKey string, uploadID string) error

	// AbortMultipartUploadToBucket 取消指定存储桶的分片上传
	AbortMultipartUploadToBucket(uploadID string, objectKey string, regionCode string, bucketName string) error

	// ListUploadedPartsToBucket 获取已上传的分片列表
	// objectKey: 对象键
	// uploadID: 上传ID
	// regionCode, bucketName: 指定的地域和存储桶
	// 返回：已上传的分片信息列表, 错误
	ListUploadedPartsToBucket(objectKey string, uploadID string, regionCode string, bucketName string) ([]Part, error)

	// GeneratePartUploadURL 生成单个分片上传的预签名URL
	GeneratePartUploadURL(objectKey string, uploadID string, partNumber int, regionCode string, bucketName string) (string, error)

	// GenerateDownloadURL 生成下载URL
	// objectKey: 对象键
	// expiration: 过期时间
	// 返回：下载URL, 过期时间, 错误
	GenerateDownloadURL(objectKey string, expiration time.Duration) (string, time.Time, error)

	// DeleteObject 删除文件
	DeleteObject(objectKey string) error

	// DeleteObjectFromBucket 删除指定存储桶中的文件
	// objectKey: 对象键
	// regionCode: 区域代码
	// bucketName: 存储桶名称
	// 返回：错误
	DeleteObjectFromBucket(objectKey string, regionCode string, bucketName string) error

	// GetObjectInfo 获取对象信息
	// objectKey: 对象键
	// 返回：对象大小, 错误
	GetObjectInfo(objectKey string) (int64, error)

	// GetObject 获取对象内容
	// objectKey: 对象键
	// 返回：对象内容读取器, 错误
	GetObject(objectKey string) (io.ReadCloser, error)

	// TriggerMD5Calculation 触发计算MD5值
	// objectKey: 对象键
	// fileID: 文件ID
	// 返回：错误
	TriggerMD5Calculation(objectKey string, fileID uint) error

	// GetDownloadURL 获取文件下载URL
	GetDownloadURL(objectKey string, expires time.Duration) (string, error)
}

// StorageFactory 存储服务工厂
type StorageFactory interface {
	// GetStorageService 获取存储服务
	// storageType: 存储类型
	GetStorageService(storageType string) (StorageService, error)

	// GetDefaultStorageService 获取默认存储服务
	GetDefaultStorageService() (StorageService, error)

	// ClearCache 清除缓存
	ClearCache()
}
