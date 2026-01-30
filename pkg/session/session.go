package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Session 表示一个用户会话
type Session struct {
	ID        string                 // 会话 ID
	Data      map[string]interface{} // 会话数据
	CreatedAt time.Time             // 创建时间
	ExpiresAt time.Time             // 过期时间
}

// SessionManager 会话管理器
type SessionManager struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

// NewSessionManager 创建新的会话管理器
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		sessions: make(map[string]*Session),
	}
	
	// 启动过期清理
	go sm.cleanupExpired()
	
	return sm
}

// CreateSession 创建新会话
func (sm *SessionManager) CreateSession(userID string, ttlSeconds int) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// 生成唯一会话 ID
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}
	
	now := time.Now()
	session := &Session{
		ID:        sessionID,
		Data:      make(map[string]interface{}),
		CreatedAt: now,
		ExpiresAt: now.Add(time.Duration(ttlSeconds) * time.Second),
	}
	
	// 设置默认数据
	session.Data["user_id"] = userID
	session.Data["created_at"] = now.Unix()
	
	sm.sessions[sessionID] = session
	
	return session, nil
}

// GetSession 获取会话
func (sm *SessionManager) GetSession(sessionID string) (*Session, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		return nil, fmt.Errorf("session expired: %s", sessionID)
	}
	
	return session, nil
}

// DeleteSession 删除会话
func (sm *SessionManager) DeleteSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	if _, exists := sm.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	delete(sm.sessions, sessionID)
	return nil
}

// UpdateSession 更新会话数据
func (sm *SessionManager) UpdateSession(sessionID string, key string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	// 检查是否过期
	if time.Now().After(session.ExpiresAt) {
		delete(sm.sessions, sessionID)
		return fmt.Errorf("session expired: %s", sessionID)
	}
	
	session.Data[key] = value
	return nil
}

// GetSessionData 获取会话数据
func (sm *SessionManager) GetSessionData(sessionID, key string) (interface{}, error) {
	session, err := sm.GetSession(sessionID)
	if err != nil {
		return nil, err
	}
	
	value, exists := session.Data[key]
	if !exists {
		return nil, fmt.Errorf("key not found in session: %s", key)
	}
	
	return value, nil
}

// ExtendSession 延长会话有效期
func (sm *SessionManager) ExtendSession(sessionID string, ttlSeconds int) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	session.ExpiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	return nil
}

// GetActiveSessionCount 获取活跃会话数
func (sm *SessionManager) GetActiveSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	count := 0
	now := time.Now()
	for _, session := range sm.sessions {
		if now.Before(session.ExpiresAt) {
			count++
		}
	}
	return count
}

// GetAllSessions 获取所有会话（用于调试）
func (sm *SessionManager) GetAllSessions() []*Session {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	sessions := make([]*Session, 0, len(sm.sessions))
	for _, session := range sm.sessions {
		sessions = append(sessions, session)
	}
	return sessions
}

// cleanupExpired 清理过期会话
func (sm *SessionManager) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		sm.mu.Lock()
		now := time.Now()
		for id, session := range sm.sessions {
			if now.After(session.ExpiresAt) {
				delete(sm.sessions, id)
			}
		}
		sm.mu.Unlock()
	}
}

// generateSessionID 生成唯一的会话 ID
func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
