package jwt

import (
	"testing"
)

// TestNewJWTManager 测试创建 JWT 管理器
func TestNewJWTManager(t *testing.T) {
	jm := NewJWTManager("secret")
	if jm == nil {
		t.Fatal("NewJWTManager() returned nil")
	}
}

// TestGenerateAndVerifyJWT 测试生成和验证 JWT
func TestGenerateAndVerifyJWT(t *testing.T) {
	jm := NewJWTManager("secret")

	claims := map[string]interface{}{
		"sub": "user123",
	}

	token, err := jm.GenerateJWT(claims)
	if err != nil {
		t.Fatalf("GenerateJWT() error = %v", err)
	}

	if token == "" {
		t.Error("GenerateJWT() returned empty token")
	}

	verified, err := jm.VerifyJWT(token)
	if err != nil {
		t.Fatalf("VerifyJWT() error = %v", err)
	}

	if verified.Subject != "user123" {
		t.Errorf("Subject = %v, want user123", verified.Subject)
	}
}

// TestVerifyJWTInvalidToken 测试验证无效 token
func TestVerifyJWTInvalidToken(t *testing.T) {
	jm := NewJWTManager("secret")
	
	_, err := jm.VerifyJWT("invalid.token")
	if err == nil {
		t.Error("VerifyJWT() should return error for invalid format")
	}
}
