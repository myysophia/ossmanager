package main

import (
	"context"
	"crypto/tls"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/myysophia/ossmanager/internal/api"
	"github.com/myysophia/ossmanager/internal/api/handlers"
	"github.com/myysophia/ossmanager/internal/api/middleware"
	"github.com/myysophia/ossmanager/internal/config"
	"github.com/myysophia/ossmanager/internal/db"
	"github.com/myysophia/ossmanager/internal/function"
	"github.com/myysophia/ossmanager/internal/logger"
	"github.com/myysophia/ossmanager/internal/oss"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

//go:embed web/build/*
var staticFiles embed.FS

func main() {
	// 加载 .env 文件
	if err := godotenv.Load(); err != nil {
		// .env 文件不存在或加载失败时，只是记录日志，不影响程序启动
		fmt.Printf("警告：无法加载 .env 文件: %v\n", err)
	}

	// 加载配置 - 使用原有后端的配置系统
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev" // 默认开发环境
	}
	cfg, err := config.LoadConfigWithEnv("configs", env)
	if err != nil {
		panic(fmt.Sprintf("加载配置失败: %v", err))
	}

	// 初始化日志系统
	if err := logger.InitLogger(&cfg.Log); err != nil {
		panic(fmt.Sprintf("初始化日志系统失败: %v", err))
	}
	defer logger.Sync()

	logger.Info("🚀 OSS Manager 简化单体服务启动中...")
	logger.Info("📦 前端静态文件已嵌入")
	logger.Info("配置加载成功", zap.String("env", cfg.App.Env))

	// 调试：打印数据库配置信息
	logger.Info("数据库配置",
		zap.String("host", cfg.Database.Host),
		zap.Int("port", cfg.Database.Port),
		zap.String("username", cfg.Database.Username),
		zap.String("dbname", cfg.Database.DBName),
		zap.String("sslmode", cfg.Database.SSLMode),
	)
	logger.Info("数据库连接字符串", zap.String("dsn", cfg.Database.GetDSN()))

	// 初始化数据库
	if err := db.Init(&cfg.Database); err != nil {
		logger.Error("初始化数据库失败，继续启动但功能可能受限", zap.Error(err))
		// 暂时不退出，允许程序继续运行以测试前端
	} else {
		logger.Info("✅ 数据库初始化成功")
	}

	// 创建存储服务工厂
	storageFactory := oss.NewStorageFactory(&cfg.OSS)

	// 创建MD5计算器
	md5Calculator := function.NewMD5Calculator(storageFactory, cfg.App.Workers)
	logger.Info("MD5计算器初始化成功", zap.Int("workers", cfg.App.Workers))

	// 设置API路由
	apiRouter := api.SetupRouter(storageFactory, md5Calculator, db.GetDB(), cfg)

	// 创建主路由器，整合API和静态文件服务
	mainRouter := setupIntegratedRouter(apiRouter)

	// 创建HTTP服务器配置
	readTimeout := time.Duration(cfg.App.ReadTimeout) * time.Second
	if cfg.App.ReadTimeout <= 0 {
		readTimeout = 0
	}
	writeTimeout := time.Duration(cfg.App.WriteTimeout) * time.Second
	if cfg.App.WriteTimeout <= 0 {
		writeTimeout = 0
	}
	idleTimeout := time.Duration(cfg.App.IdleTimeout) * time.Second
	if cfg.App.IdleTimeout <= 0 {
		idleTimeout = 0
	}

	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
		Handler: mainRouter,
		// 禁用HTTP/2，强制使用HTTP/1.1以确保SSE连接稳定性
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
		// 设置超时时间，优化长连接
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
		ReadHeaderTimeout: 10 * time.Second,
	}

	// 优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// 启动HTTP服务器
	go func() {
		logger.Info("🌐 OSS Manager 简化单体服务启动成功", zap.String("addr", server.Addr))
		logger.Info("前端访问: http://"+server.Addr)
		logger.Info("API访问: http://"+server.Addr+"/api/v1")
		logger.Info("📁 WebDAV访问: http://"+server.Addr+"/webdav/{bucket}")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("HTTP服务器启动失败", zap.Error(err))
		}
	}()

	// 等待退出信号
	<-quit
	logger.Info("正在关闭服务器...")

	// 关闭MD5计算器
	md5Calculator.Stop()
	logger.Info("MD5计算器已关闭")

	// 设置关闭超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 关闭HTTP服务器
	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("服务器关闭异常", zap.Error(err))
	}

	// 关闭数据库连接
	if err := db.Close(); err != nil {
		logger.Error("关闭数据库连接失败", zap.Error(err))
	}

	logger.Info("✅ 服务器已安全关闭")
}

