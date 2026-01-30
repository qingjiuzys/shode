# Shode API 参考文档

## 目录

- [标准库函数](#标准库函数)
  - [字符串操作](#字符串操作)
  - [文件操作](#文件操作)
  - [环境变量](#环境变量)
  - [HTTP 服务器](#http-服务器)
  - [WebSocket](#websocket)
  - [数据库操作](#数据库操作)
  - [缓存操作](#缓存操作)
- [HTTP 上下文](#http-上下文)
- [错误类型](#错误类型)

---

## 标准库函数

### 字符串操作

#### ToUpper
```bash
ToUpper "string"
```
将字符串转换为大写。

**参数:**
- `string` - 输入字符串

**返回值:**
- 大写字符串

**示例:**
```bash
result := ToUpper "hello"
# result = "HELLO"
```

#### ToLower
```bash
ToLower "string"
```
将字符串转换为小写。

**参数:**
- `string` - 输入字符串

**返回值:**
- 小写字符串

**示例:**
```bash
result := ToLower "WORLD"
# result = "world"
```

#### Trim
```bash
Trim "string"
```
去除字符串两端的空格。

**参数:**
- `string` - 输入字符串

**返回值:**
- 去除空格后的字符串

**示例:**
```bash
result := Trim "  hello  "
# result = "hello"
```

#### Replace
```bash
Replace "string" "old" "new"
```
替换字符串中的内容。

**参数:**
- `string` - 原字符串
- `old` - 要替换的子串
- `new` - 替换后的子串

**返回值:**
- 替换后的字符串

**示例:**
```bash
result := Replace "hello world" "world" "there"
# result = "hello there"
```

#### Contains
```bash
Contains "haystack" "needle"
```
检查字符串是否包含子串。

**参数:**
- `haystack` - 原字符串
- `needle` - 要查找的子串

**返回值:**
- `true` 如果包含，`false` 否则

**示例:**
```bash
if Contains "hello world" "world"; then
    echo "Found!"
fi
```

#### SHA256Hash
```bash
SHA256Hash "string"
```
计算字符串的 SHA256 哈希值。

**参数:**
- `string` - 输入字符串

**返回值:**
- 64 位十六进制哈希字符串

**示例:**
```bash
hash := SHA256Hash "password"
echo $hash
# 输出: 5e884898da28047151d0e56f8dc6292773603d0d6aabbdd62a11ef721d1542d8
```

---

### 文件操作

#### ReadFile
```bash
ReadFile "path"
```
读取文件内容。

**参数:**
- `path` - 文件路径

**返回值:**
- 文件内容字符串
- 错误信息（如果失败）

**示例:**
```bash
content := ReadFile "/etc/hostname"
echo $content
```

#### WriteFile
```bash
WriteFile "path" "content"
```
写入内容到文件。

**参数:**
- `path` - 文件路径
- `content` - 要写入的内容

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
WriteFile "/tmp/output.txt" "Hello, World!"
```

#### FileExists
```bash
FileExists "path"
```
检查文件是否存在。

**参数:**
- `path` - 文件路径

**返回值:**
- `true` 如果文件存在，`false` 否则

**示例:**
```bash
if FileExists "/tmp/file.txt"; then
    echo "File exists"
fi
```

#### ListFiles
```bash
ListFiles "directory"
```
列出目录中的文件。

**参数:**
- `directory` - 目录路径

**返回值:**
- 文件列表数组
- 错误信息（如果失败）

**示例:**
```bash
files := ListFiles "/tmp"
for file in $files; do
    echo $file
done
```

---

### 环境变量

#### GetEnv
```bash
GetEnv "key"
```
获取环境变量值。

**参数:**
- `key` - 环境变量名

**返回值:**
- 环境变量值（空字符串如果不存在）

**示例:**
```bash
path := GetEnv "PATH"
echo $path
```

#### SetEnv
```bash
SetEnv "key" "value"
```
设置环境变量。

**参数:**
- `key` - 变量名
- `value` - 变量值

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
SetEnv "MY_VAR" "my_value"
```

#### UnsetEnv
```bash
UnsetEnv "key"
```
删除环境变量。

**参数:**
- `key` - 变量名

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
UnsetEnv "MY_VAR"
```

#### Export
```bash
Export "key" "value"
```
导出环境变量到子进程。

**参数:**
- `key` - 变量名
- `value` - 变量值

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
Export "API_KEY" "secret_key"
```

---

### HTTP 服务器

#### StartHTTPServer
```bash
StartHTTPServer "port"
```
启动 HTTP 服务器。

**参数:**
- `port` - 端口号（字符串）

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
StartHTTPServer "8080"
```

#### RegisterHTTPRoute
```bash
RegisterHTTPRoute "method" "path" "type" "handler"
```
注册 HTTP 路由。

**参数:**
- `method` - HTTP 方法（GET, POST, PUT, DELETE, *）
- `path` - 路由路径（/api/users）
- `type` - 处理器类型（function, script）
- `handler` - 处理器名称或脚本

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
# 函数处理器
function handleUsers() {
    SetHTTPResponse 200 '{"users":[]}'
}
RegisterHTTPRoute "GET" "/api/users" "function" "handleUsers"

# 脚本处理器
RegisterHTTPRoute "POST" "/api/data" "script" "SetHTTPResponse 201 'Created'"
```

#### RegisterStaticRoute
```bash
RegisterStaticRoute "path" "directory"
```
注册静态文件路由。

**参数:**
- `path` - URL 路径
- `directory` - 文件系统目录

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
RegisterStaticRoute "/" "./public"
RegisterStaticRoute "/docs" "./documentation"
```

#### GetHTTPMethod
```bash
GetHTTPMethod
```
获取当前请求的 HTTP 方法。

**返回值:**
- HTTP 方法字符串（GET, POST, 等）

**示例:**
```bash
method := GetHTTPMethod
echo "Method: $method"
```

#### GetHTTPPath
```bash
GetHTTPPath
```
获取当前请求的路径。

**返回值:**
- 请求路径字符串

**示例:**
```bash
path := GetHTTPPath
echo "Path: $path"
```

#### GetHTTPQuery
```bash
GetHTTPQuery "key"
```
获取 URL 查询参数。

**参数:**
- `key` - 参数名

**返回值:**
- 参数值（空字符串如果不存在）

**示例:**
```bash
# URL: /api/users?page=2
page := GetHTTPQuery "page"
echo "Page: $page"
```

#### GetHTTPBody
```bash
GetHTTPBody
```
获取请求体内容。

**返回值:**
- 请求体字符串

**示例:**
```bash
body := GetHTTPBody
echo "Body: $body"
```

#### SetHTTPResponse
```bash
SetHTTPResponse "status" "body"
```
设置 HTTP 响应。

**参数:**
- `status` - HTTP 状态码（200, 201, 404, 等）
- `body` - 响应体（字符串或 JSON）

**返回值:**
- 无

**示例:**
```bash
SetHTTPResponse 200 '{"status":"ok"}'
SetHTTPResponse 404 '{"error":"Not found"}'
```

---

### WebSocket

#### RegisterWebSocketRoute
```bash
RegisterWebSocketRoute "path" "handler"
```
注册 WebSocket 路由。

**参数:**
- `path` - WebSocket 路径
- `handler` - 处理函数名称（可选）

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
RegisterWebSocketRoute "/ws" ""
RegisterWebSocketRoute "/chat" "handleChat"
```

#### SendWebSocketMessage
```bash
SendWebSocketMessage "connectionID" "message"
```
发送消息给特定连接。

**参数:**
- `connectionID` - 连接 ID
- `message` - 消息内容

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
SendWebSocketMessage "conn_123" "Hello!"
```

#### BroadcastWebSocketMessage
```bash
BroadcastWebSocketMessage "message"
```
广播消息到所有连接。

**参数:**
- `message` - 消息内容

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
BroadcastWebSocketMessage "Server maintenance in 5 minutes"
```

#### BroadcastWebSocketMessageToRoom
```bash
BroadcastWebSocketMessageToRoom "room" "message"
```
广播消息到特定房间。

**参数:**
- `room` - 房间名
- `message` - 消息内容

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
BroadcastWebSocketMessageToRoom "chatroom" "New message!"
```

#### JoinRoom
```bash
JoinRoom "connectionID" "room"
```
让连接加入房间。

**参数:**
- `connectionID` - 连接 ID
- `room` - 房间名

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
JoinRoom "conn_123" "general"
```

#### LeaveRoom
```bash
LeaveRoom "connectionID"
```
让连接离开当前房间。

**参数:**
- `connectionID` - 连接 ID

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
LeaveRoom "conn_123"
```

#### GetWebSocketConnectionCount
```bash
GetWebSocketConnectionCount
```
获取当前连接数。

**返回值:**
- 连接数（整数）

**示例:**
```bash
count := GetWebSocketConnectionCount
echo "Connections: $count"
```

---

### 数据库操作

#### ConnectDB
```bash
ConnectDB "type" "dsn"
```
连接数据库。

**参数:**
- `type` - 数据库类型（mysql, postgresql, sqlite）
- `dsn` - 数据源名称（连接字符串）

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
# SQLite
ConnectDB "sqlite" "./data.db"

# MySQL
ConnectDB "mysql" "user:password@tcp(localhost:3306)/dbname"

# PostgreSQL
ConnectDB "postgresql" "host=localhost port=5432 user=user password=password dbname=db"
```

#### QueryDB
```bash
QueryDB "sql" "arg1" "arg2" ...
```
执行查询 SQL。

**参数:**
- `sql` - SQL 查询语句
- `arg1`, `arg2`, ... - 参数（可选）

**返回值:**
- 查询结果
- 错误信息（如果失败）

**示例:**
```bash
result := QueryDB "SELECT * FROM users WHERE id = ?" "123"
```

#### QueryRowDB
```bash
QueryRowDB "sql" "arg1" "arg2" ...
```
查询单行数据。

**参数:**
- `sql` - SQL 查询语句
- `arg1`, `arg2`, ... - 参数（可选）

**返回值:**
- 查询结果
- 错误信息（如果失败）

**示例:**
```bash
row := QueryRowDB "SELECT name FROM users WHERE id = ?" "123"
```

#### ExecDB
```bash
ExecDB "sql" "arg1" "arg2" ...
```
执行非查询 SQL（INSERT, UPDATE, DELETE）。

**参数:**
- `sql` - SQL 语句
- `arg1`, `arg2`, ... - 参数（可选）

**返回值:**
- 执行结果
- 错误信息（如果失败）

**示例:**
```bash
ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" "Alice" "alice@example.com"
```

#### GetQueryResult
```bash
GetQueryResult
```
获取上一次查询的结果。

**返回值:**
- JSON 格式的查询结果

**示例:**
```bash
result := GetQueryResult
echo $result
```

---

### 缓存操作

#### SetCache
```bash
SetCache "key" "value" "ttl"
```
设置缓存。

**参数:**
- `key` - 缓存键
- `value` - 缓存值
- `ttl` - 过期时间（秒）

**返回值:**
- 无

**示例:**
```bash
SetCache "user:123" '{"name":"Alice"}' 3600
```

#### GetCache
```bash
GetCache "key"
```
获取缓存值。

**参数:**
- `key` - 缓存键

**返回值:**
- 缓存值
- `true` 如果存在，`false` 如果不存在

**示例:**
```bash
value, exists := GetCache "user:123"
if $exists; then
    echo $value
fi
```

#### DeleteCache
```bash
DeleteCache "key"
```
删除缓存。

**参数:**
- `key` - 缓存键

**返回值:**
- 无

**示例:**
```bash
DeleteCache "user:123"
```

#### CacheExists
```bash
CacheExists "key"
```
检查缓存是否存在。

**参数:**
- `key` - 缓存键

**返回值:**
- `true` 如果存在，`false` 否则

**示例:**
```bash
if CacheExists "user:123"; then
    echo "Cache hit"
fi
```

#### ClearCache
```bash
ClearCache
```
清空所有缓存。

**返回值:**
- 无

**示例:**
```bash
ClearCache
```

---

## HTTP 上下文

在 HTTP 处理函数中，可以使用以下上下文函数：

### GetHTTPHeader
```bash
GetHTTPHeader "name"
```
获取 HTTP 请求头。

**参数:**
- `name` - 头名称

**返回值:**
- 头值

**示例:**
```bash
auth := GetHTTPHeader "Authorization"
echo "Auth: $auth"
```

### SetHTTPHeader
```bash
SetHTTPHeader "name" "value"
```
设置 HTTP 响应头。

**参数:**
- `name` - 头名称
- `value` - 头值

**返回值:**
- 无

**示例:**
```bash
SetHTTPHeader "Content-Type" "application/json"
```

---

## 错误类型

Shode 使用自定义错误类型系统：

```go
ErrSecurityViolation   // 安全违规
ErrCommandNotFound     // 命令未找到
ErrExecutionFailed      // 执行失败
ErrParseError           // 解析错误
ErrTimeout              // 超时
ErrFileNotFound         // 文件未找到
ErrPermissionDenied     // 权限拒绝
ErrInvalidInput         // 无效输入
ErrResourceExhausted    // 资源耗尽
ErrNetworkError         // 网络错误
ErrUnknown              // 未知错误
```

---

## 完整示例

### REST API 服务器

```bash
#!/usr/bin/env shode

# 启动 HTTP 服务器
StartHTTPServer "8080"

# 获取用户列表
function GetUsers() {
    SetHTTPResponse 200 '{"users":[{"id":1,"name":"Alice"}]}'
}
RegisterHTTPRoute "GET" "/api/users" "function" "GetUsers"

# 获取单个用户
function GetUser() {
    id := GetHTTPQuery "id"
    result := QueryRowDB "SELECT * FROM users WHERE id = ?" $id
    SetHTTPResponse 200 $result
}
RegisterHTTPRoute "GET" "/api/user" "function" "GetUser"

# 创建用户
function CreateUser() {
    body := GetHTTPBody
    # 解析并插入数据库
    ExecDB "INSERT INTO users (name) VALUES (?)"
    SetHTTPResponse 201 '{"status":"created"}'
}
RegisterHTTPRoute "POST" "/api/users" "function" "CreateUser"

# 静态文件
RegisterStaticRoute "/" "./public"

# 保持运行
for i in $(seq 1 100000); do sleep 1; done
```

### WebSocket 聊天室

```bash
#!/usr/bin/env shode

StartHTTPServer "8090"

# 注册 WebSocket 路由
RegisterWebSocketRoute "/ws" ""

# 广播消息 API
function Broadcast() {
    body := GetHTTPBody
    BroadcastWebSocketMessage $body
    SetHTTPResponse 200 '{"status":"broadcasted"}'
}
RegisterHTTPRoute "POST" "/api/broadcast" "function" "Broadcast"

# 统计信息 API
function Stats() {
    count := GetWebSocketConnectionCount
    SetHTTPResponse 200 "{\"connections\":$count}"
}
RegisterHTTPRoute "GET" "/api/stats" "function" "Stats"

# 保持运行
for i in $(seq 1 100000); do sleep 1; done
```

---

## 更多信息

- [用户指南](USER_GUIDE.md)
- [WebSocket 指南](WEBSOCKET_GUIDE.md)
- [最佳实践](BEST_PRACTICES.md)
- [示例项目](../examples/)
