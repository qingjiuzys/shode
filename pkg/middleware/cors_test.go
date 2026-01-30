package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware_Preflight(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"http://example.com"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}
	mw := NewCORSMiddleware(config)

	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	req.Header.Set("Access-Control-Request-Method", "POST")
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
	}

	result := mw.Process(context.Background(), w, req, next)

	if result {
		t.Error("CORS middleware should return false for preflight requests")
	}

	if called {
		t.Error("Next function should not be called for preflight requests")
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") == "" {
		t.Error("Access-Control-Allow-Origin header should be set")
	}

	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Error("Access-Control-Allow-Methods header should be set")
	}
}

func TestCORSMiddleware_SimpleRequest(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST"},
		AllowCredentials: false,
	}
	mw := NewCORSMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://example.com")
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("CORS middleware should return true for simple requests")
	}

	if !called {
		t.Error("Next function should be called for simple requests")
	}

	// Check CORS headers
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("Access-Control-Allow-Origin should be '*', got '%s'", w.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSMiddleware_DisallowedOrigin(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins: []string{"http://example.com"},
	}
	mw := NewCORSMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://evil.com")
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("CORS middleware should return true even for disallowed origins (pass through)")
	}

	if !called {
		t.Error("Next function should be called for disallowed origins")
	}

	// No CORS headers should be set for disallowed origins
	if w.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Error("Access-Control-Allow-Origin should not be set for disallowed origins")
	}
}

func TestCORSMiddleware_NoOrigin(t *testing.T) {
	config := &CORSConfig{
		AllowedOrigins: []string{"*"},
	}
	mw := NewCORSMiddleware(config)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("CORS middleware should return true for requests without Origin header")
	}

	if !called {
		t.Error("Next function should be called for requests without Origin header")
	}
}
