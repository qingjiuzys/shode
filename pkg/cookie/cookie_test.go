package cookie

import (
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewCookieManager 测试创建 Cookie 管理器
func TestNewCookieManager(t *testing.T) {
	cm := NewCookieManager()
	if cm == nil {
		t.Fatal("NewCookieManager() returned nil")
	}
}

// TestSetAndGetCookie 测试设置和获取 Cookie
func TestSetAndGetCookie(t *testing.T) {
	cm := NewCookieManager()
	
	// 创建测试响应
	w := httptest.NewRecorder()
	
	// 设置 Cookie
	err := cm.SetCookie(w, "test", "value", "")
	if err != nil {
		t.Fatalf("SetCookie() error = %v", err)
	}
	
	// 创建测试请求
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Cookie", w.Header().Get("Set-Cookie"))
	
	// 获取 Cookie
	value, err := cm.GetCookie(req, "test")
	if err != nil {
		t.Fatalf("GetCookie() error = %v", err)
	}
	
	if value != "value" {
		t.Errorf("GetCookie() = %v, want value", value)
	}
}

// TestGetCookieNotFound 测试获取不存在的 Cookie
func TestGetCookieNotFound(t *testing.T) {
	cm := NewCookieManager()
	
	req := httptest.NewRequest("GET", "/", nil)
	
	_, err := cm.GetCookie(req, "nonexistent")
	if err == nil {
		t.Error("GetCookie() should return error for nonexistent cookie")
	}
}

// TestDeleteCookie 测试删除 Cookie
func TestDeleteCookie(t *testing.T) {
	cm := NewCookieManager()
	
	w := httptest.NewRecorder()
	
	// 先设置 Cookie
	cm.SetCookie(w, "to_delete", "value", "")
	
	// 删除 Cookie
	err := cm.DeleteCookie(w, "to_delete", "/")
	if err != nil {
		t.Fatalf("DeleteCookie() error = %v", err)
	}
	
	// 验证已删除（MaxAge = 0 表示立即删除，Go 会将 -1 序列化为 0）
	// Check all Set-Cookie headers
	headers := w.Header()["Set-Cookie"]
	found := false
	for _, h := range headers {
		if strings.Contains(h, "to_delete") && strings.Contains(h, "Max-Age=0") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Expected deleted cookie with Max-Age=0, got headers: %v", headers)
	}
}

// TestSetCookieWithOptions 测试设置带选项的 Cookie
func TestSetCookieWithOptions(t *testing.T) {
	cm := NewCookieManager()
	
	w := httptest.NewRecorder()
	
	// 设置带多个选项的 Cookie
	options := "Path=/; Domain=.example.com; Max-Age=3600; Secure; HttpOnly"
	err := cm.SetCookie(w, "session", "value", options)
	if err != nil {
		t.Fatalf("SetCookie() error = %v", err)
	}
	
	// 验证 Cookie 包含选项
	setCookie := w.Header().Get("Set-Cookie")
	if !strings.Contains(setCookie, "Path=/") {
		t.Error("Cookie should contain Path=/")
	}
	if !strings.Contains(setCookie, "HttpOnly") {
		t.Error("Cookie should be HttpOnly")
	}
}
