# 安全防护系统 (Security Protection System)

Shode 框架提供全面的安全防护解决方案。

## 🔒 功能特性

### 1. CORS 策略配置 (cors/)
- ✅ CORS 策略配置中心
- ✅ 白名单域名管理
- ✅ 允许的 HTTP 方法
- ✅ 允许的请求头
- ✅ 凭证支持
- ✅ 预检请求缓存

### 2. CSRF 防护 (csrf/)
- ✅ Token 生成
- ✅ Token 验证
- ✅ Double Submit Cookie
- ✅ 同步令牌模式
- ✅ 加密存储

### 3. XSS 防护 (xss/)
- ✅ 输入过滤
- ✅ 输出编码
- ✅ Content-Type 策略
- ✅ CSP 头设置
- ✅ HTML 转义

### 4. SQL 注入防护 (sqli/)
- ✅ 参数化查询
- ✅ 输入验证
- ✅ 特殊字符转义
- ✅ ORM 集成
- ✅ 查询白名单

### 5. 速率限制 (ratelimit/)
- ✅ 令牌桶算法
- ✅ 漏桶算法
- ✅ 滑动窗口
- ✅ 固定窗口
- ✅ 分布式限流

### 6. 安全审计 (audit/)
- ✅ 事件记录
- ✅ 异常检测
- ✅ 审计日志
- ✅ 告警通知
- ✅ 合规报告

## 🚀 快速开始

### CORS 配置

```go
import "gitee.com/com_818cloud/shode/pkg/security/cors"

func main() {
    // 创建 CORS 中间件
    corsMiddleware := cors.New(cors.Config{
        AllowedOrigins:   []string{"https://example.com"},
        AllowedMethods:   []string{"GET", "POST", "PUT"},
        AllowedHeaders:   []string{"Content-Type", "Authorization"},
        AllowCredentials: true,
        MaxAge:           3600,
    })

    // 使用中间件
    http.Use(corsMiddleware.Handler())
}
```

### CSRF 防护

```go
import "gitee.com/com_818cloud/shode/pkg/security/csrf"

func main() {
    // 创建 CSRF 中间件
    csrfMiddleware := csrf.New(csrf.Config{
        Secret:       "your-secret-key",
        CookieName:   "csrf_token",
        CookieSecure: true,
        CookieHTTPOnly: true,
        TokenLength:  32,
    })

    // 使用中间件
    http.Use(csrfMiddleware.Handler())
}
```

### XSS 防护

```go
import "gitee.com/com_818cloud/shode/pkg/security/xss"

func main() {
    // 创建 XSS 防护中间件
    xssMiddleware := xss.New(xss.Config{
        EnableInputSanitization: true,
        EnableOutputEncoding:    true,
        EnableCSP:              true,
        CSPDirective:           "default-src 'self'",
    })

    // 使用中间件
    http.Use(xssMiddleware.Handler())
}
```

### SQL 注入防护

```go
import "gitee.com/com_818cloud/shode/pkg/security/sqli"

// 使用参数化查询
func GetUser(db *sql.DB, username string) (*User, error) {
    query := "SELECT * FROM users WHERE username = ?"
    return sqli.Query(db, query, username)
}

// 输入验证
func isValidUsername(username string) bool {
    return sqli.IsValidInput(username, sqli.UsernamePattern)
}
```

### 速率限制

```go
import "gitee.com/com_818cloud/shode/pkg/security/ratelimit"

func main() {
    // 创建速率限制器
    limiter := ratelimit.NewTokenBucket(ratelimit.Config{
        Rate:   100,         // 每秒 100 个请求
        Bucket: 200,         // 桶容量 200
    })

    // 使用中间件
    http.Use(limiter.Handler())
}
```

### 安全审计

```go
import "gitee.com/com_818cloud/shode/pkg/security/audit"

func main() {
    // 创建审计日志器
    auditor := audit.New(audit.Config{
        Output:   []string{"stdout", "/var/log/audit.log"},
        Format:   "json",
        MinLevel: audit.InfoLevel,
    })

    // 记录安全事件
    auditor.Log(audit.Event{
        Type:     "login",
        User:     "user1",
        IP:       "192.168.1.1",
        Success:  true,
        Metadata: map[string]interface{}{"method": "password"},
    })
}
```

## 📋 CORS 策略

### 基础配置

