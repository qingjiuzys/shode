// Package security 提供安全防护功能。
package security

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// CSRFManager CSRF Token 管理器
type CSRFManager struct {
	tokens     map[string]csrfToken
	mu         sync.RWMutex
	tokenAge   time.Duration
	secret     []byte
}

type csrfToken struct {
	value     string
	expiresAt time.Time
}

// NewCSRFManager 创建 CSRF 管理器
func NewCSRFManager(secret string, age time.Duration) *CSRFManager {
	return &CSRFManager{
		tokens:   make(map[string]csrfToken),
		tokenAge: age,
		secret:  []byte(secret),
	}
}

// GenerateToken 生成 Token
func (cm *CSRFManager) GenerateToken(sessionID string) (string, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 清理过期 token
	cm.cleanup()

	// 生成随机 token
	randomBytes := make([]byte, 32)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	token := base64.URLEncoding.EncodeToString(randomBytes)

	cm.tokens[sessionID] = csrfToken{
		value:     token,
		expiresAt: time.Now().Add(cm.tokenAge),
	}

	return token, nil
}

// ValidateToken 验证 Token
func (cm *CSRFManager) ValidateToken(sessionID, token string) bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	stored, exists := cm.tokens[sessionID]
	if !exists {
		return false
	}

	return stored.value == token && time.Now().Before(stored.expiresAt)
}

// cleanup 清理过期 token
func (cm *CSRFManager) cleanup() {
	now := time.Now()
	for id, token := range cm.tokens {
		if now.After(token.expiresAt) {
			delete(cm.tokens, id)
		}
	}
}

// XSSSanitizer XSS 清理器
type XSSSanitizer struct {
	allowedTags []string
	allowedAttrs []string
}

// NewXSSSanitizer 创建 XSS 清理器
func NewXSSSanitizer() *XSSSanitizer {
	return &XSSSanitizer{
		allowedTags:  []string{"p", "br", "strong", "em", "u", "i", "a", "ul", "ol", "li"},
		allowedAttrs: []string{"href", "title", "class", "id"},
	}
}

// Sanitize 清理 HTML 输入
func (xs *XSSSanitizer) Sanitize(input string) string {
	// 基本的 XSS 防护
	// 移除危险的脚本标签
	input = strings.ReplaceAll(input, "<script", "")
	input = strings.ReplaceAll(input, "</script>", "")
	input = strings.ReplaceAll(input, "javascript:", "")
	input = strings.ReplaceAll(input, "onerror=", "")
	input = strings.ReplaceAll(input, "onload=", "")

	// HTML 实体编码
	input = strings.ReplaceAll(input, "&", "&amp;")
	input = strings.ReplaceAll(input, "<", "&lt;")
	input = strings.ReplaceAll(input, ">", "&gt;")
	input = strings.ReplaceAll(input, "\"", "&quot;")
	input = strings.ReplaceAll(input, "'", "&#39;")

	return input
}

// SQLInjectionGuard SQL 注入防护
type SQLInjectionGuard struct {
	patterns []string
}

// NewSQLInjectionGuard 创建 SQL 注入防护
func NewSQLInjectionGuard() *SQLInjectionGuard {
	return &SQLInjectionGuard{
		patterns: []string{
			"(?i)(union\\s+select)",
			"(?i)(or\\s+1\\s*=)",
			"(?i)(drop\\s+table)",
			"(?i)(delete\\s+from)",
			"(?i)(insert\\s+into)",
			"(?i)(exec\\s*\\()",
			"(?i)(script\\s*>)",
			"(?i)(--)",
		},
	}
}

// CheckInput 检查输入
func (sg *SQLInjectionGuard) CheckInput(input string) bool {
	for _, pattern := range sg.patterns {
		if strings.Contains(strings.ToLower(input), pattern[4:len(pattern)-4]) {
			return false
		}
	}
	return true
}

