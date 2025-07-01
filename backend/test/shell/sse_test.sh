#!/bin/bash

# SSE连接稳定性测试脚本
# 使用方法: ./sse_test.sh [task_id] [server_url]

set -e

# 默认参数
TASK_ID=${1:-"test-$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')"}
SERVER_URL=${2:-"http://localhost:8080"}

echo "🧪 SSE连接稳定性测试"
echo "=================================="
echo "任务ID: $TASK_ID"
echo "服务器: $SERVER_URL"
echo "测试时间: $(date)"
echo ""

# 测试1: 创建测试任务
echo "📝 步骤1: 创建测试任务..."
curl -s -X POST "$SERVER_URL/api/v1/uploads/init" \
    -H "Content-Type: application/json" \
    -d "{\"total\": 1000000}" | jq .

echo ""

# 测试2: 检查任务状态
echo "📊 步骤2: 检查任务状态..."
PROGRESS_RESPONSE=$(curl -s "$SERVER_URL/api/v1/uploads/$TASK_ID/progress")
echo "$PROGRESS_RESPONSE" | jq .

if echo "$PROGRESS_RESPONSE" | jq -e '.code == 404' > /dev/null; then
    echo "⚠️ 任务不存在，使用新建任务进行测试"
    TASK_ID="new-test-$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')"
    curl -s -X POST "$SERVER_URL/api/v1/uploads/init" \
        -H "Content-Type: application/json" \
        -d "{\"total\": 1000000}" > /dev/null
fi

echo ""

# 测试3: SSE连接测试
echo "🔗 步骤3: 测试SSE连接稳定性..."
echo "连接到: $SERVER_URL/api/v1/uploads/$TASK_ID/stream"
echo "等待10秒观察连接..."

# 创建临时文件记录SSE输出
TEMP_FILE=$(mktemp)
echo "临时日志文件: $TEMP_FILE"

# 启动SSE连接（后台运行）
timeout 10s curl -N -H "Accept: text/event-stream" \
    "$SERVER_URL/api/v1/uploads/$TASK_ID/stream" > "$TEMP_FILE" 2>&1 &

SSE_PID=$!
echo "SSE进程ID: $SSE_PID"

# 等待一段时间
sleep 3

# 检查SSE进程是否还在运行
if ps -p $SSE_PID > /dev/null 2>&1; then
    echo "✅ SSE连接正常建立并保持"
else
    echo "❌ SSE连接已断开"
fi

# 等待测试完成
wait $SSE_PID 2>/dev/null || true

echo ""
echo "📋 SSE连接日志:"
echo "---------------"
cat "$TEMP_FILE"
echo "---------------"

# 分析日志
echo ""
echo "📈 连接质量分析:"

# 检查连接建立
if grep -q "connected" "$TEMP_FILE"; then
    echo "✅ 连接确认消息: 正常"
else
    echo "❌ 连接确认消息: 缺失"
fi

# 检查心跳
HEARTBEAT_COUNT=$(grep -c "heartbeat" "$TEMP_FILE" || echo "0")
echo "💓 心跳消息数量: $HEARTBEAT_COUNT"

if [ "$HEARTBEAT_COUNT" -gt 0 ]; then
    echo "✅ 心跳机制: 正常"
else
    echo "⚠️ 心跳机制: 可能异常（连接时间较短）"
fi

# 检查错误
if grep -q "error" "$TEMP_FILE"; then
    echo "❌ 发现错误信息:"
    grep "error" "$TEMP_FILE"
else
    echo "✅ 无错误信息"
fi

echo ""

# 测试4: 并发连接测试
echo "🔄 步骤4: 并发连接测试..."
echo "创建3个并发SSE连接，持续5秒..."

CONCURRENT_PIDS=()
for i in {1..3}; do
    CONCURRENT_TASK_ID="concurrent-$i-$(uuidgen 2>/dev/null || python3 -c 'import uuid; print(uuid.uuid4())')"
    
    # 创建任务
    curl -s -X POST "$SERVER_URL/api/v1/uploads/init" \
        -H "Content-Type: application/json" \
        -d "{\"total\": 1000000}" > /dev/null
    
    # 启动SSE连接
    timeout 5s curl -N -H "Accept: text/event-stream" \
        "$SERVER_URL/api/v1/uploads/$CONCURRENT_TASK_ID/stream" > "/tmp/sse_concurrent_$i.log" 2>&1 &
    
    CONCURRENT_PIDS+=($!)
    echo "启动并发连接 $i, PID: ${CONCURRENT_PIDS[$i-1]}"
done

# 等待所有并发连接完成
echo "等待并发测试完成..."
for pid in "${CONCURRENT_PIDS[@]}"; do
    wait $pid 2>/dev/null || true
done

# 分析并发结果
echo ""
echo "📊 并发连接分析:"
SUCCESS_COUNT=0
for i in {1..3}; do
    if [ -f "/tmp/sse_concurrent_$i.log" ]; then
        if grep -q "connected" "/tmp/sse_concurrent_$i.log"; then
            echo "✅ 并发连接 $i: 成功"
            SUCCESS_COUNT=$((SUCCESS_COUNT + 1))
        else
            echo "❌ 并发连接 $i: 失败"
            echo "   错误日志: $(head -n 1 /tmp/sse_concurrent_$i.log)"
        fi
        rm -f "/tmp/sse_concurrent_$i.log"
    fi
done

echo "成功率: $SUCCESS_COUNT/3 ($(( SUCCESS_COUNT * 100 / 3 ))%)"

# 清理
rm -f "$TEMP_FILE"

echo ""
echo "🎯 测试总结:"
echo "============"
if [ "$SUCCESS_COUNT" -ge 2 ]; then
    echo "✅ SSE连接稳定性: 良好"
    echo "   - 连接建立正常"
    echo "   - 并发处理良好"
    echo "   - 可以投入生产使用"
else
    echo "⚠️ SSE连接稳定性: 需要改进"
    echo "   - 请检查服务器配置"
    echo "   - 确认HTTP/2已禁用"
    echo "   - 检查代理服务器设置"
fi

echo ""
echo "📚 更多信息请参考: doc/sse-stability.md"
echo "测试完成时间: $(date)" 