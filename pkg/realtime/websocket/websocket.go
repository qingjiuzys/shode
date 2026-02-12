// Package websocket WebSocket 实时通信
package websocket

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message WebSocket 消息
type Message struct {
	Type    string      `json:"type"`
	Room    string      `json:"room,omitempty"`
	From    string      `json:"from,omitempty"`
	To      string      `json:"to,omitempty"`
	Data    interface{} `json:"data"`
	Time    time.Time   `json:"time"`
	ID      string      `json:"id,omitempty"`
}

// Client WebSocket 客户端
type Client struct {
	ID     string
	Hub    *Hub
	SendCh chan Message
	Conn   *websocket.Conn
	UserID string
	Rooms  map[string]bool
	mu     sync.RWMutex
}

// Hub WebSocket Hub
type Hub struct {
	clients    map[*Client]bool
	rooms      map[string]map[*Client]bool
	presence   map[string]string
	broadcast  chan Message
	register   chan *Client
	unregister chan *Client
	rpcMethods map[string]RPCHandler
	mutex      sync.RWMutex
}

// RPCHandler RPC 处理函数
type RPCHandler func(params map[string]interface{}) (interface{}, error)

// NewHub 创建 Hub
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		rooms:      make(map[string]map[*Client]bool),
		presence:   make(map[string]string),
		broadcast:  make(chan Message, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		rpcMethods: make(map[string]RPCHandler),
	}
}

// Run 运行 Hub
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.handleBroadcast(message)
		}
	}
}

// registerClient 注册客户端
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	h.presence[client.UserID] = Online
	fmt.Printf("Client connected: %s\n", client.ID)
}

// unregisterClient 注销客户端
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.SendCh)

		// 从所有房间移除
		for room := range client.Rooms {
			h.leaveRoom(room, client)
		}

		h.presence[client.UserID] = Offline
		fmt.Printf("Client disconnected: %s\n", client.ID)
	}
}

// handleBroadcast 处理广播消息
func (h *Hub) handleBroadcast(message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	// 广播到所有客户端
	for client := range h.clients {
		select {
		case client.SendCh <- message:
		default:
			h.unregister <- client
		}
	}
}

// Broadcast 广播消息
func (h *Hub) Broadcast(message Message) {
	h.broadcast <- message
}

// SendToRoom 发送消息到房间
func (h *Hub) SendToRoom(room string, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			select {
			case client.SendCh <- message:
			default:
				h.unregister <- client
			}
		}
	}
}

// SendToUser 发送消息到用户
func (h *Hub) SendToUser(userID string, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.SendCh <- message:
			default:
				h.unregister <- client
			}
			break
		}
	}
}

// Join 加入房间
func (h *Hub) Join(room string, client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}

	h.rooms[room][client] = true
	client.Rooms[room] = true

	// 通知房间其他人加入通知
	joinMsg := Message{
		Type: "user_joined",
		Room: room,
		Data: map[string]interface{}{
			"user_id": client.UserID,
		},
		Time: time.Now(),
	}

	for c := range h.rooms[room] {
		if c != client {
			select {
			case c.SendCh <- joinMsg:
			default:
			}
		}
	}

	fmt.Printf("User %s joined room %s\n", client.UserID, room)
}

// Leave 离开房间
func (h *Hub) Leave(room string, client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.leaveRoom(room, client)
}

func (h *Hub) leaveRoom(room string, client *Client) {
	if clients, ok := h.rooms[room]; ok {
		delete(clients, client)
		delete(client.Rooms, room)

		// 通知房间其他人离开
		leaveMsg := Message{
			Type: "user_left",
			Room: room,
			Data: map[string]interface{}{
				"user_id": client.UserID,
			},
			Time: time.Now(),
		}

		for c := range clients {
			select {
			case c.SendCh <- leaveMsg:
			default:
			}
		}

		// 如果房间为空，删除房间
		if len(clients) == 0 {
			delete(h.rooms, room)
		}
	}

	fmt.Printf("User %s left room %s\n", client.UserID, room)
}

// GetUsersInRoom 获取房间用户列表
func (h *Hub) GetUsersInRoom(room string) []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0)
	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			users = append(users, client.UserID)
		}
	}
	return users
}

// GetRoomInfo 获取房间信息
func (h *Hub) GetRoomInfo(room string) map[string]interface{} {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	info := make(map[string]interface{})

	if clients, ok := h.rooms[room]; ok {
		info["count"] = len(clients)
		users := make([]string, 0)
		for client := range clients {
			users = append(users, client.UserID)
		}
		info["users"] = users
	} else {
		info["count"] = 0
		info["users"] = []string{}
	}

	return info
}

// RegisterRPC 注册 RPC 方法
func (h *Hub) RegisterRPC(name string, handler RPCHandler) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.rpcMethods[name] = handler
}

