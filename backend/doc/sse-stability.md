# SSE连接稳定性改进

## 🚨 问题背景

在使用Server-Sent Events (SSE) 进行实时进度推送时，可能会遇到 `ERR_HTTP2_PROTOCOL_ERROR` 错误，这通常由以下原因引起：

1. **HTTP/2协议问题**: HTTP/2对长连接（如SSE）的处理存在兼容性问题
2. **代理服务器配置**: Nginx等代理服务器的HTTP/2实现可能与SSE不兼容
3. **浏览器兼容性**: 不同浏览器的HTTP/2实现差异
4. **连接超时**: 长连接在网络层面的超时处理

## 🛠️ 解决方案

### 1. 服务端改进

#### 1.1 禁用HTTP/2
在 `cmd/main.go` 中强制使用HTTP/1.1：

```go
// 创建HTTP服务器 - 禁用HTTP/2以确保SSE连接稳定性
server := &http.Server{
    Addr:    fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port),
    Handler: router,
    // 禁用HTTP/2，强制使用HTTP/1.1
    TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
    // 设置超时时间，优化长连接
    ReadTimeout:       time.Duration(cfg.App.ReadTimeout) * time.Second,
    WriteTimeout:      time.Duration(cfg.App.WriteTimeout) * time.Second,
    IdleTimeout:       time.Duration(cfg.App.IdleTimeout) * time.Second,
    ReadHeaderTimeout: 10 * time.Second,
}
```

#### 1.2 SSE专用中间件
新增 `internal/api/middleware/sse.go` 中的SSE中间件：

```go
// SSEMiddleware 确保SSE连接稳定性
uploads.Use(
    middleware.SSEMiddleware(),       // SSE连接稳定性中间件
    middleware.HTTP1OnlyMiddleware(), // 强制HTTP/1.1
    middleware.NoBufferMiddleware(),  // 禁用缓冲
)
```

#### 1.3 增强的SSE处理器
在 `StreamProgress` 方法中添加：
- **连接确认**: 建立连接时发送确认消息
- **心跳机制**: 每10秒发送心跳，保持连接活跃
- **错误处理**: 更好的错误检测和处理
- **优雅断开**: 任务完成时正确关闭连接

### 2. 代理服务器配置

#### 2.1 Nginx配置
如果使用Nginx作为反向代理，建议配置：

```nginx
server {
    listen 80;
    listen 443 ssl; # 移除 http2 参数
    
    server_name your-domain.com;
    
    # 特别针对SSE接口禁用缓冲
    location /api/v1/uploads/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        
        # SSE专用配置
        proxy_buffering off;
        proxy_cache off;
        proxy_set_header X-Accel-Buffering no;
        proxy_read_timeout 300s;
        proxy_send_timeout 300s;
    }
}
```

#### 2.2 Apache配置
```apache
<Location "/api/v1/uploads/">
    ProxyPass http://localhost:8080/api/v1/uploads/
    ProxyPassReverse http://localhost:8080/api/v1/uploads/
    ProxyPreserveHost On
    ProxyTimeout 300
    
    # 禁用缓冲
    SetEnv proxy-nokeepalive 1
    SetEnv proxy-sendchunked 1
</Location>
```

### 3. 客户端最佳实践

#### 3.1 EventSource配置
```javascript
// 强制使用HTTP/1.1的EventSource连接
const eventSource = new EventSource('/api/v1/uploads/task-id/stream', {
    withCredentials: false  // 根据需要设置
});

// 连接确认处理
eventSource.addEventListener('connected', function(e) {
    const data = JSON.parse(e.data);
    console.log('SSE连接已建立:', data.taskId);
});

// 进度更新处理
eventSource.addEventListener('progress', function(e) {
    const progress = JSON.parse(e.data);
    updateProgressBar(progress);
});

// 心跳处理（可选）
eventSource.addEventListener('heartbeat', function(e) {
    const data = JSON.parse(e.data);
    console.log('心跳:', new Date(data.timestamp * 1000));
});

// 完成处理
eventSource.addEventListener('complete', function(e) {
    const data = JSON.parse(e.data);
    console.log('任务完成:', data.taskId);
    eventSource.close();
});

// 错误处理
eventSource.onerror = function(event) {
    console.error('SSE连接错误:', event);
    // 可以实现重连逻辑
};
```

