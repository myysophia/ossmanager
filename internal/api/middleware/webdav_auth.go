package middleware

import (
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/db/models"
)

// AuthFailureInfo 认证失败信息
type AuthFailureInfo struct {
	FailureCount int
	LastFailure  time.Time
	Blocked      bool
	BlockedUntil time.Time
}

// 全局失败计数器（生产环境建议使用 Redis）
var (
	authFailures = make(map[string]*AuthFailureInfo)
	failuresMutex sync.RWMutex
	// 配置参数
	maxFailures   = 5
	blockDuration = 15 * time.Minute
	cleanupTicker *time.Ticker
)

func init() {
	// 启动清理任务，定期清理过期的失败记录
	cleanupTicker = time.NewTicker(10 * time.Minute)
	go func() {
		for {
			select {
			case <-cleanupTicker.C:
				cleanExpiredFailures()
			}
		}
	}()
}

// cleanExpiredFailures 清理过期的失败记录
func cleanExpiredFailures() {
	failuresMutex.Lock()
	defer failuresMutex.Unlock()
	
	now := time.Now()
	for ip, info := range authFailures {
		// 清理 1 小时前的记录
		if now.Sub(info.LastFailure) > time.Hour {
			delete(authFailures, ip)
		}
	}
}

// recordAuthFailure 记录认证失败
func recordAuthFailure(ip string) bool {
	failuresMutex.Lock()
	defer failuresMutex.Unlock()
	
	info, exists := authFailures[ip]
	if !exists {
		info = &AuthFailureInfo{}
		authFailures[ip] = info
	}
	
	info.FailureCount++
	info.LastFailure = time.Now()
	
	// 检查是否需要封锁
	if info.FailureCount >= maxFailures {
		info.Blocked = true
		info.BlockedUntil = time.Now().Add(blockDuration)
		return true // 被封锁
	}
	
	return false
}

// isBlocked 检查 IP 是否被封锁
func isBlocked(ip string) bool {
	failuresMutex.RLock()
	defer failuresMutex.RUnlock()
	
	info, exists := authFailures[ip]
	if !exists {
		return false
	}
	
	// 检查封锁是否过期
	if info.Blocked && time.Now().After(info.BlockedUntil) {
		// 解除封锁
		info.Blocked = false
		info.FailureCount = 0
		return false
	}
	
	return info.Blocked
}

// resetAuthFailures 重置认证失败计数（认证成功时调用）
func resetAuthFailures(ip string) {
	failuresMutex.Lock()
	defer failuresMutex.Unlock()
	
	delete(authFailures, ip)
}

// WebDAVAuthMiddleware WebDAV 认证中间件，支持 Basic Auth 和 Bearer JWT
func WebDAVAuthMiddleware(db *gorm.DB, jwtConfig *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		
		// 检查 IP 是否被封锁
		if isBlocked(clientIP) {
			c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many authentication failures. Please try again later.",
				"retry_after": int(blockDuration.Seconds()),
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")
		
		// 尝试 Token 认证
		if strings.HasPrefix(authHeader, "Token ") {
			token := strings.TrimPrefix(authHeader, "Token ")
			
			// 验证 WebDAV Token - 查找匹配的令牌
			var tokens []models.WebDAVToken
			err := db.Where("expires_at > ?", time.Now()).Find(&tokens).Error
			if err != nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Token realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Token validation failed"})
				c.Abort()
				return
			}
			
			var validToken *models.WebDAVToken
			for _, t := range tokens {
				if t.CheckToken(token) {
					validToken = &t
					break
				}
			}
			
			if validToken == nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Token realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
				c.Abort()
				return
			}
			
			// 验证用户状态
			var user models.User
			err = db.Where("id = ? AND status = ?", validToken.UserID, true).First(&user).Error
			if err != nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Token realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found or disabled"})
				c.Abort()
				return
			}
			
			// 认证成功，重置失败计数
			resetAuthFailures(clientIP)
			
			// 设置用户信息到上下文
			claims := &auth.Claims{
				UserID:   user.ID,
				Username: user.Username,
			}
			c.Set("user", claims)
			c.Set("userID", user.ID)
			c.Set("username", user.Username)
			c.Set("webdav_token", validToken) // 额外设置令牌信息，用于权限检查
			c.Next()
			return
		}
		
		// 尝试 Bearer JWT 认证
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			
			// 验证 JWT Token
			claims, err := auth.ParseToken(token, jwtConfig)
			if err != nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Bearer realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired JWT token"})
				c.Abort()
				return
			}

			// 验证用户状态
			var user models.User
			err = db.Where("id = ? AND status = ?", claims.UserID, true).First(&user).Error
			if err != nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Bearer realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found or disabled"})
				c.Abort()
				return
			}

			// 认证成功，重置失败计数
			resetAuthFailures(clientIP)
			
			// 设置用户信息到上下文
			c.Set("user", claims)
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Next()
			return
		}

		// 尝试 Basic Auth 认证
		if strings.HasPrefix(authHeader, "Basic ") {
			// 解析 Basic Auth
			payload, err := base64.StdEncoding.DecodeString(authHeader[6:])
			if err != nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
				c.Abort()
				return
			}

			credentials := strings.SplitN(string(payload), ":", 2)
			if len(credentials) != 2 {
				recordAuthFailure(clientIP)
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
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				c.Abort()
				return
			}

			// 验证密码
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				recordAuthFailure(clientIP)
				c.Header("WWW-Authenticate", `Basic realm="OSS Manager WebDAV"`)
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
				c.Abort()
				return
			}

			// 认证成功，重置失败计数
			resetAuthFailures(clientIP)

			// 设置用户信息到上下文
			claims := &auth.Claims{
				UserID:   user.ID,
				Username: user.Username,
			}
			c.Set("user", claims)
			c.Set("userID", user.ID)
			c.Set("username", user.Username)
			c.Next()
			return
		}

		// 没有认证信息
		c.Header("WWW-Authenticate", `Bearer realm="OSS Manager WebDAV", Basic realm="OSS Manager WebDAV"`)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required. Please provide Basic Auth or Bearer JWT token"})
		c.Abort()
	}
}
