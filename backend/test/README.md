# OSS管理系统接口压测方案

## 压测目标

对OSS管理系统的核心接口进行压力测试，评估系统在高并发场景下的性能表现，确保系统具备足够的稳定性和可扩展性。

## 接口功能测试

在进行压测前，推荐先对各个接口进行功能单元测试，确保接口功能正常。以下是使用curl命令对各接口进行功能测试的示例：

### 认证接口

#### 用户注册

```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Password123!",
    "email": "test@example.com",
    "real_name": "测试用户"
  }'
```

#### 用户登录

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Password123!"
  }'
```

登录成功后会返回包含token的JSON，可以保存token用于后续请求：

```bash
# 提取并保存token到环境变量
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Password123!"
  }' | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

echo "获取到的Token: $TOKEN"
```

#### 获取当前用户信息

```bash
curl -X GET http://localhost:8080/api/v1/user/current \
  -H "Authorization: Bearer $TOKEN"
```

### OSS文件管理

#### 文件上传

```bash
# 创建一个测试文件
echo "测试文件内容" > test_file.txt

# 上传文件，config_id替换为实际的配置ID
curl -X POST http://localhost:8080/api/v1/oss/files \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@test_file.txt" \
  -F "config_id=1"
```

#### 文件列表查询

```bash
# 基本查询
curl -X GET "http://localhost:8080/api/v1/oss/files?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 带过滤条件的查询
curl -X GET "http://localhost:8080/api/v1/oss/files?page=1&page_size=10&file_name=test&storage_type=s3" \
  -H "Authorization: Bearer $TOKEN"
```

#### 文件删除

```bash
# 替换file_id为实际的文件ID
curl -X DELETE http://localhost:8080/api/v1/oss/files/1 \
  -H "Authorization: Bearer $TOKEN"
```

#### 获取文件下载链接

```bash
# 替换file_id为实际的文件ID
curl -X GET http://localhost:8080/api/v1/oss/files/1/download \
  -H "Authorization: Bearer $TOKEN"
```

#### 触发MD5计算

```bash
# 替换file_id为实际的文件ID
curl -X POST http://localhost:8080/api/v1/oss/files/1/md5 \
  -H "Authorization: Bearer $TOKEN"
```

#### 获取文件MD5值

```bash
# 替换file_id为实际的文件ID
curl -X GET http://localhost:8080/api/v1/oss/files/1/md5 \
  -H "Authorization: Bearer $TOKEN"
```

### OSS配置管理

#### 创建配置

```bash
# 创建S3配置
curl -X POST http://localhost:8080/api/v1/oss/configs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试S3配置",
    "storage_type": "s3",
    "access_key": "your-access-key",
    "secret_key": "your-secret-key",
    "region": "us-east-1",
    "bucket": "test-bucket",
    "endpoint": "https://s3.amazonaws.com",
    "description": "用于测试的S3配置"
  }'

# 创建本地存储配置
curl -X POST http://localhost:8080/api/v1/oss/configs \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "本地存储配置",
    "storage_type": "local",
    "root_path": "/data/storage",
    "description": "本地文件系统存储"
  }'
```

#### 更新配置

```bash
# 替换config_id为实际的配置ID
curl -X PUT http://localhost:8080/api/v1/oss/configs/1 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新后的S3配置",
    "storage_type": "s3",
    "access_key": "updated-access-key",
    "secret_key": "updated-secret-key",
    "region": "us-west-1",
    "bucket": "new-test-bucket",
    "endpoint": "https://s3.amazonaws.com",
    "description": "已更新的S3配置"
  }'
```

#### 删除配置

```bash
# 替换config_id为实际的配置ID
curl -X DELETE http://localhost:8080/api/v1/oss/configs/1 \
  -H "Authorization: Bearer $TOKEN"
```

#### 获取配置列表

```bash
# 基本查询
curl -X GET "http://localhost:8080/api/v1/oss/configs?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"

# 带过滤条件的查询
curl -X GET "http://localhost:8080/api/v1/oss/configs?page=1&page_size=10&name=测试&storage_type=s3" \
  -H "Authorization: Bearer $TOKEN"
```

#### 获取单个配置详情

```bash
# 替换config_id为实际的配置ID
curl -X GET http://localhost:8080/api/v1/oss/configs/1 \
  -H "Authorization: Bearer $TOKEN"
```

#### 设置默认配置

```bash
# 替换config_id为实际的配置ID
curl -X PUT http://localhost:8080/api/v1/oss/configs/1/default \
  -H "Authorization: Bearer $TOKEN"
```

#### 测试配置连接

```bash
# 测试已有配置连接
curl -X POST http://localhost:8080/api/v1/oss/configs/1/test \
  -H "Authorization: Bearer $TOKEN"

# 测试新配置连接而不保存
curl -X POST http://localhost:8080/api/v1/oss/configs/test \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "storage_type": "s3",
    "access_key": "test-access-key",
    "secret_key": "test-secret-key",
    "region": "ap-northeast-1",
    "bucket": "test-bucket",
    "endpoint": "https://s3.amazonaws.com"
  }'
