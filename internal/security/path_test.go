package security

import (
	"testing"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/", "/"},
		{"", "."},
		{"/test", "/test"},
		{"/test/", "/test"},
		{"/test/../", "/"},
		{"/test/./file", "/test/file"},
		{"/../../../etc/passwd", "/etc/passwd"},
		{"/test//file", "/test/file"},
		{"test/../../file", "../file"},
	}

	for _, test := range tests {
		result := CleanPath(test.input)
		if result != test.expected {
			t.Errorf("CleanPath(%s) = %s, want %s", test.input, result, test.expected)
		}
	}
}

func TestIsPathSafe(t *testing.T) {
	tests := []struct {
		basePath   string
		targetPath string
		expected   bool
	}{
		{"/bucket", "/bucket/file", true},
		{"/bucket", "/bucket/subdir/file", true},
		{"/bucket", "/bucket/../etc/passwd", false},
		{"/bucket", "../etc/passwd", false},
		{"/bucket", "/other/file", false},
		{"/bucket", "/bucket", true},
		{"/bucket", "file", true}, // 相对路径在基础路径内
	}

	for _, test := range tests {
		result := IsPathSafe(test.basePath, test.targetPath)
		if result != test.expected {
			t.Errorf("IsPathSafe(%s, %s) = %t, want %t", test.basePath, test.targetPath, result, test.expected)
		}
	}
}

func TestValidateWebDAVPath(t *testing.T) {
	tests := []struct {
		bucket   string
		filePath string
		expected bool
	}{
		{"test-bucket", "/file.txt", true},
		{"test-bucket", "/subdir/file.txt", true},
		{"test../bucket", "/file.txt", false},
		{"test-bucket", "/../../../etc/passwd", false},
		{"test-bucket", "/normal/path", true},
		{"test-bucket", "", true},
	}

	for _, test := range tests {
		_, safe := ValidateWebDAVPath(test.bucket, test.filePath)
		if safe != test.expected {
			t.Errorf("ValidateWebDAVPath(%s, %s) = %t, want %t", test.bucket, test.filePath, safe, test.expected)
		}
	}
}

func TestExtractBucketAndPath(t *testing.T) {
	tests := []struct {
		urlPath      string
		expectedBucket string
		expectedPath   string
		expectedValid  bool
	}{
		{"/bucket/file.txt", "bucket", "/file.txt", true},
		{"/bucket", "bucket", "/", true},
		{"/bucket/", "bucket", "/", true},
		{"/bucket/subdir/file.txt", "bucket", "/subdir/file.txt", true},
		{"/bucket/../../../etc/passwd", "", "", false},
		{"", "", "", false},
		{"/", "", "", false},
	}

	for _, test := range tests {
		bucket, path, valid := ExtractBucketAndPath(test.urlPath)
		if bucket != test.expectedBucket || path != test.expectedPath || valid != test.expectedValid {
			t.Errorf("ExtractBucketAndPath(%s) = (%s, %s, %t), want (%s, %s, %t)", 
				test.urlPath, bucket, path, valid, 
				test.expectedBucket, test.expectedPath, test.expectedValid)
		}
	}
}
