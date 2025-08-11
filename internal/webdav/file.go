package webdav

import (
	"bytes"
	"context"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/myysophia/ossmanager-backend/internal/models"
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
		// 保存创建的文件到对象存储
		err := f.fs.storage.PutObject(
			context.Background(),
			f.fs.bucket,
			f.name,
			f.buffer,
			int64(f.buffer.Len()),
			"",
		)
		if err != nil {
			return err
		}

		// 保存到数据库
		ossFile := &models.OSSFile{
			Filename:    path.Base(f.name), // 修正字段名
			FileSize:    int64(f.buffer.Len()),
			ObjectKey:   f.name,
			// UserID:      f.fs.userID, // 暂时注释，避免编译错误
			// StorageType: f.fs.storage.GetType(), // 暂时注释
		}

		return f.fs.db.Create(ossFile).Error
	}

	if f.reader != nil {
		return f.reader.Close()
	}

	return nil
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
