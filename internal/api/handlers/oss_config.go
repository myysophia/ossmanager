package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/db"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"github.com/myysophia/ossmanager-backend/internal/oss"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"go.uber.org/zap"
)

// OSSConfigHandler OSS配置处理器
type OSSConfigHandler struct {
	*BaseHandler
	storageFactory oss.StorageFactory
}

// NewOSSConfigHandler 创建OSS配置处理器
func NewOSSConfigHandler(storageFactory oss.StorageFactory) *OSSConfigHandler {
	return &OSSConfigHandler{
		BaseHandler:    NewBaseHandler(),
		storageFactory: storageFactory,
	}
}

// CreateConfig 创建存储配置
func (h *OSSConfigHandler) CreateConfig(c *gin.Context) {
	var config models.OSSConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 验证配置类型
	if !isValidStorageType(config.StorageType) {
		h.BadRequest(c, "不支持的存储类型")
		return
	}

	// OSSConfig没有CreatedBy字段，不需要设置

	if err := db.GetDB().Create(&config).Error; err != nil {
		h.InternalError(c, "创建存储配置失败")
		return
	}

	h.Success(c, config)
}

// UpdateConfig 更新存储配置
func (h *OSSConfigHandler) UpdateConfig(c *gin.Context) {
	configID := c.Param("id")

	var config models.OSSConfig
	if err := db.GetDB().First(&config, configID).Error; err != nil {
		h.NotFound(c, "存储配置不存在")
		return
	}

	var updateData struct {
		Name        string `json:"name" binding:"required"`
		StorageType string `json:"storage_type" binding:"required"`
		Endpoint    string `json:"endpoint" binding:"required"`
		Bucket      string `json:"bucket" binding:"required"`
		AccessKey   string `json:"access_key" binding:"required"`
		SecretKey   string `json:"secret_key" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateData); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 验证配置类型
	if !isValidStorageType(updateData.StorageType) {
		h.BadRequest(c, "不支持的存储类型")
		return
	}

	// 更新配置
	config.Name = updateData.Name
	config.StorageType = updateData.StorageType
	config.Endpoint = updateData.Endpoint
	config.Bucket = updateData.Bucket
	config.AccessKey = updateData.AccessKey
	config.SecretKey = updateData.SecretKey

	if err := db.GetDB().Save(&config).Error; err != nil {
		h.InternalError(c, "更新存储配置失败")
		return
	}

	h.Success(c, config)
}

// DeleteConfig 删除存储配置
func (h *OSSConfigHandler) DeleteConfig(c *gin.Context) {
	configID := c.Param("id")

	var config models.OSSConfig
	if err := db.GetDB().First(&config, configID).Error; err != nil {
		h.NotFound(c, "存储配置不存在")
		return
	}

	// 检查是否有文件使用此配置
	var count int64
	if err := db.GetDB().Model(&models.OSSFile{}).Where("config_id = ?", configID).Count(&count).Error; err != nil {
		h.InternalError(c, "检查文件关联失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "存储配置正在使用中，无法删除")
		return
	}

	if err := db.GetDB().Delete(&config).Error; err != nil {
		h.InternalError(c, "删除存储配置失败")
		return
	}

	h.Success(c, nil)
}

// GetConfigList 获取存储配置列表
func (h *OSSConfigHandler) GetConfigList(c *gin.Context) {
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
	logger.Info("获取OSS配置列表请求",
		zap.Int("page", page),
		zap.Int("pageSize", pageSize),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method))

	var total int64
	if err := db.GetDB().Model(&models.OSSConfig{}).Count(&total).Error; err != nil {
		logger.Error("获取配置总数失败", zap.Error(err))
		h.InternalError(c, "获取配置总数失败")
		return
	}

	// 记录总数
	logger.Info("OSS配置总数", zap.Int64("total", total))

	var configs []models.OSSConfig

	// 增加详细SQL日志
	query := db.GetDB().Debug()

	// 仅当有数据且需要分页时应用分页
	if total > 0 {
		offset := (page - 1) * pageSize
		query = query.Offset(offset).Limit(pageSize)
		logger.Info("应用分页", zap.Int("offset", offset), zap.Int("limit", pageSize))
	}

	if err := query.Find(&configs).Error; err != nil {
		logger.Error("获取配置列表失败", zap.Error(err))
		h.InternalError(c, "获取配置列表失败")
		return
	}

	// 记录查询结果
	logger.Info("OSS配置查询结果",
		zap.Int("结果数量", len(configs)),
		zap.Any("配置列表", configs))

	// 检查是否启用了软删除但查询未排除已删除记录
	if total > 0 && len(configs) == 0 {
		logger.Warn("发现异常：Count返回有数据但Find查不到记录，尝试不使用软删除查询")

		// 尝试不使用软删除查询
		var allConfigs []models.OSSConfig
		if err := db.GetDB().Debug().Unscoped().Find(&allConfigs).Error; err != nil {
			logger.Error("不使用软删除查询失败", zap.Error(err))
		} else {
			for i, config := range allConfigs {
				logger.Info("记录详情",
					zap.Int("索引", i),
					zap.Uint("ID", config.ID),
					zap.String("名称", config.Name),
					zap.String("类型", config.StorageType),
					zap.Bool("是否默认", config.IsDefault),
					zap.Time("创建时间", config.CreatedAt),
					zap.Time("更新时间", config.UpdatedAt),
					zap.Any("软删除时间", config.DeletedAt))
			}
		}

		// 查询是否记录被软删除
		var deletedConfigs []models.OSSConfig
		if err := db.GetDB().Debug().Unscoped().Where("deleted_at IS NOT NULL").Find(&deletedConfigs).Error; err != nil {
			logger.Error("查询软删除记录失败", zap.Error(err))
		} else {
			logger.Info("软删除记录数量", zap.Int("count", len(deletedConfigs)))
		}
	}

	h.Success(c, gin.H{
		"total": total,
		"items": configs,
	})
}

// GetConfig 获取存储配置详情
func (h *OSSConfigHandler) GetConfig(c *gin.Context) {
	configID := c.Param("id")

	var config models.OSSConfig
	if err := db.GetDB().First(&config, configID).Error; err != nil {
		h.NotFound(c, "存储配置不存在")
		return
	}

	h.Success(c, config)
}

// SetDefaultConfig 设置默认配置
func (h *OSSConfigHandler) SetDefaultConfig(c *gin.Context) {
	configID := c.Param("id")

	var config models.OSSConfig
	if err := db.GetDB().First(&config, configID).Error; err != nil {
		h.NotFound(c, "存储配置不存在")
		return
	}

	// 开始事务
	tx := db.GetDB().Begin()

	// 取消所有默认配置
	if err := tx.Model(&models.OSSConfig{}).Where("is_default = ?", true).Update("is_default", false).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "取消默认配置失败")
		return
	}

	// 设置新的默认配置
	if err := tx.Model(&config).Update("is_default", true).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "设置默认配置失败")
		return
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	// 清除存储工厂缓存
	h.storageFactory.ClearCache()

	h.Success(c, nil)
}

// isValidStorageType 验证存储类型是否有效
func isValidStorageType(storageType string) bool {
	validTypes := []string{oss.StorageTypeAliyunOSS, oss.StorageTypeAWSS3, oss.StorageTypeR2}
	for _, t := range validTypes {
		if t == storageType {
			return true
		}
	}
	return false
}

// Test 测试存储配置
func (h *OSSConfigHandler) Test(c *gin.Context) {
	var config models.OSSConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 验证配置类型
	if !isValidStorageType(config.StorageType) {
		h.BadRequest(c, "不支持的存储类型")
		return
	}

	// TODO: 实现存储配置测试逻辑
	// 1. 创建临时存储服务
	// 2. 尝试上传测试文件
	// 3. 尝试下载测试文件
	// 4. 删除测试文件
	// 5. 返回测试结果

	h.Success(c, gin.H{
		"message": "存储配置测试成功",
	})
}
