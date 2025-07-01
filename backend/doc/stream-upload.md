# 流式文件上传功能

## 概述

OSS Manager 后端现在支持两种文件上传方式：

1. **表单上传** (`multipart/form-data`) - 传统的表单文件上传方式
2. **流式上传** (`application/octet-stream`) - 直接流式传输文件内容，支持真正的实时进度追踪

## 流式上传的优势

### 问题背景
传统的表单上传使用 `c.FormFile("file")` 方式，这会导致 Gin 框架在返回文件之前就将整个文件读取到内存或临时文件中，从而失去了流式上传和进度追踪的机会。

### 解决方案
新的流式上传功能直接使用 `c.Request.Body` 读取请求体，实现真正的流式上传：

- **实时进度追踪**: 在网络传输过程中就能追踪上传进度
- **内存效率**: 不需要将整个文件加载到内存
- **更好的用户体验**: 能够提供准确的实时上传进度

## API 接口

### 端点
```
POST /api/v1/files/upload
```

### 请求方式自动检测
系统会根据 `Content-Type` 自动选择上传方式：
- `Content-Type: multipart/form-data` → 表单上传
- `Content-Type: application/octet-stream` → 流式上传

### 流式上传请求头

| 头部字段 | 类型 | 必需 | 说明 |
|---------|------|------|------|
| `Content-Type` | string | 是 | 必须设置为 `application/octet-stream` |
| `Content-Length` | number | 是 | 文件大小（字节） |
| `X-File-Name` | string | 是 | 原始文件名 |
| `region_code` | string | 是 | OSS 区域代码，如 `cn-hangzhou` |
| `bucket_name` | string | 是 | OSS 存储桶名称 |
| `Upload-Task-ID` | string | 否 | 上传任务ID，用于进度追踪 |
| `Authorization` | string | 是 | 认证令牌 |

### 请求体
原始二进制文件数据

## 使用示例

### 1. curl 命令行示例

```bash
#!/bin/bash

FILE_PATH="/path/to/your/file.pdf"
FILE_NAME="example.pdf"
FILE_SIZE=$(stat -c%s "$FILE_PATH")
TASK_ID=$(uuidgen)

curl -X POST \
  -H "Content-Type: application/octet-stream" \
  -H "X-File-Name: $FILE_NAME" \
  -H "Content-Length: $FILE_SIZE" \
  -H "region_code: cn-hangzhou" \
  -H "bucket_name: my-bucket" \
  -H "Upload-Task-ID: $TASK_ID" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  --data-binary "@$FILE_PATH" \
  "http://localhost:8080/api/v1/files/upload"
```

### 2. JavaScript 前端示例

```javascript
async function uploadFileStream(file, regionCode, bucketName, token) {
  const taskId = crypto.randomUUID();
  
  // 开始上传任务
  const response = await fetch('/api/v1/files/upload', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/octet-stream',
      'X-File-Name': file.name,
      'Content-Length': file.size.toString(),
      'region_code': regionCode,
      'bucket_name': bucketName,
      'Upload-Task-ID': taskId,
      'Authorization': `Bearer ${token}`
    },
    body: file
  });
  
  return { response, taskId };
}

// 使用示例
const fileInput = document.querySelector('#file-input');
const file = fileInput.files[0];

if (file) {
  const { response, taskId } = await uploadFileStream(
    file, 
    'cn-hangzhou', 
    'my-bucket', 
    'your-token'
  );
  
  // 监控上传进度
  const progressInterval = setInterval(async () => {
    const progressResponse = await fetch(`/api/v1/uploads/${taskId}/progress`);
    const progress = await progressResponse.json();
    console.log(`上传进度: ${progress.percentage}%`);
    
    if (progress.status === 'completed' || progress.status === 'failed') {
      clearInterval(progressInterval);
    }
  }, 1000);
}
```

### 3. Python 示例

```python
import requests
import uuid
import os

def upload_file_stream(file_path, region_code, bucket_name, token):
    task_id = str(uuid.uuid4())
    file_size = os.path.getsize(file_path)
    file_name = os.path.basename(file_path)
    
    headers = {
        'Content-Type': 'application/octet-stream',
        'X-File-Name': file_name,
        'Content-Length': str(file_size),
        'region_code': region_code,
        'bucket_name': bucket_name,
        'Upload-Task-ID': task_id,
        'Authorization': f'Bearer {token}'
    }
    
    with open(file_path, 'rb') as f:
        response = requests.post(
            'http://localhost:8080/api/v1/files/upload',
            headers=headers,
            data=f
        )
    
    return response, task_id

# 使用示例
response, task_id = upload_file_stream(
    '/path/to/file.pdf',
    'cn-hangzhou',
    'my-bucket',
    'your-token'
)

print(f"上传结果: {response.status_code}")
print(f"任务ID: {task_id}")
```

## 进度追踪

流式上传支持实时进度追踪，相关 API：

### 获取上传进度
```
GET /api/v1/uploads/{task_id}/progress
```

### 实时进度流 (SSE)
```
GET /api/v1/uploads/{task_id}/stream
```

详细说明请参考 [上传进度 API 文档](upload-progress.md)。

## 技术实现细节

### ProgressReader
系统实现了一个 `ProgressReader` 来包装 `c.Request.Body`：

```go
type ProgressReader struct {
    reader   io.Reader
    total    int64
    read     int64
    callback func(read, total int64)
}

func (pr *ProgressReader) Read(p []byte) (n int, err error) {
    n, err = pr.reader.Read(p)
    pr.read += int64(n)
    if pr.callback != nil {
        pr.callback(pr.read, pr.total)
    }
    return n, err
}
```

### 上传流程
1. 客户端发送包含文件元数据的请求头
2. 服务端验证权限和参数
3. 创建 `ProgressReader` 包装请求体
4. 在读取过程中实时更新上传进度
5. 将数据流式传输到 OSS

### 兼容性
- 保持对原有表单上传方式的完全兼容
- 根据 `Content-Type` 自动选择上传方式
- 现有客户端无需修改即可继续使用

## 测试

使用提供的测试脚本进行流式上传测试：

```bash
# 给脚本添加执行权限
chmod +x test/shell/stream_upload_test.sh

# 执行测试
./test/shell/stream_upload_test.sh /path/to/file.pdf cn-hangzhou my-bucket
```

该脚本会：
1. 验证文件存在性
2. 自动获取文件信息
3. 生成任务ID
4. 启动后台进度监控
5. 执行流式上传
6. 显示最终结果

## 注意事项

1. **请求头必需**: 流式上传需要正确设置所有必需的请求头
2. **文件大小限制**: 受服务器配置和OSS限制影响
3. **网络稳定性**: 大文件上传时需要考虑网络中断的处理
4. **进度精度**: 进度更新频率取决于数据读取速度
5. **错误处理**: 上传失败时需要清理进度状态

## 故障排除

### 常见错误

1. **"请在 X-File-Name 头中指定文件名"**
   - 检查是否设置了 `X-File-Name` 请求头

2. **"请求头中缺少 Content-Length"**
   - 确保设置了正确的 `Content-Length` 请求头

3. **"无效的 Content-Length"**
   - 检查 `Content-Length` 值是否为有效数字

4. **进度不更新**
   - 检查任务ID是否正确
   - 确认上传是否真正开始

5. **上传中断**
   - 检查网络连接
   - 验证文件权限和大小限制 