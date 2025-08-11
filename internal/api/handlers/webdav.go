//go:build webdav
// +build webdav

package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager-backend/internal/auth"
	"github.com/myysophia/ossmanager-backend/internal/oss"
	webdavfs "github.com/myysophia/ossmanager-backend/internal/webdav"
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
	// 从路径中提取存储桶信息
	path := c.Request.URL.Path
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// 路径格式: /webdav/{bucket}/...
	if len(pathParts) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid WebDAV path"})
		return
	}

	bucket := pathParts[1]

	// 获取当前用户
	userValue, exists := c.Get("user")
	if !exists {
		c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	user := userValue.(*auth.Claims)

	// 检查用户对该存储桶的访问权限
	if !h.checkBucketAccess(user.UserID, bucket) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to bucket"})
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
			if err != nil {
				// 可以在这里添加日志记录
			}
		},
	}

	// 修改请求路径，移除存储桶前缀
	originalPath := c.Request.URL.Path
	if len(pathParts) > 2 {
		c.Request.URL.Path = "/" + strings.Join(pathParts[2:], "/")
	} else {
		c.Request.URL.Path = "/"
	}

	// 处理 WebDAV 请求
	webdavHandler.ServeHTTP(c.Writer, c.Request)

	// 恢复原始路径
	c.Request.URL.Path = originalPath
}

// checkBucketAccess 检查用户是否有权限访问该存储桶
func (h *WebDAVHandler) checkBucketAccess(userID uint, bucket string) bool {
	// 检查用户是否有权限访问该存储桶
	var count int64
	h.db.Table("role_bucket_access rba").
		Joins("JOIN user_roles ur ON ur.role_id = rba.role_id").
		Where("ur.user_id = ? AND rba.bucket = ?", userID, bucket).
		Count(&count)

	return count > 0
}
