package stdlib

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/net/websocket"
)

// WebSocketConnection represents an active WebSocket connection
type WebSocketConnection struct {
	ID         string
	Conn       *websocket.Conn
	RemoteAddr string
	UserAgent  string
	Room       string
	mu         sync.Mutex
	closed     bool
}

// WebSocketManager manages WebSocket connections
type WebSocketManager struct {
	connections map[string]*WebSocketConnection
	rooms       map[string][]*WebSocketConnection
	handlers    map[string]func(*WebSocketConnection, []byte)
	mu          sync.RWMutex
	handlerFunc string // Shode function name to call
}

// NewWebSocketManager creates a new WebSocket manager
func NewWebSocketManager() *WebSocketManager {
	return &WebSocketManager{
		connections: make(map[string]*WebSocketConnection),
		rooms:       make(map[string][]*WebSocketConnection),
		handlers:    make(map[string]func(*WebSocketConnection, []byte)),
	}
}

// RegisterWebSocketRoute registers a WebSocket route
func (sl *StdLib) RegisterWebSocketRoute(path, handlerFunc string) error {
	if sl.httpServer == nil {
		return fmt.Errorf("HTTP server not started")
	}

	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.wsManager == nil {
		sl.wsManager = NewWebSocketManager()
	}

	sl.wsManager.handlerFunc = handlerFunc

	// Create WebSocket handler
	wsHandler := &websocket.Server{
		Handler: func(wsConn *websocket.Conn) {
			// Create connection wrapper
			conn := &WebSocketConnection{
				ID:         fmt.Sprintf("ws_%d", time.Now().UnixNano()),
				Conn:       wsConn,
				RemoteAddr: wsConn.Request().RemoteAddr,
				UserAgent:  wsConn.Request().Header.Get("User-Agent"),
				Room:       "",
				closed:     false,
			}

			// Add connection
			sl.wsManager.mu.Lock()
			sl.wsManager.connections[conn.ID] = conn
			sl.wsManager.mu.Unlock()

			fmt.Printf("[WebSocket] Client connected: %s from %s\n", conn.ID, conn.RemoteAddr)

			// Clean up on disconnect
			defer func() {
				sl.wsManager.mu.Lock()
				delete(sl.wsManager.connections, conn.ID)
				if conn.Room != "" {
					// Remove from room
					if roomConns, ok := sl.wsManager.rooms[conn.Room]; ok {
						for i, c := range roomConns {
							if c.ID == conn.ID {
								sl.wsManager.rooms[conn.Room] = append(roomConns[:i], roomConns[i+1:]...)
								break
							}
						}
						if len(sl.wsManager.rooms[conn.Room]) == 0 {
							delete(sl.wsManager.rooms, conn.Room)
						}
					}
				}
				sl.wsManager.mu.Unlock()

				conn.mu.Lock()
				conn.closed = true
				conn.mu.Unlock()

				fmt.Printf("[WebSocket] Client disconnected: %s\n", conn.ID)
			}()

			// Message loop
			buf := make([]byte, 4096)
			for {
				n, err := wsConn.Read(buf)
				if err != nil {
					break
				}

				message := string(buf[:n])
				fmt.Printf("[WebSocket] Message from %s: %s\n", conn.ID, message)

				// Call handler function if registered
				if handlerFunc != "" {
					// In a real implementation, this would execute the Shode function
					// For now, just print the message
					fmt.Printf("[WebSocket] Would call handler: %s with message: %s\n", handlerFunc, message)
				}
			}
		},
	}

	// Register with ServeMux
	sl.httpServer.mux.Handle(path, wsHandler)

	return nil
}

// BroadcastWebSocketMessage broadcasts a message to all WebSocket connections
func (sl *StdLib) BroadcastWebSocketMessage(message string) error {
	if sl.wsManager == nil {
		return fmt.Errorf("WebSocket not initialized")
	}

	sl.wsManager.mu.RLock()
	defer sl.wsManager.mu.RUnlock()

	for _, conn := range sl.wsManager.connections {
		conn.mu.Lock()
		if !conn.closed {
			websocket.Message.Send(conn.Conn, message)
		}
		conn.mu.Unlock()
	}

	return nil
}

