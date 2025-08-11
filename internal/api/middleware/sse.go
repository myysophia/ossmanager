package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"go.uber.org/zap"
)

// SSEMiddleware SSE连接专用中间件
// 用于确保SSE连接的稳定性，禁用缓存和优化连接设置
func SSEMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取请求路径，仅对SSE流接口进行特殊处理
		if isSSEEndpoint(c.Request.URL.Path) {
			logger.Debug("处理SSE请求",
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("user_agent", c.Request.UserAgent()),
				zap.String("remote_addr", c.ClientIP()),
			)

			// 强制使用HTTP/1.1，禁用HTTP/2
			c.Header("Connection", "keep-alive")
			c.Header("Upgrade", "") // 清除可能的升级头部

			// SSE专用头部设置
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache")
			c.Header("Expires", "0")

			// 禁用响应缓冲
			c.Header("X-Accel-Buffering", "no") // Nginx
			c.Header("X-Content-Type-Options", "nosniff")

			// 跨域支持（如果需要）
			origin := c.GetHeader("Origin")
			if origin != "" {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Expose-Headers", "Content-Type")
			}
		}

		c.Next()

		// 请求完成后的日志记录（仅对SSE请求）
		if isSSEEndpoint(c.Request.URL.Path) {
			logger.Debug("SSE请求完成",
				zap.String("path", c.Request.URL.Path),
				zap.Int("status", c.Writer.Status()),
				zap.String("remote_addr", c.ClientIP()),
			)
		}
	}
}

// isSSEEndpoint 判断是否为SSE端点
func isSSEEndpoint(path string) bool {
	// 检查是否为流式端点
	return len(path) >= 7 && path[len(path)-7:] == "/stream"
}

// HTTP1OnlyMiddleware 强制使用HTTP/1.1的中间件
// 用于特定路由组，确保不使用HTTP/2
func HTTP1OnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 强制使用HTTP/1.1
		c.Request.ProtoMajor = 1
		c.Request.ProtoMinor = 1
		c.Request.Proto = "HTTP/1.1"

		// 设置响应头
		c.Header("Connection", "keep-alive")
		c.Header("Upgrade", "") // 移除升级头部

		c.Next()
	}
}

// NoBufferMiddleware 禁用缓冲的中间件
// 用于需要实时响应的接口
func NoBufferMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 禁用各种缓冲
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.Header("Pragma", "no-cache")
		c.Header("Expires", "0")
		c.Header("X-Accel-Buffering", "no") // Nginx专用

		c.Next()
	}
}
