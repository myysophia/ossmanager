package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/utils"
	"go.uber.org/zap"
)

// WebDAVProxyAuthMiddleware provides authentication for WebDAV proxy endpoints
// It supports both JWT tokens and WebDAV tokens for consistent authentication
func WebDAVProxyAuthMiddleware(db *gorm.DB, jwtConfig *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var user *auth.Claims
		var webdavToken *models.WebDAVToken
		var err error

		// Try to authenticate using standard JWT token first (Authorization: Bearer <token>)
		if jwtToken := extractBearerToken(c); jwtToken != "" {
			user, err = authenticateJWT(jwtToken, jwtConfig, db)
			if err != nil {
				logger.Warn("JWT authentication failed", zap.Error(err))
				respondUnauthorized(c, "Invalid or expired JWT token")
				return
			}
		} else if webdavTokenStr := extractWebDAVToken(c); webdavTokenStr != "" {
			// Try WebDAV token authentication (X-WebDAV-Token header or query parameter)
			user, webdavToken, err = authenticateWebDAVToken(webdavTokenStr, db)
			if err != nil {
				logger.Warn("WebDAV token authentication failed", zap.Error(err))
				respondUnauthorized(c, "Invalid or expired WebDAV token")
				return
			}
		} else {
			// No authentication provided
			respondUnauthorized(c, "Authentication required")
			return
		}

		// Set user context
		c.Set("user", user)
		c.Set("userID", user.UserID)
		c.Set("username", user.Username)

		// If WebDAV token was used, also set token context for bucket validation
		if webdavToken != nil {
			c.Set("webdav_token", webdavToken)
		}

		c.Next()
	}
}

// extractBearerToken extracts JWT token from Authorization header
func extractBearerToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		return authHeader[7:]
	}
	return ""
}

// extractWebDAVToken extracts WebDAV token from headers or query parameters
func extractWebDAVToken(c *gin.Context) string {
	// Try X-WebDAV-Token header first
	if token := c.GetHeader("X-WebDAV-Token"); token != "" {
		return token
	}
	
	// Try webdav_token query parameter
	if token := c.Query("webdav_token"); token != "" {
		return token
	}
	
	// Try token query parameter (for compatibility)
	if token := c.Query("token"); token != "" {
		return token
	}
	
	return ""
}

// authenticateJWT validates JWT token and returns user claims
func authenticateJWT(token string, jwtConfig *config.JWTConfig, db *gorm.DB) (*auth.Claims, error) {
	// Parse and validate JWT token
	claims, err := auth.ParseToken(token, jwtConfig)
	if err != nil {
		return nil, err
	}

	// Verify user exists and is active
	var user models.User
	err = db.Where("id = ? AND status = ?", claims.UserID, true).First(&user).Error
	if err != nil {
		return nil, err
	}

	// Update claims with current user info (in case data changed)
	claims.Username = user.Username
	
	return claims, nil
}

// authenticateWebDAVToken validates WebDAV token and returns user claims and token info
func authenticateWebDAVToken(tokenStr string, db *gorm.DB) (*auth.Claims, *models.WebDAVToken, error) {
	// Find all active WebDAV tokens
	var tokens []models.WebDAVToken
	err := db.Where("expires_at > ?", time.Now()).Find(&tokens).Error
	if err != nil {
		return nil, nil, err
	}

	// Find matching token
	var validToken *models.WebDAVToken
	for _, token := range tokens {
		if token.CheckToken(tokenStr) {
			validToken = &token
			break
		}
	}

	if validToken == nil {
		return nil, nil, fmt.Errorf("invalid or expired WebDAV token")
	}

	// Get associated user
	var user models.User
	err = db.Where("id = ? AND status = ?", validToken.UserID, true).First(&user).Error
	if err != nil {
		return nil, nil, err
	}

	// Create claims
	claims := &auth.Claims{
		UserID:   user.ID,
		Username: user.Username,
	}

	return claims, validToken, nil
}

// respondUnauthorized sends unauthorized response
func respondUnauthorized(c *gin.Context, message string) {
	c.Header("WWW-Authenticate", `Bearer realm="OSS Manager WebDAV Proxy"`)
	c.JSON(http.StatusUnauthorized, gin.H{
		"error":   "Unauthorized",
		"message": message,
		"code":    utils.CodeUnauthorized,
	})
	c.Abort()
}
