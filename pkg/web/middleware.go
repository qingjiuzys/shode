package web

import (
	"net/http"
)

// Middleware represents an HTTP middleware function
type Middleware func(http.Handler) http.Handler

// MiddlewareFunc is a function that processes HTTP requests
type MiddlewareFunc func(w http.ResponseWriter, r *http.Request) error

// Chain chains multiple middlewares together
func Chain(middlewares ...Middleware) Middleware {
	return func(next http.Handler) http.Handler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			next = middlewares[i](next)
		}
		return next
	}
}

// Apply applies middleware to a handler
func Apply(handler http.Handler, middlewares ...Middleware) http.Handler {
	return Chain(middlewares...)(handler)
}
