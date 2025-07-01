# 阿里云OSS传输加速配置指南

## 概述

阿里云OSS传输加速是一项能显著提升全球范围内文件上传和下载速度的功能。通过在全球各地部署的加速节点，为用户提供更快速、更稳定的数据传输体验。

## 前提条件

1. **在阿里云OSS控制台开启传输加速功能**
   - 登录阿里云OSS控制台
   - 选择对应的Bucket
   - 在Bucket设置中找到"传输加速"选项
   - 开启传输加速功能

2. **确保账户有足够的权限和余额**
   - 传输加速功能会产生额外费用
   - 确保RAM用户有相应的权限

## 配置说明

### 配置文件示例

在 `configs/oss.yaml` 中配置传输加速：

```yaml
aliyun_oss:
  access_key_id: ""
  access_key_secret: ""
  endpoint: "https://oss-cn-hangzhou.aliyuncs.com"  # 默认endpoint
  bucket: "your-bucket-name"
  region: "cn-hangzhou"
  upload_dir: "uploads/"
  url_expire_time: 3600
  
  # 传输加速配置
  transfer_accelerate:
    enabled: true   # 开启传输加速
    type: "global"  # 加速类型
```

### 加速类型说明

| 类型 | 域名 | 适用场景 |
|------|------|----------|
| `global` | `https://oss-accelerate.aliyuncs.com` | 全球加速，适用于全球用户访问 |
| `overseas` | `https://oss-accelerate-overseas.aliyuncs.com` | 海外加速，适用于海外用户访问 |

## 功能特点

### 1. 智能域名选择

- **开启传输加速时**：自动使用加速域名
- **关闭传输加速时**：使用区域特定域名
- **日志记录**：详细记录使用的endpoint和加速状态

### 2. 支持的操作

- ✅ 文件上传 (`UploadToBucket`)
- ✅ 分片上传初始化 (`InitMultipartUploadToBucket`)
- ✅ 分片上传完成 (`CompleteMultipartUploadToBucket`)
- ✅ 详细的日志记录

### 3. 日志输出示例

```log
INFO: 创建OSS客户端 
  endpoint=https://oss-accelerate.aliyuncs.com 
  regionCode=cn-hangzhou 
  transferAccelerate=true 
  accelerateType=global
```

## 配置步骤

### 1. 在阿里云控制台开启传输加速

1. 登录[阿里云OSS控制台](https://oss.console.aliyun.com/)
2. 选择目标Bucket
3. 进入 `传输管理` > `传输加速`
4. 开启传输加速功能

### 2. 修改应用配置

在 `configs/oss.yaml` 中设置：

```yaml
transfer_accelerate:
  enabled: true    # 开启传输加速
  type: "global"   # 选择加速类型
```

### 3. 重启应用

重启应用以使配置生效。

## 注意事项

### 费用说明

- 传输加速会产生额外的费用
- 费用按实际使用的加速流量计算
- 具体费用请参考[阿里云OSS传输加速计费文档](https://help.aliyun.com/document_detail/173302.html)

### 兼容性说明

- 传输加速与现有的SDK完全兼容
- 无需修改客户端代码
- 通过配置即可开启/关闭

### 性能优化

- 建议在全球用户较多时开启传输加速
- 可以根据实际测试结果决定是否开启
- 支持随时开启或关闭

## 故障排除

### 常见问题

1. **配置了传输加速但没有生效**
   - 检查 `transfer_accelerate.enabled` 是否为 `true`
   - 确认Bucket是否已在控制台开启传输加速
   - 查看日志中的endpoint是否为加速域名

2. **上传失败**
   - 检查网络连接
   - 确认AccessKey权限
   - 查看详细的错误日志

3. **速度没有明显提升**
   - 传输加速效果因地区而异
   - 小文件可能效果不明显
   - 建议使用大文件进行测试

### 调试方法

查看日志中的关键信息：

```log
# 确认endpoint类型
INFO: 创建OSS客户端 endpoint=https://oss-accelerate.aliyuncs.com

# 确认加速状态
transferAccelerate=true accelerateType=global
```

## 最佳实践

1. **全球业务**：使用 `global` 类型
2. **主要面向海外**：使用 `overseas` 类型  
3. **测试环境**：可以关闭传输加速以节省费用
4. **生产环境**：根据用户分布和成本考虑开启

## 更多信息

- [阿里云OSS传输加速官方文档](https://help.aliyun.com/document_detail/31863.html)
- [传输加速计费说明](https://help.aliyun.com/document_detail/173302.html)
- [传输加速性能测试](https://help.aliyun.com/document_detail/31864.html) 