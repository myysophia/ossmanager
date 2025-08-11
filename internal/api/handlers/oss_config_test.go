package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/myysophia/ossmanager-backend/internal/tests/mocks"
	"github.com/myysophia/ossmanager-backend/internal/utils"
	"github.com/stretchr/testify/assert"
)

func TestOSSConfigHandler_Create(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewOSSConfigHandler(mockDB)

	// 设置测试数据
	config := &models.OSSConfig{
		ID:          1,
		Name:        "test-config",
		Type:        "aws_s3",
		AccessKey:   "test-key",
		SecretKey:   "test-secret",
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		Endpoint:    "https://s3.amazonaws.com",
		CreatorID:   1,
		CreatorName: "test",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().Create(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))
	c.Set("username", "test")

	// 设置请求体
	body := map[string]interface{}{
		"name":       "test-config",
		"type":       "aws_s3",
		"access_key": "test-key",
		"secret_key": "test-secret",
		"bucket":     "test-bucket",
		"region":     "us-east-1",
		"endpoint":   "https://s3.amazonaws.com",
	}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("POST", "/api/v1/configs", bytes.NewBuffer(jsonBody))

	// 执行测试
	handler.Create(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOSSConfigHandler_Update(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewOSSConfigHandler(mockDB)

	// 设置测试数据
	config := &models.OSSConfig{
		ID:          1,
		Name:        "test-config",
		Type:        "aws_s3",
		AccessKey:   "test-key",
		SecretKey:   "test-secret",
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		Endpoint:    "https://s3.amazonaws.com",
		CreatorID:   1,
		CreatorName: "test",
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
	c.Set("username", "test")

	// 设置请求参数
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// 设置请求体
	body := map[string]interface{}{
		"name":       "updated-config",
		"type":       "aws_s3",
		"access_key": "updated-key",
		"secret_key": "updated-secret",
		"bucket":     "updated-bucket",
		"region":     "us-west-2",
		"endpoint":   "https://s3.us-west-2.amazonaws.com",
	}
	jsonBody, _ := json.Marshal(body)
	c.Request = httptest.NewRequest("PUT", "/api/v1/configs/1", bytes.NewBuffer(jsonBody))

	// 执行测试
	handler.Update(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOSSConfigHandler_Delete(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewOSSConfigHandler(mockDB)

	// 设置测试数据
	config := &models.OSSConfig{
		ID:          1,
		Name:        "test-config",
		Type:        "aws_s3",
		AccessKey:   "test-key",
		SecretKey:   "test-secret",
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		Endpoint:    "https://s3.amazonaws.com",
		CreatorID:   1,
		CreatorName: "test",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)
	mockDB.EXPECT().Delete(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))
	c.Set("username", "test")

	// 设置请求参数
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// 执行测试
	handler.Delete(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOSSConfigHandler_List(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewOSSConfigHandler(mockDB)

	// 设置测试数据
	configs := []models.OSSConfig{
		{
			ID:          1,
			Name:        "test-config-1",
			Type:        "aws_s3",
			AccessKey:   "test-key-1",
			SecretKey:   "test-secret-1",
			Bucket:      "test-bucket-1",
			Region:      "us-east-1",
			Endpoint:    "https://s3.amazonaws.com",
			CreatorID:   1,
			CreatorName: "test",
		},
		{
			ID:          2,
			Name:        "test-config-2",
			Type:        "aws_s3",
			AccessKey:   "test-key-2",
			SecretKey:   "test-secret-2",
			Bucket:      "test-bucket-2",
			Region:      "us-west-2",
			Endpoint:    "https://s3.us-west-2.amazonaws.com",
			CreatorID:   1,
			CreatorName: "test",
		},
	}

	// 设置 Mock 期望
	mockDB.EXPECT().Offset(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Limit(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Find(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)
	mockDB.EXPECT().Model(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Count(gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))
	c.Set("username", "test")

	// 设置查询参数
	c.Request = httptest.NewRequest("GET", "/api/v1/configs?page=1&page_size=10", nil)

	// 执行测试
	handler.List(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOSSConfigHandler_Get(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewOSSConfigHandler(mockDB)

	// 设置测试数据
	config := &models.OSSConfig{
		ID:          1,
		Name:        "test-config",
		Type:        "aws_s3",
		AccessKey:   "test-key",
		SecretKey:   "test-secret",
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		Endpoint:    "https://s3.amazonaws.com",
		CreatorID:   1,
		CreatorName: "test",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))
	c.Set("username", "test")

	// 设置请求参数
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// 执行测试
	handler.Get(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}

func TestOSSConfigHandler_Test(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockDB(ctrl)

	handler := NewOSSConfigHandler(mockDB)

	// 设置测试数据
	config := &models.OSSConfig{
		ID:          1,
		Name:        "test-config",
		Type:        "aws_s3",
		AccessKey:   "test-key",
		SecretKey:   "test-secret",
		Bucket:      "test-bucket",
		Region:      "us-east-1",
		Endpoint:    "https://s3.amazonaws.com",
		CreatorID:   1,
		CreatorName: "test",
	}

	// 设置 Mock 期望
	mockDB.EXPECT().First(gomock.Any(), gomock.Any()).Return(mockDB)
	mockDB.EXPECT().Error().Return(nil)

	// 创建测试请求
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("userID", uint(1))
	c.Set("username", "test")

	// 设置请求参数
	c.Params = []gin.Param{{Key: "id", Value: "1"}}

	// 执行测试
	handler.Test(c)

	// 验证响应
	assert.Equal(t, http.StatusOK, w.Code)

	var resp utils.Response
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.True(t, resp.Success)
}