```go
config := cors.Config{
    AllowedOrigins:     []string{"*"},                    // 允许所有源
    AllowedMethods:     []string{"GET", "POST", "PUT"},   // 允许的方法
    AllowedHeaders:     []string{"*"},                    // 允许的请求头
    ExposedHeaders:     []string{"X-Total-Count"},        // 暴露的响应头
    AllowCredentials:   false,                            // 不允许凭证
    MaxAge:             3600,                             // 预检缓存时间
    OptionsPassthrough: false,                            // 不传递 OPTIONS 请求
}
```

### 多域名配置

```go
config := cors.Config{
    AllowedOrigins: []string{
        "https://example.com",
        "https://app.example.com",
        "https://admin.example.com",
    },
}
```

### 动态源配置

```go
config := cors.Config{
    AllowOriginFunc: func(origin string) bool {
        // 从数据库或配置文件检查
        return isAllowedOrigin(origin)
    },
}
```

## 🛡️ CSRF 防护

### Token 生成

```go
import "crypto/rand"

func generateToken() (string, error) {
    b := make([]byte, 32)
    _, err := rand.Read(b)
    if err != nil {
        return "", err
    }
    return hex.EncodeToString(b), nil
}
```

### Token 验证

```go
func validateToken(token string) bool {
    // 从 session 或 cookie 获取期望的 token
    expectedToken := getSessionToken()

    // 使用 constant-time 比较
    return subtle.ConstantTimeCompare(
        []byte(token),
        []byte(expectedToken),
    ) == 1
}
```

### Double Submit Cookie

```go
// 1. 生成 token 并设置 cookie
token := generateToken()
http.SetCookie(w, &http.Cookie{
    Name:     "csrf_token",
    Value:    token,
    Secure:   true,
    HttpOnly: true,
    SameSite: http.SameSiteStrictMode,
})

// 2. 在表单中包含 token
<input type="hidden" name="csrf_token" value="{{.CSRFToken}}">

// 3. 验证 token
if r.FormValue("csrf_token") != getCookieValue(r, "csrf_token") {
    http.Error(w, "Invalid CSRF token", http.StatusForbidden)
    return
}
```

## 🔍 XSS 防护

### 输入过滤

```go
import "regexp"

var scriptPattern = regexp.MustCompile(`<script[^>]*>.*?</script>`)

func sanitizeInput(input string) string {
    // 移除 script 标签
    input = scriptPattern.ReplaceAllString(input, "")

    // 移除事件处理器
    input = regexp.MustCompile(`on\w+\s*=`).ReplaceAllString(input, "")

    return input
}
```

### 输出编码

```go
import "html"

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
    // 自动编码输出
    tmpl.Execute(w, data)
}

// 在模板中使用
{{.Username | html}}  // HTML 编码
{{.Username | url}}   // URL 编码
{{.Username | js}}    // JavaScript 编码
```

### Content Security Policy

```go
func setCSPHeaders(w http.ResponseWriter) {
    w.Header().Set("Content-Security-Policy",
        "default-src 'self'; "+
        "script-src 'self' 'unsafe-inline' 'unsafe-eval'; "+
        "style-src 'self' 'unsafe-inline'; "+
        "img-src 'self' data: https:; "+
        "font-src 'self' data:; "+
        "connect-src 'self'; "+
        "frame-ancestors 'none';")
}
```

## 💉 SQL 注入防护

### 参数化查询

```go
// ✅ 正确 - 使用参数化查询
func getUser(db *sql.DB, username string) (*User, error) {
    var user User
    err := db.QueryRow(
        "SELECT * FROM users WHERE username = ?",
        username,
    ).Scan(&user.ID, &user.Username, &user.Email)
    return &user, err
}

// ❌ 错误 - 字符串拼接
func getUserBad(db *sql.DB, username string) (*User, error) {
    query := fmt.Sprintf("SELECT * FROM users WHERE username = '%s'", username)
    // ...
}
```

### ORM 使用

```go
import "gorm.io/gorm"

func getUser(db *gorm.DB, username string) (*User, error) {
    var user User
    result := db.Where("username = ?", username).First(&user)
    return &user, result.Error
}
```

### 输入验证

```go
func validateInput(input string) bool {
    // 检查长度
    if len(input) < 3 || len(input) > 50 {
        return false
    }

    // 检查字符集
    matched, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, input)
    return matched
}
```

## ⏱️ 速率限制

### 令牌桶算法

