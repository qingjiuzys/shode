// Package main 提供实时聊天应用示例。
// 这是一个使用 WebSocket 的实时聊天应用。
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/logger"
	"gitee.com/com_818cloud/shode/pkg/websocket"
)

// Message 聊天消息
type Message struct {
	ID        string    `json:"id"`
	Room      string    `json:"room"`
	From      string    `json:"from"`
	To        string    `json:"to,omitempty"`
	Content   string    `json:"content"`
	Type      string    `json:"type"` // text, system, typing
	Timestamp time.Time `json:"timestamp"`
}

// User 聊天用户
type User struct {
	ID       string `json:"id"`
	Nickname string `json:"nickname"`
	Room     string `json:"room"`
}

// ChatRoom 聊天室
type ChatRoom struct {
	ID      string
	Name    string
	Users   map[string]*User
	Messages []*Message
	mu      sync.RWMutex
}

// ChatService 聊天服务
type ChatService struct {
	rooms       map[string]*ChatRoom
	userConnMap map[string]string // userID -> connectionID
	mu          sync.RWMutex
	manager     *websocket.Manager
	presence    *websocket.PresenceManager
}

// NewChatService 创建聊天服务
func NewChatService(manager *websocket.Manager) *ChatService {
	return &ChatService{
		rooms:       make(map[string]*ChatRoom),
		userConnMap: make(map[string]string),
		manager:     manager,
		presence:    websocket.NewPresenceManager(manager),
	}
}

// CreateRoom 创建聊天室
func (cs *ChatService) CreateRoom(id, name string) *ChatRoom {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	room := &ChatRoom{
		ID:       id,
		Name:     name,
		Users:    make(map[string]*User),
		Messages: make([]*Message, 0),
	}
	cs.rooms[id] = room
	return room
}

// GetRoom 获取聊天室
func (cs *ChatService) GetRoom(id string) (*ChatRoom, bool) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	room, exists := cs.rooms[id]
	return room, exists
}

// ListRooms 列出所有聊天室
func (cs *ChatService) ListRooms() []*ChatRoom {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	rooms := make([]*ChatRoom, 0, len(cs.rooms))
	for _, room := range cs.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// JoinRoom 加入聊天室
func (cs *ChatService) JoinRoom(roomID, userID, nickname, connID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	room, exists := cs.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found")
	}

	user := &User{
		ID:       userID,
		Nickname: nickname,
		Room:     roomID,
	}

	room.Users[userID] = user
	cs.userConnMap[userID] = connID

	// 加入 presence
	cs.presence.Join(userID, connID, map[string]interface{}{
		"nickname": nickname,
		"room":     roomID,
	})
	cs.presence.JoinRoom(userID, roomID)

	// 发送系统消息
	cs.broadcastSystemMessage(roomID, fmt.Sprintf("%s joined the room", nickname))

	return nil
}

// LeaveRoom 离开聊天室
func (cs *ChatService) LeaveRoom(roomID, userID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	room, exists := cs.rooms[roomID]
	if !exists {
		return fmt.Errorf("room not found")
	}

	user, exists := room.Users[userID]
	if !exists {
		return fmt.Errorf("user not in room")
	}

	delete(room.Users, userID)
	delete(cs.userConnMap, userID)

	// 离开 presence
	cs.presence.LeaveRoom(userID, roomID)

	// 发送系统消息
	cs.broadcastSystemMessage(roomID, fmt.Sprintf("%s left the room", user.Nickname))

	return nil
}

// SendMessage 发送消息
func (cs *ChatService) SendMessage(msg *Message) error {
	cs.mu.RLock()
	room, exists := cs.rooms[msg.Room]
	cs.mu.RUnlock()

	if !exists {
		return fmt.Errorf("room not found")
	}

	msg.ID = generateID()
	msg.Timestamp = time.Now()

	// 保存消息
	room.mu.Lock()
	room.Messages = append(room.Messages, msg)
	if len(room.Messages) > 100 { // 保留最近 100 条消息
		room.Messages = room.Messages[len(room.Messages)-100:]
	}
	room.mu.Unlock()

	// 广播消息
	cs.broadcastToRoom(msg.Room, msg)

	return nil
}

// GetMessages 获取聊天室消息
func (cs *ChatService) GetMessages(roomID string) ([]*Message, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	room, exists := cs.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	room.mu.RLock()
	messages := make([]*Message, len(room.Messages))
	copy(messages, room.Messages)
	room.mu.RUnlock()

	return messages, nil
}

// GetRoomUsers 获取聊天室用户
func (cs *ChatService) GetRoomUsers(roomID string) ([]*User, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	room, exists := cs.rooms[roomID]
	if !exists {
		return nil, fmt.Errorf("room not found")
	}

	users := make([]*User, 0, len(room.Users))
	for _, user := range room.Users {
		users = append(users, user)
	}

	return users, nil
}

// broadcastSystemMessage 广播系统消息
func (cs *ChatService) broadcastSystemMessage(roomID, content string) {
	msg := &Message{
		ID:        generateID(),
		Room:      roomID,
		Content:   content,
		Type:      "system",
		Timestamp: time.Now(),
	}
	cs.broadcastToRoom(roomID, msg)
}

// broadcastToRoom 广播到房间
func (cs *ChatService) broadcastToRoom(roomID string, msg interface{}) {
	data, err := json.Marshal(msg)
	if err != nil {
		return
	}
	cs.manager.BroadcastToRoom(roomID, websocket.TextMessage, data)
}

// ChatHandler 聊天处理器
type ChatHandler struct {
	service *ChatService
	manager *websocket.Manager
	logger  *logger.Logger
}

