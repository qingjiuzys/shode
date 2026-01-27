package stdlib

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
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
	// WebSocket manager
	wsManager *WebSocketManager
}

// FilesManager handles file operations
type FilesManager struct{}

// SystemManager handles system operations
type SystemManager struct{}

// NetworkManager handles network operations
type NetworkManager struct{}

// ArchiveManager handles compression/archive operations
type ArchiveManager struct{}

// StaticFileConfig configures static file serving
type StaticFileConfig struct {
	Directory       string   // File directory
	IndexFiles      []string // Index files (default: ["index.html", "index.htm"])
	DirectoryBrowse bool     // Directory browsing toggle
	CacheControl    string   // Cache-Control header (e.g., "max-age=3600")
	EnableGzip      bool     // gzip compression toggle
	SPAFallback     string   // SPA fallback file
}

// routeHandler represents a route handler
type routeHandler struct {
	method      string // HTTP method (GET, POST, PUT, DELETE, PATCH, "*" for all)
	path        string
	handlerType string // "function", "script", or "static"
	handlerName string // function name or script content
	staticConfig *StaticFileConfig // Only for "static" type
}

// httpServer represents an HTTP server instance
type httpServer struct {
	server      *http.Server
	mux         *http.ServeMux
	routes      map[string]*routeHandler // routeKey (method:path) -> handler
	staticRoutes map[string]*StaticFileConfig // route prefix -> config (for static routes)
	registeredPaths map[string]bool // Track which paths have mux handlers registered
	isRunning   bool
	middlewares []web.Middleware // Global middlewares
	enableRequestLog bool // Enable request logging
	requestLogLevel string // Log level: "debug", "info", "error"
	errorPages map[int]string // Custom error pages (status code -> file path)
	mu          sync.RWMutex
}

// responseWriterWrapper wraps http.ResponseWriter to capture status code
type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

func (w *responseWriterWrapper) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.written = true
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if !w.written {
		w.statusCode = http.StatusOK
		w.written = true
	}
	return w.ResponseWriter.Write(b)
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

// Static file serving helper functions

// getContentType determines MIME type based on file extension
func (sl *StdLib) getContentType(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return "text/plain; charset=utf-8"
	}

	// Remove the leading dot
	ext = ext[1:]

	// Try to detect the MIME type
	mimeType := mime.TypeByExtension("." + ext)
	if mimeType != "" {
		return mimeType
	}

	// Fallback to common types
	switch strings.ToLower(ext) {
	case "html", "htm":
		return "text/html; charset=utf-8"
	case "css":
		return "text/css; charset=utf-8"
	case "js":
		return "application/javascript; charset=utf-8"
	case "json":
		return "application/json; charset=utf-8"
	case "xml":
		return "application/xml; charset=utf-8"
	case "png":
		return "image/png"
	case "jpg", "jpeg":
		return "image/jpeg"
	case "gif":
		return "image/gif"
	case "svg":
		return "image/svg+xml"
	case "ico":
		return "image/x-icon"
	case "woff", "woff2":
		return "font/woff2"
	case "ttf":
		return "font/ttf"
	case "eot":
		return "application/vnd.ms-fontobject"
	case "pdf":
		return "application/pdf"
	case "zip":
		return "application/zip"
	case "txt":
		return "text/plain; charset=utf-8"
	case "md":
		return "text/markdown; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

// validateStaticDirectory validates and normalizes directory path
func (sl *StdLib) validateStaticDirectory(directory string) (string, error) {
	// Convert to absolute path if relative
	if !filepath.IsAbs(directory) {
		wd, err := os.Getwd()
		if err != nil {
			return "", fmt.Errorf("failed to get working directory: %v", err)
		}
		directory = filepath.Join(wd, directory)
	}

	// Clean the path
	directory = filepath.Clean(directory)

	// Check if directory exists
	info, err := os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("directory does not exist: %s", directory)
		}
		return "", fmt.Errorf("failed to access directory: %v", err)
	}

	// Check if it's actually a directory
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", directory)
	}

	return directory, nil
}

