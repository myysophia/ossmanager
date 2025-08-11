package api

import (
	"github.com/gin-gonic/gin"
	"github.com/myysophia/ossmanager-backend/internal/api/handlers"
	"github.com/myysophia/ossmanager-backend/internal/api/middleware"
	"github.com/myysophia/ossmanager-backend/internal/function"
	"github.com/myysophia/ossmanager-backend/internal/oss"
	"gorm.io/gorm"
)

// SetupRouter 设置路由
func SetupRouter(storageFactory oss.StorageFactory, md5Calculator *function.MD5Calculator, db *gorm.DB) *gin.Engine {
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
	ossFileHandler := handlers.NewOSSFileHandler(storageFactory, db)
	ossConfigHandler := handlers.NewOSSConfigHandler(storageFactory)
	md5Handler := handlers.NewMD5Handler(md5Calculator)
	auditLogHandler := handlers.NewAuditLogHandler()           // 审计日志处理器
	userHandler := handlers.NewUserHandler()                   // 用户管理处理器
	roleHandler := handlers.NewRoleHandler(db)                 // 角色管理处理器
	permissionHandler := handlers.NewPermissionHandler(db)     // 权限管理处理器
	regionBucketHandler := handlers.NewRegionBucketHandler(db) // 区域存储桶处理器
	uploadProgressHandler := handlers.NewUploadProgressHandler()
	// webdavHandler := handlers.NewWebDAVHandler(storageFactory, db) // WebDAV 处理器 - 暂时禁用

	// 公开路由
	public := router.Group("/api/v1")
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
	authorized := router.Group("/api/v1")
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
		//authorized.GET("/oss/files/by-filename", ossFileHandler.GetByOriginalFilename)

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

	// WebDAV 支持 - 暂时禁用，避免编译错误
	/*
	webdavGroup := router.Group("/webdav")
	webdavGroup.Use(middleware.WebDAVAuthMiddleware(db))
	{
		// 处理所有 WebDAV 请求
		webdavGroup.Any("/*bucket", webdavHandler.ServeHTTP)
	}
	*/

	return router
}
