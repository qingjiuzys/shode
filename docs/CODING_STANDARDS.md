# Shode 编码规范和最佳实践

## 错误处理规范

### 1. 统一使用 pkg/errors 包

```go
// ✅ 正确 - 使用自定义错误类型
import "gitee.com/com_818cloud/shode/pkg/errors"

// 创建基础错误
err := errors.NewSecurityViolation("dangerous command blocked")
err := errors.NewCommandNotFound("git")
err := errors.NewExecutionFailed("script execution failed", cause)

// 添加上下文信息
err := errors.NewFileNotFoundError("/path/to/file").
    WithContext("operation", "read").
    WithContext("user_id", userID)

// ❌ 错误 - 使用标准库错误
err := fmt.Errorf("file not found: %s", path)
```

### 2. 错误包装和传播

```go
// ✅ 正确 - 保留错误链
func ProcessFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return errors.WrapError(
            errors.ErrFileNotFound,
            "failed to read config file",
            err,
        ).WithContext("path", path)
    }
    // ...
}

// ❌ 错误 - 丢失原始错误
func ProcessFile(path string) error {
    data, err := os.ReadFile(path)
    if err != nil {
        return fmt.Errorf("read failed")
    }
}
```

### 3. 错误类型检查

```go
// ✅ 正确 - 使用类型断言检查
if errors.IsSecurityViolation(err) {
    log.Printf("Security violation detected: %v", err)
    return
}

if errors.IsTimeout(err) {
    // 重试逻辑
    return retry()
}

// ✅ 正确 - 获取错误类型
switch errors.GetErrorType(err) {
case errors.ErrNetworkError:
    // 处理网络错误
case errors.ErrTimeout:
    // 处理超时
default:
    // 处理其他错误
}
```

### 4. 错误上下文

```go
// ✅ 正确 - 添加丰富上下文
func ProcessUser(userID string) error {
    user, err := db.GetUser(userID)
    if err != nil {
        return errors.WrapError(
            errors.ErrExecutionFailed,
            "failed to fetch user",
            err,
        ).WithContext("user_id", userID).
         WithContext("operation", "get_user").
         WithContext("table", "users")
    }
}
```

### 5. 错误日志记录

```go
// ✅ 正确 - 记录错误时包含上下文
func HandleRequest(req *Request) error {
    err := processRequest(req)
    if err != nil {
        // 获取错误上下文
        ctx := errors.GetErrorContext(err)
        log.Printf("Request failed: type=%s, msg=%s, context=%+v",
            errors.GetErrorType(err),
            err,
            ctx)
        return err
    }
}
```

## 代码风格规范

### 1. 函数命名

```go
// ✅ 导出函数 - 大驼峰
func ProcessFile() error { ... }
func GetUserByID(id string) *User { ... }

// ✅ 内部函数 - 小驼峰
func parseInput() error { ... }
func validateUser() bool { ... }
```

### 2. 错误处理

```go
// ✅ 立即处理错误
file, err := os.Open(path)
if err != nil {
    return errors.WrapError(errors.ErrFileNotFound, "open failed", err)
}
defer file.Close()

// ❌ 不要忽略错误
file, err := os.Open(path)
// 忘记检查 err
```

### 3. 资源清理

```go
// ✅ 正确 - 使用 defer 确保资源释放
func ProcessFile(path string) error {
    file, err := os.Open(path)
    if err != nil {
        return err
    }
    defer file.Close()

    // 处理文件
}

// ✅ 正确 - 多个资源按相反顺序关闭
func ProcessData() error {
    db, err := sql.Open(...)
    if err != nil {
        return err
    }
    defer db.Close()

    tx, err := db.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback() // 在 Commit 后无效果
}
```

### 4. 并发安全

```go
// ✅ 正确 - 使用互斥锁保护共享状态
type Counter struct {
    mu    sync.Mutex
    value int
}

func (c *Counter) Increment() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.value++
}

// ✅ 正确 - 使用 RWMutex 提高读性能
type Cache struct {
    mu    sync.RWMutex
    items map[string]string
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.items[key]
    return val, ok
}
```

## 性能优化建议

### 1. 字符串处理

```go
// ✅ 正确 - 使用 strings.Builder
func BuildString(parts []string) string {
    var builder strings.Builder
    builder.Grow(len(parts) * 10) // 预分配
    for _, part := range parts {
        builder.WriteString(part)
    }
    return builder.String()
}

// ❌ 错误 - 频繁字符串拼接
func BuildString(parts []string) string {
    result := ""
    for _, part := range parts {
        result += part // 每次都创建新字符串
    }
    return result
}
```

### 2. 切片预分配

```go
// ✅ 正确 - 预分配切片容量
func ProcessItems(items []Item) []Result {
    results := make([]Result, 0, len(items))
    for _, item := range items {
        results = append(results, process(item))
    }
    return results
}

// ❌ 错误 - 多次重新分配
func ProcessItems(items []Item) []Result {
    var results []Result
    for _, item := range items {
        results = append(results, process(item)) // 多次重新分配
    }
    return results
}
```

