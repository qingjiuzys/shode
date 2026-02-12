// Package cors CORS 跨域资源共享配置
package cors

import (
	"net/http"
	"strconv"
	"strings"
)

// Config CORS 配置
type Config struct {
	AllowedOrigins     []string
	AllowedMethods     []string
	AllowedHeaders     []string
	ExposedHeaders     []string
	AllowCredentials   bool
	MaxAge             int
	AllowOriginFunc    func(string) bool
	OptionsPassthrough bool
	Debug              bool
}

// CORS CORS 中间件
type CORS struct {
	config *Config
}

// New 创建 CORS 中间件
func New(config *Config) *CORS {
	return &CORS{config: config}
}

// Handler 返回 HTTP 处理器
func (c *CORS) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// 检查是否允许该源
		if !c.isAllowedOrigin(origin) {
			if c.config.OptionsPassthrough && r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		// 设置 CORS 头
		c.setHeaders(w, r, origin)

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 继续处理请求
		w.WriteHeader(http.StatusOK)
	})
}

// isAllowedOrigin 检查源是否允许
func (c *CORS) isAllowedOrigin(origin string) bool {
	if origin == "" {
		return true
	}

	// 检查自定义函数
	if c.config.AllowOriginFunc != nil {
		return c.config.AllowOriginFunc(origin)
	}

	// 检查白名单
	for _, allowed := range c.config.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		if strings.HasSuffix(allowed, "*") {
			prefix := strings.TrimSuffix(allowed, "*")
			if strings.HasPrefix(origin, prefix) {
				return true
			}
		}
	}

	return false
}

// setHeaders 设置 CORS 响应头
func (c *CORS) setHeaders(w http.ResponseWriter, r *http.Request, origin string) {
	// Access-Control-Allow-Origin
	if c.config.AllowCredentials || len(c.config.AllowedOrigins) == 1 {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	} else if len(c.config.AllowedOrigins) == 1 && c.config.AllowedOrigins[0] == "*" {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	}

	// Access-Control-Allow-Methods
	if len(c.config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods",
			strings.Join(c.config.AllowedMethods, ", "))
	}

	// Access-Control-Allow-Headers
	if len(c.config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers",
			strings.Join(c.config.AllowedHeaders, ", "))
	}

	// Access-Control-Expose-Headers
	if len(c.config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers",
			strings.Join(c.config.ExposedHeaders, ", "))
	}

	// Access-Control-Allow-Credentials
	if c.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	// Access-Control-Max-Age
	if c.config.MaxAge > 0 {
		w.Header().Set("Access-Control-Max-Age",
			strconv.Itoa(c.config.MaxAge))
	}

	if c.config.Debug {
		c.logRequest(r)
	}
}

// Middleware 返回中间件函数
func (c *CORS) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// 检查是否允许该源
		if !c.isAllowedOrigin(origin) {
			if c.config.OptionsPassthrough && r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}
			http.Error(w, "Origin not allowed", http.StatusForbidden)
			return
		}

		// 设置 CORS 头
		c.setHeaders(w, r, origin)

		// 处理预检请求
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// 继续处理请求
		next.ServeHTTP(w, r)
	})
}

// logRequest 记录请求
func (c *CORS) logRequest(r *http.Request) {
	println("[CORS]",
		"Method:", r.Method,
		"Origin:", r.Header.Get("Origin"),
		"Path:", r.URL.Path)
}

// Default 默认配置
func Default() *CORS {
	return New(&Config{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"*"},
		ExposedHeaders:   []string{},
		AllowCredentials: false,
		MaxAge:           86400,
	})
}
