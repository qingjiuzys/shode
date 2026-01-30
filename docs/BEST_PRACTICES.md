# Shode 最佳实践

## 目录

- [安全性](#安全性)
- [性能优化](#性能优化)
- [错误处理](#错误处理)
- [代码组织](#代码组织)
- [HTTP 服务器](#http-服务器)
- [WebSocket](#websocket)
- [数据库操作](#数据库操作)
- [测试](#测试)

---

## 安全性

### 1. 输入验证

始终验证用户输入：

```bash
# ✅ 正确 - 验证路径
function HandleDownload() {
    path := GetHTTPQuery "path"
    
    # 检查路径遍历
    if Contains $path ".."; then
        SetHTTPResponse 400 '{"error":"Invalid path"}'
        return
    fi
    
    # 检查路径长度
    if ${#path} > 255; then
        SetHTTPResponse 400 '{"error":"Path too long"}'
        return
    fi
    
    # 处理请求...
}
```

### 2. SQL 注入防护

使用参数化查询：

```bash
# ✅ 正确 - 参数化查询
function GetUser() {
    id := GetHTTPQuery "id"
    result := QueryRowDB "SELECT * FROM users WHERE id = ?" $id
    SetHTTPResponse 200 $result
}

# ❌ 错误 - 字符串拼接
function GetUser_Bad() {
    id := GetHTTPQuery "id"
    sql := "SELECT * FROM users WHERE id = " $id
    result := QueryRowDB $sql  # SQL 注入风险！
}
```

### 3. 敏感信息保护

不要在日志中记录敏感信息：

```bash
# ✅ 正确
function Login() {
    username := GetHTTPQuery "username"
    password := GetHTTPQuery "password"
    
    # 只记录用户名
    echo "Login attempt: $username"
    
    # ❌ 不要这样做：
    # echo "Login attempt: $username $password"
}
```

### 4. 文件系统安全

限制文件访问范围：

```bash
# ✅ 正确 - 使用白名单
function SafeReadFile() {
    path := GetHTTPQuery "path"
    
    # 验证路径在允许的目录内
    if ! MatchPrefix $path "/var/data/"; then
        SetHTTPResponse 403 '{"error":"Access denied"}'
        return
    fi
    
    content := ReadFile $path
    SetHTTPResponse 200 $content
}
```

### 5. 认证和授权

实现认证中间件：

```bash
# 认证函数
function CheckAuth() {
    token := GetHTTPHeader "Authorization"
    
    # 验证 token
    if ! IsValidToken $token; then
        SetHTTPResponse 401 '{"error":"Unauthorized"}'
        return 1
    fi
    
    return 0
}

# 受保护的 API
function ProtectedAPI() {
    # 先检查认证
    if CheckAuth; then
        return
    fi
    
    # 处理请求...
    SetHTTPResponse 200 '{"data":"sensitive"}'
}
```

---

## 性能优化

### 1. 缓存策略

使用缓存减少重复计算：

```bash
# ✅ 正确 - 使用缓存
function GetUserProfile() {
    userID := GetHTTPQuery "user_id"
    cacheKey := "user:profile:" $userID
    
    # 先检查缓存
    cached, exists := GetCache $cacheKey
    if $exists; then
        SetHTTPResponse 200 $cached
        return
    fi
    
    # 缓存未命中，查询数据库
    profile := QueryRowDB "SELECT * FROM profiles WHERE user_id = ?" $userID
    
    # 存入缓存（1 小时）
    SetCache $cacheKey $profile 3600
    
    SetHTTPResponse 200 $profile
}
```

### 2. 批量操作

批量处理数据：

```bash
# ✅ 正确 - 批量插入
function ImportUsers() {
    users := GetHTTPBody
    
    # 批量插入而非逐条插入
    for user in $users; do
        ExecDB "INSERT INTO users (name, email) VALUES (?, ?)" $user.name $user.email
    done
    
    SetHTTPResponse 201 '{"status":"imported"}'
}
```

### 3. 连接池管理

保持数据库连接：

```bash
# 启动时连接
ConnectDB "sqlite" "./data.db"

# 使用已有连接
function ProcessRequest() {
    result := QueryDB "SELECT * FROM data"
    # ...
}

# 程序结束时关闭
# （Shode 会自动处理）
```

### 4. 流式处理

大文件分块处理：

```bash
# ✅ 正确 - 分块读取
function ProcessLargeFile() {
    path := GetHTTPQuery "file"
    
    # 分块读取而非一次性读取
    lines := ListFiles $path
    for line in $lines; do
        # 处理每一行
        ProcessLine $line
    done
}
```

### 5. 避免不必要的操作

减少循环中的重复计算：

```bash
# ✅ 正确 - 提前计算
function ProcessData() {
    data := GetData
    count := Length $data
    
    # 提前计算 count，避免每次循环都计算
    for i in $(seq 1 $count); do
        ProcessItem $i
    done
}

# ❌ 错误 - 每次循环都计算长度
function ProcessData_Bad() {
    data := GetData
    for i in $(seq 1 $(Length $data)); do
        ProcessItem $i
    done
}
```

---

## 错误处理

### 1. 始终检查错误

```bash
# ✅ 正确 - 检查错误
function SafeOperation() {
    result, err := SomeOperation
    if $err; then
        SetHTTPResponse 500 '{"error":"Operation failed"}'
        return
    fi
    
    SetHTTPResponse 200 $result
}

# ❌ 错误 - 忽略错误
function UnsafeOperation() {
    result := SomeOperation
    SetHTTPResponse 200 $result  # 可能使用失败的结果
}
```

### 2. 提供有用的错误信息

```bash
# ✅ 正确 - 详细的错误信息
function GetUser() {
    userID := GetHTTPQuery "user_id"
    
    # 验证输入
    if IsEmpty $userID; then
        SetHTTPResponse 400 '{"error":"user_id is required"}'
        return
    fi
    
    user, err := QueryRowDB "SELECT * FROM users WHERE id = ?" $userID
    if $err; then
        SetHTTPResponse 404 '{"error":"User not found"}'
        return
    fi
    
    SetHTTPResponse 200 $user
}
```

### 3. 使用统一的错误格式

```bash
# 标准错误响应格式
function ErrorResponse(status, message) {
    errorJSON := '{"error":"' $message '"}'
    SetHTTPResponse $status $errorJSON
}

# 使用示例
function HandleRequest() {
    if !ValidateInput; then
        ErrorResponse 400 "Invalid input"
        return
    fi
    
    if !Authenticate; then
        ErrorResponse 401 "Unauthorized"
        return
    fi
}
```

---

## 代码组织

### 1. 函数职责单一

每个函数只做一件事：

```bash
# ✅ 正确 - 单一职责
function ValidateUsername(username) {
    if ${#username} < 3; then
        return 1
    fi
    if ${#username} > 20; then
        return 1
    fi
    return 0
}

function SaveUser(user) {
    ExecDB "INSERT INTO users (name) VALUES (?)" $user
}

function CreateUser() {
    username := GetHTTPQuery "username"
    
    if ValidateUsername $username; then
        SetHTTPResponse 400 '{"error":"Invalid username"}'
        return
    fi
    
    SaveUser $username
    SetHTTPResponse 201 '{"status":"created"}'
}

# ❌ 错误 - 一个函数做太多事
function CreateUser_Bad() {
    username := GetHTTPQuery "username"
    
    # 验证
    if ${#username} < 3; then
        SetHTTPResponse 400 '{"error":"Invalid username"}'
        return
    fi
    
    # 保存
    ExecDB "INSERT INTO users (name) VALUES (?)" $username
    
    # 发送通知
    SendNotification $username
    
    # 更新缓存
    SetCache "user:$username" "..." 3600
    
    SetHTTPResponse 201 '{"status":"created"}'
}
```

### 2. 使用注释解释复杂逻辑

```bash
# ✅ 正确 - 清晰的注释
# 使用指数退避算法重试失败的网络请求
# 最大重试次数：5 次
# 初始延迟：1 秒
function FetchWithRetry(url) {
    maxRetries := 5
    delay := 1
    
    for i in $(seq 1 $maxRetries); do
        result, err := HTTPGet $url
        if !$err; then
            return $result
        fi
        
        # 指数退避
        sleep $delay
        delay := $delay * 2
    done
    
    return ""
}
```

### 3. 配置和代码分离

```bash
# ✅ 正确 - 使用配置变量
DB_HOST := "localhost"
DB_PORT := "5432"
DB_NAME := "myapp"

function Connect() {
    dsn := "host=" $DB_HOST " port=" $DB_PORT " dbname=" $DB_NAME
    ConnectDB "postgresql" $dsn
}

# ❌ 错误 - 硬编码配置
function Connect_Bad() {
    ConnectDB "postgresql" "host=localhost port=5432 dbname=myapp"
}
```

---

## HTTP 服务器

### 1. 路由组织

按功能组织路由：

```bash
# 用户相关路由
RegisterHTTPRoute "GET" "/api/users" "function" "ListUsers"
RegisterHTTPRoute "GET" "/api/users/:id" "function" "GetUser"
RegisterHTTPRoute "POST" "/api/users" "function" "CreateUser"
RegisterHTTPRoute "PUT" "/api/users/:id" "function" "UpdateUser"
RegisterHTTPRoute "DELETE" "/api/users/:id" "function" "DeleteUser"

# 文章相关路由
RegisterHTTPRoute "GET" "/api/articles" "function" "ListArticles"
RegisterHTTPRoute "GET" "/api/articles/:id" "function" "GetArticle"
```

### 2. 中间件模式

实现中间件链：

```bash
# 日志中间件
function LogMiddleware() {
    path := GetHTTPPath
    method := GetHTTPMethod
    echo "[$method] $path" >> /var/log/shode/access.log
}

# 认证中间件
function AuthMiddleware() {
    token := GetHTTPHeader "Authorization"
    if !IsValidToken $token; then
        SetHTTPResponse 401 '{"error":"Unauthorized"}'
        return 1
    fi
    return 0
}

# 使用中间件
function ProtectedAPI() {
    # 先执行中间件
    LogMiddleware
    if AuthMiddleware; then
        return
    fi
    
    # 处理实际请求
    SetHTTPResponse 200 '{"data":"protected"}'
}
```

### 3. 静态文件优化

配置静态文件服务：

```bash
# 启用 Gzip 和缓存
RegisterStaticRouteAdvanced "/" "./public" "" "false" "public, max-age=3600" "true" ""

# SPA 应用支持
RegisterStaticRouteAdvanced "/app" "./spa-build" "" "false" "" "false" "index.html"

# 文档目录（启用浏览）
RegisterStaticRouteAdvanced "/docs" "./documentation" "" "true" "max-age=7200" "false" ""
```

---

## WebSocket

### 1. 连接管理

正确管理连接生命周期：

```bash
# 连接建立时
function OnConnect() {
    connID := GetWebSocketConnectionID
    echo "Client connected: $connID"
    
    # 加入默认房间
    JoinRoom $connID "lobby"
}

# 连接断开时
function OnDisconnect() {
    connID := GetWebSocketConnectionID
    echo "Client disconnected: $connID"
}
```

### 2. 消息验证

验证收到的消息：

```bash
# ✅ 正确 - 验证消息
function HandleMessage() {
    message := GetWebSocketMessage
    
    # 验证消息长度
    if ${#message} > 10000; then
        SendWebSocketMessage $connID "Message too long"
        return
    fi
    
    # 验证消息格式
    if !IsValidJSON $message; then
        SendWebSocketMessage $connID "Invalid message format"
        return
    fi
    
    # 处理消息
    ProcessMessage $message
}
```

### 3. 房间管理

使用房间隔离用户：

```bash
# 加入房间
function JoinChatRoom() {
    connID := GetHTTPQuery "conn_id"
    room := GetHTTPQuery "room"
    
    # 离开旧房间
    LeaveRoom $connID
    
    # 加入新房间
    JoinRoom $connID $room
    
    # 通知房间成员
    BroadcastWebSocketMessageToRoom $room "User joined"
}

# 房间消息
function SendToRoom() {
    room := GetHTTPQuery "room"
    message := GetHTTPBody
    
    # 广播到房间
    BroadcastWebSocketMessageToRoom $room $message
}
```

---

## 数据库操作

### 1. 使用事务

确保数据一致性：

```bash
function TransferMoney() {
    fromAccount := GetHTTPQuery "from"
    toAccount := GetHTTPQuery "to"
    amount := GetHTTPQuery "amount"
    
    # 开始事务
    BeginTransaction
    
    # 扣除发送方余额
    ExecDB "UPDATE accounts SET balance = balance - ? WHERE id = ?" $amount $fromAccount
    
    # 增加接收方余额
    ExecDB "UPDATE accounts SET balance = balance + ? WHERE id = ?" $amount $toAccount
    
    # 提交事务
    CommitTransaction
    
    SetHTTPResponse 200 '{"status":"transferred"}'
}
```

### 2. 索引优化

为常用查询创建索引：

```sql
-- 用户表索引
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at);

-- 查询优化
SELECT * FROM users WHERE email = 'user@example.com';  -- 使用索引
```

### 3. 连接复用

保持连接而非每次创建：

```bash
# 启动时连接
function Init() {
    ConnectDB "sqlite" "./data.db"
    echo "Database connected"
}

Init

# 在整个生命周期中复用连接
function HandleRequest() {
    result := QueryDB "SELECT * FROM data"
    # 使用已建立的连接
}
```

---

## 测试

### 1. 单元测试

为关键函数编写测试：

```bash
# test_utils.sh
source ../shode

# 测试字符串操作
function test_to_upper() {
    result := ToUpper "hello"
    if [ "$result" != "HELLO" ]; then
        echo "FAIL: ToUpper"
        return 1
    fi
    echo "PASS: ToUpper"
}

# 运行测试
test_to_upper
```

### 2. 集成测试

测试完整流程：

```bash
# test_api.sh
#!/usr/bin/env shode

# 启动测试服务器
StartHTTPServer "9999"

# 注册测试路由
function TestEndpoint() {
    SetHTTPResponse 200 '{"status":"ok"}'
}
RegisterHTTPRoute "GET" "/test" "function" "TestEndpoint"

# 运行测试
response := curl http://localhost:9999/test
if Contains $response "ok"; then
    echo "PASS: API test"
else
    echo "FAIL: API test"
fi
```

### 3. 性能测试

使用基准测试：

```bash
# benchmark.sh
#!/usr/bin/env shode

# 测试数据库查询性能
start := Time
for i in $(seq 1 1000); do
    QueryDB "SELECT * FROM users LIMIT 1"
done
end := Time

elapsed := $end - $start
echo "1000 queries in $elapsed seconds"
```

---

## 总结

遵循这些最佳实践可以：

- ✅ 提高代码质量和可维护性
- ✅ 增强安全性
- ✅ 优化性能
- ✅ 减少错误
- ✅ 改善用户体验

记住：**好的代码不仅是能运行的代码，更是易读、易维护、安全的代码。**

---

## 参考资源

- [API 参考](API_REFERENCE.md)
- [编码规范](CODING_STANDARDS.md)
- [WebSocket 指南](WEBSOCKET_GUIDE.md)
- [示例项目](../examples/)
