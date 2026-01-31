// Package websocket 提供 WebSocket 连接管理和心跳机制。
//
// WebSocket 管理器特点：
//   - 连接池管理
//   - 房间（Room）分组
//   - 消息广播
//   - 连接生命周期回调
//
// 心跳机制特点：
//   - 可配置的心跳间隔和超时
//   - 支持 WebSocket 协议层面的 ping/pong
//   - 支持应用层面的 ping/pong
//   - 自动检测并清理超时连接
//   - 详细的统计信息
//
// 使用示例：
//
//	manager := websocket.NewManager()
//	hm := websocket.NewHeartbeatManager(nil, manager)
//
//	hm.Start(func(connID string) {
//	    manager.RemoveConnection(connID)
//	})
//
//	// 处理收到的消息时更新心跳
//	hm.HandleMessage(connID, messageType, data)
//
//	defer hm.Stop()
package websocket

import (
	"log"
	"sync"
	"time"
)

// HeartbeatConfig 配置心跳参数
type HeartbeatConfig struct {
	Interval        time.Duration // 心跳间隔
	Timeout         time.Duration // 超时时间
	PingMessage     []byte        // Ping 消息内容（二进制）
	PongMessage     []byte        // Pong 消息内容（二进制）
	UseWebSocketPing bool         // 使用 WebSocket 协议层面的 ping/pong
}

// DefaultHeartbeatConfig 默认心跳配置
var DefaultHeartbeatConfig = &HeartbeatConfig{
	Interval:        30 * time.Second,
	Timeout:         60 * time.Second,
	PingMessage:     []byte("ping"),
	PongMessage:     []byte("pong"),
	UseWebSocketPing: true, // 默认使用 WebSocket 协议层面的 ping
}

// HeartbeatManager 心跳管理器
type HeartbeatManager struct {
	config        *HeartbeatConfig
	mu            sync.RWMutex
	lastPong      map[string]time.Time // 最后一次 pong 时间
	onPingTimeout func(connID string)    // 超时回调
	stopChan      chan struct{}
	manager       *Manager // WebSocket 管理器引用
}

// NewHeartbeatManager 创建心跳管理器
func NewHeartbeatManager(config *HeartbeatConfig, manager *Manager) *HeartbeatManager {
	if config == nil {
		config = DefaultHeartbeatConfig
	}

	return &HeartbeatManager{
		config:   config,
		lastPong: make(map[string]time.Time),
		stopChan: make(chan struct{}),
		manager:  manager,
	}
}

// RegisterConnection 注册连接到心跳管理
func (hm *HeartbeatManager) RegisterConnection(connID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	hm.lastPong[connID] = time.Now()
	log.Printf("[Heartbeat] Connection %s registered", connID)
}

// UnregisterConnection 从心跳管理移除连接
func (hm *HeartbeatManager) UnregisterConnection(connID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	delete(hm.lastPong, connID)
	log.Printf("[Heartbeat] Connection %s unregistered", connID)
}

// UpdatePong 更新连接的 pong 时间
func (hm *HeartbeatManager) UpdatePong(connID string) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	
	hm.lastPong[connID] = time.Now()
}

// Start 启动心跳检测
func (hm *HeartbeatManager) Start(onPingTimeout func(connID string)) {
	hm.onPingTimeout = onPingTimeout

	go hm.runHeartbeat()
	go hm.checkTimeout()

	log.Printf("[Heartbeat] Heartbeat manager started (interval: %v, timeout: %v)",
		hm.config.Interval, hm.config.Timeout)
}

// SetOnPingTimeout 设置超时回调函数
func (hm *HeartbeatManager) SetOnPingTimeout(handler func(connID string)) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.onPingTimeout = handler
}

// Stop 停止心跳检测
func (hm *HeartbeatManager) Stop() {
	close(hm.stopChan)
}

