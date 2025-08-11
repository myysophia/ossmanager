# OSS Manager Monolith Build Script

# 项目信息
APP_NAME := ossmanager
VERSION := $(shell git describe --tags --always 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date +%Y-%m-%d\ %H:%M:%S)
GO_VERSION := $(shell go version | cut -d " " -f 3)

# 路径配置
FRONTEND_DIR := frontend
WEB_DIR := web
BUILD_DIR := build
DIST_DIR := dist

# Go构建参数
LDFLAGS := -s -w -X 'main.Version=$(VERSION)' -X 'main.BuildTime=$(BUILD_TIME)' -X 'main.GoVersion=$(GO_VERSION)'
GO_BUILD := go build -ldflags "$(LDFLAGS)"

# Node.js参数
NODE_OPTIONS := --max-old-space-size=4096

# 默认目标 - 构建简化单体应用
.PHONY: all
all: build-simple

# 清理构建产物
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -rf $(WEB_DIR)/build
	@rm -f $(APP_NAME)
	@rm -f $(APP_NAME).exe

# 准备目录结构
.PHONY: prepare
prepare:
	@echo "📁 Preparing directories..."
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(DIST_DIR)
	@mkdir -p $(WEB_DIR)

# 构建前端
.PHONY: frontend
frontend: prepare
	@echo "🏗️  Building frontend..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && \
		NODE_OPTIONS="$(NODE_OPTIONS)" npm run build && \
		mkdir -p ../$(WEB_DIR)/build && \
		cp -r out/* ../$(WEB_DIR)/build/ 2>/dev/null && \
		echo "✅ Frontend build completed"; \
	else \
		echo "⚠️  Frontend directory not found, creating placeholder..."; \
		mkdir -p $(WEB_DIR)/build; \
		echo '<!DOCTYPE html><html><head><title>OSS Manager</title></head><body><h1>OSS Manager</h1><p>API is running</p></body></html>' > $(WEB_DIR)/build/index.html; \
	fi

# 构建前端 (开发模式)
.PHONY: frontend-dev
frontend-dev: prepare
	@echo "🏗️  Building frontend for development..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && \
		NODE_OPTIONS="$(NODE_OPTIONS)" npm run build && \
		mkdir -p ../$(WEB_DIR)/build && \
		cp -r .next/standalone/* ../$(WEB_DIR)/build/ && \
		cp -r .next/static ../$(WEB_DIR)/build/_next/; \
	fi

# 跨平台构建简化单体应用
.PHONY: build-cross
build-cross: prepare
	@echo "🏗️  Building monolith for multiple platforms..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		echo "📦 Building frontend..."; \
		cd $(FRONTEND_DIR) && \
		NODE_OPTIONS="$(NODE_OPTIONS)" npm run build && \
		mkdir -p ../$(WEB_DIR)/build && \
		cp -r out/* ../$(WEB_DIR)/build/ 2>/dev/null && \
		echo "✅ Frontend build completed"; \
	else \
		echo "⚠️  Frontend directory not found, creating placeholder..."; \
		mkdir -p $(WEB_DIR)/build; \
		echo '<!DOCTYPE html><html><head><title>OSS Manager</title></head><body><h1>OSS Manager</h1><p>API is running</p></body></html>' > $(WEB_DIR)/build/index.html; \
	fi
	@echo "🏗️  Building Go binaries for multiple platforms..."
	GOOS=linux GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 ./main.go && \
	GOOS=darwin GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 ./main.go && \
	GOOS=windows GOARCH=amd64 $(GO_BUILD) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe ./main.go
	@echo "✅ Cross-platform build completed"

# 生产环境构建（使用简化单体应用）
.PHONY: build-prod
build-prod: build-simple
	@echo "📦 Creating production build..."
	@mkdir -p $(DIST_DIR)
	@cp $(APP_NAME) $(DIST_DIR)/
	@cp -r configs $(DIST_DIR)/ 2>/dev/null || true
	@echo "✅ Production build completed: $(DIST_DIR)/$(APP_NAME)"

# 运行测试
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# 代码格式化
.PHONY: fmt
fmt:
	@echo "💄 Formatting code..."
	@go fmt ./...
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm run lint:fix 2>/dev/null || true; \
	fi

