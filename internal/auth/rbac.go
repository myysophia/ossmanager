package auth

import (
	"errors"
	"net/url"
	"strings"

	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// 定义错误
var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrUserDisabled      = errors.New("用户已禁用")
	ErrPermissionDenied  = errors.New("权限不足")
	ErrDatabaseOperation = errors.New("数据库操作失败")
)

// CheckUserStatus 检查用户状态
func CheckUserStatus(userID uint) error {
	gormDB := db.GetDB()
	var user models.User

	// 查询用户
	err := gormDB.First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("用户不存在", zap.Uint("userID", userID))
			return ErrUserNotFound
		}
		logger.Error("查询用户失败", zap.Uint("userID", userID), zap.Error(err))
		return ErrDatabaseOperation
	}

	// 检查用户状态
	if !user.Status {
		logger.Warn("用户已禁用", zap.Uint("userID", userID), zap.String("username", user.Username))
		return ErrUserDisabled
	}

	return nil
}

// CheckPermission 检查用户是否有对特定资源的操作权限
func CheckPermission(userID uint, resource string, action string) error {
	// 首先检查用户状态
	if err := CheckUserStatus(userID); err != nil {
		return err
	}

	gormDB := db.GetDB()

	// 使用原生SQL查询权限，因为关联查询较为复杂
	// 检查用户通过角色获得的权限
	var count int64
	err := gormDB.Raw(`
		SELECT COUNT(*) FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON rp.role_id = r.id
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ? 
		AND p.resource = ? 
		AND p.action = ?
	`, userID, resource, action).Count(&count).Error

	if err != nil {
		logger.Error("查询用户权限失败",
			zap.Uint("userID", userID),
			zap.String("resource", resource),
			zap.String("action", action),
			zap.Error(err))
		return ErrDatabaseOperation
	}

	// 如果有权限
	if count > 0 {
		return nil
	}

	// 如果用户直接拥有 "管理" 权限，也认为有操作权限
	err = gormDB.Raw(`
		SELECT COUNT(*) FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		JOIN roles r ON rp.role_id = r.id
		JOIN user_roles ur ON r.id = ur.role_id
		WHERE ur.user_id = ? 
		AND p.resource = ? 
		AND p.action = 'manage'
	`, userID, resource).Count(&count).Error

	if err != nil {
		logger.Error("查询用户管理权限失败",
			zap.Uint("userID", userID),
			zap.String("resource", resource),
			zap.Error(err))
		return ErrDatabaseOperation
	}

	// 如果有管理权限
	if count > 0 {
		return nil
	}

	// 权限不足
	logger.Warn("用户权限不足",
		zap.Uint("userID", userID),
		zap.String("resource", resource),
		zap.String("action", action))
	return ErrPermissionDenied
}

// GetUserRoles 获取用户角色
func GetUserRoles(userID uint) ([]models.Role, error) {
	var user models.User
	err := db.GetDB().Preload("Roles").First(&user, userID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Warn("用户不存在", zap.Uint("userID", userID))
			return nil, ErrUserNotFound
		}
		logger.Error("查询用户角色失败", zap.Uint("userID", userID), zap.Error(err))
		return nil, ErrDatabaseOperation
	}

	// 将 []*models.Role 转换为 []models.Role
	roles := make([]models.Role, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = *role
	}

	return roles, nil
}

// GetUserPermissions 获取用户权限
func GetUserPermissions(userID uint) ([]models.Permission, error) {
	// 首先获取用户角色
	roles, err := GetUserRoles(userID)
	if err != nil {
		return nil, err
	}

	// 如果用户没有角色
	if len(roles) == 0 {
		return []models.Permission{}, nil
	}

	// 收集角色ID
	var roleIDs []uint
	for _, role := range roles {
		roleIDs = append(roleIDs, role.ID)
	}

	// 使用 GORM 关联查询获取权限
	var permissions []models.Permission
	err = db.GetDB().Model(&models.Role{}).
		Where("id IN ?", roleIDs).
		Preload("Permissions").
		Find(&roles).Error

	if err != nil {
		logger.Error("查询角色权限失败", zap.Uints("roleIDs", roleIDs), zap.Error(err))
		return nil, ErrDatabaseOperation
	}

	// 收集所有权限并去重
	permissionMap := make(map[uint]models.Permission)
	for _, role := range roles {
		for _, perm := range role.Permissions {
			permissionMap[perm.ID] = *perm
		}
	}

	// 转换为切片
	permissions = make([]models.Permission, 0, len(permissionMap))
	for _, perm := range permissionMap {
		permissions = append(permissions, perm)
	}

	return permissions, nil
}

// CheckBucketAccess 检查用户是否有权限访问指定的桶
func CheckBucketAccess(db *gorm.DB, userID uint, downloadURL, bucketName string) bool {
	// download
	//iotdb-backup.oss-cn-hangzhou
	// care-eu.oss-eu-central-1
	region := ""
	if strings.Contains(downloadURL, "https://") {
		parsedURL, err := url.Parse(downloadURL)
		if err != nil {
			panic(err)
		}
		hostParts := strings.Split(parsedURL.Host, ".")
		regionName := hostParts[1]
		region = strings.TrimPrefix(regionName, "oss-")
		logger.Debug("检查用户桶访问权限", zap.Uint("userID", userID), zap.String("regionCode", region), zap.String("bucketName", bucketName))
	} else {
		region = downloadURL
	}

	var count int64
	err := db.Model(&models.RegionBucketMapping{}).
		Joins("JOIN role_region_bucket_access ON role_region_bucket_access.region_bucket_mapping_id = region_bucket_mapping.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_region_bucket_access.role_id").
		Where("user_roles.user_id = ?  AND region_bucket_mapping.bucket_name = ?",
			userID, bucketName).
		Count(&count).Error

	if err != nil {
		return false
	}
	logger.Debug("检查用户桶访问权限", zap.Uint("userID", userID), zap.String("regionCode", region), zap.String("bucketName", bucketName), zap.Bool("hasAccess", count > 0))
	return count > 0
}

// GetUserAccessibleBuckets 获取用户可访问的桶列表
func GetUserAccessibleBuckets(db *gorm.DB, userID uint, regionCode string) ([]string, error) {
	var buckets []string
	query := db.Model(&models.RegionBucketMapping{}).
		Joins("JOIN role_region_bucket_access ON role_region_bucket_access.region_bucket_mapping_id = region_bucket_mapping.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_region_bucket_access.role_id").
		Where("user_roles.user_id = ?", userID)

	if regionCode != "" {
		query = query.Where("region_bucket_mapping.region_code = ?", regionCode)
	}

	err := query.Distinct().
		Pluck("region_bucket_mapping.bucket_name", &buckets).
		Error
	logger.Debug("获取可访问桶列表", zap.Strings("buckets", buckets))
	return buckets, err
}