// runHeartbeat 定期发送 ping
func (hm *HeartbeatManager) runHeartbeat() {
	ticker := time.NewTicker(hm.config.Interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 获取所有活跃连接
			if hm.manager != nil {
				connections := hm.manager.GetConnections()
				activeCount := 0

				for _, conn := range connections {
					// 检查连接是否已关闭
					if conn.IsClosed() {
						hm.UnregisterConnection(conn.ID)
						continue
					}

					// 检查连接是否有有效的 WebSocket 连接对象
					if conn.Conn == nil {
						// 跳过没有实际连接的连接（可能是测试用的模拟连接）
						continue
					}

					// 发送 ping 消息
					var err error
					if hm.config.UseWebSocketPing {
						// 使用 WebSocket 协议层面的 ping ( opcode 0x9 )
						err = conn.Send(PingMessage, hm.config.PingMessage)
					} else {
						// 使用文本消息作为 ping
						err = conn.Send(TextMessage, hm.config.PingMessage)
					}

					if err != nil {
						log.Printf("[Heartbeat] Failed to send ping to %s: %v", conn.ID, err)
						// 发送失败，标记为可能已断开
						hm.UnregisterConnection(conn.ID)
					} else {
						activeCount++
					}
				}

				if activeCount > 0 {
					log.Printf("[Heartbeat] Sent ping to %d active connections", activeCount)
				}
			}
		case <-hm.stopChan:
			log.Printf("[Heartbeat] Heartbeat stopped")
			return
		}
	}
}

// checkTimeout 检查超时连接
func (hm *HeartbeatManager) checkTimeout() {
	ticker := time.NewTicker(hm.config.Interval)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			hm.mu.Lock()
			now := time.Now()
			for connID, lastPong := range hm.lastPong {
				if now.Sub(lastPong) > hm.config.Timeout {
					log.Printf("[Heartbeat] Connection %s timeout", connID)
					
					if hm.onPingTimeout != nil {
						go hm.onPingTimeout(connID)
					}
					
					delete(hm.lastPong, connID)
				}
			}
			hm.mu.Unlock()
		case <-hm.stopChan:
			return
		}
	}
}

// GetLastPong 获取最后一次 pong 时间
func (hm *HeartbeatManager) GetLastPong(connID string) time.Time {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	
	return hm.lastPong[connID]
}

// IsActive 检查连接是否活跃
func (hm *HeartbeatManager) IsActive(connID string) bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	lastPong, exists := hm.lastPong[connID]
	if !exists {
		return false
	}

	return time.Since(lastPong) < hm.config.Timeout
}

// HandlePong 处理接收到的 pong 消息
// 这个方法应该在收到 pong 消息时调用
func (hm *HeartbeatManager) HandlePong(connID string) {
	hm.UpdatePong(connID)
}

// HandleMessage 处理消息（自动检测是否为 pong）
// 返回 true 表示消息是 pong 并已处理
func (hm *HeartbeatManager) HandleMessage(connID string, messageType MessageType, data []byte) bool {
	// 检查是否为 WebSocket 协议层面的 pong
	if messageType == PongMessage {
		hm.UpdatePong(connID)
		return true
	}

	// 检查是否为应用层面的 pong（文本消息）
	if messageType == TextMessage && hm.config != nil {
		if string(data) == string(hm.config.PongMessage) {
			hm.UpdatePong(connID)
			return true
		}
	}

	return false
}

// GetStats 获取心跳统计信息
type HeartbeatStats struct {
	TotalConnections int
	ActiveConnections int
	Config          *HeartbeatConfig
}

func (hm *HeartbeatManager) GetStats() HeartbeatStats {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	now := time.Now()
	activeCount := 0
	for _, lastPong := range hm.lastPong {
		if now.Sub(lastPong) < hm.config.Timeout {
			activeCount++
		}
	}

	return HeartbeatStats{
		TotalConnections:  len(hm.lastPong),
		ActiveConnections: activeCount,
		Config:           hm.config,
	}
}
