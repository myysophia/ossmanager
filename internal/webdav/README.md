# WebDAV 文件系统

这个包实现了一个跨云兼容的 WebDAV 文件系统，支持阿里云 OSS、AWS S3、Cloudflare R2 等多种对象存储后端。

## 功能特性

### 核心功能
- **跨云兼容**：支持阿里云 OSS、AWS S3、Cloudflare R2 等存储服务
- **WebDAV 协议完整支持**：支持文件上传、下载、创建目录、重命名、删除等操作
- **自动数据库同步**：文件操作自动同步到 `oss_files` 数据库表
- **智能目录判定**：通过前缀查询自动判定目录结构

### 高级特性
- **大文件处理**：超过 100MB 的文件使用分片上传（TODO）
- **MD5 校验**：自动计算和存储文件 MD5 值
- **元数据管理**：完整的文件元数据记录，包括上传者、时间戳等
- **平台差异处理**：统一不同存储平台的 API 差异

## 架构设计

### 组件结构

```
webdav/
├── handler.go      # WebDAV HTTP 处理器
├── filesystem.go   # 文件系统实现
├── file.go        # 文件对象实现
└── README.md      # 此文档
```

### 依赖关系

```
WebDAV Handler
      ↓
OSS FileSystem (实现 webdav.FileSystem 接口)
      ↓
StorageServiceAdapter (适配器模式)
      ↓
StorageService (阿里云OSS/AWS S3/R2 实现)
```

## 使用示例

### 基本使用

```go
package main

import (
    "log"
    "net/http"
    
    "github.com/myysophia/ossmanager/internal/webdav"
    "github.com/myysophia/ossmanager/internal/oss"
    "github.com/myysophia/ossmanager/internal/db"
)

func main() {
    // 初始化数据库
    database := db.GetDB()
    
    // 创建存储工厂
    factory := oss.NewStorageFactory(ossConfig)
    
    // 获取存储服务
    service, err := factory.GetStorageService("ALIYUN_OSS")
    if err != nil {
        log.Fatal(err)
    }
    
    // 创建 WebDAV 处理器
    handler := webdav.NewHandler(service, database, userID, "my-bucket")
    
    // 启动 HTTP 服务器
    http.Handle("/webdav/", http.StripPrefix("/webdav", handler))
    log.Println("WebDAV server started on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### 客户端连接示例

#### 使用 curl
```bash
# 列出根目录
curl -X PROPFIND http://localhost:8080/webdav/ \
  -H "X-User-ID: 1" \
  -H "X-Bucket: my-bucket"

# 上传文件
curl -X PUT http://localhost:8080/webdav/test.txt \
  -H "X-User-ID: 1" \
  -H "X-Bucket: my-bucket" \
  --data-binary @local-file.txt

# 创建目录
curl -X MKCOL http://localhost:8080/webdav/newfolder/ \
  -H "X-User-ID: 1" \
  -H "X-Bucket: my-bucket"
```

#### 使用 WebDAV 客户端
大多数操作系统都支持 WebDAV：

**Windows**：在文件资源管理器地址栏输入 `http://localhost:8080/webdav/`

**macOS**：Finder -> 前往 -> 连接服务器 -> `http://localhost:8080/webdav/`

**Linux**：使用 `davfs2` 或在文件管理器中输入地址

## 实现细节

### 文件操作流程

#### 上传文件（小文件 < 100MB）
1. WebDAV 客户端发起 PUT 请求
2. `OSSFile.Write()` 将数据写入内存缓冲区
3. `OSSFile.Close()` 触发实际上传：
   - 调用 `StorageService.PutObjectToBucket()` 上传到对象存储
   - 计算文件 MD5 值
   - 调用 `syncToDatabase()` 同步到数据库

#### 上传文件（大文件 ≥ 100MB）
1. 系统检测文件大小超过阈值
2. 返回错误并建议使用分片上传 API（TODO: 待实现）

