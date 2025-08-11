//go:build !webdav
// +build !webdav

package handlers

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager-backend/internal/oss"
)

// WebDAVHandler stub implementation when WebDAV is disabled
type WebDAVHandler struct {
	storageFactory oss.StorageFactory
	db             *gorm.DB
}

// NewWebDAVHandler creates a stub WebDAV handler
func NewWebDAVHandler(storageFactory oss.StorageFactory, db *gorm.DB) *WebDAVHandler {
	return &WebDAVHandler{
		storageFactory: storageFactory,
		db:             db,
	}
}

// ServeHTTP provides a disabled message when WebDAV is not available
func (h *WebDAVHandler) ServeHTTP(c *gin.Context) {
	c.JSON(503, gin.H{
		"error":   "WebDAV support is disabled",
		"message": "This build does not include WebDAV functionality",
	})
}
