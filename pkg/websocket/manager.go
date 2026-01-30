package websocket

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

// MessageType represents WebSocket message types
type MessageType int

const (
	// TextMessage represents a text message
	TextMessage MessageType = iota
	// BinaryMessage represents a binary message
	BinaryMessage
	// PingMessage represents a ping message
	PingMessage
	// PongMessage represents a pong message
	PongMessage
	// CloseMessage represents a close message
	CloseMessage
)

// Connection represents a WebSocket connection
type Connection struct {
	ID         string
	Room       string
	Conn       *websocket.Conn
	Request    *http.Request
	mu         sync.Mutex
	WriteChan  chan []byte
	CloseChan  chan bool
	RemoteAddr string
	UserAgent  string
}

// Manager manages WebSocket connections
type Manager struct {
	connections map[string]*Connection
	rooms       map[string][]*Connection
	mu          sync.RWMutex
	onMessage   func(*Connection, MessageType, []byte)
	onConnect   func(*Connection)
	onDisconnect func(*Connection)
}

// NewManager creates a new WebSocket manager
func NewManager() *Manager {
	return &Manager{
		connections: make(map[string]*Connection),
		rooms:       make(map[string][]*Connection),
	}
}

// SetMessageHandler sets the message handler callback
func (m *Manager) SetMessageHandler(handler func(*Connection, MessageType, []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onMessage = handler
}

// SetConnectHandler sets the connection handler callback
func (m *Manager) SetConnectHandler(handler func(*Connection)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onConnect = handler
}

// SetDisconnectHandler sets the disconnect handler callback
func (m *Manager) SetDisconnectHandler(handler func(*Connection)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.onDisconnect = handler
}

// AddConnection adds a new WebSocket connection
func (m *Manager) AddConnection(conn *Connection) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connections[conn.ID] = conn

	// Add to room if specified
	if conn.Room != "" {
		m.rooms[conn.Room] = append(m.rooms[conn.Room], conn)
	}

	// Call connect handler if set
	if m.onConnect != nil {
		go m.onConnect(conn)
	}
}

// RemoveConnection removes a WebSocket connection
func (m *Manager) RemoveConnection(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	conn, exists := m.connections[connID]
	if !exists {
		return
	}

	// Remove from room
	if conn.Room != "" {
		room := m.rooms[conn.Room]
		for i, c := range room {
			if c.ID == connID {
				m.rooms[conn.Room] = append(room[:i], room[i+1:]...)
				break
			}
		}
		// Clean up empty rooms
		if len(m.rooms[conn.Room]) == 0 {
			delete(m.rooms, conn.Room)
		}
	}

	delete(m.connections, connID)

	// Call disconnect handler if set
	if m.onDisconnect != nil {
		go m.onDisconnect(conn)
	}
}

// GetConnection retrieves a connection by ID
func (m *Manager) GetConnection(connID string) (*Connection, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[connID]
	return conn, exists
}

// GetConnections returns all connections
func (m *Manager) GetConnections() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		conns = append(conns, conn)
	}
	return conns
}

// GetRoomConnections returns all connections in a room
func (m *Manager) GetRoomConnections(room string) []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns, exists := m.rooms[room]
	if !exists {
		return []*Connection{}
	}

	result := make([]*Connection, len(conns))
	copy(result, conns)
	return result
}

// Broadcast sends a message to all connections
func (m *Manager) Broadcast(messageType MessageType, data []byte) {
	conns := m.GetConnections()
	for _, conn := range conns {
		conn.Send(messageType, data)
	}
}

// BroadcastToRoom sends a message to all connections in a room
func (m *Manager) BroadcastToRoom(room string, messageType MessageType, data []byte) {
	conns := m.GetRoomConnections(room)
	for _, conn := range conns {
		conn.Send(messageType, data)
	}
}

// BroadcastExcept sends a message to all connections except one
func (m *Manager) BroadcastExcept(excludeConnID string, messageType MessageType, data []byte) {
	conns := m.GetConnections()
	for _, conn := range conns {
		if conn.ID != excludeConnID {
			conn.Send(messageType, data)
		}
	}
}