### 3. 避免不必要的转换

```go
// ✅ 正确 - 直接使用
if bytes.Contains(data, []byte("pattern")) {
    // ...
}

// ❌ 错误 - 不必要的字符串转换
if strings.Contains(string(data), "pattern") {
    // 可能导致大量内存分配
}
```

## 测试规范

### 1. 测试命名

```go
// ✅ 正确 - 清晰的测试名称
func TestUserRepository_GetUserByID_Success(t *testing.T) { ... }
func TestUserRepository_GetUserByID_NotFound(t *testing.T) { ... }
func TestUserRepository_CreateUser_Duplicate(t *testing.T) { ... }

// ❌ 错误 - 不清晰的命名
func TestUser1(t *testing.T) { ... }
func TestUserError(t *testing.T) { ... }
```

### 2. 表格驱动测试

```go
// ✅ 正确 - 使用表格驱动测试
func TestValidateEmail(t *testing.T) {
    tests := []struct {
        name  string
        email string
        want  bool
    }{
        {"valid email", "user@example.com", true},
        {"invalid format", "invalid", false},
        {"empty string", "", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := ValidateEmail(tt.email)
            if got != tt.want {
                t.Errorf("ValidateEmail() = %v, want %v", got, tt.want)
            }
        })
    }
}
```

### 3. 子测试

```go
// ✅ 正确 - 使用 t.Run 组织测试
func TestProcessFile(t *testing.T) {
    t.Run("success case", func(t *testing.T) {
        // 测试成功情况
    })

    t.Run("file not found", func(t *testing.T) {
        // 测试文件不存在
    })

    t.Run("permission denied", func(t *testing.T) {
        // 测试权限错误
    })
}
```

## 注释规范

### 1. 包注释

```go
// ✅ 正确 - 每个包都应该有包注释
// Package stdlib provides standard library functions for Shode scripts.
//
// This package includes functions for file operations, string manipulation,
// HTTP server management, database operations, and more.
package stdlib
```

### 2. 导出函数注释

```go
// ✅ 正确 - 完整的函数注释
// ReadFile reads the contents of a file and returns it as a string.
//
// It returns an error if the file does not exist or cannot be read.
// The caller is responsible for closing the file.
//
// Parameters:
//   filename - the path to the file to read
//
// Returns:
//   string - the contents of the file
//   error  - an error if the file cannot be read
func (sl *StdLib) ReadFile(filename string) (string, error) {
    // ...
}
```

### 3. 复杂逻辑注释

```go
// ✅ 正确 - 解释为什么这样做
// 使用指数退避算法重试，避免服务器过载
for i := 0; i < maxRetries; i++ {
    err := try()
    if err == nil {
        return nil
    }
    time.Sleep(backoff(i)) // 指数退避
}

// ❌ 错误 - 重复代码逻辑
// 循环 maxRetries 次
for i := 0; i < maxRetries; i++ {
    // ...
}
```

## 安全规范

### 1. 输入验证

```go
// ✅ 正确 - 验证用户输入
func ProcessPath(path string) error {
    // 检查路径遍历攻击
    if strings.Contains(path, "..") {
        return errors.NewSecurityViolation("path traversal detected")
    }

    // 验证路径格式
    if !isValidPath(path) {
        return errors.NewInvalidInput("invalid path format")
    }

    return nil
}
```

### 2. 敏感信息处理

```go
// ✅ 正确 - 不在日志中记录敏感信息
func Login(username, password string) error {
    // 记录用户名但不记录密码
    log.Printf("Login attempt for user: %s", username)

    // ❌ 不要这样做：
    // log.Printf("Login attempt: %s:%s", username, password)
}
```

### 3. SQL 注入防护

```go
// ✅ 正确 - 使用参数化查询
func GetUser(id string) (*User, error) {
    row := db.QueryRow("SELECT * FROM users WHERE id = $1", id)
    // ...
}

// ❌ 错误 - 字符串拼接导致 SQL 注入
func GetUser(id string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE id = '%s'", id)
    row := db.QueryRow(query)
    // ...
}
```

## 依赖管理

### 1. 最小化依赖

```go
// ✅ 优先使用标准库
import "encoding/json"

// ❌ 避免不必要的第三方库
// import "github.com/some/json-lib" // 除非有特殊需求
```

### 2. 依赖版本

```go
// 在 go.mod 中明确指定版本
require (
    github.com/spf13/cobra v1.8.0
    golang.org/x/net v0.20.0
)
```

## 总结

遵循这些规范可以：
- ✅ 提高代码质量和可维护性
- ✅ 减少 bug 和错误
- ✅ 提升性能
- ✅ 增强安全性
- ✅ 改善团队协作

记住：**代码被阅读的次数远多于被编写的次数**。编写清晰、规范的代码是对团队和未来的自己的最好投资。
