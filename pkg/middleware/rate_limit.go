package middleware

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimitMiddleware 请求限流中间件
type RateLimitMiddleware struct {
	*BaseMiddleware
	config  *RateLimitConfig
	limiter  *RateLimiter
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	// RequestsPerMinute 每分钟请求数
	RequestsPerMinute int
	// BurstSize 突发大小
	BurstSize int
	// KeyExtractor 用于提取限流 key 的函数
	KeyExtractor func(r *http.Request) string
}

// DefaultRateLimitConfig 默认限流配置
var DefaultRateLimitConfig = &RateLimitConfig{
	RequestsPerMinute: 60,
	BurstSize:        10,
	KeyExtractor:      func(r *http.Request) string { return "global" },
}

// RateLimiter 限流器
type RateLimiter struct {
	mu    sync.Mutex
	tokens map[string]*tokenBucket
}

// tokenBucket 令牌桶
type tokenBucket struct {
	tokens     int
	maxTokens  int
	lastRefill time.Time
}

// NewRateLimitMiddleware 创建限流中间件
func NewRateLimitMiddleware(config *RateLimitConfig) *RateLimitMiddleware {
	if config == nil {
		config = DefaultRateLimitConfig
	}

	if config.KeyExtractor == nil {
		config.KeyExtractor = func(r *http.Request) string { return "global" }
	}

	return &RateLimitMiddleware{
		BaseMiddleware: NewBaseMiddleware("rate_limit", 200, nil),
		config:        config,
		limiter: &RateLimiter{
			tokens: make(map[string]*tokenBucket),
		},
	}
}

// Process 处理限流逻辑
func (rl *RateLimitMiddleware) Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
	// 提取限流 key
	key := rl.config.KeyExtractor(r)

	// 检查是否超过限流
	allowed, resetTime := rl.limiter.Allow(key, rl.config.RequestsPerMinute, rl.config.BurstSize)

	if !allowed {
		// 设置限流响应头
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", rl.config.RequestsPerMinute))
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.Header().Set("X-RateLimit-Reset", resetTime.Format(time.RFC3339))
		w.Header().Set("Retry-After", "60")

		// 返回 429 Too Many Requests
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprintf(w, `{"error":"Rate limit exceeded","retry_after":60}`)
		return false
	}

	// 继续执行
	next(ctx, w, r)
	return true
}

// Allow 检查并消耗令牌
func (rl *RateLimiter) Allow(key string, rate int, burst int) (bool, time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	bucket, exists := rl.tokens[key]

	if !exists {
		bucket = &tokenBucket{
			tokens:     burst - 1,
			maxTokens:  burst,
			lastRefill: now,
		}
		rl.tokens[key] = bucket
		return true, now.Add(time.Minute)
	}

	// 计算需要补充的令牌数
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Minutes() * float64(rate))

	// 补充令牌（不超过上限）
	bucket.tokens += tokensToAdd
	if bucket.tokens > bucket.maxTokens {
		bucket.tokens = bucket.maxTokens
	}
	bucket.lastRefill = now

	// 检查是否有足够令牌
	if bucket.tokens > 0 {
		bucket.tokens--
		return true, now.Add(time.Minute)
	}

	return false, bucket.lastRefill.Add(time.Minute)
}

// GetStats 获取限流统计
func (rl *RateLimiter) GetStats(key string, burstSize int) (current int, max int, resetTime time.Time) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.tokens[key]
	if !exists {
		return 0, burstSize, time.Now().Add(time.Minute)
	}

	return bucket.tokens, bucket.maxTokens, bucket.lastRefill.Add(time.Minute)
}
