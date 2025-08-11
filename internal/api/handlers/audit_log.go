package handlers

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
)

// AuditLogHandler 审计日志处理器
type AuditLogHandler struct {
	*BaseHandler
}

// NewAuditLogHandler 创建审计日志处理器
func NewAuditLogHandler() *AuditLogHandler {
	return &AuditLogHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// ListAuditLogs 获取审计日志列表
func (h *AuditLogHandler) ListAuditLogs(c *gin.Context) {
	// 获取分页参数
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	// 转换并验证分页参数
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
		logger.Warn("无效的page参数，使用默认值1", zap.String("原始值", pageStr))
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize <= 0 {
		pageSize = 10
		logger.Warn("无效的page_size参数，使用默认值10", zap.String("原始值", pageSizeStr))
	}

	// 记录请求参数
	logger.Info("获取审计日志列表请求",
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method))

	// 构建查询条件
	query := db.GetDB().Model(&models.AuditLog{})

	// 处理时间范围筛选
	startTimeStr := c.Query("start_time")
	endTimeStr := c.Query("end_time")

	if startTimeStr != "" {
		startTime, err := time.Parse(time.RFC3339, startTimeStr)
		if err == nil {
			query = query.Where("created_at >= ?", startTime)
			logger.Info("应用开始时间筛选", zap.Time("startTime", startTime))
		} else {
			logger.Warn("无效的start_time参数", zap.String("原始值", startTimeStr), zap.Error(err))
		}
	}

	if endTimeStr != "" {
		endTime, err := time.Parse(time.RFC3339, endTimeStr)
		if err == nil {
			query = query.Where("created_at <= ?", endTime)
			logger.Info("应用结束时间筛选", zap.Time("endTime", endTime))
		} else {
			logger.Warn("无效的end_time参数", zap.String("原始值", endTimeStr), zap.Error(err))
		}
	}

	// 处理用户筛选
	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err == nil {
			query = query.Where("user_id = ?", userID)
			logger.Info("应用用户ID筛选", zap.Int("userID", userID))
		} else {
			logger.Warn("无效的user_id参数", zap.String("原始值", userIDStr), zap.Error(err))
		}
	}

	username := c.Query("username")
	if username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
		logger.Info("应用用户名筛选", zap.String("username", username))
	}

	// 处理操作类型筛选
	action := c.Query("action")
	if action != "" {
		query = query.Where("action = ?", action)
		logger.Info("应用操作类型筛选", zap.String("action", action))
	}

	// 处理资源类型筛选
	resourceType := c.Query("resource_type")
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
		logger.Info("应用资源类型筛选", zap.String("resourceType", resourceType))
	}

	// 处理状态筛选
	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
		logger.Info("应用状态筛选", zap.String("status", status))
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取审计日志总数失败", zap.Error(err))
		h.InternalError(c, "获取审计日志总数失败")
		return
	}

	// 记录总数
	logger.Info("审计日志总数", zap.Int64("total", total))

	// 应用分页
	offset := (page - 1) * pageSize
	query = query.Offset(offset).Limit(pageSize)
	logger.Info("应用分页", zap.Int("offset", offset), zap.Int("limit", pageSize))

	// 按时间倒序
	query = query.Order("created_at DESC")

	// 执行查询
	var logs []models.AuditLog
	if err := query.Find(&logs).Error; err != nil {
		logger.Error("获取审计日志列表失败", zap.Error(err))
		h.InternalError(c, "获取审计日志列表失败")
		return
	}

	// 记录查询结果
	logger.Info("审计日志查询结果", zap.Int("结果数量", len(logs)))

	// 返回结果
	h.Success(c, gin.H{
		"total":     total,
		"page":      page,
		"page_size": pageSize,
		"items":     logs,
	})
}