// serveFile serves a single file with proper headers, streaming gzip, and cache support
func (sl *StdLib) serveFile(w http.ResponseWriter, r *http.Request, filePath string, config *StaticFileConfig) error {
	// Check if file exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found")
		}
		return fmt.Errorf("failed to access file: %v", err)
	}

	// Check if it's a directory
	if info.IsDir() {
		return fmt.Errorf("is directory")
	}

	fileSize := info.Size()
	modTime := info.ModTime()

	// Generate ETag based on file metadata (strong ETag)
	// Format: "mtime-size" in hex for uniqueness and cache validation
	etag := fmt.Sprintf("%x-%x", modTime.Unix(), fileSize)

	// Set Last-Modified header (RFC 1123 format)
	w.Header().Set("Last-Modified", modTime.UTC().Format(http.TimeFormat))
	w.Header().Set("ETag", etag)

	// Check conditional requests - If-None-Match (ETag)
	ifNoneMatch := r.Header.Get("If-None-Match")
	if ifNoneMatch != "" && ifNoneMatch == etag {
		w.WriteHeader(http.StatusNotModified)
		return nil
	}

	// Check conditional requests - If-Modified-Since
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince != "" {
		ifModTime, err := http.ParseTime(ifModifiedSince)
		if err == nil && !modTime.After(ifModTime) {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
	}

	// Parse Range header (supports both single and multiple ranges)
	type byteRange struct {
		start int64
		end   int64
	}

	var ranges []byteRange
	var sendPartial bool

	rangeHeader := r.Header.Get("Range")
	if rangeHeader != "" && strings.HasPrefix(rangeHeader, "bytes=") {
		rangeSpec := strings.TrimPrefix(rangeHeader, "bytes=")

		// Check if this is a multi-range request (contains comma)
		if strings.Contains(rangeSpec, ",") {
			// Multi-range request
			rangeParts := strings.Split(rangeSpec, ",")
			for _, part := range rangeParts {
				part = strings.TrimSpace(part)
				partRanges := strings.Split(part, "-")

				if len(partRanges) != 2 {
					// Invalid range format
					w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
					http.Error(w, "Invalid Range Format", http.StatusRequestedRangeNotSatisfiable)
					return nil
				}

				var rStart, rEnd int64
				var err error

				// Parse start position
				if partRanges[0] != "" {
					rStart, err = strconv.ParseInt(partRanges[0], 10, 64)
					if err != nil || rStart < 0 || rStart >= fileSize {
						w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
						http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
						return nil
					}
				}

				// Parse end position
				if partRanges[1] != "" {
					rEnd, err = strconv.ParseInt(partRanges[1], 10, 64)
					if err != nil || rEnd < rStart || rEnd >= fileSize {
						w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
						http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
						return nil
					}
				} else {
					// No end specified, use end of file
					rEnd = fileSize - 1
				}

				ranges = append(ranges, byteRange{start: rStart, end: rEnd})
			}

			if len(ranges) > 0 {
				sendPartial = true
			}
		} else {
			// Single range request (original logic)
			parts := strings.Split(rangeSpec, "-")

			if len(parts) == 2 {
				var start, end int64

				// Parse start position
				if parts[0] != "" {
					start, err = strconv.ParseInt(parts[0], 10, 64)
					if err != nil || start < 0 || start >= fileSize {
						w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
						http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
						return nil
					}
				}

				// Parse end position
				if parts[1] != "" {
					end, err = strconv.ParseInt(parts[1], 10, 64)
					if err != nil || end < start || end >= fileSize {
						w.Header().Set("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
						http.Error(w, "Range Not Satisfiable", http.StatusRequestedRangeNotSatisfiable)
						return nil
					}
				} else {
					end = fileSize - 1
				}

				ranges = append(ranges, byteRange{start: start, end: end})
				sendPartial = true
			}
		}
	}

	// Check if gzip compression is enabled and client supports it
	// Note: Don't use gzip with Range requests
	shouldGzip := config.EnableGzip && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && !sendPartial

	// Set cache control header if specified
	if config.CacheControl != "" {
		w.Header().Set("Cache-Control", config.CacheControl)
	}

	// Set content type
	contentType := sl.getContentType(filePath)
	w.Header().Set("Accept-Ranges", "bytes")

	// Open file for streaming
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	// Handle different response types
	if !sendPartial {
		// No range request - send entire file
		w.Header().Set("Content-Type", contentType)

		if !shouldGzip {
			// Only set Content-Length if not using gzip
			w.Header().Set("Content-Length", fmt.Sprintf("%d", fileSize))
		}

		// Apply gzip compression with streaming if enabled
		if shouldGzip {
			// Set gzip headers
			w.Header().Set("Content-Encoding", "gzip")
			w.Header().Set("Vary", "Accept-Encoding")

			// Create streaming gzip writer
			gzipWriter := gzip.NewWriter(w)
			defer gzipWriter.Close()

			// Stream and compress in chunks
			_, err := io.Copy(gzipWriter, file)
			if err != nil {
				return fmt.Errorf("failed to write compressed content: %v", err)
			}

			// Flush gzip writer
			if err := gzipWriter.Close(); err != nil {
				return fmt.Errorf("failed to close gzip writer: %v", err)
			}
		} else {
			// Stream file directly without compression
			if _, err := io.Copy(w, file); err != nil {
				return fmt.Errorf("failed to write content: %v", err)
			}
		}
	} else if len(ranges) == 1 {
		// Single range request
		r := ranges[0]
		contentLength := r.end - r.start + 1

		w.Header().Set("Content-Type", contentType)
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", r.start, r.end, fileSize))
		w.Header().Set("Content-Length", fmt.Sprintf("%d", contentLength))
		w.WriteHeader(http.StatusPartialContent)

		// Seek to start position and send the range
		file.Seek(r.start, 0)
		limitedReader := io.LimitReader(file, contentLength)
		if _, err := io.Copy(w, limitedReader); err != nil {
			return fmt.Errorf("failed to write content: %v", err)
		}
	} else {
		// Multi-range request - send multipart/byteranges response
		boundary := fmt.Sprintf("%x", time.Now().UnixNano())

		w.Header().Set("Content-Type", fmt.Sprintf("multipart/byteranges; boundary=%s", boundary))
		w.WriteHeader(http.StatusPartialContent)

		for _, r := range ranges {
			// Write boundary
			fmt.Fprintf(w, "--%s\r\n", boundary)
			fmt.Fprintf(w, "Content-Type: %s\r\n", contentType)
			fmt.Fprintf(w, "Content-Range: bytes %d-%d/%d\r\n", r.start, r.end, fileSize)
			fmt.Fprintf(w, "\r\n")

			// Seek to start position and write the range
			file.Seek(r.start, 0)
			contentLength := r.end - r.start + 1
			limitedReader := io.LimitReader(file, contentLength)
			if _, err := io.Copy(w, limitedReader); err != nil {
				return fmt.Errorf("failed to write content: %v", err)
			}

			// Write CRLF after each part
			fmt.Fprintf(w, "\r\n")
		}

		// Write final boundary
		fmt.Fprintf(w, "--%s--\r\n", boundary)
	}

	return nil
}

// multipartWriter helps with writing multipart responses
type multipartWriter struct {
	w        http.ResponseWriter
	boundary string
}

// serveDirectoryListing generates a directory browsing page
func (sl *StdLib) serveDirectoryListing(w http.ResponseWriter, r *http.Request, dirPath, requestPath string) error {
	// Read directory entries
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory: %v", err)
	}

	// Generate HTML listing
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<!DOCTYPE html>\n")
	fmt.Fprintf(w, "<html>\n")
	fmt.Fprintf(w, "<head>\n")
	fmt.Fprintf(w, "  <title>Index of %s</title>\n", requestPath)
	fmt.Fprintf(w, "  <style>\n")
	fmt.Fprintf(w, "    body { font-family: monospace; margin: 2em; }\n")
	fmt.Fprintf(w, "    h1 { font-size: 1.2em; }\n")
	fmt.Fprintf(w, "    table { border-collapse: collapse; }\n")
	fmt.Fprintf(w, "    td, th { padding: 0.5em; text-align: left; }\n")
	fmt.Fprintf(w, "    a { text-decoration: none; color: #0066cc; }\n")
	fmt.Fprintf(w, "    a:hover { text-decoration: underline; }\n")
	fmt.Fprintf(w, "  </style>\n")
	fmt.Fprintf(w, "</head>\n")
	fmt.Fprintf(w, "<body>\n")
	fmt.Fprintf(w, "  <h1>Index of %s</h1>\n", requestPath)
	fmt.Fprintf(w, "  <table>\n")
	fmt.Fprintf(w, "    <tr><th>Name</th><th>Size</th></tr>\n")

	// Parent directory link
	if requestPath != "/" {
		fmt.Fprintf(w, "    <tr><td><a href=\"..\">../</a></td><td>-</td></tr>\n")
	}

	// Directory entries
	for _, entry := range entries {
		name := entry.Name()
		isDir := entry.IsDir()

		// Get size from FileInfo
		info, err := entry.Info()
		if err != nil {
			continue // Skip entries we can't get info for
		}
		size := info.Size()

		// Skip hidden files
		if strings.HasPrefix(name, ".") {
			continue
		}

		displayName := name
		if isDir {
			displayName += "/"
		}

		link := filepath.Join(requestPath, name)
		fmt.Fprintf(w, "    <tr><td><a href=\"%s\">%s</a></td><td>%d</td></tr>\n", link, displayName, size)
	}

	fmt.Fprintf(w, "  </table>\n")
	fmt.Fprintf(w, "</body>\n")
	fmt.Fprintf(w, "</html>\n")

	return nil
}

