package webdav

import (
	"context"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager-backend/internal/models"
	"github.com/myysophia/ossmanager-backend/internal/oss"
)

// 临时类型别名，避免编译错误
// type Storage = oss.StorageService

// OSSFileSystem 实现 WebDAV FileSystem 接口
type OSSFileSystem struct {
	storage oss.Storage
	db      *gorm.DB
	userID  uint
	bucket  string
}

// NewOSSFileSystem 创建新的OSS文件系统
func NewOSSFileSystem(storage oss.Storage, db *gorm.DB, userID uint, bucket string) *OSSFileSystem {
	return &OSSFileSystem{
		storage: storage,
		db:      db,
		userID:  userID,
		bucket:  bucket,
	}
}

// Mkdir 创建目录
func (fs *OSSFileSystem) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	// 在对象存储中创建空对象表示目录
	name = strings.TrimPrefix(name, "/")
	if name != "" && !strings.HasSuffix(name, "/") {
		name += "/"
	}

	// 创建空对象表示目录
	return fs.storage.PutObject(ctx, fs.bucket, name, strings.NewReader(""), 0, "application/x-directory")
}

// OpenFile 打开或创建文件
func (fs *OSSFileSystem) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	name = strings.TrimPrefix(name, "/")

	if flag&os.O_CREATE != 0 {
		// 创建新文件
		return NewOSSFile(fs, name, true), nil
	}

	// 检查是否是目录
	if fs.isDirectory(ctx, name) {
		return NewOSSFile(fs, name, false), nil
	}

	// 打开现有文件
	reader, err := fs.storage.GetObject(ctx, fs.bucket, name)
	if err != nil {
		return nil, err
	}

	file := NewOSSFile(fs, name, false)
	file.reader = reader
	return file, nil
}

// RemoveAll 删除文件或目录
func (fs *OSSFileSystem) RemoveAll(ctx context.Context, name string) error {
	name = strings.TrimPrefix(name, "/")

	// 如果是目录，删除所有子对象
	if strings.HasSuffix(name, "/") || fs.isDirectory(ctx, name) {
		return fs.removeDirectory(ctx, name)
	}

	// 删除单个文件
	err := fs.storage.DeleteObject(ctx, fs.bucket, name)
	if err != nil {
		return err
	}

	// 从数据库中删除记录
	return fs.db.Where("object_key = ? AND user_id = ?", name, fs.userID).
		Delete(&models.OSSFile{}).Error
}

// Rename 重命名文件或目录
func (fs *OSSFileSystem) Rename(ctx context.Context, oldName, newName string) error {
	oldName = strings.TrimPrefix(oldName, "/")
	newName = strings.TrimPrefix(newName, "/")

	// 复制到新位置
	err := fs.storage.CopyObject(ctx, fs.bucket, oldName, fs.bucket, newName)
	if err != nil {
		return err
	}

	// 删除原文件
	err = fs.storage.DeleteObject(ctx, fs.bucket, oldName)
	if err != nil {
		return err
	}

	// 更新数据库记录
	return fs.db.Model(&models.OSSFile{}).
		Where("object_key = ? AND user_id = ?", oldName, fs.userID).
		Update("object_key", newName).Error
}

// Stat 获取文件信息
func (fs *OSSFileSystem) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	name = strings.TrimPrefix(name, "/")

	if name == "" {
		// 根目录
		return &OSSFileInfo{
			name:    "/",
			size:    0,
			mode:    os.ModeDir | 0755,
			modTime: time.Now(),
			isDir:   true,
		}, nil
	}

	// 检查是否是目录
	if fs.isDirectory(ctx, name) {
		return &OSSFileInfo{
			name:    path.Base(name),
			size:    0,
			mode:    os.ModeDir | 0755,
			modTime: time.Now(),
			isDir:   true,
		}, nil
	}

	// 查询数据库获取文件信息
	var ossFile models.OSSFile
	err := fs.db.Where("object_key = ? AND user_id = ?", name, fs.userID).
		First(&ossFile).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// 尝试从对象存储直接获取
			objects, err := fs.storage.ListObjects(ctx, fs.bucket, name, 1)
			if err != nil || len(objects) == 0 {
				return nil, os.ErrNotExist
			}
			obj := objects[0]
			return &OSSFileInfo{
				name:    path.Base(name),
				size:    obj.Size,
				mode:    0644,
				modTime: obj.LastModified,
				isDir:   false,
			}, nil
		}
		return nil, err
	}

	return &OSSFileInfo{
		name:    path.Base(name),
		size:    ossFile.FileSize,
		mode:    0644,
		modTime: ossFile.UpdatedAt,
		isDir:   false,
	}, nil
}

// 辅助方法：检查是否是目录
func (fs *OSSFileSystem) isDirectory(ctx context.Context, name string) bool {
	if name == "" {
		return true
	}

	// 检查名称是否以 / 结尾
	if strings.HasSuffix(name, "/") {
		return true
	}

	// 检查是否存在以该路径为前缀的对象
	prefix := name
	if !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	objects, err := fs.storage.ListObjects(ctx, fs.bucket, prefix, 1)
	if err != nil {
		return false
	}

	return len(objects) > 0
}

// 删除目录及其所有内容
func (fs *OSSFileSystem) removeDirectory(ctx context.Context, dirName string) error {
	if dirName != "" && !strings.HasSuffix(dirName, "/") {
		dirName += "/"
	}

	// 列出所有子对象
	objects, err := fs.storage.ListObjects(ctx, fs.bucket, dirName, 1000)
	if err != nil {
		return err
	}

	// 删除所有子对象
	for _, obj := range objects {
		err = fs.storage.DeleteObject(ctx, fs.bucket, obj.Key)
		if err != nil {
			return fmt.Errorf("failed to delete object %s: %v", obj.Key, err)
		}

		// 从数据库中删除记录
		fs.db.Where("object_key = ? AND user_id = ?", obj.Key, fs.userID).
			Delete(&models.OSSFile{})
	}

	return nil
}

// OSSFileInfo 实现 os.FileInfo 接口
type OSSFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
	isDir   bool
}

func (fi *OSSFileInfo) Name() string       { return fi.name }
func (fi *OSSFileInfo) Size() int64        { return fi.size }
func (fi *OSSFileInfo) Mode() os.FileMode  { return fi.mode }
func (fi *OSSFileInfo) ModTime() time.Time { return fi.modTime }
func (fi *OSSFileInfo) IsDir() bool        { return fi.isDir }
func (fi *OSSFileInfo) Sys() interface{}   { return nil }
