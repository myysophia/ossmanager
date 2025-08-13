# Security, Rate-Limiting and Pagination Enhancements - Step 3

## Overview

This document describes the security, rate-limiting, and pagination enhancements implemented for the new endpoints as part of Step 3 in the broader plan.

## 1. JWT Authentication Middleware Reuse

### Implementation
- **File**: `internal/api/middleware/auth.go`
- **Function**: `AuthMiddleware()`
- **Status**: ✅ Already implemented and reused across all secured endpoints

### Features
- JWT token validation with proper error handling
- Support for both `Bearer <token>` and direct token formats
- User context injection (`userID`, `username`) for downstream handlers
- Comprehensive error logging with request details

### Usage
```go
authorized.Use(middleware.AuthMiddleware())
```

## 2. Bucket Access Validation

### Implementation
- **File**: `internal/auth/rbac.go`
- **Function**: `CheckBucketAccess(db, userID, regionCode, bucketName)`
- **Status**: ✅ Already implemented and applied to all endpoints

### Features
- Role-based bucket access control
- Region-specific bucket validation
- User permission verification through role mappings
- Database-backed access control with proper error handling

### Usage
```go
if !auth.CheckBucketAccess(h.DB, userID, regionCode, bucketName) {
    h.Error(c, utils.CodeForbidden, "没有权限访问该存储桶")
    return
}
```

## 3. Rate Limiting Implementation

### New Implementation
- **File**: `internal/api/middleware/rate_limit.go`
- **Dependencies**: `github.com/didip/tollbooth/v6`

### Rate Limiting Configurations

#### Default Rate Limiting
- **Requests per second**: 10
- **Burst size**: 20
- **TTL**: 1 hour
- **Applied to**: General file operations

#### WebDAV Rate Limiting
- **Requests per second**: 5
- **Burst size**: 10
- **TTL**: 1 hour
- **Applied to**: WebDAV operations (both proxy API and native WebDAV)

#### Upload Rate Limiting
- **Requests per second**: 2
- **Burst size**: 5
- **TTL**: 1 hour
- **Applied to**: Upload operations and multipart uploads

### Middleware Functions
```go
// General rate limiting
middleware.RateLimitMiddleware(config)

// WebDAV specific rate limiting
middleware.WebDAVRateLimitMiddleware()

// Upload specific rate limiting
middleware.UploadRateLimitMiddleware()
```

### Per-User Rate Limiting
- Each user gets their own rate limit bucket
- Rate limiting key format: `user:{userID}`
- Automatic cleanup after TTL expiration
- No IP-based limiting (user-based only)

## 4. Pagination Support

### WebDAV Directory Listing
**File**: `internal/api/handlers/webdav_proxy.go`

#### Endpoint
```
GET /api/v1/webdav/objects/{bucket}?path=/dir&offset=0&limit=100
```

#### Parameters
- `offset`: Starting position (default: 0)
- `limit`: Maximum items per page (default: 100, max: 1000)

#### Response Format
```json
{
  "items": [...],
  "path": "/requested/path",
  "total": 500,
  "offset": 0,
  "limit": 100,
  "hasMore": true
}
```

### OSS File Listing
**File**: `internal/api/handlers/oss_file.go`

#### Endpoint
```
GET /v1/oss/files?offset=0&limit=10
```

#### Backward Compatibility
- Still supports legacy `page` and `page_size` parameters
- Automatically converts legacy parameters to offset/limit format

#### Parameters
- `offset`: Starting position (default: 0)
- `limit`: Maximum items per page (default: 10, max: 100)

#### Response Format
```json
{
  "total": 250,
  "items": [...],
  "offset": 0,
  "limit": 10,
  "hasMore": true
}
```

## 5. Router Configuration Updates

### Applied Security Enhancements
**File**: `internal/api/router.go`

#### OSS File Operations
```go
ossFiles := authorized.Group("/oss/files")
ossFiles.Use(middleware.RateLimitMiddleware(nil)) // Default rate limiting
{
    // Upload operations use stricter rate limiting
    ossFiles.POST("", middleware.UploadRateLimitMiddleware(), ossFileHandler.Upload)
    ossFiles.GET("", ossFileHandler.List)
    ossFiles.DELETE("/:id", ossFileHandler.Delete)
    ossFiles.GET("/:id/download", ossFileHandler.GetDownloadURL)
    ossFiles.GET("/check-duplicate", ossFileHandler.CheckDuplicateFile)
}
```

#### Multipart Upload Operations
```go
multipart := authorized.Group("/oss/multipart")
multipart.Use(middleware.UploadRateLimitMiddleware()) // Upload rate limiting
{
    multipart.POST("/init", ossFileHandler.InitMultipartUpload)
    multipart.POST("/complete", ossFileHandler.CompleteMultipartUpload)
    multipart.DELETE("/abort", ossFileHandler.AbortMultipartUpload)
    multipart.GET("/parts", ossFileHandler.ListUploadedParts)
}
```

