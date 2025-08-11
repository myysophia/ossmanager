package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/db"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"go.uber.org/zap"
)

// PermissionMiddleware 权限检查中间件
func PermissionMiddleware(requiredPermissions ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			utils.ResponseError(c, utils.CodeUnauthorized, errors.New("未登录"))
			c.Abort()
			return
		}

		// 检查是否有权限
		hasPermission, err := checkPermissions(userID.(uint), requiredPermissions...)
		if err != nil {
			logger.Error("检查权限失败", zap.Error(err))
			utils.ResponseError(c, utils.CodeInternalError, errors.New("检查权限失败"))
			c.Abort()
			return
		}

		if !hasPermission {
			utils.ResponseError(c, utils.CodeForbidden, errors.New("没有权限执行此操作"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RoleMiddleware 角色检查中间件
func RoleMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从上下文获取用户ID
		userID, exists := c.Get("userID")
		if !exists {
			utils.ResponseError(c, utils.CodeUnauthorized, errors.New("未登录"))
			c.Abort()
			return
		}

		// 检查是否有角色
		hasRole, err := checkRoles(userID.(uint), requiredRoles...)
		if err != nil {
			logger.Error("检查角色失败", zap.Error(err))
			utils.ResponseError(c, utils.CodeInternalError, errors.New("检查角色失败"))
			c.Abort()
			return
		}

		if !hasRole {
			utils.ResponseError(c, utils.CodeForbidden, errors.New("没有权限执行此操作"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// AdminMiddleware 管理员检查中间件
func AdminMiddleware() gin.HandlerFunc {
	return RoleMiddleware()
}

// checkPermissions 检查用户是否拥有指定权限
func checkPermissions(userID uint, permissions ...string) (bool, error) {
	if len(permissions) == 0 {
		return true, nil
	}

	// 获取用户权限
	var userPermissions []models.Permission
	err := db.GetDB().Model(&models.User{}).
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Joins("JOIN role_permissions ON role_permissions.role_id = roles.id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("users.id = ?", userID).
		Select("permissions.*").
		Find(&userPermissions).Error
	if err != nil {
		return false, err
	}

	// 转换为权限代码集合，方便查找
	permissionMap := make(map[string]bool)
	for _, perm := range userPermissions {
		permissionMap[perm.Name] = true
	}

	// 检查是否拥有所有指定权限
	for _, permission := range permissions {
		if !permissionMap[permission] {
			return false, nil
		}
	}

	return true, nil
}

// checkRoles 检查用户是否拥有指定角色
func checkRoles(userID uint, roles ...string) (bool, error) {
	if len(roles) == 0 {
		return true, nil
	}

	// 获取用户角色
	var userRoles []models.Role
	err := db.GetDB().Model(&models.User{}).
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("users.id = ?", userID).
		Select("roles.*").
		Find(&userRoles).Error
	if err != nil {
		return false, err
	}

	// 转换为角色代码集合，方便查找
	roleMap := make(map[string]bool)
	for _, role := range userRoles {
		roleMap[role.Name] = true
	}

	// 检查是否拥有所有指定角色
	for _, role := range roles {
		if !roleMap[role] {
			return false, nil
		}
	}

	return true, nil
}
