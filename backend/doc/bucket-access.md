# 桶访问权限实现方案

## 1. 数据库设计

### 1.1 地域-桶映射表 (region_bucket_mapping)
```sql
CREATE TABLE region_bucket_mapping (
    id SERIAL PRIMARY KEY,
    region_code VARCHAR(50) NOT NULL,
    bucket_name VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(region_code, bucket_name)
);
```

### 1.2 角色-地域桶访问权限表 (role_region_bucket_access)
```sql
CREATE TABLE role_region_bucket_access (
    id SERIAL PRIMARY KEY,
    role_id INTEGER NOT NULL REFERENCES roles(id),
    region_bucket_mapping_id INTEGER NOT NULL REFERENCES region_bucket_mapping(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(role_id, region_bucket_mapping_id)
);
```

## 2. 模型层实现

### 2.1 RegionBucketMapping 模型
```go
type RegionBucketMapping struct {
    Model
    RegionCode string  `gorm:"size:50;not null;index" json:"region_code"`  // 地域代码
    BucketName string  `gorm:"size:255;not null;index" json:"bucket_name"` // 桶的名称
    Roles      []*Role `gorm:"many2many:role_region_bucket_access;" json:"roles,omitempty"`
}
```

### 2.2 RoleRegionBucketAccess 模型
```go
type RoleRegionBucketAccess struct {
    Model
    RoleID                uint                 `gorm:"not null;index" json:"role_id"`
    Role                  *Role                `json:"role,omitempty"`
    RegionBucketMappingID uint                 `gorm:"not null;index" json:"region_bucket_mapping_id"`
    RegionBucketMapping   *RegionBucketMapping `json:"region_bucket_mapping,omitempty"`
}
```

### 2.3 Role 模型扩展
```go
type Role struct {
    Model
    Name          string                 `gorm:"size:50;uniqueIndex;not null" json:"name"`
    Description   string                 `gorm:"type:text" json:"description"`
    Users         []*User                `gorm:"many2many:user_roles;" json:"users,omitempty"`
    Permissions   []*Permission          `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
    RegionBuckets []*RegionBucketMapping `gorm:"many2many:role_region_bucket_access;" json:"region_buckets,omitempty"`
}
```

## 3. API 实现

### 3.1 地域-桶映射管理接口
```go
// RegionBucketHandler 地域-桶映射处理器
type RegionBucketHandler struct {
    BaseHandler
    DB *gorm.DB
}

// List 获取地域-桶映射列表
func (h *RegionBucketHandler) List(c *gin.Context)

// Create 创建地域-桶映射
func (h *RegionBucketHandler) Create(c *gin.Context)

// Get 获取地域-桶映射详情
func (h *RegionBucketHandler) Get(c *gin.Context)

// Update 更新地域-桶映射
func (h *RegionBucketHandler) Update(c *gin.Context)

// Delete 删除地域-桶映射
func (h *RegionBucketHandler) Delete(c *gin.Context)

// GetRegionList 获取地域列表
func (h *RegionBucketHandler) GetRegionList(c *gin.Context)

// GetBucketList 获取指定地域下的桶列表
func (h *RegionBucketHandler) GetBucketList(c *gin.Context)

// GetUserAccessibleBuckets 获取用户可访问的桶列表
func (h *RegionBucketHandler) GetUserAccessibleBuckets(c *gin.Context)
```

### 3.2 角色桶权限管理接口
```go
// RoleHandler 角色管理处理器
type RoleHandler struct {
    BaseHandler
    DB *gorm.DB
}

// ListRoleBucketAccess 获取角色存储桶访问权限列表
func (h *RoleHandler) ListRoleBucketAccess(c *gin.Context)

// CreateRoleBucketAccess 创建角色存储桶访问权限
func (h *RoleHandler) CreateRoleBucketAccess(c *gin.Context)

// GetRoleBucketAccess 获取角色存储桶访问权限详情
func (h *RoleHandler) GetRoleBucketAccess(c *gin.Context)

// UpdateRoleBucketAccess 更新角色存储桶访问权限
func (h *RoleHandler) UpdateRoleBucketAccess(c *gin.Context)

