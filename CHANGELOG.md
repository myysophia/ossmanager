# Changelog

所有重要的更改都将记录在此文件中。

版本格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [0.2.0] - 2024-12-13

### 新增

- **WebDAV 文件浏览器** - 全新的现代化浏览器内文件管理界面
  - 🚀 即时访问：登录后直接使用，无需额外配置
  - 🔒 自动认证：使用用户的 JWT Token 自动完成身份验证
  - 📁 完整功能：支持文件浏览、上传、下载、创建文件夹等所有操作
  - 🌐 跨平台兼容：支持所有现代浏览器，包括桌面和移动端
  - ⚡ 高性能：通过代理 API 优化传输，支持大文件和分片上传
  - 👥 多存储桶：一键切换不同存储桶，统一管理界面

- **WebDAV API 扩展** - 完整的 RESTful API 支持
  - 存储桶管理 API（列表、连接信息）
  - WebDAV 令牌管理 API（创建、列表、删除）
  - 文件操作代理 API（列表、上传、下载、删除、移动、复制）
  - 统计和监控 API（存储桶统计、访问日志）

- **增强的 WebDAV 认证**
  - 双重认证模式：Basic Auth + Bearer Token
  - 浏览器内自动 JWT 认证
  - 外部客户端令牌认证

### 改进

- **WebDAV 用户体验** - 优化传统 WebDAV 客户端支持
  - 改进的令牌管理界面
  - 更清晰的连接配置指南
  - 增强的错误处理和用户反馈

- **文档更新**
  - 全面更新 WebDAV 使用指南
  - 添加浏览器端功能详细说明
  - 增加完整的 API 参考文档
  - 提供 JavaScript/TypeScript 和 Python SDK 示例

- **性能优化**
  - WebDAV 代理 API 传输优化
  - 大文件上传和分片处理改进
  - 缓存策略优化

### 修复

- 修复 WebDAV 连接在某些客户端上的兼容性问题
- 解决中文文件名编码问题
- 改进网络错误处理和重试机制

### 安全

- 加强 WebDAV 令牌安全性
- 改进审计日志记录
- 增强权限验证机制

### 技术栈

- Go 1.24+ (后端)
- Next.js 15+ (前端)
- PostgreSQL 15+ (数据库)
- Docker & Kubernetes 支持

## [0.1.0] - 2024-08-12

### 新增

- **核心功能**
  - 多云存储支持（阿里云 OSS、AWS S3、CloudFlare R2）
  - 基础 WebDAV 支持
  - 用户权限管理（基于 RBAC）
  - 文件管理（上传、下载、删除、预览）
  - 分片上传和断点续传
  - MD5 校验和文件完整性验证
  - 完整的审计日志

- **架构特性**
  - 单体应用架构
  - 前后端集成部署
  - 嵌入式静态资源
  - Docker 容器化支持

- **安全功能**
  - JWT Token 认证
  - bcrypt 密码加密
  - RBAC 权限控制
  - API 安全验证

### 技术实现

- Go 1.23+ 后端服务
- React + Next.js 14 前端
- PostgreSQL 14+ 数据库
- WebDAV 标准协议实现

---

## 版本说明

- **主要版本** (x.0.0): 包含破坏性变更或重大新功能
- **次要版本** (0.x.0): 添加新功能，向后兼容
- **修订版本** (0.0.x): Bug 修复和小的改进

## 贡献

我们欢迎社区贡献！请查看 [贡献指南](CONTRIBUTING.md) 了解如何参与项目开发。

## 支持

- 📖 [项目文档](https://docs.ossmanager.io)
- 🐛 [问题报告](https://github.com/myysophia/ossmanager/issues)
- 💬 [社区讨论](https://github.com/myysophia/ossmanager/discussions)
- 📧 [邮件支持](mailto:support@ossmanager.io)
