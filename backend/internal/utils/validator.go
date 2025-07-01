package utils

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/myysophia/ossmanager-backend/internal/logger"
	"go.uber.org/zap"
	"reflect"
	"strings"
)

// 全局验证器
var (
	validate *validator.Validate
	trans    ut.Translator
)

// InitValidator 初始化验证器
func InitValidator() {
	// 创建验证器
	validate = validator.New()

	// 注册自定义标签名称
	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return fld.Name
		}
		return name
	})

	// 创建中文翻译器
	zhTrans := zh.New()
	uni := ut.New(zhTrans, zhTrans)
	trans, _ = uni.GetTranslator("zh")

	// 注册中文翻译
	err := zh_translations.RegisterDefaultTranslations(validate, trans)
	if err != nil {
		logger.Error("注册验证器翻译失败", zap.Error(err))
		return
	}

	// 注册自定义验证器
	registerCustomValidators()
}

// registerCustomValidators 注册自定义验证器
func registerCustomValidators() {
	// 示例：添加一个OSS存储类型验证器
	_ = validate.RegisterValidation("storage_type", func(fl validator.FieldLevel) bool {
		value := fl.Field().String()
		return value == "ALIYUN_OSS" || value == "AWS_S3" || value == "CLOUDFLARE_R2"
	})
}

// BindAndValidate 绑定并验证请求数据
func BindAndValidate(c *gin.Context, obj interface{}) error {
	// 根据请求类型选择绑定方法
	var err error
	switch c.Request.Method {
	case "GET":
		err = c.ShouldBindQuery(obj)
	case "POST", "PUT", "PATCH":
		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "application/json") {
			err = c.ShouldBindJSON(obj)
		} else if strings.Contains(contentType, "multipart/form-data") {
			err = c.ShouldBindWith(obj, binding.FormMultipart)
		} else {
			err = c.ShouldBind(obj)
		}
	default:
		err = c.ShouldBind(obj)
	}

	// 处理绑定错误
	if err != nil {
		logger.Warn("请求数据绑定失败",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err))
		return err
	}

	// 验证
	err = validate.Struct(obj)
	if err != nil {
		logger.Warn("数据验证失败",
			zap.String("path", c.Request.URL.Path),
			zap.Error(err))

		// 如果是验证错误，翻译错误信息
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errMsgs := []string{}
			for _, e := range validationErrors {
				errMsgs = append(errMsgs, e.Translate(trans))
			}
			return errors.New(strings.Join(errMsgs, "; "))
		}

		return err
	}

	return nil
}