// serveStaticFile handles static file requests
func (sl *StdLib) serveStaticFile(w http.ResponseWriter, r *http.Request, config *StaticFileConfig, routePrefix string) {
	// Get the requested path relative to the route prefix
	requestPath := r.URL.Path

	// Remove the route prefix from the request path
	relativePath := strings.TrimPrefix(requestPath, routePrefix)
	if relativePath == "" || relativePath == "/" {
		relativePath = "/"
	} else {
		// Ensure relative path starts with /
		if !strings.HasPrefix(relativePath, "/") {
			relativePath = "/" + relativePath
		}
	}

	// Security check: prevent path traversal attacks
	if strings.Contains(relativePath, "..") {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Build the full file path
	filePath := filepath.Join(config.Directory, relativePath)

	// Clean the path to prevent any path traversal attempts
	filePath = filepath.Clean(filePath)

	// Verify the file is still within the configured directory
	absFilePath, err := filepath.Abs(filePath)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	absDir, err := filepath.Abs(config.Directory)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !strings.HasPrefix(absFilePath, absDir) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	// Check if the path exists
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			// Check for SPA fallback
			if config.SPAFallback != "" {
				fallbackPath := filepath.Join(config.Directory, config.SPAFallback)
				if _, err := os.Stat(fallbackPath); err == nil {
					// Serve fallback file
					if err := sl.serveFile(w, r, fallbackPath, config); err != nil {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
					return
				}
			}
			// File not found - render error page
			sl.renderErrorPage(w, r, http.StatusNotFound)
			return
		}
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// If it's a directory, try to serve index file
	if info.IsDir() {
		// Try each index file
		indexServed := false
		for _, indexFile := range config.IndexFiles {
			indexPath := filepath.Join(filePath, indexFile)
			if _, err := os.Stat(indexPath); err == nil {
				if err := sl.serveFile(w, r, indexPath, config); err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
				indexServed = true
				break
			}
		}

		// If no index file found, show directory listing or 404
		if !indexServed {
			if config.DirectoryBrowse {
				if err := sl.serveDirectoryListing(w, r, filePath, relativePath); err != nil {
					http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				}
			} else {
				http.NotFound(w, r)
			}
		}
		return
	}

	// Serve the file
	if err := sl.serveFile(w, r, filePath, config); err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// FileSystem functions

// ReadFile reads the contents of a file (replaces 'cat')
func (sl *StdLib) ReadFile(filename string) (string, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file %s: %v", filename, err)
	}
	return string(content), nil
}

// WriteFile writes content to a file (replaces echo > file)
func (sl *StdLib) WriteFile(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

// ListFiles lists files in a directory (replaces 'ls')
func (sl *StdLib) ListFiles(dirpath string) ([]string, error) {
	files, err := os.ReadDir(dirpath)
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

// logRequest logs HTTP request details
func (sl *StdLib) logRequest(r *http.Request, statusCode int, duration time.Duration) {
	if sl.httpServer == nil || !sl.httpServer.enableRequestLog {
		return
	}

	// Filter by log level
	level := sl.httpServer.requestLogLevel
	if level == "" {
		level = "info" // default level
	}

	// Only log errors if level is "error"
	if level == "error" && statusCode < 400 {
		return
	}

	// Format: [timestamp] method path status duration
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logMessage := fmt.Sprintf("[%s] %s %s %d %vms\n",
		timestamp,
		r.Method,
		r.URL.Path,
		statusCode,
		duration.Milliseconds(),
	)

	// Add query string if present
	if r.URL.RawQuery != "" {
		logMessage = fmt.Sprintf("[%s] %s %s?%s %d %vms\n",
			timestamp,
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			statusCode,
			duration.Milliseconds(),
		)
	}

	if level == "debug" || statusCode >= 400 {
		fmt.Fprintf(os.Stderr, "[HTTP-LOG] %s", logMessage)
	}
}

// EnableRequestLog enables HTTP request logging
func (sl *StdLib) EnableRequestLog(level string) error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	if sl.httpServer == nil {
		return fmt.Errorf("HTTP server not started. Call StartHTTPServer first")
	}

	// Validate log level
	if level != "" && level != "debug" && level != "info" && level != "error" {
		return fmt.Errorf("invalid log level: %s (must be 'debug', 'info', or 'error')", level)
	}

	sl.httpServer.enableRequestLog = true
	sl.httpServer.requestLogLevel = level

	return nil
}

// StartHTTPServer starts an HTTP server on the specified port
// Usage: StartHTTPServer "9188"
func (sl *StdLib) StartHTTPServer(port string) error {
	sl.httpMu.Lock()
	defer sl.httpMu.Unlock()

	// Remove surrounding quotes if present
	port = strings.Trim(port, "\"")

	// Debug output
	fmt.Fprintf(os.Stderr, "[DEBUG] StartHTTPServer called with port %s (trimmed)\n", port)

	// Parse port
	portNum, err := strconv.Atoi(port)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] StartHTTPServer: invalid port %s: %v\n", port, err)
		return fmt.Errorf("invalid port: %s", port)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] StartHTTPServer: port parsed as %d\n", portNum)

	// Check if server already exists and is running
	if sl.httpServer != nil && sl.httpServer.isRunning {
		fmt.Fprintf(os.Stderr, "[DEBUG] StartHTTPServer: server already running\n")
		return fmt.Errorf("HTTP server is already running")
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] StartHTTPServer: no existing server, proceeding...\n")

	// Create new server
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", portNum),
		Handler: mux,
	}

	sl.httpServer = &httpServer{
		server:          server,
		mux:             mux,
		routes:          make(map[string]*routeHandler),
		staticRoutes:    make(map[string]*StaticFileConfig),
		registeredPaths: make(map[string]bool),
		errorPages:      make(map[int]string),
		isRunning:       true, // Set to true immediately, before starting goroutine
	}

	// Debug: confirm httpServer was created
	fmt.Fprintf(os.Stderr, "[DEBUG] StartHTTPServer: httpServer created, isRunning=%v\n", sl.httpServer.isRunning)

	// Start server in goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(os.Stderr, "HTTP server error: %v\n", err)
		}

		// Mark server as not running when it stops
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

	// Debug output
	fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: method=%s, path=%s, handlerType=%s, handler=%s\n", method, path, handlerType, handler)

	// Normalize method to uppercase
	method = strings.ToUpper(method)
	fmt.Fprintf(os.Stderr, "[DEBUG] Method normalized to: %s\n", method)
	if method == "" {
		method = "*"
	}

	// Validate handler type
	if handlerType != "function" && handlerType != "script" && handlerType != "static" {
		return fmt.Errorf("invalid handler type: %s (must be 'function', 'script', or 'static')", handlerType)
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] Handler type validated: %s\n", handlerType)

	sl.httpServer.mu.Lock()
	defer sl.httpServer.mu.Unlock()
	fmt.Fprintf(os.Stderr, "[DEBUG] Acquired httpServer lock\n")

	// Create route key: method:path
	routeKey := fmt.Sprintf("%s:%s", method, path)

	// Create handler with static config if needed
	var routeHdlr *routeHandler
	if handlerType == "static" {
		fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: Creating static handler for path=%s, directory=%s\n", path, handler)
		// Validate and prepare static file configuration
		absDir, err := sl.validateStaticDirectory(handler)
		if err != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: Static directory validation failed: %v\n", err)
			return fmt.Errorf("invalid static directory: %v", err)
		}
		fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: Static directory validated: %s\n", absDir)
		routeHdlr = &routeHandler{
			method:      method,
			path:        path,
			handlerType: handlerType,
			handlerName: handler,
			staticConfig: &StaticFileConfig{
				Directory:       absDir,
				IndexFiles:      []string{"index.html", "index.htm"},
				DirectoryBrowse: false,
				CacheControl:    "",
				EnableGzip:      false,
				SPAFallback:     "",
			},
		}
	} else {
		routeHdlr = &routeHandler{
			method:      method,
			path:        path,
			handlerType: handlerType,
			handlerName: handler,
		}
	}

	// Store the handler
	sl.httpServer.routes[routeKey] = routeHdlr

	// If this is a static route, also store in staticRoutes for prefix matching
	if handlerType == "static" && routeHdlr.staticConfig != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: Storing static route for path=%s\n", path)
		sl.httpServer.staticRoutes[path] = routeHdlr.staticConfig
	}

	// Check if path is already registered
	// If not, register a method-aware handler
	fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: Checking path=%s, registered=%v\n", path, sl.httpServer.registeredPaths[path])
	if !sl.httpServer.registeredPaths[path] {
		// Register the route with method checking
		fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRoute: Registering mux handler for path %s\n", path)
		sl.httpServer.registeredPaths[path] = true
		sl.httpServer.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			// Record start time for logging
			startTime := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: 200}

			fmt.Fprintf(os.Stderr, "[DEBUG] mux handler called: %s %s\n", r.Method, r.URL.Path)
			sl.httpServer.mu.RLock()
			defer sl.httpServer.mu.RUnlock()

			// Log request at the end
			defer func() {
				duration := time.Since(startTime)
				sl.logRequest(r, wrapped.statusCode, duration)
			}()

			// Check for exact method match
			exactKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
			handler, exactExists := sl.httpServer.routes[exactKey]
			fmt.Fprintf(os.Stderr, "[DEBUG] Looking for exact key: %s, found=%v\n", exactKey, exactExists)

			// Check for wildcard method match
			wildcardKey := fmt.Sprintf("*:%s", r.URL.Path)
			wildcardHandler, wildcardExists := sl.httpServer.routes[wildcardKey]
			fmt.Fprintf(os.Stderr, "[DEBUG] Looking for wildcard key: %s, found=%v\n", wildcardKey, wildcardExists)

			var selectedHandler *routeHandler
			if exactExists {
				selectedHandler = handler
			} else if wildcardExists {
				selectedHandler = wildcardHandler
			}

			if selectedHandler == nil {
				// No exact handler found, check for static routes
				var staticConfig *StaticFileConfig
				for prefix, config := range sl.httpServer.staticRoutes {
					if strings.HasPrefix(r.URL.Path, prefix) {
						staticConfig = config
						break
					}
				}
				if staticConfig != nil {
					// Find the longest matching prefix
					var longestPrefix string
					for prefix, config := range sl.httpServer.staticRoutes {
						if strings.HasPrefix(r.URL.Path, prefix) && len(prefix) > len(longestPrefix) {
							longestPrefix = prefix
							staticConfig = config
						}
					}
					// Serve static file
					sl.serveStaticFile(wrapped, r, staticConfig, longestPrefix)
					return
				}
				wrapped.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(wrapped, "Method Not Allowed\n")
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

							// Store request context globally before executing function
							// This ensures SetHTTPResponse can find it
							goroutineID := fmt.Sprintf("func-%s", selectedHandler.handlerName)
							sl.requestContexts.Store(goroutineID, reqCtx)
							sl.requestContexts.Store("current", reqCtx)

							// Call ExecuteCommand using reflection
							ctxValue := reflect.ValueOf(ctx)
							cmdNodeValue := reflect.ValueOf(cmdNode)
							results := executeCommandMethod.Call([]reflect.Value{ctxValue, cmdNodeValue})

							// Clean up context
							sl.requestContexts.Delete(goroutineID)

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
			} else if selectedHandler.handlerType == "static" {
				// Serve static files
				sl.serveStaticFile(w, r, selectedHandler.staticConfig, path)
				return
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
				wrapped.Header().Set(k, v)
			}
			if len(headers) == 0 {
				wrapped.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}

			wrapped.WriteHeader(status)
			if body != "" {
				fmt.Fprintf(wrapped, "%s", body)
			} else {
				fmt.Fprintf(wrapped, "Handler: %s (type: %s, method: %s)\n",
					selectedHandler.handlerName, selectedHandler.handlerType, selectedHandler.method)
			}
		})
	}

	return nil
}