// GetConnectionCount returns the number of active connections
func (m *Manager) GetConnectionCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connections)
}

// GetRoomCount returns the number of connections in a room
func (m *Manager) GetRoomCount(room string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.rooms[room])
}

// GetAllRooms returns all room names
func (m *Manager) GetAllRooms() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	rooms := make([]string, 0, len(m.rooms))
	for room := range m.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

// Send sends a message to the connection
func (c *Connection) Send(messageType MessageType, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.CloseChan:
		return fmt.Errorf("connection closed")
	default:
	}

	// gorilla/websocket expects int for message type
	var mt int
	switch messageType {
	case TextMessage:
		mt = websocket.TextMessage
	case BinaryMessage:
		mt = websocket.BinaryMessage
	case PingMessage:
		mt = websocket.PingMessage
	case PongMessage:
		mt = websocket.PongMessage
	case CloseMessage:
		mt = websocket.CloseMessage
	default:
		mt = websocket.TextMessage
	}

	return c.Conn.WriteMessage(mt, data)
}

// Close closes the connection
func (c *Connection) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case <-c.CloseChan:
		return
	default:
		close(c.CloseChan)
		c.Conn.Close()
	}
}

// IsClosed checks if the connection is closed
func (c *Connection) IsClosed() bool {
	select {
	case <-c.CloseChan:
		return true
	default:
		return false
	}
}

// GenerateConnectionID generates a unique connection ID
func GenerateConnectionID() string {
	return fmt.Sprintf("conn_%d", makeTimestamp())
}

func makeTimestamp() int64 {
	return 0 // Placeholder, will be replaced with actual timestamp
}

// buildFrame builds a WebSocket frame
func buildFrame(messageType MessageType, data []byte) ([]byte, error) {
	frame := make([]byte, 2+len(data))

	var opcode byte
	switch messageType {
	case TextMessage:
		opcode = 0x1
	case BinaryMessage:
		opcode = 0x2
	case PingMessage:
		opcode = 0x9
	case PongMessage:
		opcode = 0xA
	case CloseMessage:
		opcode = 0x8
	default:
		opcode = 0x1 // Default to text
	}

	// FIN bit + opcode
	frame[0] = 0x80 | opcode

	// Payload length
	length := len(data)
	if length < 126 {
		frame[1] = byte(length)
	} else if length < 65536 {
		frame[1] = 126
		binary.BigEndian.PutUint16(frame[2:4], uint16(length))
		frame = append(frame[:4], data...)
		return frame, nil
	} else {
		frame[1] = 127
		binary.BigEndian.PutUint64(frame[2:10], uint64(length))
		frame = append(frame[:10], data...)
		return frame, nil
	}

	copy(frame[2:], data)
	return frame, nil
}

// AcceptWebSocket accepts a WebSocket connection
func AcceptWebSocket(w http.ResponseWriter, r *http.Request) (*websocket.Conn, error) {
	// Validate WebSocket handshake
	if !isWebSocketUpgrade(r) {
		return nil, fmt.Errorf("not a WebSocket upgrade request")
	}

	// Create WebSocket upgrader
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for now
		},
	}

	// Upgrade the connection
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

// isWebSocketUpgrade checks if the request is a WebSocket upgrade
func isWebSocketUpgrade(r *http.Request) bool {
	return strings.ToLower(r.Header.Get("Upgrade")) == "websocket" &&
		strings.Contains(strings.ToLower(r.Header.Get("Connection")), "upgrade")
}

// GenerateSecWebSocketAccept generates Sec-WebSocket-Accept header value
func GenerateSecWebSocketAccept(secWebSocketKey string) string {
	const magicString = "258EAFA5-E914-47DA-95CA-C5AB0DC85B11"
	h := sha1.New()
	h.Write([]byte(secWebSocketKey + magicString))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}
