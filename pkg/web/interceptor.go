package web

import (
	"net/http"
)

// Interceptor intercepts HTTP requests and responses
type Interceptor interface {
	Before(w http.ResponseWriter, r *http.Request) error
	After(w http.ResponseWriter, r *http.Request) error
}

// InterceptorFunc is a function-based interceptor
type InterceptorFunc struct {
	BeforeFunc func(w http.ResponseWriter, r *http.Request) error
	AfterFunc  func(w http.ResponseWriter, r *http.Request) error
}

// Before executes before the handler
func (f *InterceptorFunc) Before(w http.ResponseWriter, r *http.Request) error {
	if f.BeforeFunc != nil {
		return f.BeforeFunc(w, r)
	}
	return nil
}

// After executes after the handler
func (f *InterceptorFunc) After(w http.ResponseWriter, r *http.Request) error {
	if f.AfterFunc != nil {
		return f.AfterFunc(w, r)
	}
	return nil
}

// InterceptorMiddleware converts an interceptor to middleware
func InterceptorMiddleware(interceptor Interceptor) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Execute before
			if err := interceptor.Before(w, r); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Execute handler
			next.ServeHTTP(w, r)

			// Execute after
			interceptor.After(w, r)
		})
	}
}