// BroadcastWebSocketMessageToRoom broadcasts a message to connections in a room
func (sl *StdLib) BroadcastWebSocketMessageToRoom(room, message string) error {
	if sl.wsManager == nil {
		return fmt.Errorf("WebSocket not initialized")
	}

	sl.wsManager.mu.RLock()
	defer sl.wsManager.mu.RUnlock()

	roomConns, exists := sl.wsManager.rooms[room]
	if !exists {
		return fmt.Errorf("room not found: %s", room)
	}

	for _, conn := range roomConns {
		conn.mu.Lock()
		if !conn.closed {
			websocket.Message.Send(conn.Conn, message)
		}
		conn.mu.Unlock()
	}

	return nil
}

// SendWebSocketMessage sends a message to a specific connection
func (sl *StdLib) SendWebSocketMessage(connID, message string) error {
	if sl.wsManager == nil {
		return fmt.Errorf("WebSocket not initialized")
	}

	sl.wsManager.mu.RLock()
	conn, exists := sl.wsManager.connections[connID]
	sl.wsManager.mu.RUnlock()

	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	conn.mu.Lock()
	defer conn.mu.Unlock()

	if conn.closed {
		return fmt.Errorf("connection is closed")
	}

	return websocket.Message.Send(conn.Conn, message)
}

// JoinRoom adds a connection to a room
func (sl *StdLib) JoinRoom(connID, room string) error {
	if sl.wsManager == nil {
		return fmt.Errorf("WebSocket not initialized")
	}

	sl.wsManager.mu.Lock()
	defer sl.wsManager.mu.Unlock()

	conn, exists := sl.wsManager.connections[connID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	// Remove from old room
	if conn.Room != "" {
		if roomConns, ok := sl.wsManager.rooms[conn.Room]; ok {
			for i, c := range roomConns {
				if c.ID == connID {
					sl.wsManager.rooms[conn.Room] = append(roomConns[:i], roomConns[i+1:]...)
					break
				}
			}
			if len(sl.wsManager.rooms[conn.Room]) == 0 {
				delete(sl.wsManager.rooms, conn.Room)
			}
		}
	}

	// Add to new room
	conn.Room = room
	sl.wsManager.rooms[room] = append(sl.wsManager.rooms[room], conn)

	return nil
}

// LeaveRoom removes a connection from its current room
func (sl *StdLib) LeaveRoom(connID string) error {
	if sl.wsManager == nil {
		return fmt.Errorf("WebSocket not initialized")
	}

	sl.wsManager.mu.Lock()
	defer sl.wsManager.mu.Unlock()

	conn, exists := sl.wsManager.connections[connID]
	if !exists {
		return fmt.Errorf("connection not found: %s", connID)
	}

	if conn.Room == "" {
		return fmt.Errorf("connection not in any room")
	}

	// Remove from room
	if roomConns, ok := sl.wsManager.rooms[conn.Room]; ok {
		for i, c := range roomConns {
			if c.ID == connID {
				sl.wsManager.rooms[conn.Room] = append(roomConns[:i], roomConns[i+1:]...)
				break
			}
		}
		if len(sl.wsManager.rooms[conn.Room]) == 0 {
			delete(sl.wsManager.rooms, conn.Room)
		}
	}

	conn.Room = ""
	return nil
}

// GetWebSocketConnectionCount returns the number of active WebSocket connections
func (sl *StdLib) GetWebSocketConnectionCount() int {
	if sl.wsManager == nil {
		return 0
	}

	sl.wsManager.mu.RLock()
	defer sl.wsManager.mu.RUnlock()
	return len(sl.wsManager.connections)
}

// GetWebSocketRoomCount returns the number of connections in a room
func (sl *StdLib) GetWebSocketRoomCount(room string) int {
	if sl.wsManager == nil {
		return 0
	}

	sl.wsManager.mu.RLock()
	defer sl.wsManager.mu.RUnlock()

	if roomConns, ok := sl.wsManager.rooms[room]; ok {
		return len(roomConns)
	}
	return 0
}

// ListWebSocketRooms returns all active room names
func (sl *StdLib) ListWebSocketRooms() []string {
	if sl.wsManager == nil {
		return []string{}
	}

	sl.wsManager.mu.RLock()
	defer sl.wsManager.mu.RUnlock()

	rooms := make([]string, 0, len(sl.wsManager.rooms))
	for room := range sl.wsManager.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}

