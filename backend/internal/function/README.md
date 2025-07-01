# 函数计算模块

本模块实现了通过阿里云函数计算来异步计算OSS文件的MD5值的功能。

## 功能介绍

当文件上传到阿里云OSS后，通过OSS的事件通知触发函数计算，函数计算会:

1. 从OSS读取文件内容
2. 流式计算文件的MD5值
3. 更新数据库中文件记录的MD5值

这种异步处理方式有以下优势:
- 不阻塞正常的文件上传流程
- 对于大文件，避免在API服务中占用过多资源
- 利用函数计算的弹性伸缩能力

## 文件说明

- `md5_calculator.go`: 包含阿里云函数计算的入口函数和MD5计算的实现

## 使用方法

### 部署函数

1. 在阿里云函数计算控制台创建一个新的函数
2. 上传该模块的代码
3. 设置函数入口为 `CalculateOSSFileMD5`
4. 配置环境变量，包括数据库连接信息和OSS访问凭证

### 配置OSS事件触发器

1. 在OSS的指定Bucket中配置事件通知
2. 选择 `PutObject` 和 `CompleteMultipartUpload` 等事件
3. 选择触发函数计算服务
4. 选择刚才创建的函数

## 开发指南

### 本地调试

为了便于本地调试，可以使用模拟的OSS事件:

```go
// 模拟OSS事件
event := function.OSSEvent{
    Events: []struct{
        EventName    string `json:"eventName"`
        EventSource  string `json:"eventSource"`
        EventTime    string `json:"eventTime"`
        EventVersion string `json:"eventVersion"`
        OSS          struct{
            Bucket struct{
                Arn  string `json:"arn"`
                Name string `json:"name"`
            } `json:"bucket"`
            Object struct{
                Key       string `json:"key"`
                Size      int64  `json:"size"`
                ETag      string `json:"eTag"`
                Type      string `json:"type"`
                URL       string `json:"url"`
                FileACL   string `json:"fileACL"`
                ObjectACL string `json:"objectACL"`
            } `json:"object"`
        } `json:"oss"`
    }{
        {
            EventName:    "ObjectCreated:PutObject",
            EventSource:  "acs:oss",
            EventTime:    "2023-01-01T00:00:00.000Z",
            EventVersion: "1.0",
            OSS: struct{
                Bucket struct{
                    Arn  string `json:"arn"`
                    Name string `json:"name"`
                } `json:"bucket"`
                Object struct{
                    Key       string `json:"key"`
                    Size      int64  `json:"size"`
                    ETag      string `json:"eTag"`
                    Type      string `json:"type"`
                    URL       string `json:"url"`
                    FileACL   string `json:"fileACL"`
                    ObjectACL string `json:"objectACL"`
                } `json:"object"`
            }{
                Bucket: struct{
                    Arn  string `json:"arn"`
                    Name string `json:"name"`
                }{
                    Arn:  "acs:oss:cn-hangzhou:123456789:test-bucket",
                    Name: "test-bucket",
                },
                Object: struct{
                    Key       string `json:"key"`
                    Size      int64  `json:"size"`
                    ETag      string `json:"eTag"`
                    Type      string `json:"type"`
                    URL       string `json:"url"`
                    FileACL   string `json:"fileACL"`
                    ObjectACL string `json:"objectACL"`
                }{
                    Key:  "test-file.txt",
                    Size: 1024,
                    ETag: "etagvalue",
                    Type: "text/plain",
                },
            },
        },
    },
}

// 调用函数
result, err := function.CalculateOSSFileMD5(context.Background(), event)
```

### 性能优化

- 使用流式读取和计算，避免将整个文件加载到内存
- 对于大文件，可以考虑分片计算MD5
- 确保函数计算的超时时间足够长，特别是对于大文件

### 错误处理

函数计算模块包含完善的错误处理和日志记录:

1. OSS连接和读取错误
2. MD5计算过程中的错误
3. 数据库连接和更新错误

所有错误都会被记录到日志中，并返回到函数计算平台，可以在控制台查看。

## 注意事项

1. 函数计算需要有访问OSS和数据库的权限
2. 对于特别大的文件，需要设置合适的函数超时时间
3. 如果文件很多，需要考虑函数计算的并发数量
4. 确保数据库连接能够处理并发请求 