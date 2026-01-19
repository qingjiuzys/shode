# Shode 命令速查表

> 像使用命令行工具一样使用 Shode

---

## 🌐 HTTP 服务器

| 命令 | 说明 | 示例 |
|------|------|------|
| `StartHTTPServer <port>` | 启动服务器 | `StartHTTPServer 8080` |
| `StopHTTPServer` | 停止服务器 | `StopHTTPServer` |
| `IsHTTPServerRunning` | 检查状态 | `IsHTTPServerRunning` |

---

## 🗄 数据库

| 命令 | 说明 | 示例 |
|------|------|------|
| `ConnectDB <type> <dsn>` | 连接数据库 | `ConnectDB "sqlite" "app.db"` |
| `QueryDB <sql> <args>` | 执行查询 | `QueryDB "SELECT * FROM users"` |
| `ExecDB <sql> <args>` | 执行修改 | `ExecDB "INSERT INTO users..."` |
| `GetQueryResult` | 获取结果 | `result = GetQueryResult` |
| `CloseDB` | 关闭连接 | `CloseDB` |

---

## 💾 缓存

| 命令 | 说明 | 示例 |
|------|------|------|
| `SetCache <key> <value> <ttl>` | 设置缓存 | `SetCache "users" 'json' 300` |
| `GetCache <key>` | 获取缓存 | `cached = GetCache "users"` |
| `DeleteCache <key>` | 删除缓存 | `DeleteCache "users"` |
| `ClearCache` | 清空缓存 | `ClearCache` |
| `CacheExists <key>` | 检查存在 | `exists = CacheExists "users"` |
| `GetCacheKeys <pattern>` | 获取键列表 | `keys = GetCacheKeys "*"` |

---

## 📡 HTTP 请求

| 命令 | 说明 | 示例 |
|------|------|------|
| `GetHTTPMethod` | 获取方法 | `method = GetHTTPMethod` |
| `GetHTTPPath` | 获取路径 | `path = GetHTTPPath` |
| `GetHTTPQuery <param>` | 获取参数 | `name = GetHTTPQuery "name"` |
| `GetHTTPHeader <header>` | 获取头 | `auth = GetHTTPHeader "Authorization"` |

---

## 📤 HTTP 响应

| 命令 | 说明 | 示例 |
|------|------|------|
| `SetHTTPResponse <code> <body>` | 设置响应 | `SetHTTPResponse 200 '{"data":"ok"}'` |
| `SetHTTPHeader <name> <value>` | 设置头 | `SetHTTPHeader "Content-Type" "application/json"` |

---

## 🛣 路由注册

| 命令 | 说明 | 示例 |
|------|------|------|
| `RegisterHTTPRoute <method> <path> <type> <handler>` | 注册路由 | `RegisterHTTPRoute "GET" "/" "function" "handler"` |
| `RegisterRouteWithResponse <path> <response>` | 简单路由 | `RegisterRouteWithResponse "/" "Hello"` |

---

## 🔧 字符串处理

| 命令 | 说明 | 示例 |
|------|------|------|
| `Upper <text>` | 大写 | `Upper "hello"` → `HELLO` |
| `Lower <text>` | 小写 | `Lower "HELLO"` → `hello` |
| `Trim <text>` | 去除空格 | `Trim "  text  "` → `text` |
| `Contains <text> <substring>` | 包含检查 | `Contains "hello" "ell"` → `true` |

---

## 📁 文件操作

| 命令 | 说明 | 示例 |
|------|------|------|
| `ReadFile <path>` | 读文件 | `content = ReadFile "file.txt"` |
| `WriteFile <path> <content>` | 写文件 | `WriteFile "file.txt" "content"` |
| `FileExists <path>` | 检查存在 | `if FileExists "file.txt"` |

---

## 🔄 控制流

```sh
# If 语句
if FileExists "file.txt" {
    Println "File exists"
}

# For 循环
for item in 1 2 3 {
    Println "Item: " + item
}

# While 循环
counter = 0
while counter < 10 {
    Println "Counter: " + counter
    counter = counter + 1
}
```