// setupSimplifiedAPIRouter 设置简化的API路由，跳过WebDAV功能
func setupSimplifiedAPIRouter(storageFactory oss.StorageFactory, md5Calculator *function.MD5Calculator, database *gorm.DB) *gin.Engine {
	// 创建Gin实例
	router := gin.New()

	// 全局中间件
	router.Use(
		gin.Recovery(),                  // 内置恢复中间件
		middleware.RecoveryMiddleware(), // 自定义恢复中间件
		middleware.LoggerMiddleware(),   // 日志中间件
		middleware.CorsMiddleware(),     // 跨域中间件
	)

	// 创建处理器
	authHandler := handlers.NewAuthHandler()
	ossFileHandler := handlers.NewOSSFileHandler(storageFactory, database)
	ossConfigHandler := handlers.NewOSSConfigHandler(storageFactory)
	md5Handler := handlers.NewMD5Handler(md5Calculator)
	auditLogHandler := handlers.NewAuditLogHandler()           // 审计日志处理器
	userHandler := handlers.NewUserHandler()                   // 用户管理处理器
	roleHandler := handlers.NewRoleHandler(database)                 // 角色管理处理器
	permissionHandler := handlers.NewPermissionHandler(database)     // 权限管理处理器
	regionBucketHandler := handlers.NewRegionBucketHandler(database) // 区域存储桶处理器
	uploadProgressHandler := handlers.NewUploadProgressHandler()
	// 暂时禁用WebDAV相关的handler

	// 公开路由
	public := router.Group("/v1")
	{
		// 认证相关
		public.POST("/auth/login", authHandler.Login)
		public.POST("/auth/register", authHandler.Register)

		// 上传进度查询（不需要认证，因为taskId本身就是安全的UUID）
		uploads := public.Group("/uploads")
		uploads.Use(
			middleware.SSEMiddleware(),       // SSE连接稳定性中间件
			middleware.HTTP1OnlyMiddleware(), // 强制HTTP/1.1
			middleware.NoBufferMiddleware(),  // 禁用缓冲
		)
		{
			uploads.POST("/init", uploadProgressHandler.Init)
			uploads.GET("/:id/progress", uploadProgressHandler.GetProgress)
			uploads.GET("/:id/stream", uploadProgressHandler.StreamProgress)
		}
	}

	// 需要认证的路由
	authorized := router.Group("/v1")
	authorized.Use(
		middleware.AuthMiddleware(),     // 认证中间件
		middleware.AuditLogMiddleware(), // 审计日志中间件
	)
	{
		// 用户相关
		authorized.GET("/user/current", authHandler.GetCurrentUser)

		// 用户管理
		users := authorized.Group("/users")
		{
			users.GET("", userHandler.List)
			users.POST("", userHandler.Create)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
			users.GET("/:id/bucket-access", userHandler.GetUserBucketAccess)
		}

		// 角色管理
		roles := authorized.Group("/roles")
		{
			roles.GET("", roleHandler.List)
			roles.POST("", roleHandler.Create)
			roles.GET("/:id", roleHandler.Get)
			roles.PUT("/:id", roleHandler.Update)
			roles.DELETE("/:id", roleHandler.Delete)
			roles.GET("/:id/bucket-access", roleHandler.GetRoleBucketAccess)
			roles.PUT("/:id/bucket-access", roleHandler.UpdateRoleBucketAccess)
		}

		// 添加 region-bucket-mappings 路由组
		regionBucketMappings := authorized.Group("/region-bucket-mappings")
		{
			regionBucketMappings.GET("", roleHandler.ListRegionBucketMappings)
		}

		// 权限管理
		permissions := authorized.Group("/permissions")
		{
			permissions.GET("", permissionHandler.List)
			permissions.POST("", permissionHandler.Create)
			permissions.GET("/:id", permissionHandler.Get)
			permissions.PUT("/:id", permissionHandler.Update)
			permissions.DELETE("/:id", permissionHandler.Delete)
		}

		// OSS文件管理
		authorized.POST("/oss/files", ossFileHandler.Upload)
		authorized.GET("/oss/files", ossFileHandler.List)
		authorized.DELETE("/oss/files/:id", ossFileHandler.Delete)
		authorized.GET("/oss/files/:id/download", ossFileHandler.GetDownloadURL)
		authorized.GET("/oss/files/check-duplicate", ossFileHandler.CheckDuplicateFile)

		// 分片上传
		authorized.POST("/oss/multipart/init", ossFileHandler.InitMultipartUpload)
		authorized.POST("/oss/multipart/complete", ossFileHandler.CompleteMultipartUpload)
		authorized.DELETE("/oss/multipart/abort", ossFileHandler.AbortMultipartUpload)
		authorized.GET("/oss/multipart/parts", ossFileHandler.ListUploadedParts)

		// MD5计算相关
		authorized.POST("/oss/files/:id/md5", md5Handler.TriggerCalculation)
		authorized.GET("/oss/files/:id/md5", md5Handler.GetMD5)

		// OSS配置管理（仅管理员可访问）
		configs := authorized.Group("/oss/configs")
		configs.Use(middleware.AdminMiddleware()) // 管理员权限中间件
		{
			configs.POST("", ossConfigHandler.CreateConfig)
			configs.PUT("/:id", ossConfigHandler.UpdateConfig)
			configs.DELETE("/:id", ossConfigHandler.DeleteConfig)
			configs.GET("", ossConfigHandler.GetConfigList)
			configs.GET("/:id", ossConfigHandler.GetConfig)
			configs.PUT("/:id/default", ossConfigHandler.SetDefaultConfig)
		}

		// 审计日志管理（仅管理员可访问）
		audit := authorized.Group("/audit")
		audit.Use(middleware.AdminMiddleware()) // 管理员权限中间件
		{
			audit.GET("/logs", auditLogHandler.ListAuditLogs)
		}

		// 区域存储桶管理
		regionBuckets := authorized.Group("/oss/region-buckets")
		{
			regionBuckets.GET("", regionBucketHandler.List)
			regionBuckets.POST("", regionBucketHandler.Create)
			regionBuckets.GET("/:id", regionBucketHandler.Get)
			regionBuckets.PUT("/:id", regionBucketHandler.Update)
			regionBuckets.DELETE("/:id", regionBucketHandler.Delete)
			regionBuckets.GET("/regions", regionBucketHandler.GetRegionList)
			regionBuckets.GET("/buckets", regionBucketHandler.GetBucketList)
			regionBuckets.GET("/user-accessible", regionBucketHandler.GetUserAccessibleBuckets)
		}

		// 角色存储桶访问权限管理
		roleBucketAccess := authorized.Group("/oss/role-bucket-access")
		{
			roleBucketAccess.GET("", roleHandler.ListRoleBucketAccess)
			roleBucketAccess.POST("", roleHandler.CreateRoleBucketAccess)
			roleBucketAccess.GET("/:id", roleHandler.GetRoleBucketAccess)
			roleBucketAccess.PUT("/:id", roleHandler.UpdateRoleBucketAccess)
			roleBucketAccess.DELETE("/:id", roleHandler.DeleteRoleBucketAccess)
		}
	}

	return router
}

