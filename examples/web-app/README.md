# Web 应用示例 - RESTful API 服务

一个使用 Shode 框架构建的现代化 RESTful API 服务示例。

## 功能特性

- ✅ RESTful API 设计
- ✅ 用户认证和授权
- ✅ 数据库集成
- ✅ 缓存优化
- ✅ 请求限流
- ✅ 日志记录
- ✅ 性能监控
- ✅ 错误处理
- ✅ API 文档（Swagger）

## 技术栈

- **框架**: Shode v0.6.0
- **数据库**: PostgreSQL (通过 pkg/database)
- **缓存**: Redis (通过 pkg/cache)
- **监控**: OpenTelemetry (通过 pkg/trace)
- **部署**: Docker + Docker Compose

## 项目结构

```
web-app/
├── main.shode          # 主应用入口
├── config.shode        # 配置文件
├── Dockerfile          # Docker 镜像
├── docker-compose.yml  # Docker Compose 配置
├── README.md           # 项目说明
└── tests/              # 测试用例
    └── api_test.shode
```

## 快速开始

### 1. 本地运行

```bash
# 安装依赖
shode install

# 启动服务（开发模式）
shode run main.shode

# 运行测试
shode test tests/
```

### 2. Docker 运行

```bash
# 构建镜像
docker-compose build

# 启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down
```

### 3. 使用 API

```bash
# 健康检查
curl http://localhost:8080/health

# 用户注册
curl -X POST http://localhost:8080/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","email":"test@example.com","password":"password123"}'

# 用户登录
curl -X POST http://localhost:8080/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"password123"}'

# 获取用户信息（需要认证）
curl http://localhost:8080/api/v1/users/profile \
  -H "Authorization: Bearer YOUR_TOKEN"

# 创建文章
curl -X POST http://localhost:8080/api/v1/articles \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"My First Article","content":"This is the content..."}'

# 获取文章列表
curl http://localhost:8080/api/v1/articles?page=1&limit=10
```

## API 端点

### 用户管理

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| POST | `/api/v1/users/register` | 用户注册 | 否 |
| POST | `/api/v1/users/login` | 用户登录 | 否 |
| GET | `/api/v1/users/profile` | 获取用户信息 | 是 |
| PUT | `/api/v1/users/profile` | 更新用户信息 | 是 |
| DELETE | `/api/v1/users/profile` | 删除用户 | 是 |

### 文章管理

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/api/v1/articles` | 获取文章列表 | 否 |
| GET | `/api/v1/articles/:id` | 获取文章详情 | 否 |
| POST | `/api/v1/articles` | 创建文章 | 是 |
| PUT | `/api/v1/articles/:id` | 更新文章 | 是 |
| DELETE | `/api/v1/articles/:id` | 删除文章 | 是 |

### 系统管理

| 方法 | 端点 | 描述 | 认证 |
|------|------|------|------|
| GET | `/health` | 健康检查 | 否 |
| GET | `/metrics` | 性能指标 | 否 |
| GET | `/api/v1/stats` | 统计信息 | 是 |

## 配置说明

### config.shode

```javascript
// 服务器配置
server {
    host: "0.0.0.0"
    port: 8080
    mode: "production"  // "development" or "production"
}

// 数据库配置
database {
    driver: "postgres"
    host: "localhost"
    port: 5432
    name: "shode_demo"
    user: "shode"
    password: "password"
    max_open_conns: 25
    max_idle_conns: 5
}

// Redis 配置
redis {
    host: "localhost"
    port: 6379
    password: ""
    db: 0
    pool_size: 10
}

// JWT 配置
jwt {
    secret: "your-secret-key"
    expire_hours: 24
}

// 限流配置
rate_limit {
    enabled: true
    requests_per_minute: 60
    burst: 10
}

// 日志配置
logging {
    level: "info"  // "debug", "info", "warn", "error"
    format: "json"  // "json" or "text"
    output: "stdout"  // "stdout" or "file"
}
```

## 性能优化

### 缓存策略

1. **用户缓存** - 缓存用户信息，TTL: 1小时
2. **文章缓存** - 缓存热门文章，TTL: 30分钟
3. **列表缓存** - 缓存文章列表，TTL: 5分钟

### 数据库优化

1. **连接池** - 使用数据库连接池
2. **索引** - 为常用查询字段添加索引
3. **查询优化** - 使用分页和字段选择

### 并发处理

1. **协程池** - 使用 GoroutinePool 处理并发请求
2. **批量操作** - 批量处理数据库操作

## 监控和日志

### 性能监控

访问 `/metrics` 端点获取性能指标：

- `http_requests_total` - 总请求数
- `http_request_duration_ms` - 请求耗时
- `http_requests_in_flight` - 当前并发请求数
- `cache_hits_total` - 缓存命中数
- `db_connections_active` - 活跃数据库连接数

### 日志格式

```json
{
  "timestamp": "2026-02-01T10:30:45Z",
  "level": "info",
  "method": "GET",
  "path": "/api/v1/articles",
  "status": 200,
  "duration_ms": 45,
  "user_id": "123456"
}
```

## 测试

### 运行测试

```bash
# 运行所有测试
shode test tests/

# 运行单个测试
shode test tests/api_test.shode

# 查看测试覆盖率
shode test --coverage tests/
```

### 测试示例

```javascript
// tests/api_test.shode
import { assert, assertEquals } from "testing"

// 测试健康检查
test("health check", () => {
    response = http_get("/health")
    assertEquals(response.status, 200)
    assert(response.body.status == "ok")
})

// 测试用户注册
test("user registration", () => {
    data = {
        username: "testuser",
        email: "test@example.com",
        password: "password123"
    }
    response = http_post("/api/v1/users/register", data)
    assertEquals(response.status, 201)
    assert(response.body.user.id != null)
})

// 测试用户登录
test("user login", () => {
    data = {
        username: "testuser",
        password: "password123"
    }
    response = http_post("/api/v1/users/login", data)
    assertEquals(response.status, 200)
    assert(response.body.token != null)
})
```

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t shode-web-app .

# 运行容器
docker run -p 8080:8080 \
  -e DATABASE_HOST=postgres \
  -e REDIS_HOST=redis \
  shode-web-app
```

### Docker Compose 部署

```bash
# 启动所有服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f web-app
```

### Kubernetes 部署

```bash
# 创建部署
kubectl apply -f k8s/

# 查看状态
kubectl get pods -l app=shode-web-app

# 查看日志
kubectl logs -f deployment/shode-web-app
```

## 故障排查

### 常见问题

1. **数据库连接失败**
   - 检查数据库是否运行
   - 检查连接配置是否正确
   - 检查网络连接

2. **Redis 连接失败**
   - 检查 Redis 是否运行
   - 检查 Redis 配置

3. **性能问题**
   - 查看性能指标 `/metrics`
   - 检查缓存命中率
   - 检查数据库慢查询

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
