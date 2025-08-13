package webdav

import (
	"net/http"
	"strconv"

	"golang.org/x/net/webdav"
	"gorm.io/gorm"

	"github.com/myysophia/ossmanager/internal/oss"
)

// Handler WebDAV处理器
type Handler struct {
	*webdav.Handler
}

// NewHandler 创建WebDAV处理器
func NewHandler(storageService oss.StorageService, db *gorm.DB, userID uint, bucket string) *Handler {
	fs := NewOSSFileSystem(storageService, db, userID, bucket)

	handler := &webdav.Handler{
		FileSystem: fs,
		LockSystem: webdav.NewMemLS(),
		Logger: func(r *http.Request, err error) {
			if err != nil {
				// 记录WebDAV错误日志
				// logger.Error("WebDAV error", zap.Error(err), zap.String("method", r.Method), zap.String("url", r.URL.String()))
			}
		},
	}

	return &Handler{Handler: handler}
}

// ServeHTTP 处理HTTP请求
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 添加浏览器兼容的CORS头
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, MKCOL, COPY, MOVE, PROPFIND, PROPPATCH, LOCK, UNLOCK, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, Depth, If, If-None-Match, Lock-Token, Overwrite, Timeout, Destination, X-User-ID, X-Bucket")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Length, Content-Type, Last-Modified, ETag, DAV")
	w.Header().Set("Access-Control-Max-Age", "86400") // 24小时预检缓存

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// 调用底层WebDAV处理器
	h.Handler.ServeHTTP(w, r)
}

// ExtractUserAndBucketFromRequest 从请求中提取用户ID和存储桶信息
// 这是一个示例函数，实际使用中需要根据认证方式调整
func ExtractUserAndBucketFromRequest(r *http.Request) (uint, string, error) {
	// 示例：从头部或查询参数中获取
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}

	bucket := r.Header.Get("X-Bucket")
	if bucket == "" {
		bucket = r.URL.Query().Get("bucket")
	}

	if userIDStr == "" || bucket == "" {
		return 0, "", http.ErrNoCookie // 使用标准错误表示认证信息缺失
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, "", err
	}

	return uint(userID), bucket, nil
}
