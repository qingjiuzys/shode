package middleware

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLoggingMiddleware_Basic(t *testing.T) {
	var logs []string
	config := &LoggingConfig{
		OutputWriter: func(format string, args ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	}
	mw := NewLoggingMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("Logging middleware should return true")
	}

	if len(logs) != 2 {
		t.Errorf("Expected 2 log entries (request + response), got %d", len(logs))
	}

	// Check request log
	if !strings.Contains(logs[0], "GET") || !strings.Contains(logs[0], "/test") {
		t.Errorf("Request log should contain method and path, got: %s", logs[0])
	}

	// Check response log
	if !strings.Contains(logs[1], "Status: 200") {
		t.Errorf("Response log should contain status code, got: %s", logs[1])
	}
}

func TestLoggingMiddleware_WithQuery(t *testing.T) {
	var logs []string
	config := &LoggingConfig{
		OutputWriter: func(format string, args ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	}
	mw := NewLoggingMiddleware(config)

	req := httptest.NewRequest("GET", "/test?foo=bar&baz=qux", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	mw.Process(context.Background(), w, req, next)

	if !strings.Contains(logs[0], "?foo=bar&baz=qux") {
		t.Errorf("Request log should contain query string, got: %s", logs[0])
	}
}

func TestLoggingMiddleware_WithHeaders(t *testing.T) {
	var logs []string
	config := &LoggingConfig{
		LogHeaders: true,
		OutputWriter: func(format string, args ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	}
	mw := NewLoggingMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("Accept", "application/json")
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	mw.Process(context.Background(), w, req, next)

	if !strings.Contains(logs[0], "Headers:") {
		t.Errorf("Request log should contain headers when LogHeaders is true, got: %s", logs[0])
	}
}

func TestLoggingMiddleware_WithBody(t *testing.T) {
	var logs []string
	config := &LoggingConfig{
		LogBody: true,
		OutputWriter: func(format string, args ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	}
	mw := NewLoggingMiddleware(config)

	req := httptest.NewRequest("POST", "/test", strings.NewReader("test body"))
	req.Header.Set("Content-Length", "9")
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	mw.Process(context.Background(), w, req, next)

	// Should log body length for non-GET requests
	if !strings.Contains(logs[0], "Body Length:") {
		t.Errorf("Request log should contain body length when LogBody is true, got: %s", logs[0])
	}
}

func TestLoggingMiddleware_ResponseSize(t *testing.T) {
	var logs []string
	config := &LoggingConfig{
		LogResponse: true,
		OutputWriter: func(format string, args ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	}
	mw := NewLoggingMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	}

	mw.Process(context.Background(), w, req, next)

	// Check response log contains size
	if !strings.Contains(logs[1], "Size:") {
		t.Errorf("Response log should contain size when LogResponse is true, got: %s", logs[1])
	}
}

func TestLoggingMiddleware_Non200Status(t *testing.T) {
	var logs []string
	config := &LoggingConfig{
		OutputWriter: func(format string, args ...interface{}) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	}
	mw := NewLoggingMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}

	mw.Process(context.Background(), w, req, next)

	if !strings.Contains(logs[1], "Status: 404") {
		t.Errorf("Response log should contain 404 status, got: %s", logs[1])
	}
}

func TestLoggingMiddleware_DefaultConfig(t *testing.T) {
	mw := NewLoggingMiddleware(nil)

	if mw.config == nil {
		t.Error("Config should not be nil when nil is passed to NewLoggingMiddleware")
	}

	if mw.config.LogHeaders {
		t.Error("Default LogHeaders should be false")
	}

	if mw.config.LogBody {
		t.Error("Default LogBody should be false")
	}

	if mw.config.LogResponse {
		t.Error("Default LogResponse should be false")
	}
}

func TestResponseWriterWrapper(t *testing.T) {
	underlying := httptest.NewRecorder()
	wrapped := &responseWriter{
		ResponseWriter: underlying,
		status:         200,
		size:           0,
	}

	// Test WriteHeader
	wrapped.WriteHeader(http.StatusCreated)
	if wrapped.status != http.StatusCreated {
		t.Errorf("Expected status %d, got %d", http.StatusCreated, wrapped.status)
	}

	// Test Write
	n, err := wrapped.Write([]byte("test"))
	if err != nil {
		t.Errorf("Write should not error, got %v", err)
	}
	if n != 4 {
		t.Errorf("Expected to write 4 bytes, wrote %d", n)
	}
	if wrapped.size != 4 {
		t.Errorf("Expected size 4, got %d", wrapped.size)
	}

	// Test Hijack on non-hijackable writer
	_, _, err = wrapped.Hijack()
	if err == nil {
		t.Error("Hijack should return error for non-hijackable writer")
	}
}
