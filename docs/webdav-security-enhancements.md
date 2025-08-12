# WebDAV 权限与安全增强

本文档描述了 WebDAV 功能的权限与安全增强措施的实现。

## 功能概述

根据任务要求，我们实现了以下安全增强：

### 1. 增强 `WebDAVAuthMiddleware`

#### 支持的认证方式
- **Basic Auth (用户名/密码)**
  - 支持标准的 HTTP Basic Authentication
  - 用户名密码存储在数据库中，密码使用 bcrypt 加密
  - 认证成功后设置用户上下文信息

- **Bearer JWT**
  - 支持 JWT Token 认证
  - 使用现有的 JWT 配置和密钥
  - 验证 Token 有效性和过期时间
  - 检查用户状态（是否被禁用）

#### 限速与失败计数
- **IP 级别限速**：基于客户端 IP 地址进行失败计数
- **失败阈值**：默认 5 次失败后进入封锁状态
- **封锁时长**：默认 15 分钟自动解除封锁
- **自动清理**：定期清理过期的失败记录（1小时）

#### WWW-Authenticate 头
- 认证失败时返回适当的 `WWW-Authenticate` 头
- 支持同时提示 Basic 和 Bearer 认证方式
- 限速时返回 429 状态码和重试时间

### 2. RBAC 校验增强

#### 桶级别访问控制
- 复用现有的角色-桶映射表进行权限检查
- 使用 `auth.CheckBucketAccess` 函数进行统一权限验证
- 支持多区域桶的权限管理

#### 资源级别权限校验（预留）
- 实现了资源级别权限检查的框架
- 根据 HTTP 方法映射到不同的操作权限：
  - `GET/HEAD/OPTIONS/PROPFIND` → `read` 权限
  - `PUT/POST/MKCOL/COPY/MOVE` → `write` 权限
  - `DELETE` → `delete` 权限
  - `LOCK/UNLOCK/PROPPATCH` → `manage` 权限
- 预留扩展接口支持基于文件路径的细粒度权限控制

### 3. 路径安全

#### 目录穿越防护
实现了 `security` 包，提供以下安全功能：

- **CleanPath**：清理路径，移除 `..` 和多余的 `/`
- **IsPathSafe**：检查目标路径是否在基础路径内
- **ValidateWebDAVPath**：验证 WebDAV 路径是否安全
- **ExtractBucketAndPath**：安全地从 URL 路径提取桶名和文件路径

#### 安全措施
- 使用 `filepath.Clean` 标准化路径
- 检查相对路径中的 `..` 元素
- 验证存储桶名称不包含危险字符
- 确保文件路径不能访问存储桶外的内容

### 4. 日志与审计

#### WebDAV 操作审计
- 记录所有 WebDAV 操作到审计日志表
- 包含以下信息：
  - 用户ID和用户名
  - 操作方法（GET/PUT/DELETE等）
  - 存储桶和资源路径
  - 客户端IP地址和User-Agent
  - 操作结果（成功/失败）
  - 错误详情（如果有）

#### 双重日志记录
- **数据库审计日志**：结构化存储到 `audit_logs` 表
- **应用日志**：同时记录到应用日志系统
- **异步处理**：审计日志保存采用异步处理，不影响主要业务流程

## 实现文件

### 核心文件
1. **`internal/api/middleware/webdav_auth.go`** - 增强的认证中间件
2. **`internal/security/path.go`** - 路径安全工具
3. **`internal/api/handlers/webdav.go`** - 增强的 WebDAV 处理器

### 测试文件
- **`internal/security/path_test.go`** - 路径安全功能的单元测试

## 配置说明

### 认证限流配置
```go
// 可配置的参数（目前硬编码，可以移到配置文件）
maxFailures   = 5                // 最大失败次数
blockDuration = 15 * time.Minute // 封锁时长
```

### JWT 配置
使用现有的 `config.JWTConfig` 配置：
- `SecretKey`：JWT 签名密钥
- `ExpiresIn`：Token 过期时间
- `Issuer`：Token 发行者

## 使用方法

### 启用 WebDAV 功能
由于 WebDAV 功能使用了条件编译，需要使用特定标签编译：
```bash
go build -tags webdav
```

### 认证示例

#### Basic Auth
```bash
curl -u username:password http://localhost:8080/webdav/bucket/file.txt
```

#### Bearer JWT
```bash
curl -H "Authorization: Bearer your-jwt-token" http://localhost:8080/webdav/bucket/file.txt
```

## 安全特性

### 1. 认证安全
- 密码使用 bcrypt 加密存储
- JWT Token 验证签名和过期时间
- IP 级别的失败计数和限速
- 用户状态检查（防止已禁用用户访问）

### 2. 授权安全
- 基于 RBAC 的桶级别权限控制
- 操作级别的权限检查
- 预留资源级别的细粒度权限控制

### 3. 路径安全
- 目录穿越攻击防护
- 存储桶名称安全验证
- 路径规范化和清理

### 4. 审计安全
- 完整的操作审计日志
- 异步日志记录，防止性能影响
- 结构化日志便于分析和监控

## 扩展建议

### 1. 生产环境优化
- 将认证失败计数器迁移到 Redis（支持分布式）
- 配置化限流参数
- 增加更多的安全头（如 CSRF 防护）

### 2. 权限细化
- 实现基于文件路径的权限控制
- 支持文件夹级别的权限设置
- 增加只读/读写权限的细分

### 3. 监控告警
- 异常访问模式检测
- 失败率监控和告警
- 性能指标收集

## 测试验证

运行路径安全测试：
```bash
go test ./internal/security/
```

测试覆盖了以下场景：
- 路径清理功能
- 路径安全检查
- WebDAV 路径验证
- 存储桶和路径提取

## 注意事项

1. **编译标签**：WebDAV 功能需要使用 `-tags webdav` 编译
2. **性能影响**：认证限流使用内存存储，重启后计数器会重置
3. **日志存储**：审计日志会持续增长，需要定期清理或归档
4. **权限检查**：每次请求都会进行多层权限验证，可能影响性能

这个实现提供了全面的安全增强，既满足了当前的安全需求，也为未来的扩展预留了接口。
