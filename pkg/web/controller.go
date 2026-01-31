package web

import (
	"net/http"
	"strings"
	"sync"
)

// Controller represents a controller with routes
type Controller struct {
	BasePath string
	Routes   []*Route
	Middlewares []Middleware
}

// Route represents an HTTP route
type Route struct {
	Method      string
	Path        string
	Handler     http.HandlerFunc
	Middlewares []Middleware
}

// ControllerRegistry manages controllers
type ControllerRegistry struct {
	controllers map[string]*Controller
	mu          sync.RWMutex
}

// NewControllerRegistry creates a new controller registry
func NewControllerRegistry() *ControllerRegistry {
	return &ControllerRegistry{
		controllers: make(map[string]*Controller),
	}
}

// RegisterController registers a controller
func (cr *ControllerRegistry) RegisterController(name string, controller *Controller) {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.controllers[name] = controller
}

// GetController retrieves a controller
func (cr *ControllerRegistry) GetController(name string) (*Controller, bool) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	controller, exists := cr.controllers[name]
	return controller, exists
}

// RegisterRoutes registers all routes from controllers to a mux
func (cr *ControllerRegistry) RegisterRoutes(mux *http.ServeMux) {
	cr.mu.RLock()
	defer cr.mu.RUnlock()

	for _, controller := range cr.controllers {
		for _, route := range controller.Routes {
			// Capture loop variables to avoid closure issues
			route := route
			basePath := controller.BasePath

			fullPath := basePath + route.Path
			if !strings.HasPrefix(fullPath, "/") {
				fullPath = "/" + fullPath
			}

			// Build handler with middlewares
			var handler http.Handler = route.Handler

			// Apply route-specific middlewares
			if len(route.Middlewares) > 0 {
				handler = Apply(handler, route.Middlewares...)
			}
			// Apply controller-level middlewares
			if len(controller.Middlewares) > 0 {
				handler = Apply(handler, controller.Middlewares...)
			}

			mux.HandleFunc(fullPath, func(w http.ResponseWriter, r *http.Request) {
				// Check method
				if route.Method != "*" && r.Method != route.Method {
					http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
					return
				}
				handler.ServeHTTP(w, r)
			})
		}
	}
}