# 代码检查
.PHONY: lint
lint:
	@echo "🔍 Running code analysis..."
	@go vet ./... && \
	golangci-lint run 2>/dev/null || echo "⚠️  golangci-lint not installed"

# 依赖管理
.PHONY: deps
deps:
	@echo "📦 Installing dependencies..."
	@go mod download && go mod tidy
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		cd $(FRONTEND_DIR) && npm ci; \
	fi

# Docker构建
.PHONY: docker
docker:
	@echo "🐳 Building Docker image..."
	@docker build -t $(APP_NAME):$(VERSION) .

# 创建发布包
.PHONY: release
release: build-prod
	@echo "📦 Creating release package..."
	@mkdir -p $(DIST_DIR)/release
	@cp $(DIST_DIR)/$(APP_NAME) $(DIST_DIR)/release/
	@cp -r configs $(DIST_DIR)/release/ 2>/dev/null || true
	@cp README.md $(DIST_DIR)/release/ 2>/dev/null || true
	@cd $(DIST_DIR) && tar -czf $(APP_NAME)-$(VERSION).tar.gz release/
	@echo "✅ Release package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION).tar.gz"

# 安装开发工具
.PHONY: install-tools
install-tools:
	@echo "🛠️  Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install github.com/air-verse/air@latest

# 监控文件变化并重新构建
.PHONY: watch
watch:
	@echo "👀 Watching for changes..."
	@air

# 构建简化单体应用
.PHONY: build-simple
build-simple: prepare
	@echo "🏗️  Building simplified monolith application..."
	@if [ -d "$(FRONTEND_DIR)" ]; then \
		echo "📦 Building frontend..."; \
		cd $(FRONTEND_DIR) && \
		NODE_OPTIONS="$(NODE_OPTIONS)" npm run build && \
		mkdir -p ../$(WEB_DIR)/build && \
		cp -r out/* ../$(WEB_DIR)/build/ 2>/dev/null && \
		echo "✅ Frontend build completed"; \
	else \
		echo "⚠️  Frontend directory not found, creating placeholder..."; \
		mkdir -p $(WEB_DIR)/build; \
		echo '<!DOCTYPE html><html><head><title>OSS Manager</title></head><body><h1>OSS Manager</h1><p>API is running</p></body></html>' > $(WEB_DIR)/build/index.html; \
	fi
	@echo "🏗️  Building Go monolith binary..."
	@$(GO_BUILD) -o $(APP_NAME) ./main.go
	@echo "✅ Simplified monolith build completed: ./$(APP_NAME)"

# 启动简化单体应用 (开发模式)
.PHONY: dev-simple
dev-simple:
	@echo "🚀 Starting simplified monolith in development mode..."
	@APP_ENV=dev go run ./main.go

# 启动简化单体应用 (生产模式)
.PHONY: run-simple
run-simple: build-simple
	@echo "🚀 Starting simplified monolith application..."
	@APP_ENV=prod ./$(APP_NAME)

# 显示帮助信息
.PHONY: help
help:
	@echo "OSS Manager Monolith Build Commands:"
	@echo ""
	@echo "  🚀 Main Commands:"
	@echo "  make all           - 构建单体应用 (默认)"
	@echo "  make build-simple  - 构建单体应用"
	@echo "  make dev-simple    - 开发模式运行"
	@echo "  make run-simple    - 生产模式运行"
	@echo "  make clean         - 清理构建产物"
	@echo ""
	@echo "  🔧 Development:"
	@echo "  make frontend      - 单独构建前端"
	@echo "  make test          - 运行测试"
	@echo "  make fmt           - 代码格式化"
	@echo "  make lint          - 代码检查"
	@echo "  make deps          - 安装依赖"
	@echo ""
	@echo "  📦 Production:"
	@echo "  make build-prod    - 生产环境构建"
	@echo "  make build-cross   - 跨平台构庺"
	@echo "  make release       - 创建发布包"
	@echo "  make docker        - 构建Docker镜像"
	@echo ""
	@echo "  ⚙️  Utilities:"
	@echo "  make install-tools - 安装开发工具"
	@echo "  make watch         - 监控文件变化"
	@echo "  make help          - 显示此帮助信息"
