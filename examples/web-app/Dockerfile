# 多阶段构建 Dockerfile

# 第一阶段：构建
FROM golang:1.21-alpine AS builder

# 安装依赖
RUN apk add --no-cache git make

# 设置工作目录
WORKDIR /app

# 复制 go mod 文件
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建 Shode 应用
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o shode-app ./cmd/shode

# 第二阶段：运行
FROM alpine:latest

# 安装必要的运行时依赖
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai

# 创建非 root 用户
RUN addgroup -g 1000 shode && \
    adduser -D -u 1000 -G shode shode

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /app/shode-app .
COPY --from=builder /app/examples/web-app ./examples/web-app

# 修改权限
RUN chown -R shode:shode /app

# 切换到非 root 用户
USER shode

# 暴露端口
EXPOSE 8080

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# 启动应用
CMD ["./shode-app", "run", "examples/web-app/main.shode"]
