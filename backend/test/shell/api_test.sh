#!/bin/bash
# test/shell/api_test.sh

# 设置API基础URL
BASE_URL="http://localhost:8080/api/v1"
USERNAME="testuser_$(date +%s)"  # 使用时间戳确保用户名唯一
PASSWORD="Password123!"
EMAIL="test_$(date +%s)@example.com"
CONFIG_ID=""
FILE_ID=""
UPLOAD_ID=""

echo "===== OSS管理系统API功能测试 ====="

# 1. 用户注册
echo -e "\n1. 测试用户注册..."
REGISTER_RESULT=$(curl -s -X POST "${BASE_URL}/auth/register" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\",
    \"email\": \"${EMAIL}\",
    \"real_name\": \"测试用户\"
  }")
echo "注册结果: $REGISTER_RESULT"

# 2. 用户登录
echo -e "\n2. 测试用户登录..."
LOGIN_RESULT=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")
echo "登录结果: $LOGIN_RESULT"

# 提取token
TOKEN=$(echo $LOGIN_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "登录失败，无法获取token，测试终止"
  exit 1
fi
echo "获取到token: ${TOKEN:0:15}..."

# 3. 获取当前用户信息
echo -e "\n3. 测试获取用户信息..."
USER_INFO=$(curl -s -X GET "${BASE_URL}/user/current" \
  -H "Authorization: Bearer $TOKEN")
echo "用户信息: $USER_INFO"

# 4. 创建存储配置
echo -e "\n4. 测试创建存储配置..."
CONFIG_RESULT=$(curl -s -X POST "${BASE_URL}/oss/configs" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试本地存储",
    "storage_type": "local",
    "root_path": "/tmp/storage",
    "description": "测试用本地存储"
  }')
echo "创建配置结果: $CONFIG_RESULT"

# 提取配置ID
CONFIG_ID=$(echo $CONFIG_RESULT | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
if [ -z "$CONFIG_ID" ]; then
  echo "创建配置失败，无法获取配置ID，使用默认值1"
  CONFIG_ID="1"
fi
echo "获取到配置ID: $CONFIG_ID"

# 5. 测试配置连接
echo -e "\n5. 测试配置连接..."
TEST_RESULT=$(curl -s -X POST "${BASE_URL}/oss/configs/${CONFIG_ID}/test" \
  -H "Authorization: Bearer $TOKEN")
echo "配置测试结果: $TEST_RESULT"

# 6. 更新配置
echo -e "\n6. 测试更新配置..."
UPDATE_CONFIG_RESULT=$(curl -s -X PUT "${BASE_URL}/oss/configs/${CONFIG_ID}" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"name\": \"更新后的本地存储\",
    \"storage_type\": \"local\",
    \"root_path\": \"/tmp/storage_updated\",
    \"description\": \"已更新的测试用本地存储\"
  }")
echo "更新配置结果: $UPDATE_CONFIG_RESULT"

# 7. 创建测试文件
echo -e "\n7. 准备测试文件..."
TEST_FILE="/tmp/test_file_$(date +%s).txt"
echo "这是测试文件内容，用于OSS上传测试" > $TEST_FILE
echo "测试文件已创建: $TEST_FILE"

# 8. 上传文件
echo -e "\n8. 测试文件上传..."
UPLOAD_RESULT=$(curl -s -X POST "${BASE_URL}/oss/files" \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@${TEST_FILE}" \
  -F "config_id=${CONFIG_ID}")
echo "文件上传结果: $UPLOAD_RESULT"

# 提取文件ID
FILE_ID=$(echo $UPLOAD_RESULT | grep -o '"id":[0-9]*' | head -1 | cut -d':' -f2)
if [ -z "$FILE_ID" ]; then
  echo "文件上传失败，无法获取文件ID，使用默认值1"
  FILE_ID="1"
fi
echo "获取到文件ID: $FILE_ID"

# 9. 文件列表查询
echo -e "\n9. 测试文件列表查询..."
LIST_RESULT=$(curl -s -X GET "${BASE_URL}/oss/files?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
echo "文件列表查询结果: $LIST_RESULT"

# 10. 获取文件下载链接
echo -e "\n10. 测试获取文件下载链接..."
DOWNLOAD_RESULT=$(curl -s -X GET "${BASE_URL}/oss/files/${FILE_ID}/download" \
  -H "Authorization: Bearer $TOKEN")
echo "获取下载链接结果: $DOWNLOAD_RESULT"

# 11. 触发文件MD5计算
echo -e "\n11. 测试触发MD5计算..."
MD5_TRIGGER_RESULT=$(curl -s -X POST "${BASE_URL}/oss/files/${FILE_ID}/md5" \
  -H "Authorization: Bearer $TOKEN")
echo "触发MD5计算结果: $MD5_TRIGGER_RESULT"

# 12. 获取文件MD5值
echo -e "\n12. 测试获取MD5值..."
MD5_RESULT=$(curl -s -X GET "${BASE_URL}/oss/files/${FILE_ID}/md5" \
  -H "Authorization: Bearer $TOKEN")
echo "获取MD5值结果: $MD5_RESULT"

# 13. 获取配置列表
echo -e "\n13. 测试获取配置列表..."
CONFIG_LIST=$(curl -s -X GET "${BASE_URL}/oss/configs?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN")
echo "配置列表查询结果: $CONFIG_LIST"

# 14. 获取配置详情
echo -e "\n14. 测试获取配置详情..."
CONFIG_DETAIL=$(curl -s -X GET "${BASE_URL}/oss/configs/${CONFIG_ID}" \
  -H "Authorization: Bearer $TOKEN")
echo "配置详情结果: $CONFIG_DETAIL"

# 15. 设置默认配置
echo -e "\n15. 测试设置默认配置..."
DEFAULT_CONFIG_RESULT=$(curl -s -X PUT "${BASE_URL}/oss/configs/${CONFIG_ID}/default" \
  -H "Authorization: Bearer $TOKEN")
echo "设置默认配置结果: $DEFAULT_CONFIG_RESULT"

# 16. 初始化分片上传
echo -e "\n16. 测试初始化分片上传..."
INIT_RESULT=$(curl -s -X POST "${BASE_URL}/oss/multipart/init" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"file_name\": \"test_large_file.zip\",
    \"file_size\": 10485760,
    \"config_id\": ${CONFIG_ID}
  }")
echo "初始化分片上传结果: $INIT_RESULT"

# 提取上传ID和对象键
UPLOAD_ID=$(echo $INIT_RESULT | grep -o '"upload_id":"[^"]*"' | cut -d'"' -f4)
OBJECT_KEY=$(echo $INIT_RESULT | grep -o '"object_key":"[^"]*"' | cut -d'"' -f4)
if [ -z "$UPLOAD_ID" ]; then
  echo "初始化分片上传失败，无法获取上传ID，跳过分片上传测试"
else
  echo "获取到上传ID: $UPLOAD_ID"
  echo "获取到对象键: $OBJECT_KEY"
  
  # 17. 获取分片上传URL
  echo -e "\n17. 测试获取分片上传URL..."
  PART_URLS=$(curl -s -X GET "${BASE_URL}/oss/multipart/${CONFIG_ID}/urls?upload_id=${UPLOAD_ID}&part_numbers=1,2" \
    -H "Authorization: Bearer $TOKEN")
  echo "获取分片上传URL结果: $PART_URLS"
  
  # 18. 取消分片上传
  echo -e "\n18. 测试取消分片上传..."
  CANCEL_RESULT=$(curl -s -X DELETE "${BASE_URL}/oss/multipart/abort?upload_id=${UPLOAD_ID}" \
    -H "Authorization: Bearer $TOKEN")
  echo "取消分片上传结果: $CANCEL_RESULT"
fi

# 19. 删除文件
echo -e "\n19. 测试删除文件..."
DELETE_FILE_RESULT=$(curl -s -X DELETE "${BASE_URL}/oss/files/${FILE_ID}" \
  -H "Authorization: Bearer $TOKEN")
echo "删除文件结果: $DELETE_FILE_RESULT"

# 20. 删除配置
echo -e "\n20. 测试删除配置..."
DELETE_CONFIG_RESULT=$(curl -s -X DELETE "${BASE_URL}/oss/configs/${CONFIG_ID}" \
  -H "Authorization: Bearer $TOKEN")
echo "删除配置结果: $DELETE_CONFIG_RESULT"

# 清理
rm -f $TEST_FILE

echo -e "\n===== API功能测试完成 ====="
echo "用户名: $USERNAME"
echo "配置ID: $CONFIG_ID"
echo "文件ID: $FILE_ID"
echo "测试时间: $(date)" 