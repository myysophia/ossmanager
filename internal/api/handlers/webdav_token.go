package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/utils"
	"go.uber.org/zap"
)

// WebDAVTokenHandler WebDAV 令牌处理器
type WebDAVTokenHandler struct {
	db *gorm.DB
}

// NewWebDAVTokenHandler 创建 WebDAV 令牌处理器
func NewWebDAVTokenHandler(db *gorm.DB) *WebDAVTokenHandler {
	return &WebDAVTokenHandler{db: db}
}

// CreateTokenRequest 创建令牌请求结构
type CreateTokenRequest struct {
	Bucket    string `json:"bucket" binding:"required"`
	ExpiresIn int    `json:"expires_in"` // 过期时间（小时），默认 24 小时
}

// generateSecureToken 生成安全的随机令牌（32 字节）
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Create 生成新的 WebDAV 访问令牌
func (h *WebDAVTokenHandler) Create(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request format", err)
		return
	}

	// 获取当前用户
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 检查用户对该存储桶的访问权限
	if !auth.CheckBucketAccess(h.db, userID, "", req.Bucket) {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied to bucket", nil)
		return
	}

	// 设置默认过期时间
	if req.ExpiresIn <= 0 {
		req.ExpiresIn = 24 // 默认 24 小时
	}
	
	// 限制最大过期时间为 30 天
	if req.ExpiresIn > 30*24 {
		req.ExpiresIn = 30 * 24
	}

	expiresAt := time.Now().Add(time.Duration(req.ExpiresIn) * time.Hour)

	// 生成随机令牌
	token, err := generateSecureToken()
	if err != nil {
		logger.Error("Failed to generate secure token", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token", err)
		return
	}

	// 创建令牌记录
	webdavToken := &models.WebDAVToken{
		UserID:    userID,
		Bucket:    req.Bucket,
		ExpiresAt: expiresAt,
	}

	// 设置令牌哈希
	if err := webdavToken.SetToken(token); err != nil {
		logger.Error("Failed to hash token", zap.Error(err))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to process token", err)
		return
	}

	// 保存到数据库
	if err := h.db.Create(webdavToken).Error; err != nil {
		logger.Error("Failed to create WebDAV token", zap.Error(err), zap.Uint("userID", userID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create token", err)
		return
	}

	// 构造响应，包含原始令牌
	response := webdavToken.ToResponse()
	response.Token = token // 只在创建时返回原始令牌

	logger.Info("WebDAV token created successfully", 
		zap.Uint("userID", userID),
		zap.String("bucket", req.Bucket),
		zap.Uint("tokenID", webdavToken.ID))

	utils.SuccessResponse(c, "Token created successfully", response)
}

// List 获取当前用户的所有 WebDAV 令牌
func (h *WebDAVTokenHandler) List(c *gin.Context) {
	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	var tokens []models.WebDAVToken
	query := h.db.Where("user_id = ?", userID)

	// 支持按存储桶过滤
	if bucket := c.Query("bucket"); bucket != "" {
		query = query.Where("bucket = ?", bucket)
	}

	// 可选择是否包含过期的令牌
	if includeExpired := c.Query("include_expired"); includeExpired != "true" {
		query = query.Where("expires_at > ?", time.Now())
	}

	if err := query.Order("created_at DESC").Find(&tokens).Error; err != nil {
		logger.Error("Failed to list WebDAV tokens", zap.Error(err), zap.Uint("userID", userID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to retrieve tokens", err)
		return
	}

	// 转换为响应格式
	responses := make([]*models.WebDAVTokenResponse, 0)
	for _, token := range tokens {
		responses = append(responses, token.ToResponse())
	}

	utils.SuccessResponse(c, "Tokens retrieved successfully", responses)
}

// Delete 撤销指定的 WebDAV 令牌
func (h *WebDAVTokenHandler) Delete(c *gin.Context) {
	tokenIDStr := c.Param("id")
	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid token ID", err)
		return
	}

	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 查找令牌
	var token models.WebDAVToken
	if err := h.db.Where("id = ? AND user_id = ?", tokenID, userID).First(&token).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			utils.ErrorResponse(c, http.StatusNotFound, "Token not found", nil)
			return
		}
		logger.Error("Failed to find WebDAV token", zap.Error(err), zap.Uint("userID", userID))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to find token", err)
		return
	}

	// 软删除令牌
	if err := h.db.Delete(&token).Error; err != nil {
		logger.Error("Failed to delete WebDAV token", zap.Error(err), zap.Uint("tokenID", uint(tokenID)))
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete token", err)
		return
	}

	logger.Info("WebDAV token deleted successfully", 
		zap.Uint("userID", userID),
		zap.Uint("tokenID", uint(tokenID)),
		zap.String("bucket", token.Bucket))

	utils.SuccessResponse(c, "Token deleted successfully", nil)
}

// GetConnectionInfo 获取 WebDAV 连接信息
func (h *WebDAVTokenHandler) GetConnectionInfo(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bucket parameter is required", nil)
		return
	}

	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 检查用户对该存储桶的访问权限
	if !auth.CheckBucketAccess(h.db, userID, "", bucket) {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied to bucket", nil)
		return
	}

	// 获取用户信息
	var user models.User
	if err := h.db.Where("id = ?", userID).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user info", err)
		return
	}

	// 构建连接信息
	protocol := "http"
	if c.Request.TLS != nil {
		protocol = "https"
	}

	connectionInfo := map[string]interface{}{
		"url":      fmt.Sprintf("%s://%s/webdav/%s", protocol, c.Request.Host, bucket),
		"bucket":   bucket,
		"username": user.Username,
		"supported_methods": []string{"GET", "PUT", "DELETE", "MKCOL", "COPY", "MOVE", "PROPFIND", "PROPPATCH", "LOCK", "UNLOCK"},
		"instructions": map[string]interface{}{
			"windows": []string{
				"Open File Explorer",
				"Right-click 'This PC' → 'Map network drive'",
				"Select drive letter and enter address",
				"Use your username and WebDAV token as credentials",
			},
			"macos": []string{
				"Open Finder",
				"Press Cmd+K or go to 'Go' → 'Connect to Server'",
				"Enter server address",
				"Use your username and WebDAV token as credentials",
			},
			"linux": []string{
				"Use davfs2: sudo mount -t davfs [URL] /mnt/webdav",
				"Or use cadaver: cadaver [URL]",
				"Enter your username and WebDAV token when prompted",
			},
		},
	}

	utils.SuccessResponse(c, "Connection info retrieved successfully", connectionInfo)
}

