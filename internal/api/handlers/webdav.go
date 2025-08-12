//go:build webdav
// +build webdav

package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/oss"
	"github.com/myysophia/ossmanager/internal/security"
	webdavfs "github.com/myysophia/ossmanager/internal/webdav"
	"go.uber.org/zap"
)

type WebDAVHandler struct {
	storageFactory oss.StorageFactory
	db             *gorm.DB
}

func NewWebDAVHandler(storageFactory oss.StorageFactory, db *gorm.DB) *WebDAVHandler {
	return &WebDAVHandler{
		storageFactory: storageFactory,
		db:             db,
	}
}

func (h *WebDAVHandler) ServeHTTP(c *gin.Context) {
	// 从路径中提取存储桶信息，使用安全的路径处理
	path := c.Request.URL.Path
	
	// 移除 /webdav 前缀
	webdavPath := strings.TrimPrefix(path, "/webdav")
	
	// 使用安全的路径提取
	bucket, filePath, valid := security.ExtractBucketAndPath(webdavPath)
	if !valid || bucket == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or unsafe WebDAV path"})
		return
	}

	// 获取当前用户
	userValue, exists := c.Get("user")
	if !exists {
		c.Header("WWW-Authenticate", `Bearer realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user := userValue.(*auth.Claims)

	// 检查用户对该存储桶的访问权限
	// 如果使用 WebDAV Token，还要检查令牌是否允许访问该存储桶
	if webdavToken, exists := c.Get("webdav_token"); exists {
		token := webdavToken.(*models.WebDAVToken)
		if token.Bucket != bucket {
			c.JSON(http.StatusForbidden, gin.H{"error": "Token does not allow access to this bucket"})
			return
		}
	} else {
		// 使用常规权限检查
		if !h.checkBucketAccess(user.UserID, bucket) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to bucket"})
			return
		}
	}
	
	// 检查资源级别权限（针对具体文件操作）
	if !h.checkResourceAccess(user.UserID, bucket, filePath, c.Request.Method) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to resource"})
		return
	}

	// 获取存储服务
	storage, err := h.storageFactory.GetDefaultStorageService()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Storage service unavailable"})
		return
	}

	// 创建 WebDAV 文件系统
	fs := webdavfs.NewOSSFileSystem(storage, h.db, user.UserID, bucket)

	// 创建 WebDAV Handler
	webdavHandler := &webdav.Handler{
		Prefix:     "/webdav/" + bucket,
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			// 记录所有 WebDAV 操作到审计日志
			h.logWebDAVOperation(user.UserID, user.Username, r.Method, bucket, filePath, r.RemoteAddr, r.UserAgent(), err)
		},
	}

	// 修改请求路径，使用安全的文件路径
	originalPath := c.Request.URL.Path
	c.Request.URL.Path = filePath

	// 处理 WebDAV 请求
	webdavHandler.ServeHTTP(c.Writer, c.Request)

	// 恢复原始路径
	c.Request.URL.Path = originalPath
}

// checkBucketAccess 检查用户是否有权限访问该存储桶，复用角色-桶映射表
func (h *WebDAVHandler) checkBucketAccess(userID uint, bucket string) bool {
	// 使用现有的 RBAC 函数检查桶访问权限
	return auth.CheckBucketAccess(h.db, userID, "", bucket)
}

// checkResourceAccess 检查用户对特定资源的访问权限（资源级别校验预留）
func (h *WebDAVHandler) checkResourceAccess(userID uint, bucket, resourcePath, method string) bool {
	// 这里预留资源级别的权限校验
	// 可以根据具体需求实现：
	// 1. 检查用户对特定文件/目录的权限
	// 2. 根据操作类型（GET/PUT/DELETE等）进行不同的权限检查
	// 3. 支持基于文件路径的细粒度权限控制
	
	// 目前实现基本的操作权限检查
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
	case "LOCK", "UNLOCK", "PROPPATCH":
		resource = "webdav"
		action = "manage"
	default:
		// 未知方法，拒绝访问
		return false
	}
	
	// 检查用户权限
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

// logWebDAVOperation 记录 WebDAV 操作到审计日志
func (h *WebDAVHandler) logWebDAVOperation(userID uint, username, method, bucket, resourcePath, remoteAddr, userAgent string, err error) {
	// 构建审计日志
	status := "SUCCESS"
	if err != nil {
		status = "FAILED"
	}
	
	// 构建详细信息
	details := fmt.Sprintf(`{"method":"%s","bucket":"%s","resource":"%s","error":"%v"}`,
		method, bucket, resourcePath, err)
	
	// 创建审计日志记录
	auditLog := &models.AuditLog{
		UserID:       userID,
		Username:     username,
		Action:       method,
		ResourceType: "webdav",
		ResourceID:   bucket + ":" + resourcePath,
		Details:      details,
		IPAddress:    strings.Split(remoteAddr, ":")[0], // 提取IP地址
		UserAgent:    userAgent,
		Status:       status,
	}
	
	// 异步保存审计日志
	go func() {
		if saveErr := h.db.Create(auditLog).Error; saveErr != nil {
			logger.Error("Failed to save WebDAV audit log",
				zap.Error(saveErr),
				zap.Uint("userID", userID),
				zap.String("method", method),
				zap.String("bucket", bucket))
		}
	}()
	
	// 同时记录到应用日志
	if err != nil {
		logger.Error("WebDAV operation failed",
			zap.Uint("userID", userID),
			zap.String("username", username),
			zap.String("method", method),
			zap.String("bucket", bucket),
			zap.String("resource", resourcePath),
			zap.String("remoteAddr", remoteAddr),
			zap.Error(err))
	} else {
		logger.Info("WebDAV operation success",
			zap.Uint("userID", userID),
			zap.String("username", username),
			zap.String("method", method),
			zap.String("bucket", bucket),
			zap.String("resource", resourcePath),
			zap.String("remoteAddr", remoteAddr))
	}
}
