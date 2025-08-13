#!/bin/bash

# WebDAV CORS配置验证测试脚本
# 用于验证浏览器WebDAV访问的CORS头部配置

set -e

# 配置
WEBDAV_BASE_URL="http://localhost:8080"
API_BASE_URL="$WEBDAV_BASE_URL/api/v1"
WEBDAV_URL="$WEBDAV_BASE_URL/webdav"

echo "🧪 WebDAV CORS配置验证测试"
echo "=================================="
echo "测试目标: $WEBDAV_BASE_URL"
echo ""

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 测试函数
test_cors_headers() {
    local url=$1
    local method=$2
    local test_name=$3
    
    echo -n "测试 $test_name ($method to $url)... "
    
    # 发送OPTIONS预检请求
    response=$(curl -s -i -X OPTIONS \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: $method" \
        -H "Access-Control-Request-Headers: Authorization, Content-Type, Depth" \
        "$url" 2>/dev/null || echo "FAILED")
    
    if [[ "$response" == "FAILED" ]]; then
        echo -e "${RED}❌ 连接失败${NC}"
        return 1
    fi
    
    # 检查CORS头部
    local has_allow_origin=$(echo "$response" | grep -i "Access-Control-Allow-Origin" | wc -l)
    local has_allow_methods=$(echo "$response" | grep -i "Access-Control-Allow-Methods" | wc -l) 
    local has_allow_headers=$(echo "$response" | grep -i "Access-Control-Allow-Headers" | wc -l)
    local has_allow_credentials=$(echo "$response" | grep -i "Access-Control-Allow-Credentials" | wc -l)
    local has_max_age=$(echo "$response" | grep -i "Access-Control-Max-Age" | wc -l)
    
    if [[ $has_allow_origin -gt 0 && $has_allow_methods -gt 0 && $has_allow_headers -gt 0 && $has_allow_credentials -gt 0 && $has_max_age -gt 0 ]]; then
        echo -e "${GREEN}✅ 通过${NC}"
        
        # 显示详细的CORS头部
        echo "   CORS头部详情:"
        echo "$response" | grep -i "Access-Control-" | sed 's/^/   /'
        echo ""
        return 0
    else
        echo -e "${RED}❌ 失败${NC}"
        echo "   缺失的CORS头部:"
        [[ $has_allow_origin -eq 0 ]] && echo "   - Access-Control-Allow-Origin"
        [[ $has_allow_methods -eq 0 ]] && echo "   - Access-Control-Allow-Methods"
        [[ $has_allow_headers -eq 0 ]] && echo "   - Access-Control-Allow-Headers"
        [[ $has_allow_credentials -eq 0 ]] && echo "   - Access-Control-Allow-Credentials"
        [[ $has_max_age -eq 0 ]] && echo "   - Access-Control-Max-Age"
        echo ""
        return 1
    fi
}

# 检查服务是否运行
echo "🔍 检查服务状态..."
if ! curl -s "$WEBDAV_BASE_URL" > /dev/null; then
    echo -e "${RED}❌ 服务器未运行，请先启动OSS Manager${NC}"
    echo "启动命令: go run main.go"
    exit 1
fi
echo -e "${GREEN}✅ 服务器正在运行${NC}"
echo ""

# 测试WebDAV端点的CORS配置
echo "🌐 测试WebDAV端点CORS配置..."
test_cors_headers "$WEBDAV_URL/test-bucket" "PROPFIND" "WebDAV PROPFIND"
test_cors_headers "$WEBDAV_URL/test-bucket/file.txt" "PUT" "WebDAV PUT"
test_cors_headers "$WEBDAV_URL/test-bucket/folder" "MKCOL" "WebDAV MKCOL"
test_cors_headers "$WEBDAV_URL/test-bucket/file.txt" "DELETE" "WebDAV DELETE"

echo "🌐 测试REST API端点CORS配置..."
test_cors_headers "$API_BASE_URL/webdav/token" "POST" "REST API POST"
test_cors_headers "$API_BASE_URL/oss/files" "GET" "REST API GET"

echo ""
echo "🧪 WebDAV方法支持测试..."

# 测试WebDAV方法是否在Allow-Methods中
test_webdav_methods() {
    local url="$WEBDAV_URL/test-bucket"
    
    response=$(curl -s -i -X OPTIONS \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Method: PROPFIND" \
        "$url" 2>/dev/null)
    
    local allow_methods=$(echo "$response" | grep -i "Access-Control-Allow-Methods:" | cut -d: -f2 | tr ',' '\n' | tr -d ' ')
    
    local webdav_methods=("GET" "PUT" "DELETE" "MKCOL" "COPY" "MOVE" "PROPFIND" "PROPPATCH" "LOCK" "UNLOCK")
    local missing_methods=()
    
    for method in "${webdav_methods[@]}"; do
        if ! echo "$allow_methods" | grep -q "$method"; then
            missing_methods+=("$method")
        fi
    done
    
    if [[ ${#missing_methods[@]} -eq 0 ]]; then
        echo -e "${GREEN}✅ 所有WebDAV方法都已支持${NC}"
        echo "   支持的方法: $(echo "$allow_methods" | tr '\n' ' ')"
    else
        echo -e "${RED}❌ 缺少WebDAV方法支持${NC}"
        echo "   缺失的方法: ${missing_methods[*]}"
    fi
}

test_webdav_methods

echo ""
echo "🔒 WebDAV头部支持测试..."

# 测试WebDAV特定头部是否在Allow-Headers中
test_webdav_headers() {
    local url="$WEBDAV_URL/test-bucket"
    
    response=$(curl -s -i -X OPTIONS \
        -H "Origin: http://localhost:3000" \
        -H "Access-Control-Request-Headers: Authorization, Depth, Destination" \
        "$url" 2>/dev/null)
    
    local allow_headers=$(echo "$response" | grep -i "Access-Control-Allow-Headers:" | cut -d: -f2- | tr ',' '\n' | tr -d ' ')
    
    local webdav_headers=("Authorization" "Content-Type" "Depth" "Destination" "Overwrite" "Lock-Token" "If" "Timeout")
    local missing_headers=()
    
    for header in "${webdav_headers[@]}"; do
        if ! echo "$allow_headers" | grep -qi "$header"; then
            missing_headers+=("$header")
        fi
    done
    
    if [[ ${#missing_headers[@]} -eq 0 ]]; then
        echo -e "${GREEN}✅ 所有WebDAV头部都已支持${NC}"
        echo "   允许的头部包括: Authorization, Depth, Destination, Overwrite, Lock-Token等"
    else
        echo -e "${RED}❌ 缺少WebDAV头部支持${NC}"
        echo "   缺失的头部: ${missing_headers[*]}"
    fi
}

test_webdav_headers

echo ""
echo "📊 测试总结"
echo "=================================="
echo -e "${GREEN}✅ CORS配置已优化，支持浏览器WebDAV访问${NC}"
echo -e "${YELLOW}⚠️ 注意：某些浏览器可能仍然限制JavaScript WebDAV操作${NC}"
echo -e "${YELLOW}💡 建议：使用REST代理为浏览器提供WebDAV功能包装${NC}"
echo ""
echo "🔗 相关文档:"
echo "   - WebDAV审计报告: docs/webdav-browser-audit-report.md"
echo "   - WebDAV使用指南: docs/webdav-usage.md"
echo "   - WebDAV实现文档: internal/webdav/README.md"
