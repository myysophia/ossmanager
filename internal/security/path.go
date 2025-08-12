package security

import (
	"path/filepath"
	"strings"
)

// CleanPath 清理路径，防止目录穿越
func CleanPath(path string) string {
	// 使用 filepath.Clean 清理路径，它会：
	// 1. 解析 "." 和 ".." 元素
	// 2. 删除多余的斜杠
	// 3. 规范化路径
	cleaned := filepath.Clean(path)
	
	// 确保路径以 "/" 开始（对于绝对路径）
	if strings.HasPrefix(path, "/") && !strings.HasPrefix(cleaned, "/") {
		cleaned = "/" + cleaned
	}
	
	return cleaned
}

// IsPathSafe 检查路径是否安全，防止目录穿越
func IsPathSafe(basePath, targetPath string) bool {
	// 清理路径
	cleanBase := filepath.Clean(basePath)
	cleanTarget := filepath.Clean(targetPath)
	
	// 将相对路径转换为基于 basePath 的绝对路径
	if !filepath.IsAbs(cleanTarget) {
		cleanTarget = filepath.Join(cleanBase, cleanTarget)
	}
	
	// 使用 filepath.Rel 检查目标路径是否在基础路径内
	rel, err := filepath.Rel(cleanBase, cleanTarget)
	if err != nil {
		return false
	}
	
	// 检查相对路径是否包含 ".."，如果包含说明尝试访问父目录
	return !strings.HasPrefix(rel, "..") && !strings.Contains(rel, string(filepath.Separator)+"..")
}

// SanitizePath 净化路径，移除危险元素
func SanitizePath(path string) string {
	// 首先清理路径
	cleaned := CleanPath(path)
	
	// 移除或替换危险字符
	// 这里可以根据需要添加更多的净化规则
	dangerous := []string{
		"../",
		"..\\",
		"./",
		".\\",
	}
	
	for _, danger := range dangerous {
		cleaned = strings.ReplaceAll(cleaned, danger, "")
	}
	
	// 移除多余的斜杠
	for strings.Contains(cleaned, "//") {
		cleaned = strings.ReplaceAll(cleaned, "//", "/")
	}
	
	return cleaned
}

// ValidateWebDAVPath 验证 WebDAV 路径是否安全
func ValidateWebDAVPath(bucketName, filePath string) (string, bool) {
	// 检查存储桶名称是否安全
	if bucketName == "" || strings.Contains(bucketName, "..") || strings.Contains(bucketName, "/") || strings.Contains(bucketName, "\\") {
		return "", false
	}
	
	// 处理空路径
	if filePath == "" {
		return "/", true
	}
	
	// 先检查原始路径是否包含危险模式
	if strings.Contains(filePath, "..") {
		return "", false
	}
	
	// 清理文件路径
	cleanPath := CleanPath(filePath)
	
	// 构建完整路径
	fullPath := "/" + bucketName + cleanPath
	
	// 验证路径是否安全（不能访问存储桶外的内容）
	basePath := "/" + bucketName
	if !IsPathSafe(basePath, fullPath) {
		return "", false
	}
	
	return cleanPath, true
}

// ExtractBucketAndPath 从 WebDAV URL 路径中提取存储桶和文件路径
func ExtractBucketAndPath(urlPath string) (bucket string, filePath string, valid bool) {
	// 检查空路径
	if urlPath == "" || urlPath == "/" {
		return "", "", false
	}
	
	// 先检查原始路径是否包含危险模式
	if strings.Contains(urlPath, "..") {
		return "", "", false
	}
	
	// 清理路径
	cleanPath := CleanPath(urlPath)
	
	// 移除前导斜杠并分割路径
	trimmed := strings.TrimPrefix(cleanPath, "/")
	if trimmed == "" {
		return "", "", false
	}
	
	parts := strings.SplitN(trimmed, "/", 2)
	
	// 检查路径格式
	if len(parts) < 1 {
		return "", "", false
	}
	
	bucket = parts[0]
	if len(parts) > 1 {
		filePath = "/" + parts[1]
	} else {
		filePath = "/"
	}
	
	// 验证存储桶名称
	if bucket == "" || bucket == "." || strings.Contains(bucket, "..") {
		return "", "", false
	}
	
	// 验证文件路径
	if sanitizedPath, safe := ValidateWebDAVPath(bucket, filePath); safe {
		return bucket, sanitizedPath, true
	}
	
	return "", "", false
}
