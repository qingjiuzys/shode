// Package sse Server-Sent Events
package sse

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Event SSE 事件
type Event struct {
	ID    string
	Event string
	Data  interface{}
	Retry int
}

// Client SSE 客户端
type Client struct {
	ID     string
	Events chan Event
	SendCh chan Event
	Done   chan bool
	mu     sync.RWMutex
}

// SSE SSE 服务端
type SSE struct {
	clients   map[string]*Client
	channels  map[string]map[string]bool
	broadcast chan Event
	register  chan *Client
	unregister chan string
	mu        sync.RWMutex
}

// NewSSE 创建 SSE 服务
func NewSSE() *SSE {
	return &SSE{
		clients:   make(map[string]*Client),
		channels:  make(map[string]map[string]bool),
		broadcast: make(chan Event, 256),
		register:  make(chan *Client),
		unregister: make(chan string),
	}
}

// NewClient 创建客户端
func NewClient(w http.ResponseWriter, r *http.Request) *Client {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	client := &Client{
		ID:     generateID(),
		Events: make(chan Event, 256),
		SendCh: make(chan Event, 256),
		Done:   make(chan bool),
	}

	go func() {
		<-r.Context().Done()
		client.Done <- true
	}()

	return client
}

// SendEvent 发送事件
func (c *Client) SendEvent(event Event) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	select {
	case c.SendCh <- event:
		return nil
	default:
		return fmt.Errorf("client channel full")
	}
}

// Send 发送事件
func (c *Client) Send(event Event) error {
	return c.SendEvent(event)
}

// Close 关闭客户端
func (c *Client) Close() error {
	close(c.Events)
	close(c.SendCh)
	return nil
}

// Run 运行 SSE 服务
func (s *SSE) Run() {
	for {
		select {
		case client := <-s.register:
			s.registerClient(client)
		case id := <-s.unregister:
			s.unregisterClient(id)
		case event := <-s.broadcast:
			s.handleBroadcast(event)
		}
	}
}

// registerClient 注册客户端
func (s *SSE) registerClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client.ID] = client
}

// unregisterClient 注销客户端
func (s *SSE) unregisterClient(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if client, ok := s.clients[id]; ok {
		delete(s.clients, id)
		client.Close()
	}
}

// handleBroadcast 处理广播
func (s *SSE) handleBroadcast(event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, client := range s.clients {
		client.SendEvent(event)
	}
}

// Broadcast 广播事件
func (s *SSE) Broadcast(event Event) {
	s.broadcast <- event
}

// Subscribe 订阅频道
func (s *SSE) Subscribe(channel, clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.channels[channel]; !ok {
		s.channels[channel] = make(map[string]bool)
	}

	s.channels[channel][clientID] = true
}

// Unsubscribe 取消订阅
func (s *SSE) Unsubscribe(channel, clientID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if clients, ok := s.channels[channel]; ok {
		delete(clients, clientID)
	}
}

// Publish 发布到频道
func (s *SSE) Publish(channel string, event Event) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if clients, ok := s.channels[channel]; ok {
		for clientID := range clients {
			if client, ok := s.clients[clientID]; ok {
				client.SendEvent(event)
			}
		}
	}
}

// GetClientCount 获取客户端数量
func (s *SSE) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Default 默认 SSE 服务
var Default = NewSSE()

// Start 启动默认服务
func Start() {
	go Default.Run()
}
