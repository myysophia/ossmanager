package webdav

import (
	"bytes"
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"gorm.io/gorm"
	models "github.com/myysophia/ossmanager/internal/db/models"
)

// OSSFile 实现 WebDAV File 接口
type OSSFile struct {
	fs       *OSSFileSystem
	name     string
	reader   io.ReadCloser
	buffer   *bytes.Buffer
	isCreate bool
	closed   bool
	isDir    bool
}

// NewOSSFile 创建新的OSS文件对象
func NewOSSFile(fs *OSSFileSystem, name string, isCreate bool) *OSSFile {
	file := &OSSFile{
		fs:       fs,
		name:     name,
		isCreate: isCreate,
		closed:   false,
	}

	if isCreate {
		file.buffer = &bytes.Buffer{}
	}

	// 检查是否是目录
	if name == "" || fs.isDirectory(context.Background(), name) {
		file.isDir = true
	}

	return file
}

// Close 关闭文件
func (f *OSSFile) Close() error {
	if f.closed {
		return nil
	}

	f.closed = true

	if f.isCreate && f.buffer != nil && !f.isDir {
		fileSize := int64(f.buffer.Len())

		// 检查是否是大文件（超过100MB使用分片上传）
		const chunkThreshold = 100 * 1024 * 1024 // 100MB
		if fileSize > chunkThreshold {
			// TODO: 实现分片上传
			return fmt.Errorf("large file upload not implemented yet (size: %d bytes)", fileSize)
		}

		// 小文件直接上传
		err := f.fs.storage.PutObject(
			context.Background(),
			f.fs.bucket,
			f.name,
			f.buffer,
			fileSize,
			"", // content type 留空，让存储服务自行判断
		)
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}

		// 计算文件MD5
		bufferData := f.buffer.Bytes()
		hash := md5.Sum(bufferData)
		md5Hash := fmt.Sprintf("%x", hash)

		// 同步到数据库
		err = f.syncToDatabase(fileSize, md5Hash)
		if err != nil {
			return fmt.Errorf("failed to sync to database: %w", err)
		}
	}

	if f.reader != nil {
		return f.reader.Close()
	}

	return nil
}

// syncToDatabase 同步文件信息到数据库
func (f *OSSFile) syncToDatabase(fileSize int64, md5Hash string) error {
	// 检查文件是否已存在
	var existingFile models.OSSFile
	err := f.fs.db.Where("object_key = ? AND bucket = ?", f.name, f.fs.bucket).
		First(&existingFile).Error

	if err == nil {
		// 文件已存在，更新记录
		updates := map[string]interface{}{
			"file_size":    fileSize,
			"md5":          md5Hash,
			"md5_status":   models.MD5StatusCompleted,
			"updated_at":   time.Now(),
			"status":       "ACTIVE",
		}

		return f.fs.db.Model(&existingFile).Updates(updates).Error
	} else if err == gorm.ErrRecordNotFound {
		// 文件不存在，创建新记录
		ossFile := &models.OSSFile{
			Filename:         path.Base(f.name),
			OriginalFilename: path.Base(f.name),
			FileSize:         fileSize,
			MD5:              md5Hash,
			MD5Status:        models.MD5StatusCompleted,
			StorageType:      f.fs.storage.GetType(),
			Bucket:           f.fs.bucket,
			ObjectKey:        f.name,
			UploaderID:       f.fs.userID,
			Status:           "ACTIVE",
		}

		return f.fs.db.Create(ossFile).Error
	}

	// 其他数据库错误
	return fmt.Errorf("database error: %w", err)
}

// Read 读取文件内容
func (f *OSSFile) Read(p []byte) (n int, err error) {
	if f.reader != nil {
		return f.reader.Read(p)
	}
	return 0, io.EOF
}

// Write 写入文件内容
func (f *OSSFile) Write(p []byte) (n int, err error) {
	if f.buffer != nil {
		return f.buffer.Write(p)
	}
	return 0, os.ErrInvalid
}

// Seek 定位文件位置（不支持）
func (f *OSSFile) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid // WebDAV 通常不需要 seek 操作
}

// Readdir 读取目录内容
func (f *OSSFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, os.ErrInvalid
	}

	// 列出目录内容
	dirPath := f.name
	if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}

	// 获取所有对象
	objects, err := f.fs.storage.ListObjects(
		context.Background(),
		f.fs.bucket,
		dirPath,
		1000, // 获取更多对象，后面会过滤
	)
	if err != nil {
		return nil, err
	}

	var fileInfos []os.FileInfo
	seen := make(map[string]bool)

	for _, obj := range objects {
		if obj.Key == dirPath {
			continue // 跳过目录本身
		}

		// 获取相对路径
		relativePath := strings.TrimPrefix(obj.Key, dirPath)
		if relativePath == "" {
			continue
		}

		var name string
		var isDir bool

		// 检查是否是直接子项
		if strings.Contains(relativePath, "/") {
			// 这是一个子目录中的文件，只显示子目录
			name = strings.Split(relativePath, "/")[0]
			isDir = true
		} else {
			// 这是一个直接文件
			name = relativePath
			isDir = false
		}

		// 避免重复添加相同的目录
		if seen[name] {
			continue
		}
		seen[name] = true

		var size int64
		var modTime time.Time

		if isDir {
			size = 0
			modTime = time.Now()
		} else {
			size = obj.Size
			modTime = obj.LastModified
		}

		fileInfo := &OSSFileInfo{
			name:    name,
			size:    size,
			mode:    getFileMode(isDir),
			modTime: modTime,
			isDir:   isDir,
		}

		fileInfos = append(fileInfos, fileInfo)

		// 如果指定了数量限制且已达到，则停止
		if count > 0 && len(fileInfos) >= count {
			break
		}
	}

	return fileInfos, nil
}

// Stat 获取文件信息
func (f *OSSFile) Stat() (os.FileInfo, error) {
	return f.fs.Stat(context.Background(), f.name)
}

// getFileMode 根据是否是目录返回文件模式
func getFileMode(isDir bool) os.FileMode {
	if isDir {
		return os.ModeDir | 0755
	}
	return 0644
}