// DeleteRoleBucketAccess 删除角色存储桶访问权限
func (h *RoleHandler) DeleteRoleBucketAccess(c *gin.Context)
```

## 4. 权限控制实现

### 4.1 桶访问权限检查
```go
// CheckBucketAccess 检查用户是否有权限访问指定的桶
func CheckBucketAccess(db *gorm.DB, userID uint, regionCode, bucketName string) bool {
    var count int64
    err := db.Model(&models.RegionBucketMapping{}).
        Joins("JOIN role_region_bucket_access ON role_region_bucket_access.region_bucket_mapping_id = region_bucket_mapping.id").
        Joins("JOIN user_roles ON user_roles.role_id = role_region_bucket_access.role_id").
        Where("user_roles.user_id = ? AND region_bucket_mapping.region_code = ? AND region_bucket_mapping.bucket_name = ?",
            userID, regionCode, bucketName).
        Count(&count).Error

    if err != nil {
        return false
    }

    return count > 0
}

// GetUserAccessibleBuckets 获取用户可访问的桶列表
func GetUserAccessibleBuckets(db *gorm.DB, userID uint, regionCode string) ([]string, error) {
    var buckets []string
    query := db.Model(&models.RegionBucketMapping{}).
        Joins("JOIN role_region_bucket_access ON role_region_bucket_access.region_bucket_mapping_id = region_bucket_mapping.id").
        Joins("JOIN user_roles ON user_roles.role_id = role_region_bucket_access.role_id").
        Where("user_roles.user_id = ?", userID)

    if regionCode != "" {
        query = query.Where("region_bucket_mapping.region_code = ?", regionCode)
    }

    err := query.Distinct().
        Pluck("region_bucket_mapping.bucket_name", &buckets).
        Error

    return buckets, err
}
```

## 5. 路由配置

```go
// 地域-桶映射管理
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
```

## 6. 使用说明

### 6.1 地域-桶映射管理
1. 创建地域-桶映射
   - 请求：`POST /api/v1/oss/region-buckets`
   - 请求体：
   ```json
   {
       "region_code": "us-east-1",
       "bucket_name": "my-bucket"
   }
   ```

2. 获取地域-桶映射列表
   - 请求：`GET /api/v1/oss/region-buckets`
   - 查询参数：
     - `page`: 页码（默认1）
     - `page_size`: 每页记录数（默认10）
     - `region_code`: 按地域代码筛选
     - `bucket_name`: 按桶名称筛选

3. 获取地域列表
   - 请求：`GET /api/v1/oss/region-buckets/regions`

4. 获取指定地域下的桶列表
   - 请求：`GET /api/v1/oss/region-buckets/buckets?region_code=us-east-1`

### 6.2 角色存储桶访问权限管理
1. 创建角色存储桶访问权限
   - 请求：`POST /api/v1/oss/role-bucket-access`
   - 请求体：
   ```json
   {
       "role_id": 1,
       "region_bucket_mapping_id": 1
   }
   ```

2. 获取角色存储桶访问权限列表
   - 请求：`GET /api/v1/oss/role-bucket-access`
   - 查询参数：
     - `page`: 页码（默认1）
     - `page_size`: 每页记录数（默认10）
     - `role_id`: 按角色ID筛选
     - `region_code`: 按地域代码筛选
     - `bucket_name`: 按桶名称筛选

3. 获取用户可访问的桶列表
   - 请求：`GET /api/v1/oss/region-buckets/user-accessible`
   - 查询参数：`region_code` - 指定地域代码（可选）

### 6.3 文件操作权限检查
在以下文件操作中会自动检查用户是否有权限访问对应的桶：
- 上传文件
- 获取文件列表
- 获取文件详情
- 删除文件
- 获取文件下载链接

## 7. 注意事项

1. 所有文件操作都会进行桶访问权限的检查
2. 用户只能访问其角色被授权的桶
3. 管理员可以为角色配置桶访问权限
4. 文件列表接口只会返回用户有权限访问的桶中的文件
5. 如果用户尝试访问未授权的桶，将返回 403 Forbidden 错误
6. 删除地域-桶映射时会自动删除相关的角色访问权限
7. 创建和更新地域-桶映射时会检查是否已存在相同的映射
8. 所有接口都需要用户认证
9. 部分管理接口需要管理员权限
