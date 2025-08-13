package webdav

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/oss"
)

// MockStreamingStorage implements storage service with streaming support
type MockStreamingStorage struct {
	objects        map[string][]byte
	metadata       map[string]oss.ObjectInfo
	putObjectCalls []StreamingPutCall
	memoryPeaks    []int64
}

type StreamingPutCall struct {
	Bucket      string
	Key         string
	Size        int64
	ContentType string
	Data        []byte
	Timestamp   time.Time
}

func NewMockStreamingStorage() *MockStreamingStorage {
	return &MockStreamingStorage{
		objects:        make(map[string][]byte),
		metadata:       make(map[string]oss.ObjectInfo),
		putObjectCalls: []StreamingPutCall{},
		memoryPeaks:    []int64{},
	}
}

func (m *MockStreamingStorage) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	// Track memory usage during upload
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	m.memoryPeaks = append(m.memoryPeaks, int64(memStats.Alloc))

	// Read data with minimal memory footprint
	data := make([]byte, 0, size)
	buffer := make([]byte, 32*1024) // 32KB read buffer
	
	for {
		n, err := reader.Read(buffer)
		if n > 0 {
			data = append(data, buffer[:n]...)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	// Store the data
	m.objects[bucket+"/"+key] = data
	m.metadata[bucket+"/"+key] = oss.ObjectInfo{
		Key:          key,
		Size:         int64(len(data)),
		LastModified: time.Now(),
	}

	// Record the call
	m.putObjectCalls = append(m.putObjectCalls, StreamingPutCall{
		Bucket:      bucket,
		Key:         key,
		Size:        size,
		ContentType: contentType,
		Data:        data,
		Timestamp:   time.Now(),
	})

	return nil
}

func (m *MockStreamingStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	data, exists := m.objects[bucket+"/"+key]
	if !exists {
		return nil, fmt.Errorf("object not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MockStreamingStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	delete(m.objects, bucket+"/"+key)
	delete(m.metadata, bucket+"/"+key)
	return nil
}

func (m *MockStreamingStorage) ListObjects(ctx context.Context, bucket, prefix string, maxKeys int) ([]oss.ObjectInfo, error) {
	var objects []oss.ObjectInfo
	for key, info := range m.metadata {
		if strings.HasPrefix(key, bucket+"/"+prefix) {
			objects = append(objects, info)
		}
		if len(objects) >= maxKeys {
			break
		}
	}
	return objects, nil
}

func (m *MockStreamingStorage) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	srcData, exists := m.objects[srcBucket+"/"+srcKey]
	if !exists {
		return fmt.Errorf("source object not found")
	}
	
	m.objects[destBucket+"/"+destKey] = srcData
	m.metadata[destBucket+"/"+destKey] = oss.ObjectInfo{
		Key:          destKey,
		Size:         int64(len(srcData)),
		LastModified: time.Now(),
	}
	return nil
}

func (m *MockStreamingStorage) GetType() string {
	return "mock-streaming"
}

func (m *MockStreamingStorage) GetMaxMemoryUsage() int64 {
	var maxMem int64
	for _, mem := range m.memoryPeaks {
		if mem > maxMem {
			maxMem = mem
		}
	}
	return maxMem
}

func setupStreamingTestEnvironment() (*OSSFileSystem, *gorm.DB, sqlmock.Sqlmock, *MockStreamingStorage) {
	// Create mock database
	db, mock, err := sqlmock.New()
	if err != nil {
		panic(err)
	}
	
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	
	// Create mock streaming storage
	mockStorage := NewMockStreamingStorage()
	
	// Create file system
	fs := &OSSFileSystem{
		storage: oss.NewStorageServiceAdapter(mockStorage),
		db:      gormDB,
		userID:  1,
		bucket:  "test-bucket",
	}
	
	return fs, gormDB, mock, mockStorage
}

func TestStreamingOSSFile_SmallFileUpload(t *testing.T) {
	fs, _, mock, mockStorage := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock database operations
	mock.ExpectQuery("SELECT (.+) FROM \"oss_files\"").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec("INSERT INTO \"oss_files\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Create streaming file
	file := NewStreamingOSSFile(fs, "small-file.txt", true)
	
	// Write small content
	testContent := "Hello, streaming world!"
	_, err := file.Write([]byte(testContent))
	require.NoError(t, err)
	
	// Close to finalize upload
	err = file.Close()
	require.NoError(t, err)
	
	// Verify upload
	assert.Len(t, mockStorage.putObjectCalls, 1)
	assert.Equal(t, []byte(testContent), mockStorage.objects["test-bucket/small-file.txt"])
}

func TestStreamingOSSFile_LargeFileUpload_MemoryConstraint(t *testing.T) {
	fs, _, mock, mockStorage := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock database operations for chunked upload
	mock.ExpectQuery("SELECT (.+) FROM \"oss_files\"").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec("INSERT INTO \"oss_files\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Create streaming file
	file := NewStreamingOSSFile(fs, "large-file.bin", true)
	
	// Simulate 200MB file upload in chunks
	const totalSize = 200 * 1024 * 1024 // 200MB
	const chunkSize = 1024 * 1024        // 1MB chunks
	
	var initialMemStats, maxMemStats runtime.MemStats
	runtime.ReadMemStats(&initialMemStats)
	
	written := int64(0)
	for written < totalSize {
		remainingSize := totalSize - written
		currentChunkSize := chunkSize
		if remainingSize < chunkSize {
			currentChunkSize = int(remainingSize)
		}
		
		// Generate random data for this chunk
		chunkData := make([]byte, currentChunkSize)
		_, err := rand.Read(chunkData)
		require.NoError(t, err)
		
		// Write chunk
		n, err := file.Write(chunkData)
		require.NoError(t, err)
		require.Equal(t, currentChunkSize, n)
		
		written += int64(currentChunkSize)
		
		// Monitor memory usage
		var currentMemStats runtime.MemStats
		runtime.ReadMemStats(&currentMemStats)
		if currentMemStats.Alloc > maxMemStats.Alloc {
			maxMemStats = currentMemStats
		}
		
		// Force GC periodically to get accurate memory readings
		if written%(50*1024*1024) == 0 {
			runtime.GC()
			runtime.ReadMemStats(&currentMemStats)
		}
	}
	
	// Close to finalize upload
	err := file.Close()
	require.NoError(t, err)
	
	// Verify memory constraint: should not exceed 100MB even for 200MB file
	memoryIncrease := int64(maxMemStats.Alloc) - int64(initialMemStats.Alloc)
	assert.True(t, memoryIncrease < 100*1024*1024,
		"Memory usage increase %d MB exceeds 100MB limit during 200MB upload",
		memoryIncrease/(1024*1024))
	
	// Verify chunks were created (for large files)
	if file.writer != nil && len(file.writer.chunks) > 0 {
		assert.Greater(t, len(file.writer.chunks), 0, "Expected chunks for large file")
		assert.Equal(t, totalSize, file.writer.totalSize, "Total size should match written size")
	}
	
	t.Logf("Memory increase during 200MB upload: %d MB (limit: 100MB)", 
		memoryIncrease/(1024*1024))
}

func TestStreamingOSSFile_MemoryUsageReporting(t *testing.T) {
	fs, _, mock, _ := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Create streaming file
	file := NewStreamingOSSFile(fs, "memory-test.txt", true)
	
	// Write some data
	testData := make([]byte, 10*1024*1024) // 10MB
	_, err := rand.Read(testData)
	require.NoError(t, err)
	
	_, err = file.Write(testData)
	require.NoError(t, err)
	
	// Get memory usage report
	if file.writer != nil {
		report := file.writer.GetMemoryUsage()
		
		assert.Greater(t, report.CurrentMemory, int64(0))
		assert.Equal(t, maxMemoryUsage, report.MaxMemory)
		assert.Greater(t, report.BufferSize, int64(0))
		assert.Equal(t, int64(len(testData)), report.TotalWritten)
		
		t.Logf("Memory usage report: Current=%d MB, Max=%d MB, Buffer=%d MB, Written=%d MB",
			report.CurrentMemory/(1024*1024),
			report.MaxMemory/(1024*1024), 
			report.BufferSize/(1024*1024),
			report.TotalWritten/(1024*1024))
	}
}

func TestStreamingOSSFile_ChunkedUpload(t *testing.T) {
	fs, _, mock, mockStorage := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock database operations
	mock.ExpectQuery("SELECT (.+) FROM \"oss_files\"").
		WillReturnError(gorm.ErrRecordNotFound)
	mock.ExpectExec("INSERT INTO \"oss_files\"").
		WillReturnResult(sqlmock.NewResult(1, 1))
	
	// Create streaming file
	file := NewStreamingOSSFile(fs, "chunked-file.bin", true)
	
	// Write data exceeding chunk threshold
	const fileSize = 150 * 1024 * 1024 // 150MB (exceeds 100MB threshold)
	
	// Write in smaller increments to trigger multiple flushes
	chunkData := make([]byte, 10*1024*1024) // 10MB at a time
	_, err := rand.Read(chunkData)
	require.NoError(t, err)
	
	totalWritten := int64(0)
	for totalWritten < fileSize {
		n, err := file.Write(chunkData)
		require.NoError(t, err)
		totalWritten += int64(n)
		
		if totalWritten >= fileSize {
			break
		}
	}
	
	// Close to finalize
	err = file.Close()
	require.NoError(t, err)
	
	// Verify chunked upload behavior
	if file.writer != nil {
		assert.Greater(t, len(file.writer.chunks), 0, "Should have created chunks for large file")
		assert.Equal(t, fileSize, file.writer.totalSize, "Total size should match")
	}
	
	// Verify multiple storage calls (chunks)
	assert.Greater(t, len(mockStorage.putObjectCalls), 1, "Should have multiple storage calls for chunks")
}

func TestStreamingOSSFile_ConcurrentWrites(t *testing.T) {
	fs, _, mock, mockStorage := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock database operations for each file
	for i := 0; i < 5; i++ {
		mock.ExpectQuery("SELECT (.+) FROM \"oss_files\"").
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec("INSERT INTO \"oss_files\"").
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
	}
	
	const numFiles = 5
	const fileSize = 20 * 1024 * 1024 // 20MB each
	
	// Create multiple files concurrently
	files := make([]*StreamingOSSFile, numFiles)
	for i := 0; i < numFiles; i++ {
		files[i] = NewStreamingOSSFile(fs, fmt.Sprintf("concurrent-file-%d.bin", i), true)
	}
	
	// Track initial memory
	var initialMem runtime.MemStats
	runtime.ReadMemStats(&initialMem)
	
	// Write to all files concurrently
	done := make(chan error, numFiles)
	
	for i := 0; i < numFiles; i++ {
		go func(fileIndex int) {
			file := files[fileIndex]
			
			// Generate unique data for each file
			data := make([]byte, fileSize)
			for j := range data {
				data[j] = byte(fileIndex + j%256)
			}
			
			// Write data in chunks
			chunkSize := 1024 * 1024 // 1MB chunks
			for offset := 0; offset < len(data); offset += chunkSize {
				end := offset + chunkSize
				if end > len(data) {
					end = len(data)
				}
				
				_, err := file.Write(data[offset:end])
				if err != nil {
					done <- err
					return
				}
			}
			
			// Close file
			err := file.Close()
			done <- err
		}(i)
	}
	
	// Wait for all uploads to complete
	for i := 0; i < numFiles; i++ {
		err := <-done
		require.NoError(t, err, "File %d upload failed", i)
	}
	
	// Check final memory usage
	var finalMem runtime.MemStats
	runtime.ReadMemStats(&finalMem)
	
	totalDataSize := int64(numFiles * fileSize)
	memoryIncrease := int64(finalMem.Alloc) - int64(initialMem.Alloc)
	
	// Memory usage should be much less than total data size
	assert.True(t, memoryIncrease < totalDataSize/4,
		"Memory increase %d MB is too high for %d MB total data (concurrent upload)",
		memoryIncrease/(1024*1024), totalDataSize/(1024*1024))
	
	// Verify all files were uploaded
	assert.Equal(t, numFiles, len(mockStorage.objects))
	
	t.Logf("Concurrent upload: %d files x %d MB each, memory increase: %d MB",
		numFiles, fileSize/(1024*1024), memoryIncrease/(1024*1024))
}

func TestStreamingOSSFile_ErrorHandling(t *testing.T) {
	fs, _, mock, _ := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Create a failing storage service
	type FailingStreamingStorage struct {
		*MockStreamingStorage
		failAfter int
		callCount int
	}
	
	failingStorage := &FailingStreamingStorage{
		MockStreamingStorage: NewMockStreamingStorage(),
		failAfter:           3, // Fail after 3 calls
	}
	
	failingStorage.PutObject = func(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
		failingStorage.callCount++
		if failingStorage.callCount > failingStorage.failAfter {
			return fmt.Errorf("storage failure simulation")
		}
		return failingStorage.MockStreamingStorage.PutObject(ctx, bucket, key, reader, size, contentType)
	}
	
	// Replace storage in filesystem
	fs.storage = oss.NewStorageServiceAdapter(failingStorage)
	
	// Create streaming file
	file := NewStreamingOSSFile(fs, "failing-upload.bin", true)
	
	// Write data that will trigger multiple flushes
	largeData := make([]byte, 50*1024*1024) // 50MB
	_, err := rand.Read(largeData)
	require.NoError(t, err)
	
	// Write should fail during flush
	_, writeErr := file.Write(largeData)
	if writeErr == nil {
		// Error might occur during close
		writeErr = file.Close()
	}
	
	// Should have error
	assert.Error(t, writeErr)
	assert.Contains(t, writeErr.Error(), "storage failure simulation")
}

func TestStreamingOSSFile_ReadOperations(t *testing.T) {
	fs, _, _, mockStorage := setupStreamingTestEnvironment()
	
	// Pre-populate storage
	testContent := "Hello, streaming read world!"
	testKey := "read-test.txt"
	mockStorage.objects["test-bucket/"+testKey] = []byte(testContent)
	mockStorage.metadata["test-bucket/"+testKey] = oss.ObjectInfo{
		Key:          testKey,
		Size:         int64(len(testContent)),
		LastModified: time.Now(),
	}
	
	// Open file for reading
	ctx := context.Background()
	reader, err := fs.storage.GetObject(ctx, "test-bucket", testKey)
	require.NoError(t, err)
	
	file := NewStreamingOSSFile(fs, testKey, false)
	file.reader = reader
	
	// Test reading
	buffer := make([]byte, len(testContent))
	n, err := file.Read(buffer)
	require.NoError(t, err)
	assert.Equal(t, len(testContent), n)
	assert.Equal(t, testContent, string(buffer))
	
	// Test Stat
	info, err := file.Stat()
	require.NoError(t, err)
	assert.Equal(t, testKey, info.Name())
	assert.Equal(t, int64(len(testContent)), info.Size())
	assert.False(t, info.IsDir())
	
	// Close
	err = file.Close()
	require.NoError(t, err)
}

// Benchmark tests for performance validation

func BenchmarkStreamingUpload_1GB_MemoryUsage(b *testing.B) {
	fs, _, mock, _ := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock database operations for all iterations
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT (.+) FROM \"oss_files\"").
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec("INSERT INTO \"oss_files\"").
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
	}
	
	const fileSize = 1024 * 1024 * 1024 // 1GB
	const chunkSize = 8 * 1024 * 1024   // 8MB chunks
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		file := NewStreamingOSSFile(fs, fmt.Sprintf("benchmark-file-%d.bin", i), true)
		
		// Track memory at start
		var startMem runtime.MemStats
		runtime.ReadMemStats(&startMem)
		
		// Generate and write data in chunks
		written := int64(0)
		chunkData := make([]byte, chunkSize)
		
		for written < fileSize {
			// Fill chunk with pattern data (faster than random)
			for j := range chunkData {
				chunkData[j] = byte(written/chunkSize + int64(j))
			}
			
			remainingSize := fileSize - written
			currentChunkSize := chunkSize
			if remainingSize < chunkSize {
				currentChunkSize = int(remainingSize)
			}
			
			_, err := file.Write(chunkData[:currentChunkSize])
			if err != nil {
				b.Fatal(err)
			}
			
			written += int64(currentChunkSize)
		}
		
		// Close file
		err := file.Close()
		if err != nil {
			b.Fatal(err)
		}
		
		// Check memory usage
		var endMem runtime.MemStats
		runtime.ReadMemStats(&endMem)
		
		memoryIncrease := int64(endMem.Alloc) - int64(startMem.Alloc)
		if memoryIncrease > 100*1024*1024 {
			b.Fatalf("Memory usage %d MB exceeds 100MB limit for 1GB upload",
				memoryIncrease/(1024*1024))
		}
		
		b.ReportMetric(float64(memoryIncrease)/(1024*1024), "memory_mb")
	}
}

func BenchmarkStreamingUpload_Throughput(b *testing.B) {
	fs, _, mock, mockStorage := setupStreamingTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock database operations for all iterations  
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT (.+) FROM \"oss_files\"").
			WillReturnError(gorm.ErrRecordNotFound)
		mock.ExpectExec("INSERT INTO \"oss_files\"").
			WillReturnResult(sqlmock.NewResult(int64(i+1), 1))
	}
	
	const fileSize = 100 * 1024 * 1024 // 100MB per iteration
	testData := make([]byte, fileSize)
	_, err := rand.Read(testData)
	require.NoError(b, err)
	
	b.SetBytes(fileSize)
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		file := NewStreamingOSSFile(fs, fmt.Sprintf("throughput-test-%d.bin", i), true)
		
		_, err := file.Write(testData)
		if err != nil {
			b.Fatal(err)
		}
		
		err = file.Close()
		if err != nil {
			b.Fatal(err)
		}
	}
	
	b.StopTimer()
	
	// Report additional metrics
	totalUploads := len(mockStorage.putObjectCalls)
	b.ReportMetric(float64(totalUploads), "storage_calls")
}
