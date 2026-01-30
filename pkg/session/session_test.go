package session

import (
	"testing"
	"time"
)

// TestNewSessionManager 测试创建会话管理器
func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager()
	if sm == nil {
		t.Fatal("NewSessionManager() returned nil")
	}
}

// TestCreateSession 测试创建会话
func TestCreateSession(t *testing.T) {
	sm := NewSessionManager()
	
	session, err := sm.CreateSession("user123", 3600)
	if err != nil {
		t.Fatalf("CreateSession() error = %v", err)
	}
	
	if session.ID == "" {
		t.Error("Session ID should not be empty")
	}
	
	if session.Data["user_id"] != "user123" {
		t.Errorf("user_id = %v, want user123", session.Data["user_id"])
	}
}

// TestGetSession 测试获取会话
func TestGetSession(t *testing.T) {
	sm := NewSessionManager()
	
	// 创建会话
	session, _ := sm.CreateSession("user123", 3600)
	
	// 获取会话
	retrieved, err := sm.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession() error = %v", err)
	}
	
	if retrieved.Data["user_id"] != "user123" {
		t.Errorf("user_id = %v, want user123", retrieved.Data["user_id"])
	}
}

// TestGetSessionNotFound 测试获取不存在的会话
func TestGetSessionNotFound(t *testing.T) {
	sm := NewSessionManager()
	
	_, err := sm.GetSession("nonexistent")
	if err == nil {
		t.Error("GetSession() should return error for non-existent session")
	}
}

// TestDeleteSession 测试删除会话
func TestDeleteSession(t *testing.T) {
	sm := NewSessionManager()
	
	session, _ := sm.CreateSession("user123", 3600)
	
	err := sm.DeleteSession(session.ID)
	if err != nil {
		t.Fatalf("DeleteSession() error = %v", err)
	}
	
	// 验证已删除
	_, err = sm.GetSession(session.ID)
	if err == nil {
		t.Error("GetSession() should return error after DeleteSession()")
	}
}

// TestUpdateSession 测试更新会话数据
func TestUpdateSession(t *testing.T) {
	sm := NewSessionManager()
	
	session, _ := sm.CreateSession("user123", 3600)
	
	// 更新会话数据
	err := sm.UpdateSession(session.ID, "last_login", int64(1234567890))
	if err != nil {
		t.Fatalf("UpdateSession() error = %v", err)
	}
	
	// 验证更新
	value, err := sm.GetSessionData(session.ID, "last_login")
	if err != nil {
		t.Fatalf("GetSessionData() error = %v", err)
	}
	
	if value.(int64) != 1234567890 {
		t.Errorf("last_login = %v, want 1234567890", value)
	}
}

// TestGetSessionData 测试获取会话数据
func TestGetSessionData(t *testing.T) {
	sm := NewSessionManager()
	
	session, _ := sm.CreateSession("user123", 3600)
	session.Data["custom_key"] = "custom_value"
	
	// 获取数据
	value, err := sm.GetSessionData(session.ID, "custom_key")
	if err != nil {
		t.Fatalf("GetSessionData() error = %v", err)
	}
	
	if value != "custom_value" {
		t.Errorf("custom_key = %v, want custom_value", value)
	}
}

// TestExtendSession 测试延长会话
func TestExtendSession(t *testing.T) {
	sm := NewSessionManager()
	
	session, _ := sm.CreateSession("user123", 10)
	
	originalExpiry := session.ExpiresAt
	
	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
	
	// 延长会话
	sm.ExtendSession(session.ID, 20)
	
	// 获取更新后的会话
	updated, _ := sm.GetSession(session.ID)
	
	if updated.ExpiresAt.Before(originalExpiry) {
		t.Error("ExpiresAt should be extended")
	}
}

// TestGetActiveSessionCount 测试获取活跃会话数
func TestGetActiveSessionCount(t *testing.T) {
	sm := NewSessionManager()
	
	// 创建多个会话
	sm.CreateSession("user1", 3600)
	sm.CreateSession("user2", 3600)
	sm.CreateSession("user3", 3600)
	
	count := sm.GetActiveSessionCount()
	if count != 3 {
		t.Errorf("GetActiveSessionCount() = %d, want 3", count)
	}
}

// TestSessionExpiration 测试会话过期
func TestSessionExpiration(t *testing.T) {
	sm := NewSessionManager()
	
	// 创建 1 秒过期的会话
	session, _ := sm.CreateSession("user123", 1)
	
	// 等待过期
	time.Sleep(2 * time.Second)
	
	// 应该返回过期错误
	_, err := sm.GetSession(session.ID)
	if err == nil {
		t.Error("GetSession() should return error for expired session")
	}
}

// TestConcurrentAccess 测试并发访问
func TestConcurrentAccess(t *testing.T) {
	sm := NewSessionManager()
	
	session, _ := sm.CreateSession("user123", 3600)
	
	// 并发读取
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			sm.GetSession(session.ID)
			done <- true
		}()
	}
	
	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}
	
	// 验证会话仍然有效
	_, err := sm.GetSession(session.ID)
	if err != nil {
		t.Error("Session should still be valid after concurrent access")
	}
}
