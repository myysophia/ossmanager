package oss

import (
	"context"
	"io"
)

// StorageServiceAdapter WebDAV存储服务适配器
// 将StorageService接口适配为WebDAV所需的接口
type StorageServiceAdapter struct {
	service StorageService
}

// NewStorageServiceAdapter 创建存储服务适配器
func NewStorageServiceAdapter(service StorageService) *StorageServiceAdapter {
	return &StorageServiceAdapter{service: service}
}

// PutObject 上传对象
func (a *StorageServiceAdapter) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	// 使用StorageService的PutObjectToBucket方法
	return a.service.PutObjectToBucket(bucket, key, reader, size, contentType)
}

// GetObject 获取对象
func (a *StorageServiceAdapter) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return a.service.GetObject(key)
}

// DeleteObject 删除对象
func (a *StorageServiceAdapter) DeleteObject(ctx context.Context, bucket, key string) error {
	return a.service.DeleteObjectFromBucket(key, "", bucket)
}

// CopyObject 复制对象
func (a *StorageServiceAdapter) CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error {
	return a.service.CopyObject(srcBucket, srcKey, dstBucket, dstKey)
}

// ListObjects 列出对象
func (a *StorageServiceAdapter) ListObjects(ctx context.Context, bucket, prefix string, limit int) ([]ObjectInfo, error) {
	return a.service.ListObjects(bucket, prefix, limit)
}

// GetType 获取存储类型
func (a *StorageServiceAdapter) GetType() string {
	return a.service.GetType()
}
