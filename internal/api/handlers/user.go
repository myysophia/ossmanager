package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/utils"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

// UserHandler 用户管理处理器
type UserHandler struct {
	*BaseHandler
}

// NewUserHandler 创建用户管理处理器
func NewUserHandler() *UserHandler {
	return &UserHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// List 获取用户列表
func (h *UserHandler) List(c *gin.Context) {
	// 获取分页参数
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	// 构建查询
	query := db.GetDB().Model(&models.User{})

	// 处理筛选条件
	if username := c.Query("username"); username != "" {
		query = query.Where("username LIKE ?", "%"+username+"%")
	}
	if email := c.Query("email"); email != "" {
		query = query.Where("email LIKE ?", "%"+email+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status == "true")
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取用户总数失败", zap.Error(err))
		h.InternalError(c, "获取用户总数失败")
		return
	}

	// 获取用户列表
	var users []models.User
	if err := query.Preload("Roles").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		logger.Error("获取用户列表失败", zap.Error(err))
		h.InternalError(c, "获取用户列表失败")
		return
	}

	h.Success(c, gin.H{
		"total": total,
		"items": users,
	})
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=32"`
		Password string `json:"password" binding:"required,min=6,max=32"`
		Email    string `json:"email" binding:"required,email"`
		RealName string `json:"real_name"`
		RoleIDs  []uint `json:"role_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查用户名是否已存在
	var count int64
	if err := db.GetDB().Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		h.InternalError(c, "检查用户名失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "用户名已存在")
		return
	}

	// 检查邮箱是否已存在
	if err := db.GetDB().Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		h.InternalError(c, "检查邮箱失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "邮箱已存在")
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.InternalError(c, "密码加密失败")
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		RealName: req.RealName,
		Status:   true,
	}

	// 开启事务
	tx := db.GetDB().Begin()

	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "创建用户失败")
		return
	}

	// 分配角色
	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
			tx.Rollback()
			h.InternalError(c, "获取角色失败")
			return
		}

		if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
			tx.Rollback()
			h.InternalError(c, "分配角色失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "CREATE", "USER", strconv.FormatUint(uint64(user.ID), 10), "创建用户")

	h.Success(c, user)
}

// Get 获取用户详情
func (h *UserHandler) Get(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := db.GetDB().Preload("Roles").First(&user, userID).Error; err != nil {
		h.NotFound(c, "用户不存在")
		return
	}

	h.Success(c, user)
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		h.NotFound(c, "用户不存在")
		return
	}

	var req struct {
		Email    string `json:"email" binding:"required,email"`
		RealName string `json:"real_name"`
		Status   *bool  `json:"status"`
		RoleIDs  []uint `json:"role_ids"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查邮箱是否已被其他用户使用
	var count int64
	if err := db.GetDB().Model(&models.User{}).Where("email = ? AND id != ?", req.Email, userID).Count(&count).Error; err != nil {
		h.InternalError(c, "检查邮箱失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "邮箱已存在")
		return
	}

	// 开启事务
	tx := db.GetDB().Begin()

	// 更新基本信息
	updates := map[string]interface{}{
		"email":     req.Email,
		"real_name": req.RealName,
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}

	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "更新用户失败")
		return
	}

	// 更新角色
	if len(req.RoleIDs) > 0 {
		var roles []models.Role
		if err := tx.Where("id IN ?", req.RoleIDs).Find(&roles).Error; err != nil {
			tx.Rollback()
			h.InternalError(c, "获取角色失败")
			return
		}

		if err := tx.Model(&user).Association("Roles").Replace(roles); err != nil {
			tx.Rollback()
			h.InternalError(c, "更新角色失败")
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "UPDATE", "USER", userID, "更新用户信息")

	// 重新获取用户信息（包含角色）
	if err := db.GetDB().Preload("Roles").First(&user, userID).Error; err != nil {
		h.NotFound(c, "获取更新后的用户信息失败")
		return
	}

	h.Success(c, user)
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	userID := c.Param("id")

	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		h.NotFound(c, "用户不存在")
		return
	}

	// 开启事务
	tx := db.GetDB().Begin()

	// 清除用户角色关联
	if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
		tx.Rollback()
		h.InternalError(c, "清除用户角色关联失败")
		return
	}

	// 删除用户
	if err := tx.Delete(&user).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "删除用户失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 记录审计日志
	h.createAuditLog(c, "DELETE", "USER", userID, "删除用户")

	h.Success(c, nil)
}

// createAuditLog 创建审计日志
func (h *UserHandler) createAuditLog(c *gin.Context, action, resourceType, resourceID, details string) {
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

// GetUserBucketAccess 获取用户可访问的存储桶列表
func (h *UserHandler) GetUserBucketAccess(c *gin.Context) {
	userID := c.Param("id")

	// 获取用户信息
	var user models.User
	if err := db.GetDB().Preload("Roles").First(&user, userID).Error; err != nil {
		h.NotFound(c, "用户不存在")
		return
	}

	// 获取用户所有角色的ID
	var roleIDs []uint
	for _, role := range user.Roles {
		roleIDs = append(roleIDs, role.ID)
	}

	// 如果没有角色，返回空列表
	if len(roleIDs) == 0 {
		h.Success(c, []models.RegionBucketMapping{})
		return
	}

	// 查询用户角色关联的所有存储桶
	var buckets []models.RegionBucketMapping
	if err := db.GetDB().
		Joins("JOIN role_region_bucket_access ON role_region_bucket_access.region_bucket_mapping_id = region_bucket_mapping.id").
		Where("role_region_bucket_access.role_id IN ?", roleIDs).
		Distinct().
		Find(&buckets).Error; err != nil {
		logger.Error("获取用户存储桶列表失败", zap.Error(err))
		h.InternalError(c, "获取用户存储桶列表失败")
		return
	}

	h.Success(c, buckets)
}