---

## 💡 常见模式

### RESTful API

```sh
# GET - 获取所有
RegisterHTTPRoute "GET" "/items" "function" "getAll"

# POST - 创建
RegisterHTTPRoute "POST" "/items" "function" "create"

# PUT - 更新
RegisterHTTPRoute "PUT" "/items/:id" "function" "update"

# DELETE - 删除
RegisterHTTPRoute "DELETE" "/items/:id" "function" "delete"
```

### 数据库操作

```sh
# 创建表
ExecDB "CREATE TABLE users (id INTEGER, name TEXT)"

# 插入数据
ExecDB "INSERT INTO users (name) VALUES (?)" "Alice"

# 查询数据
QueryDB "SELECT * FROM users"
result = GetQueryResult

# 更新数据
ExecDB "UPDATE users SET name = ? WHERE id = ?" "Bob" userId

# 删除数据
ExecDB "DELETE FROM users WHERE id = ?" userId
```

### 缓存策略

```sh
# 检查缓存
cached = GetCache "data:all"

if cached != "" {
    SetHTTPResponse 200 "$cached"
    return
}

# 查询数据库
QueryDB "SELECT * FROM data"
result = GetQueryResult

# 存入缓存（5 分钟）
SetCache "data:all" result 300
```

---

## 🎯 快速参考

### 最小 API

```sh
StartHTTPServer 8080
function api() {
    SetHTTPResponse 200 '{"status":"ok"}'
}
RegisterHTTPRoute "GET" "/" "function" "api"
```

### 带 CRUD 的完整 API

```sh
StartHTTPServer 8080
ConnectDB "sqlite" "app.db"
ExecDB "CREATE TABLE items (id INTEGER, name TEXT)"

function create() {
    ExecDB "INSERT INTO items (name) VALUES (?)" GetHTTPQuery "name"
    SetHTTPResponse 201 "{}"
}

function read() {
    QueryDB "SELECT * FROM items"
    SetHTTPResponse 200 GetQueryResult
}

function update() {
    ExecDB "UPDATE items SET name = ? WHERE id = ?" GetHTTPQuery "name" GetHTTPQuery "id")
    SetHTTPResponse 200 "{}"
}

function delete() {
    ExecDB "DELETE FROM items WHERE id = ?" GetHTTPQuery "id")
    SetHTTPResponse 204 ""
}

RegisterHTTPRoute "POST" "/items" "function" "create"
RegisterHTTPRoute "GET" "/items" "function" "read"
RegisterHTTPRoute "PUT" "/items" "function" "update"
RegisterHTTPRoute "DELETE" "/items" "function" "delete"
```

---

## 🔐 安全最佳实践

```sh
# 参数化查询（防注入）
QueryDB "SELECT * FROM users WHERE id = ?" userId

# 密码哈希
passwordHash = SHA256Hash password
ExecDB "UPDATE users SET password = ? WHERE id = ?" passwordHash userId

# 会话管理
token = SHA256Hash username + "salt"
SetCache "session:" + token username 3600
```

---

## 🚀 部署

### Docker 部署

```dockerfile
FROM alpine:latest

COPY shode /usr/local/bin/shode
RUN chmod +x /usr/local/bin/shode

COPY api.sh /app/api.sh
WORKDIR /app

EXPOSE 8080

CMD ["/usr/local/bin/shode", "run", "api.sh"]
```

```bash
docker build -t shode-app .
docker run -p 8080:8080 shode-app
```

### systemd 服务

```ini
[Unit]
Description=Shode HTTP API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/shode-app
ExecStart=/usr/local/bin/shode run /opt/shode-app/api.sh
Restart=always

[Install]
WantedBy=multi-user.target
```

```bash
systemctl enable shode-app
systemctl start shode-app
```

---

## 📖 更多文档

- [极简入门指南](index.md) - 30 秒上手
- [用户指南](guides/user-guide.md) - 详细操作指南
- [示例集合](../examples/index.md) - 完整示例
- [API 参考](../api/stdlib.md) - 完整 API 文档

---

**开始使用 Shode！**
