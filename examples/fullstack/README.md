# Shode TODO 应用 - 全栈示例

这是一个使用 Shode 框架构建的完整 TODO 应用，展示了框架的核心功能。

## ✨ 功能特性

### 后端功能
- ✅ RESTful API 设计
- ✅ WebSocket 实时更新
- ✅ CRUD 操作
- ✅ 结构化日志
- ✅ 实时通信

### 前端功能
- ✅ 响应式设计
- ✅ 实时UI更新（通过WebSocket）
- ✅ 优雅的动画效果
- ✅ 连接状态显示

## 🚀 快速开始

### 运行应用

```bash
cd examples/fullstack
go run main.go
```

### 访问应用

打开浏览器访问：http://localhost:8080

## 📡 API 端点

| 方法 | 端点 | 描述 |
|------|------|------|
| GET | `/api/todos` | 获取所有TODO |
| POST | `/api/todos` | 创建新TODO |
| GET | `/api/todos/:id` | 获取单个TODO |
| PUT | `/api/todos/:id` | 更新TODO |
| DELETE | `/api/todos/:id` | 删除TODO |
| POST | `/api/todos/:id/toggle` | 切换TODO状态 |
| WS | `/ws` | WebSocket连接 |

## 🔧 技术栈

### 后端
- **Shode Framework** - 核心框架
- **WebSocket** - 实时通信
- **Logger** - 结构化日志
- **Router** - HTTP路由
- **Middleware** - 中间件系统

### 前端
- **原生 JavaScript** - 无框架依赖
- **WebSocket API** - 实时更新
- **Fetch API** - HTTP请求
- **CSS3** - 现代样式

## 📁 项目结构

```
examples/fullstack/
├── main.go           # 后端代码
└── static/
    └── index.html    # 前端代码
```

## 💡 使用示例

### 添加TODO

```bash
curl -X POST http://localhost:8080/api/todos \
  -H "Content-Type: application/json" \
  -d '{"title":"学习 Shode 框架","completed":false}'
```

### 切换TODO状态

```bash
curl -X POST http://localhost:8080/api/todos/1/toggle
```

### 删除TODO

```bash
curl -X DELETE http://localhost:8080/api/todos/1
```

## 🔄 WebSocket 实时更新

当任何客户端修改TODO时，所有连接的客户端都会自动收到更新：

```javascript
// WebSocket消息格式
{
  "type": "todo_created",
  "data": {
    "id": 1,
    "title": "...",
    "completed": false
  },
  "time": "2024-01-01T00:00:00Z"
}
```

消息类型：
- `todo_created` - 新TODO创建
- `todo_updated` - TODO更新
- `todo_deleted` - TODO删除

## 🎨 特性展示

本示例展示了 Shode 框架的以下功能：

1. **Web路由系统** (`pkg/web/`)
   - RESTful路由定义
   - 路径参数提取
   - 中间件支持

2. **实时通信** (`pkg/realtime/websocket/`)
   - WebSocket Hub模式
   - 消息广播
   - 连接管理

3. **日志系统** (`pkg/logger/`)
   - 结构化日志
   - 多级别日志
   - 日志格式化

4. **中间件系统** (`pkg/middleware/`)
   - 日志中间件
   - 恢复中间件
   - CORS中间件

## 🎯 学习要点

### RESTful API设计
```go
// 注册路由
r.Get("/api/todos", h.ListTodos)
r.Post("/api/todos", h.CreateTodo)
r.Get("/api/todos/:id", h.GetTodo)
```

### WebSocket广播
```go
// 广播到所有客户端
hub.Broadcast(websocket.Message{
    Type: "todo_created",
    Data: todo,
    Time: time.Now(),
})
```

### 路径参数提取
```go
id := web.PathParam(r, "id")
```

## 📝 扩展建议

1. **添加数据持久化**
   - 集成数据库 (PostgreSQL/MySQL)
   - 使用数据库迁移工具

2. **添加用户认证**
   - JWT token认证
   - 用户管理

3. **添加更多功能**
   - TODO标签/分类
   - 过期日期
   - 优先级

4. **添加测试**
   - 单元测试
   - 集成测试
   - 使用测试工具包

5. **添加文档**
   - API文档 (使用apidoc工具)
   - 使用指南

## 🚀 部署

### 使用 Shode CLI 部署

```bash
# 构建应用
shode build

# 运行
./fullstack
```

### Docker部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o fullstack ./examples/fullstack

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/fullstack .
EXPOSE 8080
CMD ["./fullstack"]
```

## 📚 相关文档

- [Shode 文档](../../docs/)
- [Web路由指南](../../docs/web.md)
- [WebSocket指南](../../docs/websocket.md)
- [中间件文档](../../docs/middleware.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
