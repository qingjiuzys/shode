package stdlib

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/cache"
	"gitee.com/com_818cloud/shode/pkg/config"
	"gitee.com/com_818cloud/shode/pkg/database"
	"gitee.com/com_818cloud/shode/pkg/ioc"
	"gitee.com/com_818cloud/shode/pkg/parser"
	"gitee.com/com_818cloud/shode/pkg/types"
	"gitee.com/com_818cloud/shode/pkg/web"
)

// HTTPRequestContext holds HTTP request information
type HTTPRequestContext struct {
	Method      string
	Path        string
	QueryParams map[string]string
	Headers     map[string]string
	Body        string
	Response    *HTTPResponseContext
	mu          sync.RWMutex
}

// HTTPResponseContext holds HTTP response information
type HTTPResponseContext struct {
	Status  int
	Body    string
	Headers map[string]string
	mu      sync.RWMutex
}

// StdLib provides built-in functions to replace external commands
type StdLib struct {
	httpServer *httpServer
	httpMu     sync.Mutex
	// Request context storage (per-goroutine)
	requestContexts sync.Map // map[goroutineID]*HTTPRequestContext
	// Cache instance
	cache *cache.Cache
	// Database manager
	dbManager *database.DatabaseManager
	// IoC container
	iocContainer *ioc.Container
	// Config manager
	configManager *config.ConfigManager
	// Execution engine factory (to avoid circular dependency)
	engineFactory func() interface{} // Returns *engine.ExecutionEngine
	// Files manager
	filesManager *FilesManager
	// System manager
	systemManager *SystemManager
	// Network manager
	networkManager *NetworkManager
	// Archive manager
	archiveManager *ArchiveManager
}

// FilesManager handles file operations
type FilesManager struct{}

// SystemManager handles system operations
type SystemManager struct{}

// NetworkManager handles network operations
type NetworkManager struct{}

// ArchiveManager handles compression/archive operations
type ArchiveManager struct{}

// routeHandler represents a route handler
type routeHandler struct {
	method      string // HTTP method (GET, POST, PUT, DELETE, PATCH, "*" for all)
	path        string
	handlerType string // "function" or "script"
	handlerName string // function name or script content
}

// httpServer represents an HTTP server instance
type httpServer struct {
	server      *http.Server
	mux         *http.ServeMux
	routes      map[string]*routeHandler // routeKey (method:path) -> handler
	isRunning   bool
	middlewares []web.Middleware // Global middlewares
	mu          sync.RWMutex
}

// New creates a new standard library instance
func New() *StdLib {
	return &StdLib{
		cache:          cache.NewCache(),
		dbManager:      database.NewDatabaseManager(),
		iocContainer:   ioc.NewContainer(),
		configManager:  config.NewConfigManager(),
		filesManager:   &FilesManager{},
		systemManager:  &SystemManager{},
		networkManager: &NetworkManager{},
		archiveManager: &ArchiveManager{},
	}
}

// FileSystem functions

// ReadFile reads the contents of a file (replaces 'cat')
func (sl *StdLib) ReadFile(filename string) (string, error) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return string(content), nil
}

// WriteFile writes content to a file (replaces echo > file)
func (sl *StdLib) WriteFile(filename, content string) error {
	return ioutil.WriteFile(filename, []byte(content), 0644)
}

// ListFiles lists files in a directory (replaces 'ls')
func (sl *StdLib) ListFiles(dirpath string) ([]string, error) {
	files, err := ioutil.ReadDir(dirpath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory %s: %v", dirpath, err)
	}

	var fileNames []string
	for _, file := range files {
		fileNames = append(fileNames, file.Name())
	}
	return fileNames, nil
}

// FileExists checks if a file exists (replaces test -f)
func (sl *StdLib) FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// String functions

// Contains checks if a string contains another string (replaces grep)
func (sl *StdLib) Contains(haystack, needle string) bool {
	return strings.Contains(haystack, needle)
}

