# 本地开发环境启动指南

## 🎯 当前项目状态

✅ **已完成功能**:
- WebDAV Token管理API
- WebDAV管理界面（前端页面）
- 用户认证和权限管理

⚠️ **待完成功能**:
- WebDAV协议服务端点（被注释掉了）

## 📋 前置条件

1. **Go 1.21+**
2. **Node.js 18+** 和 **npm**
3. **PostgreSQL** 数据库运行在 localhost:5432
4. **环境配置文件** (.env, configs/*.yaml)

## 🚀 本地启动步骤

### 1. 准备数据库
```bash
# 确保PostgreSQL运行
brew services start postgresql  # MacOS
# 或 sudo systemctl start postgresql  # Linux

# 确保数据库存在
psql -c "CREATE DATABASE ossmanager;" -U postgres || echo "数据库已存在"
```

### 2. 启动后端 (端口 8081)
```bash
# 在项目根目录
cd /Users/ninesun/GolandProjects/ossmanager

# 方式一: 开发模式 (推荐)
go run main.go

# 方式二: 编译模式
go build -o ossmanager && ./ossmanager
```

**后端启动成功后会显示**:
```
🌐 OSS Manager 简化单体服务启动成功 {"addr": ":8081"}
前端访问: http://:8081
API访问: http://:8081/api/v1  
⚠️ WebDAV功能暂时不可用
```

### 3. 启动前端 (端口 3000) - 新终端窗口
```bash
cd /Users/ninesun/GolandProjects/ossmanager/frontend

# 安装依赖 (首次运行)
npm install

# 启动开发服务器
npm run dev
```

**前端启动成功后会显示**:
```
- Local:        http://localhost:3000
- Network:      http://192.168.x.x:3000
```

## 🌐 访问地址说明

### 开发模式 (当前状态)
| 功能 | 地址 | 说明 |
|------|------|------|
| **前端主页** | http://localhost:3000 | Next.js开发服务器 |
| **WebDAV管理页面** | http://localhost:3000/main/webdav | Token管理界面 |
| **后端API** | http://localhost:8081/api/v1 | RESTful API |
| **WebDAV服务** | ❌ 暂未启用 | 需要启用注释的代码 |

### 生产模式 (单体部署)
| 功能 | 地址 | 说明 |
|------|------|------|
| **应用入口** | http://localhost:8081 | 包含前后端 |
| **WebDAV管理页面** | http://localhost:8081/main/webdav | Token管理界面 |
| **后端API** | http://localhost:8081/api/v1 | RESTful API |
| **WebDAV服务** | http://localhost:8081/webdav/{bucket} | 当启用时 |

## 🔧 开发工作流

### 日常开发
1. **启动顺序**: 先后端 → 再前端
2. **前端修改**: 自动热重载
3. **后端修改**: 需手动重启 `go run main.go`
4. **访问应用**: http://localhost:3000

### API测试
- **直接访问**: http://localhost:8081/api/v1/user/current
- **前端请求**: 自动代理到8081端口 (通过CORS)

## 🧪 功能测试

### WebDAV Token管理测试
1. 访问: http://localhost:3000/main/webdav
2. 点击"创建新令牌"
3. 选择存储桶，设置过期时间
4. 生成并复制Token
5. 查看Token列表和状态

### API直接测试
```bash
# 创建Token (需要先登录获取JWT)
curl -X POST http://localhost:8081/api/v1/webdav/token \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"bucket": "test-bucket", "expires_in": 24}'

# 列出Token
curl http://localhost:8081/api/v1/webdav/token \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

## ⚠️ 当前限制

1. **WebDAV协议服务未启用**: 
   - router.go 中183-196行被注释
   - 需要使用 `go build -tags webdav` 编译才能启用

2. **Token管理功能完整**: 
   - ✅ 可以创建、列出、删除Token
   - ✅ 前端管理界面完整
   - ❌ 实际WebDAV协议端点不可用

## 🎯 下一步工作

要完整启用WebDAV功能，需要:
1. 取消注释 router.go 中的WebDAV路由
2. 使用 `go build -tags webdav` 编译
3. 配置WebDAV中间件和处理器
4. 测试WebDAV客户端连接

## 🐛 常见问题

**Q: 端口冲突怎么办？**
```bash
# 查看端口占用
lsof -i :3000  # 前端
lsof -i :8081  # 后端

# 杀掉进程
kill -9 <PID>
```

**Q: 数据库连接失败？**
- 检查PostgreSQL是否运行
- 验证.env文件中的数据库配置
- 确保数据库ossmanager已创建

**Q: 前端API请求失败？**  
- 确保后端在8081端口运行
- 检查浏览器控制台网络请求
- 验证CORS配置是否正确

**Q: WebDAV为什么不可用？**
- 当前WebDAV协议服务被注释掉了
- 只有Token管理功能可用
- 需要额外的构建标签来启用
