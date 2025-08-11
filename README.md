# OSS Manager - 单体应用

OSS Manager 是一个统一的对象存储管理系统，支持阿里云OSS、AWS S3、CloudFlare R2等多种云存储服务。现已整合为单体应用，前后端打包在一个可执行文件中。

## 特性

### 🚀 核心功能
- **多云存储支持** - 阿里云OSS、AWS S3、CloudFlare R2
- **WebDAV支持** - 通过文件管理器直接访问云存储
- **用户权限管理** - 基于RBAC的细粒度权限控制
- **文件管理** - 上传、下载、删除、预览
- **分片上传** - 支持大文件分片上传和断点续传
- **MD5校验** - 异步文件完整性验证
- **审计日志** - 完整的操作审计记录

### 🏗️ 技术架构
- **单体应用** - 前后端集成，单文件部署
- **嵌入式前端** - 静态资源嵌入到二进制文件
- **Go后端** - 高性能、低内存占用
- **React前端** - 现代化Web界面

## 快速开始

### 📋 系统要求

- Go 1.24+
- Node.js 18+ (仅构建时需要)
- PostgreSQL 14+
- 云存储账号 (阿里云OSS、AWS S3 或 CloudFlare R2)

### 🚀 一键运行

```bash
# 1. 从源码构建
git clone https://github.com/myysophia/ossmanager.git
cd ossmanager

# 2. 安装依赖并构建
make deps
make all

# 3. 配置
cp configs/app.example.yaml configs/app.yaml
cp configs/oss.example.yaml configs/oss.yaml
# 编辑配置文件...

# 4. 启动应用
./build/ossmanager
```

应用启动后：
- Web界面: http://localhost:8080
- API接口: http://localhost:8080/api/v1
- WebDAV: http://localhost:8080/webdav/{bucket}

### 🛠️ 开发模式

```bash
# 启动开发服务器
make dev

# 监控文件变化
make watch
```

## 构建选项

```bash
# 基础构建
make all                 # 构建完整项目
make clean              # 清理构建产物
make frontend           # 仅构建前端
make backend            # 仅构建后端

# 开发相关
make dev                # 启动开发服务器
make watch              # 监控文件变化
make test               # 运行测试
make fmt                # 代码格式化
make lint               # 代码检查

# 部署相关
make build-prod         # 生产环境构建
make backend-cross      # 跨平台构建
make docker             # Docker 镜像构建
make release            # 创建发布包

# 工具相关
make deps               # 安装依赖
make install-tools      # 安装开发工具
make help               # 显示帮助信息
```

## 🐳 Docker 部署

```bash
# 构建镜像
make docker

# 运行容器
docker run -d \
  --name ossmanager \
  -p 8080:8080 \
  -v ./configs:/app/configs \
  -v ./data:/data \
  ossmanager:latest
```

## WebDAV 使用

### Windows
1. 打开文件资源管理器
2. 右键"此电脑" → "映射网络驱动器"
3. 输入地址：`http://your-server:8080/webdav/your-bucket`
4. 输入用户名和密码

### macOS
1. 打开 Finder
2. 按 Cmd+K 或菜单"前往" → "连接服务器"
3. 输入服务器地址：`http://your-server:8080/webdav/your-bucket`
4. 输入用户名和密码

### Linux
```bash
# 使用 davfs2
sudo mount -t davfs http://your-server:8080/webdav/your-bucket /mnt/webdav

# 使用 cadaver
cadaver http://your-server:8080/webdav/your-bucket
```

## 项目结构

```
ossmanager/
├── cmd/server/          # 主程序入口
├── internal/            # 内部包
│   ├── api/            # API路由和处理器
│   ├── auth/           # 认证模块
│   ├── config/         # 配置管理
│   ├── db/             # 数据库层
│   ├── embed/          # 静态文件嵌入
│   ├── oss/            # 存储服务
│   └── webdav/         # WebDAV支持
├── web/build/          # 前端构建产物
├── configs/            # 配置文件
├── build/              # 构建产物
├── Dockerfile          # Docker构建文件
├── Makefile           # 构建脚本
└── README.md          # 项目说明
```