// CallRPC 调用 RPC 方法
func (h *Hub) CallRPC(name string, params map[string]interface{}) (interface{}, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if handler, ok := h.rpcMethods[name]; ok {
		return handler(params)
	}
	return nil, fmt.Errorf("RPC method not found: %s", name)
}

// SetPresence 设置用户状态
func (h *Hub) SetPresence(userID, status string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.presence[userID] = status

	// 广播状态变化
	h.broadcast <- Message{
		Type: "presence_changed",
		Data: map[string]interface{}{
			"user_id": userID,
			"status":  status,
		},
		Time: time.Now(),
	}
}

// GetPresence 获取用户状态
func (h *Hub) GetPresence(userID string) string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if status, ok := h.presence[userID]; ok {
		return status
	}
	return Offline
}

// GetAllPresence 获取所有用户状态
func (h *Hub) GetAllPresence() map[string]string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	result := make(map[string]string)
	for k, v := range h.presence {
		result[k] = v
	}
	return result
}

// Upgrader WebSocket 升级器
var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ServeWebSocket 处理 WebSocket 连接
func ServeWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := Upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}

	// 获取用户 ID（简化实现）
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		userID = generateID()
	}

	client := &Client{
		ID:     generateID(),
		Hub:    hub,
		SendCh: make(chan Message, 256),
		Conn:   conn,
		UserID: userID,
		Rooms:  make(map[string]bool),
	}

	// 注册客户端
	hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// readPump 读取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				fmt.Printf("WebSocket error: %v\n", err)
			}
			break
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			fmt.Printf("Message parse error: %v\n", err)
			continue
		}

		msg.From = c.UserID
		msg.Time = time.Now()

		// 处理消息
		c.handleMessage(msg)
	}
}

// writePump 写入消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.SendCh:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			data, err := json.Marshal(message)
			if err != nil {
				fmt.Printf("Message marshal error: %v\n", err)
				continue
			}

			if err := c.Conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage 处理消息
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case "join_room":
		c.Hub.Join(msg.Room, c)
	case "leave_room":
		c.Hub.Leave(msg.Room, c)
	case "room_message":
		c.Hub.SendToRoom(msg.Room, msg)
	case "broadcast":
		c.Hub.Broadcast(msg)
	case "direct_message":
		c.Hub.SendToUser(msg.To, msg)
	case "rpc":
		c.handleRPC(msg)
	case "presence":
		c.Hub.SetPresence(c.UserID, msg.Data.(string))
	}
}

// handleRPC 处理 RPC 调用
func (c *Client) handleRPC(msg Message) {
	if data, ok := msg.Data.(map[string]interface{}); ok {
		methodName, _ := data["method"].(string)
		params, _ := data["params"].(map[string]interface{})

		result, err := c.Hub.CallRPC(methodName, params)

		response := Message{
			Type: "rpc_response",
			From: c.ID,
			Data: map[string]interface{}{
				"id":     msg.ID,
				"result": result,
				"error":  err,
			},
			Time: time.Now(),
		}

		select {
		case c.SendCh <- response:
		default:
		}
	}
}

// 在线状态常量
const (
	Online    = "online"
	Offline   = "offline"
	Away      = "away"
	Busy      = "busy"
	Invisible = "invisible"
)

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// GetOnlineUsers 获取在线用户
func (h *Hub) GetOnlineUsers() []string {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	users := make([]string, 0)
	for client := range h.clients {
		users = append(users, client.UserID)
	}
	return users
}

// GetClientCount 获取客户端数量
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.clients)
}

// GetRoomCount 获取房间数量
func (h *Hub) GetRoomCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	return len(h.rooms)
}

// BroadcastExcept 广播到除指定客户端外的所有人
func (h *Hub) BroadcastExcept(message Message, except *Client) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client != except {
			select {
			case client.SendCh <- message:
			default:
				h.unregister <- client
			}
		}
	}
}

// CreateRoom 创建房间
func (h *Hub) CreateRoom(room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.rooms[room]; !ok {
		h.rooms[room] = make(map[*Client]bool)
	}
}

// DeleteRoom 删除房间
func (h *Hub) DeleteRoom(room string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if clients, ok := h.rooms[room]; ok {
		// 移除所有客户端
		for client := range clients {
			delete(client.Rooms, room)
		}
		delete(h.rooms, room)
	}
}

// RoomExists 检查房间是否存在
func (h *Hub) RoomExists(room string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	_, exists := h.rooms[room]
	return exists
}

// IsUserInRoom 检查用户是否在房间中
func (h *Hub) IsUserInRoom(room, userID string) bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if clients, ok := h.rooms[room]; ok {
		for client := range clients {
			if client.UserID == userID {
				return true
			}
		}
	}
	return false
}

// GetClientRooms 获取客户端所在房间
func (c *Client) GetClientRooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	rooms := make([]string, 0, len(c.Rooms))
	for room := range c.Rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// DefaultHub 默认 Hub
var DefaultHub = NewHub()

// Start 启动默认 Hub
func Start() {
	go DefaultHub.Run()
}