// setupIntegratedRouter 设置集成的路由器，包含API路由和静态文件服务
func setupIntegratedRouter(apiRouter *gin.Engine) *gin.Engine {
	// 创建主路由器
	mainRouter := gin.New()

	// 全局中间件
	mainRouter.Use(
		gin.Recovery(),
		ginLoggerMiddleware(),
	)

	// API路由组 - 优先级最高
	mainRouter.Any("/api/*proxyPath", gin.WrapH(http.StripPrefix("/api", apiRouter)))
	
	// WebDAV路由 - 高优先级，代理到API路由器但不去掉前缀
	mainRouter.Any("/webdav/*proxyPath", gin.WrapH(apiRouter))

	// 静态文件和SPA路由 - 最低优先级
	mainRouter.NoRoute(func(c *gin.Context) {
		serveStaticFile(c)
	})

	return mainRouter
}

// ginLoggerMiddleware 简单的日志中间件
func ginLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		c.Next()

		latency := time.Since(start)

		if raw != "" {
			path = path + "?" + raw
		}

		// 只记录非静态文件的请求
		if !strings.HasPrefix(path, "/_next/") && 
		   !strings.HasSuffix(path, ".js") && 
		   !strings.HasSuffix(path, ".css") && 
		   !strings.HasSuffix(path, ".ico") &&
		   !strings.HasPrefix(path, "/favicon") {
			logger.Info("HTTP Request",
				zap.String("method", c.Request.Method),
				zap.String("path", path),
				zap.Int("status", c.Writer.Status()),
				zap.Duration("latency", latency),
				zap.String("ip", c.ClientIP()),
			)
		}
	}
}

