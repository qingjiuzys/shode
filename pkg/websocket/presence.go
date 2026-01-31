// Package websocket 提供在线状态管理功能。
package websocket

import (
	"encoding/json"
	"sync"
	"time"
)

// PresenceManager 管理用户在线状态
type PresenceManager struct {
	mu         sync.RWMutex
	presence   map[string]*PresenceInfo      // userID -> PresenceInfo
	rooms      map[string]map[string]bool    // roomID -> userID set
	userRooms  map[string]map[string]bool    // userID -> roomID set
	manager    *Manager
}

// PresenceInfo 用户在线状态信息
type PresenceInfo struct {
	UserID      string                 `json:"user_id"`
	ConnectionID string                `json:"connection_id"`
	Online      bool                   `json:"online"`
	LastSeen    time.Time              `json:"last_seen"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Rooms       []string               `json:"rooms"`
}

// PresenceEvent 在线状态事件
type PresenceEvent struct {
	Type    string                 `json:"type"` // join, leave, update
	UserID  string                 `json:"user_id"`
	Room    string                 `json:"room,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Time    time.Time              `json:"time"`
}

// NewPresenceManager 创建在线状态管理器
func NewPresenceManager(manager *Manager) *PresenceManager {
	return &PresenceManager{
		presence:  make(map[string]*PresenceInfo),
		rooms:     make(map[string]map[string]bool),
		userRooms: make(map[string]map[string]bool),
		manager:   manager,
	}
}

// Join 用户加入
func (pm *PresenceManager) Join(userID, connectionID string, metadata map[string]interface{}) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()

	// 更新用户状态
	pm.presence[userID] = &PresenceInfo{
		UserID:       userID,
		ConnectionID: connectionID,
		Online:       true,
		LastSeen:     now,
		Metadata:     metadata,
		Rooms:        make([]string, 0),
	}

	// 广播加入事件
	pm.broadcastEvent(PresenceEvent{
		Type:   "join",
		UserID: userID,
		Data:   metadata,
		Time:   now,
	})
}

// Leave 用户离开
func (pm *PresenceManager) Leave(userID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	info, exists := pm.presence[userID]
	if !exists {
		return
	}

	// 从所有房间移除
	for roomID := range pm.userRooms[userID] {
		pm.leaveRoom(userID, roomID)
	}

	// 更新状态
	info.Online = false
	info.LastSeen = time.Now()

	// 广播离开事件
	pm.broadcastEvent(PresenceEvent{
		Type:   "leave",
		UserID: userID,
		Time:   time.Now(),
	})
}

// UpdateHeartbeat 更新心跳时间
func (pm *PresenceManager) UpdateHeartbeat(userID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if info, exists := pm.presence[userID]; exists {
		info.LastSeen = time.Now()
		info.Online = true
	}
}

// UpdateMetadata 更新用户元数据
func (pm *PresenceManager) UpdateMetadata(userID string, metadata map[string]interface{}) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if info, exists := pm.presence[userID]; exists {
		info.Metadata = metadata
		info.LastSeen = time.Now()

		// 广播更新事件
		pm.broadcastEvent(PresenceEvent{
			Type:   "update",
			UserID: userID,
			Data:   metadata,
			Time:   time.Now(),
		})
	}
}

// JoinRoom 加入房间
func (pm *PresenceManager) JoinRoom(userID, roomID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// 初始化映射
	if pm.rooms[roomID] == nil {
		pm.rooms[roomID] = make(map[string]bool)
	}
	if pm.userRooms[userID] == nil {
		pm.userRooms[userID] = make(map[string]bool)
	}

	// 添加到房间
	pm.rooms[roomID][userID] = true
	pm.userRooms[userID][roomID] = true

	// 更新用户房间列表
	if info, exists := pm.presence[userID]; exists {
		info.Rooms = pm.getRoomList(userID)
	}

	// 广播房间加入事件
	pm.broadcastEventToRoom(roomID, PresenceEvent{
		Type:   "join",
		UserID: userID,
		Room:   roomID,
		Time:   time.Now(),
	})
}

