package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/myysophia/ossmanager/internal/webdav"
	"github.com/myysophia/ossmanager/internal/oss"
)

// PerformanceTest validates memory usage during large file operations
type PerformanceTest struct {
	memoryLimit    int64 // Maximum memory allowed (100MB)
	testFileSizes  []int64
	memoryMonitor  *MemoryMonitor
}

// MemoryMonitor tracks memory usage during operations
type MemoryMonitor struct {
	samples      []MemorySample
	peakMemory   int64
	startMemory  int64
}

type MemorySample struct {
	timestamp time.Time
	heapAlloc int64
	heapSys   int64
	stackSys  int64
}

// NewPerformanceTest creates a new performance test instance
func NewPerformanceTest() *PerformanceTest {
	return &PerformanceTest{
		memoryLimit: 100 * 1024 * 1024, // 100MB
		testFileSizes: []int64{
			50 * 1024 * 1024,   // 50MB
			100 * 1024 * 1024,  // 100MB
			500 * 1024 * 1024,  // 500MB
			1024 * 1024 * 1024, // 1GB
		},
		memoryMonitor: &MemoryMonitor{
			samples: make([]MemorySample, 0),
		},
	}
}

// StartMemoryMonitoring begins tracking memory usage
func (m *MemoryMonitor) StartMemoryMonitoring() {
	var initialStats runtime.MemStats
	runtime.ReadMemStats(&initialStats)
	m.startMemory = int64(initialStats.Alloc)
	
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		
		for range ticker.C {
			var stats runtime.MemStats
			runtime.ReadMemStats(&stats)
			
			sample := MemorySample{
				timestamp: time.Now(),
				heapAlloc: int64(stats.Alloc),
				heapSys:   int64(stats.HeapSys),
				stackSys:  int64(stats.StackSys),
			}
			
			m.samples = append(m.samples, sample)
			
			if sample.heapAlloc > m.peakMemory {
				m.peakMemory = sample.heapAlloc
			}
		}
	}()
}

// StopMemoryMonitoring stops tracking and returns results
func (m *MemoryMonitor) StopMemoryMonitoring() MemoryReport {
	// Allow final memory collection
	time.Sleep(200 * time.Millisecond)
	
	var finalStats runtime.MemStats
	runtime.ReadMemStats(&finalStats)
	
	return MemoryReport{
		StartMemory:    m.startMemory,
		PeakMemory:     m.peakMemory,
		FinalMemory:    int64(finalStats.Alloc),
		MemoryIncrease: m.peakMemory - m.startMemory,
		Samples:        m.samples,
	}
}

type MemoryReport struct {
	StartMemory    int64
	PeakMemory     int64
	FinalMemory    int64
	MemoryIncrease int64
	Samples        []MemorySample
}

// RunStreamingUploadTest tests memory usage during streaming uploads
func (pt *PerformanceTest) RunStreamingUploadTest() error {
	fmt.Println("=== Streaming Upload Performance Test ===")
	
	for i, fileSize := range pt.testFileSizes {
		fmt.Printf("\nTest %d: Testing %s file upload\n", i+1, formatBytes(fileSize))
		
		// Create test data generator (memory efficient)
		dataReader := &PatternDataReader{
			size:      fileSize,
			remaining: fileSize,
			pattern:   []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"),
		}
		
		// Start memory monitoring
		pt.memoryMonitor = &MemoryMonitor{samples: make([]MemorySample, 0)}
		pt.memoryMonitor.StartMemoryMonitoring()
		
		// Create streaming file system (mock)
		fs := createMockFileSystem()
		
		// Perform streaming upload
		startTime := time.Now()
		file := webdav.NewStreamingOSSFile(fs, fmt.Sprintf("test-file-%d.bin", i), true)
		
		// Stream data in chunks
		buffer := make([]byte, 64*1024) // 64KB buffer
		for dataReader.remaining > 0 {
			n, err := dataReader.Read(buffer)
			if err != nil && err != io.EOF {
				return fmt.Errorf("error reading test data: %w", err)
			}
			
			if n > 0 {
				_, err = file.Write(buffer[:n])
				if err != nil {
					return fmt.Errorf("error writing to streaming file: %w", err)
				}
			}
			
			if err == io.EOF {
				break
			}
			
			// Force garbage collection periodically
			if dataReader.size-dataReader.remaining > 0 && (dataReader.size-dataReader.remaining)%(100*1024*1024) == 0 {
				runtime.GC()
			}
		}
		
		// Finalize upload
		err := file.Close()
		if err != nil {
			return fmt.Errorf("error closing streaming file: %w", err)
		}
		
		duration := time.Since(startTime)
		
		// Stop memory monitoring and get report
		report := pt.memoryMonitor.StopMemoryMonitoring()
		
		// Validate memory usage
		success := pt.validateMemoryUsage(report, fileSize)
		
		// Print results
		pt.printTestResults(fileSize, duration, report, success)
		
		if !success {
			return fmt.Errorf("memory usage exceeded limit for %s file", formatBytes(fileSize))
		}
		
		// Cleanup before next test
		runtime.GC()
		time.Sleep(1 * time.Second)
	}
	
	fmt.Println("\n=== All Streaming Upload Tests Passed ===")
	return nil
}