// NewChatHandler 创建聊天处理器
func NewChatHandler(service *ChatService, manager *websocket.Manager, log *logger.Logger) *ChatHandler {
	return &ChatHandler{
		service: service,
		manager: manager,
		logger:  log,
	}
}

// HandleWebSocket 处理 WebSocket 连接
func (ch *ChatHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 升级到 WebSocket
	conn, err := websocket.AcceptWebSocket(w, r)
	if err != nil {
		ch.logger.Error("Failed to accept WebSocket:", err)
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusBadRequest)
		return
	}

	// 获取参数
	roomID := r.URL.Query().Get("room")
	userID := r.URL.Query().Get("user")
	nickname := r.URL.Query().Get("nickname")

	if roomID == "" || userID == "" || nickname == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		conn.Close()
		return
	}

	// 创建连接
	connection := &websocket.Connection{
		ID:      websocket.GenerateConnectionID(),
		Room:    roomID,
		Conn:    conn,
		Request: r,
	}

	// 添加到 manager
	ch.manager.AddConnection(connection)

	// 加入聊天室
	if err := ch.service.JoinRoom(roomID, userID, nickname, connection.ID); err != nil {
		ch.logger.Error("Failed to join room:", err)
		conn.Close()
		return
	}

	ch.logger.Info(fmt.Sprintf("User %s (%s) joined room %s", userID, nickname, roomID))

	// 发送欢迎消息
	welcome := &Message{
		ID:        generateID(),
		Room:      roomID,
		Content:   fmt.Sprintf("Welcome %s!", nickname),
		Type:      "system",
		Timestamp: time.Now(),
	}
	data, _ := json.Marshal(welcome)
	connection.Send(websocket.TextMessage, data)

	// 启动消息循环
	go ch.handleMessages(connection, userID, roomID)
}

// handleMessages 处理消息
func (ch *ChatHandler) handleMessages(conn *websocket.Connection, userID, roomID string) {
	defer func() {
		// 离开聊天室
		ch.service.LeaveRoom(roomID, userID)
		ch.manager.RemoveConnection(conn.ID)
		conn.Close()
		ch.logger.Info(fmt.Sprintf("User %s disconnected", userID))
	}()

	for {
		messageType, data, err := conn.Conn.ReadMessage()
		if err != nil {
			return
		}

		if websocket.MessageType(messageType) != websocket.TextMessage {
			continue
		}

		// 解析消息
		var msg Message
		if err := json.Unmarshal(data, &msg); err != nil {
			ch.logger.Error("Failed to parse message:", err)
			continue
		}

		msg.Room = roomID
		msg.From = userID

		// 处理不同类型的消息
		switch msg.Type {
		case "text":
			if err := ch.service.SendMessage(&msg); err != nil {
				ch.logger.Error("Failed to send message:", err)
			}
		case "typing":
			// 转发输入状态
			ch.manager.BroadcastToRoom(roomID, websocket.TextMessage, data)
		default:
			ch.logger.Warn("Unknown message type:", msg.Type)
		}
	}
}

// HandleListRooms 处理列出房间
func (ch *ChatHandler) HandleListRooms(w http.ResponseWriter, r *http.Request) {
	rooms := ch.service.ListRooms()

	// 转换为 JSON 格式
	result := make([]map[string]interface{}, 0)
	for _, room := range rooms {
		room.mu.RLock()
		info := map[string]interface{}{
			"id":       room.ID,
			"name":     room.Name,
			"userCount": len(room.Users),
		}
		room.mu.RUnlock()
		result = append(result, info)
	}

	respondJSON(w, http.StatusOK, result)
}

// HandleGetMessages 处理获取消息
func (ch *ChatHandler) HandleGetMessages(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Missing room parameter", http.StatusBadRequest)
		return
	}

	messages, err := ch.service.GetMessages(roomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, messages)
}

// HandleGetUsers 处理获取用户
func (ch *ChatHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")
	if roomID == "" {
		http.Error(w, "Missing room parameter", http.StatusBadRequest)
		return
	}

	users, err := ch.service.GetRoomUsers(roomID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	respondJSON(w, http.StatusOK, users)
}

// respondJSON 响应 JSON
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func main() {
	// 初始化日志
	log := logger.NewLogger(logger.DefaultConfig)

	// 创建 WebSocket manager
	wsManager := websocket.NewManager()

	// 创建聊天服务
	chatService := NewChatService(wsManager)

	// 创建示例聊天室
	chatService.CreateRoom("general", "General Discussion")
	chatService.CreateRoom("random", "Random Chat")
	chatService.CreateRoom("tech", "Tech Talk")

	// 创建聊天处理器
	chatHandler := NewChatHandler(chatService, wsManager, log)

	// 创建路由
	mux := http.NewServeMux()

	// WebSocket 端点
	mux.HandleFunc("/ws", chatHandler.HandleWebSocket)

	// REST API 端点
	mux.HandleFunc("/api/rooms", chatHandler.HandleListRooms)
	mux.HandleFunc("/api/messages", chatHandler.HandleGetMessages)
	mux.HandleFunc("/api/users", chatHandler.HandleGetUsers)

	// 静态文件服务
	mux.Handle("/", http.FileServer(http.Dir("examples/chat-frontend")))

	log.Info("Starting Shode Chat Application on http://localhost:8080")
	log.Info("WebSocket endpoint: ws://localhost:8080/ws?room=<room_id>&user=<user_id>&nickname=<nickname>")
	log.Info("Available rooms: general, random, tech")

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Error("Failed to start server:", err)
	}
}