// LeaveRoom 离开房间
func (pm *PresenceManager) LeaveRoom(userID, roomID string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pm.leaveRoom(userID, roomID)
}

// leaveRoom 内部离开房间方法（不加锁）
func (pm *PresenceManager) leaveRoom(userID, roomID string) {
	if pm.rooms[roomID] != nil {
		delete(pm.rooms[roomID], userID)
		if len(pm.rooms[roomID]) == 0 {
			delete(pm.rooms, roomID)
		}
	}

	if pm.userRooms[userID] != nil {
		delete(pm.userRooms[userID], roomID)
	}

	// 更新用户房间列表
	if info, exists := pm.presence[userID]; exists {
		info.Rooms = pm.getRoomList(userID)
	}

	// 广播房间离开事件
	pm.broadcastEventToRoom(roomID, PresenceEvent{
		Type:   "leave",
		UserID: userID,
		Room:   roomID,
		Time:   time.Now(),
	})
}

// GetPresence 获取用户状态
func (pm *PresenceManager) GetPresence(userID string) (*PresenceInfo, bool) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	info, exists := pm.presence[userID]
	if !exists {
		return nil, false
	}

	// 返回副本
	copy := *info
	return &copy, true
}

// GetRoomPresence 获取房间内所有在线用户
func (pm *PresenceManager) GetRoomPresence(roomID string) []*PresenceInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	users, exists := pm.rooms[roomID]
	if !exists {
		return []*PresenceInfo{}
	}

	result := make([]*PresenceInfo, 0, len(users))
	for userID := range users {
		if info, ok := pm.presence[userID]; ok && info.Online {
			copy := *info
			result = append(result, &copy)
		}
	}

	return result
}

// GetAllPresence 获取所有在线用户
func (pm *PresenceManager) GetAllPresence() []*PresenceInfo {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	result := make([]*PresenceInfo, 0)
	for _, info := range pm.presence {
		if info.Online {
			copy := *info
			result = append(result, &copy)
		}
	}

	return result
}

// GetOnlineCount 获取在线用户数
func (pm *PresenceManager) GetOnlineCount() int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	count := 0
	for _, info := range pm.presence {
		if info.Online {
			count++
		}
	}
	return count
}

// GetRoomCount 获取房间在线用户数
func (pm *PresenceManager) GetRoomCount(roomID string) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	users, exists := pm.rooms[roomID]
	if !exists {
		return 0
	}

	count := 0
	for userID := range users {
		if info, ok := pm.presence[userID]; ok && info.Online {
			count++
		}
	}
	return count
}

// GetAllRooms 获取所有房间
func (pm *PresenceManager) GetAllRooms() []string {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	rooms := make([]string, 0, len(pm.rooms))
	for roomID := range pm.rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}

// CleanupStale 清理过期用户
func (pm *PresenceManager) CleanupStale(timeout time.Duration) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	now := time.Now()
	for userID, info := range pm.presence {
		if info.Online && now.Sub(info.LastSeen) > timeout {
			info.Online = false

			// 广播离开事件
			pm.broadcastEvent(PresenceEvent{
				Type:   "leave",
				UserID: userID,
				Time:   now,
			})
		}
	}
}

// getRoomList 获取用户房间列表
func (pm *PresenceManager) getRoomList(userID string) []string {
	rooms := make([]string, 0)
	for roomID := range pm.userRooms[userID] {
		rooms = append(rooms, roomID)
	}
	return rooms
}

// broadcastEvent 广播事件到所有连接
func (pm *PresenceManager) broadcastEvent(event PresenceEvent) {
	if pm.manager == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	pm.manager.Broadcast(TextMessage, data)
}

// broadcastEventToRoom 广播事件到房间
func (pm *PresenceManager) broadcastEventToRoom(roomID string, event PresenceEvent) {
	if pm.manager == nil {
		return
	}

	data, err := json.Marshal(event)
	if err != nil {
		return
	}

	pm.manager.BroadcastToRoom(roomID, TextMessage, data)
}

// StartCleanupLoop 启动定期清理循环
func (pm *PresenceManager) StartCleanupLoop(interval time.Duration, timeout time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			pm.CleanupStale(timeout)
		}
	}()
}
