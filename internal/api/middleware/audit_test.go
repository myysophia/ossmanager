package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/tests/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuditLogMiddleware(t *testing.T) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 创建控制器
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 创建模拟数据库
	mockDB := mocks.NewMockDB(ctrl)

	// 创建测试路由
	r := gin.New()
	r.Use(AuditLogMiddleware(mockDB))

	// 创建测试处理器
	r.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 测试记录审计日志
	t.Run("Record Audit Log", func(t *testing.T) {
		// 设置 Mock 期望
		mockDB.EXPECT().Create(gomock.Any()).Return(mockDB)
		mockDB.EXPECT().Error().Return(nil)

		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Request = httptest.NewRequest("POST", "/test", nil)
		c.Request.Header.Set("User-Agent", "test-agent")
		c.Request.RemoteAddr = "127.0.0.1"

		// 执行中间件
		AuditLogMiddleware(mockDB)(c)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
	})

	// 测试不记录审计日志
	t.Run("Skip Audit Log", func(t *testing.T) {
		// 创建测试请求
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("userID", uint(1))
		c.Set("username", "test")
		c.Request = httptest.NewRequest("GET", "/test", nil)

		// 执行中间件
		AuditLogMiddleware(mockDB)(c)

		// 验证响应
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestGetResourceType(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "OSS File",
			path:     "/api/v1/files",
			expected: "oss_file",
		},
		{
			name:     "OSS Config",
			path:     "/api/v1/configs",
			expected: "oss_config",
		},
		{
			name:     "User",
			path:     "/api/v1/users",
			expected: "user",
		},
		{
			name:     "Unknown",
			path:     "/api/v1/unknown",
			expected: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getResourceType(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetResourceID(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{
			name:     "With ID",
			path:     "/api/v1/files/1",
			expected: "1",
		},
		{
			name:     "With Operation",
			path:     "/api/v1/files/1/download",
			expected: "1",
		},
		{
			name:     "No ID",
			path:     "/api/v1/files",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getResourceID(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}