// Replace replaces all occurrences of old with new in a string (replaces sed)
func (sl *StdLib) Replace(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// ToUpper converts string to uppercase (replaces tr '[:lower:]' '[:upper:]')
func (sl *StdLib) ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower converts string to lowercase (replaces tr '[:upper:]' '[:lower:]')
func (sl *StdLib) ToLower(s string) string {
	return strings.ToLower(s)
}

// Trim removes leading and trailing whitespace (replaces sed trimming)
func (sl *StdLib) Trim(s string) string {
	return strings.TrimSpace(s)
}

// Environment functions

// GetEnv gets an environment variable (replaces $VAR)
func (sl *StdLib) GetEnv(key string) string {
	return os.Getenv(key)
}

// SetEnv sets an environment variable (replaces export)
func (sl *StdLib) SetEnv(key, value string) error {
	return os.Setenv(key, value)
}

// WorkingDir gets the current working directory (replaces pwd)
func (sl *StdLib) WorkingDir() (string, error) {
	return os.Getwd()
}

// ChangeDir changes the current directory (replaces cd)
func (sl *StdLib) ChangeDir(dirpath string) error {
	return os.Chdir(dirpath)
}

// Utility functions

// Print outputs text to stdout (replaces echo)
func (sl *StdLib) Print(text string) {
	fmt.Print(text)
}

// Println outputs text with newline to stdout (replaces echo)
func (sl *StdLib) Println(text string) {
	fmt.Println(text)
}

// Error outputs text to stderr (replaces echo >&2)
func (sl *StdLib) Error(text string) {
	fmt.Fprint(os.Stderr, text)
}

// Errorln outputs text with newline to stderr (replaces echo >&2)
func (sl *StdLib) Errorln(text string) {
	fmt.Fprintln(os.Stderr, text)
}

// HTTP Server functions

// StartHTTPServer starts an HTTP server on the specified port
// Usage: StartHTTPServer "9188"
func (sl *StdLib) StartHTTPServer(port string) error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	// Parse port
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return fmt.Errorf("invalid port: %s", port)
	}

	// Check if server already exists and is running
	if sl.httpServer != nil && sl.httpServer.isRunning {
		return fmt.Errorf("HTTP server is already running")
	}

	// Create new server
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", portNum),
		Handler: mux,
	}

	sl.httpServer = &httpServer{
		server:    server,
		mux:       mux,
		routes:    make(map[string]*routeHandler),
		isRunning: false,
	}

	// Start server in goroutine
	go func() {
		sl.httpServer.mu.Lock()
		sl.httpServer.isRunning = true
		sl.httpServer.mu.Unlock()

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}

		sl.httpServer.mu.Lock()
		sl.httpServer.isRunning = false
		sl.httpServer.mu.Unlock()
	}()

	return nil
}

// RegisterRoute registers a route handler (deprecated, use RegisterHTTPRoute)
// Usage: RegisterRoute "/" "handleRoot"
func (sl *StdLib) RegisterRoute(path, handlerName string) error {
	return sl.RegisterHTTPRoute("*", path, "function", handlerName)
}

