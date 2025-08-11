package db

import (
	"fmt"
	"github.com/myysophia/ossmanager-backend/internal/config"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var db *gorm.DB

// Init 初始化数据库连接
func Init(cfg *config.DatabaseConfig) error {
	var err error

	// 配置 GORM
	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名
		},
		// 禁用默认事务
		SkipDefaultTransaction: true,
	}

	// 连接数据库
	db, err = gorm.Open(postgres.Open(cfg.GetDSN()), gormConfig)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}

	// 配置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取数据库实例失败: %w", err)
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	// 自动迁移数据库表
	//if err := autoMigrate(); err != nil {
	//	return fmt.Errorf("数据库迁移失败: %w", err)
	//}

	// 初始化基础数据
	//if err := initBaseData(); err != nil {
	//	return fmt.Errorf("初始化基础数据失败: %w", err)
	//}

	logger.Info("数据库初始化成功")
	return nil
}

// GetDB 获取数据库连接
func GetDB() *gorm.DB {
	return db
}

// autoMigrate 自动迁移数据库表
func autoMigrate() error {
	return db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.OSSFile{},
		&models.OSSConfig{},
		&models.AuditLog{},
	)
}

// initBaseData 初始化基础数据
func initBaseData() error {
	// 创建默认管理员角色
	adminRole := &models.Role{
		Name:        "admin",
		Description: "系统管理员",
	}

	if err := db.FirstOrCreate(adminRole, models.Role{Name: "admin"}).Error; err != nil {
		return err
	}

	// 创建基础权限
	permissions := []models.Permission{
		{Name: "user_manage", Description: "用户管理", Resource: "user", Action: "manage"},
		{Name: "role_manage", Description: "角色管理", Resource: "role", Action: "manage"},
		{Name: "file_manage", Description: "文件管理", Resource: "file", Action: "manage"},
		{Name: "oss_config", Description: "OSS配置管理", Resource: "oss_config", Action: "manage"},
	}

	for _, perm := range permissions {
		if err := db.FirstOrCreate(&perm, models.Permission{Name: perm.Name}).Error; err != nil {
			return err
		}
	}

	// 为管理员角色分配所有权限
	if err := db.Model(adminRole).Association("Permissions").Replace(&permissions); err != nil {
		return err
	}

	// 创建默认管理员用户
	adminUser := &models.User{
		Username: "admin",
		Email:    "admin@example.com",
		RealName: "系统管理员",
		Status:   true,
	}

	if err := adminUser.SetPassword("admin123"); err != nil {
		return err
	}

	if err := db.FirstOrCreate(adminUser, models.User{Username: "admin"}).Error; err != nil {
		return err
	}

	// 为管理员用户分配管理员角色
	if err := db.Model(adminUser).Association("Roles").Replace([]*models.Role{adminRole}); err != nil {
		return err
	}

	return nil
}