```

### 自动化测试脚本

以下是将上述所有接口集成到一个自动化测试脚本中的示例：

```bash
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
  CANCEL_RESULT=$(curl -s -X DELETE "${BASE_URL}/oss/multipart/abort?config_id=${CONFIG_ID}&upload_id=${UPLOAD_ID}" \
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
```

使用方法：
```bash
mkdir -p test/shell
chmod +x test/shell/api_test.sh
./test/shell/api_test.sh
```

此脚本将自动测试所有关键接口，并以正确的顺序执行操作（先创建配置，然后上传文件，最后删除资源）。测试过程中会显示每个接口的请求结果，方便排查问题。

## 压测工具

本方案主要使用 [k6](https://k6.io/) 作为压测工具，k6是一个现代化的负载测试工具，支持HTTP/HTTPS、WebSocket等协议的压测。

### 安装k6

```bash
# macOS
brew install k6

# Ubuntu/Debian
sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# CentOS/RHEL
sudo yum install https://dl.k6.io/rpm/repo.rpm
sudo yum install k6
```

### 使用curl命令进行简单压测

如果不想安装专业压测工具，可以使用curl配合shell脚本进行简单的压测。这种方法适用于快速验证接口的可用性和简单的负载测试。

#### 1. 基本的curl压测脚本

```bash
#!/bin/bash
# test/shell/simple_load_test.sh

# 设置参数
URL="http://localhost:8080/api/v1/oss/files"
TOKEN="your-auth-token-here"
REQUESTS=100  # 总请求数
CONCURRENCY=10  # 并发数

# 创建临时文件存储结果
TEMP_FILE=$(mktemp)

# 压测函数
run_test() {
  local start_time=$(date +%s.%N)
  
  # 发送请求
  curl -s -w "%{http_code},%{time_total}\n" -o /dev/null \
    -H "Authorization: Bearer $TOKEN" \
    "$URL" >> "$TEMP_FILE"
    
  local end_time=$(date +%s.%N)
  echo "完成请求: $1, 耗时: $(echo "$end_time - $start_time" | bc) 秒"
}

# 显示开始信息
echo "开始对 $URL 进行压测"
echo "总请求数: $REQUESTS, 并发数: $CONCURRENCY"

# 记录开始时间
TEST_START=$(date +%s.%N)

# 并发执行请求
for i in $(seq 1 $REQUESTS); do
  # 控制并发数
  if [ $(jobs -r | wc -l) -ge $CONCURRENCY ]; then
    wait -n  # 等待一个任务完成
  fi
  
  run_test $i &  # 后台执行
done

# 等待所有请求完成
wait

# 记录结束时间
TEST_END=$(date +%s.%N)
TOTAL_TIME=$(echo "$TEST_END - $TEST_START" | bc)

# 分析结果
SUCCESSFUL=$(grep -c "^200," "$TEMP_FILE")
TOTAL=$(wc -l < "$TEMP_FILE")
SUCCESS_RATE=$(echo "scale=2; $SUCCESSFUL / $TOTAL * 100" | bc)

# 计算响应时间统计
echo "分析响应时间..."
TIMES=$(cut -d',' -f2 "$TEMP_FILE")
TOTAL_RESP_TIME=$(echo "$TIMES" | paste -sd+ | bc)
AVG_RESP_TIME=$(echo "scale=3; $TOTAL_RESP_TIME / $TOTAL" | bc)

# 计算RPS
RPS=$(echo "scale=2; $REQUESTS / $TOTAL_TIME" | bc)

# 输出结果
echo "=============== 压测结果 ==============="
echo "总请求数: $REQUESTS"
echo "并发数: $CONCURRENCY"
echo "总耗时: $TOTAL_TIME 秒"
echo "成功请求数: $SUCCESSFUL"
echo "成功率: $SUCCESS_RATE%"
echo "平均响应时间: $AVG_RESP_TIME 秒"
echo "每秒请求数(RPS): $RPS"
echo "========================================="

# 清理临时文件
rm "$TEMP_FILE"
```

执行权限和运行：
```bash
chmod +x test/shell/simple_load_test.sh
./test/shell/simple_load_test.sh
```

#### 2. 登录并进行文件上传压测

```bash
#!/bin/bash
# test/shell/file_upload_test.sh

# 设置参数
LOGIN_URL="http://localhost:8080/api/v1/auth/login"
UPLOAD_URL="http://localhost:8080/api/v1/oss/files"
USERNAME="admin"
PASSWORD="admin123"
CONFIG_ID="1"
REQUESTS=20  # 总请求数
CONCURRENCY=5  # 并发数
FILE_SIZE=10240  # 10KB

# 创建临时文件存储结果
TEMP_FILE=$(mktemp)
TEST_FILE=$(mktemp)

# 创建测试文件
dd if=/dev/urandom of=$TEST_FILE bs=1024 count=$((FILE_SIZE/1024)) 2>/dev/null

# 获取认证令牌
echo "登录获取Token..."
TOKEN=$(curl -s -X POST $LOGIN_URL \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | 
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "获取Token失败，请检查登录接口"
  exit 1
fi
echo "获取Token成功"

# 压测函数
run_upload_test() {
  local start_time=$(date +%s.%N)
  
  # 发送上传请求
  local response=$(curl -s -w ",HTTP_CODE:%{http_code},TIME:%{time_total}" \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@$TEST_FILE" \
    -F "config_id=$CONFIG_ID" \
    "$UPLOAD_URL")
    
  local end_time=$(date +%s.%N)
  local duration=$(echo "$end_time - $start_time" | bc)
  
  # 提取HTTP状态码和响应时间
  local http_code=$(echo "$response" | grep -o 'HTTP_CODE:[0-9]*' | cut -d':' -f2)
  local time_total=$(echo "$response" | grep -o 'TIME:[0-9.]*' | cut -d':' -f2)
  
  echo "$http_code,$time_total" >> "$TEMP_FILE"
  echo "完成请求: $1, HTTP状态: $http_code, 耗时: $duration 秒"
}

# 显示开始信息
echo "开始对 $UPLOAD_URL 进行文件上传压测"
echo "总请求数: $REQUESTS, 并发数: $CONCURRENCY, 文件大小: $FILE_SIZE 字节"

# 记录开始时间
TEST_START=$(date +%s.%N)

# 并发执行请求
for i in $(seq 1 $REQUESTS); do
  # 控制并发数
  if [ $(jobs -r | wc -l) -ge $CONCURRENCY ]; then
    wait -n  # 等待一个任务完成
  fi
  
  run_upload_test $i &  # 后台执行
done

# 等待所有请求完成
wait

# 记录结束时间
TEST_END=$(date +%s.%N)
TOTAL_TIME=$(echo "$TEST_END - $TEST_START" | bc)

# 分析结果
SUCCESSFUL=$(grep -c "^200," "$TEMP_FILE")
TOTAL=$(wc -l < "$TEMP_FILE")
SUCCESS_RATE=$(echo "scale=2; $SUCCESSFUL / $TOTAL * 100" | bc)

# 计算响应时间统计
echo "分析响应时间..."
TIMES=$(cut -d',' -f2 "$TEMP_FILE")
TOTAL_RESP_TIME=$(echo "$TIMES" | paste -sd+ | bc)
AVG_RESP_TIME=$(echo "scale=3; $TOTAL_RESP_TIME / $TOTAL" | bc)

# 计算RPS和吞吐量
RPS=$(echo "scale=2; $REQUESTS / $TOTAL_TIME" | bc)
THROUGHPUT=$(echo "scale=2; $RPS * $FILE_SIZE / 1024" | bc)

# 输出结果
echo "=============== 文件上传压测结果 ==============="
echo "总请求数: $REQUESTS"
echo "并发数: $CONCURRENCY"
echo "文件大小: $FILE_SIZE 字节"
echo "总耗时: $TOTAL_TIME 秒"
echo "成功请求数: $SUCCESSFUL"
echo "成功率: $SUCCESS_RATE%"
echo "平均响应时间: $AVG_RESP_TIME 秒"
echo "每秒请求数(RPS): $RPS"
echo "吞吐量: $THROUGHPUT KB/s"
echo "================================================="

# 清理临时文件
rm "$TEMP_FILE" "$TEST_FILE"
```

#### 3. 使用ApacheBench (ab) 进行简单压测

如果系统中安装了Apache工具套件，也可以使用更专业的ab工具：

```bash
# 首先使用curl登录获取token
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' | 
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

# 使用ab进行文件列表查询测试
ab -n 1000 -c 50 -H "Authorization: Bearer $TOKEN" \
  "http://localhost:8080/api/v1/oss/files?page=1&page_size=10"

# 带JSON请求体的POST请求测试
echo '{"name":"test-config","region":"us-east-1","bucket":"test-bucket","storage_type":"s3"}' > /tmp/post_data.json
ab -n 500 -c 20 -T "application/json" -H "Authorization: Bearer $TOKEN" \
  -p /tmp/post_data.json "http://localhost:8080/api/v1/oss/configs"
```

#### 4. 多轮次递增压测脚本

以下脚本执行多轮次压测，并在每轮逐步增加并发数：

```bash
#!/bin/bash
# test/shell/incremental_load_test.sh

# 压测配置
URL="http://localhost:8080/api/v1/oss/files"
USERNAME="admin"
PASSWORD="admin123"
LOGIN_URL="http://localhost:8080/api/v1/auth/login"
ROUNDS=5
START_CONCURRENCY=10
INCREMENT=10
REQUESTS_PER_ROUND=100

# 获取认证令牌
echo "登录获取Token..."
TOKEN=$(curl -s -X POST $LOGIN_URL \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"password\":\"$PASSWORD\"}" | 
  grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
  echo "获取Token失败，请检查登录接口"
  exit 1
fi

# 结果文件
RESULTS_FILE="test_results_$(date +%Y%m%d_%H%M%S).csv"
echo "并发数,总请求数,成功请求数,成功率,总耗时(秒),平均响应时间(秒),RPS" > $RESULTS_FILE

# 循环执行压测轮次
for ((round=1; round<=ROUNDS; round++)); do
  CONCURRENCY=$((START_CONCURRENCY + (round-1)*INCREMENT))
  
  echo "===== 开始第 $round 轮压测 ====="
  echo "并发数: $CONCURRENCY, 请求数: $REQUESTS_PER_ROUND"
  
  # 临时结果文件
  TEMP_FILE=$(mktemp)
  
  # 记录开始时间
  TEST_START=$(date +%s.%N)
  
  # 执行并发请求
  for ((i=1; i<=REQUESTS_PER_ROUND; i++)); do
    # 控制并发数
    if [ $(jobs -r | wc -l) -ge $CONCURRENCY ]; then
      wait -n
    fi
    
    # 发送请求
    (curl -s -w "%{http_code},%{time_total}\n" -o /dev/null \
      -H "Authorization: Bearer $TOKEN" \
      "$URL" >> "$TEMP_FILE") &
  done
  
  # 等待所有请求完成
  wait
  
  # 记录结束时间
  TEST_END=$(date +%s.%N)
  TOTAL_TIME=$(echo "$TEST_END - $TEST_START" | bc)
  
  # 分析结果
  SUCCESSFUL=$(grep -c "^200," "$TEMP_FILE")
  TOTAL=$(wc -l < "$TEMP_FILE")
  SUCCESS_RATE=$(echo "scale=2; $SUCCESSFUL / $TOTAL * 100" | bc)
  
  # 计算响应时间统计
  TIMES=$(cut -d',' -f2 "$TEMP_FILE")
  TOTAL_RESP_TIME=$(echo "$TIMES" | paste -sd+ | bc)
  AVG_RESP_TIME=$(echo "scale=3; $TOTAL_RESP_TIME / $TOTAL" | bc)
  
  # 计算RPS
  RPS=$(echo "scale=2; $REQUESTS_PER_ROUND / $TOTAL_TIME" | bc)
  
  # 输出结果
  echo "=============== 第 $round 轮压测结果 ==============="
  echo "并发数: $CONCURRENCY"
  echo "总请求数: $REQUESTS_PER_ROUND"
  echo "成功请求数: $SUCCESSFUL"
  echo "成功率: $SUCCESS_RATE%"
  echo "总耗时: $TOTAL_TIME 秒"
  echo "平均响应时间: $AVG_RESP_TIME 秒"
  echo "每秒请求数(RPS): $RPS"
  echo "================================================="
  
  # 保存结果到CSV
  echo "$CONCURRENCY,$REQUESTS_PER_ROUND,$SUCCESSFUL,$SUCCESS_RATE%,$TOTAL_TIME,$AVG_RESP_TIME,$RPS" >> $RESULTS_FILE
  
  # 清理临时文件
  rm "$TEMP_FILE"
  
  # 每轮之间休息一下，避免过度压力
  if [ $round -lt $ROUNDS ]; then
    echo "休息5秒后开始下一轮..."
    sleep 5
  fi
done

echo "压测完成，结果已保存到: $RESULTS_FILE"

# 可以使用以下命令生成简单的结果图表（需要安装gnuplot）
if command -v gnuplot >/dev/null 2>&1; then
  echo "使用gnuplot生成结果图表..."
  
  # 生成gnuplot脚本
  cat > plot.gp << EOL
set terminal png size 800,600
set output "test_results_$(date +%Y%m%d_%H%M%S).png"
set title "API压测结果"
set xlabel "并发数"
set ylabel "每秒请求数(RPS)"
set y2label "平均响应时间(秒)"
set y2tics
set grid
plot "$RESULTS_FILE" using 1:7 with linespoints title "RPS", \
     "$RESULTS_FILE" using 1:6 with linespoints axes x1y2 title "响应时间"
EOL

  # 执行gnuplot生成图表
  gnuplot plot.gp
  rm plot.gp
  echo "图表生成完成"
fi
```

#### 使用curl压测的注意事项

1. **资源消耗**: 这些脚本在本地执行时会消耗较多系统资源，特别是在高并发测试时
2. **网络因素**: 本地网络延迟可能影响测试结果
3. **适用场景**: 这些方法适合开发和测试环境下的简单验证，不适合正式的性能测试
4. **结果精度**: 结果不如专业压测工具精确，但足以提供基本的性能评估
5. **Token有效期**: 注意登录获取的token可能有效期限制，长时间测试可能需要定期刷新

## 压测指标

- **平均响应时间 (Average Response Time)**: 所有请求的平均响应时间
- **请求/秒 (RPS)**: 每秒处理的请求数
- **错误率 (Error Rate)**: 请求失败的百分比
- **95/99百分位响应时间**: 95%/99%的请求能在这个时间内得到响应
- **资源使用率**: CPU、内存、网络I/O等资源使用情况

## 压测场景

### 1. 认证接口压测

```javascript
// test/k6/auth_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter } from 'k6/metrics';

// 定义指标
const loginErrors = new Counter('login_errors');

// 定义压测配置
export const options = {
  stages: [
    { duration: '1m', target: 50 }, // 逐步增加到50个并发用户
    { duration: '3m', target: 50 }, // 保持50个并发用户3分钟
    { duration: '1m', target: 0 },  // 逐步减少到0个并发用户
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95%的请求响应时间<500ms
    'login_errors': ['count<10'],     // 登录错误次数<10
  },
};

// 测试逻辑
export default function() {
  const url = 'http://localhost:8080/api/v1/auth/login';
  const payload = JSON.stringify({
    username: `user_${__VU}`, // 使用虚拟用户ID作为用户名，确保不同用户登录
    password: 'password123',
  });
  
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
  };
  
  // 发送登录请求
  const res = http.post(url, payload, params);
  
  // 检查响应
  const success = check(res, {
    'login successful': (r) => r.status === 200,
    'token received': (r) => JSON.parse(r.body).data && JSON.parse(r.body).data.token,
  });
  
  if (!success) {
    loginErrors.add(1);
    console.log(`Login failed: ${res.status} ${res.body}`);
  }
  
  sleep(1);
}
```

### 2. 文件上传接口压测

```javascript
// test/k6/file_upload_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// 定义指标
const uploadErrors = new Counter('upload_errors');
const uploadTime = new Trend('upload_time');

// 预先登录获取token
const users = new SharedArray('users', function() {
  // 这个函数在初始化阶段只执行一次
  const res = http.post('http://localhost:8080/api/v1/auth/login', 
    JSON.stringify({ username: 'admin', password: 'admin123' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  return [{ token: JSON.parse(res.body).data.token }];
});

// 定义压测配置
export const options = {
  stages: [
    { duration: '1m', target: 10 },  // 逐步增加到10个并发用户
    { duration: '3m', target: 10 },  // 保持10个并发用户3分钟
    { duration: '1m', target: 0 },   // 逐步减少到0个并发用户
  ],
  thresholds: {
    http_req_duration: ['p(95)<3000'], // 95%的请求响应时间<3s
    'upload_errors': ['count<5'],      // 上传错误次数<5
  },
};

// 测试逻辑
export default function() {
  const token = users[0].token;
  
  // 创建二进制数据作为文件内容（约100KB）
  const fileContent = new Array(100 * 1024).fill('A').join('');
  
  // 创建FormData格式的请求
  const data = {
    file: http.file(fileContent, `test_${__VU}_${Date.now()}.txt`, 'text/plain'),
    config_id: '1',
  };
  
  // 发送文件上传请求
  const startTime = new Date();
  const res = http.post('http://localhost:8080/api/v1/oss/files', data, {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  const endTime = new Date();
  
  // 记录上传时间
  uploadTime.add(endTime - startTime);
  
  // 检查响应
  const success = check(res, {
    'upload successful': (r) => r.status === 200,
    'file info received': (r) => JSON.parse(r.body).data && JSON.parse(r.body).data.download_url,
  });
  
  if (!success) {
    uploadErrors.add(1);
    console.log(`Upload failed: ${res.status} ${res.body}`);
  }
  
  sleep(3);
}
```

### 3. 文件列表查询接口压测

```javascript
// test/k6/file_list_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// 定义指标
const queryErrors = new Counter('query_errors');
const queryTime = new Trend('query_time');

// 预先登录获取token
const users = new SharedArray('users', function() {
  const res = http.post('http://localhost:8080/api/v1/auth/login', 
    JSON.stringify({ username: 'admin', password: 'admin123' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  return [{ token: JSON.parse(res.body).data.token }];
});

// 定义压测配置
export const options = {
  stages: [
    { duration: '1m', target: 30 },  // 逐步增加到30个并发用户
    { duration: '3m', target: 30 },  // 保持30个并发用户3分钟
    { duration: '1m', target: 0 },   // 逐步减少到0个并发用户
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'], // 95%的请求响应时间<500ms
    'query_errors': ['count<5'],      // 查询错误次数<5
  },
};

// 测试逻辑
export default function() {
  const token = users[0].token;
  
  // 随机选择页码和每页数量
  const page = Math.floor(Math.random() * 5) + 1;
  const pageSize = 10;
  
  // 发送查询请求
  const startTime = new Date();
  const res = http.get(`http://localhost:8080/api/v1/oss/files?page=${page}&page_size=${pageSize}`, {
    headers: {
      'Authorization': `Bearer ${token}`,
    },
  });
  const endTime = new Date();
  
  // 记录查询时间
  queryTime.add(endTime - startTime);
  
  // 检查响应
  const success = check(res, {
    'query successful': (r) => r.status === 200,
    'file list received': (r) => {
      const data = JSON.parse(r.body).data;
      return data && Array.isArray(data.items);
    },
  });
  
  if (!success) {
    queryErrors.add(1);
    console.log(`Query failed: ${res.status} ${res.body}`);
  }
  
  sleep(1);
}
```

### 4. 下载链接生成接口压测

```javascript
// test/k6/download_url_test.js
import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Trend } from 'k6/metrics';
import { SharedArray } from 'k6/data';

// 定义指标
const urlGenErrors = new Counter('url_gen_errors');
const urlGenTime = new Trend('url_gen_time');

// 预先登录获取token和文件列表
const testData = new SharedArray('testData', function() {
  // 登录获取token
  const loginRes = http.post('http://localhost:8080/api/v1/auth/login', 
    JSON.stringify({ username: 'admin', password: 'admin123' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  const token = JSON.parse(loginRes.body).data.token;
  
  // 获取文件列表
  const listRes = http.get('http://localhost:8080/api/v1/oss/files?page=1&page_size=20', {
    headers: { 'Authorization': `Bearer ${token}` }
  });
  
  const files = JSON.parse(listRes.body).data.items;
  return {
    token: token,
    fileIds: files.map(file => file.id)
  };
});

// 定义压测配置
export const options = {
  stages: [
    { duration: '1m', target: 50 },  // 逐步增加到50个并发用户
    { duration: '3m', target: 50 },  // 保持50个并发用户3分钟
    { duration: '1m', target: 0 },   // 逐步减少到0个并发用户
  ],
  thresholds: {
    http_req_duration: ['p(95)<300'], // 95%的请求响应时间<300ms
    'url_gen_errors': ['count<5'],    // 生成URL错误次数<5
  },
};

// 测试逻辑
export default function() {
  if (!testData.fileIds || testData.fileIds.length === 0) {
    console.log('No files found in database, skipping test');
    sleep(1);
    return;
  }
  
  // 随机选择一个文件ID
  const fileId = testData.fileIds[Math.floor(Math.random() * testData.fileIds.length)];
  
  // 发送生成下载链接请求
  const startTime = new Date();
  const res = http.get(`http://localhost:8080/api/v1/oss/files/${fileId}/download`, {
    headers: {
      'Authorization': `Bearer ${testData.token}`,
    },
  });
  const endTime = new Date();
  
  // 记录生成时间
  urlGenTime.add(endTime - startTime);
  
  // 检查响应
  const success = check(res, {
    'url generation successful': (r) => r.status === 200,
    'download url received': (r) => {
      const data = JSON.parse(r.body).data;
      return data && data.download_url;
    },
  });
  
  if (!success) {
    urlGenErrors.add(1);
    console.log(`URL generation failed: ${res.status} ${res.body}`);
  }
  
  sleep(1);
}
```

### 5. 混合压测场景

```javascript
// test/k6/mixed_test.js
import { group, sleep } from 'k6';
import { SharedArray } from 'k6/data';
import http from 'k6/http';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

// 定义指标
const errors = new Counter('errors');

// 预先登录获取token
const users = new SharedArray('users', function() {
  const res = http.post('http://localhost:8080/api/v1/auth/login', 
    JSON.stringify({ username: 'admin', password: 'admin123' }),
    { headers: { 'Content-Type': 'application/json' } }
  );
  return [{ token: JSON.parse(res.body).data.token }];
});

// 定义压测配置
export const options = {
  scenarios: {
    // 查询场景 - 高并发，占70%的流量
    queries: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 35 },
        { duration: '3m', target: 35 },
        { duration: '1m', target: 0 },
      ],
      gracefulRampDown: '30s',
      exec: 'queryFiles',
    },
    // 上传场景 - 中等并发，占20%的流量
    uploads: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 10 },
        { duration: '3m', target: 10 },
        { duration: '1m', target: 0 },
      ],
      gracefulRampDown: '30s',
      exec: 'uploadFile',
    },
    // 下载链接生成场景 - 低并发，占10%的流量
    downloads: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 5 },
        { duration: '3m', target: 5 },
        { duration: '1m', target: 0 },
      ],
      gracefulRampDown: '30s',
      exec: 'generateDownloadUrl',
    },
  },
  thresholds: {
    http_req_duration: ['p(95)<1000'], // 95%的请求响应时间<1s
    'errors': ['count<50'],           // 总错误数<50
  },
};

// 文件查询场景
export function queryFiles() {
  const token = users[0].token;
  
  group('List Files', function() {
    const page = Math.floor(Math.random() * 5) + 1;
    const res = http.get(`http://localhost:8080/api/v1/oss/files?page=${page}&page_size=10`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    
    const success = check(res, {
      'list files successful': (r) => r.status === 200,
    });
    
    if (!success) errors.add(1);
  });
  
  sleep(Math.random() * 2);
}

// 文件上传场景
export function uploadFile() {
  const token = users[0].token;
  
  group('Upload File', function() {
    // 创建一个小文件（10KB左右）
    const fileContent = new Array(10 * 1024).fill('A').join('');
    const data = {
      file: http.file(fileContent, `test_${__VU}_${Date.now()}.txt`, 'text/plain'),
      config_id: '1',
    };
    
    const res = http.post('http://localhost:8080/api/v1/oss/files', data, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    
    const success = check(res, {
      'upload file successful': (r) => r.status === 200,
    });
    
    if (!success) errors.add(1);
  });
  
  sleep(Math.random() * 3 + 2);
}

// 生成下载链接场景
export function generateDownloadUrl() {
  const token = users[0].token;
  
  group('Generate Download URL', function() {
    // 首先获取文件列表
    const listRes = http.get('http://localhost:8080/api/v1/oss/files?page=1&page_size=5', {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    
    const success = check(listRes, {
      'list files successful': (r) => r.status === 200,
    });
    
    if (!success) {
      errors.add(1);
      return;
    }
    
    const files = JSON.parse(listRes.body).data.items;
    if (!files || files.length === 0) {
      console.log('No files found, skipping download URL generation');
      return;
    }
    
    // 随机选择一个文件
    const fileId = files[Math.floor(Math.random() * files.length)].id;
    
    // 生成下载链接
    const dlRes = http.get(`http://localhost:8080/api/v1/oss/files/${fileId}/download`, {
      headers: { 'Authorization': `Bearer ${token}` },
    });
    
    const dlSuccess = check(dlRes, {
      'generate download url successful': (r) => r.status === 200,
    });
    
    if (!dlSuccess) errors.add(1);
  });
  
  sleep(Math.random() * 2 + 1);
}
```

## 压测执行

### 基本命令

```bash
# 执行单一场景压测
k6 run test/k6/auth_test.js

# 执行混合场景压测
k6 run test/k6/mixed_test.js

# 指定虚拟用户数和持续时间
k6 run --vus 50 --duration 30s test/k6/file_list_test.js

# 保存结果到JSON文件
k6 run --out json=result.json test/k6/file_upload_test.js
```

### 输出结果到InfluxDB和Grafana（可选）

如果需要更好的可视化，可以将结果输出到InfluxDB，然后使用Grafana展示：

```bash
# 启动InfluxDB和Grafana（使用Docker）
docker-compose up -d influxdb grafana

# 执行压测并输出到InfluxDB
k6 run --out influxdb=http://localhost:8086/k6 test/k6/mixed_test.js
```

docker-compose.yml 配置示例：

```yaml
version: '3'
services:
  influxdb:
    image: influxdb:1.8
    ports:
      - "8086:8086"
    environment:
      - INFLUXDB_DB=k6
      - INFLUXDB_ADMIN_USER=admin
      - INFLUXDB_ADMIN_PASSWORD=admin
    volumes:
      - influxdb-data:/var/lib/influxdb

  grafana:
    image: grafana/grafana:latest
    ports:
      - "3000:3000"
    environment:
      - GF_AUTH_ANONYMOUS_ENABLED=true
      - GF_AUTH_ANONYMOUS_ORG_ROLE=Admin
    volumes:
      - grafana-data:/var/lib/grafana
    depends_on:
      - influxdb

volumes:
  influxdb-data:
  grafana-data:
```

## 压测结果分析

### 预期目标

| 接口 | 并发用户数 | 平均响应时间 | 95%响应时间 | RPS | 错误率 |
|------|------------|--------------|-------------|-----|--------|
| 登录 | 50         | <200ms       | <500ms      | >100| <1%    |
| 文件上传 | 10     | <1s          | <3s         | >5  | <2%    |
| 文件列表 | 30     | <200ms       | <500ms      | >50 | <1%    |
| 下载链接 | 50     | <100ms       | <300ms      | >100| <1%    |

### 性能瓶颈判断

在压测过程中，可以通过以下指标判断系统的瓶颈：

1. **CPU 使用率**: 如果 CPU 使用率接近 100%，则可能是应用程序逻辑或并发处理能力存在瓶颈
2. **内存使用率**: 如果内存使用持续增长，可能存在内存泄漏
3. **I/O 等待**: 如果 I/O 等待时间较长，可能是数据库或文件系统存在瓶颈
4. **网络流量**: 如果网络带宽接近饱和，可能需要优化网络或考虑CDN加速

### 优化建议

根据压测结果，可能的优化方向包括：

1. **连接池优化**: 调整数据库连接池大小，确保高并发场景下有足够的连接可用
2. **缓存策略**: 对频繁访问的数据（如文件列表、文件元数据）进行缓存
3. **异步处理**: 将文件上传、MD5计算等耗时操作改为异步处理
4. **负载均衡**: 在多实例部署时，使用负载均衡分散请求压力
5. **资源限制**: 对单个用户的请求频率和并发数进行限制，防止恶意攻击

## 压测注意事项

1. **测试数据准备**: 确保数据库中有足够的测试数据，特别是文件记录
2. **隔离环境**: 在非生产环境进行压测，避免影响正常业务
3. **监控系统资源**: 在压测过程中监控系统资源使用情况
4. **逐步增加负载**: 从小负载开始，逐步增加，找到系统的极限
5. **清理测试数据**: 测试完成后清理测试产生的数据

## 参考资料

- [k6 官方文档](https://k6.io/docs/)
- [Grafana k6 Dashboard](https://grafana.com/grafana/dashboards/2587)
- [性能测试最佳实践](https://k6.io/docs/testing-guides/api-load-testing)

## OSS管理系统分片上传测试指南

本文档提供了OSS管理系统分片上传功能的详细测试方法和示例，帮助开发者和测试人员了解如何测试大文件分片上传功能。

### 1. 分片上传流程概述

OSS管理系统的分片上传功能允许用户将大文件分成多个片段分别上传，从而提高上传成功率和效率。完整的分片上传流程包括以下步骤：

1. **初始化分片上传**：获取上传ID和对象键
2. **上传分片**：使用初始化返回的信息，分别上传各个分片
3. **完成分片上传**：所有分片上传完成后，通知服务器合并分片
4. **（可选）取消分片上传**：在任何时候可以取消分片上传

### 2. API接口说明

分片上传相关API接口如下：

| 接口 | 请求方法 | 路径 | 功能描述 |
|-----|---------|-----|---------|
| 初始化分片上传 | POST | /api/v1/oss/multipart/init | 获取上传ID和对象键 |
| 完成分片上传 | POST | /api/v1/oss/multipart/complete | 所有分片上传完成后调用 |
| 取消分片上传 | DELETE | /api/v1/oss/multipart/abort | 取消分片上传 |

### 3. 测试准备

在测试分片上传功能前，需要做以下准备：

1. 确保已经获取有效的认证令牌（通过登录接口获取）
2. 准备一个测试用的大文件（建议大于10MB）
3. 准备分片上传所需的工具（如curl、Postman或自定义脚本）
4. 确认已创建有效的OSS配置（使用OSS配置管理接口）

### 4. 初始化分片上传测试

#### 使用curl测试初始化分片上传

```bash
# 设置认证令牌
TOKEN="您的认证令牌"

# 设置配置ID（替换为您的实际配置ID）
CONFIG_ID="1"

# 初始化分片上传
curl -X POST "http://localhost:8080/api/v1/oss/multipart/init" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config_id": "'"$CONFIG_ID"'",
    "file_name": "large_test_file.zip",
    "file_size": 10485760
  }'
```

#### 预期响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "upload_id": "abc123def456ghi789",
    "object_key": "uploads/202401010001_large_test_file.zip"
  }
}
```

#### 检查点

- 响应状态码应为200
- 返回的数据中包含`upload_id`和`object_key`，这两个值需要保存用于后续操作

### 5. 获取分片上传URL测试

对于支持直接上传到对象存储的场景，需要为每个分片获取一个预签名URL。

#### 使用curl测试获取分片上传URL

```bash
# 使用上一步获取的upload_id和对象键
UPLOAD_ID="上一步获取的upload_id"
OBJECT_KEY="上一步获取的object_key"

# 获取分片上传URL
curl -X GET "http://localhost:8080/api/v1/oss/multipart/${CONFIG_ID}/urls?upload_id=${UPLOAD_ID}&part_numbers=1,2,3" \
  -H "Authorization: Bearer $TOKEN"
```

#### 预期响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "urls": [
      {
        "part_number": 1,
        "url": "https://bucket.s3.region.amazonaws.com/object_key?partNumber=1&uploadId=upload_id&X-Amz-Algorithm=..."
      },
      {
        "part_number": 2,
        "url": "https://bucket.s3.region.amazonaws.com/object_key?partNumber=2&uploadId=upload_id&X-Amz-Algorithm=..."
      },
      {
        "part_number": 3,
        "url": "https://bucket.s3.region.amazonaws.com/object_key?partNumber=3&uploadId=upload_id&X-Amz-Algorithm=..."
      }
    ]
  }
}
```

#### 检查点

- 响应状态码应为200
- 返回的数据中包含与请求的part_numbers数量相同的URL条目

### 6. 上传分片测试

使用获取的预签名URL上传各个分片。这一步通常直接与云存储服务交互，不经过我们的服务器。

#### 使用curl测试上传分片

```bash
# 分割文件（Linux环境）
split -b 5M large_test_file.zip part_

# 上传第一个分片并保存返回的ETag（示例使用AWS S3格式）
PART1_ETAG=$(curl -X PUT -T part_aa "第一个分片的预签名URL" -H "Content-Type: application/octet-stream" -v 2>&1 | grep -i "ETag" | awk '{print $3}' | tr -d '"')

# 上传第二个分片并保存返回的ETag
PART2_ETAG=$(curl -X PUT -T part_ab "第二个分片的预签名URL" -H "Content-Type: application/octet-stream" -v 2>&1 | grep -i "ETag" | awk '{print $3}' | tr -d '"')

# 上传第三个分片并保存返回的ETag
PART3_ETAG=$(curl -X PUT -T part_ac "第三个分片的预签名URL" -H "Content-Type: application/octet-stream" -v 2>&1 | grep -i "ETag" | awk '{print $3}' | tr -d '"')

# 打印所有ETag
echo "Part 1 ETag: $PART1_ETAG"
echo "Part 2 ETag: $PART2_ETAG"
echo "Part 3 ETag: $PART3_ETAG"
```

#### 检查点

- 每个分片上传应返回一个ETag
- 所有分片上传请求应返回200或204状态码

### 7. 完成分片上传测试

所有分片都上传完成后，调用完成分片上传接口。

#### 使用curl测试完成分片上传

```bash
# 使用之前保存的upload_id、object_key和ETags
curl -X POST "http://localhost:8080/api/v1/oss/multipart/complete" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config_id": "'"$CONFIG_ID"'",
    "upload_id": "'"$UPLOAD_ID"'",
    "object_key": "'"$OBJECT_KEY"'",
    "parts": ["'"$PART1_ETAG"'", "'"$PART2_ETAG"'", "'"$PART3_ETAG"'"]
  }'
```

#### 预期响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {
    "id": 123,
    "file_name": "large_test_file.zip",
    "file_size": 15728640,
    "storage_type": "AWS_S3",
    "object_key": "uploads/202401010001_large_test_file.zip",
    "download_url": "https://bucket.s3.region.amazonaws.com/object_key?X-Amz-Algorithm=...",
    "created_at": "2024-01-01T12:00:00Z"
  }
}
```

#### 检查点

- 响应状态码应为200
- 返回的数据中包含已合并文件的信息，包括ID、文件名、大小、下载URL等

### 8. 取消分片上传测试

测试取消分片上传功能。

#### 使用curl测试取消分片上传

```bash
# 初始化一个新的分片上传用于测试取消功能
NEW_UPLOAD_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/oss/multipart/init" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "config_id": "'"$CONFIG_ID"'",
    "file_name": "to_be_aborted.zip",
    "file_size": 10485760
  }')

# 提取新的upload_id
NEW_UPLOAD_ID=$(echo $NEW_UPLOAD_RESPONSE | grep -o '"upload_id":"[^"]*"' | cut -d'"' -f4)

# 取消分片上传
curl -X DELETE "http://localhost:8080/api/v1/oss/multipart/abort?config_id=${CONFIG_ID}&upload_id=${NEW_UPLOAD_ID}" \
  -H "Authorization: Bearer $TOKEN"
```

#### 预期响应

```json
{
  "code": 0,
  "message": "成功",
  "data": {}
}
```

#### 检查点

- 响应状态码应为200
- 取消成功后，使用相同的upload_id尝试完成上传应该失败

### 9. 自动化测试脚本

在`test/shell/`目录下创建以下脚本，可以用来测试分片上传的完整流程：

```bash
#!/bin/bash
# test/shell/multipart_upload_test.sh

# 配置信息
BASE_URL="http://localhost:8080/api/v1"
USERNAME="admin"
PASSWORD="admin123"
CONFIG_ID="1"  # 使用现有的配置ID，或者通过API创建一个新的配置
TEST_FILE="/tmp/large_test_file.tmp"
FILE_SIZE=$((10 * 1024 * 1024))  # 10MB
CHUNK_SIZE=$((5 * 1024 * 1024))  # 5MB

# 创建测试文件
echo "创建测试文件 (${FILE_SIZE} 字节)..."
dd if=/dev/urandom of=$TEST_FILE bs=1M count=10 2>/dev/null

# 登录获取token
echo "登录获取Token..."
LOGIN_RESULT=$(curl -s -X POST "${BASE_URL}/auth/login" \
  -H "Content-Type: application/json" \
  -d "{
    \"username\": \"${USERNAME}\",
    \"password\": \"${PASSWORD}\"
  }")

TOKEN=$(echo $LOGIN_RESULT | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
if [ -z "$TOKEN" ]; then
  echo "登录失败，无法获取token"
  exit 1
fi
echo "获取Token成功"

# 初始化分片上传
echo "初始化分片上传..."
INIT_RESULT=$(curl -s -X POST "${BASE_URL}/oss/multipart/init" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"config_id\": \"${CONFIG_ID}\",
    \"file_name\": \"large_test_file.bin\",
    \"file_size\": ${FILE_SIZE}
  }")

echo "初始化结果: $INIT_RESULT"

# 提取上传ID和对象键
UPLOAD_ID=$(echo $INIT_RESULT | grep -o '"upload_id":"[^"]*"' | cut -d'"' -f4)
OBJECT_KEY=$(echo $INIT_RESULT | grep -o '"object_key":"[^"]*"' | cut -d'"' -f4)

if [ -z "$UPLOAD_ID" ] || [ -z "$OBJECT_KEY" ]; then
  echo "初始化失败，无法获取上传ID或对象键"
  exit 1
fi

echo "上传ID: $UPLOAD_ID"
echo "对象键: $OBJECT_KEY"

# 分割文件并上传各个分片
echo "分割文件..."
split -b $CHUNK_SIZE $TEST_FILE ${TEST_FILE}_part_

# 获取分片列表
PARTS=(${TEST_FILE}_part_*)
PART_COUNT=${#PARTS[@]}
echo "文件已分割为 $PART_COUNT 个分片"

# 获取上传URL
echo "获取分片上传URL..."
PART_NUMBERS=$(seq -s, 1 $PART_COUNT)
URL_RESULT=$(curl -s -X GET "${BASE_URL}/oss/multipart/${CONFIG_ID}/urls?upload_id=${UPLOAD_ID}&part_numbers=${PART_NUMBERS}" \
  -H "Authorization: Bearer $TOKEN")

echo "URL结果: $URL_RESULT"

# 上传分片并收集ETags
ETAGS=()

echo "上传分片..."
for i in $(seq 1 $PART_COUNT); do
  PART_FILE="${TEST_FILE}_part_$(printf '%s' $(echo "a" | tr "a" $(printf "\\$(printf '%o' $(($i+96)))")))"
  
  # 获取当前分片的URL (这里需要根据实际返回格式调整提取方式)
  PART_URL=$(echo $URL_RESULT | jq -r ".data.urls[$((i-1))].url")
  
  if [ -z "$PART_URL" ] || [ "$PART_URL" = "null" ]; then
    echo "无法获取分片 $i 的URL，跳过直接上传"
    # 如果无法获取预签名URL，可以使用系统可能提供的其他上传方式
    # 这里仅作示例，实际实现需要根据系统设计调整
    continue
  fi
  
  echo "上传分片 $i: $PART_FILE"
  # 使用预签名URL上传分片 (AWS S3格式示例)
  UPLOAD_RESPONSE=$(curl -s -X PUT -T $PART_FILE "$PART_URL" -H "Content-Type: application/octet-stream" -v 2>&1)
  
  # 提取ETag (根据实际返回格式调整)
  ETAG=$(echo "$UPLOAD_RESPONSE" | grep -i "ETag" | awk '{print $3}' | tr -d '"')
  
  if [ -z "$ETAG" ]; then
    echo "无法获取分片 $i 的ETag，使用模拟值"
    ETAG="etag_part_$i"  # 模拟值，实际测试中需要真实值
  fi
  
  ETAGS+=("$ETAG")
  echo "分片 $i ETag: $ETAG"
done

# 构建分片信息JSON
PARTS_JSON="["
for i in $(seq 0 $((${#ETAGS[@]}-1))); do
  if [ $i -gt 0 ]; then
    PARTS_JSON+=","
  fi
  PARTS_JSON+="\"${ETAGS[$i]}\""
done
PARTS_JSON+="]"

# 完成分片上传
echo "完成分片上传..."
COMPLETE_RESULT=$(curl -s -X POST "${BASE_URL}/oss/multipart/complete" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"config_id\": \"${CONFIG_ID}\",
    \"upload_id\": \"${UPLOAD_ID}\",
    \"object_key\": \"${OBJECT_KEY}\",
    \"parts\": ${PARTS_JSON}
  }")

echo "完成结果: $COMPLETE_RESULT"

# 测试取消分片上传
echo "测试取消分片上传..."
NEW_INIT_RESULT=$(curl -s -X POST "${BASE_URL}/oss/multipart/init" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"config_id\": \"${CONFIG_ID}\",
    \"file_name\": \"to_be_aborted.bin\",
    \"file_size\": ${FILE_SIZE}
  }")

NEW_UPLOAD_ID=$(echo $NEW_INIT_RESULT | grep -o '"upload_id":"[^"]*"' | cut -d'"' -f4)

if [ -z "$NEW_UPLOAD_ID" ]; then
  echo "初始化新上传失败，无法测试取消功能"
else
  echo "取消分片上传 (Upload ID: $NEW_UPLOAD_ID)..."
  ABORT_RESULT=$(curl -s -X DELETE "${BASE_URL}/oss/multipart/abort?config_id=${CONFIG_ID}&upload_id=${NEW_UPLOAD_ID}" \
    -H "Authorization: Bearer $TOKEN")
  
  echo "取消结果: $ABORT_RESULT"
fi

# 清理临时文件
echo "清理临时文件..."
rm -f $TEST_FILE ${TEST_FILE}_part_*

echo "分片上传测试完成"
```

使用方法：
```bash
chmod +x test/shell/multipart_upload_test.sh
./test/shell/multipart_upload_test.sh
```

### 10. 前端测试

以下是使用JavaScript在浏览器中测试分片上传的示例代码：

```javascript
// 分片上传测试函数
async function testMultipartUpload() {
  const token = 'your_token_here'; // 从登录响应中获取
  const configId = '1';
  const file = document.getElementById('fileInput').files[0];
  const chunkSize = 5 * 1024 * 1024; // 5MB
  const chunks = Math.ceil(file.size / chunkSize);
  
  // 初始化分片上传
  const initResponse = await fetch('http://localhost:8080/api/v1/oss/multipart/init', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      config_id: configId,
      file_name: file.name,
      file_size: file.size
    })
  }).then(res => res.json());
  
  console.log('初始化分片上传响应:', initResponse);
  
  const uploadId = initResponse.data.upload_id;
  const objectKey = initResponse.data.object_key;
  
  if (!uploadId || !objectKey) {
    console.error('初始化失败，无法获取上传ID或对象键');
    return;
  }
  
  // 获取分片上传URL
  const partNumbers = Array.from({length: chunks}, (_, i) => i + 1).join(',');
  const urlResponse = await fetch(`http://localhost:8080/api/v1/oss/multipart/${configId}/urls?upload_id=${uploadId}&part_numbers=${partNumbers}`, {
    method: 'GET',
    headers: {
      'Authorization': `Bearer ${token}`
    }
  }).then(res => res.json());
  
  console.log('获取分片上传URL响应:', urlResponse);
  
  // 上传分片
  const etags = [];
  for (let i = 0; i < chunks; i++) {
    const start = i * chunkSize;
    const end = Math.min(file.size, start + chunkSize);
    const chunk = file.slice(start, end);
    
    const partUrl = urlResponse.data.urls[i].url;
    
    // 上传分片到预签名URL
    const uploadResponse = await fetch(partUrl, {
      method: 'PUT',
      body: chunk,
      headers: {
        'Content-Type': 'application/octet-stream'
      }
    });
    
    const etag = uploadResponse.headers.get('ETag').replace(/"/g, '');
    etags.push(etag);
    console.log(`分片 ${i+1} 上传完成, ETag: ${etag}`);
  }
  
  // 完成分片上传
  const completeResponse = await fetch('http://localhost:8080/api/v1/oss/multipart/complete', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      config_id: configId,
      upload_id: uploadId,
      object_key: objectKey,
      parts: etags
    })
  }).then(res => res.json());
  
  console.log('完成分片上传响应:', completeResponse);
  
  return completeResponse;
}