#### 目录判定
系统通过以下规则判定目录：
1. 路径以 `/` 结尾
2. 存在以该路径为前缀的对象

#### 数据库同步
文件操作时会自动同步以下信息到 `oss_files` 表：
- `filename`: 文件名
- `original_filename`: 原始文件名
- `file_size`: 文件大小
- `md5`: MD5 哈希值
- `storage_type`: 存储类型（ALIYUN_OSS/AWS_S3/CLOUDFLARE_R2）
- `bucket`: 存储桶名称
- `object_key`: 对象键
- `uploader_id`: 上传者 ID
- `status`: 文件状态（ACTIVE）

### 平台差异处理

不同存储平台在 API 语义上存在差异，系统通过以下方式统一：

#### 阿里云 OSS
- **Copy**: 使用 `CopyObject(srcPath, dstKey)` 
- **Delete**: 支持批量删除
- **List**: 支持前缀和分页查询

#### AWS S3
- **Copy**: 使用 `CopySource` 格式 `bucket/key`
- **Delete**: 单个删除，需要循环处理批量操作
- **List**: 使用 `ListObjectsV2`，参数略有不同

#### Cloudflare R2
- 基于 S3 兼容 API，与 AWS S3 类似
- TODO: 待实现具体差异处理

### 错误处理

系统实现了完善的错误处理机制：

1. **存储服务错误**：网络问题、认证失败等
2. **数据库错误**：同步失败时记录日志但不阻塞主流程
3. **权限错误**：用户权限验证失败
4. **文件系统错误**：路径非法、文件不存在等

## 配置说明

### 认证配置
目前支持两种认证方式：
- **请求头认证**：`X-User-ID` 和 `X-Bucket`
- **查询参数认证**：`?user_id=1&bucket=my-bucket`

### 存储配置
支持的存储类型：
- `ALIYUN_OSS`: 阿里云对象存储
- `AWS_S3`: Amazon S3
- `CLOUDFLARE_R2`: Cloudflare R2

## 性能优化

### 内存使用
- 小文件（<100MB）：完全在内存中处理
- 大文件（≥100MB）：TODO - 实现流式处理和分片上传

### 网络优化
- 使用存储服务的原生 SDK 以获得最佳性能
- 支持传输加速（阿里云 OSS）
- 自动重试机制

### 数据库优化
- 使用索引：`object_key + bucket` 联合索引
- 批量操作：目录删除时批量处理数据库更新

## 扩展性

### 添加新的存储服务
1. 实现 `StorageService` 接口
2. 在 `StorageFactory` 中注册新类型
3. 处理该平台特有的 API 差异

### 自定义认证
实现自己的认证逻辑，替换 `ExtractUserAndBucketFromRequest()` 函数

### 中间件支持
可以在 `Handler.ServeHTTP()` 中添加自定义中间件，如：
- 访问日志
- 速率限制
- 缓存控制

## TODO 和已知限制

### 待实现功能
- [ ] 大文件分片上传支持
- [ ] Cloudflare R2 完整实现
- [ ] WebDAV 锁定机制优化
- [ ] 更细粒度的权限控制
- [ ] 文件版本管理

### 已知限制
1. 大文件（>100MB）暂不支持，会返回错误
2. 不支持文件的随机读写（Seek 操作）
3. 目录操作性能在对象数量很大时可能较慢
4. 暂不支持 WebDAV 的高级功能（如属性、锁定等）

### 性能考虑
- 大型目录列表可能较慢（当前限制 1000 个对象）
- 频繁的小文件操作会产生较多的 API 调用
- 数据库同步是同步操作，可能影响响应时间

## 贡献指南

欢迎提交 Issue 和 Pull Request！

### 开发环境设置
```bash
go mod download
go test ./internal/webdav/...
```

### 测试
```bash
# 运行单元测试
go test -v ./internal/webdav/

# 运行基准测试
go test -bench=. ./internal/webdav/
```
