# 实时通信增强系统 (Realtime Communication System)

Shode 框架提供完整的实时通信解决方案。

## 🔌 功能特性

### 1. WebSocket 通信 (websocket/)
- ✅ 房间管理
- ✅ 消息广播
- ✅ Presence 状态管理
- ✅ 连接池管理
- ✅ 心跳检测
- ✅ 自动重连

### 2. Server-Sent Events (sse/)
- ✅ 单向推送
- ✅ 事件流
- ✅ 自动重连
- ✅ Last-Event-ID 支持

### 3. gRPC Streaming (grpc/)
- ✅ 双向流
- ✅ 服务端流
- ✅ 客户端流
- ✅ RPC 支持

### 4. WebRTC (webrtc/)
- ✅ P2P 连接
- ✅ 音视频通话
- ✅ 数据通道
- ✅ ICE/STUN/TURN

## 🚀 快速开始

### WebSocket 房间管理

```go
import "gitee.com/com_818cloud/shode/pkg/realtime/websocket"

func main() {
    // 创建 Hub
    hub := websocket.NewHub()

    // 启动 Hub
    go hub.Run()

    // WebSocket 处理器
    http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
        websocket.ServeWebSocket(hub, w, r)
    })
}
```

### 消息广播

```go
// 广播消息到所有客户端
hub.Broadcast(websocket.Message{
    Type: "chat",
    Data: "Hello everyone!",
})

// 发送到特定房间
hub.SendToRoom("room1", websocket.Message{
    Type: "notification",
    Data: "New message",
})
```

### Presence 状态管理

```go
// 用户上线
hub.Join("room1", client)

// 用户离线
hub.Leave("room1", client)

// 获取房间在线用户
users := hub.GetUsersInRoom("room1")
```

### SSE 推送

```go
import "gitee.com/com_818cloud/shode/pkg/realtime/sse"

func main() {
    http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
        // 创建 SSE 客户端
        client := sse.NewClient(w, r)

        // 发送事件
        client.Send(sse.Event{
            ID:    "1",
            Event: "message",
            Data:  "Hello!",
        })
    })
}
```

## 📡 WebSocket 房间管理

### Hub 架构

```go
type Hub struct {
    clients    map[*Client]bool
    rooms      map[string]map[*Client]bool
    broadcast  chan Message
    register   chan *Client
    unregister chan *Client
    mutex      sync.RWMutex
}
```

### 消息类型

```go
type Message struct {
    Type    string      `json:"type"`
    Room    string      `json:"room,omitempty"`
    From    string      `json:"from,omitempty"`
    To      string      `json:"to,omitempty"`
    Data    interface{} `json:"data"`
    Time    time.Time   `json:"time"`
}
```

### 房间操作

```go
// 创建房间
hub.CreateRoom("room1")

// 加入房间
hub.Join("room1", client)

// 离开房间
hub.Leave("room1", client")

// 发送到房间
hub.SendToRoom("room1", message)

// 获取房间信息
info := hub.GetRoomInfo("room1")
```

## 👥 Presence 状态管理

### 状态类型

```go
const (
    Online      = "online"
    Offline     = "offline"
    Away        = "away"
    Busy        = "busy"
    Invisible   = "invisible"
)
```

### Presence 操作

```go
// 更新状态
presence.SetStatus(userID, Online)

// 获取用户状态
status := presence.GetStatus(userID)

// 获取多个用户状态
statuses := presence.GetStatuses(userIDs)

// 监听状态变化
presence.Subscribe(userID, func(status string) {
    fmt.Println("User status changed:", status)
})
```

## 🔔 消息广播系统

### 广播策略

```go
// 广播到所有客户端
hub.Broadcast(message)

// 广播到房间
hub.SendToRoom("room1", message)

// 发送到特定用户
hub.SendToUser(userID, message)

// 除了发送者外的所有人
hub.BroadcastExcept(message, sender)
```

### 消息队列

```go
// 可靠消息传递
hub.Enqueue(message)

// 批量发送
hub.SendBatch(messages)

// 消息确认
message.Ack()
```

## 🔄 RPC 支持

### 远程过程调用

```go
// 注册 RPC 方法
hub.RegisterRPC("getUser", func(params map[string]interface{}) (interface{}, error) {
    userID := params["userId"].(string)
    return getUser(userID), nil
})

// 调用 RPC
result, err := hub.CallRPC("getUser", map[string]interface{}{
    "userId": "123",
})
```

## 📡 Server-Sent Events

### 创建 SSE 连接

```go
func handleSSE(w http.ResponseWriter, r *http.Request) {
    // 设置 SSE 头
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // 创建客户端
    client := sse.NewClient(w, r)

    // 发送事件
    for {
        select {
        case event := <-events:
            client.Send(event)
        case <-r.Context().Done():
            return
        }
    }
}
```

### 事件格式

```go
type Event struct {
    ID    string
    Event string
    Data  interface{}
    Retry int
}
```

## 🔌 gRPC Streaming

### 双向流

```go
func (s *server) StreamChat(stream pb.ChatService_StreamChatServer) error {
    for {
        req, err := stream.Recv()
        if err == io.EOF {
            break
        }

        // 处理消息
        resp := &pb.ChatResponse{
            Message: req.Message,
        }

        stream.Send(resp)
    }
    return nil
}
```

## 🎬 WebRTC

### PeerConnection

```go
// 创建 PeerConnection
pc, err := webrtc.NewPeerConnection(webrtc.Configuration{
    ICEServers: []webrtc.ICEServer{
        {URLs: []string{"stun:stun.l.google.com:19302"}},
    },
})

// 添加轨道
track, err := webrtc.NewTrackLocalStaticSample(
    webrtc.RTPCodecCapability{MimeType: "video/vp8"},
    "video",
    "pion",
)

pc.AddTrack(track)

// 创建 Offer
offer, err := pc.CreateOffer(nil)
pc.SetLocalDescription(offer)

// 设置 Answer
pc.SetRemoteDescription(answer)
```

## 🔧 配置选项

### WebSocket 配置

```go
type WebSocketConfig struct {
    ReadBufferSize    int
    WriteBufferSize   int
    PingPeriod        time.Duration
    PongTimeout       time.Duration
    MaxMessageSize    int64
    EnableCompression bool
}
```

### Presence 配置

```go
type PresenceConfig struct {
    HeartbeatInterval time.Duration
    TimeoutDuration   time.Duration
    CleanupInterval   time.Duration
}
```

## 📚 最佳实践

1. **心跳检测**: 定期发送心跳消息保持连接
2. **自动重连**: 客户端断线后自动重连
3. **消息确认**: 重要消息需要确认机制
4. **限流保护**: 防止消息洪泛
5. **状态同步**: 及时同步用户状态
6. **错误处理**: 优雅处理各种错误情况
7. **资源清理**: 及时清理断开的连接
8. **安全认证**: 验证连接和消息的合法性

## 🤝 贡献

欢迎贡献新的实时通信功能！

## 📄 许可证

MIT License
