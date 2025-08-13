package middleware

import (
	"fmt"
	"time"

	"github.com/didip/tollbooth/v6"
	"github.com/didip/tollbooth/v6/limiter"
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager/internal/utils"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	BurstSize         int
	TTL               time.Duration
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 10.0, // 10 requests per second per user
		BurstSize:         20,   // Allow burst of 20 requests
		TTL:               time.Hour, // Clean up rate limit records after 1 hour
	}
}

// RateLimitMiddleware creates a rate limiting middleware that limits requests per user
func RateLimitMiddleware(config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	// Create limiter with custom configuration
	lmt := tollbooth.NewLimiter(config.RequestsPerSecond, &limiter.ExpirableOptions{
		DefaultExpirationTTL: config.TTL,
	})

	// Set burst size
	lmt.SetBurst(config.BurstSize)

	// Configure to limit by custom key (user ID)
	lmt.SetIPLookups([]string{}) // Disable IP-based limiting
	lmt.SetMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH"})

	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("userID")
		if !exists {
			// If no user ID, skip rate limiting (shouldn't happen after auth)
			c.Next()
			return
		}

		// Convert userID to string for use as rate limit key
		userKey := fmt.Sprintf("user:%v", userID)

		// Check rate limit for this user
		httpError := tollbooth.LimitByKeys(lmt, []string{userKey})
		if httpError != nil {
			utils.ResponseError(c, utils.CodeTooManyRequests, fmt.Errorf("rate limit exceeded: %s", httpError.Message))
			c.Abort()
			return
		}

		c.Next()
	}
}

// WebDAVRateLimitMiddleware creates a rate limiting middleware specifically for WebDAV endpoints
// with more restrictive limits for heavy operations
func WebDAVRateLimitMiddleware() gin.HandlerFunc {
	config := &RateLimitConfig{
		RequestsPerSecond: 5.0, // 5 requests per second for WebDAV operations
		BurstSize:         10,  // Lower burst size for WebDAV
		TTL:               time.Hour,
	}
	
	return RateLimitMiddleware(config)
}

// UploadRateLimitMiddleware creates a rate limiting middleware for upload endpoints
// with more restrictive limits due to resource intensity
func UploadRateLimitMiddleware() gin.HandlerFunc {
	config := &RateLimitConfig{
		RequestsPerSecond: 2.0, // 2 requests per second for uploads
		BurstSize:         5,   // Very low burst for uploads
		TTL:               time.Hour,
	}
	
	return RateLimitMiddleware(config)
}
