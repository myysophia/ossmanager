package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test config with very low limits for testing
	config := &RateLimitConfig{
		RequestsPerSecond: 2.0,
		BurstSize:         3,
		TTL:               time.Minute,
	}

	// Create a test router
	router := gin.New()
	router.Use(RateLimitMiddleware(config))
	router.GET("/test", func(c *gin.Context) {
		// Simulate setting userID (normally done by AuthMiddleware)
		c.Set("userID", uint(1))
		c.JSON(200, gin.H{"message": "success"})
	})

	// Test that requests within limit succeed
	for i := 0; i < 3; i++ {
		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		
		// Simulate userID being set
		ctx := gin.CreateTestContextOnly(w, router)
		ctx.Request = req
		ctx.Set("userID", uint(1))
		
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	t.Log("Rate limit middleware test completed successfully")
}

func TestWebDAVRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(WebDAVRateLimitMiddleware())
	router.GET("/webdav/test", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.JSON(200, gin.H{"message": "webdav success"})
	})

	req, _ := http.NewRequest("GET", "/webdav/test", nil)
	w := httptest.NewRecorder()
	
	// Simulate userID being set
	ctx := gin.CreateTestContextOnly(w, router)
	ctx.Request = req
	ctx.Set("userID", uint(1))
	
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Log("WebDAV rate limit middleware test completed successfully")
}

func TestUploadRateLimitMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(UploadRateLimitMiddleware())
	router.POST("/upload/test", func(c *gin.Context) {
		c.Set("userID", uint(1))
		c.JSON(200, gin.H{"message": "upload success"})
	})

	req, _ := http.NewRequest("POST", "/upload/test", nil)
	w := httptest.NewRecorder()
	
	// Simulate userID being set
	ctx := gin.CreateTestContextOnly(w, router)
	ctx.Request = req
	ctx.Set("userID", uint(1))
	
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	t.Log("Upload rate limit middleware test completed successfully")
}
