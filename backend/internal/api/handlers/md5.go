package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/db"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/function"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"go.uber.org/zap"
	"strconv"
)

// MD5Handler MD5计算处理器
type MD5Handler struct {
	*BaseHandler
	md5Calculator *function.MD5Calculator
}

// NewMD5Handler 创建MD5计算处理器
func NewMD5Handler(md5Calculator *function.MD5Calculator) *MD5Handler {
	return &MD5Handler{
		BaseHandler:   NewBaseHandler(),
		md5Calculator: md5Calculator,
	}
}

// TriggerCalculation 触发MD5计算
func (h *MD5Handler) TriggerCalculation(c *gin.Context) {
	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		h.BadRequest(c, "无效的文件ID")
		return
	}

	// 从数据库获取文件信息
	var file models.OSSFile
	if err := db.GetDB().First(&file, fileID).Error; err != nil {
		h.NotFound(c, "文件不存在")
		return
	}

	// 触发MD5计算
	if err := h.md5Calculator.TriggerCalculation(&file); err != nil {
		logger.Error("触发MD5计算失败",
			zap.Uint("file_id", file.ID),
			zap.String("object_key", file.ObjectKey),
			zap.Error(err))
		h.InternalError(c, "触发MD5计算失败")
		return
	}

	h.Success(c, gin.H{
		"message": "MD5计算已触发，请稍后查询结果",
		"file_id": file.ID,
	})
}

// GetMD5 获取文件的MD5值
func (h *MD5Handler) GetMD5(c *gin.Context) {
	// 获取文件ID
	fileIDStr := c.Param("id")
	fileID, err := strconv.ParseUint(fileIDStr, 10, 32)
	if err != nil {
		h.BadRequest(c, "无效的文件ID")
		return
	}

	// 从数据库获取文件信息
	var file models.OSSFile
	if err := db.GetDB().First(&file, fileID).Error; err != nil {
		h.NotFound(c, "文件不存在")
		return
	}

	if file.MD5 == "" {
		h.Success(c, gin.H{
			"message": "MD5值尚未计算或正在计算中",
			"file_id": file.ID,
			"status":  "pending",
		})
		return
	}

	h.Success(c, gin.H{
		"file_id": file.ID,
		"md5":     file.MD5,
		"status":  "completed",
	})
}
