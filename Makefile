# Shode 开发工具 Makefile

.PHONY: help test lint fmt build clean install coverage benchmark refcheck

help: ## 显示帮助信息
	@echo "Shode 开发命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

## 代码质量

lint: ## 运行代码检查
	@echo "运行 golangci-lint..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "⚠️  golangci-lint 未安装，跳过"; \
		echo "安装: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b \$$(go env GOPATH)/bin v1.55.2"; \
	fi

fmt: ## 格式化代码
	@echo "格式化代码..."
	go fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

vet: ## 运行 go vet
	@echo "运行 go vet..."
	go vet ./...

## 测试

test: ## 运行所有测试
	@echo "运行测试..."
	go test -v ./...

test-coverage: ## 运行测试并生成覆盖率报告
	@echo "运行测试覆盖率检查..."
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "覆盖率报告: coverage.html"

test-race: ## 运行竞态检测
	@echo "运行竞态检测..."
	go test -race ./...

benchmark: ## 运行性能基准测试
	@echo "运行性能基准测试..."
	go test -bench=. -benchmem ./...

## 构建

build: ## 构建项目
	@echo "构建 shode..."
	go build -o shode ./cmd/shode
	@echo "✅ 构建完成: ./shode"

build-all: ## 构建所有平台
	@echo "构建多平台版本..."
	@mkdir -p bin
	GOOS=linux GOARCH=amd64 go build -o bin/shode-linux-amd64 ./cmd/shode
	GOOS=darwin GOARCH=amd64 go build -o bin/shode-darwin-amd64 ./cmd/shode
	GOOS=windows GOARCH=amd64 go build -o bin/shode-windows-amd64.exe ./cmd/shode
	@echo "✅ 多平台构建完成: bin/"

install: ## 安装到 $GOPATH/bin
	@echo "安装 shode..."
	go install ./cmd/shode

clean: ## 清理构建文件
	@echo "清理构建文件..."
	rm -f shode
	rm -f coverage.out coverage.html
	rm -rf bin/
	@echo "✅ 清理完成"

## 代码分析

complexity: ## 分析代码复杂度
	@echo "分析代码复杂度..."
	@if command -v gocyclo >/dev/null 2>&1; then \
		gocyclo -over 15 pkg/; \
	else \
		echo "⚠️  gocyclo 未安装"; \
		echo "安装: go install github.com/fzipp/gocyclo/cmd/gocyclo@latest"; \
	fi

deps: ## 分析依赖关系
	@echo "分析依赖..."
	go mod graph
	go mod verify

update-deps: ## 更新依赖
	@echo "更新依赖..."
	go get -u ./...
	go mod tidy

## 开发工具

refcheck: ## 检查需要重构的代码
	@echo "检查重构候选..."
	@echo "超过 100 行的文件:"
	@find pkg -name "*.go" -not -name "*_test.go" -exec wc -l {} \; | awk '$$1 > 100 {print $$0}' | sort -rn

dev: ## 开发模式（自动重载）
	@echo "启动开发模式..."
	@if command -v air >/dev/null 2>&1; then \
		air; \
	else \
		echo "⚠️  air 未安装，请先安装: go install github.com/cosmtrek/air@latest"; \
	fi

## CI/CD

ci: lint test vet ## 运行完整 CI 检查
	@echo "✅ CI 检查通过"

ci-local: ## 本地运行 CI 检查（包含覆盖率）
	@echo "运行本地 CI..."
	@$(MAKE) fmt
	@$(MAKE) lint
	@$(MAKE) vet
	@$(MAKE) test-coverage
	@echo "✅ 本地 CI 完成"

## 发布

release: clean ci test-race build ## 发布前检查
	@echo "准备发布..."
	@$(MAKE) build-all
	@echo "✅ 发布准备完成"

## 文档

docs: ## 生成文档
	@echo "生成文档..."
	go doc -all ./...

## 快速命令

quick: fmt build test ## 快速检查（格式化 + 构建 + 测试）
	@echo "✅ 快速检查完成"

all: clean fmt lint test build ## 完整构建流程
	@echo "✅ 所有步骤完成"