// RunConcurrentUploadTest tests memory usage during concurrent uploads
func (pt *PerformanceTest) RunConcurrentUploadTest() error {
	fmt.Println("\n=== Concurrent Upload Performance Test ===")
	
	const numConcurrentUploads = 5
	const fileSize = 100 * 1024 * 1024 // 100MB each
	
	fmt.Printf("Testing %d concurrent uploads of %s each\n", numConcurrentUploads, formatBytes(fileSize))
	
	// Start memory monitoring
	pt.memoryMonitor = &MemoryMonitor{samples: make([]MemorySample, 0)}
	pt.memoryMonitor.StartMemoryMonitoring()
	
	startTime := time.Now()
	
	// Channel to collect results
	results := make(chan error, numConcurrentUploads)
	
	// Start concurrent uploads
	for i := 0; i < numConcurrentUploads; i++ {
		go func(uploadIndex int) {
			dataReader := &PatternDataReader{
				size:      fileSize,
				remaining: fileSize,
				pattern:   []byte(fmt.Sprintf("UPLOAD_%d_PATTERN_", uploadIndex)),
			}
			
			fs := createMockFileSystem()
			file := webdav.NewStreamingOSSFile(fs, fmt.Sprintf("concurrent-file-%d.bin", uploadIndex), true)
			
			buffer := make([]byte, 32*1024) // 32KB buffer per upload
			for dataReader.remaining > 0 {
				n, err := dataReader.Read(buffer)
				if err != nil && err != io.EOF {
					results <- fmt.Errorf("upload %d read error: %w", uploadIndex, err)
					return
				}
				
				if n > 0 {
					_, err = file.Write(buffer[:n])
					if err != nil {
						results <- fmt.Errorf("upload %d write error: %w", uploadIndex, err)
						return
					}
				}
				
				if err == io.EOF {
					break
				}
			}
			
			err := file.Close()
			if err != nil {
				results <- fmt.Errorf("upload %d close error: %w", uploadIndex, err)
				return
			}
			
			results <- nil // Success
		}(i)
	}
	
	// Wait for all uploads to complete
	var errors []error
	for i := 0; i < numConcurrentUploads; i++ {
		if err := <-results; err != nil {
			errors = append(errors, err)
		}
	}
	
	duration := time.Since(startTime)
	
	// Stop memory monitoring
	report := pt.memoryMonitor.StopMemoryMonitoring()
	
	// Validate memory usage (should be much less than total file size)
	totalDataSize := int64(numConcurrentUploads) * fileSize
	maxAllowedMemory := pt.memoryLimit * 2 // Allow double limit for concurrent operations
	success := report.MemoryIncrease < maxAllowedMemory
	
	// Print results
	fmt.Printf("\nConcurrent Upload Results:\n")
	fmt.Printf("  Total data processed: %s\n", formatBytes(totalDataSize))
	fmt.Printf("  Duration: %v\n", duration)
	fmt.Printf("  Throughput: %s/s\n", formatBytes(totalDataSize/int64(duration.Seconds())))
	fmt.Printf("  Memory increase: %s (limit: %s)\n", formatBytes(report.MemoryIncrease), formatBytes(maxAllowedMemory))
	fmt.Printf("  Peak memory: %s\n", formatBytes(report.PeakMemory))
	fmt.Printf("  Success: %t\n", success && len(errors) == 0)
	
	if len(errors) > 0 {
		fmt.Printf("  Errors: %d\n", len(errors))
		for _, err := range errors {
			fmt.Printf("    - %v\n", err)
		}
		return fmt.Errorf("concurrent upload test failed with %d errors", len(errors))
	}
	
	if !success {
		return fmt.Errorf("concurrent upload memory usage exceeded limit")
	}
	
	fmt.Println("=== Concurrent Upload Test Passed ===")
	return nil
}

