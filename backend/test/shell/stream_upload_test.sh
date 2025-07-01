#!/bin/bash

# 流式上传测试脚本
# 使用方法: ./stream_upload_test.sh <file_path> <region_code> <bucket_name>

set -e

# 检查参数
if [ $# -ne 3 ]; then
    echo "使用方法: $0 <file_path> <region_code> <bucket_name>"
    echo "示例: $0 /path/to/file.pdf cn-hangzhou my-bucket"
    exit 1
fi

FILE_PATH="$1"
REGION_CODE="$2"
BUCKET_NAME="$3"

# 检查文件是否存在
if [ ! -f "$FILE_PATH" ]; then
    echo "错误: 文件 $FILE_PATH 不存在"
    exit 1
fi

# 获取文件信息
FILE_NAME=$(basename "$FILE_PATH")
FILE_SIZE=$(stat -c%s "$FILE_PATH" 2>/dev/null || stat -f%z "$FILE_PATH" 2>/dev/null)

echo "开始流式上传测试..."
echo "文件: $FILE_PATH"
echo "文件名: $FILE_NAME" 
echo "文件大小: $FILE_SIZE bytes"
echo "区域: $REGION_CODE"
echo "存储桶: $BUCKET_NAME"
echo "----------------------------------------"

# 服务器地址
SERVER_URL="http://localhost:8080"

# 生成任务ID
TASK_ID=$(uuidgen 2>/dev/null || python3 -c "import uuid; print(uuid.uuid4())" 2>/dev/null || echo "stream-upload-$(date +%s)")

echo "任务ID: $TASK_ID"

# 启动进度监控（在后台）
echo "启动进度监控..."
(
    while true; do
        curl -s "$SERVER_URL/api/v1/uploads/$TASK_ID/progress" | jq '.' 2>/dev/null || true
        sleep 1
    done
) &
PROGRESS_PID=$!

# 确保在脚本退出时停止进度监控
trap "kill $PROGRESS_PID 2>/dev/null || true" EXIT

echo "开始上传文件..."

# 执行流式上传
RESPONSE=$(curl -X POST \
    -H "Content-Type: application/octet-stream" \
    -H "X-File-Name: $FILE_NAME" \
    -H "Content-Length: $FILE_SIZE" \
    -H "region_code: $REGION_CODE" \
    -H "bucket_name: $BUCKET_NAME" \
    -H "Upload-Task-ID: $TASK_ID" \
    -H "Authorization: Bearer YOUR_TOKEN_HERE" \
    --data-binary "@$FILE_PATH" \
    "$SERVER_URL/api/v1/files/upload" \
    -w "\nHTTP_CODE:%{http_code}\n")

echo "----------------------------------------"
echo "上传响应:"
echo "$RESPONSE"

# 停止进度监控
kill $PROGRESS_PID 2>/dev/null || true

# 获取最终进度
echo "----------------------------------------"
echo "最终进度:"
curl -s "$SERVER_URL/api/v1/uploads/$TASK_ID/progress" | jq '.' 2>/dev/null || echo "无法获取进度信息"

echo "流式上传测试完成！" 