# 多阶段构建 Dockerfile for OSS Manager Monolith

# 阶段1：构建前端
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# 安装构建依赖
RUN apk add --no-cache python3 make g++

# 复制前端源码
COPY frontend/package*.json ./
RUN npm ci --legacy-peer-deps

COPY frontend/ ./

# 构建前端
ENV NEXT_TELEMETRY_DISABLED=1
ENV NODE_ENV=production
RUN npm run build

# 准备静态文件
RUN mkdir -p /app/web/build && \
    cp -r .next/standalone/* /app/web/ && \
    cp -r .next/static /app/web/build/_next/static && \
    cp -r public/* /app/web/build/ 2>/dev/null || true

# 阶段2：构建后端
FROM golang:1.23-alpine AS backend-builder

# 安装必要的系统依赖
RUN apk add --no-cache gcc musl-dev git

WORKDIR /app

# 复制go mod文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源码
COPY . ./

# 复制前端构建产物
COPY --from=frontend-builder /app/web/ ./web/

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-w -s' -o ossmanager .

# 阶段3：运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata curl

# 设置时区和环境变量
ENV TZ=Asia/Shanghai
ENV GIN_MODE=release
ENV APP_ENV=prod

# 创建非root用户
RUN adduser -D -g '' appuser

# 设置工作目录
WORKDIR /app

# 创建必要的目录并设置权限
RUN mkdir -p /app/logs /app/tmp/uploads /data/oss && \
    chown -R appuser:appuser /app /data/oss

# 从构建阶段复制二进制文件和配置
COPY --from=backend-builder /app/ossmanager .
COPY --from=backend-builder /app/configs ./configs

# 使用非root用户运行
USER appuser

# 健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=60s --retries=3 \
  CMD curl -f http://localhost:8080/api/v1/health || exit 1

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./ossmanager"]