// RegisterHTTPRoute registers an HTTP route with method, path, handler type and handler
// Usage: RegisterHTTPRoute "GET" "/api/users" "function" "handleGetUsers"
//
//	RegisterHTTPRoute "POST" "/api/users" "script" "SetHTTPResponse 201 'Created'"
func (sl *StdLib) RegisterHTTPRoute(method, path, handlerType, handler string) error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer == nil {
		return fmt.Errorf("HTTP server not started. Call StartHTTPServer first")
	}

	// Normalize method to uppercase
	method = strings.ToUpper(method)
	if method == "" {
		method = "*"
	}

	// Validate handler type
	if handlerType != "function" && handlerType != "script" {
		return fmt.Errorf("invalid handler type: %s (must be 'function' or 'script')", handlerType)
	}

	sl.httpServer.mu.Lock()
	defer sl.httpServer.mu.Unlock()

	// Create route key: method:path
	routeKey := fmt.Sprintf("%s:%s", method, path)

	// Store the handler
	sl.httpServer.routes[routeKey] = &routeHandler{
		method:      method,
		path:        path,
		handlerType: handlerType,
		handlerName: handler,
	}

	// Check if path is already registered
	// If not, register a method-aware handler
	pathKey := fmt.Sprintf("path:%s", path)
	if _, exists := sl.httpServer.routes[pathKey]; !exists {
		// Register the route with method checking
		sl.httpServer.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			sl.httpServer.mu.RLock()
			defer sl.httpServer.mu.RUnlock()

			// Check for exact method match
			exactKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
			handler, exactExists := sl.httpServer.routes[exactKey]

			// Check for wildcard method match
			wildcardKey := fmt.Sprintf("*:%s", r.URL.Path)
			wildcardHandler, wildcardExists := sl.httpServer.routes[wildcardKey]

			var selectedHandler *routeHandler
			if exactExists {
				selectedHandler = handler
			} else if wildcardExists {
				selectedHandler = wildcardHandler
			}

			if selectedHandler == nil {
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "Method Not Allowed\n")
				return
			}

			// Create and store request context
			reqCtx := sl.createRequestContext(r)
			goroutineID := fmt.Sprintf("%p", &r) // Use request pointer as unique ID
			sl.requestContexts.Store(goroutineID, reqCtx)
			// Also store as "current" for getCurrentRequestContext to find
			sl.requestContexts.Store("current", reqCtx)
			defer func() {
				sl.requestContexts.Delete(goroutineID)
				sl.requestContexts.Delete("current")
			}()

			// Execute handler based on type
			if selectedHandler.handlerType == "function" {
				// Execute function handler using reflection to avoid circular dependency
				if sl.engineFactory != nil {
					engineInterface := sl.engineFactory()
					if engineInterface != nil {
						// Use reflection to call ExecuteCommand method
						engineValue := reflect.ValueOf(engineInterface)
						executeCommandMethod := engineValue.MethodByName("ExecuteCommand")

						if executeCommandMethod.IsValid() {
							ctx := context.Background()

							// Create a command node to call the function
							cmdNode := &types.CommandNode{
								Pos:  types.Position{Line: 0, Column: 0, Offset: 0},
								Name: selectedHandler.handlerName,
								Args: []string{}, // Function arguments would come from query params if needed
							}

							// Call ExecuteCommand using reflection
							ctxValue := reflect.ValueOf(ctx)
							cmdNodeValue := reflect.ValueOf(cmdNode)
							results := executeCommandMethod.Call([]reflect.Value{ctxValue, cmdNodeValue})

							// Check for errors
							if len(results) == 2 && !results[1].IsNil() {
								// Error occurred
								err := results[1].Interface().(error)
								reqCtx.Response.mu.Lock()
								reqCtx.Response.Status = http.StatusInternalServerError
								reqCtx.Response.Body = fmt.Sprintf("Error executing handler: %v", err)
								reqCtx.Response.mu.Unlock()
							} else {
								// Function executed successfully
								// Wait a moment for SetHTTPResponse to be called
								// (in case it's called asynchronously or needs time)
								time.Sleep(10 * time.Millisecond)

								// The function should have called SetHTTPResponse
								// Check if response was set (after execution)
								reqCtx.Response.mu.RLock()
								statusSet := reqCtx.Response.Status != 0
								bodySet := reqCtx.Response.Body != ""
								reqCtx.Response.mu.RUnlock()

								// If response wasn't set by function, check command result output
								if !statusSet || !bodySet {
									if len(results) >= 1 && !results[0].IsNil() {
										resultValue := results[0].Interface()
										resultReflect := reflect.ValueOf(resultValue)
										if resultReflect.Kind() == reflect.Ptr {
											resultReflect = resultReflect.Elem()
										}

										outputField := resultReflect.FieldByName("Output")
										if outputField.IsValid() && outputField.Kind() == reflect.String {
											output := outputField.String()
											if output != "" && !bodySet {
												reqCtx.Response.mu.Lock()
												reqCtx.Response.Body = output
												reqCtx.Response.mu.Unlock()
											}
										}
									}
								}
							}
						}
					}
				}

				// If function didn't set response, set default
				reqCtx.Response.mu.RLock()
				if reqCtx.Response.Status == 0 {
					reqCtx.Response.mu.RUnlock()
					reqCtx.Response.mu.Lock()
					reqCtx.Response.Status = http.StatusOK
					reqCtx.Response.mu.Unlock()
				} else {
					reqCtx.Response.mu.RUnlock()
				}

				reqCtx.Response.mu.RLock()
				if reqCtx.Response.Body == "" {
					reqCtx.Response.mu.RUnlock()
					reqCtx.Response.mu.Lock()
					reqCtx.Response.Body = fmt.Sprintf("Handler function: %s (call SetHTTPResponse in function)", selectedHandler.handlerName)
					reqCtx.Response.mu.Unlock()
				} else {
					reqCtx.Response.mu.RUnlock()
				}
			} else {
				// Execute script handler
				if sl.engineFactory != nil {
					engineInterface := sl.engineFactory()
					if engineInterface != nil {
						// Parse the script
						p := parser.NewSimpleParser()
						script, err := p.ParseString(selectedHandler.handlerName)
						if err == nil {
							// Use reflection to call Execute method
							engineValue := reflect.ValueOf(engineInterface)
							executeMethod := engineValue.MethodByName("Execute")

							if executeMethod.IsValid() {
								ctx := context.Background()
								ctxValue := reflect.ValueOf(ctx)
								scriptValue := reflect.ValueOf(script)
								results := executeMethod.Call([]reflect.Value{ctxValue, scriptValue})

								if len(results) >= 1 && !results[0].IsNil() {
									resultValue := results[0].Interface()
									resultReflect := reflect.ValueOf(resultValue)
									if resultReflect.Kind() == reflect.Ptr {
										resultReflect = resultReflect.Elem()
									}

									outputField := resultReflect.FieldByName("Output")
									if outputField.IsValid() && outputField.Kind() == reflect.String {
										output := outputField.String()
										if output != "" {
											reqCtx.Response.mu.Lock()
											if reqCtx.Response.Status == 0 {
												reqCtx.Response.Status = http.StatusOK
											}
											if reqCtx.Response.Body == "" {
												reqCtx.Response.Body = output
											}
											reqCtx.Response.mu.Unlock()
										}
									}
								}
							}
						}
					}
				}

				// Set default if script didn't set response
				reqCtx.Response.mu.RLock()
				if reqCtx.Response.Status == 0 {
					reqCtx.Response.mu.RUnlock()
					reqCtx.Response.mu.Lock()
					reqCtx.Response.Status = http.StatusOK
					reqCtx.Response.mu.Unlock()
				} else {
					reqCtx.Response.mu.RUnlock()
				}

				reqCtx.Response.mu.RLock()
				if reqCtx.Response.Body == "" {
					reqCtx.Response.mu.RUnlock()
					reqCtx.Response.mu.Lock()
					reqCtx.Response.Body = fmt.Sprintf("Handler script: %s", selectedHandler.handlerName)
					reqCtx.Response.mu.Unlock()
				} else {
					reqCtx.Response.mu.RUnlock()
				}
			}

			// Write response
			reqCtx.Response.mu.RLock()
			status := reqCtx.Response.Status
			body := reqCtx.Response.Body
			headers := reqCtx.Response.Headers
			reqCtx.Response.mu.RUnlock()

			// Set headers
			for k, v := range headers {
				w.Header().Set(k, v)
			}
			if len(headers) == 0 {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}

			w.WriteHeader(status)
			if body != "" {
				fmt.Fprintf(w, "%s", body)
			} else {
				fmt.Fprintf(w, "Handler: %s (type: %s, method: %s)\n",
					selectedHandler.handlerName, selectedHandler.handlerType, selectedHandler.method)
			}
		})

		// Mark path as registered
		sl.httpServer.routes[pathKey] = &routeHandler{path: path}
	}

	return nil
}

