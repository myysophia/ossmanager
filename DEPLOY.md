# OSS Manager Docker 部署指南

本文档详细说明如何使用 Docker 部署 OSS Manager 单体应用。

## 📋 系统要求

- Docker 20.10 或更高版本
- Docker Compose 2.0 或更高版本
- 至少 2GB 可用内存
- 至少 5GB 可用磁盘空间

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/myysophia/ossmanager.git
cd ossmanager
```

### 2. 配置环境变量

```bash
# 复制环境变量示例文件
cp .env.example .env

# 编辑配置文件
nano .env  # 或使用其他编辑器
```

**重要配置项：**

```bash
# 应用端口
APP_PORT=8080

# JWT密钥（必须修改！）
JWT_SECRET_KEY=your-super-secret-jwt-key-at-least-32-characters-long

# 数据库配置
DB_HOST=your-database-host
DB_PORT=5432
DB_USERNAME=your-db-username
DB_PASSWORD=your-db-password
DB_NAME=your-database-name
DB_SSLMODE=require

# OSS存储配置（可选）
OSS_ACCESS_KEY_ID=your-oss-access-key
OSS_ACCESS_KEY_SECRET=your-oss-secret-key
OSS_ENDPOINT=https://oss-cn-hangzhou.aliyuncs.com
OSS_REGION=cn-hangzhou
OSS_BUCKET=your-bucket-name
```

### 3. 部署应用

使用提供的部署脚本：

```bash
# 构建并启动服务
./deploy.sh build
./deploy.sh start

# 或者一次性执行
./deploy.sh build && ./deploy.sh start
```

### 4. 验证部署

```bash
# 检查服务状态
docker-compose ps

# 查看健康检查
curl http://localhost:8080/api/v1/health

# 访问前端
open http://localhost:8080
```

## 📁 部署结构

```
ossmanager/
├── Dockerfile              # Docker镜像构建文件
├── docker-compose.yml      # Docker Compose配置
├── deploy.sh               # 部署脚本
├── .env.example            # 环境变量示例
├── .env                    # 实际环境变量（需要创建）
├── configs/
│   ├── app.prod.yaml      # 生产环境配置
│   ├── app.dev.yaml       # 开发环境配置
│   └── oss.yaml           # OSS配置
└── frontend/              # 前端源码
```

## 🛠️ 部署脚本使用

`deploy.sh` 脚本提供了完整的部署管理功能：

```bash
./deploy.sh build      # 构建Docker镜像
./deploy.sh start      # 启动服务
./deploy.sh stop       # 停止服务
./deploy.sh restart    # 重启服务
./deploy.sh logs       # 查看实时日志
./deploy.sh clean      # 清理所有Docker资源
./deploy.sh help       # 显示帮助信息
```

## 🐳 手动Docker命令

如果不使用脚本，也可以手动执行：

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 清理资源
docker-compose down -v --rmi all
```

## 📊 监控和维护

### 查看服务状态

```bash
# 查看容器状态
docker-compose ps

# 查看资源使用情况
docker stats ossmanager

# 查看健康检查
docker inspect ossmanager | grep Health -A 10
```

### 日志管理

```bash
# 查看实时日志
docker-compose logs -f

# 查看最近100行日志
docker-compose logs --tail=100

# 查看特定时间段日志
docker-compose logs --since "2024-01-01T00:00:00"
```

### 数据备份

```bash
# 备份数据卷
docker run --rm -v ossmanager_ossmanager_data:/data -v $(pwd):/backup alpine tar czf /backup/ossmanager-data-backup.tar.gz /data

# 恢复数据卷
docker run --rm -v ossmanager_ossmanager_data:/data -v $(pwd):/backup alpine tar xzf /backup/ossmanager-data-backup.tar.gz -C /
```

## 🔧 配置说明

### 环境变量详解

| 变量名 | 说明 | 默认值 | 必需 |
|--------|------|--------|------|
| `APP_PORT` | 应用端口 | 8080 | 否 |
| `JWT_SECRET_KEY` | JWT签名密钥 | - | **是** |
| `DB_HOST` | 数据库主机 | - | **是** |
| `DB_PORT` | 数据库端口 | 5432 | 否 |
| `DB_USERNAME` | 数据库用户名 | - | **是** |
| `DB_PASSWORD` | 数据库密码 | - | **是** |
| `DB_NAME` | 数据库名称 | - | **是** |
| `DB_SSLMODE` | SSL模式 | require | 否 |
| `OSS_ACCESS_KEY_ID` | OSS访问密钥ID | - | 否 |
| `OSS_ACCESS_KEY_SECRET` | OSS访问密钥 | - | 否 |
| `OSS_ENDPOINT` | OSS端点 | - | 否 |
| `OSS_REGION` | OSS区域 | - | 否 |
| `OSS_BUCKET` | OSS存储桶 | - | 否 |

### Docker Compose 配置

- **端口映射**：容器8080端口映射到主机端口
- **数据持久化**：使用Docker卷存储数据、日志和上传文件
- **健康检查**：30秒间隔检查应用健康状态
- **重启策略**：`unless-stopped` 确保服务自动重启

## 🔒 安全建议

1. **修改默认JWT密钥**：确保`JWT_SECRET_KEY`至少32字符
2. **数据库连接**：使用SSL连接数据库
3. **端口限制**：生产环境中限制端口访问
4. **定期备份**：定期备份数据卷和数据库
5. **日志监控**：监控应用日志以发现异常

## 🚨 故障排除

### 常见问题

1. **端口冲突**
   ```bash
   # 检查端口占用
   lsof -i :8080
   # 修改 .env 中的 APP_PORT
   ```

2. **数据库连接失败**
   ```bash
   # 检查数据库配置
   docker-compose logs | grep database
   # 验证数据库连通性
   ```

3. **前端资源404**
   ```bash
   # 重新构建镜像
   ./deploy.sh clean
   ./deploy.sh build
   ```

4. **内存不足**
   ```bash
   # 检查资源使用
   docker stats
   # 增加系统内存或优化配置
   ```

### 日志分析

```bash
# 查看错误日志
docker-compose logs | grep ERROR

# 查看启动日志
docker-compose logs --tail=50

# 查看特定组件日志
docker-compose logs ossmanager
```

## 📈 性能优化

1. **资源限制**：在docker-compose.yml中设置内存和CPU限制
2. **数据库优化**：调整数据库连接池大小
3. **缓存策略**：配置适当的缓存头
4. **日志轮转**：配置日志轮转避免磁盘占满

## 🔄 更新升级

```bash
# 拉取最新代码
git pull origin main

# 重新构建和部署
./deploy.sh build
./deploy.sh restart
```

## 📞 技术支持

如果遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查项目的 Issues 页面
3. 提交新的 Issue 并附带详细日志

---

**注意**：首次部署前请确保已正确配置所有必需的环境变量，特别是数据库连接信息和JWT密钥。