// TestConnection 测试 WebDAV 连接
func (h *WebDAVTokenHandler) TestConnection(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		utils.ErrorResponse(c, http.StatusBadRequest, "Bucket parameter is required", nil)
		return
	}

	userID := utils.GetUserID(c)
	if userID == 0 {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// 检查用户对该存储桶的访问权限
	if !auth.CheckBucketAccess(h.db, userID, "", bucket) {
		utils.ErrorResponse(c, http.StatusForbidden, "Access denied to bucket", nil)
		return
	}

	// 简单的连接测试，检查存储桶是否可访问
	testResult := map[string]interface{}{
		"bucket":        bucket,
		"accessible":    true,
		"test_time":     time.Now(),
		"supported":     true,
		"message":       "Connection test successful",
	}

	utils.SuccessResponse(c, "Connection test completed", testResult)
}

// CleanExpiredTokens 清理过期的令牌（可以通过定时任务调用）
func (h *WebDAVTokenHandler) CleanExpiredTokens() error {
	result := h.db.Where("expires_at < ?", time.Now()).Delete(&models.WebDAVToken{})
	if result.Error != nil {
		logger.Error("Failed to clean expired WebDAV tokens", zap.Error(result.Error))
		return result.Error
	}

	if result.RowsAffected > 0 {
		logger.Info("Cleaned expired WebDAV tokens", zap.Int64("count", result.RowsAffected))
	}

	return nil
}