// RegisterRouteWithResponse registers a route with a direct response
// Usage: RegisterRouteWithResponse "/" "hello world"
func (sl *StdLib) RegisterRouteWithResponse(path, response string) error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer == nil {
		return fmt.Errorf("HTTP server not started. Call StartHTTPServer first")
	}

	sl.httpServer.mu.Lock()
	defer sl.httpServer.mu.Unlock()

	// Register the route with a direct response
	sl.httpServer.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "%s\n", response)
	})

	return nil
}

// StopHTTPServer stops the HTTP server gracefully
func (sl *StdLib) StopHTTPServer() error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer == nil {
		return fmt.Errorf("HTTP server is not running")
	}

	sl.httpServer.mu.Lock()
	running := sl.httpServer.isRunning
	server := sl.httpServer.server
	sl.httpServer.mu.Unlock()

	if !running {
		return fmt.Errorf("HTTP server is not running")
	}

	// Use Shutdown for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		// If shutdown fails, try force close
		server.Close()
		return fmt.Errorf("failed to stop HTTP server: %v", err)
	}

	sl.httpServer.mu.Lock()
	sl.httpServer.isRunning = false
	sl.httpServer.mu.Unlock()

	return nil
}

// IsHTTPServerRunning checks if the HTTP server is running
func (sl *StdLib) IsHTTPServerRunning() bool {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer == nil {
		return false
	}

	sl.httpServer.mu.RLock()
	defer sl.httpServer.mu.RUnlock()

	return sl.httpServer.isRunning
}

