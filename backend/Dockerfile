# 构建阶段
FROM golang:1.24-alpine AS builder

# 设置工作目录
WORKDIR /app

# 安装必要的系统依赖
RUN apk add --no-cache gcc musl-dev

# 复制go mod文件
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -o ossmanager ./cmd

# 运行阶段
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk add --no-cache ca-certificates tzdata

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
COPY --from=builder /app/ossmanager .
COPY --from=builder /app/configs ./configs

# 使用非root用户运行
USER appuser

# 暴露端口
EXPOSE 8080

# 启动应用
CMD ["./ossmanager"] 