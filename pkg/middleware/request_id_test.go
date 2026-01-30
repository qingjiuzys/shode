package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestIDMiddleware_GenerateNew(t *testing.T) {
	mw := NewRequestIDMiddleware("X-Request-ID")

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
		// Check context has request ID
		requestID := ctx.Value("request_id")
		if requestID == nil {
			t.Error("Request ID should be set in context")
		}
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("Request ID middleware should return true")
	}

	if !called {
		t.Error("Next function should be called")
	}

	// Check response header
	requestID := w.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Error("X-Request-ID header should be set")
	}
}

func TestRequestIDMiddleware_UseExisting(t *testing.T) {
	mw := NewRequestIDMiddleware("X-Request-ID")

	existingID := "existing-request-123"
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", existingID)
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
		// Check context has the existing request ID
		requestID := ctx.Value("request_id")
		if requestID != existingID {
			t.Errorf("Context should have existing request ID '%s', got '%v'", existingID, requestID)
		}
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("Request ID middleware should return true")
	}

	if !called {
		t.Error("Next function should be called")
	}

	// Check response header has the existing ID
	requestID := w.Header().Get("X-Request-ID")
	if requestID != existingID {
		t.Errorf("X-Request-ID should be '%s', got '%s'", existingID, requestID)
	}
}

func TestRequestIDMiddleware_DefaultHeader(t *testing.T) {
	mw := NewRequestIDMiddleware("")

	if mw.HeaderName != "X-Request-ID" {
		t.Errorf("Default header name should be 'X-Request-ID', got '%s'", mw.HeaderName)
	}
}

func TestGetRequestID(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "test-123")

	requestID := GetRequestID(ctx)
	if requestID != "test-123" {
		t.Errorf("Expected 'test-123', got '%s'", requestID)
	}
}

func TestGetRequestID_NotFound(t *testing.T) {
	ctx := context.Background()

	requestID := GetRequestID(ctx)
	if requestID != "unknown" {
		t.Errorf("Expected 'unknown', got '%s'", requestID)
	}
}

func TestGetRequestIDInt(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "12345")

	requestIDInt := GetRequestIDInt(ctx)
	if requestIDInt != 12345 {
		t.Errorf("Expected 12345, got %d", requestIDInt)
	}
}

func TestGetRequestIDInt_NotANumber(t *testing.T) {
	ctx := context.WithValue(context.Background(), "request_id", "not-a-number")

	requestIDInt := GetRequestIDInt(ctx)
	if requestIDInt != 0 {
		t.Errorf("Expected 0 for non-numeric ID, got %d", requestIDInt)
	}
}
