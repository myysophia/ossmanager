# OSS 文件管理系统 (OSS Manager Backend)

对象存储服务(OSS)文件管理系统的后端服务，支持阿里云OSS、AWS S3、CloudFlare R2等多种对象存储服务。

## 功能特点

- 多种对象存储服务支持（阿里云OSS、AWS S3、CloudFlare R2）
- 完善的用户权限管理（RBAC）
  - 权限粒度可分配至bucket级别
- 支持大文件分片上传与断点续传
- 文件MD5异步计算
- 详细的审计日志
- RESTful API 接口

## 技术栈

- Go 1.24
- Gin Web 框架
- GORM (PostgreSQL)
- JWT 认证
- Zap 日志
- Viper 配置管理
- AWS SDK
- Alibaba Cloud SDK

## 目录结构

```
ossmanager-backend/
├── cmd/
│   └── api/
│       └── main.go     # 应用入口
├── configs/            # 配置文件目录
│   ├── app.yaml        # 应用配置文件
│   └── oss.yaml        # OSS存储配置文件
├── internal/           # 项目核心代码目录
│   ├── api/            # API 请求处理
│   │   ├── handlers/   # 请求处理程序
│   │   ├── middleware/ # 中间件
│   │   └── router.go   # 路由定义
│   ├── auth/           # 认证相关
│   ├── config/         # 配置管理
│   ├── db/             # 数据库相关
│   ├── logger/         # 日志模块
│   ├── oss/            # 对象存储服务
│   └── utils/          # 工具函数
└── configs/            # 配置文件
    ├── app.yaml
    └── oss.yaml
```


## 安装与运行

### 前提条件

- Go 1.24+
- PostgreSQL 14+
- 对象存储服务（阿里云OSS、AWS S3 或 CloudFlare R2）

### 配置

1. 配置文件并进行修改

```bash
cp configs/app.example.yaml configs/app.yaml
cp configs/oss.example.yaml configs/oss.yaml
```

2. 修改 `configs/app.yaml` 和 `configs/oss.yaml` 中的配置项

### 构建与运行

```bash
# 编译
go build -o bin/ossmanager-api ./cmd/api

# 运行
./bin/ossmanager-api
```

### 使用 Docker 运行

```bash
# 构建镜像
docker build -t ossmanager-backend .

# 运行容器
docker run -p 8080:8080 --env-file .env ossmanager-backend
```

## API 文档

API 文档请参考 [API.md](./API.md) 文件。

## 数据库迁移

项目使用的数据库迁移文件位于 `internal/db/migrations` 目录下，可以使用以下命令进行迁移：

```bash
# 创建数据库（如果不存在）
createdb -U postgres ossmanager

# 执行迁移
psql -U postgres -d ossmanager -f internal/db/migrations/001_init_schema.sql
```

## 开发

### 目录说明

- `cmd/api`: 应用程序入口
- `internal/api`: API 请求处理逻辑
- `internal/auth`: 认证与授权
- `internal/config`: 配置管理
- `internal/db`: 数据库相关
- `internal/logger`: 日志管理
- `internal/oss`: 对象存储服务
- `internal/utils`: 工具函数

### 编码规范

项目遵循 Go 语言最佳实践和编码规范。请参考 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) 进行开发。

## 许可证

[MIT License](LICENSE)

## 区域存储桶映射

系统支持配置不同区域的存储桶映射关系，方便管理和访问不同区域的存储资源。

### 功能特点

- 支持配置区域与存储桶的映射关系
- 提供完整的CRUD接口
- 支持按区域和存储桶名称筛选
- 支持添加描述信息

### API接口

- GET /oss/region-buckets - 获取区域存储桶映射列表
- POST /oss/region-buckets - 创建区域存储桶映射
- GET /oss/region-buckets/{id} - 获取区域存储桶映射详情
- PUT /oss/region-buckets/{id} - 更新区域存储桶映射
- DELETE /oss/region-buckets/{id} - 删除区域存储桶映射
- GET /oss/region-buckets/regions - 获取所有区域列表
- GET /oss/region-buckets/buckets - 获取指定区域下的所有存储桶列表
- GET /oss/region-buckets/user-accessible - 获取用户可访问的所有存储桶列表

## 角色存储桶访问权限

系统实现了基于角色的存储桶访问控制，可以精确控制不同角色对存储桶的访问权限。

### 功能特点

- 支持为不同角色配置存储桶访问权限
- 支持细粒度的权限控制（读、写、删除）
- 提供完整的CRUD接口
- 支持按角色ID和存储桶名称筛选

### API接口

- GET /oss/role-bucket-access - 获取角色存储桶访问权限列表
- POST /oss/role-bucket-access - 创建角色存储桶访问权限
- GET /oss/role-bucket-access/{id} - 获取角色存储桶访问权限详情
- PUT /oss/role-bucket-access/{id} - 更新角色存储桶访问权限
- DELETE /oss/role-bucket-access/{id} - 删除角色存储桶访问权限

### 权限说明

系统支持以下存储桶访问权限：

- READ: 读取权限，允许查看和下载存储桶中的文件
- WRITE: 写入权限，允许上传和修改存储桶中的文件
- DELETE: 删除权限，允许删除存储桶中的文件


## 问题

1. GORM 和数据库交互报错: ERROR: prepared statement "stmtcache_464a186ac913f304ddb716c8d1cc0951f7d514b166a797c4" already exists (SQLSTATE 42P05)
   Supabase 使用了 PgBouncer 作为其连接池。PgBouncer 可以工作在不同的模式下，其中最常见的是：
   Session 模式（session pooling）： 连接在整个客户端会话期间都保持不变。在这种模式下，预处理语句会话内是持久的。
   Transaction 模式（transaction pooling）： 连接在每个事务结束后会返回到池中。最重要的是，在事务模式下，PgBouncer 在将连接返回到池中时，会清除该连接上所有已准备的语句。 这是为了确保从池中获取的连接是“干净”的，不带有前一个事务的残余状态。

从报错看使用的是session pool模式。
GORM 的 SkipDefaultTransaction: true 配置：你禁用了 GORM 的默认事务行为。这可能导致 GORM 认为它可以在不显式事务边界的情况下，持续在一个连接上操作。当 GORM 从连接池中获取一个连接时，它可能认为这个连接是一个全新的逻辑会话，然后尝试重新准备之前已经用过的同名语句。但如果 PgBouncer 处于 Session 模式，或者由于某些原因没有清除该语句，PostgreSQL 就会报错“已经存在”。

解决方案：
- 在数据库连接字符串中添加 `statement_cache_mode=describe` 参数
- 或者在 GORM 配置中设置 `PrepareStmt: false`
