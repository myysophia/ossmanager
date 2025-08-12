# 🔒 OSS Manager 安全指南

## 概述

本文档描述了 OSS Manager 项目的安全最佳实践和配置指南。

## ⚠️ 重要安全原则

### 1. **永远不要提交敏感信息到代码仓库**

❌ **禁止提交的内容：**
- 数据库密码
- API 密钥 (如阿里云 AccessKey、AWS 凭据)
- JWT 密钥
- 真实的服务器 IP 地址
- 生产环境配置文件

✅ **安全的做法：**
- 使用环境变量占位符: `${VAR_NAME}`
- 提供 `.env.example` 文件作为模板
- 在 `.gitignore` 中排除 `.env` 文件

### 2. **配置文件安全**

我们的配置文件已经安全化，使用环境变量替代硬编码值：

```yaml
# ✅ 安全的配置
jwt:
  secret_key: "${JWT_SECRET_KEY}"

database:
  host: "${DB_HOST:-localhost}"
  password: "${DB_PASSWORD}"

# ❌ 不安全的配置
jwt:
  secret_key: "hardcoded-secret-key"
  
database:
  host: "192.168.1.100"
  password: "mypassword123"
```

## 🛡️ 环境变量管理

### 开发环境

1. **复制环境变量模板：**
   ```bash
   cp .env.example .env
   ```

2. **编辑 `.env` 文件，设置真实值：**
   ```bash
   # 必需的变量
   JWT_SECRET_KEY=your-secure-jwt-key-at-least-32-characters
   POSTGRES_PASSWORD=your-secure-database-password
   
   # 可选的 OSS 配置
   ALIYUN_ACCESS_KEY_ID=your-access-key-id
   ALIYUN_ACCESS_KEY_SECRET=your-access-key-secret
   ```

### 生产环境

使用以下方法之一管理生产环境的敏感信息：

1. **Docker Compose（推荐）：**
   ```bash
   # 在服务器上创建 .env 文件
   echo "JWT_SECRET_KEY=your-production-jwt-key" > .env
   echo "POSTGRES_PASSWORD=your-production-db-password" >> .env
   ```

2. **系统环境变量：**
   ```bash
   export JWT_SECRET_KEY="your-production-jwt-key"
   export POSTGRES_PASSWORD="your-production-db-password"
   ```

3. **密钥管理服务：**
   - AWS Secrets Manager
   - Azure Key Vault
   - HashiCorp Vault
   - Kubernetes Secrets

## 🔍 安全检查工具

我们提供了自动化安全检查脚本：

```bash
# 运行安全检查
./scripts/security-check.sh
```

这个脚本会：
- 扫描配置文件中的硬编码敏感信息
- 检查是否意外提交了真实凭据
- 验证环境变量的正确使用

## 📋 安全检查清单

### 提交代码前

- [ ] 运行 `./scripts/security-check.sh` 确保通过
- [ ] 确认没有 `.env` 文件被意外添加到 Git
- [ ] 检查配置文件只包含环境变量占位符
- [ ] 验证 `.gitignore` 包含所有敏感文件模式

### 部署前

- [ ] 确认生产环境的 `.env` 文件包含所有必需变量
- [ ] 验证 JWT 密钥至少 32 字符且随机生成
- [ ] 确认数据库密码强度足够
- [ ] 检查 OSS 访问权限是否最小化

### 运行时

- [ ] 定期轮换 JWT 密钥和数据库密码
- [ ] 监控访问日志以发现异常行为
- [ ] 保持依赖项更新以修复安全漏洞

## 🔐 密钥生成建议

### JWT 密钥生成

```bash
# 生成安全的 JWT 密钥
openssl rand -base64 32

# 或者
node -e "console.log(require('crypto').randomBytes(32).toString('base64'))"
```

### 数据库密码生成

```bash
# 生成强密码
openssl rand -base64 24
```

## 🚨 已知安全风险

### 已修复的问题

1. ✅ **配置文件硬编码凭据** - 已替换为环境变量
2. ✅ **Supabase 连接字符串泄露** - 已移除
3. ✅ **阿里云 OSS 密钥泄露** - 已移除
4. ✅ **生产数据库凭据暴露** - 已参数化

### 当前安全措施

1. **环境变量隔离** - 所有敏感配置通过环境变量管理
2. **自动安全扫描** - 提供安全检查脚本
3. **Git 忽略规则** - 防止意外提交敏感文件
4. **配置模板** - 提供安全的配置示例

## 📞 报告安全问题

如果发现安全问题，请：

1. **不要**在公开 Issue 中报告安全漏洞
2. 发送邮件到项目维护者
3. 描述问题的详细信息和影响范围
4. 提供复现步骤（如果适用）

## 🔄 安全更新流程

1. **监控依赖漏洞**
   ```bash
   # Go 依赖检查
   go list -json -deps | nancy sleuth
   
   # 前端依赖检查
   npm audit
   ```

2. **定期更新**
   - 每月检查并更新依赖
   - 关注安全公告
   - 测试更新后的功能

3. **安全测试**
   - 渗透测试
   - 代码安全审计
   - 配置安全检查

---

**记住：安全是一个持续的过程，不是一次性的任务。请定期审查和更新安全措施。** 🛡️
