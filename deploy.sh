#!/bin/bash

# OSS Manager Docker 部署脚本
# 使用方法: ./deploy.sh [action]
# action: build, start, stop, restart, logs, clean

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_step() {
    echo -e "${BLUE}[STEP]${NC} $1"
}

# 检查Docker和Docker Compose
check_prerequisites() {
    log_step "检查系统依赖..."
    
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装Docker Compose"
        exit 1
    fi
    
    log_info "系统依赖检查完成"
}

# 检查环境变量文件
check_env_file() {
    log_step "检查环境变量配置..."
    
    if [[ ! -f ".env" ]]; then
        if [[ -f ".env.example" ]]; then
            log_warn ".env 文件不存在，正在从 .env.example 创建"
            cp .env.example .env
            log_warn "请编辑 .env 文件并配置您的环境变量"
            echo -e "${YELLOW}特别注意以下配置：${NC}"
            echo -e "  - JWT_SECRET_KEY: JWT密钥（请修改默认值）"
            echo -e "  - DB_HOST, DB_USERNAME, DB_PASSWORD: 数据库连接信息"
            echo -e "  - OSS相关配置（如果使用外部存储）"
            read -p "配置完成后按Enter继续..." -r
        else
            log_error ".env 和 .env.example 文件都不存在"
            exit 1
        fi
    fi
    
    # 检查关键环境变量
    source .env
    if [[ -z "$JWT_SECRET_KEY" || "$JWT_SECRET_KEY" == "your-super-secret-jwt-key-change-me-in-production-32chars" ]]; then
        log_error "请在 .env 文件中设置安全的 JWT_SECRET_KEY"
        exit 1
    fi
    
    if [[ -z "$DB_HOST" || -z "$DB_USERNAME" || -z "$DB_PASSWORD" || -z "$DB_NAME" ]]; then
        log_error "请在 .env 文件中配置数据库连接信息"
        exit 1
    fi
    
    log_info "环境变量配置检查完成"
}

# 构建镜像
build_image() {
    log_step "构建Docker镜像..."
    docker-compose build --no-cache
    log_info "Docker镜像构建完成"
}

# 启动服务
start_service() {
    log_step "启动OSS Manager服务..."
    docker-compose up -d
    
    log_info "等待服务启动..."
    sleep 10
    
    # 检查服务状态
    if docker-compose ps | grep -q "Up"; then
        log_info "服务启动成功！"
        echo -e "${GREEN}访问地址：${NC} http://localhost:${APP_PORT:-8080}"
        echo -e "${GREEN}健康检查：${NC} http://localhost:${APP_PORT:-8080}/api/v1/health"
    else
        log_error "服务启动失败，请检查日志"
        docker-compose logs
    fi
}

# 停止服务
stop_service() {
    log_step "停止OSS Manager服务..."
    docker-compose down
    log_info "服务已停止"
}

# 重启服务
restart_service() {
    log_step "重启OSS Manager服务..."
    stop_service
    start_service
}

# 查看日志
show_logs() {
    log_step "显示服务日志..."
    docker-compose logs -f
}

# 清理资源
clean_resources() {
    log_step "清理Docker资源..."
    read -p "这将删除所有容器、镜像和数据卷，确认继续？(y/N): " -r
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose down -v --rmi all
        docker system prune -f
        log_info "Docker资源清理完成"
    else
        log_info "操作已取消"
    fi
}

# 显示帮助信息
show_help() {
    echo -e "${BLUE}OSS Manager Docker 部署工具${NC}"
    echo ""
    echo "使用方法:"
    echo "  $0 [action]"
    echo ""
    echo "可用操作:"
    echo -e "  ${GREEN}build${NC}    - 构建Docker镜像"
    echo -e "  ${GREEN}start${NC}    - 启动服务"
    echo -e "  ${GREEN}stop${NC}     - 停止服务"
    echo -e "  ${GREEN}restart${NC}  - 重启服务"
    echo -e "  ${GREEN}logs${NC}     - 查看日志"
    echo -e "  ${GREEN}clean${NC}    - 清理所有Docker资源"
    echo -e "  ${GREEN}help${NC}     - 显示此帮助信息"
    echo ""
    echo "示例:"
    echo "  $0 build && $0 start    # 构建并启动"
    echo "  $0 logs                 # 查看实时日志"
    echo ""
}

# 主函数
main() {
    local action=${1:-help}
    
    case $action in
        build)
            check_prerequisites
            check_env_file
            build_image
            ;;
        start)
            check_prerequisites
            check_env_file
            start_service
            ;;
        stop)
            stop_service
            ;;
        restart)
            restart_service
            ;;
        logs)
            show_logs
            ;;
        clean)
            clean_resources
            ;;
        help|--help|-h)
            show_help
            ;;
        *)
            log_error "未知操作: $action"
            show_help
            exit 1
            ;;
    esac
}

# 运行主函数
main "$@"