// createRequestContext creates a request context from an HTTP request
func (sl *StdLib) createRequestContext(r *http.Request) *HTTPRequestContext {
	// Read body
	bodyBytes, _ := ioutil.ReadAll(r.Body)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for potential re-read

	// Parse query parameters
	queryParams := make(map[string]string)
	rawQuery := r.URL.RawQuery
	fmt.Printf("[DEBUG] createRequestContext: RawQuery=%s, Path=%s\n", rawQuery, r.URL.Path)
	for k, v := range r.URL.Query() {
		if len(v) > 0 {
			queryParams[k] = v[0]
			fmt.Printf("[DEBUG] createRequestContext: parsed param %s=%s\n", k, v[0])
		}
	}
	// Debug: ensure query params are parsed correctly
	if rawQuery != "" && len(queryParams) == 0 {
		fmt.Printf("[DEBUG] createRequestContext: RawQuery not empty but no params parsed, trying manual parse\n")
		// Try manual parsing as fallback
		parts := strings.Split(rawQuery, "&")
		for _, part := range parts {
			if idx := strings.Index(part, "="); idx > 0 {
				key := part[:idx]
				value := part[idx+1:]
				// URL decode if needed
				if decoded, err := url.QueryUnescape(value); err == nil {
					value = decoded
				}
				queryParams[key] = value
				fmt.Printf("[DEBUG] createRequestContext: manually parsed param %s=%s\n", key, value)
			}
		}
	}
	fmt.Printf("[DEBUG] createRequestContext: final queryParams=%v\n", queryParams)

	// Parse headers
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	return &HTTPRequestContext{
		Method:      r.Method,
		Path:        r.URL.Path,
		QueryParams: queryParams,
		Headers:     headers,
		Body:        string(bodyBytes),
		Response: &HTTPResponseContext{
			Status:  0,
			Body:    "",
			Headers: make(map[string]string),
		},
	}
}

// getCurrentRequestContext gets the current request context for the calling goroutine
func (sl *StdLib) getCurrentRequestContext() *HTTPRequestContext {
	// First try to get "current" context (set during request handling)
	if ctx, ok := sl.requestContexts.Load("current"); ok {
		if httpCtx, ok := ctx.(*HTTPRequestContext); ok {
			return httpCtx
		}
	}

	// Fallback: find any context (for backward compatibility)
	var foundCtx *HTTPRequestContext
	sl.requestContexts.Range(func(key, value interface{}) bool {
		// Skip "current" key as we already checked it
		if keyStr, ok := key.(string); ok && keyStr == "current" {
			return true
		}
		if ctx, ok := value.(*HTTPRequestContext); ok {
			foundCtx = ctx
			return false // Stop at first match
		}
		return true
	})
	return foundCtx
}

