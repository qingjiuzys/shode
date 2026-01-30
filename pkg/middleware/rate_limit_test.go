package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddleware_Allow(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 10,
		BurstSize:         5,
		KeyExtractor:      func(r *http.Request) string { return "test-key" },
	}
	mw := NewRateLimitMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("Rate limit middleware should allow first request")
	}

	if !called {
		t.Error("Next function should be called when under limit")
	}
}

func TestRateLimitMiddleware_Exceed(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		KeyExtractor:      func(r *http.Request) string { return "test-key" },
	}
	mw := NewRateLimitMiddleware(config)

	// First request should pass
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	next1 := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	result1 := mw.Process(context.Background(), w1, req1, next1)

	if !result1 {
		t.Error("First request should be allowed")
	}

	// Second request should be rate limited
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	next2 := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		t.Error("Next function should not be called when rate limited")
	}
	result2 := mw.Process(context.Background(), w2, req2, next2)

	if result2 {
		t.Error("Second request should be rate limited")
	}

	if w2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w2.Code)
	}
}

func TestRateLimitMiddleware_DifferentKeys(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		KeyExtractor:      func(r *http.Request) string { return r.Header.Get("X-API-Key") },
	}
	mw := NewRateLimitMiddleware(config)

	// First request with key1
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-API-Key", "key1")
	w1 := httptest.NewRecorder()
	next1 := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	result1 := mw.Process(context.Background(), w1, req1, next1)

	if !result1 {
		t.Error("Request with key1 should be allowed")
	}

	// Second request with key2 (different key, should also be allowed)
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-API-Key", "key2")
	w2 := httptest.NewRecorder()
	next2 := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {}
	result2 := mw.Process(context.Background(), w2, req2, next2)

	if !result2 {
		t.Error("Request with key2 should be allowed")
	}
}

func TestRateLimiter_TokenRefill(t *testing.T) {
	limiter := &RateLimiter{
		tokens: make(map[string]*tokenBucket),
	}

	// Create a bucket with rate of 60 tokens per minute (1 token per second)
	key := "test-key"
	rate := 60
	burst := 5

	// Consume all burst tokens
	for i := 0; i < burst; i++ {
		allowed, _ := limiter.Allow(key, rate, burst)
		if !allowed {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// Next request should be denied
	allowed, _ := limiter.Allow(key, rate, burst)
	if allowed {
		t.Error("Request after burst exhaustion should be denied")
	}

	// Wait for 2 seconds (2 tokens should be refilled)
	time.Sleep(2 * time.Second)

	// Should be allowed again after refill
	allowed, _ = limiter.Allow(key, rate, burst)
	if !allowed {
		t.Error("Request should be allowed after token refill")
	}
}

func TestRateLimitMiddleware_ResponseHeaders(t *testing.T) {
	config := &RateLimitConfig{
		RequestsPerMinute: 1,
		BurstSize:         1,
		KeyExtractor:      func(r *http.Request) string { return "test" },
	}
	mw := NewRateLimitMiddleware(config)

	// First request (allowed)
	req1 := httptest.NewRequest("GET", "/test", nil)
	w1 := httptest.NewRecorder()
	mw.Process(context.Background(), w1, req1, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Second request (rate limited)
	req2 := httptest.NewRequest("GET", "/test", nil)
	w2 := httptest.NewRecorder()
	mw.Process(context.Background(), w2, req2, func(ctx context.Context, w http.ResponseWriter, r *http.Request) {})

	// Check response headers
	if w2.Header().Get("X-RateLimit-Limit") == "" {
		t.Error("X-RateLimit-Limit header should be set")
	}

	if w2.Header().Get("X-RateLimit-Remaining") != "0" {
		t.Errorf("X-RateLimit-Remaining should be 0, got '%s'", w2.Header().Get("X-RateLimit-Remaining"))
	}

	if w2.Header().Get("Retry-After") != "60" {
		t.Errorf("Retry-After should be 60, got '%s'", w2.Header().Get("Retry-After"))
	}

	// Check response body
	body := w2.Body.String()
	expectedBody := `{"error":"Rate limit exceeded","retry_after":60}`
	if body != expectedBody {
		t.Errorf("Expected body '%s', got '%s'", expectedBody, body)
	}
}

func TestRateLimiter_GetStats(t *testing.T) {
	limiter := &RateLimiter{
		tokens: make(map[string]*tokenBucket),
	}

	key := "test-key"
	rate := 60
	burst := 10

	// Consume 3 tokens
	for i := 0; i < 3; i++ {
		limiter.Allow(key, rate, burst)
	}

	current, max, resetTime := limiter.GetStats(key, burst)

	if max != burst {
		t.Errorf("Expected max %d, got %d", burst, max)
	}

	if current != 7 { // 10 - 3 = 7 remaining
		t.Errorf("Expected 7 remaining tokens, got %d", current)
	}

	if resetTime.IsZero() {
		t.Error("Reset time should not be zero")
	}
}

func ExampleRateLimiter() {
	config := &RateLimitConfig{
		RequestsPerMinute: 60,
		BurstSize:         10,
		KeyExtractor: func(r *http.Request) string {
			// Rate limit by IP address
			return r.RemoteAddr
		},
	}
	mw := NewRateLimitMiddleware(config)

	fmt.Println("Created rate limiter:", mw.Name())
	// Output: Created rate limiter: rate_limit
}
