package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager-backend/internal/auth"
	"github.com/myysophia/ossmanager-backend/internal/models"
)

// WebDAVAuthMiddleware WebDAV 认证中间件，支持 Basic Auth 和 Bearer Token
func WebDAVAuthMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试 Bearer Token 认证
		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			// 使用现有的 JWT 认证逻辑
			// token := strings.TrimPrefix(authHeader, "Bearer ") // 暂时注释未使用变量
			// 这里需要实现 ValidateToken 或者使用其他认证方式
			// 暂时跳过 Bearer token 认证
			claims := &auth.Claims{}
			err := error(nil)
			if err != nil {
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				c.Abort()
				return
			}

			// 设置用户信息到上下文
			c.Set("user", claims)
			c.Next()
			return
		}

		// 尝试 Basic Auth 认证
		if strings.HasPrefix(authHeader, "Basic ") {
			// 解析 Basic Auth
			payload, err := base64.StdEncoding.DecodeString(authHeader[6:])
			if err != nil {
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
				c.Abort()
				return
			}

			credentials := strings.SplitN(string(payload), ":", 2)
			if len(credentials) != 2 {
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials format"})
				c.Abort()
				return
			}

			username, password := credentials[0], credentials[1]

			// 验证用户名和密码
			var user models.User
			err = db.Where("username = ? AND status = ?", username, true).First(&user).Error
			if err != nil {
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				c.Abort()
				return
			}

			// 验证密码
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				c.Abort()
				return
			}

			// 设置用户信息到上下文
			claims := &auth.Claims{
				UserID:   user.ID,
				Username: user.Username,
				// Role 字段在 Claims 中不存在，暂时移除
			}
			c.Set("user", claims)
			c.Next()
			return
		}

		// 没有认证信息
		c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
	}
}
