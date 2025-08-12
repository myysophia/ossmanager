#!/bin/bash

# OSS Manager 安全性检查脚本
# 检查配置文件中是否包含敏感信息

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔍 OSS Manager 安全性检查${NC}"
echo "========================================"

# 检查结果统计
ISSUES_FOUND=0

# 定义敏感信息模式（实际的硬编码值，不包括环境变量占位符）
SENSITIVE_PATTERNS=(
    # 真实的阿里云访问密钥
    'LTAI[A-Za-z0-9]{12,}'
    # 真实的AWS访问密钥
    'AKIA[A-Z0-9]{16}'
    # 真实的IP地址（不是localhost或0.0.0.0）
    '[1-9][0-9]+\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}'
    # Supabase连接字符串
    'postgres\.[a-zA-Z0-9]+'
)

# 检查文件函数
check_file() {
    local file=$1
    echo -e "\n${YELLOW}检查文件: $file${NC}"
    
    if [[ ! -f "$file" ]]; then
        echo -e "${YELLOW}⚠️  文件不存在，跳过检查${NC}"
        return
    fi
    
    local file_issues=0
    
    # 检查每个模式
    for pattern in "${SENSITIVE_PATTERNS[@]}"; do
        if grep -iE "$pattern" "$file" >/dev/null 2>&1; then
            echo -e "${RED}❌ 发现可疑内容:${NC}"
            grep -iE --color=always "$pattern" "$file" || true
            ((file_issues++))
            ((ISSUES_FOUND++))
        fi
    done
    
    # 特殊检查：硬编码的凭据
    if grep -E "(LTAI|AK[A-Z0-9]{10,})" "$file" >/dev/null 2>&1; then
        echo -e "${RED}❌ 发现阿里云访问密钥:${NC}"
        grep -E --color=always "(LTAI|AK[A-Z0-9]{10,})" "$file" || true
        ((file_issues++))
        ((ISSUES_FOUND++))
    fi
    
    if grep -E "[A-Za-z0-9+/]{40,}=*" "$file" | grep -v "\${" >/dev/null 2>&1; then
        echo -e "${RED}❌ 发现可能的 Base64 编码凭据:${NC}"
        grep -E --color=always "[A-Za-z0-9+/]{40,}=*" "$file" | grep -v "\${" || true
        ((file_issues++))
        ((ISSUES_FOUND++))
    fi
    
    if [[ $file_issues -eq 0 ]]; then
        echo -e "${GREEN}✅ 未发现敏感信息${NC}"
    fi
}

# 检查配置文件
echo -e "\n${BLUE}检查配置文件...${NC}"
for config_file in configs/*.yaml configs/*.yml; do
    [[ -f "$config_file" ]] && check_file "$config_file"
done

# 检查环境变量文件
echo -e "\n${BLUE}检查环境变量文件...${NC}"
for env_file in .env .env.local .env.production .env.example; do
    [[ -f "$env_file" ]] && check_file "$env_file"
done

# 检查其他可能包含敏感信息的文件
echo -e "\n${BLUE}检查其他配置文件...${NC}"
for other_file in docker-compose.yml docker-compose.*.yml; do
    [[ -f "$other_file" ]] && check_file "$other_file"
done

echo -e "\n========================================"
if [[ $ISSUES_FOUND -eq 0 ]]; then
    echo -e "${GREEN}🎉 安全检查通过！未发现敏感信息。${NC}"
    exit 0
else
    echo -e "${RED}⚠️  发现 $ISSUES_FOUND 个潜在的安全问题！${NC}"
    echo -e "${YELLOW}请检查并移除上述敏感信息，使用环境变量替代。${NC}"
    echo ""
    echo -e "${BLUE}建议操作：${NC}"
    echo "1. 将敏感信息替换为环境变量，如 \${VAR_NAME}"
    echo "2. 在 .env 文件中设置实际值（.env 已在 .gitignore 中）"
    echo "3. 更新 .env.example 文件提供配置示例"
    echo "4. 确认 .gitignore 包含所有敏感文件"
    exit 1
fi