// HTML代码示例
// <input type="file" id="fileInput" />
// <button onclick="testMultipartUpload()">测试分片上传</button>
```

### 11. 注意事项

测试分片上传功能时，请注意以下几点：

1. **文件大小**：建议使用较大的文件（如10MB以上）测试分片上传，小文件可能不会触发分片机制
2. **网络环境**：在不同网络环境下测试，包括稳定网络和不稳定网络
3. **存储服务兼容性**：针对不同的存储服务（AWS S3、阿里云OSS、Cloudflare R2等）进行测试，它们的分片上传实现可能略有不同
4. **恢复测试**：测试网络中断后恢复上传的场景
5. **并发测试**：测试同时进行多个分片上传任务
6. **大文件测试**：测试超大文件（如1GB以上）的分片上传性能
7. **文件类型测试**：测试不同类型的文件（如图片、视频、文档等）

### 12. 进阶测试

完成基本测试后，可以进行以下进阶测试：

1. **断点续传**：模拟网络中断，然后使用相同的uploadId继续上传剩余分片
2. **超时处理**：测试上传超时的情况，验证系统是否能正确处理
3. **并发上传**：测试多个用户同时进行分片上传
4. **分片大小优化**：测试不同分片大小对上传速度的影响
5. **加密文件**：测试上传加密文件的场景
6. **性能测试**：在高负载条件下测试分片上传功能的性能和稳定性 