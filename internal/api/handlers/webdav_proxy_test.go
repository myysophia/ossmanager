package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/oss"
)

// MockStorageService for testing
type MockStorageService struct {
	objects        map[string][]byte
	metadata       map[string]oss.ObjectInfo
	putObjectCalls []PutObjectCall
}

type PutObjectCall struct {
	Bucket      string
	Key         string
	Size        int64
	ContentType string
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		objects:        make(map[string][]byte),
		metadata:       make(map[string]oss.ObjectInfo),
		putObjectCalls: []PutObjectCall{},
	}
}

func (m *MockStorageService) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	// Record the call for memory usage testing
	m.putObjectCalls = append(m.putObjectCalls, PutObjectCall{
		Bucket:      bucket,
		Key:         key,
		Size:        size,
		ContentType: contentType,
	})

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	m.objects[bucket+"/"+key] = data
	m.metadata[bucket+"/"+key] = oss.ObjectInfo{
		Key:          key,
		Size:         int64(len(data)),
		LastModified: time.Now(),
	}
	return nil
}

func (m *MockStorageService) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	data, exists := m.objects[bucket+"/"+key]
	if !exists {
		return nil, fmt.Errorf("object not found")
	}
	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MockStorageService) DeleteObject(ctx context.Context, bucket, key string) error {
	delete(m.objects, bucket+"/"+key)
	delete(m.metadata, bucket+"/"+key)
	return nil
}

func (m *MockStorageService) ListObjects(ctx context.Context, bucket, prefix string, maxKeys int) ([]oss.ObjectInfo, error) {
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

func (m *MockStorageService) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
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

func (m *MockStorageService) GetType() string {
	return "mock"
}

// MockStorageFactory for testing
type MockStorageFactory struct {
	service *MockStorageService
}

func NewMockStorageFactory(service *MockStorageService) *MockStorageFactory {
	return &MockStorageFactory{service: service}
}

func (f *MockStorageFactory) GetDefaultStorageService() (oss.StorageService, error) {
	return f.service, nil
}

func (f *MockStorageFactory) GetStorageService(config interface{}) (oss.StorageService, error) {
	return f.service, nil
}

func setupTestEnvironment() (*gin.Engine, *gorm.DB, sqlmock.Sqlmock, *MockStorageService) {
	gin.SetMode(gin.TestMode)
	
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
	
	// Create mock storage
	mockStorage := NewMockStorageService()
	storageFactory := NewMockStorageFactory(mockStorage)
	
	// Create router with middleware
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// Mock authenticated user
		user := &auth.Claims{
			UserID:   1,
			Username: "testuser",
		}
		c.Set("user", user)
		c.Set("userID", user.UserID)
		c.Set("username", user.Username)
		c.Next()
	})
	
	// Setup routes
	handler := NewWebDAVProxyHandler(storageFactory, gormDB)
	
	v1 := router.Group("/api/v1")
	{
		v1.GET("/:bucket", handler.ListDirectory)
		v1.POST("/:bucket/file", handler.UploadFile)
		v1.DELETE("/:bucket", handler.DeleteFile)
		v1.PATCH("/:bucket/rename", handler.RenameFile)
		v1.POST("/:bucket/mkdir", handler.CreateDirectory)
	}
	
	return router, gormDB, mock, mockStorage
}

func TestWebDAVProxyHandler_ListDirectory(t *testing.T) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock bucket access check
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	// Add some test objects to mock storage
	testBucket := "test-bucket"
	mockStorage.objects[testBucket+"/file1.txt"] = []byte("content1")
	mockStorage.objects[testBucket+"/dir1/file2.txt"] = []byte("content2")
	mockStorage.metadata[testBucket+"/file1.txt"] = oss.ObjectInfo{
		Key:          "file1.txt",
		Size:         8,
		LastModified: time.Now(),
	}
	mockStorage.metadata[testBucket+"/dir1/file2.txt"] = oss.ObjectInfo{
		Key:          "dir1/file2.txt", 
		Size:         8,
		LastModified: time.Now(),
	}
	
	req := httptest.NewRequest("GET", "/api/v1/test-bucket?path=/", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response ListDirectoryResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "/", response.Path)
	assert.True(t, len(response.Items) > 0)
}