// validateMemoryUsage checks if memory usage is within limits
func (pt *PerformanceTest) validateMemoryUsage(report MemoryReport, fileSize int64) bool {
	// Memory increase should be much less than file size
	// and within our absolute limit
	return report.MemoryIncrease < pt.memoryLimit && 
		   report.MemoryIncrease < fileSize/4 // Memory should be less than 25% of file size
}

// printTestResults displays test results
func (pt *PerformanceTest) printTestResults(fileSize int64, duration time.Duration, report MemoryReport, success bool) {
	fmt.Printf("Results:\n")
	fmt.Printf("  Duration: %v\n", duration)
	fmt.Printf("  Throughput: %s/s\n", formatBytes(fileSize/int64(duration.Seconds())))
	fmt.Printf("  Memory increase: %s\n", formatBytes(report.MemoryIncrease))
	fmt.Printf("  Peak memory: %s\n", formatBytes(report.PeakMemory))
	fmt.Printf("  Memory efficiency: %.1f%% (lower is better)\n", float64(report.MemoryIncrease)/float64(fileSize)*100)
	fmt.Printf("  Success: %t\n", success)
	
	if success {
		fmt.Printf("  ✅ Memory usage within limits\n")
	} else {
		fmt.Printf("  ❌ Memory usage exceeded limits\n")
	}
}

// PatternDataReader generates test data on-demand without storing it in memory
type PatternDataReader struct {
	size      int64
	remaining int64
	pattern   []byte
	position  int64
}

func (r *PatternDataReader) Read(p []byte) (n int, err error) {
	if r.remaining <= 0 {
		return 0, io.EOF
	}
	
	// Fill buffer with pattern data
	bytesToWrite := len(p)
	if int64(bytesToWrite) > r.remaining {
		bytesToWrite = int(r.remaining)
	}
	
	for i := 0; i < bytesToWrite; i++ {
		p[i] = r.pattern[(r.position+int64(i))%int64(len(r.pattern))]
	}
	
	r.position += int64(bytesToWrite)
	r.remaining -= int64(bytesToWrite)
	
	return bytesToWrite, nil
}

// createMockFileSystem creates a mock file system for testing
func createMockFileSystem() *webdav.OSSFileSystem {
	// Create a minimal mock storage service for testing
	storage := &MockStorage{}
	return &webdav.OSSFileSystem{
		// Initialize with minimal required fields for testing
	}
}

// MockStorage provides a minimal storage interface for testing
type MockStorage struct{}

func (s *MockStorage) PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error {
	// Simulate storage operation by consuming the reader
	buffer := make([]byte, 32*1024)
	for {
		_, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *MockStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader([]byte{})), nil
}

func (s *MockStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	return nil
}

func (s *MockStorage) ListObjects(ctx context.Context, bucket, prefix string, maxKeys int) ([]oss.ObjectInfo, error) {
	return []oss.ObjectInfo{}, nil
}

func (s *MockStorage) CopyObject(ctx context.Context, srcBucket, srcKey, destBucket, destKey string) error {
	return nil
}

func (s *MockStorage) GetType() string {
	return "mock"
}

// formatBytes formats byte count as human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	
	units := []string{"B", "KB", "MB", "GB", "TB"}
	return fmt.Sprintf("%.1f %s", float64(bytes)/float64(div), units[exp])
}

// RunMemoryBenchmark runs comprehensive memory benchmarks
func (pt *PerformanceTest) RunMemoryBenchmark() error {
	fmt.Println("Starting OSS Manager Memory Performance Validation")
	fmt.Printf("Memory limit: %s\n", formatBytes(pt.memoryLimit))
	fmt.Println("Testing streaming upload with memory constraints...")
	
	// Test streaming uploads
	if err := pt.RunStreamingUploadTest(); err != nil {
		return fmt.Errorf("streaming upload test failed: %w", err)
	}
	
	// Test concurrent uploads
	if err := pt.RunConcurrentUploadTest(); err != nil {
		return fmt.Errorf("concurrent upload test failed: %w", err)
	}
	
	return nil
}

func main() {
	test := NewPerformanceTest()
	
	if err := test.RunMemoryBenchmark(); err != nil {
		fmt.Printf("Performance test failed: %v\n", err)
		return
	}
	
	fmt.Println("\n🎉 All performance tests passed!")
	fmt.Printf("✅ 1GB file upload uses less than %s RAM\n", formatBytes(test.memoryLimit))
	fmt.Println("✅ Memory-efficient streaming implementation verified")
	fmt.Println("✅ Concurrent upload memory usage within limits")
}
