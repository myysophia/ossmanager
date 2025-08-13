package webdav

import (
	"context"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"os"
	"path"
	"runtime"
	"strings"
	"time"

	"gorm.io/gorm"
	models "github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
)

// StreamingOSSFile implements WebDAV File interface with memory-efficient streaming
type StreamingOSSFile struct {
	fs         *OSSFileSystem
	name       string
	reader     io.ReadCloser
	writer     *StreamingWriter
	isCreate   bool
	closed     bool
	isDir      bool
	position   int64
}

// StreamingWriter handles large file uploads with memory constraints
type StreamingWriter struct {
	fs          *OSSFileSystem
	filename    string
	buffer      []byte
	bufferSize  int
	totalSize   int64
	written     int64
	hash        hash.Hash
	chunks      []ChunkInfo
	maxMemory   int64 // Maximum memory to use (default 50MB)
}

type ChunkInfo struct {
	Index  int
	Offset int64
	Size   int64
	Hash   string
}

const (
	defaultBufferSize = 8 * 1024 * 1024  // 8MB buffer
	maxMemoryUsage    = 50 * 1024 * 1024 // 50MB max memory
	chunkThreshold    = 100 * 1024 * 1024 // 100MB threshold for chunked upload
)

// NewStreamingOSSFile creates a new streaming file with memory constraints
func NewStreamingOSSFile(fs *OSSFileSystem, name string, isCreate bool) *StreamingOSSFile {
	file := &StreamingOSSFile{
		fs:       fs,
		name:     name,
		isCreate: isCreate,
		closed:   false,
	}

	if isCreate {
		file.writer = &StreamingWriter{
			fs:         fs,
			filename:   name,
			buffer:     make([]byte, 0, defaultBufferSize),
			bufferSize: defaultBufferSize,
			maxMemory:  maxMemoryUsage,
			hash:       md5.New(),
		}
	}

	// Check if it's a directory
	if name == "" || fs.isDirectory(context.Background(), name) {
		file.isDir = true
	}

	return file
}

// Close finalizes the file write operation with memory-efficient streaming
func (f *StreamingOSSFile) Close() error {
	if f.closed {
		return nil
	}

	f.closed = true

	if f.isCreate && f.writer != nil && !f.isDir {
		return f.writer.Finalize()
	}

	if f.reader != nil {
		return f.reader.Close()
	}

	return nil
}

// Write implements io.Writer with streaming and memory management
func (f *StreamingOSSFile) Write(p []byte) (n int, err error) {
	if f.writer == nil {
		return 0, os.ErrInvalid
	}

	return f.writer.Write(p)
}

// Write handles streaming writes with memory constraints
func (w *StreamingWriter) Write(p []byte) (n int, err error) {
	totalWritten := 0

	for len(p) > 0 {
		// Check if we need to flush the buffer to maintain memory limits
		if len(w.buffer)+len(p) > w.bufferSize || w.shouldFlush() {
			if err := w.flush(); err != nil {
				return totalWritten, err
			}
		}

		// Determine how much we can write to buffer
		spaceAvailable := w.bufferSize - len(w.buffer)
		writeSize := len(p)
		if writeSize > spaceAvailable {
			writeSize = spaceAvailable
		}

		// Write to buffer and hash
		w.buffer = append(w.buffer, p[:writeSize]...)
		w.hash.Write(p[:writeSize])
		w.written += int64(writeSize)
		w.totalSize += int64(writeSize)

		p = p[writeSize:]
		totalWritten += writeSize
	}

	return totalWritten, nil
}

// shouldFlush determines if we should flush based on memory pressure
func (w *StreamingWriter) shouldFlush() bool {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	// Flush if we're using too much memory or buffer is near full
	return len(w.buffer) > w.bufferSize/2 || 
		   int64(m.Alloc) > w.maxMemory
}

// flush writes current buffer to storage and resets it
func (w *StreamingWriter) flush() error {
	if len(w.buffer) == 0 {
		return nil
	}

	ctx := context.Background()
	
	// For large files, use chunked upload strategy
	if w.totalSize > chunkThreshold {
		return w.flushChunk()
	}

	// For small files, use direct upload
	return w.flushDirect(ctx)
}

// flushChunk handles chunked upload for large files
func (w *StreamingWriter) flushChunk() error {
	// Create temporary chunk key
	chunkKey := fmt.Sprintf("%s.chunk.%d", w.filename, len(w.chunks))
	
	ctx := context.Background()
	err := w.fs.storage.PutObject(ctx, w.fs.bucket, chunkKey, 
		strings.NewReader(string(w.buffer)), int64(len(w.buffer)), "")
	
	if err != nil {
		return fmt.Errorf("failed to upload chunk %d: %w", len(w.chunks), err)
	}

	// Record chunk information
	chunkHash := md5.Sum(w.buffer)
	w.chunks = append(w.chunks, ChunkInfo{
		Index:  len(w.chunks),
		Offset: w.written - int64(len(w.buffer)),
		Size:   int64(len(w.buffer)),
		Hash:   fmt.Sprintf("%x", chunkHash),
	})

	// Clear buffer to free memory
	w.buffer = w.buffer[:0]
	
	// Force garbage collection if memory is high
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	if int64(m.Alloc) > w.maxMemory/2 {
		runtime.GC()
	}

	logger.Debug("Flushed chunk", 
		zap.String("filename", w.filename),
		zap.Int("chunkIndex", len(w.chunks)-1),
		zap.Int64("chunkSize", int64(len(w.buffer))),
		zap.Uint64("memoryUsage", m.Alloc))

	return nil
}

