package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryMiddleware_NoPanic(t *testing.T) {
	mw := NewRecoveryMiddleware(nil)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	called := false
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}

	result := mw.Process(context.Background(), w, req, next)

	if !result {
		t.Error("Recovery middleware should return true when no panic")
	}

	if !called {
		t.Error("Next function should be called when no panic")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRecoveryMiddleware_WithPanic(t *testing.T) {
	mw := NewRecoveryMiddleware(nil)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}

	result := mw.Process(context.Background(), w, req, next)

	if result {
		t.Error("Recovery middleware should return false after panic")
	}

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	body := w.Body.String()
	if body == "" {
		t.Error("Response body should not be empty after panic")
	}
}

func TestRecoveryMiddleware_CustomErrorHandler(t *testing.T) {
	customCalled := false
	errorHandler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, err interface{}) {
		customCalled = true
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("Custom error handler"))
	}
	mw := NewRecoveryMiddleware(errorHandler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	}

	mw.Process(context.Background(), w, req, next)

	if !customCalled {
		t.Error("Custom error handler should be called")
	}

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("Expected status 503, got %d", w.Code)
	}

	body := w.Body.String()
	if body != "Custom error handler" {
		t.Errorf("Expected custom error message, got '%s'", body)
	}
}

func TestRecoveryMiddleware_StackTrace(t *testing.T) {
	errorHandler := func(ctx context.Context, w http.ResponseWriter, r *http.Request, err interface{}) {}
	mw := NewRecoveryMiddleware(errorHandler)

	// By default, StackTrace should be true
	if !mw.StackTrace {
		t.Error("Default StackTrace should be true")
	}

	// Set StackTrace to false
	mw.StackTrace = false
	if mw.StackTrace {
		t.Error("StackTrace should now be false")
	}
}

func TestRecoveryMiddleware_ErrorType(t *testing.T) {
	mw := NewRecoveryMiddleware(nil)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	testError := errors.New("test error")
	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		panic(testError)
	}

	mw.Process(context.Background(), w, req, next)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	body := w.Body.String()
	// Should contain the error message
	if body == "" {
		t.Error("Response body should contain error information")
	}
}

func TestRecoveryMiddleware_StringPanic(t *testing.T) {
	mw := NewRecoveryMiddleware(nil)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	next := func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		panic("string panic")
	}

	mw.Process(context.Background(), w, req, next)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}
}

func TestRecoveryMiddleware_DefaultValues(t *testing.T) {
	mw := NewRecoveryMiddleware(nil)

	if mw.ErrorHandler != nil {
		t.Error("ErrorHandler should be nil when nil is passed to NewRecoveryMiddleware")
	}

	if !mw.StackTrace {
		t.Error("Default StackTrace should be true")
	}
}
