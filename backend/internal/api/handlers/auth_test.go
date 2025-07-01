package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/tests/mocks"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHandler_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewAuthHandler(mockDB)

	// 设置测试数据
	user := &models.User{
		ID:       1,
		Username: "test",
		Password: "$2a$10$test", // 加密后的密码
		Role:     "user",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().Where(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置请求体
	body := map[string]interface{}{
		"username": "test",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewBuffer(jsonBody))

	// 执行测试
	handler.Login(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestAuthHandler_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewAuthHandler(mockDB)

	// 设置测试数据
	user := &models.User{
		ID:       1,
		Username: "newuser",
		Password: "$2a$10$test", // 加密后的密码
		Role:     "user",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().Where(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)
	mockDB.EXPECT().Create(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 设置请求体
	body := map[string]interface{}{
		"username": "newuser",
		"password": "password123",
		"email":    "newuser@example.com",
	}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/register", bytes.NewBuffer(jsonBody))

	// 执行测试
	handler.Register(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestAuthHandler_GetUserInfo(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewAuthHandler(mockDB)

	// 设置测试数据
	user := &models.User{
		ID:       1,
		Username: "test",
		Email:    "test@example.com",
		Role:     "user",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))

	// 执行测试
	handler.GetUserInfo(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestAuthHandler_UpdatePassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewAuthHandler(mockDB)

	// 设置测试数据
	user := &models.User{
		ID:       1,
		Username: "test",
		Password: "$2a$10$test", // 加密后的密码
		Role:     "user",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)
	mockDB.EXPECT().Save(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))

	// 设置请求体
	body := map[string]interface{}{
		"old_password": "password123",
		"new_password": "newpassword123",
	}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("PUT", "/api/v1/auth/password", bytes.NewBuffer(jsonBody))

	// 执行测试
	handler.UpdatePassword(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}
