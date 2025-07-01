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

// RegionBucketHandler 地域-桶映射处理器
type RegionBucketHandler struct {
	BaseHandler
	DB *gorm.DB
}

// NewRegionBucketHandler 创建地域-桶映射处理器
func NewRegionBucketHandler(db *gorm.DB) *RegionBucketHandler {
	return &RegionBucketHandler{
		BaseHandler: BaseHandler{},
		DB:          db,
	}
}

// GetRegionList 获取地域列表
func (h *RegionBucketHandler) GetRegionList(c *gin.Context) {
	var regions []string
	if err := h.DB.Model(&models.RegionBucketMapping{}).
		Distinct().
		Pluck("region_code", &regions).
		Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取地域列表失败")
		return
	}

	h.Success(c, regions)
}

// GetBucketList 获取指定地域下的桶列表
func (h *RegionBucketHandler) GetBucketList(c *gin.Context) {
	regionCode := c.Query("region_code")
	if regionCode == "" {
		h.Error(c, utils.CodeInvalidParams, "地域代码不能为空")
		return
	}

	var buckets []string
	if err := h.DB.Model(&models.RegionBucketMapping{}).
		Where("region_code = ?", regionCode).
		Pluck("bucket_name", &buckets).
		Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取桶列表失败")
		return
	}

	h.Success(c, buckets)
}

// GetUserAccessibleBuckets 获取用户可访问的桶列表
func (h *RegionBucketHandler) GetUserAccessibleBuckets(c *gin.Context) {
	userID := c.GetUint("user_id")
	regionCode := c.Query("region_code")

	query := h.DB.Model(&models.RegionBucketMapping{}).
		Joins("JOIN role_region_bucket_access ON role_region_bucket_access.region_bucket_mapping_id = region_bucket_mapping.id").
		Joins("JOIN user_roles ON user_roles.role_id = role_region_bucket_access.role_id").
		Where("user_roles.user_id = ?", userID)

	if regionCode != "" {
		query = query.Where("region_bucket_mapping.region_code = ?", regionCode)
	}

	var buckets []string
	if err := query.Distinct().
		Pluck("region_bucket_mapping.bucket_name", &buckets).
		Error; err != nil {
		h.Error(c, utils.CodeServerError, "获取可访问桶列表失败")
		return
	}

	h.Success(c, buckets)
}

// List 获取地域-桶映射列表
func (h *RegionBucketHandler) List(c *gin.Context) {
	// 修复分页参数解析
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 1 {
		pageSize = 10
	}

	// 修复查询参数名称
	regionCode := c.Query("region") // 改为 region
	bucketName := c.Query("bucket") // 改为 bucket

	logger.Info("开始获取地域-桶映射列表",
		zap.Int("page", page),
		zap.Int("page_size", pageSize),
		zap.String("region_code", regionCode),
		zap.String("bucket_name", bucketName))

	query := h.DB.Model(&models.RegionBucketMapping{})

	if regionCode != "" {
		query = query.Where("region_code = ?", regionCode)
		logger.Info("添加地域代码筛选条件", zap.String("region_code", regionCode))
	}
	if bucketName != "" {
		query = query.Where("bucket_name = ?", bucketName)
		logger.Info("添加桶名称筛选条件", zap.String("bucket_name", bucketName))
	}

	// 添加调试日志，打印实际执行的 SQL
	sql := query.ToSQL(func(tx *gorm.DB) *gorm.DB {
		return tx.Model(&models.RegionBucketMapping{})
	})
	logger.Info("执行的SQL查询", zap.String("sql", sql))

	var total int64
	if err := query.Count(&total).Error; err != nil {
		logger.Error("获取总数失败", zap.Error(err))
		h.InternalError(c, "获取总数失败")
		return
	}
	logger.Info("查询到总记录数", zap.Int64("total", total))

	var mappings []models.RegionBucketMapping
	if err := query.Offset((page - 1) * pageSize).Limit(pageSize).Find(&mappings).Error; err != nil {
		logger.Error("获取列表失败", zap.Error(err))
		h.InternalError(c, "获取列表失败")
		return
	}
	logger.Info("查询结果",
		zap.Int("offset", (page-1)*pageSize),
		zap.Int("limit", pageSize),
		zap.Int("result_count", len(mappings)))

	// 打印每条记录的详细信息
	for i, mapping := range mappings {
		logger.Info("记录详情",
			zap.Int("index", i),
			zap.Uint("id", mapping.ID),
			zap.String("region_code", mapping.RegionCode),
			zap.String("bucket_name", mapping.BucketName),
			zap.Time("created_at", mapping.CreatedAt),
			zap.Time("updated_at", mapping.UpdatedAt))
	}

	h.Success(c, gin.H{
		"total": total,
		"items": mappings,
	})
}

// Create 创建地域-桶映射
func (h *RegionBucketHandler) Create(c *gin.Context) {
	var input models.RegionBucketMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查是否已存在相同的映射
	var count int64
	if err := h.DB.Model(&models.RegionBucketMapping{}).
		Where("region_code = ? AND bucket_name = ?", input.RegionCode, input.BucketName).
		Count(&count).Error; err != nil {
		h.InternalError(c, "检查映射是否存在失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "该地域-桶映射已存在")
		return
	}

	if err := h.DB.Create(&input).Error; err != nil {
		h.InternalError(c, "创建失败")
		return
	}

	h.Success(c, input)
}

// Get 获取地域-桶映射详情
func (h *RegionBucketHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var mapping models.RegionBucketMapping
	if err := h.DB.First(&mapping, id).Error; err != nil {
		h.NotFound(c, "映射不存在")
		return
	}

	h.Success(c, mapping)
}

// Update 更新地域-桶映射
func (h *RegionBucketHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var mapping models.RegionBucketMapping
	if err := h.DB.First(&mapping, id).Error; err != nil {
		h.NotFound(c, "映射不存在")
		return
	}

	var input models.RegionBucketMapping
	if err := c.ShouldBindJSON(&input); err != nil {
		h.BadRequest(c, "参数错误")
		return
	}

	// 检查是否已存在相同的映射（排除当前记录）
	var count int64
	if err := h.DB.Model(&models.RegionBucketMapping{}).
		Where("region_code = ? AND bucket_name = ? AND id != ?", input.RegionCode, input.BucketName, id).
		Count(&count).Error; err != nil {
		h.InternalError(c, "检查映射是否存在失败")
		return
	}
	if count > 0 {
		h.Error(c, utils.CodeInvalidParams, "该地域-桶映射已存在")
		return
	}

	mapping.RegionCode = input.RegionCode
	mapping.BucketName = input.BucketName

	if err := h.DB.Save(&mapping).Error; err != nil {
		h.InternalError(c, "更新失败")
		return
	}

	h.Success(c, mapping)
}

// Delete 删除地域-桶映射
func (h *RegionBucketHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	var mapping models.RegionBucketMapping
	if err := h.DB.First(&mapping, id).Error; err != nil {
		h.NotFound(c, "映射不存在")
		return
	}

	// 开启事务
	tx := h.DB.Begin()

	// 删除相关的角色访问权限
	if err := tx.Where("region_bucket_mapping_id = ?", id).Delete(&models.RoleRegionBucketAccess{}).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "删除相关访问权限失败")
		return
	}

	// 删除映射
	if err := tx.Delete(&mapping).Error; err != nil {
		tx.Rollback()
		h.InternalError(c, "删除失败")
		return
	}

	if err := tx.Commit().Error; err != nil {
		h.InternalError(c, "提交事务失败")
		return
	}

	h.Success(c, nil)
}
