package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager/internal/auth"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/db/models"
	"github.com/myysophia/ossmanager/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Password string `json:"password" binding:"required,min=6,max=32"`
	Email    string `json:"email" binding:"required,email"`
}

// AuthHandler 认证处理器
type AuthHandler struct {
	*BaseHandler
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler() *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandler(),
	}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.CodeInvalidParams, err)
		return
	}

	var user models.User
	if err := db.GetDB().Where("username = ?", req.Username).First(&user).Error; err != nil {
		utils.ResponseError(c, utils.CodeInternalError, errors.New("用户名或密码错误"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.ResponseError(c, utils.CodeInternalError, errors.New("用户名或密码错误"))
		return
	}

	// 生成 JWT token
	jwtConfig := config.GetConfig().JWT
	token, err := auth.GenerateToken(&user, &jwtConfig)
	if err != nil {
		utils.ResponseError(c, utils.CodeInternalError, errors.New("生成令牌失败"))
		return
	}

	utils.ResponseWithData(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"roles":    user.Roles,
		},
	})
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	// 记录请求头信息
	fmt.Printf("【DEBUG】请求头信息: %+v\n", c.Request.Header)

	// 读取原始请求体并打印
	bodyBytes, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		fmt.Printf("【ERROR】读取请求体失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("读取请求失败"))
		return
	}

	// 打印原始请求体
	fmt.Printf("【DEBUG】原始请求体: %s\n", string(bodyBytes))

	// 直接使用 json.Unmarshal 解析 JSON 数据
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=32"`
		Password string `json:"password" binding:"required,min=6,max=32"`
		Email    string `json:"email" binding:"required,email"`
		RealName string `json:"real_name"`
	}

	// 打印绑定前的结构体
	fmt.Printf("【DEBUG】绑定前的请求结构体: %+v\n", req)

	// 使用自定义绑定方式替代 ShouldBindJSON
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		fmt.Printf("【ERROR】解析JSON失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("参数错误"))
		return
	}

	// 手动验证必填字段
	if req.Username == "" || len(req.Username) < 3 || len(req.Username) > 32 {
		fmt.Printf("【ERROR】用户名验证失败: %s\n", req.Username)
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("用户名长度应为3-32个字符"))
		return
	}
	if req.Password == "" || len(req.Password) < 6 || len(req.Password) > 32 {
		fmt.Printf("【ERROR】密码验证失败\n")
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("密码长度应为6-32个字符"))
		return
	}
	// 简单的邮箱格式验证
	if req.Email == "" || !strings.Contains(req.Email, "@") {
		fmt.Printf("【ERROR】邮箱验证失败: %s\n", req.Email)
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("请输入有效的邮箱地址"))
		return
	}

	// 打印绑定后的结构体
	fmt.Printf("【DEBUG】绑定后的请求结构体: %+v\n", req)

	// 检查用户名是否已存在
	var count int64
	if err := db.GetDB().Model(&models.User{}).Where("username = ?", req.Username).Count(&count).Error; err != nil {
		fmt.Printf("【ERROR】检查用户名失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("检查用户名失败"))
		return
	}
	if count > 0 {
		fmt.Printf("【WARN】用户名已存在: %s\n", req.Username)
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("用户名已存在"))
		return
	}

	// 检查邮箱是否已存在
	if err := db.GetDB().Model(&models.User{}).Where("email = ?", req.Email).Count(&count).Error; err != nil {
		fmt.Printf("【ERROR】检查邮箱失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("检查邮箱失败"))
		return
	}
	if count > 0 {
		fmt.Printf("【WARN】邮箱已存在: %s\n", req.Email)
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("邮箱已存在"))
		return
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		fmt.Printf("【ERROR】密码加密失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("密码加密失败"))
		return
	}

	// 创建用户
	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		RealName: req.RealName,
		Status:   true,
	}

	fmt.Printf("【DEBUG】准备创建用户: %+v\n", user)

	if err := db.GetDB().Create(&user).Error; err != nil {
		fmt.Printf("【ERROR】创建用户失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("创建用户失败"))
		return
	}

	fmt.Printf("【INFO】用户创建成功: ID=%d, Username=%s\n", user.ID, user.Username)

	// 分配默认角色
	var defaultRole models.Role
	if err := db.GetDB().Where("name = ?", "user").First(&defaultRole).Error; err != nil {
		fmt.Printf("【ERROR】获取默认角色失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("获取默认角色失败"))
		return
	}

	if err := db.GetDB().Model(&user).Association("Roles").Append(&defaultRole); err != nil {
		fmt.Printf("【ERROR】分配角色失败: %v\n", err)
		utils.ResponseError(c, utils.CodeInternalError, errors.New("分配角色失败"))
		return
	}

	fmt.Printf("【INFO】角色分配成功: UserID=%d, Role=%s\n", user.ID, defaultRole.Name)

	utils.ResponseWithData(c, gin.H{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"real_name": user.RealName,
	})
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// 从上下文获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ResponseError(c, utils.CodeUnauthorized, errors.New("未登录"))
		return
	}

	var user models.User
	if err := db.GetDB().Preload("Roles").First(&user, userID).Error; err != nil {
		utils.ResponseError(c, utils.CodeNotFound, errors.New("用户不存在"))
		return
	}

	// 获取用户权限
	permissions, err := auth.GetUserPermissions(userID.(uint))
	if err != nil {
		utils.ResponseError(c, utils.CodeInternalError, errors.New("获取用户权限失败"))
		return
	}

	utils.ResponseWithData(c, gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"real_name":   user.RealName,
		"roles":       user.Roles,
		"permissions": permissions,
	})
}

// UpdatePassword 更新密码
func (h *AuthHandler) UpdatePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=6,max=32"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("参数错误"))
		return
	}

	// 从上下文获取用户ID
	userID, exists := c.Get("userID")
	if !exists {
		utils.ResponseError(c, utils.CodeUnauthorized, errors.New("未登录"))
		return
	}

	var user models.User
	if err := db.GetDB().First(&user, userID).Error; err != nil {
		utils.ResponseError(c, utils.CodeNotFound, errors.New("用户不存在"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword)); err != nil {
		utils.ResponseError(c, utils.CodeInvalidParams, errors.New("原密码错误"))
		return
	}

	// 加密新密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		utils.ResponseError(c, utils.CodeInternalError, errors.New("密码加密失败"))
		return
	}

	if err := db.GetDB().Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		utils.ResponseError(c, utils.CodeInternalError, errors.New("更新密码失败"))
		return
	}

	utils.ResponseWithData(c, nil)
}