#### WebDAV Proxy API
```go
webdavProxyAPI.Use(
    middleware.WebDAVProxyAuthMiddleware(db, &cfg.JWT),
    middleware.WebDAVRateLimitMiddleware(), // WebDAV-specific rate limiting
)
```

#### Native WebDAV
```go
webdavGroup.Use(
    middleware.WebDAVAuthMiddleware(db, &cfg.JWT),
    middleware.WebDAVRateLimitMiddleware(), // WebDAV-specific rate limiting
)
```

## 6. Error Handling

### Rate Limiting Errors
- **HTTP Status**: 429 (Too Many Requests)
- **Error Code**: `utils.CodeTooManyRequests`
- **Response Format**:
```json
{
  "code": 429,
  "message": "请求过于频繁，请稍后再试"
}
```

### Access Control Errors
- **HTTP Status**: 403 (Forbidden)
- **Error Code**: `utils.CodeForbidden`
- **Common Messages**:
  - "没有权限访问该存储桶"
  - "write access denied"
  - "delete access denied"

## 7. Testing

### Rate Limiting Tests
**File**: `internal/api/middleware/rate_limit_test.go`

#### Test Coverage
- ✅ Default rate limiting middleware
- ✅ WebDAV rate limiting middleware
- ✅ Upload rate limiting middleware
- ✅ Per-user rate limiting verification

#### Test Results
```
=== RUN   TestRateLimitMiddleware
--- PASS: TestRateLimitMiddleware (0.00s)
=== RUN   TestWebDAVRateLimitMiddleware
--- PASS: TestWebDAVRateLimitMiddleware (0.00s)
=== RUN   TestUploadRateLimitMiddleware
--- PASS: TestUploadRateLimitMiddleware (0.00s)
```

## 8. Security Features Summary

### ✅ Implemented Security Features

1. **JWT Authentication**
   - Token validation and parsing
   - User context injection
   - Proper error handling

2. **Bucket Access Control**
   - Role-based access validation
   - Region-specific bucket checks
   - Database-backed permission system

3. **Rate Limiting**
   - Per-user request limits
   - Differentiated limits by operation type
   - Automatic cleanup and memory management

4. **Input Validation**
   - Path validation for WebDAV operations
   - Parameter sanitization
   - Dangerous character filtering

5. **Pagination**
   - Memory-efficient large directory handling
   - Configurable page sizes with maximum limits
   - Backward compatibility support

### 🔒 Security Benefits

- **Abuse Prevention**: Rate limiting prevents API abuse and DoS attacks
- **Resource Protection**: Upload-specific limits protect server resources
- **Access Control**: Fine-grained bucket access prevents unauthorized data access
- **Memory Safety**: Pagination prevents memory exhaustion on large directories
- **Audit Trail**: All operations are logged through existing audit middleware

## 9. Configuration Examples

### Environment Variables
```bash
# Rate limiting is configured in code, but can be made configurable
RATE_LIMIT_REQUESTS_PER_SEC=10
RATE_LIMIT_BURST_SIZE=20
WEBDAV_RATE_LIMIT_REQUESTS_PER_SEC=5
UPLOAD_RATE_LIMIT_REQUESTS_PER_SEC=2
```

### Usage Examples

#### Directory listing with pagination
```bash
# Get first 50 items
curl "https://api.example.com/api/v1/webdav/objects/my-bucket?path=/&limit=50" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Get next 50 items
curl "https://api.example.com/api/v1/webdav/objects/my-bucket?path=/&offset=50&limit=50" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

#### File listing with pagination
```bash
# New format
curl "https://api.example.com/v1/oss/files?offset=0&limit=20" \
  -H "Authorization: Bearer $JWT_TOKEN"

# Legacy format (still supported)
curl "https://api.example.com/v1/oss/files?page=1&page_size=20" \
  -H "Authorization: Bearer $JWT_TOKEN"
```

## 10. Performance Considerations

### Rate Limiting Performance
- **Memory Usage**: Minimal with automatic cleanup
- **CPU Overhead**: Low overhead per request
- **Scalability**: Scales with user count, not request count

### Pagination Performance
- **Memory Usage**: Constant memory usage regardless of directory size
- **Database Impact**: Reduced database load for large result sets
- **Network Efficiency**: Smaller response payloads

## 11. Monitoring and Observability

### Rate Limiting Metrics
- Rate limit violations are logged with user context
- Failed requests include rate limiting details
- TTL-based cleanup prevents memory leaks

### Access Control Monitoring
- All access control failures are logged
- Audit trail includes user, resource, and action details
- Database query performance is logged

## Conclusion

The security, rate-limiting, and pagination enhancements have been successfully implemented with:

1. ✅ **JWT Authentication**: Reused existing middleware across all endpoints
2. ✅ **Bucket Access Validation**: Applied existing RBAC system
3. ✅ **Rate Limiting**: New per-user rate limiting with different tiers
4. ✅ **Pagination**: Enhanced directory listing for large folders
5. ✅ **Testing**: Comprehensive test coverage for new functionality
6. ✅ **Documentation**: Complete implementation documentation

All new endpoints now have comprehensive security controls while maintaining backward compatibility and performance efficiency.
