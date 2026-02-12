// Package ratelimit 速率限制
package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Config 速率限制配置
type Config struct {
	Rate    float64       // 每秒请求数
	Burst   int           // 突发请求数
	Window  time.Duration // 时间窗口
	KeyFunc func(*http.Request) string
}

// RateLimiter 速率限制器
type RateLimiter struct {
	config *Config
}

// TokenBucket 令牌桶算法
type TokenBucket struct {
	rate     float64
	capacity float64
	tokens   float64
	lastTime time.Time
	mu       sync.Mutex
}

// NewTokenBucket 创建令牌桶
func NewTokenBucket(config Config) *TokenBucket {
	return &TokenBucket{
		rate:     config.Rate,
		capacity: float64(config.Burst),
		tokens:   float64(config.Burst),
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastTime).Seconds()

	// 添加令牌
	tb.tokens += elapsed * tb.rate
	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}
	tb.lastTime = now

	// 消费令牌
	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// SlidingWindow 滑动窗口算法
type SlidingWindow struct {
	window time.Duration
	limit  int
	events []time.Time
	mu     sync.Mutex
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(config Config) *SlidingWindow {
	return &SlidingWindow{
		window: config.Window,
		limit:  config.Burst,
		events: make([]time.Time, 0),
	}
}

// Allow 检查是否允许请求
func (sw *SlidingWindow) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-sw.window)

	// 移除窗口外的事件
	for len(sw.events) > 0 && sw.events[0].Before(cutoff) {
		sw.events = sw.events[1:]
	}

	// 检查是否超过限制
	if len(sw.events) >= sw.limit {
		return false
	}

	sw.events = append(sw.events, now)
	return true
}

// FixedWindow 固定窗口算法
type FixedWindow struct {
	limit    int
	window   time.Duration
	count    int
	lastTime time.Time
	mu       sync.Mutex
}

// NewFixedWindow 创建固定窗口
func NewFixedWindow(config Config) *FixedWindow {
	return &FixedWindow{
		limit:    config.Burst,
		window:   config.Window,
		lastTime: time.Now(),
	}
}

// Allow 检查是否允许请求
func (fw *FixedWindow) Allow() bool {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(fw.lastTime)

	// 重置窗口
	if elapsed >= fw.window {
		fw.count = 0
		fw.lastTime = now
	}

	// 检查是否超过限制
	if fw.count >= fw.limit {
		return false
	}

	fw.count++
	return true
}

// Middleware 返回速率限制中间件
func Middleware(limiter interface{}) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var allowed bool

			switch l := limiter.(type) {
			case *TokenBucket:
				allowed = l.Allow()
			case *SlidingWindow:
				allowed = l.Allow()
			case *FixedWindow:
				allowed = l.Allow()
			default:
				allowed = true
			}

			if !allowed {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// New 创建速率限制器
func New(algorithm string, config Config) interface{} {
	switch algorithm {
	case "token-bucket":
		return NewTokenBucket(config)
	case "sliding-window":
		return NewSlidingWindow(config)
	case "fixed-window":
		return NewFixedWindow(config)
	default:
		return NewTokenBucket(config)
	}
}

// Default 默认配置
func Default() *TokenBucket {
	return NewTokenBucket(Config{
		Rate:   100,
		Burst:  200,
		Window: time.Second,
	})
}
