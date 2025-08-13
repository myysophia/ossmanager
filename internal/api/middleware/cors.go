package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// CorsMiddleware 跨域中间件
func CorsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		// 支持WebDAV和REST API所需的所有头部
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Depth, If, If-None-Match, Lock-Token, Overwrite, Timeout, Destination, X-User-ID, X-Bucket, X-Requested-With, Accept, Origin")
		// 支持WebDAV和REST API所需的所有方法
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, MKCOL, COPY, MOVE, PROPFIND, PROPPATCH, LOCK, UNLOCK")
		// 暴露浏览器可以访问的响应头部
		c.Writer.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Last-Modified, ETag, DAV, X-Total-Count")
		// 预检请求缓存24小时
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