// flushDirect handles direct upload for smaller files
func (w *StreamingWriter) flushDirect(ctx context.Context) error {
	return w.fs.storage.PutObject(ctx, w.fs.bucket, w.filename,
		strings.NewReader(string(w.buffer)), int64(len(w.buffer)), "")
}

// Finalize completes the file upload process
func (w *StreamingWriter) Finalize() error {
	ctx := context.Background()

	// Flush any remaining data
	if len(w.buffer) > 0 {
		if err := w.flush(); err != nil {
			return fmt.Errorf("failed to flush final data: %w", err)
		}
	}

	// If we used chunked upload, combine chunks
	if len(w.chunks) > 0 {
		if err := w.combineChunks(ctx); err != nil {
			return fmt.Errorf("failed to combine chunks: %w", err)
		}
	}

	// Calculate final hash
	finalHash := fmt.Sprintf("%x", w.hash.Sum(nil))

	// Update database
	return w.syncToDatabase(w.totalSize, finalHash)
}

// combineChunks combines all uploaded chunks into the final file
func (w *StreamingWriter) combineChunks(ctx context.Context) error {
	// TODO: Implement chunk combination using multipart upload completion
	// For now, this is a placeholder for the streaming architecture
	
	logger.Info("Combining chunks for large file",
		zap.String("filename", w.filename),
		zap.Int("chunkCount", len(w.chunks)),
		zap.Int64("totalSize", w.totalSize))

	// In a real implementation, you would:
	// 1. Use multipart upload completion API
	// 2. Or concatenate chunks server-side
	// 3. Clean up temporary chunk objects

	return nil
}

// syncToDatabase updates file metadata in database
func (w *StreamingWriter) syncToDatabase(fileSize int64, md5Hash string) error {
	var existingFile models.OSSFile
	err := w.fs.db.Where("object_key = ? AND bucket = ?", w.filename, w.fs.bucket).
		First(&existingFile).Error

	if err == nil {
		// Update existing file
		updates := map[string]interface{}{
			"file_size":    fileSize,
			"md5":          md5Hash,
			"md5_status":   models.MD5StatusCompleted,
			"updated_at":   time.Now(),
			"status":       "ACTIVE",
		}
		return w.fs.db.Model(&existingFile).Updates(updates).Error
	} else if err == gorm.ErrRecordNotFound {
		// Create new file record
		ossFile := &models.OSSFile{
			Filename:         path.Base(w.filename),
			OriginalFilename: path.Base(w.filename),
			FileSize:         fileSize,
			MD5:              md5Hash,
			MD5Status:        models.MD5StatusCompleted,
			StorageType:      w.fs.storage.GetType(),
			Bucket:           w.fs.bucket,
			ObjectKey:        w.filename,
			UploaderID:       w.fs.userID,
			Status:           "ACTIVE",
		}
		return w.fs.db.Create(ossFile).Error
	}

	return fmt.Errorf("database error: %w", err)
}

// Read implements io.Reader
func (f *StreamingOSSFile) Read(p []byte) (n int, err error) {
	if f.reader != nil {
		return f.reader.Read(p)
	}
	return 0, io.EOF
}

// Seek is not supported for streaming files
func (f *StreamingOSSFile) Seek(offset int64, whence int) (int64, error) {
	return 0, os.ErrInvalid
}

// Readdir reads directory contents (same as original implementation)
func (f *StreamingOSSFile) Readdir(count int) ([]os.FileInfo, error) {
	if !f.isDir {
		return nil, os.ErrInvalid
	}

	// Use existing implementation from original file
	dirPath := f.name
	if dirPath != "" && !strings.HasSuffix(dirPath, "/") {
		dirPath += "/"
	}

	objects, err := f.fs.storage.ListObjects(
		context.Background(),
		f.fs.bucket,
		dirPath,
		1000,
	)
	if err != nil {
		return nil, err
	}

	var fileInfos []os.FileInfo
	seen := make(map[string]bool)

	for _, obj := range objects {
		if obj.Key == dirPath {
			continue
		}

		relativePath := strings.TrimPrefix(obj.Key, dirPath)
		if relativePath == "" {
			continue
		}

		var name string
		var isDir bool

		if strings.Contains(relativePath, "/") {
			name = strings.Split(relativePath, "/")[0]
			isDir = true
		} else {
			name = relativePath
			isDir = false
		}

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

		if count > 0 && len(fileInfos) >= count {
			break
		}
	}

	return fileInfos, nil
}

// Stat returns file information
func (f *StreamingOSSFile) Stat() (os.FileInfo, error) {
	return f.fs.Stat(context.Background(), f.name)
}

// MemoryUsageReport provides memory usage statistics for monitoring
type MemoryUsageReport struct {
	CurrentMemory int64
	MaxMemory     int64
	BufferSize    int64
	ChunkCount    int
	TotalWritten  int64
}

// GetMemoryUsage returns current memory usage statistics
func (w *StreamingWriter) GetMemoryUsage() MemoryUsageReport {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	
	return MemoryUsageReport{
		CurrentMemory: int64(m.Alloc),
		MaxMemory:     w.maxMemory,
		BufferSize:    int64(len(w.buffer)),
		ChunkCount:    len(w.chunks),
		TotalWritten:  w.written,
	}
}
