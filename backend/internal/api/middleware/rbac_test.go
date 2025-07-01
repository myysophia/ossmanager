package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestPermissionMiddleware(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	r := gin.New()
	r.Use(PermissionMiddleware("test:permission"))

	// 创建测试处理器
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试有权限
	t.Run("Has Permission", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Set("permissions", []string{"test:permission"})
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		PermissionMiddleware("test:permission")(c)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试无权限
	t.Run("No Permission", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Set("permissions", []string{})
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		PermissionMiddleware("test:permission")(c)

		// 验证响应
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestRoleMiddleware(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	r := gin.New()
	r.Use(RoleMiddleware("admin"))

	// 创建测试处理器
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试有角色
	t.Run("Has Role", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Set("roles", []string{"admin"})
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		RoleMiddleware("admin")(c)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试无角色
	t.Run("No Role", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Set("roles", []string{})
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		RoleMiddleware("admin")(c)

		// 验证响应
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestAdminMiddleware(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建测试路由
	r := gin.New()
	r.Use(AdminMiddleware())

	// 创建测试处理器
	r.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试管理员
	t.Run("Is Admin", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Set("roles", []string{"admin"})
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		AdminMiddleware()(c)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试非管理员
	t.Run("Not Admin", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Set("roles", []string{"user"})
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		AdminMiddleware()(c)

		// 验证响应
		assert.Equal(t, http.StatusForbidden, w.Code)
	})
} 