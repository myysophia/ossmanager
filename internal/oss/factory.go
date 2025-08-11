package oss

import (
	"fmt"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
	"sync"
)

// DefaultStorageFactory 默认存储服务工厂
type DefaultStorageFactory struct {
	ossConfig     *config.OSSConfig
	serviceCache  map[string]StorageService
	lock          sync.RWMutex
	defaultConfig *models.OSSConfig
}

// NewStorageFactory 创建存储服务工厂
func NewStorageFactory(ossConfig *config.OSSConfig) *DefaultStorageFactory {
	return &DefaultStorageFactory{
		ossConfig:    ossConfig,
		serviceCache: make(map[string]StorageService),
	}
}

// GetStorageService 获取存储服务
func (f *DefaultStorageFactory) GetStorageService(storageType string) (StorageService, error) {
	// 先从缓存中获取
	f.lock.RLock()
	service, ok := f.serviceCache[storageType]
	f.lock.RUnlock()
	if ok {
		return service, nil
	}

	// 缓存中没有，创建新的服务
	f.lock.Lock()
	defer f.lock.Unlock()

	// 再次检查，防止在获取锁的过程中被其他协程创建
	service, ok = f.serviceCache[storageType]
	if ok {
		return service, nil
	}

	// 创建存储服务
	var err error
	switch storageType {
	case StorageTypeAliyunOSS:
		service, err = NewAliyunOSSService(&f.ossConfig.AliyunOSS)
	//case StorageTypeAWSS3:
	//	service, err = NewAWSS3Service(&f.ossConfig.AWSS3)
	//case StorageTypeR2:
	//	service, err = NewCloudflareR2Service(&f.ossConfig.CloudflareR2)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", storageType)
	}

	if err != nil {
		logger.Error("创建存储服务失败", zap.String("storageType", storageType), zap.Error(err))
		return nil, err
	}

	// 加入缓存
	f.serviceCache[storageType] = service
	return service, nil
}

// GetDefaultStorageService 获取默认存储服务
func (f *DefaultStorageFactory) GetDefaultStorageService() (StorageService, error) {
	// 如果已有默认配置，直接使用
	if f.defaultConfig != nil {
		return f.GetStorageService(f.defaultConfig.StorageType)
	}

	// 从数据库中获取默认配置
	gormDB := db.GetDB()
	var ossConfig models.OSSConfig
	err := gormDB.Where("is_default = ?", true).First(&ossConfig).Error
	if err != nil {
		logger.Error("从数据库获取默认OSS配置失败", zap.Error(err))

		// 降级为使用配置文件中的阿里云OSS
		logger.Info("降级为使用配置文件中的阿里云OSS作为默认存储")
		return f.GetStorageService(StorageTypeAliyunOSS)
	}

	f.defaultConfig = &ossConfig

	// 根据配置创建存储服务
	var service StorageService
	switch ossConfig.StorageType {
	case StorageTypeAliyunOSS:
		service, err = f.GetStorageService(StorageTypeAliyunOSS)
	case StorageTypeAWSS3:
		service, err = f.GetStorageService(StorageTypeAWSS3)
	case StorageTypeR2:
		service, err = f.GetStorageService(StorageTypeR2)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", ossConfig.StorageType)
	}

	if err != nil {
		logger.Error("创建默认存储服务失败", zap.String("storageType", ossConfig.StorageType), zap.Error(err))
		return nil, err
	}

	return service, nil
}

// ClearCache 清除缓存
func (f *DefaultStorageFactory) ClearCache() {
	f.lock.Lock()
	defer f.lock.Unlock()
	f.serviceCache = make(map[string]StorageService)
	f.defaultConfig = nil
}
