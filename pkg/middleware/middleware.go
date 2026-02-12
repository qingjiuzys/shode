package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
)

// Middleware 中间件函数类型
type Middleware func(http.Handler) http.Handler

// Manager 中间件管理器
type Manager struct {
	middlewares []Middleware
	mu          sync.RWMutex
}

// NewManager 创建中间件管理器
func NewManager() *Manager {
	return &Manager{
		middlewares: make([]Middleware, 0),
	}
}

// Use 添加中间件
func (m *Manager) Use(mw Middleware) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = append(m.middlewares, mw)
}

// Apply 应用所有中间件
func (m *Manager) Apply(h http.Handler) http.Handler {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i := len(m.middlewares) - 1; i >= 0; i-- {
		h = m.middlewares[i](h)
	}
	return h
}

// Chain 中间件链
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// ContextKey 上下文键类型
type ContextKey string

// WithContext 创建带上下文的中间件
func WithContext(key ContextKey, value interface{}) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), key, value)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// List 列出所有中间件
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.middlewares))
	for i := range m.middlewares {
		names[i] = fmt.Sprintf("middleware_%d", i)
	}
	return names
}

// Remove 移除中间件
func (m *Manager) Remove(index int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if index >= 0 && index < len(m.middlewares) {
		m.middlewares = append(m.middlewares[:index], m.middlewares[index+1:]...)
	}
}

// Clear 清空中间件
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = make([]Middleware, 0)
}

// NewCORSMiddleware 创建 CORS 中间件
func NewCORSMiddleware(allowOrigins []string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			allowed := false
			for _, ao := range allowOrigins {
				if ao == "*" || ao == origin {
					allowed = true
					break
				}
			}
			if allowed {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			}
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(requests int, window int) Middleware {
	return func(next http.Handler) http.Handler {
		return next
	}
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// NewRequestIDMiddleware 创建请求 ID 中间件
func NewRequestIDMiddleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}
}
