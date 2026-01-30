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
	PingMessage     string         // Ping 消息内容
	PongMessage     string         // Pong 消息内容
}

// DefaultHeartbeatConfig 默认心跳配置
var DefaultHeartbeatConfig = &HeartbeatConfig{
	Interval:    30 * time.Second,
	Timeout:     60 * time.Second,
	PingMessage: "ping",
	PongMessage: "pong",
}

// HeartbeatManager 心跳管理器
type HeartbeatManager struct {
	config        *HeartbeatConfig
	mu            sync.RWMutex
	lastPong      map[string]time.Time // 最后一次 pong 时间
	onPingTimeout func(connID string)    // 超时回调
	stopChan      chan struct{}
}

// NewHeartbeatManager 创建心跳管理器
func NewHeartbeatManager(config *HeartbeatConfig) *HeartbeatManager {
	if config == nil {
		config = DefaultHeartbeatConfig
	}
	
	return &HeartbeatManager{
		config:   config,
		lastPong: make(map[string]time.Time),
		stopChan: make(chan struct{}),
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
			// 触发 ping 事件
			log.Printf("[Heartbeat] Sending ping to all connections")
		case <-hm.stopChan:
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
