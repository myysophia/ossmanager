package mocks

import (
	"io"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/myysophia/ossmanager/internal/oss"
)

// MockStorageService 模拟存储服务
type MockStorageService struct {
	mock.Mock
}

// GetType 获取存储类型
func (m *MockStorageService) GetType() string {
	args := m.Called()
	return args.String(0)
}

// GetName 获取存储服务名称
func (m *MockStorageService) GetName() string {
	args := m.Called()
	return args.String(0)
}

// GetBucketName 获取Bucket名称
func (m *MockStorageService) GetBucketName() string {
	args := m.Called()
	return args.String(0)
}

// Upload 上传文件
func (m *MockStorageService) Upload(reader io.Reader, objectKey string) (string, error) {
	args := m.Called(reader, objectKey)
	return args.String(0), args.Error(1)
}

// UploadToBucket 上传文件到指定存储桶
func (m *MockStorageService) UploadToBucket(file io.Reader, objectKey string, regionCode string, bucketName string) (string, error) {
	args := m.Called(file, objectKey, regionCode, bucketName)
	return args.String(0), args.Error(1)
}

// UploadToBucketWithProgress 上传文件到指定的存储桶并回调上传进度
func (m *MockStorageService) UploadToBucketWithProgress(file io.Reader, objectKey string, regionCode string, bucketName string, progressCallback func(consumedBytes, totalBytes int64)) (string, error) {
	args := m.Called(file, objectKey, regionCode, bucketName, progressCallback)
	return args.String(0), args.Error(1)
}

// GetObject 获取对象
func (m *MockStorageService) GetObject(objectKey string) (io.ReadCloser, error) {
	args := m.Called(objectKey)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// DeleteObject 删除对象
func (m *MockStorageService) DeleteObject(objectKey string) error {
	return nil
}

// DeleteObjectFromBucket 删除指定存储桶中的文件
func (m *MockStorageService) DeleteObjectFromBucket(objectKey string, regionCode string, bucketName string) error {
	return nil
}

// GetObjectInfo 获取对象信息
func (m *MockStorageService) GetObjectInfo(objectKey string) (int64, error) {
	args := m.Called(objectKey)
	return args.Get(0).(int64), args.Error(1)
}

// GenerateDownloadURL 生成下载URL
func (m *MockStorageService) GenerateDownloadURL(objectKey string, expiration time.Duration) (string, time.Time, error) {
	args := m.Called(objectKey, expiration)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

// InitMultipartUpload 初始化分片上传
func (m *MockStorageService) InitMultipartUpload(objectKey string) (string, []string, error) {
	args := m.Called(objectKey)
	return args.String(0), args.Get(1).([]string), args.Error(2)
}

// InitMultipartUploadToBucket 初始化分片上传到指定存储桶
func (m *MockStorageService) InitMultipartUploadToBucket(objectKey string, regionCode string, bucketName string) (string, []string, error) {
	args := m.Called(objectKey, regionCode, bucketName)
	return args.String(0), args.Get(1).([]string), args.Error(2)
}

// CompleteMultipartUpload 完成分片上传
func (m *MockStorageService) CompleteMultipartUpload(objectKey string, uploadID string, parts []oss.Part) (string, error) {
	args := m.Called(objectKey, uploadID, parts)
	return args.String(0), args.Error(1)
}

// CompleteMultipartUploadToBucket 完成分片上传到指定存储桶
func (m *MockStorageService) CompleteMultipartUploadToBucket(objectKey string, uploadID string, parts []oss.Part, regionCode string, bucketName string) (string, error) {
	args := m.Called(objectKey, uploadID, parts, regionCode, bucketName)
	return args.String(0), args.Error(1)
}

// AbortMultipartUpload 取消分片上传
func (m *MockStorageService) AbortMultipartUpload(objectKey string, uploadID string) error {
	args := m.Called(objectKey, uploadID)
	return args.Error(0)
}

func (m *MockStorageService) AbortMultipartUploadToBucket(uploadID string, objectKey string, regionCode string, bucketName string) error {
	args := m.Called(uploadID, objectKey, regionCode, bucketName)
	return args.Error(0)
}

func (m *MockStorageService) ListUploadedPartsToBucket(objectKey string, uploadID string, regionCode string, bucketName string) ([]oss.Part, error) {
	args := m.Called(objectKey, uploadID, regionCode, bucketName)
	return args.Get(0).([]oss.Part), args.Error(1)
}

func (m *MockStorageService) GeneratePartUploadURL(objectKey string, uploadID string, partNumber int, regionCode string, bucketName string) (string, error) {
	args := m.Called(objectKey, uploadID, partNumber, regionCode, bucketName)
	return args.String(0), args.Error(1)
}

// TriggerMD5Calculation 触发MD5计算
func (m *MockStorageService) TriggerMD5Calculation(objectKey string, fileID uint) error {
	args := m.Called(objectKey, fileID)
	return args.Error(0)
}

// GetDownloadURL 获取文件下载URL
func (m *MockStorageService) GetDownloadURL(objectKey string, expires time.Duration) (string, error) {
	args := m.Called(objectKey, expires)
	return args.String(0), args.Error(1)
}

// MockStorageFactory 模拟存储工厂
type MockStorageFactory struct {
	mock.Mock
}

// GetStorageService 获取存储服务
func (m *MockStorageFactory) GetStorageService(storageType string) (oss.StorageService, error) {
	args := m.Called(storageType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(oss.StorageService), args.Error(1)
}

// GetDefaultStorageService 获取默认存储服务
func (m *MockStorageFactory) GetDefaultStorageService() (oss.StorageService, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(oss.StorageService), args.Error(1)
}

// ClearCache 清除缓存
func (m *MockStorageFactory) ClearCache() {
	m.Called()
}