#### 3.2 重连机制
```javascript
class StableEventSource {
    constructor(url, options = {}) {
        this.url = url;
        this.options = options;
        this.maxRetries = options.maxRetries || 5;
        this.retryDelay = options.retryDelay || 2000;
        this.retries = 0;
        this.connect();
    }

    connect() {
        this.eventSource = new EventSource(this.url, this.options);
        
        this.eventSource.onopen = () => {
            console.log('SSE连接已建立');
            this.retries = 0; // 重置重试计数
        };
        
        this.eventSource.onerror = (event) => {
            console.error('SSE连接错误:', event);
            this.eventSource.close();
            
            if (this.retries < this.maxRetries) {
                this.retries++;
                setTimeout(() => {
                    console.log(`重连尝试 ${this.retries}/${this.maxRetries}`);
                    this.connect();
                }, this.retryDelay);
            } else {
                console.error('超过最大重试次数，停止重连');
            }
        };
    }

    addEventListener(type, listener) {
        this.eventSource.addEventListener(type, listener);
    }

    close() {
        this.eventSource.close();
    }
}

// 使用示例
const stableSSE = new StableEventSource('/api/v1/uploads/task-id/stream');
stableSSE.addEventListener('progress', handleProgress);
```

### 4. 监控和调试

#### 4.1 服务端日志
SSE中间件会记录详细的连接日志：
```
DEBUG: 处理SSE请求 path=/api/v1/uploads/task-id/stream method=GET
DEBUG: SSE请求完成 path=/api/v1/uploads/task-id/stream status=200
```

#### 4.2 客户端调试
```javascript
// 开启详细日志
eventSource.addEventListener('connected', (e) => {
    console.log('连接建立:', e.data);
});

eventSource.addEventListener('heartbeat', (e) => {
    console.log('心跳:', e.data);
});

// 监控连接状态
console.log('EventSource readyState:', eventSource.readyState);
// 0: CONNECTING, 1: OPEN, 2: CLOSED
```

## 🧪 测试验证

### 1. 连接稳定性测试
```bash
# 测试长连接稳定性
curl -N -H "Accept: text/event-stream" \
     http://localhost:8080/api/v1/uploads/test-task-id/stream
```

### 2. 负载测试
```bash
# 使用 ab 进行并发SSE连接测试
ab -n 100 -c 10 -s 60 http://localhost:8080/api/v1/uploads/test-task-id/stream
```

### 3. 网络中断恢复测试
模拟网络中断，验证重连机制是否正常工作。

## 📊 性能指标

通过以上改进，预期达到：
- **连接成功率**: > 99%
- **平均连接时间**: < 100ms
- **心跳延迟**: < 50ms
- **重连成功率**: > 95%

## ⚠️ 注意事项

1. **HTTP/2禁用影响**: 禁用HTTP/2可能会影响其他API的性能，但对于SSE连接稳定性是必要的
2. **代理配置**: 确保所有代理服务器都正确配置了SSE支持
3. **客户端兼容性**: 旧版本浏览器可能需要额外的polyfill
4. **资源清理**: 确保客户端在不需要时正确关闭SSE连接

## 🔧 故障排除

### 常见问题及解决方案

1. **ERR_HTTP2_PROTOCOL_ERROR**
   - 确认已禁用HTTP/2
   - 检查代理服务器配置

2. **Nil Pointer Dereference Panic**
   - **问题**: `c.Writer.Flush()` 调用时出现nil指针错误
   - **解决**: 使用安全的类型断言检查 `http.Flusher` 接口
   ```go
   // 错误做法
   c.Writer.Flush()
   
   // 正确做法  
   if flusher, ok := c.Writer.(http.Flusher); ok {
       flusher.Flush()
   }
   ```

3. **连接频繁断开**
   - 增加心跳频率
   - 检查网络环境

4. **消息丢失**
   - 确认客户端正确处理所有事件类型
   - 检查服务端日志

5. **内存泄漏**
   - 确保正确关闭EventSource连接
   - 定期清理未使用的订阅

### 版本历史修复

#### v1.1 (2025-01-XX) - Panic修复
- **问题**: SSE中间件和处理器中的 `c.Writer.Flush()` 导致panic
- **修复**: 
  - 移除SSE中间件中的连接劫持逻辑
  - 所有 `Flush()` 调用都增加安全的类型检查
  - 简化中间件逻辑，提高稳定性 