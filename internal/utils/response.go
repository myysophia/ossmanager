package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager/internal/logger"
	"go.uber.org/zap"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// 定义状态码
const (
	CodeSuccess        = 200 // 成功
	CodeInvalidParams  = 400 // 参数错误
	CodeUnauthorized   = 401 // 未授权
	CodeForbidden      = 403 // 禁止访问
	CodeNotFound       = 404 // 资源不存在
	CodeTooManyRequests = 429 // 请求过多
	CodeInternalError  = 500 // 服务器内部错误

	// OSS相关状态码
	CodeServerError    = 50001 // 服务器错误
	CodeConfigNotFound = 40404 // 配置不存在
	CodeFileNotFound   = 40405 // 文件不存在
	CodeConfigInUse    = 40001 // 配置正在使用中
	CodeFileExists     = 40009 // 文件已存在
)

// 对应的消息
var codeMsgMap = map[int]string{
	CodeSuccess:        "操作成功",
	CodeInvalidParams:  "参数错误",
	CodeUnauthorized:   "未授权",
	CodeForbidden:      "禁止访问",
	CodeNotFound:       "资源不存在",
	CodeTooManyRequests: "请求过于频繁，请稍后再试",
	CodeInternalError:  "服务器内部错误",

	// OSS相关状态码消息
	CodeServerError:    "服务器错误",
	CodeConfigNotFound: "存储配置不存在",
	CodeFileNotFound:   "文件不存在",
	CodeConfigInUse:    "配置正在使用中",
	CodeFileExists:     "文件已存在",
}

// ResponseWithJSON 返回JSON响应
func ResponseWithJSON(c *gin.Context, code int, data interface{}) {
	msg, ok := codeMsgMap[code]
	if !ok {
		msg = "未知错误"
	}

	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
		Data:    data,
	})
}

// ResponseWithData 返回成功响应，包含数据
func ResponseWithData(c *gin.Context, data interface{}) {
	ResponseWithJSON(c, CodeSuccess, data)
}

// ResponseSuccess 返回成功响应，不包含数据
func ResponseSuccess(c *gin.Context) {
	ResponseWithJSON(c, CodeSuccess, nil)
}

// ResponseWithMsg 返回带自定义消息的成功响应
func ResponseWithMsg(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
	})
}

// ResponseWithMsgAndData 返回带自定义消息和数据的成功响应
func ResponseWithMsgAndData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code:    CodeSuccess,
		Message: message,
		Data:    data,
	})
}

// ResponseError 返回错误响应
func ResponseError(c *gin.Context, code int, err error) {
	msg, ok := codeMsgMap[code]
	if !ok {
		msg = "未知错误"
	}

	// 如果提供了错误信息，则使用错误信息
	if err != nil {
		msg = err.Error()
	}

	// 记录错误日志
	logger.Error("API错误响应",
		zap.Int("code", code),
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("message", msg))

	c.JSON(http.StatusOK, Response{
		Code:    code,
		Message: msg,
	})
}

// ResponseBadRequest 返回参数错误响应
func ResponseBadRequest(c *gin.Context, err error) {
	ResponseError(c, CodeInvalidParams, err)
}

// ResponseUnauthorized 返回未授权响应
func ResponseUnauthorized(c *gin.Context, err error) {
	ResponseError(c, CodeUnauthorized, err)
}

// ResponseForbidden 返回禁止访问响应
func ResponseForbidden(c *gin.Context, err error) {
	ResponseError(c, CodeForbidden, err)
}

// ResponseNotFound 返回资源不存在响应
func ResponseNotFound(c *gin.Context, err error) {
	ResponseError(c, CodeNotFound, err)
}

// ResponseInternalError 返回服务器内部错误响应
func ResponseInternalError(c *gin.Context, err error) {
	ResponseError(c, CodeInternalError, err)
}

// GetUserID 从上下文中获取用户ID
func GetUserID(c *gin.Context) uint {
	// 从上下文中获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		// 尝试从 userID 获取
		userID, exists = c.Get("userID")
		if !exists {
			return 1 // 默认返回ID为1的管理员用户，避免外键约束错误
		}
	}

	// 尝试转换为uint
	switch v := userID.(type) {
	case uint:
		return v
	case float64:
		return uint(v)
	case int:
		return uint(v)
	case int64:
		return uint(v)
	default:
		return 0
	}
}

// SuccessResponse 返回成功响应（新版本格式）
func SuccessResponse(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": message,
		"data":    data,
	})
}

// ErrorResponse 返回错误响应（新版本格式）
func ErrorResponse(c *gin.Context, httpCode int, message string, err error) {
	errorMsg := message
	if err != nil {
		errorMsg = err.Error()
	}

	// 记录错误日志
	logger.Error("API错误响应",
		zap.String("path", c.Request.URL.Path),
		zap.String("method", c.Request.Method),
		zap.String("message", errorMsg),
		zap.Error(err))

	c.JSON(httpCode, gin.H{
		"success": false,
		"message": errorMsg,
	})
}
