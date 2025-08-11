package handlers

import (
	"gorm.io/gorm"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/utils"
	"go.uber.org/zap"
)

// PermissionHandler 权限管理处理器
type PermissionHandler struct {
	*BaseHandler
	DB *gorm.DB
}

// NewPermissionHandler 创建权限管理处理器
func NewPermissionHandler(db *gorm.DB) *PermissionHandler {
	return &PermissionHandler{
		BaseHandler: NewBaseHandler(),
		DB:          db.Debug(),
	}
}

// List 获取权限列表
func (h *PermissionHandler) List(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 构建查询
	query := db.GetDB().Model(&models.Permission{})

	// 处理筛选条件
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if resource := c.Query("resource"); resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if action := c.Query("action"); action != "" {
		query = query.Where("action = ?", action)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取权限总数失败", zap.Error(err))
		h.InternalError(c, "获取权限总数失败")
		return
	}

	// 获取权限列表
	var permissions []models.Permission
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&permissions).Error; err != nil {
		logger.Error("获取权限列表失败", zap.Error(err))
		h.InternalError(c, "获取权限列表失败")
		return
	}

	h.Success(c, gin.H{
		"total": total,
		"items": permissions,
	})
}

// Create 创建权限
func (h *PermissionHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required,min=2,max=32"`
		Description string `json:"description"`
		Resource    string `json:"resource" binding:"required"`
		Action      string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查权限是否已存在
	var count int64
	if err := db.GetDB().Model(&models.Permission{}).
		Where("resource = ? AND action = ?", req.Resource, req.Action).
		Count(&count).Error; err != nil {
		h.InternalError(c, "检查权限失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "权限已存在")
		return
	}

	// 创建权限
	permission := models.Permission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}

	if err := db.GetDB().Create(&permission).Error; err != nil {
		h.InternalError(c, "创建权限失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "CREATE", "PERMISSION", strconv.FormatUint(uint64(permission.ID), 10), "创建权限")

	h.Success(c, permission)
}

// Get 获取权限详情
func (h *PermissionHandler) Get(c *gin.Context) {
	permissionID := c.Param("id")

	var permission models.Permission
	if err := db.GetDB().First(&permission, permissionID).Error; err != nil {
		h.NotFound(c, "权限不存在")
		return
	}

	h.Success(c, permission)
}

// Update 更新权限
func (h *PermissionHandler) Update(c *gin.Context) {
	permissionID := c.Param("id")

	var permission models.Permission
	if err := db.GetDB().First(&permission, permissionID).Error; err != nil {
		h.NotFound(c, "权限不存在")
		return
	}

	var req struct {
		Name        string `json:"name" binding:"required,min=2,max=32"`
		Description string `json:"description"`
		Resource    string `json:"resource" binding:"required"`
		Action      string `json:"action" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查权限是否已被其他记录使用
	var count int64
	if err := db.GetDB().Model(&models.Permission{}).
		Where("resource = ? AND action = ? AND id != ?", req.Resource, req.Action, permissionID).
		Count(&count).Error; err != nil {
		h.InternalError(c, "检查权限失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "权限已存在")
		return
	}

	// 开启事务
	tx := db.GetDB().Begin()

	// 更新权限信息
	if err := tx.Model(&permission).Updates(models.Permission{
		Name:        req.Name,
		Description: req.Description,
		Resource:    req.Resource,
		Action:      req.Action,
	}).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "更新权限失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "UPDATE", "PERMISSION", permissionID, "更新权限信息")

	h.Success(c, permission)
}

// Delete 删除权限
func (h *PermissionHandler) Delete(c *gin.Context) {
	permissionID := c.Param("id")

	var permission models.Permission
	if err := db.GetDB().First(&permission, permissionID).Error; err != nil {
		h.NotFound(c, "权限不存在")
		return
	}

	// 开启事务
	tx := db.GetDB().Begin()

	// 清除角色权限关联
	if err := tx.Model(&permission).Association("Roles").Clear(); err != nil {
		tx.Rollback()
		h.InternalError(c, "清除角色权限关联失败")
		return
	}

	// 删除权限
	if err := tx.Delete(&permission).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "删除权限失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "DELETE", "PERMISSION", permissionID, "删除权限")

	h.Success(c, nil)
}

// createAuditLog 创建审计日志
func (h *PermissionHandler) createAuditLog(c *gin.Context, action, resourceType, resourceID, details string) {
	userID, _ := c.Get("userID")
	username, _ := c.Get("username")

	auditLog := models.AuditLog{
		UserID:       userID.(uint),
		Username:     username.(string),
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
		Status:       "SUCCESS",
	}

	if err := db.GetDB().Create(&auditLog).Error; err != nil {
		logger.Error("创建审计日志失败",
			zap.String("action", action),
			zap.String("resource_type", resourceType),
			zap.String("resource_id", resourceID),
			zap.Error(err))
	}
}
