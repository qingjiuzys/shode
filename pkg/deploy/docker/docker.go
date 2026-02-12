// Package docker Docker 部署工具
package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// DockerDeployer Docker 部署器
type DockerDeployer struct {
	config     *DockerConfig
	projectDir string
	dryRun     bool
}

// DockerConfig Docker 配置
type DockerConfig struct {
	ImageName      string
	ImageTag       string
	Registry       string
	BaseImage      string
	ExposePorts    []int
	EnvVars        map[string]string
	Volumes        []Volume
	HealthCheck    *HealthCheck
	Resources      *ResourceLimits
}

// Volume 卷挂载
type Volume struct {
	HostPath      string
	ContainerPath string
	Mode          string // "rw" or "ro"
}

// HealthCheck 健康检查
type HealthCheck struct {
	Command      []string
	Interval     int // 秒
	Timeout      int // 秒
	Retries      int
	StartPeriod  int // 秒
}

// ResourceLimits 资源限制
type ResourceLimits struct {
	Memory      string // "512Mi"
	CPU         string // "0.5"
	MemorySwap  string
}

// NewDockerDeployer 创建 Docker 部署器
func NewDockerDeployer(config *DockerConfig) *DockerDeployer {
	return &DockerDeployer{
		config:     config,
		projectDir: ".",
		dryRun:     false,
	}
}

// Init 初始化 Docker 项目
func (dd *DockerDeployer) Init() error {
	fmt.Println("🐳 Initializing Docker project...")

	// 检查是否已有 Dockerfile
	if _, err := os.Stat("Dockerfile"); err == nil {
		return fmt.Errorf("Dockerfile already exists")
	}

	// 生成 Dockerfile
	if err := dd.GenerateDockerfile(); err != nil {
		return fmt.Errorf("failed to generate Dockerfile: %w", err)
	}

	// 生成 .dockerignore
	if err := dd.GenerateDockerignore(); err != nil {
		return fmt.Errorf("failed to generate .dockerignore: %w", err)
	}

	// 生成 docker-compose.yml
	if err := dd.GenerateComposeFile(); err != nil {
		return fmt.Errorf("failed to generate docker-compose.yml: %w", err)
	}

	fmt.Println("✓ Docker project initialized")
	fmt.Println("\nNext steps:")
	fmt.Println("  shode deploy docker build    # Build image")
	fmt.Println("  shode deploy docker compose up  # Start services")

	return nil
}

// GenerateDockerfile 生成 Dockerfile
func (dd *DockerDeployer) GenerateDockerfile() error {
	dockerfile := `# Multi-stage build for Shode application
# Stage 1: Build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/shode

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Set timezone
ENV TZ=Asia/Shanghai

# Create non-root user
RUN addgroup -g 1000 shode && \
    adduser -D -u 1000 -G shode shode

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/app .

# Copy configuration
COPY config ./config

# Change ownership
RUN chown -R shode:shode /app

# Switch to non-root user
USER shode

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run application
CMD ["./app"]
`

	return os.WriteFile("Dockerfile", []byte(dockerfile), 0644)
}

// GenerateDockerignore 生成 .dockerignore
func (dd *DockerDeployer) GenerateDockerignore() error {
	dockerignore := `# Git
.git
.gitignore

# Documentation
*.md
docs/

# Dependencies
vendor/

# Test files
*_test.shode
tests/
test/

# IDE
.vscode/
.idea/
*.swp
*.swo
*~

# OS
.DS_Store
Thumbs.db

# Build artifacts
dist/
build/
*.shodec
*.exe

# Environment
.env
.env.local
.env.*.local

# Logs
*.log
logs/
`

	return os.WriteFile(".dockerignore", []byte(dockerignore), 0644)
}

// GenerateComposeFile 生成 docker-compose.yml
func (dd *DockerDeployer) GenerateComposeFile() error {
	composeFile := `version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=production
      - LOG_LEVEL=info
    volumes:
      - ./config:/app/config:ro
      - app-data:/app/data
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "wget", "--spider", "localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
    networks:
      - shode-network

  # Optional: PostgreSQL database
  postgres:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: shode_db
      POSTGRES_USER: shode
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    networks:
      - shode-network
    restart: unless-stopped

  # Optional: Redis cache
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    networks:
      - shode-network
    restart: unless-stopped

  # Optional: Nginx reverse proxy
  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - app
    networks:
      - shode-network
    restart: unless-stopped

volumes:
  app-data:
  postgres_data:
  redis_data:

networks:
  shode-network:
    driver: bridge
`

	return os.WriteFile("docker-compose.yml", []byte(composeFile), 0644)
}

// Build 构建镜像
func (dd *DockerDeployer) Build(ctx context.Context, tag string) error {
	imageName := dd.getImageName(tag)

	fmt.Printf("🔨 Building Docker image: %s\n", imageName)

	args := []string{"build", "-t", imageName, "."}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Push 推送镜像
func (dd *DockerDeployer) Push(ctx context.Context, tag string) error {
	imageName := dd.getImageName(tag)

	fmt.Printf("📤 Pushing Docker image: %s\n", imageName)

	args := []string{"push", imageName}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Run 运行容器
func (dd *DockerDeployer) Run(ctx context.Context, tag string) error {
	imageName := dd.getImageName(tag)

	fmt.Printf("▶️  Running container: %s\n", imageName)

	args := []string{"run", "-d", "-p", "8080:8080", "--name", dd.config.ImageName, imageName}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Stop 停止容器
func (dd *DockerDeployer) Stop(ctx context.Context) error {
	fmt.Printf("⏹️  Stopping container: %s\n", dd.config.ImageName)

	args := []string{"stop", dd.config.ImageName}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Logs 查看日志
func (dd *DockerDeployer) Logs(ctx context.Context, follow bool) error {
	fmt.Printf("📋 Showing logs for: %s\n", dd.config.ImageName)

	args := []string{"logs", dd.config.ImageName}

	if follow {
		args = append(args, "-f")
	}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ComposeUp 使用 Docker Compose 启动服务
func (dd *DockerDeployer) ComposeUp(ctx context.Context) error {
	fmt.Println("🚀 Starting services with Docker Compose...")

	args := []string{"compose", "up", "-d"}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker-compose %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// ComposeDown 使用 Docker Compose 停止服务
func (dd *DockerDeployer) ComposeDown(ctx context.Context) error {
	fmt.Println("⏹️  Stopping services with Docker Compose...")

	args := []string{"compose", "down"}

	if dd.dryRun {
		fmt.Printf("[DRY RUN] docker-compose %s\n", strings.Join(args, " "))
		return nil
	}

	cmd := exec.Command("docker-compose", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// getImageName 获取镜像名称
func (dd *DockerDeployer) getImageName(tag string) string {
	if tag == "" {
		tag = dd.config.ImageTag
	}
	if tag == "" {
		tag = "latest"
	}

	imageName := dd.config.ImageName + ":" + tag

	if dd.config.Registry != "" {
		imageName = dd.config.Registry + "/" + imageName
	}

	return imageName
}

// SetDryRun 设置是否为模拟运行
func (dd *DockerDeployer) SetDryRun(dryRun bool) {
	dd.dryRun = dryRun
}

// OptimizeImage 优化镜像大小
func (dd *DockerDeployer) OptimizeImage(ctx context.Context) error {
	fmt.Println("🔧 Optimizing Docker image...")

	// 使用多阶段构建
	// 使用 alpine 基础镜像
	// 清理不需要的文件
	// 压缩层

	return nil
}