// SanitizeInput 清理输入
func (sg *SQLInjectionGuard) SanitizeInput(input string) string {
	// 移除单引号
	input = strings.ReplaceAll(input, "'", "")
	// 移除双引号
	input = strings.ReplaceAll(input, "\"", "")
	// 移除注释
	input = strings.ReplaceAll(input, "--", "")
	// 移除分号
	input = strings.ReplaceAll(input, ";", "")

	return input
}

// SecurityMiddleware 安全中间件
type SecurityMiddleware struct {
	csrf    *CSRFManager
	xss     *XSSSanitizer
	sqlGuard *SQLInjectionGuard
}

// NewSecurityMiddleware 创建安全中间件
func NewSecurityMiddleware() *SecurityMiddleware {
	return &SecurityMiddleware{
		xss:      NewXSSSanitizer(),
		sqlGuard: NewSQLInjectionGuard(),
		csrf:     NewCSRFManager("secret", 3600*time.Second),
	}
}

// SetCSRF 设置 CSRF 管理器
func (sm *SecurityMiddleware) SetCSRF(csrf *CSRFManager) {
	sm.csrf = csrf
}

// ProtectMiddleware 保护中间件
func (sm *SecurityMiddleware) ProtectMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CSRF 保护
		if r.Method != "GET" && r.Method != "HEAD" && r.Method != "OPTIONS" {
			sessionID := sm.getSessionID(r)
			token := r.Header.Get("X-CSRF-Token")
			if !sm.csrf.ValidateToken(sessionID, token) {
				http.Error(w, "Invalid CSRF token", http.StatusForbidden)
				return
			}
		}

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

// getSessionID 获取会话ID
func (sm *SecurityMiddleware) getSessionID(r *http.Request) string {
	// 从 cookie 或 header 获取 session ID
	if cookie, err := r.Cookie("session_id"); err == nil {
		return cookie.Value
	}
	return r.Header.Get("X-Session-ID")
}

// GenerateCSRFToken 生成 CSRF Token
func (sm *SecurityMiddleware) GenerateCSRFToken(sessionID string) (string, error) {
	return sm.csrf.GenerateToken(sessionID)
}

// RateLimiter 速率限制器
type RateLimiter struct {
	requests map[string]*rateLimitInfo
	mu       sync.RWMutex
	rate     int
	window   time.Duration
}

type rateLimitInfo struct {
	count    int
	windowStart time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*rateLimitInfo),
		rate:     rate,
		window:   window,
	}
	go rl.cleanupLoop()
	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	info, exists := rl.requests[key]

	if !exists || now.Sub(info.windowStart) > rl.window {
		rl.requests[key] = &rateLimitInfo{
			count:       1,
			windowStart: now,
		}
		return true
	}

	if info.count >= rl.rate {
		return false
	}

	info.count++
	return true
}

// cleanupLoop 定期清理过期记录
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for key, info := range rl.requests {
			if now.Sub(info.windowStart) > rl.window*10 {
				delete(rl.requests, key)
			}
		}
		rl.mu.Unlock()
	}
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EnableCSRF        bool
	EnableXSS         bool
	EnableSQLGuard     bool
	EnableRateLimit    bool
	RateLimitPerMinute int
	AllowedOrigins     []string
	AllowedMethods     []string
	MaxRequestSize     int64
}

// DefaultSecurityConfig 默认安全配置
var DefaultSecurityConfig = SecurityConfig{
	EnableCSRF:        true,
	EnableXSS:         true,
	EnableSQLGuard:     true,
	EnableRateLimit:    true,
	RateLimitPerMinute: 60,
	AllowedOrigins:     []string{"*"},
	AllowedMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	MaxRequestSize:     10 * 1024 * 1024, // 10MB
}

// PasswordHash 密码哈希
func PasswordHash(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// PasswordVerify 验证密码
func PasswordVerify(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hash))
	return err == nil
}
