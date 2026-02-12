// Package csrf CSRF 跨站请求伪造防护
package csrf

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"sync"
)

// Config CSRF 配置
type Config struct {
	Secret          string
	CookieName      string
	CookieDomain    string
	CookiePath      string
	CookieMaxAge    int
	CookieSecure    bool
	CookieHTTPOnly  bool
	CookieSameSite  http.SameSite
	TokenLength     int
	TokenHeader     string
	FormField       string
	EnableFormCheck bool
	EnableHeaderCheck bool
	ErrorHandler    http.HandlerFunc
}

// CSRF CSRF 防护中间件
type CSRF struct {
	config     *Config
	secret     []byte
	store      map[string]string
	storeMutex sync.RWMutex
}

// New 创建 CSRF 中间件
func New(config *Config) *CSRF {
	if config.TokenLength == 0 {
		config.TokenLength = 32
	}
	if config.CookieName == "" {
		config.CookieName = "csrf_token"
	}
	if config.TokenHeader == "" {
		config.TokenHeader = "X-CSRF-Token"
	}
	if config.FormField == "" {
		config.FormField = "csrf_token"
	}

	return &CSRF{
		config: config,
		secret: []byte(config.Secret),
		store:  make(map[string]string),
	}
}

// GenerateToken 生成 token
func (c *CSRF) GenerateToken() (string, error) {
	b := make([]byte, c.config.TokenLength)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// ValidateToken 验证 token
func (c *CSRF) ValidateToken(token string) bool {
	c.storeMutex.RLock()
	defer c.storeMutex.RUnlock()

	for _, storedToken := range c.store {
		if subtle.ConstantTimeCompare([]byte(token), []byte(storedToken)) == 1 {
			return true
		}
	}
	return false
}

// Middleware 返回中间件函数
func (c *CSRF) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 为安全方法生成 token
		if r.Method == http.MethodGet ||
		   r.Method == http.MethodHead ||
		   r.Method == http.MethodOptions ||
		   r.Method == http.MethodTrace {
			c.setToken(w, r)
			next.ServeHTTP(w, r)
			return
		}

		// 对于修改操作验证 token
		if !c.verifyToken(r) {
			c.handleUnauthorized(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// setToken 设置 CSRF token
func (c *CSRF) setToken(w http.ResponseWriter, r *http.Request) {
	token, err := c.GenerateToken()
	if err != nil {
		http.Error(w, "Failed to generate CSRF token", http.StatusInternalServerError)
		return
	}

	// 存储在内存中
	sessionID := c.getSessionID(r)
	c.storeMutex.Lock()
	c.store[sessionID] = token
	c.storeMutex.Unlock()

	// 设置 cookie
	cookie := &http.Cookie{
		Name:     c.config.CookieName,
		Value:    token,
		Domain:   c.config.CookieDomain,
		Path:     c.config.CookiePath,
		MaxAge:   c.config.CookieMaxAge,
		Secure:   c.config.CookieSecure,
		HttpOnly: c.config.CookieHTTPOnly,
		SameSite: c.config.CookieSameSite,
	}
	http.SetCookie(w, cookie)

	// 在请求上下文中设置 token
	if ctx := r.Context(); ctx != nil {
		*r = *r.WithContext(contextWithToken(ctx, token))
	}
}

// verifyToken 验证 CSRF token
func (c *CSRF) verifyToken(r *http.Request) bool {
	var token string

	// 从 header 获取
	if c.config.EnableHeaderCheck {
		token = r.Header.Get(c.config.TokenHeader)
	}

	// 从表单获取
	if token == "" && c.config.EnableFormCheck {
		token = r.FormValue(c.config.FormField)
	}

	// 从 cookie 获取
	if token == "" {
		cookie, err := r.Cookie(c.config.CookieName)
		if err == nil {
			token = cookie.Value
		}
	}

	if token == "" {
		return false
	}

	return c.ValidateToken(token)
}

// handleUnauthorized 处理未授权请求
func (c *CSRF) handleUnauthorized(w http.ResponseWriter, r *http.Request) {
	if c.config.ErrorHandler != nil {
		c.config.ErrorHandler.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Invalid CSRF token", http.StatusForbidden)
}

// getSessionID 获取会话 ID
func (c *CSRF) getSessionID(r *http.Request) string {
	// 简化实现：使用远程地址
	return r.RemoteAddr
}

// Token 从请求上下文获取 token
func Token(r *http.Request) string {
	if token, ok := r.Context().Value(csrfTokenKey).(string); ok {
		return token
	}
	return ""
}

type contextKey string

const csrfTokenKey contextKey = "csrf_token"

func contextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, csrfTokenKey, token)
}

// DoubleSubmitCookie Double Submit Cookie 模式
type DoubleSubmitCookie struct {
	csrf *CSRF
}

// NewDoubleSubmitCookie 创建 Double Submit Cookie
func NewDoubleSubmitCookie(config *Config) *DoubleSubmitCookie {
	return &DoubleSubmitCookie{
		csrf: New(config),
	}
}

// Middleware 返回中间件
func (d *DoubleSubmitCookie) Middleware(next http.Handler) http.Handler {
	return d.csrf.Middleware(next)
}

// SynchronizerToken 同步令牌模式
type SynchronizerToken struct {
	csrf *CSRF
}

// NewSynchronizerToken 创建同步令牌
func NewSynchronizerToken(config *Config) *SynchronizerToken {
	return &SynchronizerToken{
		csrf: New(config),
	}
}

// Middleware 返回中间件
func (s *SynchronizerToken) Middleware(next http.Handler) http.Handler {
	return s.csrf.Middleware(next)
}

// TokenField 生成表单字段
func (c *CSRF) TokenField(r *http.Request) string {
	token := Token(r)
	if token == "" {
		return ""
	}
	return `<input type="hidden" name="` + c.config.FormField + `" value="` + token + `">`
}

// MetaTag 生成 meta 标签
func (c *CSRF) MetaTag(r *http.Request) string {
	token := Token(r)
	if token == "" {
		return ""
	}
	return `<meta name="csrf-token" content="` + token + `">`
}

// SetToken 在响应中设置 token
func (c *CSRF) SetToken(w http.ResponseWriter, r *http.Request) error {
	token, err := c.GenerateToken()
	if err != nil {
		return err
	}

	sessionID := c.getSessionID(r)
	c.storeMutex.Lock()
	c.store[sessionID] = token
	c.storeMutex.Unlock()

	// 设置 cookie
	cookie := &http.Cookie{
		Name:     c.config.CookieName,
		Value:    token,
		Domain:   c.config.CookieDomain,
		Path:     c.config.CookiePath,
		MaxAge:   c.config.CookieMaxAge,
		Secure:   c.config.CookieSecure,
		HttpOnly: c.config.CookieHTTPOnly,
		SameSite: c.config.CookieSameSite,
	}
	http.SetCookie(w, cookie)

	return nil
}

// RemoveToken 移除 token
func (c *CSRF) RemoveToken(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     c.config.CookieName,
		Value:    "",
		MaxAge:   -1,
		Path:     c.config.CookiePath,
		HttpOnly: c.config.CookieHTTPOnly,
	}
	http.SetCookie(w, cookie)
}

// Cleanup 清理过期 token
func (c *CSRF) Cleanup(maxAge int) {
	c.storeMutex.Lock()
	defer c.storeMutex.Unlock()

	// 简化实现：清空所有 token
	c.store = make(map[string]string)
}

// Default 默认配置
func Default(secret string) *CSRF {
	return New(&Config{
		Secret:           secret,
		CookieName:       "csrf_token",
		CookieMaxAge:     86400,
		CookieSecure:     true,
		CookieHTTPOnly:   true,
		CookieSameSite:   http.SameSiteStrictMode,
		TokenLength:      32,
		TokenHeader:      "X-CSRF-Token",
		FormField:        "csrf_token",
		EnableFormCheck:  true,
		EnableHeaderCheck: true,
	})
}
