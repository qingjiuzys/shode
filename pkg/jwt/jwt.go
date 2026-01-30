package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// Claims JWT 声明
type Claims struct {
	Issuer    string                 `json:"iss"`
	Subject   string                 `json:"sub"`
	ExpiresAt int64                  `json:"exp"`
	IssuedAt  int64                  `json:"iat"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// JWTManager JWT 管理器
type JWTManager struct {
	secretKey string
}

// NewJWTManager 创建 JWT 管理器
func NewJWTManager(secretKey string) *JWTManager {
	return &JWTManager{
		secretKey: secretKey,
	}
}

// GenerateJWT 生成 JWT token
func (jm *JWTManager) GenerateJWT(claims map[string]interface{}) (string, error) {
	now := time.Now().Unix()
	
	claimsMap := map[string]interface{}{
		"iss":      "shode",
		"iat":      now,
		"exp":      now + 3600,
		"data":     claims,
	}
	
	if exp, ok := claims["exp"].(int); ok {
		claimsMap["exp"] = int64(exp)
	}
	if sub, ok := claims["sub"].(string); ok {
		claimsMap["sub"] = sub
	}
	
	header := base64.URLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payload := base64.URLEncoding.EncodeToString(mustMarshalJSON(claimsMap))
	
	signature := jm.sign(header + "." + payload)
	
	return header + "." + payload + "." + signature, nil
}

// VerifyJWT 验证 JWT token
func (jm *JWTManager) VerifyJWT(token string) (*Claims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	
	payloadBytes, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("failed to decode payload: %w", err)
	}
	
	var claims Claims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims: %w", err)
	}
	
	if time.Now().Unix() > claims.ExpiresAt {
		return nil, fmt.Errorf("token expired")
	}
	
	return &claims, nil
}

// sign 生成签名
func (jm *JWTManager) sign(data string) string {
	h := hmac.New(sha256.New, []byte(jm.secretKey))
	h.Write([]byte(data))
	return base64.URLEncoding.EncodeToString(h.Sum(nil))
}

func mustMarshalJSON(v interface{}) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}
