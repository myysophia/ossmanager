package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/auth"
	"github.com/myysophia/ossmanager-backend/internal/config"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"go.uber.org/zap"
)

// AuthMiddleware 认证中间件
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取 JWT 令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.ResponseError(c, utils.CodeUnauthorized, errors.New("未提供认证令牌"))
			c.Abort()
			return
		}

		// 获取实际的token值
		var tokenString string

		// 检查 Authorization 头格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			// 标准Bearer格式: Bearer <token>
			tokenString = parts[1]
		} else {
			// 直接传递token值
			tokenString = authHeader
		}

		// 输出令牌前几个字符用于调试（不要泄露完整令牌）
		tokenPreview := tokenString
		if len(tokenString) > 10 {
			tokenPreview = tokenString[:10] + "..."
		}
		logger.Debug("收到的令牌", zap.String("tokenPreview", tokenPreview))

		// 获取JWT配置
		jwtConfig := config.GetConfig().JWT
		logger.Debug("JWT配置",
			zap.String("issuer", jwtConfig.Issuer),
			zap.Int("expiresIn", jwtConfig.ExpiresIn),
			zap.String("secretKeyLength", fmt.Sprintf("%d字符", len(jwtConfig.SecretKey))))

		// 解析JWT令牌
		claims, err := auth.ParseToken(tokenString, &jwtConfig)
		if err != nil {
			// 记录详细错误
			logger.Warn("解析JWT令牌失败",
				zap.Error(err),
				zap.String("errorType", fmt.Sprintf("%T", err)))

			// 根据具体错误类型返回更有用的错误信息
			if errors.Is(err, auth.ErrExpiredToken) {
				utils.ResponseError(c, utils.CodeUnauthorized, errors.New("认证令牌已过期"))
			} else if errors.Is(err, auth.ErrInvalidToken) {
				utils.ResponseError(c, utils.CodeUnauthorized, errors.New("认证令牌无效"))
			} else {
				utils.ResponseError(c, utils.CodeUnauthorized, errors.New("认证令牌验证失败: "+err.Error()))
			}
			c.Abort()
			return
		}

		// 将用户信息保存到上下文
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Next()
	}
}
