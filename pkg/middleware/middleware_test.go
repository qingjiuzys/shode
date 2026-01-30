package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManager_Use(t *testing.T) {
	m := NewManager()
	mw := NewBaseMiddleware("test", 100, nil)

	m.Use(mw)

	list := m.List()
	if len(list) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(list))
	}

	if list[0] != "test" {
		t.Errorf("Expected middleware name 'test', got '%s'", list[0])
	}
}

func TestManager_Remove(t *testing.T) {
	m := NewManager()
	mw := NewBaseMiddleware("test", 100, nil)
	m.Use(mw)

	m.Remove("test")

	list := m.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 middlewares after removal, got %d", len(list))
	}
}

func TestManager_Clear(t *testing.T) {
	m := NewManager()
	m.Use(NewBaseMiddleware("test1", 100, nil))
	m.Use(NewBaseMiddleware("test2", 200, nil))

	m.Clear()

	list := m.List()
	if len(list) != 0 {
		t.Errorf("Expected 0 middlewares after clear, got %d", len(list))
	}
}

func TestManager_PrioritySort(t *testing.T) {
	m := NewManager()
	m.Use(NewBaseMiddleware("high", 300, nil))
	m.Use(NewBaseMiddleware("low", 100, nil))
	m.Use(NewBaseMiddleware("medium", 200, nil))

	list := m.List()
	if len(list) != 3 {
		t.Errorf("Expected 3 middlewares, got %d", len(list))
	}

	// Should be sorted: low, medium, high
	if list[0] != "low" {
		t.Errorf("Expected first middleware to be 'low', got '%s'", list[0])
	}
	if list[1] != "medium" {
		t.Errorf("Expected second middleware to be 'medium', got '%s'", list[1])
	}
	if list[2] != "high" {
		t.Errorf("Expected third middleware to be 'high', got '%s'", list[2])
	}
}

func TestManager_Execute(t *testing.T) {
	m := NewManager()

	executed := false
	finalHandler := func(w http.ResponseWriter, r *http.Request) {
		executed = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	m.Use(NewBaseMiddleware("test", 100, func(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
		next(ctx, w, r)
		return true
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	m.Execute(context.Background(), w, req, finalHandler)

	if !executed {
		t.Error("Final handler was not executed")
	}

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestManager_Execute_WithSkipCondition(t *testing.T) {
	m := NewManager()

	executed := false
	finalHandler := func(w http.ResponseWriter, r *http.Request) {
		executed = true
		w.WriteHeader(http.StatusOK)
	}

	middlewareExecuted := false
	mw := NewBaseMiddleware("test", 100, func(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
		middlewareExecuted = true
		next(ctx, w, r)
		return true
	})
	mw.SetSkipFunc(func(r *http.Request) bool {
		return r.URL.Path == "/skip"
	})

	m.Use(mw)

	req := httptest.NewRequest("GET", "/skip", nil)
	w := httptest.NewRecorder()

	m.Execute(context.Background(), w, req, finalHandler)

	// When middleware is skipped, its Process function should NOT be called
	if middlewareExecuted {
		t.Error("Middleware Process should not be executed when skip condition returns true")
	}

	// But the final handler should still be executed
	if !executed {
		t.Error("Final handler should still be executed when middleware is skipped")
	}
}

func TestManager_Execute_WithInterruption(t *testing.T) {
	m := NewManager()

	executed := false
	finalHandler := func(w http.ResponseWriter, r *http.Request) {
		executed = true
	}

	m.Use(NewBaseMiddleware("blocking", 100, func(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
		w.WriteHeader(http.StatusForbidden)
		return false // Interrupt the chain
	}))

	m.Use(NewBaseMiddleware("after", 200, func(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
		t.Error("This middleware should not be executed")
		return true
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	m.Execute(context.Background(), w, req, finalHandler)

	if executed {
		t.Error("Final handler should not be executed when chain is interrupted")
	}

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}
}

func TestChain(t *testing.T) {
	mw1 := NewBaseMiddleware("first", 100, nil)
	mw2 := NewBaseMiddleware("second", 200, nil)

	m := Chain(mw1, mw2)

	list := m.List()
	if len(list) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(list))
	}
}

func TestBaseMiddleware_SkipCondition(t *testing.T) {
	mw := NewBaseMiddleware("test", 100, nil)

	if mw.SkipCondition(nil) {
		t.Error("Default skip condition should return false")
	}

	mw.SetSkipFunc(func(r *http.Request) bool {
		return true
	})

	if !mw.SkipCondition(nil) {
		t.Error("Custom skip condition should return true")
	}
}
