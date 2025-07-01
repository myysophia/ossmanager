#!/bin/bash

# 快速SSE测试脚本 - 验证panic修复
# 使用方法: ./quick_sse_test.sh [server_url]

SERVER_URL=${1:-"http://localhost:8080"}

echo "🔧 快速SSE连接测试 - 验证panic修复"
echo "======================================"
echo "服务器: $SERVER_URL"
echo "测试时间: $(date)"
echo ""

# 测试1: 创建测试任务
echo "📝 创建测试任务..."
TASK_RESPONSE=$(curl -s -X POST "$SERVER_URL/api/v1/uploads/init" \
    -H "Content-Type: application/json" \
    -d '{"total": 100000}')

echo "响应: $TASK_RESPONSE"

if echo "$TASK_RESPONSE" | jq -e '.data.id' > /dev/null 2>&1; then
    TASK_ID=$(echo "$TASK_RESPONSE" | jq -r '.data.id')
    echo "✅ 任务创建成功，ID: $TASK_ID"
else
    echo "❌ 任务创建失败，使用固定ID进行测试"
    TASK_ID="test-fixed-id"
fi

echo ""

# 测试2: 快速SSE连接测试
echo "🔗 测试SSE连接（5秒）..."
TEMP_FILE=$(mktemp)

# 5秒连接测试
timeout 5s curl -N -H "Accept: text/event-stream" \
    "$SERVER_URL/api/v1/uploads/$TASK_ID/stream" > "$TEMP_FILE" 2>&1 &

TEST_PID=$!
echo "测试进程ID: $TEST_PID"

# 等待测试完成
wait $TEST_PID 2>/dev/null || true

echo ""
echo "📋 连接结果:"
echo "------------"
if [ -s "$TEMP_FILE" ]; then
    cat "$TEMP_FILE"
    echo ""
    echo "------------"
    
    # 检查是否有panic
    if grep -q "panic" "$TEMP_FILE"; then
        echo "❌ 发现panic错误"
    elif grep -q "connected" "$TEMP_FILE"; then
        echo "✅ 连接成功建立"
    elif grep -q "data:" "$TEMP_FILE"; then
        echo "✅ 接收到SSE数据"
    else
        echo "⚠️ 未知状态，请检查输出"
    fi
else
    echo "❌ 无输出，可能连接失败"
fi

# 清理
rm -f "$TEMP_FILE"

echo ""
echo "�� 快速测试完成: $(date)" 