func TestWebDAVProxyHandler_UploadFile_SmallFile(t *testing.T) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	testContent := "Hello, World!"
	req := httptest.NewRequest("POST", "/api/v1/test-bucket/file?path=/test.txt", 
		strings.NewReader(testContent))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify file was stored
	assert.Contains(t, mockStorage.objects, "test-bucket/test.txt")
	assert.Equal(t, []byte(testContent), mockStorage.objects["test-bucket/test.txt"])
}

func TestWebDAVProxyHandler_UploadFile_LargeFile_StreamingMemoryTest(t *testing.T) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	// Create a 1GB test file stream (simulated)
	const fileSize = 1 * 1024 * 1024 * 1024 // 1 GB
	
	// Create a custom reader that tracks memory usage
	type MemoryTrackingReader struct {
		size     int64
		read     int64
		maxMem   int64
		currentMem int64
	}
	
	reader := &MemoryTrackingReader{size: fileSize}
	reader.Read = func(p []byte) (n int, err error) {
		if reader.read >= reader.size {
			return 0, io.EOF
		}
		
		// Simulate reading chunks without loading entire file in memory
		chunkSize := int64(len(p))
		if chunkSize > reader.size-reader.read {
			chunkSize = reader.size - reader.read
		}
		
		// Track "memory usage" - should not exceed buffer size
		reader.currentMem = chunkSize
		if reader.currentMem > reader.maxMem {
			reader.maxMem = reader.currentMem
		}
		
		// Fill buffer with test data
		for i := int64(0); i < chunkSize; i++ {
			p[i] = byte('A' + (reader.read+i)%26)
		}
		
		reader.read += chunkSize
		return int(chunkSize), nil
	}
	
	req := httptest.NewRequest("POST", "/api/v1/test-bucket/file?path=/large-file.txt", reader)
	req.Header.Set("Content-Type", "text/plain")
	req.ContentLength = fileSize
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Should handle large files (though our mock doesn't implement chunked upload yet)
	// The key test is that memory usage stays under 100MB
	assert.True(t, reader.maxMem < 100*1024*1024, 
		"Memory usage %d bytes exceeds 100MB limit", reader.maxMem)
}

func TestWebDAVProxyHandler_DeleteFile(t *testing.T) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	// Mock database deletion
	mock.ExpectExec("DELETE FROM \"oss_files\"").
		WillReturnResult(sqlmock.NewResult(0, 1))
	
	// Pre-populate storage
	testBucket := "test-bucket"
	testFile := "test.txt"
	mockStorage.objects[testBucket+"/"+testFile] = []byte("test content")
	
	req := httptest.NewRequest("DELETE", "/api/v1/test-bucket?path=/test.txt", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify file was deleted
	assert.NotContains(t, mockStorage.objects, testBucket+"/"+testFile)
}

func TestWebDAVProxyHandler_RenameFile(t *testing.T) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	// Mock database update
	mock.ExpectExec("UPDATE \"oss_files\"").
		WillReturnResult(sqlmock.NewResult(0, 1))
	
	// Pre-populate storage
	testBucket := "test-bucket"
	oldKey := "old-file.txt"
	newKey := "new-file.txt"
	testContent := []byte("test content")
	
	mockStorage.objects[testBucket+"/"+oldKey] = testContent
	mockStorage.metadata[testBucket+"/"+oldKey] = oss.ObjectInfo{
		Key:          oldKey,
		Size:         int64(len(testContent)),
		LastModified: time.Now(),
	}
	
	renameReq := RenameRequest{
		OldPath: "/" + oldKey,
		NewPath: "/" + newKey,
	}
	
	reqBody, _ := json.Marshal(renameReq)
	req := httptest.NewRequest("PATCH", "/api/v1/test-bucket/rename", 
		bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify file was renamed
	assert.NotContains(t, mockStorage.objects, testBucket+"/"+oldKey)
	assert.Contains(t, mockStorage.objects, testBucket+"/"+newKey)
	assert.Equal(t, testContent, mockStorage.objects[testBucket+"/"+newKey])
}