// RegisterStaticRoute registers a simple static file route
// Usage: RegisterStaticRoute "/" "./public"
func (sl *StdLib) RegisterStaticRoute(path, directory string) error {
	return sl.RegisterHTTPRoute("GET", path, "static", directory)
}

// RegisterStaticRouteAdvanced registers a static file route with advanced options
// Usage: RegisterStaticRouteAdvanced "/" "./public" "true" "false" "" "false" ""
func (sl *StdLib) RegisterStaticRouteAdvanced(path, directory, indexFiles, directoryBrowse, cacheControl, enableGzip, spaFallback string) error {
	return sl.RegisterHTTPRouteAdvanced("GET", path, directory, indexFiles, directoryBrowse, cacheControl, enableGzip, spaFallback)
}

// RegisterHTTPRouteAdvanced registers an HTTP route with advanced static file configuration
func (sl *StdLib) RegisterHTTPRouteAdvanced(method, path, directory, indexFiles, directoryBrowse, cacheControl, enableGzip, spaFallback string) error {
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
	handlerType := "static"

	sl.httpServer.mu.Lock()
	defer sl.httpServer.mu.Unlock()

	// Create route key: method:path
	routeKey := fmt.Sprintf("%s:%s", method, path)

	// Validate and prepare static file configuration
	absDir, err := sl.validateStaticDirectory(directory)
	if err != nil {
		return fmt.Errorf("invalid static directory: %v", err)
	}

	// Parse index files (comma-separated)
	var indexFileList []string
	if indexFiles != "" && indexFiles != "false" {
		indexFileList = strings.Split(indexFiles, ",")
		for i, f := range indexFileList {
			indexFileList[i] = strings.TrimSpace(f)
		}
	} else {
		// Default index files
		indexFileList = []string{"index.html", "index.htm"}
	}

	// Parse directory browse (boolean)
	dirBrowse := false
	if directoryBrowse == "true" || directoryBrowse == "1" {
		dirBrowse = true
	}

	// Parse enable gzip (boolean)
	enableGzipFlag := false
	if enableGzip == "true" || enableGzip == "1" {
		enableGzipFlag = true
	}

	// Create route handler with full static config
	routeHdlr := &routeHandler{
		method:      method,
		path:        path,
		handlerType: handlerType,
		handlerName: directory,
		staticConfig: &StaticFileConfig{
			Directory:       absDir,
			IndexFiles:      indexFileList,
			DirectoryBrowse: dirBrowse,
			CacheControl:    cacheControl,
			EnableGzip:      enableGzipFlag,
			SPAFallback:     spaFallback,
		},
	}

	// Store the handler
	sl.httpServer.routes[routeKey] = routeHdlr

	// If this is a static route, also store in staticRoutes for prefix matching
	if routeHdlr.staticConfig != nil {
		fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRouteAdvanced: Storing static route for path=%s\n", path)
		sl.httpServer.staticRoutes[path] = routeHdlr.staticConfig
	}

	// Check if path is already registered
	fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRouteAdvanced: Checking path=%s, registered=%v\n", path, sl.httpServer.registeredPaths[path])
	if !sl.httpServer.registeredPaths[path] {
		// Register the route with method checking
		fmt.Fprintf(os.Stderr, "[DEBUG] RegisterHTTPRouteAdvanced: Registering mux handler for path %s\n", path)
		sl.httpServer.registeredPaths[path] = true
		sl.httpServer.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			// Record start time for logging
			startTime := time.Now()

			// Wrap response writer to capture status code
			wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: 200}

			fmt.Fprintf(os.Stderr, "[DEBUG] mux handler called: %s %s\n", r.Method, r.URL.Path)
			sl.httpServer.mu.RLock()
			defer sl.httpServer.mu.RUnlock()

			// Log request at the end
			defer func() {
				duration := time.Since(startTime)
				sl.logRequest(r, wrapped.statusCode, duration)
			}()

			// Check for exact method match
			exactKey := fmt.Sprintf("%s:%s", r.Method, r.URL.Path)
			handler, exactExists := sl.httpServer.routes[exactKey]
			fmt.Fprintf(os.Stderr, "[DEBUG] Looking for exact key: %s, found=%v\n", exactKey, exactExists)

			// Check for wildcard method match
			wildcardKey := fmt.Sprintf("*:%s", r.URL.Path)
			wildcardHandler, wildcardExists := sl.httpServer.routes[wildcardKey]
			fmt.Fprintf(os.Stderr, "[DEBUG] Looking for wildcard key: %s, found=%v\n", wildcardKey, wildcardExists)

			var selectedHandler *routeHandler
			if exactExists {
				selectedHandler = handler
			} else if wildcardExists {
				selectedHandler = wildcardHandler
			}

			if selectedHandler == nil {
				// No exact handler found, check for static routes
				var staticConfig *StaticFileConfig
				for prefix, config := range sl.httpServer.staticRoutes {
					if strings.HasPrefix(r.URL.Path, prefix) {
						staticConfig = config
						break
					}
				}
				if staticConfig != nil {
					// Find the longest matching prefix
					var longestPrefix string
					for prefix, config := range sl.httpServer.staticRoutes {
						if strings.HasPrefix(r.URL.Path, prefix) && len(prefix) > len(longestPrefix) {
							longestPrefix = prefix
							staticConfig = config
						}
					}
					// Serve static file
					sl.serveStaticFile(wrapped, r, staticConfig, longestPrefix)
					return
				}
				wrapped.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(wrapped, "Method Not Allowed\n")
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

							// Store request context globally before executing function
							// This ensures SetHTTPResponse can find it
							goroutineID := fmt.Sprintf("func-%s", selectedHandler.handlerName)
							sl.requestContexts.Store(goroutineID, reqCtx)
							sl.requestContexts.Store("current", reqCtx)

							// Call ExecuteCommand using reflection
							ctxValue := reflect.ValueOf(ctx)
							cmdNodeValue := reflect.ValueOf(cmdNode)
							results := executeCommandMethod.Call([]reflect.Value{ctxValue, cmdNodeValue})

							// Clean up context
							sl.requestContexts.Delete(goroutineID)

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
			} else if selectedHandler.handlerType == "static" {
				// Serve static files
				sl.serveStaticFile(w, r, selectedHandler.staticConfig, path)
				return
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
				wrapped.Header().Set(k, v)
			}
			if len(headers) == 0 {
				wrapped.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}

			wrapped.WriteHeader(status)
			if body != "" {
				fmt.Fprintf(wrapped, "%s", body)
			} else {
				fmt.Fprintf(wrapped, "Handler: %s (type: %s, method: %s)\n",
					selectedHandler.handlerName, selectedHandler.handlerType, selectedHandler.method)
			}
		})
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
		startTime := time.Now()
		wrapped := &responseWriterWrapper{ResponseWriter: w, statusCode: 200}

		wrapped.Header().Set("Content-Type", "text/plain; charset=utf-8")
		wrapped.WriteHeader(http.StatusOK)
		fmt.Fprintf(wrapped, "%s\n", response)

		duration := time.Since(startTime)
		sl.logRequest(r, wrapped.statusCode, duration)
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
		// Debug output
		fmt.Fprintf(os.Stderr, "[DEBUG] IsHTTPServerRunning: httpServer is nil\n")
		return false
	}

	sl.httpServer.mu.RLock()
	defer sl.httpServer.mu.RUnlock()

	running := sl.httpServer.isRunning
	// Debug output
	fmt.Fprintf(os.Stderr, "[DEBUG] IsHTTPServerRunning: isRunning=%v\n", running)
	return running
}

