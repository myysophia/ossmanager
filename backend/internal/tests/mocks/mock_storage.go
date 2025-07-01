package mocks

import (
	"io"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/myysophia/ossmanager-backend/internal/oss"
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

// GetObject 获取对象
func (m *MockStorageService) GetObject(objectKey string) (io.ReadCloser, error) {
	args := m.Called(objectKey)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

// DeleteObject 删除对象
func (m *MockStorageService) DeleteObject(objectKey string) error {
	args := m.Called(objectKey)
	return args.Error(0)
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
func (m *MockStorageService) InitMultipartUpload(objectKey string) (string, map[int]string, error) {
	args := m.Called(objectKey)
	return args.String(0), args.Get(1).(map[int]string), args.Error(2)
}

// CompleteMultipartUpload 完成分片上传
func (m *MockStorageService) CompleteMultipartUpload(uploadID string, parts []interface{}, objectKey string) (string, error) {
	args := m.Called(uploadID, parts, objectKey)
	return args.String(0), args.Error(1)
}

// AbortMultipartUpload 取消分片上传
func (m *MockStorageService) AbortMultipartUpload(uploadID string, objectKey string) error {
	args := m.Called(uploadID, objectKey)
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

// MockStorageFactory 模拟存储工厂
type MockStorageFactory struct {
	mock.Mock
}

// GetStorageService 获取存储服务
func (m *MockStorageFactory) GetStorageService(storageType string) (interface{}, error) {
	args := m.Called(storageType)
	return args.Get(0), args.Error(1)
}

// GetDefaultStorageService 获取默认存储服务
func (m *MockStorageFactory) GetDefaultStorageService() (interface{}, error) {
	args := m.Called()
	return args.Get(0), args.Error(1)
}
