// Package gateway 提供 API 网关功能。
package gateway

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter 速率限制器接口
type RateLimiter interface {
	Allow(key string) bool
	Reset(key string)
	GetLimit(key string) int
	GetRemaining(key string) int
}

// TokenBucketRateLimiter 令牌桶限流器
type TokenBucketRateLimiter struct {
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	rate    int           // 令牌生成速率（每秒）
	capacity int          // 桶容量
}

// TokenBucket 令牌桶
type TokenBucket struct {
	tokens    float64
	capacity  int
	lastRefill time.Time
	mu        sync.Mutex
}

// NewTokenBucketRateLimiter 创建令牌桶限流器
func NewTokenBucketRateLimiter(rate, capacity int) *TokenBucketRateLimiter {
	return &TokenBucketRateLimiter{
		buckets:  make(map[string]*TokenBucket),
		rate:     rate,
		capacity: capacity,
	}
}

// Allow 检查是否允许请求
func (rl *TokenBucketRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &TokenBucket{
			tokens:     float64(rl.capacity),
			capacity:   rl.capacity,
			lastRefill: time.Now(),
		}
		rl.buckets[key] = bucket
	}
	rl.mu.Unlock()

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	// 补充令牌
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill).Seconds()
	tokensToAdd := elapsed * float64(rl.rate)

	bucket.tokens += tokensToAdd
	if bucket.tokens > float64(bucket.capacity) {
		bucket.tokens = float64(bucket.capacity)
	}
	bucket.lastRefill = now

	// 检查是否有足够令牌
	if bucket.tokens >= 1 {
		bucket.tokens--
		return true
	}

	return false
}

// Reset 重置限流器
func (rl *TokenBucketRateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if bucket, exists := rl.buckets[key]; exists {
		bucket.mu.Lock()
		bucket.tokens = float64(bucket.capacity)
		bucket.lastRefill = time.Now()
		bucket.mu.Unlock()
	}
}

// GetLimit 获取限制
func (rl *TokenBucketRateLimiter) GetLimit(key string) int {
	return rl.rate
}

// GetRemaining 获取剩余令牌
func (rl *TokenBucketRateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	bucket, exists := rl.buckets[key]
	rl.mu.RUnlock()

	if !exists {
		return rl.capacity
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	return int(bucket.tokens)
}

// SlidingWindowRateLimiter 滑动窗口限流器
type SlidingWindowRateLimiter struct {
	windows map[string]*SlidingWindow
	mu      sync.RWMutex
	limit   int
	window  time.Duration
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	requests []time.Time
	limit    int
	window   time.Duration
	mu       sync.Mutex
}

// NewSlidingWindowRateLimiter 创建滑动窗口限流器
func NewSlidingWindowRateLimiter(limit int, window time.Duration) *SlidingWindowRateLimiter {
	return &SlidingWindowRateLimiter{
		windows: make(map[string]*SlidingWindow),
		limit:   limit,
		window:  window,
	}
}

// Allow 检查是否允许请求
func (rl *SlidingWindowRateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	window, exists := rl.windows[key]
	if !exists {
		window = &SlidingWindow{
			requests: make([]time.Time, 0),
			limit:    rl.limit,
			window:   rl.window,
		}
		rl.windows[key] = window
	}
	rl.mu.Unlock()

	window.mu.Lock()
	defer window.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// 清理过期请求
	valid := make([]time.Time, 0)
	for _, t := range window.requests {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	window.requests = valid

	// 检查是否超限
	if len(window.requests) >= rl.limit {
		return false
	}

	// 记录请求
	window.requests = append(window.requests, now)
	return true
}

// Reset 重置限流器
func (rl *SlidingWindowRateLimiter) Reset(key string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if window, exists := rl.windows[key]; exists {
		window.mu.Lock()
		window.requests = make([]time.Time, 0)
		window.mu.Unlock()
	}
}

// GetLimit 获取限制
func (rl *SlidingWindowRateLimiter) GetLimit(key string) int {
	return rl.limit
}

// GetRemaining 获取剩余请求数
func (rl *SlidingWindowRateLimiter) GetRemaining(key string) int {
	rl.mu.RLock()
	window, exists := rl.windows[key]
	rl.mu.RUnlock()

	if !exists {
		return rl.limit
	}

	window.mu.Lock()
	defer window.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	count := 0
	for _, t := range window.requests {
		if t.After(cutoff) {
			count++
		}
	}

	return rl.limit - count
}

// CircuitBreakerState 熔断器状态
type CircuitBreakerState int

const (
	StateClosed CircuitBreakerState = iota
	StateHalfOpen
	StateOpen
)

// CircuitBreaker 熔断器
type CircuitBreaker struct {
	name           string
	state          CircuitBreakerState
	maxFailures    int
	resetTimeout   time.Duration
	failureCount   int
	lastFailureTime time.Time
	mu             sync.RWMutex
	onStateChange  func(name string, from, to CircuitBreakerState)
}

// CircuitBreakerConfig 熔断器配置
type CircuitBreakerConfig struct {
	Name          string
	MaxFailures   int
	ResetTimeout  time.Duration
	OnStateChange func(name string, from, to CircuitBreakerState)
}

// NewCircuitBreaker 创建熔断器
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		name:          config.Name,
		state:         StateClosed,
		maxFailures:   config.MaxFailures,
		resetTimeout:  config.ResetTimeout,
		onStateChange: config.OnStateChange,
	}
}

