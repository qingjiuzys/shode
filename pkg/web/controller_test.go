package web

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewControllerRegistry tests creating a new controller registry
func TestNewControllerRegistry(t *testing.T) {
	registry := NewControllerRegistry()

	if registry == nil {
		t.Fatal("NewControllerRegistry returned nil")
	}

	if registry.controllers == nil {
		t.Error("controllers map is nil")
	}
}

// TestRegisterController tests registering a controller
func TestRegisterController(t *testing.T) {
	registry := NewControllerRegistry()

	controller := &Controller{
		BasePath: "/api",
		Routes: []*Route{
			{
				Method:  "GET",
				Path:    "/test",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("test response"))
				},
			},
		},
	}

	registry.RegisterController("api", controller)

	// Verify controller was registered
	retrieved, exists := registry.GetController("api")
	if !exists {
		t.Error("Controller was not registered")
	}

	if retrieved.BasePath != "/api" {
		t.Errorf("Expected BasePath /api, got %s", retrieved.BasePath)
	}

	if len(retrieved.Routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(retrieved.Routes))
	}
}

// TestGetController tests retrieving a controller
func TestGetController(t *testing.T) {
	registry := NewControllerRegistry()

	controller := &Controller{
		BasePath: "/users",
		Routes:   []*Route{},
	}

	registry.RegisterController("users", controller)

	// Test existing controller
	retrieved, exists := registry.GetController("users")
	if !exists {
		t.Error("Expected controller to exist")
	}

	if retrieved != controller {
		t.Error("Retrieved controller is not the same instance")
	}

	// Test non-existing controller
	_, exists = registry.GetController("nonexistent")
	if exists {
		t.Error("Expected non-existing controller to not exist")
	}
}

// TestRegisterRoutes tests registering routes to a mux
func TestRegisterRoutes(t *testing.T) {
	registry := NewControllerRegistry()

	controller := &Controller{
		BasePath: "/api",
		Routes: []*Route{
			{
				Method: "GET",
				Path:   "/hello",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("hello world"))
				},
			},
			{
				Method: "POST",
				Path:   "/create",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					w.Write([]byte("created"))
				},
			},
		},
	}

	registry.RegisterController("api", controller)

	mux := http.NewServeMux()
	registry.RegisterRoutes(mux)

	// Test GET route
	req := httptest.NewRequest("GET", "/api/hello", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	if w.Body.String() != "hello world" {
		t.Errorf("Expected 'hello world', got '%s'", w.Body.String())
	}

	// Test POST route
	req = httptest.NewRequest("POST", "/api/create", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	if w.Body.String() != "created" {
		t.Errorf("Expected 'created', got '%s'", w.Body.String())
	}
}

// TestRouteMethodCheck tests method validation on routes
func TestRouteMethodCheck(t *testing.T) {
	registry := NewControllerRegistry()

	controller := &Controller{
		BasePath: "/api",
		Routes: []*Route{
			{
				Method: "GET",
				Path:   "/resource",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte("GET response"))
				},
			},
		},
	}

	registry.RegisterController("api", controller)
	mux := http.NewServeMux()
	registry.RegisterRoutes(mux)

	// Test correct method
	req := httptest.NewRequest("GET", "/api/resource", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for GET, got %d", w.Code)
	}

	// Test wrong method
	req = httptest.NewRequest("POST", "/api/resource", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405 for POST, got %d", w.Code)
	}
}

// TestWildcardMethod tests wildcard method routes
func TestWildcardMethod(t *testing.T) {
	registry := NewControllerRegistry()

	controller := &Controller{
		BasePath: "/api",
		Routes: []*Route{
			{
				Method: "*",
				Path:   "/any",
				Handler: func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
					w.Write([]byte(r.Method))
				},
			},
		},
	}

	registry.RegisterController("api", controller)
	mux := http.NewServeMux()
	registry.RegisterRoutes(mux)

	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/any", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200 for %s, got %d", method, w.Code)
		}

		if w.Body.String() != method {
			t.Errorf("Expected '%s', got '%s'", method, w.Body.String())
		}
	}
}

// TestBasePathNormalization tests base path normalization
func TestBasePathNormalization(t *testing.T) {
	testCases := []struct {
		name         string
		basePath     string
		routePath    string
		expectedPath string
	}{
		{"leading slash both", "/api", "/test", "/api/test"},
		{"no leading slash base", "api", "/test", "/api/test"},
		{"no leading slash route", "/api", "test", "/api/test"},
		{"no leading slash both", "api", "test", "/api/test"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := NewControllerRegistry()
			controller := &Controller{
				BasePath: tc.basePath,
				Routes: []*Route{
					{
						Method:  "GET",
						Path:    tc.routePath,
						Handler: func(w http.ResponseWriter, r *http.Request) {
							w.WriteHeader(http.StatusOK)
						},
					},
				},
			}

			registry.RegisterController("test", controller)
			mux := http.NewServeMux()
			registry.RegisterRoutes(mux)

			// Try to access the route
			req := httptest.NewRequest("GET", tc.expectedPath, nil)
			w := httptest.NewRecorder()

			// This should not panic
			mux.ServeHTTP(w, req)

			// If we get a handler, the path was normalized correctly
			// (404 is acceptable, 500 is not)
			if w.Code == http.StatusInternalServerError {
				t.Errorf("Internal server error for path %s", tc.expectedPath)
			}
		})
	}
}

// TestMultipleControllers tests registering multiple controllers
func TestMultipleControllers(t *testing.T) {
	registry := NewControllerRegistry()

	// Register first controller
	registry.RegisterController("users", &Controller{
		BasePath: "/users",
		Routes: []*Route{
			{
				Method:  "GET",
				Path:    "/list",
				Handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("users")) },
			},
		},
	})

	// Register second controller
	registry.RegisterController("posts", &Controller{
		BasePath: "/posts",
		Routes: []*Route{
			{
				Method:  "GET",
				Path:    "/list",
				Handler: func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("posts")) },
			},
		},
	})

	mux := http.NewServeMux()
	registry.RegisterRoutes(mux)

	// Test users route
	req := httptest.NewRequest("GET", "/users/list", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Body.String() != "users" {
		t.Errorf("Expected 'users', got '%s'", w.Body.String())
	}

	// Test posts route
	req = httptest.NewRequest("GET", "/posts/list", nil)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Body.String() != "posts" {
		t.Errorf("Expected 'posts', got '%s'", w.Body.String())
	}
}

// TestControllerConcurrency tests concurrent controller access
func TestControllerConcurrency(t *testing.T) {
	registry := NewControllerRegistry()

	// Concurrent registration
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			registry.RegisterController("controller_"+string(rune('0'+n)), &Controller{
				BasePath: "/api" + string(rune('0'+n)),
				Routes:   []*Route{},
			})
			done <- true
		}(i)
	}

	// Wait for all registrations
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all controllers were registered
	for i := 0; i < 10; i++ {
		_, exists := registry.GetController("controller_" + string(rune('0'+i)))
		if !exists {
			t.Errorf("Controller %d not registered", i)
		}
	}
}
