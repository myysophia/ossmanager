package handlers

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/oss"
	"github.com/myysophia/ossmanager/internal/security"
	webdavfs "github.com/myysophia/ossmanager/internal/webdav"
	"go.uber.org/zap"
)

// WebDAVProxyHandler handles RESTful WebDAV proxy requests
type WebDAVProxyHandler struct {
	*BaseHandler
	storageFactory oss.StorageFactory
	db             *gorm.DB
}

// NewWebDAVProxyHandler creates a new WebDAV proxy handler
func NewWebDAVProxyHandler(storageFactory oss.StorageFactory, db *gorm.DB) *WebDAVProxyHandler {
	return &WebDAVProxyHandler{
		BaseHandler:    NewBaseHandler(),
		storageFactory: storageFactory,
		db:             db,
	}
}

// FileItem represents a file or directory item in JSON response
type FileItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"isDir"`
	Size     int64     `json:"size"`
	MTime    time.Time `json:"mtime"`
	MimeType string    `json:"mimeType,omitempty"`
}

// ListDirectoryResponse represents the response for directory listing
type ListDirectoryResponse struct {
	Items  []FileItem `json:"items"`
	Path   string     `json:"path"`
	Total  int        `json:"total"`
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
	HasMore bool      `json:"hasMore"`
}

// OperationResponse represents a generic operation response
type OperationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Path    string `json:"path,omitempty"`
}

// RenameRequest represents the request body for rename operation
type RenameRequest struct {
	OldPath string `json:"oldPath" binding:"required"`
	NewPath string `json:"newPath" binding:"required"`
}

// MkdirRequest represents the request body for mkdir operation
type MkdirRequest struct {
	Path string `json:"path" binding:"required"`
}

// ListDirectory handles GET /{bucket}?path=/dir → list directory (maps to PROPFIND)
func (h *WebDAVProxyHandler) ListDirectory(c *gin.Context) {
	bucket := c.Param("bucket")
	dirPath := c.Query("path")
	
	// Get pagination parameters
	offset := 0
	limit := 100 // Default limit
	
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 1000 { // Max limit of 1000 items
				limit = 1000
			}
		}
	}
	
	if bucket == "" {
		h.BadRequest(c, "bucket parameter is required")
		return
	}

	// Clean and validate bucket name
	if bucket == "" || strings.Contains(bucket, "..") || strings.ContainsAny(bucket, "/\\<>:\"|?*") {
		h.BadRequest(c, "invalid bucket name")
		return
	}

	if dirPath == "" {
		dirPath = "/"
	}
	
	// Validate directory path using WebDAV path validation
	if _, valid := security.ValidateWebDAVPath(bucket, dirPath); !valid {
		h.BadRequest(c, "invalid directory path")
		return
	}

	// Get current user
	user, exists := h.getCurrentUser(c)
	if !exists {
		h.Unauthorized(c, "authentication required")
		return
	}

	// Check bucket access
	if !h.checkBucketAccess(user.UserID, bucket) {
		h.Forbidden(c, "access denied to bucket")
		return
	}

	// Get WebDAV filesystem
	fs, err := h.getWebDAVFileSystem(user.UserID, bucket)
	if err != nil {
		logger.Error("Failed to get WebDAV filesystem", zap.Error(err))
		h.InternalError(c, "storage service unavailable")
		return
	}

	// Normalize directory path
	cleanPath := strings.TrimPrefix(dirPath, "/")
	if cleanPath != "" && !strings.HasSuffix(cleanPath, "/") {
		cleanPath += "/"
	}

	// Create a fake WebDAV request for PROPFIND
	ctx := context.Background()
	file, err := fs.OpenFile(ctx, cleanPath, 0, 0)
	if err != nil {
		logger.Error("Failed to open directory", zap.String("path", cleanPath), zap.Error(err))
		h.NotFound(c, "directory not found")
		return
	}
	defer file.Close()

	// Read directory contents
	fileInfos, err := file.Readdir(-1)
	if err != nil {
		logger.Error("Failed to read directory", zap.String("path", cleanPath), zap.Error(err))
		h.InternalError(c, "failed to read directory")
		return
	}

	// Apply pagination
	totalItems := len(fileInfos)
	hasMore := false
	
	// Calculate pagination bounds
	start := offset
	if start > totalItems {
		start = totalItems
	}
	
	end := start + limit
	if end > totalItems {
		end = totalItems
	} else {
		hasMore = end < totalItems
	}
	
	// Apply pagination slice
	paginatedFileInfos := fileInfos[start:end]

	// Convert to JSON response
	items := make([]FileItem, 0, len(paginatedFileInfos))
	for _, fi := range paginatedFileInfos {
		itemPath := path.Join(dirPath, fi.Name())
		if fi.IsDir() && !strings.HasSuffix(itemPath, "/") {
			itemPath += "/"
		}
		
		item := FileItem{
			Name:  fi.Name(),
			Path:  itemPath,
			IsDir: fi.IsDir(),
			Size:  fi.Size(),
			MTime: fi.ModTime(),
		}
		
		if !fi.IsDir() {
			item.MimeType = h.getMimeType(fi.Name())
		}
		
		items = append(items, item)
	}

	response := ListDirectoryResponse{
		Items:   items,
		Path:    dirPath,
		Total:   totalItems,
		Offset:  offset,
		Limit:   limit,
		HasMore: hasMore,
	}

	h.Success(c, response)
}

// UploadFile handles POST /{bucket}/file → upload file (maps to PUT)
func (h *WebDAVProxyHandler) UploadFile(c *gin.Context) {
	bucket := c.Param("bucket")
	filePath := c.Query("path")
	
	if bucket == "" {
		h.BadRequest(c, "bucket parameter is required")
		return
	}
	
	if filePath == "" {
		h.BadRequest(c, "path parameter is required")
		return
	}

	// Clean and validate bucket name
	if bucket == "" || strings.Contains(bucket, "..") || strings.ContainsAny(bucket, "/\\<>:\"|?*") {
		h.BadRequest(c, "invalid bucket name")
		return
	}
	
	// Validate file path using WebDAV path validation
	if _, valid := security.ValidateWebDAVPath(bucket, filePath); !valid {
		h.BadRequest(c, "invalid file path")
		return
	}

	// Get current user
	user, exists := h.getCurrentUser(c)
	if !exists {
		h.Unauthorized(c, "authentication required")
		return
	}

	// Check bucket access and write permissions
	if !h.checkBucketAccess(user.UserID, bucket) {
		h.Forbidden(c, "access denied to bucket")
		return
	}
	
	if !h.checkResourceAccess(user.UserID, bucket, filePath, "PUT") {
		h.Forbidden(c, "write access denied")
		return
	}

	// Get WebDAV filesystem
	fs, err := h.getWebDAVFileSystem(user.UserID, bucket)
	if err != nil {
		logger.Error("Failed to get WebDAV filesystem", zap.Error(err))
		h.InternalError(c, "storage service unavailable")
		return
	}

	// Clean file path
	cleanPath := strings.TrimPrefix(filePath, "/")

	// Create file for writing
	ctx := context.Background()
	file, err := fs.OpenFile(ctx, cleanPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		logger.Error("Failed to create file", zap.String("path", cleanPath), zap.Error(err))
		h.InternalError(c, "failed to create file")
		return
	}
	defer file.Close()

	// Copy request body to file
	_, err = io.Copy(file, c.Request.Body)
	if err != nil {
		logger.Error("Failed to write file", zap.String("path", cleanPath), zap.Error(err))
		h.InternalError(c, "failed to write file")
		return
	}

	response := OperationResponse{
		Success: true,
		Message: "file uploaded successfully",
		Path:    filePath,
	}

	h.Success(c, response)
}

// DeleteFile handles DELETE /{bucket}?path=/file → delete file/folder (DELETE)
func (h *WebDAVProxyHandler) DeleteFile(c *gin.Context) {
	bucket := c.Param("bucket")
	targetPath := c.Query("path")
	
	if bucket == "" {
		h.BadRequest(c, "bucket parameter is required")
		return
	}
	
	if targetPath == "" {
		h.BadRequest(c, "path parameter is required")
		return
	}

	// Clean and validate bucket name
	if bucket == "" || strings.Contains(bucket, "..") || strings.ContainsAny(bucket, "/\\<>:\"|?*") {
		h.BadRequest(c, "invalid bucket name")
		return
	}
	
	// Validate target path using WebDAV path validation
	if _, valid := security.ValidateWebDAVPath(bucket, targetPath); !valid {
		h.BadRequest(c, "invalid target path")
		return
	}

	// Get current user
	user, exists := h.getCurrentUser(c)
	if !exists {
		h.Unauthorized(c, "authentication required")
		return
	}

	// Check bucket access and delete permissions
	if !h.checkBucketAccess(user.UserID, bucket) {
		h.Forbidden(c, "access denied to bucket")
		return
	}
	
	if !h.checkResourceAccess(user.UserID, bucket, targetPath, "DELETE") {
		h.Forbidden(c, "delete access denied")
		return
	}

	// Get WebDAV filesystem
	fs, err := h.getWebDAVFileSystem(user.UserID, bucket)
	if err != nil {
		logger.Error("Failed to get WebDAV filesystem", zap.Error(err))
		h.InternalError(c, "storage service unavailable")
		return
	}

	// Clean target path
	cleanPath := strings.TrimPrefix(targetPath, "/")

	// Delete the file or directory
	ctx := context.Background()
	err = fs.RemoveAll(ctx, cleanPath)
	if err != nil {
		logger.Error("Failed to delete", zap.String("path", cleanPath), zap.Error(err))
		h.InternalError(c, "failed to delete file or directory")
		return
	}

	response := OperationResponse{
		Success: true,
		Message: "file or directory deleted successfully",
		Path:    targetPath,
	}

	h.Success(c, response)
}

// RenameFile handles PATCH /{bucket}/rename → rename/move (MOVE)
func (h *WebDAVProxyHandler) RenameFile(c *gin.Context) {
	bucket := c.Param("bucket")
	
	if bucket == "" {
		h.BadRequest(c, "bucket parameter is required")
		return
	}

	// Clean and validate bucket name
	if bucket == "" || strings.Contains(bucket, "..") || strings.ContainsAny(bucket, "/\\<>:\"|?*") {
		h.BadRequest(c, "invalid bucket name")
		return
	}

	var req RenameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Validate paths using WebDAV path validation
	if _, valid := security.ValidateWebDAVPath(bucket, req.OldPath); !valid {
		h.BadRequest(c, "invalid old path")
		return
	}
	if _, valid := security.ValidateWebDAVPath(bucket, req.NewPath); !valid {
		h.BadRequest(c, "invalid new path")
		return
	}

	// Get current user
	user, exists := h.getCurrentUser(c)
	if !exists {
		h.Unauthorized(c, "authentication required")
		return
	}

	// Check bucket access and move permissions
	if !h.checkBucketAccess(user.UserID, bucket) {
		h.Forbidden(c, "access denied to bucket")
		return
	}
	
	if !h.checkResourceAccess(user.UserID, bucket, req.OldPath, "MOVE") {
		h.Forbidden(c, "move access denied")
		return
	}

	// Get WebDAV filesystem
	fs, err := h.getWebDAVFileSystem(user.UserID, bucket)
	if err != nil {
		logger.Error("Failed to get WebDAV filesystem", zap.Error(err))
		h.InternalError(c, "storage service unavailable")
		return
	}

	// Clean paths
	oldPath := strings.TrimPrefix(req.OldPath, "/")
	newPath := strings.TrimPrefix(req.NewPath, "/")

	// Rename/move the file or directory
	ctx := context.Background()
	err = fs.Rename(ctx, oldPath, newPath)
	if err != nil {
		logger.Error("Failed to rename", zap.String("oldPath", oldPath), zap.String("newPath", newPath), zap.Error(err))
		h.InternalError(c, "failed to rename file or directory")
		return
	}

	response := OperationResponse{
		Success: true,
		Message: fmt.Sprintf("renamed from %s to %s successfully", req.OldPath, req.NewPath),
		Path:    req.NewPath,
	}

	h.Success(c, response)
}

// CreateDirectory handles POST /{bucket}/mkdir → create folder (MKCOL)
func (h *WebDAVProxyHandler) CreateDirectory(c *gin.Context) {
	bucket := c.Param("bucket")
	
	if bucket == "" {
		h.BadRequest(c, "bucket parameter is required")
		return
	}

	// Clean and validate bucket name
	if bucket == "" || strings.Contains(bucket, "..") || strings.ContainsAny(bucket, "/\\<>:\"|?*") {
		h.BadRequest(c, "invalid bucket name")
		return
	}

	var req MkdirRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "invalid request body: "+err.Error())
		return
	}

	// Validate directory path using WebDAV path validation
	if _, valid := security.ValidateWebDAVPath(bucket, req.Path); !valid {
		h.BadRequest(c, "invalid directory path")
		return
	}

	// Get current user
	user, exists := h.getCurrentUser(c)
	if !exists {
		h.Unauthorized(c, "authentication required")
		return
	}

	// Check bucket access and create permissions
	if !h.checkBucketAccess(user.UserID, bucket) {
		h.Forbidden(c, "access denied to bucket")
		return
	}
	
	if !h.checkResourceAccess(user.UserID, bucket, req.Path, "MKCOL") {
		h.Forbidden(c, "create access denied")
		return
	}

	// Get WebDAV filesystem
	fs, err := h.getWebDAVFileSystem(user.UserID, bucket)
	if err != nil {
		logger.Error("Failed to get WebDAV filesystem", zap.Error(err))
		h.InternalError(c, "storage service unavailable")
		return
	}

	// Clean directory path
	cleanPath := strings.TrimPrefix(req.Path, "/")

	// Create directory
	ctx := context.Background()
	err = fs.Mkdir(ctx, cleanPath, 0755)
	if err != nil {
		logger.Error("Failed to create directory", zap.String("path", cleanPath), zap.Error(err))
		h.InternalError(c, "failed to create directory")
		return
	}

	response := OperationResponse{
		Success: true,
		Message: "directory created successfully",
		Path:    req.Path,
	}

	h.Success(c, response)
}

// Helper methods

// getCurrentUser extracts user information from context
func (h *WebDAVProxyHandler) getCurrentUser(c *gin.Context) (*auth.Claims, bool) {
	userValue, exists := c.Get("user")
	if !exists {
		return nil, false
	}
	
	user, ok := userValue.(*auth.Claims)
	return user, ok
}

// getWebDAVFileSystem creates a WebDAV filesystem instance
func (h *WebDAVProxyHandler) getWebDAVFileSystem(userID uint, bucket string) (*webdavfs.OSSFileSystem, error) {
	storage, err := h.storageFactory.GetDefaultStorageService()
	if err != nil {
		return nil, err
	}
	
	return webdavfs.NewOSSFileSystem(storage, h.db, userID, bucket), nil
}

// checkBucketAccess checks if user has access to the bucket
func (h *WebDAVProxyHandler) checkBucketAccess(userID uint, bucket string) bool {
	return auth.CheckBucketAccess(h.db, userID, "", bucket)
}

// checkResourceAccess checks if user has specific access to a resource
func (h *WebDAVProxyHandler) checkResourceAccess(userID uint, bucket, resourcePath, method string) bool {
	var resource, action string
	
	switch method {
	case "GET", "HEAD", "OPTIONS", "PROPFIND":
		resource = "webdav"
		action = "read"
	case "PUT", "POST", "MKCOL", "COPY", "MOVE":
		resource = "webdav"
		action = "write"
	case "DELETE":
		resource = "webdav"
		action = "delete"
	default:
		return false
	}
	
	err := auth.CheckPermission(userID, resource, action)
	if err != nil {
		logger.Warn("WebDAV resource access denied",
			zap.Uint("userID", userID),
			zap.String("bucket", bucket),
			zap.String("resource", resourcePath),
			zap.String("method", method),
			zap.Error(err))
		return false
	}
	
	return true
}

// getMimeType returns the MIME type for a file based on its extension
func (h *WebDAVProxyHandler) getMimeType(filename string) string {
	ext := strings.ToLower(path.Ext(filename))
	
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".svg":  "image/svg+xml",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".zip":  "application/zip",
		".tar":  "application/x-tar",
		".gz":   "application/gzip",
	}
	
	if mimeType, exists := mimeTypes[ext]; exists {
		return mimeType
	}
	
	return "application/octet-stream"
}