// serveStaticFile 提供静态文件服务
func serveStaticFile(c *gin.Context) {
	logger.Info("静态文件请求", zap.String("path", c.Request.URL.Path))
	
	subFS, err := fs.Sub(staticFiles, "web/build")
	if err != nil {
		logger.Error("静态文件系统初始化失败", zap.Error(err))
		c.String(http.StatusInternalServerError, "Static files not available")
		return
	}

	requestPath := strings.TrimPrefix(c.Request.URL.Path, "/")

	// 如果是根路径，返回 index.html
	if requestPath == "" {
		requestPath = "index.html"
	}

	// API 路径应该已经被前面的路由处理，这里不应该到达
	if strings.HasPrefix(requestPath, "api/") {
		c.Status(http.StatusNotFound)
		return
	}

	// 尝试的文件路径列表
	possiblePaths := []string{
		requestPath,
		requestPath + "/index.html",
		requestPath + ".html",
	}

	var file fs.File
	var finalPath string
	var fileFound bool

	for _, possiblePath := range possiblePaths {
		file, err = subFS.Open(possiblePath)
		if err == nil {
			// 检查是否为目录
			if stat, err := file.Stat(); err == nil && stat.IsDir() {
				file.Close() // 关闭目录
				continue
			}
			finalPath = possiblePath
			fileFound = true
			break
		}
	}

	if !fileFound {
		// SPA fallback - 返回 index.html
		file, err = subFS.Open("index.html")
		if err != nil {
			logger.Error("无法打开 index.html", zap.Error(err))
			c.String(http.StatusNotFound, "File not found")
			return
		}
		finalPath = "index.html"
	}
	defer file.Close()

	// 获取文件信息
	stat, err := file.Stat()
	if err != nil {
		logger.Error("获取文件信息失败", zap.String("path", finalPath), zap.Error(err))
		c.String(http.StatusInternalServerError, "Could not get file info")
		return
	}

	// 设置正确的 Content-Type
	contentType := getContentType(finalPath)

	// 设置缓存头
	if shouldCache(finalPath) {
		c.Header("Cache-Control", "public, max-age=31536000") // 1 year
	} else {
		c.Header("Cache-Control", "no-cache")
	}

	// 使用 Gin 的 DataFromReader 方法来提供文件
	c.DataFromReader(http.StatusOK, stat.Size(), contentType, file, nil)
}

// getContentType 根据文件扩展名返回正确的 Content-Type
func getContentType(filename string) string {
	ext := path.Ext(filename)
	switch ext {
	case ".html":
		return "text/html; charset=utf-8"
	case ".js":
		return "application/javascript; charset=utf-8"
	case ".css":
		return "text/css; charset=utf-8"
	case ".json":
		return "application/json"
	case ".png":
		return "image/png"
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".ico":
		return "image/x-icon"
	case ".woff":
		return "font/woff"
	case ".woff2":
		return "font/woff2"
	case ".ttf":
		return "font/ttf"
	case ".otf":
		return "font/otf"
	default:
		return "application/octet-stream"
	}
}

// shouldCache 判断文件是否应该被缓存
func shouldCache(filename string) bool {
	ext := path.Ext(filename)
	switch ext {
	case ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".ico", ".woff", ".woff2", ".ttf", ".otf":
		return true
	default:
		return false
	}
}
