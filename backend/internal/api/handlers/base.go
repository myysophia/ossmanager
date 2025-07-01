package handlers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/utils"
)

// BaseHandler 基础处理器
type BaseHandler struct {
	// 可以添加一些共享的依赖，比如数据库连接、服务实例等
}

// NewBaseHandler 创建基础处理器
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{}
}

// Success 成功响应
func (h *BaseHandler) Success(c *gin.Context, data interface{}) {
	utils.ResponseWithData(c, data)
}

// Error 错误响应
func (h *BaseHandler) Error(c *gin.Context, code int, message string) {
	utils.ResponseError(c, code, errors.New(message))
}

// BadRequest 请求参数错误
func (h *BaseHandler) BadRequest(c *gin.Context, message string) {
	utils.ResponseError(c, utils.CodeInvalidParams, errors.New(message))
}

// Unauthorized 未授权
func (h *BaseHandler) Unauthorized(c *gin.Context, message string) {
	utils.ResponseError(c, utils.CodeUnauthorized, errors.New(message))
}

// Forbidden 禁止访问
func (h *BaseHandler) Forbidden(c *gin.Context, message string) {
	utils.ResponseError(c, utils.CodeForbidden, errors.New(message))
}

// NotFound 资源不存在
func (h *BaseHandler) NotFound(c *gin.Context, message string) {
	utils.ResponseError(c, utils.CodeNotFound, errors.New(message))
}

// InternalError 内部错误
func (h *BaseHandler) InternalError(c *gin.Context, message string) {
	utils.ResponseError(c, utils.CodeInternalError, errors.New(message))
}
