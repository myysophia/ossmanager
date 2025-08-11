package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"go.uber.org/zap"
	"runtime/debug"
)

// RecoveryMiddleware 错误恢复中间件
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 记录堆栈信息
				stack := string(debug.Stack())
				logger.Error("服务发生panic",
					zap.Any("error", err),
					zap.String("stack", stack),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("ip", c.ClientIP()),
				)

				// 响应内部服务器错误
				utils.ResponseError(c, utils.CodeInternalError, err.(error))
				c.Abort()
			}
		}()
		c.Next()
	}
}