// createRequestContext creates a request context from an HTTP request
func (sl *StdLib) createRequestContext(r *http.Request) *HTTPRequestContext {
	// Read body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body for potential re-read

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
	fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse called: status=%d, body=%s\n", status, body)
	ctx := sl.getCurrentRequestContext()
	fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse: ctx=%v\n", ctx != nil)
	if ctx == nil {
		// Fallback: try to find any context
		fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse: ctx is nil, trying fallback\n")
		var foundCtx *HTTPRequestContext
		sl.requestContexts.Range(func(key, value interface{}) bool {
			if httpCtx, ok := value.(*HTTPRequestContext); ok {
				foundCtx = httpCtx
				return false // Stop at first match
			}
			return true
		})
		if foundCtx != nil {
			fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse: found fallback ctx\n")
			foundCtx.Response.mu.Lock()
			foundCtx.Response.Status = status
			foundCtx.Response.Body = body
			foundCtx.Response.mu.Unlock()
		} else {
			fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse: no fallback ctx found\n")
		}
		return
	}
	fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse: setting response on ctx\n")
	ctx.Response.mu.Lock()
	defer ctx.Response.mu.Unlock()
	ctx.Response.Status = status
	ctx.Response.Body = body
	fmt.Fprintf(os.Stderr, "[DEBUG] SetHTTPResponse: response set successfully\n")
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
