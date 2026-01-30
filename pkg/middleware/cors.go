package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// CORSMiddleware CORS 跨域中间件
type CORSMiddleware struct {
	*BaseMiddleware
	config *CORSConfig
}

// CORSConfig CORS 配置
type CORSConfig struct {
	// AllowedOrigins 允许的来源，* 表示全部
	AllowedOrigins []string
	// AllowedMethods 允许的 HTTP 方法
	AllowedMethods []string
	// AllowedHeaders 允许的请求头
	AllowedHeaders []string
	// ExposedHeaders 暴露的响应头
	ExposedHeaders []string
	// AllowCredentials 是否允许携带凭证
	AllowCredentials bool
	// MaxAge 预检请求缓存时间（秒）
	MaxAge int
}

// DefaultCORSConfig 默认 CORS 配置
var DefaultCORSConfig = &CORSConfig{
	AllowedOrigins:   []string{"*"},
	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
	AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
	ExposedHeaders:   []string{},
	AllowCredentials: false,
	MaxAge:           86400, // 24 小时
}

// NewCORSMiddleware 创建 CORS 中间件
func NewCORSMiddleware(config *CORSConfig) *CORSMiddleware {
	if config == nil {
		config = DefaultCORSConfig
	}

	// 设置默认值
	if len(config.AllowedMethods) == 0 {
		config.AllowedMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	}
	if len(config.AllowedHeaders) == 0 {
		config.AllowedHeaders = []string{"Origin", "Content-Type", "Authorization"}
	}
	if config.MaxAge == 0 {
		config.MaxAge = 86400
	}

	return &CORSMiddleware{
		BaseMiddleware: NewBaseMiddleware("cors", 100, nil),
		config:        config,
	}
}

// Process 处理 CORS 请求
func (c *CORSMiddleware) Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
	origin := r.Header.Get("Origin")

	// 检查是否允许该来源
	if !c.isOriginAllowed(origin) {
		next(ctx, w, r)
		return true
	}

	// 处理预检请求
	if r.Method == http.MethodOptions {
		c.handlePreflight(w, r)
		return false // 不继续执行后续中间件
	}

	// 添加 CORS 响应头到普通请求
	c.setCORSHeaders(w)

	// 继续执行后续中间件
	next(ctx, w, r)
	return true
}

// isOriginAllowed 检查来源是否被允许
func (c *CORSMiddleware) isOriginAllowed(origin string) bool {
	for _, allowed := range c.config.AllowedOrigins {
		if allowed == "*" {
			return true
		}
		if allowed == origin {
			return true
		}
	}
	return false
}

// handlePreflight 处理预检请求
func (c *CORSMiddleware) handlePreflight(w http.ResponseWriter, r *http.Request) {
	// 设置 CORS 响应头
	w.Header().Set("Access-Control-Allow-Origin", r.Header.Get("Origin"))
	w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.config.AllowedMethods, ", "))
	w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.config.AllowedHeaders, ", "))
	w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", c.config.MaxAge))

	if c.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if len(c.config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.config.ExposedHeaders, ", "))
	}

	w.WriteHeader(http.StatusNoContent)
}

// setCORSHeaders 设置 CORS 响应头
func (c *CORSMiddleware) setCORSHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	if len(c.config.AllowedMethods) > 0 {
		w.Header().Set("Access-Control-Allow-Methods", strings.Join(c.config.AllowedMethods, ", "))
	}

	if len(c.config.AllowedHeaders) > 0 {
		w.Header().Set("Access-Control-Allow-Headers", strings.Join(c.config.AllowedHeaders, ", "))
	}

	if c.config.AllowCredentials {
		w.Header().Set("Access-Control-Allow-Credentials", "true")
	}

	if len(c.config.ExposedHeaders) > 0 {
		w.Header().Set("Access-Control-Expose-Headers", strings.Join(c.config.ExposedHeaders, ", "))
	}
}
