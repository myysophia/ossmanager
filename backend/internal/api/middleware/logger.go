package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"go.uber.org/zap"
	"time"
)

// LoggerMiddleware 日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		ip := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 处理请求
		c.Next()

		// 结束时间
		end := time.Now()
		latency := end.Sub(start)
		status := c.Writer.Status()

		// 获取用户信息
		userID, exists := c.Get("userID")
		username, _ := c.Get("username")

		// 日志记录字段
		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", ip),
			zap.String("user-agent", userAgent),
			zap.Int("status", status),
			zap.Duration("latency", latency),
		}

		// 添加用户信息
		if exists {
			fields = append(fields, zap.Any("user_id", userID))
			fields = append(fields, zap.Any("username", username))
		}

		// 记录错误信息
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				fields = append(fields, zap.String("error", e.Error()))
			}
			logger.Error("请求处理失败", fields...)
			return
		}

		// 根据状态码决定日志级别
		if status >= 500 {
			logger.Error("服务器错误", fields...)
		} else if status >= 400 {
			logger.Warn("客户端错误", fields...)
		} else {
			logger.Info("请求处理成功", fields...)
		}

		// 记录审计日志（可选）
		if exists && c.Request.Method != "GET" {
			go recordAuditLog(c, userID.(uint), username.(string), path, method, status)
		}
	}
}

// recordAuditLog 记录审计日志
func recordAuditLog(c *gin.Context, userID uint, username, path, method string, status int) {
	// 这里可以根据实际需求记录审计日志到数据库
	// 例如：
	/*
		auditLog := models.AuditLog{
			UserID:       userID,
			Username:     username,
			Action:       method,
			ResourceType: "API",
			ResourceID:   path,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.Request.UserAgent(),
			Status:       status,
		}
		db.GetDB().Create(&auditLog)
	*/
}
