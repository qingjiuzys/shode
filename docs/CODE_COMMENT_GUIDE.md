# Shode 代码注释指南

## 目录

- [注释原则](#注释原则)
- [Go 文档注释](#go-文档注释)
- [函数注释](#函数注释)
- [包注释](#包注释)
- [常量和变量注释](#常量和变量注释)
- [复杂逻辑注释](#复杂逻辑注释)
- [TODO 和 FIXME](#todo-和-fixme)
- [注释模板](#注释模板)

---

## 注释原则

### 1. 注释应该解释"为什么"而不是"是什么"

```go
// ✅ 好的注释 - 解释原因
// 使用缓冲通道避免生产者阻塞等待消费者
ch := make(chan int, 100)

// ❌ 不好的注释 - 重复代码
// 创建一个容量为 100 的缓冲通道
ch := make(chan int, 100)
```

### 2. 注释应该保持最新

```go
// ✅ 保持注释同步
// 支持的数据库类型：MySQL, PostgreSQL, SQLite
const (
    DBTypeMySQL     = "mysql"
    DBTypePostgres  = "postgresql"
    DBTypeSQLite    = "sqlite"
)

// ❌ 过时的注释
// 支持的数据库类型：MySQL, PostgreSQL
const (
    DBTypeMySQL     = "mysql"
    DBTypePostgres  = "postgresql"
    DBTypeSQLite    = "sqlite"  // 新增但注释未更新
)
```

### 3. 不要注释显而易见的代码

```go
// ✅ 不需要注释
name := "Alice"
if name == "Alice" {
    fmt.Println("Hello, Alice!")
}

// ❌ 多余的注释
// 设置名字为 Alice
name := "Alice"

// 如果名字是 Alice，打印问候
if name == "Alice" {
    fmt.Println("Hello, Alice!")
}
```

---

## Go 文档注释

### 包注释

每个包都应该有包注释：

```go
// Package stdlib provides standard library functions for Shode scripts.
//
// This package includes functions for:
//   - String manipulation (ToUpper, ToLower, Trim, Replace)
//   - File operations (ReadFile, WriteFile, ListFiles)
//   - HTTP server management (StartHTTPServer, RegisterHTTPRoute)
//   - Database operations (ConnectDB, QueryDB, ExecDB)
//   - Caching (SetCache, GetCache, DeleteCache)
//
// Example usage:
//
//	sl := stdlib.New()
//	sl.SetEnv("MY_VAR", "value")
//	data, err := sl.ReadFile("/path/to/file")
package stdlib
```

### 函数注释

导出的函数必须有文档注释：

```go
// ReadFile reads the contents of a file and returns it as a string.
//
// It returns an error if the file does not exist or cannot be read.
// The caller is responsible for validating the file path and handling errors.
//
// Parameters:
//   filename - the absolute or relative path to the file
//
// Returns:
//   string - the complete contents of the file
//   error  - an error if the file cannot be read (os.IsNotExist for missing files)
//
// Example:
//
//	data, err := sl.ReadFile("/etc/hostname")
//	if err != nil {
//	    log.Printf("Failed to read file: %v", err)
//	    return
//	}
//	fmt.Printf("Contents: %s\n", data)
func (sl *StdLib) ReadFile(filename string) (string, error) {
	// 实现代码...
}
```

### 类型注释

```go
// StdLib represents the standard library for Shode scripts.
//
// It provides access to file operations, string manipulation,
// HTTP server management, database operations, and caching.
//
// The StdLib is thread-safe for concurrent use.
type StdLib struct {
	// configuration holds the library settings
	config *Config

	// httpServer manages the HTTP server instance
	httpServer *httpServer

	// cache provides in-memory caching
	cache *cache.Cache
}
```

### 方法注释

```go
// GetEnv retrieves the value of an environment variable.
//
// If the environment variable is not set, it returns an empty string.
// Use UnsetEnv to delete an environment variable.
//
// It is safe to call GetEnv from multiple goroutines simultaneously.
//
// Parameters:
//   key - the name of the environment variable (case-sensitive)
//
// Returns:
//   string - the value of the environment variable, or empty string if not set
//
// Example:
//
//	path := sl.GetEnv("PATH")
//	if path == "" {
//	    log.Println("PATH not set")
//	}
func (sl *StdLib) GetEnv(key string) string {
	// 实现代码...
}
```

---

## 函数注释

### 简单函数

```go
// ToUpper converts a string to uppercase.
func (sl *StdLib) ToUpper(s string) string {
	return strings.ToUpper(s)
}
```

### 带参数验证的函数

```go
// SetCache stores a key-value pair in the cache with an expiration time.
//
// The key must be non-empty. If the key already exists, its value and TTL
// will be updated. Setting TTL to 0 means the cache entry will not expire.
//
// Parameters:
//   key        - the cache key (must not be empty)
//   value      - the value to store
//   ttlSeconds - time to live in seconds (0 for no expiration)
//
// Returns:
//   error - returns an error if key is empty
func (sl *StdLib) SetCache(key, value string, ttlSeconds int) error {
	if key == "" {
		return fmt.Errorf("cache key cannot be empty")
	}
	// 实现代码...
}
```

### 复杂函数

```go
// RegisterHTTPRoute registers an HTTP route with a handler.
//
// The route can be registered with three types of handlers:
//   1. function - Calls a user-defined function
//   2. script  - Executes inline Shode script
//   3. static  - Serves static files from a directory
//
// For function handlers, the function name must be defined in the script
// before calling RegisterHTTPRoute.
//
// Parameters:
//   method      - HTTP method (GET, POST, PUT, DELETE, or * for all methods)
//   path        - URL path (e.g., "/api/users", "/api/users/:id")
//   handlerType - Type of handler: "function", "script", or "static"
//   handler     - Handler name or script content
//
// Returns:
//   error - returns an error if:
//           - HTTP server is not started
//           - Invalid handler type
//           - Route already exists
//
// Example:
//
//	// Function handler
//	function handleUsers() {
//	    SetHTTPResponse 200 '{"users":[]}'
//	}
//	RegisterHTTPRoute "GET" "/api/users" "function" "handleUsers"
//
//	// Script handler
//	RegisterHTTPRoute "POST" "/api/data" "script" "SetHTTPResponse 201 'Created'"
//
//	// Static file handler
//	RegisterHTTPRoute "/" "static" "./public"
func (sl *StdLib) RegisterHTTPRoute(method, path, handlerType, handler string) error {
	// 实现代码...
}
```

---

## 包注释

### 完整的包注释示例

```go
// Package sandbox provides security sandboxing for Shode script execution.
//
// The sandbox enforces security policies to prevent malicious or accidental
// damage to the system. It includes:
//
// # Blacklist Commands
//
// Certain commands are blocked by default:
//   - File system: rm, dd, mkfs, fdisk
//   - System: reboot, shutdown, init
//   - Network: iptables, ifconfig, route
//
// # File Protection
//
// Sensitive system files and directories are protected:
//   - System directories: /bin, /sbin, /usr/bin, /usr/sbin
//   - Configuration: /etc/passwd, /etc/shadow
//   - Other users' home directories
//
// # Pattern Detection
//
// Dangerous patterns are detected and blocked:
//   - Recursive deletion: rm -rf /
//   - Chain loading: wget | sh
//   - File descriptor attacks
//
// # Usage
//
//	sb := sandbox.NewSandbox()
//	sb.EnableBlacklist(true)
//	sb.EnableFileProtection(true)
//
//	if err := sb.CheckCommand("rm -rf /"); err != nil {
//	    log.Printf("Command blocked: %v", err)
//	}
//
// # Configuration
//
// The sandbox can be configured with:
//   - Custom blacklist commands
//   - Protected file patterns
//   - Detection patterns
//   - Whitelist mode (allow-only mode)
//
// See the examples package for usage examples.
package sandbox
```

---

## 常量和变量注释

### 常量注释

```go
// Default HTTP server port
const DefaultHTTPPort = 8080

// Maximum cache entry size (10 MB)
const MaxCacheSize = 10 * 1024 * 1024

// Default cache TTL in seconds (1 hour)
const DefaultCacheTTL = 3600

// Supported database types
const (
	DBTypeMySQL     = "mysql"     // MySQL/MariaDB
	DBTypePostgres  = "postgresql" // PostgreSQL
	DBTypeSQLite    = "sqlite"     // SQLite
)
```

### 变量注释

```go
type StdLib struct {
	// config holds the library configuration
	config *Config

	// httpServer manages the HTTP server instance (lazy-initialized)
	httpServer *httpServer

	// cache provides in-memory key-value storage
	cache *cache.Cache

	// httpMu protects concurrent access to httpServer
	httpMu sync.RWMutex
}
```

---

## 复杂逻辑注释

### 算法说明

```go
// parseQueryParams parses URL query parameters from the raw query string.
//
// The parsing follows these rules:
//  1. Split by '&' to get key=value pairs
//  2. For each pair, split by '=' to separate key and value
//  3. URL-decode both key and value
//  4. Handle duplicate keys by keeping the last value
//
// Example:
//
//	Input:  "name=John&age=30&city=NYC"
//	Output: map[name:John age:30 city:NYC]
//
// Edge cases:
// - Empty values: "key=" -> map[key:]
// - No value: "key" -> map[key:]
// - Multiple "=": "a=b=c" -> map[a:b=c]
func (sl *StdLib) parseQueryParams(rawQuery string) map[string]string {
	// 实现代码...
}
```

### 并发安全说明

```go
// BroadcastWebSocketMessage sends a message to all active WebSocket connections.
//
// This function is thread-safe and can be called concurrently from multiple
// goroutines. It uses a read lock to allow concurrent broadcasts.
//
// Connections that are in the process of closing will be skipped.
// If a message send fails for a particular connection, it continues to
// send to other connections rather than failing entirely.
func (sl *StdLib) BroadcastWebSocketMessage(message string) error {
	sl.wsManager.mu.RLock()
	defer sl.wsManager.mu.RUnlock()

	for id, conn := range sl.wsManager.connections {
		if err := conn.SendMessage(message); err != nil {
			log.Printf("Failed to send to %s: %v", id, err)
		}
	}
	return nil
}
```

---

## TODO 和 FIXME

### 使用标准格式

```go
// TODO: Add support for streaming large files
// This will require implementing chunked transfer encoding
func (sl *StdLib) ReadFile(filename string) (string, error) {
	// ...
}

// FIXME: This is a temporary workaround for the race condition
// Use proper mutex locking instead
var cache map[string]string

// TODO(lujw): Implement connection pooling for database queries
// Target: v0.7.0
func QueryDB(sql string) (*Result, error) {
	// ...
}

// HACK: This is a workaround for the Go 1.21 bug
// Remove when upgrading to 1.22
func workaround() {
	// ...
}
```

### TODO 模板

```go
// TODO: [功能描述]
// [详细说明]
// [相关 issue]: #[issue number]
// [目标版本]: vX.X.X
// [负责人]: @username

// 示例：
// TODO: Implement WebSocket authentication
// Add JWT validation to the WebSocket handshake
// Related: #123
// Target: v0.7.0
// Owner: @devteam
```

---

## 注释模板

### HTTP 处理函数模板

```go
// [FunctionName] handles [HTTP method] [endpoint]
//
// [功能描述]
//
// Request:
//   [Method] [Path]
//   Headers: [required headers]
//   Body: [request body format]
//
// Response:
//   200 OK - [success response]
//   400 Bad Request - [error condition]
//   401 Unauthorized - [error condition]
//   404 Not Found - [error condition]
//   500 Internal Server Error - [error condition]
//
// Example:
//
//	[usage example]
func (sl *StdLib) [FunctionName]() {
	// 实现代码...
}
```

### 数据库操作函数模板

```go
// [FunctionName] [操作描述]
//
// [详细说明]
//
// Parameters:
//   [param1] - [描述]
//   [param2] - [描述]
//
// Returns:
//   [return value] - [描述]
//   error - [可能的错误情况]
//
// SQL Query:
//   [实际的 SQL 查询]
//
// Example:
//
//	[使用示例]
func (sl *StdLib) [FunctionName]() {
	// 实现代码...
}
```

---

## 注释检查清单

在提交代码前，确保：

- [ ] 所有导出的函数都有文档注释
- [ ] 包有完整的包注释
- [ ] 复杂的算法有解释注释
- [ ] 并发代码有安全说明
- [ ] 常量和变量有清晰注释
- [ ] TODO/FIXME 有明确的责任人和时间线
- [ ] 注释与代码保持同步
- [ ] 没有明显的、无用的注释

---

## 工具

### 生成文档

```bash
# 生成包文档
go doc -all ./pkg/stdlib

# 在浏览器中查看
godoc -http=:6060
```

### 检查注释

```bash
# 使用 golangci-lint 检查注释规范
golangci-lint run

# 使用 staticcheck
staticcheck ./...
```

---

## 参考资源

- [Effective Go - Commentary](https://go.dev/doc/effective_go#commentary)
- [Go Doc Comments](https://tip.golang.org/doc/comment)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout)
