package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// RoleHandler 角色管理处理器
type RoleHandler struct {
	BaseHandler
	DB *gorm.DB
}

// NewRoleHandler 创建角色管理处理器
func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{
		BaseHandler: BaseHandler{},
		DB:          db.Debug(),
	}
}

// List 获取角色列表
func (h *RoleHandler) List(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 构建查询
	query := h.DB.Model(&models.Role{})

	// 处理筛选条件
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取角色总数失败", zap.Error(err))
		h.InternalError(c, "获取角色总数失败")
		return
	}

	// 获取角色列表
	var roles []models.Role
	if err := query.Preload("Permissions").
		Preload("Users").
		Preload("RegionBuckets").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&roles).Error; err != nil {
		logger.Error("获取角色列表失败", zap.Error(err))
		h.InternalError(c, "获取角色列表失败")
		return
	}

	// 构建返回数据
	response := gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"total": total,
			"page":  page,
			"limit": pageSize,
			"items": roles,
		},
	}

	h.Success(c, response)
}

// Create 创建角色
func (h *RoleHandler) Create(c *gin.Context) {
	var req struct {
		Name          string `json:"name" binding:"required,min=2,max=32"`
		Description   string `json:"description"`
		PermissionIDs []uint `json:"permission_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查角色名是否已存在
	var count int64
	if err := h.DB.Model(&models.Role{}).Where("name = ?", req.Name).Count(&count).Error; err != nil {
		h.InternalError(c, "检查角色名失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "角色名已存在")
		return
	}

	// 创建角色
	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
	}

	// 开启事务
	tx := h.DB.Begin()

	if err := tx.Create(&role).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "创建角色失败")
		return
	}

	// 分配权限
	if len(req.PermissionIDs) > 0 {
		var permissions []models.Permission
		if err := tx.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
			tx.Rollback()
			h.InternalError(c, "获取权限失败")
			return
		}

		if err := tx.Model(&role).Association("Permissions").Replace(permissions); err != nil {
			tx.Rollback()
			h.InternalError(c, "分配权限失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "CREATE", "ROLE", strconv.FormatUint(uint64(role.ID), 10), "创建角色")

	h.Success(c, role)
}

// Get 获取角色详情
func (h *RoleHandler) Get(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := h.DB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		h.NotFound(c, "角色不存在")
		return
	}

	h.Success(c, role)
}

// Update 更新角色
func (h *RoleHandler) Update(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := h.DB.First(&role, roleID).Error; err != nil {
		h.NotFound(c, "角色不存在")
		return
	}

	var req struct {
		Name          string `json:"name" binding:"required,min=2,max=32"`
		Description   string `json:"description"`
		PermissionIDs []uint `json:"permission_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查角色名是否已被其他角色使用
	var count int64
	if err := h.DB.Model(&models.Role{}).Where("name = ? AND id != ?", req.Name, roleID).Count(&count).Error; err != nil {
		h.InternalError(c, "检查角色名失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "角色名已存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 更新基本信息
	if err := tx.Model(&role).Updates(models.Role{
		Name:        req.Name,
		Description: req.Description,
	}).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "更新角色失败")
		return
	}

	// 更新权限
	if len(req.PermissionIDs) > 0 {
		var permissions []models.Permission
		if err := tx.Where("id IN ?", req.PermissionIDs).Find(&permissions).Error; err != nil {
			tx.Rollback()
			h.InternalError(c, "获取权限失败")
			return
		}

		if err := tx.Model(&role).Association("Permissions").Replace(permissions); err != nil {
			tx.Rollback()
			h.InternalError(c, "更新权限失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "UPDATE", "ROLE", roleID, "更新角色信息")

	// 重新获取角色信息（包含权限）
	if err := h.DB.Preload("Permissions").First(&role, roleID).Error; err != nil {
		h.NotFound(c, "获取更新后的角色信息失败")
		return
	}

	h.Success(c, role)
}

// Delete 删除角色
func (h *RoleHandler) Delete(c *gin.Context) {
	roleID := c.Param("id")

	var role models.Role
	if err := h.DB.First(&role, roleID).Error; err != nil {
		h.NotFound(c, "角色不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 清除角色权限关联
	if err := tx.Model(&role).Association("Permissions").Clear(); err != nil {
		tx.Rollback()
		h.InternalError(c, "清除角色权限关联失败")
		return
	}

	// 清除用户角色关联
	if err := tx.Model(&role).Association("Users").Clear(); err != nil {
		tx.Rollback()
		h.InternalError(c, "清除用户角色关联失败")
		return
	}

	// 删除角色
	if err := tx.Delete(&role).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "删除角色失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "DELETE", "ROLE", roleID, "删除角色")

	h.Success(c, nil)
}

// createAuditLog 创建审计日志
func (h *RoleHandler) createAuditLog(c *gin.Context, action, resourceType, resourceID, details string) {
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

	if err := h.DB.Create(&auditLog).Error; err != nil {
		logger.Error("创建审计日志失败",
			zap.String("action", action),
			zap.String("resource_type", resourceType),
			zap.String("resource_id", resourceID),
			zap.Error(err))
	}
}

// GetRoleBucketAccess 获取角色的桶访问权限
func (h *RoleHandler) GetRoleBucketAccess(c *gin.Context) {
	roleID := c.Param("id")
	var role models.Role
	if err := h.DB.Preload("RegionBuckets").First(&role, roleID).Error; err != nil {
		h.NotFound(c, "角色不存在")
		return
	}

	h.Success(c, role.RegionBuckets)
}

// UpdateRoleBucketAccess 更新角色的桶访问权限
func (h *RoleHandler) UpdateRoleBucketAccess(c *gin.Context) {
	roleID := c.Param("id")
	var role models.Role
	if err := h.DB.First(&role, roleID).Error; err != nil {
		h.NotFound(c, "角色不存在")
		return
	}

	var request struct {
		RegionBucketMappingIDs []uint `json:"region_bucket_mapping_ids" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		h.BadRequest(c, "请求参数错误")
		return
	}

	// 开启事务
	tx := h.DB.Begin()
	if tx.Error != nil {
		h.InternalError(c, "开启事务失败")
		return
	}

	// 删除原有的桶访问权限
	if err := tx.Model(&role).Association("RegionBuckets").Clear(); err != nil {
		tx.Rollback()
		h.InternalError(c, "清除原有权限失败")
		return
	}

	// 添加新的桶访问权限
	var mappings []models.RegionBucketMapping
	if err := tx.Find(&mappings, request.RegionBucketMappingIDs).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "查询桶映射失败")
		return
	}

	if err := tx.Model(&role).Association("RegionBuckets").Replace(mappings); err != nil {
		tx.Rollback()
		h.InternalError(c, "更新桶访问权限失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	h.Success(c, nil)
}

// ListRoleBucketAccess 获取角色存储桶访问权限列表
func (h *RoleHandler) ListRoleBucketAccess(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))
	roleID := c.Query("role_id")
	bucket := c.Query("bucket")

	query := h.DB.Model(&models.RoleRegionBucketAccess{})

	if roleID != "" {
		query = query.Where("role_id = ?", roleID)
	}
	if bucket != "" {
		query = query.Where("bucket_name = ?", bucket)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		h.InternalError(c, "获取总数失败")
		return
	}

	var accesses []models.RoleRegionBucketAccess
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&accesses).Error; err != nil {
		h.InternalError(c, "获取列表失败")
		return
	}

	h.Success(c, gin.H{
		"total": total,
		"items": accesses,
	})
}

// CreateRoleBucketAccess 创建角色存储桶访问权限
func (h *RoleHandler) CreateRoleBucketAccess(c *gin.Context) {
	var input models.RoleRegionBucketAccess
	if err := c.ShouldBindJSON(&input); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	if err := h.DB.Create(&input).Error; err != nil {
		h.InternalError(c, "创建失败")
		return
	}

	h.Success(c, input)
}

// DeleteRoleBucketAccess 删除角色存储桶访问权限
func (h *RoleHandler) DeleteRoleBucketAccess(c *gin.Context) {
	id := c.Param("id")
	var access models.RoleRegionBucketAccess
	if err := h.DB.First(&access, id).Error; err != nil {
		h.NotFound(c, "访问权限不存在")
		return
	}

	if err := h.DB.Delete(&access).Error; err != nil {
		h.InternalError(c, "删除失败")
		return
	}

	h.Success(c, nil)
}

// ListRegionBucketMappings 获取所有 region-bucket 映射
func (h *RoleHandler) ListRegionBucketMappings(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "300"))

	// 构建查询
	query := h.DB.Model(&models.RegionBucketMapping{})

	// 处理筛选条件
	if regionCode := c.Query("region_code"); regionCode != "" {
		query = query.Where("region_code LIKE ?", "%"+regionCode+"%")
	}
	if bucketName := c.Query("bucket_name"); bucketName != "" {
		query = query.Where("bucket_name LIKE ?", "%"+bucketName+"%")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取 region-bucket 映射总数失败", zap.Error(err))
		h.InternalError(c, "获取总数失败")
		return
	}

	// 获取列表
	var mappings []models.RegionBucketMapping
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&mappings).Error; err != nil {
		logger.Error("获取 region-bucket 映射列表失败", zap.Error(err))
		h.InternalError(c, "获取列表失败")
		return
	}

	// 构建返回数据
	response := gin.H{
		"code": 0,
		"msg":  "success",
		"data": gin.H{
			"total": total,
			"page":  page,
			"limit": pageSize,
			"items": mappings,
		},
	}

	h.Success(c, response)
}
