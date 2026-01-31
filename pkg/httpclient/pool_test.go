package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestNewConnectionPool tests creating a new connection pool
func TestNewConnectionPool(t *testing.T) {
	pool := NewConnectionPool(nil)

	if pool == nil {
		t.Fatal("NewConnectionPool returned nil")
	}

	if pool.client == nil {
		t.Error("Pool client is nil")
	}

	if pool.config == nil {
		t.Error("Pool config is nil")
	}
}

// TestNewConnectionPoolWithConfig tests creating a pool with custom config
func TestNewConnectionPoolWithConfig(t *testing.T) {
	config := &PoolConfig{
		MaxConnsPerHost:       50,
		MaxIdleConns:          50,
		IdleConnTimeout:       60 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		Timeout:               15 * time.Second,
	}

	pool := NewConnectionPool(config)

	if pool.config.MaxConnsPerHost != 50 {
		t.Errorf("Expected MaxConnsPerHost 50, got %d", pool.config.MaxConnsPerHost)
	}

	if pool.config.Timeout != 15*time.Second {
		t.Errorf("Expected Timeout 15s, got %v", pool.config.Timeout)
	}
}

// TestConnectionPoolGet tests GET requests
func TestConnectionPoolGet(t *testing.T) {
	// Create a test server
	requestCount := int64(0)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&requestCount, 1)
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	pool := NewConnectionPool(nil)
	defer pool.Close()

	ctx := context.Background()
	resp, err := pool.Get(ctx, server.URL)
	if err != nil {
		t.Fatalf("GET request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if atomic.LoadInt64(&requestCount) != 1 {
		t.Errorf("Expected 1 request, got %d", requestCount)
	}
}

// TestConnectionPoolDo tests generic Do requests
func TestConnectionPoolDo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("Created"))
	}))
	defer server.Close()

	pool := NewConnectionPool(nil)
	defer pool.Close()

	req, err := http.NewRequest("POST", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := pool.Do(req)
	if err != nil {
		t.Fatalf("POST request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}
}

// TestConnectionPoolStats tests statistics tracking
func TestConnectionPoolStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	pool := NewConnectionPool(nil)
	defer pool.Close()
	pool.ResetStats()

	ctx := context.Background()

	// Make some requests
	for i := 0; i < 5; i++ {
		resp, err := pool.Get(ctx, server.URL)
		if err != nil {
			t.Fatalf("Request %d failed: %v", i, err)
		}
		resp.Body.Close()
	}

	stats := pool.GetStats()

	if stats.TotalRequests != 5 {
		t.Errorf("Expected 5 total requests, got %d", stats.TotalRequests)
	}

	if stats.CompletedRequests != 5 {
		t.Errorf("Expected 5 completed requests, got %d", stats.CompletedRequests)
	}

	if stats.FailedRequests != 0 {
		t.Errorf("Expected 0 failed requests, got %d", stats.FailedRequests)
	}

	if stats.AvgResponseTime == 0 {
		t.Error("Expected non-zero average response time")
	}
}

// TestConnectionPoolConcurrent tests concurrent requests
func TestConnectionPoolConcurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}))
	defer server.Close()

	pool := NewConnectionPool(nil)
	defer pool.Close()
	pool.ResetStats()

	ctx := context.Background()
	concurrent := 20

	done := make(chan bool, concurrent)
	for i := 0; i < concurrent; i++ {
		go func() {
			resp, err := pool.Get(ctx, server.URL)
			if err == nil {
				resp.Body.Close()
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < concurrent; i++ {
		<-done
	}

	stats := pool.GetStats()

	if stats.TotalRequests != int64(concurrent) {
		t.Errorf("Expected %d total requests, got %d", concurrent, stats.TotalRequests)
	}

	if stats.CompletedRequests != int64(concurrent) {
		t.Errorf("Expected %d completed requests, got %d", concurrent, stats.CompletedRequests)
	}
}

// TestConnectionPoolFailedRequest tests failed request tracking
func TestConnectionPoolFailedRequest(t *testing.T) {
	pool := NewConnectionPool(nil)
	defer pool.Close()
	pool.ResetStats()

	ctx := context.Background()

	// Make a request to an invalid URL
	_, err := pool.Get(ctx, "http://invalid.example.local:9999")
	if err == nil {
		t.Error("Expected error for invalid request")
	}

	stats := pool.GetStats()

	if stats.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request, got %d", stats.FailedRequests)
	}
}

// TestConnectionPoolCloseIdleConnections tests closing idle connections
func TestConnectionPoolCloseIdleConnections(t *testing.T) {
	pool := NewConnectionPool(nil)

	// Should not panic
	pool.CloseIdleConnections()
}

// TestConnectionPoolSetMaxConnsPerHost tests setting max connections per host
func TestConnectionPoolSetMaxConnsPerHost(t *testing.T) {
	pool := NewConnectionPool(nil)

	pool.SetMaxConnsPerHost(50)

	if pool.config.MaxConnsPerHost != 50 {
		t.Errorf("Expected MaxConnsPerHost 50, got %d", pool.config.MaxConnsPerHost)
	}
}

// TestConnectionPoolSetMaxIdleConns tests setting max idle connections
func TestConnectionPoolSetMaxIdleConns(t *testing.T) {
	pool := NewConnectionPool(nil)

	pool.SetMaxIdleConns(50)

	if pool.config.MaxIdleConns != 50 {
		t.Errorf("Expected MaxIdleConns 50, got %d", pool.config.MaxIdleConns)
	}
}

// TestConnectionPoolTimeout tests request timeout
func TestConnectionPoolTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	config := &PoolConfig{
		Timeout: 50 * time.Millisecond,
	}
	pool := NewConnectionPool(config)
	defer pool.Close()
	pool.ResetStats()

	ctx := context.Background()
	_, err := pool.Get(ctx, server.URL)
	if err == nil {
		t.Error("Expected timeout error")
	}

	stats := pool.GetStats()
	if stats.FailedRequests != 1 {
		t.Errorf("Expected 1 failed request due to timeout, got %d", stats.FailedRequests)
	}
}

// TestConnectionPoolResponseTimeStats tests response time statistics
func TestConnectionPoolResponseTimeStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	pool := NewConnectionPool(nil)
	defer pool.Close()
	pool.ResetStats()

	ctx := context.Background()

	// Make multiple requests
	for i := 0; i < 3; i++ {
		resp, err := pool.Get(ctx, server.URL)
		if err != nil {
			t.Fatalf("Request failed: %v", err)
		}
		resp.Body.Close()
	}

	stats := pool.GetStats()

	if stats.MaxResponseTime == 0 {
		t.Error("Expected non-zero max response time")
	}

	if stats.MinResponseTime == time.Hour {
		t.Error("Expected min response time to be updated")
	}

	if stats.AvgResponseTime == 0 {
		t.Error("Expected non-zero average response time")
	}

	// Min should be <= Avg <= Max
	if stats.MinResponseTime > stats.AvgResponseTime {
		t.Error("Min response time should be <= average")
	}

	if stats.AvgResponseTime > stats.MaxResponseTime {
		t.Error("Average response time should be <= max")
	}
}