// GetHTTPMethod returns the HTTP method of the current request
func (sl *StdLib) GetHTTPMethod() string {
	ctx := sl.getCurrentRequestContext()
	if ctx == nil {
		return ""
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.Method
}

// GetHTTPPath returns the HTTP path of the current request
func (sl *StdLib) GetHTTPPath() string {
	ctx := sl.getCurrentRequestContext()
	if ctx == nil {
		return ""
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.Path
}

// GetHTTPQuery returns a query parameter value
func (sl *StdLib) GetHTTPQuery(key string) string {
	fmt.Printf("[DEBUG] GetHTTPQuery called with key: %s\n", key)
	ctx := sl.getCurrentRequestContext()
	fmt.Printf("[DEBUG] GetHTTPQuery: getCurrentRequestContext returned: %v\n", ctx != nil)
	if ctx == nil {
		// Fallback: try to find any context
		var foundCtx *HTTPRequestContext
		sl.requestContexts.Range(func(mapKey, value interface{}) bool {
			if httpCtx, ok := value.(*HTTPRequestContext); ok {
				foundCtx = httpCtx
				return false // Stop at first match
			}
			return true
		})
		if foundCtx != nil {
			foundCtx.mu.RLock()
			defer foundCtx.mu.RUnlock()
			// Debug: print all query params
			if len(foundCtx.QueryParams) > 0 {
				fmt.Printf("[DEBUG] GetHTTPQuery: found context with %d params: %v\n", len(foundCtx.QueryParams), foundCtx.QueryParams)
			}
			return foundCtx.QueryParams[key]
		}
		fmt.Printf("[DEBUG] GetHTTPQuery: no context found for key: %s\n", key)
		return ""
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	// Debug: print all query params
	if len(ctx.QueryParams) > 0 {
		fmt.Printf("[DEBUG] GetHTTPQuery: context has %d params: %v, looking for: %s\n", len(ctx.QueryParams), ctx.QueryParams, key)
	}
	result := ctx.QueryParams[key]
	fmt.Printf("[DEBUG] GetHTTPQuery: returning value: '%s' for key: %s\n", result, key)
	return result
}

// GetHTTPHeader returns a request header value
func (sl *StdLib) GetHTTPHeader(name string) string {
	ctx := sl.getCurrentRequestContext()
	if ctx == nil {
		return ""
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.Headers[name]
}

// GetHTTPBody returns the request body
func (sl *StdLib) GetHTTPBody() string {
	ctx := sl.getCurrentRequestContext()
	if ctx == nil {
		return ""
	}
	ctx.mu.RLock()
	defer ctx.mu.RUnlock()
	return ctx.Body
}

// SetHTTPResponse sets the HTTP response status and body
func (sl *StdLib) SetHTTPResponse(status int, body string) {
	ctx := sl.getCurrentRequestContext()
	if ctx == nil {
		// Fallback: try to find any context
		var foundCtx *HTTPRequestContext
		sl.requestContexts.Range(func(key, value interface{}) bool {
			if httpCtx, ok := value.(*HTTPRequestContext); ok {
				foundCtx = httpCtx
				return false // Stop at first match
			}
			return true
		})
		if foundCtx != nil {
			foundCtx.Response.mu.Lock()
			foundCtx.Response.Status = status
			foundCtx.Response.Body = body
			foundCtx.Response.mu.Unlock()
		}
		return
	}
	ctx.Response.mu.Lock()
	defer ctx.Response.mu.Unlock()
	ctx.Response.Status = status
	ctx.Response.Body = body
}

// SetHTTPHeader sets a response header
func (sl *StdLib) SetHTTPHeader(name, value string) {
	ctx := sl.getCurrentRequestContext()
	if ctx == nil {
		return
	}
	ctx.Response.mu.Lock()
	defer ctx.Response.mu.Unlock()
	if ctx.Response.Headers == nil {
		ctx.Response.Headers = make(map[string]string)
	}
	ctx.Response.Headers[name] = value
}

// Cache functions

// SetCache sets a value in the cache with optional TTL
func (sl *StdLib) SetCache(key, value string, ttlSeconds int) {
	sl.cache.Set(key, value, ttlSeconds)
}

// GetCache retrieves a value from the cache
func (sl *StdLib) GetCache(key string) (string, bool) {
	return sl.cache.Get(key)
}

// DeleteCache removes a key from the cache
func (sl *StdLib) DeleteCache(key string) {
	sl.cache.Delete(key)
}

// ClearCache removes all entries from the cache
func (sl *StdLib) ClearCache() {
	sl.cache.Clear()
}

// CacheExists checks if a key exists in the cache
func (sl *StdLib) CacheExists(key string) bool {
	return sl.cache.Exists(key)
}

// GetCacheTTL returns the remaining TTL in seconds for a key
func (sl *StdLib) GetCacheTTL(key string) int {
	return sl.cache.GetTTL(key)
}

// SetCacheBatch sets multiple key-value pairs at once
func (sl *StdLib) SetCacheBatch(keyValues map[string]string, ttlSeconds int) {
	sl.cache.SetBatch(keyValues, ttlSeconds)
}

// GetCacheKeys returns all keys matching a pattern
func (sl *StdLib) GetCacheKeys(pattern string) []string {
	return sl.cache.GetKeys(pattern)
}

// Database functions

// ConnectDB connects to a database
func (sl *StdLib) ConnectDB(dbType, dsn string) error {
	return sl.dbManager.Connect(dbType, dsn)
}

// CloseDB closes the database connection
func (sl *StdLib) CloseDB() error {
	return sl.dbManager.Close()
}

// IsDBConnected checks if the database is connected
func (sl *StdLib) IsDBConnected() bool {
	return sl.dbManager.IsConnected()
}

// QueryDB executes a SELECT query
func (sl *StdLib) QueryDB(sql string, args ...string) (*database.QueryResult, error) {
	// Convert string args to interface{}
	interfaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		interfaceArgs[i] = arg
	}
	return sl.dbManager.Query(sql, interfaceArgs...)
}

// QueryRowDB executes a SELECT query and returns a single row
func (sl *StdLib) QueryRowDB(sql string, args ...string) (*database.QueryResult, error) {
	interfaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		interfaceArgs[i] = arg
	}
	return sl.dbManager.QueryRow(sql, interfaceArgs...)
}

// ExecDB executes a non-query SQL statement
func (sl *StdLib) ExecDB(sql string, args ...string) (*database.QueryResult, error) {
	interfaceArgs := make([]interface{}, len(args))
	for i, arg := range args {
		interfaceArgs[i] = arg
	}
	return sl.dbManager.Exec(sql, interfaceArgs...)
}

// GetQueryResult returns the last query result as JSON
func (sl *StdLib) GetQueryResult() (string, error) {
	return sl.dbManager.GetLastResultJSON()
}

// SetEngineFactory sets the execution engine factory
// This allows the HTTP server to execute handlers
func (sl *StdLib) SetEngineFactory(factory func() interface{}) {
	sl.engineFactory = factory
}

// AddMiddleware adds a global middleware to the HTTP server
func (sl *StdLib) AddMiddleware(middleware web.Middleware) error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer == nil {
		return fmt.Errorf("HTTP server not started. Call StartHTTPServer first")
	}

	sl.httpServer.mu.Lock()
	defer sl.httpServer.mu.Unlock()
	sl.httpServer.middlewares = append(sl.httpServer.middlewares, middleware)
	return nil
}

// ClearMiddlewares clears all global middlewares
func (sl *StdLib) ClearMiddlewares() {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer != nil {
		sl.httpServer.mu.Lock()
		defer sl.httpServer.mu.Unlock()
		sl.httpServer.middlewares = make([]web.Middleware, 0)
	}
}

// SHA256Hash computes the SHA256 hash of a string and returns it as a hex string
func (sl *StdLib) SHA256Hash(text string) string {
	hash := sha256.Sum256([]byte(text))
	return hex.EncodeToString(hash[:])
}

// Source executes a Shode script file in the current execution context
// This allows modular code organization by loading functions from separate files
// Usage: Source "path/to/module.sh"
// Note: Actual execution is handled by ExecutionEngine, this is just a placeholder
func (sl *StdLib) Source(filepath string) (string, error) {
	// This function is handled by ExecutionEngine.executeSourceFile
	// It's registered here for consistency with other stdlib functions
	return fmt.Sprintf("Source file: %s", filepath), nil
}
