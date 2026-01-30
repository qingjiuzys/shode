# Shode WebSocket 使用指南

## 目录

- [快速开始](#快速开始)
- [核心概念](#核心概念)
- [API 参考](#api-参考)
- [最佳实践](#最佳实践)
- [示例项目](#示例项目)

---

## 快速开始

### 最简单的 WebSocket 服务器

```bash
#!/usr/bin/env shode

# 启动 HTTP 服务器
StartHTTPServer "8096"

# 注册 WebSocket 路由
RegisterWebSocketRoute "/ws" ""

# 保持运行
for i in $(seq 1 100000); do sleep 1; done
```

**运行：**
```bash
shode run websocket-chat.sh
```

**访问：**
- WebSocket: `ws://localhost:8096/ws`

---

## 核心概念

### 1. 连接管理

每个 WebSocket 连接都有唯一 ID：

```bash
# 连接自动分配 ID
# 格式: conn_<timestamp>_<counter>
# 示例: conn_1706612345_1
```

### 2. 消息类型

支持文本消息：

```bash
# 发送文本消息
SendWebSocketMessage "conn_id" "Hello, World!"
```

### 3. 广播机制

三种广播方式：

```bash
# 1. 全局广播 - 所有连接
BroadcastWebSocketMessage "Server announcement"

# 2. 房间广播 - 特定房间
BroadcastWebSocketMessageToRoom "chatroom" "Hello room!"

# 3. 单播 - 特定连接
SendWebSocketMessage "conn_id" "Private message"
```

---

## API 参考

### 服务器管理

#### RegisterWebSocketRoute
```bash
RegisterWebSocketRoute "path" "handler"
```

注册 WebSocket 路由。

**参数:**
- `path` - WebSocket 端点路径（例如：`/ws`, `/chat`）
- `handler` - 处理函数名称（可选，留空则使用默认处理）

**返回值:**
- 错误信息（如果失败）

**示例:**
```bash
RegisterWebSocketRoute "/ws" ""
RegisterWebSocketRoute "/chat" "handleChat"
```

---

### 消息发送

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
BroadcastWebSocketMessageToRoom "general" "New message!"
```

---

### 房间管理

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

**说明:**
- 连接会自动离开旧房间
- 房间为空时自动删除

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

---

### 状态查询

#### GetWebSocketConnectionCount
```bash
GetWebSocketConnectionCount
```

获取当前连接总数。

**返回值:**
- 连接数（整数）

**示例:**
```bash
count := GetWebSocketConnectionCount
echo "Total connections: $count"
```

#### GetWebSocketRoomCount
```bash
GetWebSocketRoomCount "room"
```

获取特定房间的连接数。

**参数:**
- `room` - 房间名

**返回值:**
- 连接数（整数）

**示例:**
```bash
count := GetWebSocketRoomCount "general"
echo "Room connections: $count"
```

#### ListWebSocketRooms
```bash
ListWebSocketRooms
```

列出所有活跃房间。

**返回值:**
- 房间列表（JSON 数组）

**示例:**
```bash
rooms := ListWebSocketRooms
echo "Active rooms: $rooms"
```

---

## 最佳实践

### 1. 房间隔离

使用房间实现用户隔离：

```bash
function OnConnect() {
    # 每个用户加入自己的房间
    userID := GetHTTPQuery "user_id"
    userRoom := "user:" $userID
    JoinRoom $conn_id $userRoom
}

function SendToUser() {
    userID := GetHTTPQuery "user_id"
    userRoom := "user:" $userID
    BroadcastWebSocketMessageToRoom $userRoom "Private message"
}
```

### 2. 消息验证

验证消息格式和长度：

```bash
function HandleMessage() {
    message := GetWebSocketMessage
    
    # 检查长度
    if ${#message} > 10000; then
        SendWebSocketMessage $conn_id "Message too long"
        return
    fi
    
    # 检查格式
    if !IsValidJSON $message; then
        SendWebSocketMessage $conn_id "Invalid format"
        return
    fi
    
    # 处理消息
    ProcessMessage $message
}
```

### 3. 优雅关闭

处理连接关闭：

```bash
function OnDisconnect() {
    # 离开房间
    LeaveRoom $conn_id
    
    # 清理资源
    DeleteCache "session:" $conn_id
    
    # 记录日志
    echo "Client disconnected: $conn_id"
}
```

### 4. 错误处理

处理发送失败：

```bash
function BroadcastSafely() {
    rooms := ListWebSocketRooms
    
    for room in $rooms; do
        err := BroadcastWebSocketMessageToRoom $room "Message"
        if $err; then
            echo "Failed to broadcast to room: $room"
        fi
    done
}
```

---

## 完整示例

### 实时聊天室

```bash
#!/usr/bin/env shode

StartHTTPServer "8096"

# 注册 WebSocket
RegisterWebSocketRoute "/ws" ""

# 广播 API
function Broadcast() {
    body := GetHTTPBody
    BroadcastWebSocketMessage $body
    SetHTTPResponse 200 '{"status":"sent"}'
}
RegisterHTTPRoute "POST" "/api/broadcast" "function" "Broadcast"

# 统计 API
function Stats() {
    count := GetWebSocketConnectionCount
    SetHTTPResponse 200 '{"connections":' $count '}'
}
RegisterHTTPRoute "GET" "/api/stats" "function" "Stats"

# 房间广播
function BroadcastRoom() {
    room := GetHTTPQuery "room"
    body := GetHTTPBody
    BroadcastWebSocketMessageToRoom $room $body
    SetHTTPResponse 200 '{"status":"sent"}'
}
RegisterHTTPRoute "POST" "/api/broadcast-room" "function" "BroadcastRoom"

# 保持运行
for i in $(seq 1 100000); do sleep 1; done
```

### 实时通知推送

```bash
#!/usr/bin/env shode

StartHTTPServer "8097"

RegisterWebSocketRoute "/notify" ""

# 触发通知
function SendNotification() {
    title := GetHTTPQuery "title"
    message := GetHTTPQuery "message"
    
    notification := '{"title":"' $title '","message":"' $message '"}'
    BroadcastWebSocketMessage $notification
    
    SetHTTPResponse 200 '{"status":"notified"}'
}
RegisterHTTPRoute "POST" "/api/notify" "function" "SendNotification"

for i in $(seq 1 100000); do sleep 1; done
```

---

## 性能优化

### 1. 连接数管理

监控连接数：

```bash
function CheckConnections() {
    count := GetWebSocketConnectionCount
    
    if $count > 1000; then
        echo "Warning: Too many connections"
        # 触发告警
    fi
}
```

### 2. 房间清理

自动清理空房间（内置功能）：

```bash
# 当房间为空时自动删除
# 无需手动清理
```

### 3. 消息队列

对于高并发场景，考虑使用消息队列：

```bash
function QueueMessage() {
    # 将消息加入队列
    Enqueue "message_queue" $message
}

function ProcessQueue() {
    for i in $(seq 1 100000); do
        # 批量处理队列中的消息
        messages := DequeueBatch "message_queue" 100
        
        for msg in $messages; do
            BroadcastWebSocketMessage $msg
        done
        
        sleep 1
    done
}
```

---

## 安全性

### 1. 验证连接

在 WebSocket 握手时验证用户：

```bash
function OnConnect() {
    token := GetHTTPHeader "Authorization"
    
    # 验证 token
    if !IsValidToken $token; then
        # 拒绝连接
        return
    fi
    
    # 加入用户房间
    userID := ExtractUserID $token
    JoinRoom $conn_id "user:" $userID
}
```

### 2. 消息过滤

过滤恶意消息：

```bash
function FilterMessage() {
    message := GetWebSocketMessage
    
    # 检查敏感词
    if ContainsSensitiveWord $message; then
        return
    fi
    
    # 转义 HTML
    message := EscapeHTML $message
    
    BroadcastWebSocketMessage $message
}
```

---

## 故障排查

### 常见问题

#### 1. 连接立即断开

**原因**: 端口被占用或服务器未启动

**解决**:
```bash
# 检查端口
lsof -i :8096

# 确保服务器已启动
StartHTTPServer "8096"
```

#### 2. 消息发送失败

**原因**: 连接已关闭或 ID 错误

**解决**:
```bash
# 检查连接是否存在
count := GetWebSocketConnectionCount
echo "Active connections: $count"

# 验证连接 ID
conn_id := GetHTTPQuery "conn_id"
# 确保使用正确的 ID
```

#### 3. 房间广播无反应

**原因**: 房间名错误或连接未加入房间

**解决**:
```bash
# 列出所有房间
rooms := ListWebSocketRooms
echo "Active rooms: $rooms"

# 检查房间连接数
count := GetWebSocketRoomCount "chatroom"
echo "Room connections: $count"
```

---

## 更多资源

- [API 参考](API_REFERENCE.md)
- [最佳实践](BEST_PRACTICES.md)
- [示例项目](../examples/projects/)

---

**Happy Coding with Shode WebSocket!** 🚀