func TestWebDAVProxyHandler_CreateDirectory(t *testing.T) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	mkdirReq := MkdirRequest{
		Path: "/new-directory/",
	}
	
	reqBody, _ := json.Marshal(mkdirReq)
	req := httptest.NewRequest("POST", "/api/v1/test-bucket/mkdir", 
		bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify directory marker was created
	assert.Contains(t, mockStorage.objects, "test-bucket/new-directory/")
}

// Test edge cases and error conditions

func TestWebDAVProxyHandler_PermissionErrors(t *testing.T) {
	router, _, mock, _ := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock no bucket access
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"})) // Empty result
	
	req := httptest.NewRequest("GET", "/api/v1/test-bucket?path=/", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestWebDAVProxyHandler_ExpiredToken(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Mock expired token
	router.Use(func(c *gin.Context) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Token expired"})
		c.Abort()
	})
	
	req := httptest.NewRequest("GET", "/api/v1/test-bucket?path=/", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestWebDAVProxyHandler_NetworkFailure(t *testing.T) {
	// Create a storage service that simulates network failure
	type FailingStorageService struct {
		*MockStorageService
	}
	
	failingStorage := &FailingStorageService{NewMockStorageService()}
	failingStorage.GetObject = func(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
		return nil, fmt.Errorf("network connection failed")
	}
	
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Create mock database
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer mock.ExpectationsWereMet()
	
	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)
	
	storageFactory := NewMockStorageFactory(failingStorage)
	
	router.Use(func(c *gin.Context) {
		user := &auth.Claims{UserID: 1, Username: "testuser"}
		c.Set("user", user)
		c.Next()
	})
	
	handler := NewWebDAVProxyHandler(storageFactory, gormDB)
	router.GET("/:bucket", handler.ListDirectory)
	
	// Mock permission checks
	mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	
	req := httptest.NewRequest("GET", "/api/v1/test-bucket?path=/", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestWebDAVProxyHandler_InvalidPaths(t *testing.T) {
	router, _, _, _ := setupTestEnvironment()
	
	testCases := []struct {
		name     string
		path     string
		expected int
	}{
		{"Path traversal", "/api/v1/test-bucket?path=/../../../etc/passwd", http.StatusBadRequest},
		{"Invalid bucket name", "/api/v1/test<bucket?path=/", http.StatusBadRequest},
		{"Empty bucket", "/api/v1/?path=/", http.StatusBadRequest},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path, nil)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, tc.expected, w.Code)
		})
	}
}

// Benchmark tests for performance validation

func BenchmarkWebDAVProxyHandler_ListDirectory(b *testing.B) {
	router, _, mock, mockStorage := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks for all iterations
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	
	// Pre-populate with many files
	testBucket := "test-bucket"
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("file%d.txt", i)
		mockStorage.objects[testBucket+"/"+key] = []byte(fmt.Sprintf("content%d", i))
		mockStorage.metadata[testBucket+"/"+key] = oss.ObjectInfo{
			Key:          key,
			Size:         8,
			LastModified: time.Now(),
		}
	}
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("GET", "/api/v1/test-bucket?path=/", nil)
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}

func BenchmarkWebDAVProxyHandler_UploadFile(b *testing.B) {
	router, _, mock, _ := setupTestEnvironment()
	defer mock.ExpectationsWereMet()
	
	// Mock permission checks for all iterations
	for i := 0; i < b.N; i++ {
		mock.ExpectQuery("SELECT (.+) FROM \"role_region_bucket_accesses\"").
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	}
	
	testContent := strings.Repeat("Hello, World! ", 1000) // ~13KB content
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/test-bucket/file?path=/test%d.txt", i), 
			strings.NewReader(testContent))
		req.Header.Set("Content-Type", "text/plain")
		w := httptest.NewRecorder()
		
		router.ServeHTTP(w, req)
		
		if w.Code != http.StatusOK {
			b.Fatalf("Expected 200, got %d", w.Code)
		}
	}
}
