package oss

import (
	"context"
	"io"
	"time"
)

// ObjectInfo 对象信息
type ObjectInfo struct {
	Key          string
	Size         int64
	LastModified time.Time
}

// Storage WebDAV需要的存储接口
type Storage interface {
	PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
	GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
	DeleteObject(ctx context.Context, bucket, key string) error
	CopyObject(ctx context.Context, srcBucket, srcKey, dstBucket, dstKey string) error
	ListObjects(ctx context.Context, bucket, prefix string, limit int) ([]ObjectInfo, error)
	GetType() string
}
