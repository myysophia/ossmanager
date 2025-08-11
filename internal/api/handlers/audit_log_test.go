package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/db"
	"github.com/myysophia/ossmanager-backend/internal/db/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestListAuditLogs(t *testing.T) {
	// 设置Gin为测试模式
	gin.SetMode(gin.TestMode)

	// 创建模拟的SQL数据库连接
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a stub database connection", err)
	}
	defer mockDB.Close()

	// 使用GORM连接模拟数据库
	dialector := postgres.New(postgres.Config{
		Conn:       mockDB,
		DriverName: "postgres",
	})
	gdb, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("An error '%s' was not expected when opening a gorm database", err)
	}

	// 替换全局DB实例
	originalDB := db.GetDB()
	db.SetDB(gdb)
	defer db.SetDB(originalDB)

	// 模拟查询计数的SQL
	countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
	mock.ExpectQuery(`SELECT count(.+) FROM "audit_logs"`).WillReturnRows(countRows)

	// 模拟查询结果的SQL
	now := time.Now()
	logRows := sqlmock.NewRows([]string{"id", "user_id", "username", "action", "resource_type", "resource_id", "details", "ip_address", "user_agent", "status", "created_at", "updated_at"}).
		AddRow(1, 1, "admin", "GET", "FILE", "1", `{"path":"/api/v1/oss/files/1","method":"GET"}`, "127.0.0.1", "test-agent", "SUCCESS", now, now).
		AddRow(2, 1, "admin", "POST", "FILE", "2", `{"path":"/api/v1/oss/files","method":"POST"}`, "127.0.0.1", "test-agent", "SUCCESS", now, now)

	mock.ExpectQuery(`SELECT (.+) FROM "audit_logs"`).WillReturnRows(logRows)

	// 创建测试上下文
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/api/v1/audit/logs?page=1&page_size=10", nil)
	c.Set("userID", uint(1))
	c.Set("username", "admin")

	// 创建处理器并调用方法
	handler := NewAuditLogHandler()
	handler.ListAuditLogs(c)

	// 断言
	assert.Equal(t, http.StatusOK, w.Code)

	// 解析响应
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Total    int64             `json:"total"`
			Page     int               `json:"page"`
			PageSize int               `json:"page_size"`
			Items    []models.AuditLog `json:"items"`
		} `json:"data"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 验证响应内容
	assert.Equal(t, 0, response.Code)
	assert.Equal(t, "成功", response.Message)
	assert.Equal(t, int64(2), response.Data.Total)
	assert.Equal(t, 1, response.Data.Page)
	assert.Equal(t, 10, response.Data.PageSize)
	assert.Equal(t, 2, len(response.Data.Items))
	assert.Equal(t, "admin", response.Data.Items[0].Username)
	assert.Equal(t, "GET", response.Data.Items[0].Action)
	assert.Equal(t, "FILE", response.Data.Items[0].ResourceType)
	assert.Equal(t, "SUCCESS", response.Data.Items[0].Status)
}
