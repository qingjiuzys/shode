package web

import (
	"net/http"
	"sync"
)

// MiddlewareChain manages a chain of middlewares
type MiddlewareChain struct {
	middlewares []Middleware
	mu          sync.RWMutex
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: make([]Middleware, 0),
	}
}

// Add adds a middleware to the chain
func (mc *MiddlewareChain) Add(middleware Middleware) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.middlewares = append(mc.middlewares, middleware)
}

// Apply applies the chain to a handler
func (mc *MiddlewareChain) Apply(handler http.Handler) http.Handler {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return Chain(mc.middlewares...)(handler)
}

// Clear clears all middlewares
func (mc *MiddlewareChain) Clear() {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.middlewares = make([]Middleware, 0)
}

// GetMiddlewares returns all middlewares
func (mc *MiddlewareChain) GetMiddlewares() []Middleware {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.middlewares
}