```go
type TokenBucket struct {
    rate     float64    // 令牌生成速率
    capacity float64    // 桶容量
    tokens   float64    // 当前令牌数
    lastTime time.Time  // 上次访问时间
    mu       sync.Mutex
}

func (tb *TokenBucket) Allow() bool {
    tb.mu.Lock()
    defer tb.mu.Unlock()

    now := time.Now()
    elapsed := now.Sub(tb.lastTime).Seconds()

    // 添加令牌
    tb.tokens += elapsed * tb.rate
    if tb.tokens > tb.capacity {
        tb.tokens = tb.capacity
    }
    tb.lastTime = now

    // 消费令牌
    if tb.tokens >= 1 {
        tb.tokens--
        return true
    }
    return false
}
```

### 滑动窗口

```go
type SlidingWindow struct {
    window time.Duration
    limit  int
    events []time.Time
    mu     sync.Mutex
}

func (sw *SlidingWindow) Allow() bool {
    sw.mu.Lock()
    defer sw.mu.Unlock()

    now := time.Now()
    cutoff := now.Add(-sw.window)

    // 移除窗口外的事件
    for len(sw.events) > 0 && sw.events[0].Before(cutoff) {
        sw.events = sw.events[1:]
    }

    // 检查是否超过限制
    if len(sw.events) >= sw.limit {
        return false
    }

    sw.events = append(sw.events, now)
    return true
}
```

## 📊 安全审计

### 事件记录

```go
type AuditEvent struct {
    ID        string                 `json:"id"`
    Type      string                 `json:"type"`
    User      string                 `json:"user"`
    IP        string                 `json:"ip"`
    Action    string                 `json:"action"`
    Resource  string                 `json:"resource"`
    Success   bool                   `json:"success"`
    Error     string                 `json:"error,omitempty"`
    Timestamp time.Time              `json:"timestamp"`
    Metadata  map[string]interface{} `json:"metadata,omitempty"`
}
```

### 审计日志

```go
func (a *Auditor) Log(event AuditEvent) error {
    event.ID = generateID()
    event.Timestamp = time.Now()

    // 写入日志
    for _, output := range a.outputs {
        if err := output.Write(event); err != nil {
            return err
        }
    }

    return nil
}
```

### 异常检测

```go
func (a *Auditor) detectAnomalies() {
    // 检测暴力破解
    if a.countFailedLogups(ip) > 5 {
        a.alert("Multiple failed login attempts", ip)
    }

    // 检测异常访问时间
    if hour >= 2 && hour <= 5 && isSensitiveAccess(action) {
        a.alert("Off-hours sensitive access", user)
    }

    // 检测权限提升
    if !user.HasPermission(resource) {
        a.alert("Unauthorized access attempt", user)
    }
}
```

## 🔧 配置选项

### CORS 配置

```go
type CORSConfig struct {
    AllowedOrigins     []string
    AllowedMethods     []string
    AllowedHeaders     []string
    ExposedHeaders     []string
    AllowCredentials   bool
    MaxAge             int
    AllowOriginFunc    func(string) bool
    OptionsPassthrough bool
}
```

### CSRF 配置

```go
type CSRFConfig struct {
    Secret          string
    CookieName      string
    CookieDomain    string
    CookiePath      string
    CookieMaxAge    int
    CookieSecure    bool
    CookieHTTPOnly  bool
    CookieSameSite  http.SameSite
    TokenLength     int
    TokenHeader     string
    FormField       string
}
```

### XSS 配置

```go
type XSSConfig struct {
    EnableInputSanitization bool
    EnableOutputEncoding    bool
    EnableCSP              bool
    CSPDirective           string
    EnableXSSProtection    bool
    EnableContentTypeNosniff bool
}
```

### RateLimit 配置

```go
type RateLimitConfig struct {
    Algorithm string  // "token-bucket", "leaky-bucket", "sliding-window", "fixed-window"
    Rate      float64 // 每秒请求数
    Burst     int     // 突发请求数
    Window    time.Duration
    KeyFunc   func(*http.Request) string
    Store     Store  // 存储后端
}
```

## 📚 最佳实践

1. **分层防护**: 在多个层次应用安全措施
2. **默认拒绝**: 默认拒绝所有访问，明确允许所需访问
3. **最小权限**: 只授予必要的最小权限
4. **纵深防御**: 使用多层安全控制
5. **安全编码**: 遵循安全编码规范
6. **定期审计**: 定期进行安全审计和渗透测试
7. **及时更新**: 及时更新依赖和安全补丁
8. **监控告警**: 建立安全监控和告警机制

## 🤝 贡献

欢迎贡献新的安全防护功能！

## 📄 许可证

MIT License