// Allow 检查是否允许请求
func (cb *CircuitBreaker) Allow() bool {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	// 如果是打开状态，检查是否应该尝试恢复
	if cb.state == StateOpen {
		if time.Since(cb.lastFailureTime) > cb.resetTimeout {
			cb.setState(StateHalfOpen)
			return true
		}
		return false
	}

	return true
}

// RecordSuccess 记录成功
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if cb.state == StateHalfOpen {
		cb.setState(StateClosed)
	}

	cb.failureCount = 0
}

// RecordFailure 记录失败
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.failureCount++
	cb.lastFailureTime = time.Now()

	if cb.failureCount >= cb.maxFailures {
		cb.setState(StateOpen)
	}
}

// setState 设置状态
func (cb *CircuitBreaker) setState(state CircuitBreakerState) {
	if cb.state != state {
		oldState := cb.state
		cb.state = state

		if cb.onStateChange != nil {
			cb.onStateChange(cb.name, oldState, state)
		}
	}
}

// GetState 获取状态
func (cb *CircuitBreaker) GetState() CircuitBreakerState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.state
}

// GetFailureCount 获取失败次数
func (cb *CircuitBreaker) GetFailureCount() int {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return cb.failureCount
}

// RateLimitMiddleware 速率限制中间件
type RateLimitMiddleware struct {
	limiter      RateLimiter
	keyExtractor func(*http.Request) string
	onLimitExceeded func(http.ResponseWriter, *http.Request)
}

// NewRateLimitMiddleware 创建速率限制中间件
func NewRateLimitMiddleware(limiter RateLimiter) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:      limiter,
		keyExtractor: defaultKeyExtractor,
		onLimitExceeded: defaultLimitExceededHandler,
	}
}

// ServeHTTP 处理 HTTP 请求
func (m *RateLimitMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	key := m.keyExtractor(r)

	if !m.limiter.Allow(key) {
		m.onLimitExceeded(w, r)
		return
	}

	next(w, r)
}

// defaultKeyExtractor 默认密钥提取器
func defaultKeyExtractor(r *http.Request) string {
	return r.RemoteAddr
}

// defaultLimitExceededHandler 默认限流处理
func defaultLimitExceededHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusTooManyRequests)
	w.Write([]byte(`{"error":"rate limit exceeded"}`))
}

// SetKeyExtractor 设置密钥提取器
func (m *RateLimitMiddleware) SetKeyExtractor(fn func(*http.Request) string) {
	m.keyExtractor = fn
}

// SetOnLimitExceeded 设置限流处理
func (m *RateLimitMiddleware) SetOnLimitExceeded(fn func(http.ResponseWriter, *http.Request)) {
	m.onLimitExceeded = fn
}

// CircuitBreakerMiddleware 熔断中间件
type CircuitBreakerMiddleware struct {
	breaker            *CircuitBreaker
	onCircuitOpen      func(http.ResponseWriter, *http.Request)
}

// NewCircuitBreakerMiddleware 创建熔断中间件
func NewCircuitBreakerMiddleware(breaker *CircuitBreaker) *CircuitBreakerMiddleware {
	return &CircuitBreakerMiddleware{
		breaker:       breaker,
		onCircuitOpen: defaultCircuitOpenHandler,
	}
}

// ServeHTTP 处理 HTTP 请求
func (m *CircuitBreakerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	if !m.breaker.Allow() {
		m.onCircuitOpen(w, r)
		return
	}

	// 包装 ResponseWriter 来检测错误
	wrapped := &circuitBreakerResponseWriter{
		ResponseWriter: w,
		breaker:        m.breaker,
	}

	next(wrapped, r)
}

// defaultCircuitOpenHandler 默认熔断处理
func defaultCircuitOpenHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write([]byte(`{"error":"circuit breaker is open"}`))
}

// circuitBreakerResponseWriter 熔断响应包装器
type circuitBreakerResponseWriter struct {
	http.ResponseWriter
	breaker *CircuitBreaker
	statusCode int
	written bool
}

// WriteHeader 写入状态码
func (w *circuitBreakerResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.written = true
	w.ResponseWriter.WriteHeader(statusCode)

	// 记录失败
	if statusCode >= 500 {
		w.breaker.RecordFailure()
	} else {
		w.breaker.RecordSuccess()
	}
}

// Write 写入响应
func (w *circuitBreakerResponseWriter) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}
