// Package xss XSS 跨站脚本防护
package xss

import (
	"html"
	"net/http"
	"regexp"
	"strings"
)

// Config XSS 配置
type Config struct {
	EnableInputSanitization  bool
	EnableOutputEncoding     bool
	EnableCSP               bool
	CSPDirective            string
	EnableXSSProtection     bool
	EnableContentTypeNosniff bool
}

// XSS XSS 防护中间件
type XSS struct {
	config *Config
}

// New 创建 XSS 中间件
func New(config *Config) *XSS {
	return &XSS{config: config}
}

// Middleware 返回中间件函数
func (x *XSS) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 设置安全头
		x.setHeaders(w)

		// 清理输入
		if x.config.EnableInputSanitization {
			x.sanitizeRequest(r)
		}

		// 包装 ResponseWriter
		wrapped := &responseWriter{ResponseWriter: w, encode: x.config.EnableOutputEncoding}

		next.ServeHTTP(wrapped, r)
	})
}

// setHeaders 设置安全响应头
func (x *XSS) setHeaders(w http.ResponseWriter) {
	// X-XSS-Protection
	if x.config.EnableXSSProtection {
		w.Header().Set("X-XSS-Protection", "1; mode=block")
	}

	// X-Content-Type-Options
	if x.config.EnableContentTypeNosniff {
		w.Header().Set("X-Content-Type-Options", "nosniff")
	}

	// Content-Security-Policy
	if x.config.EnableCSP && x.config.CSPDirective != "" {
		w.Header().Set("Content-Security-Policy", x.config.CSPDirective)
	}
}

// sanitizeRequest 清理请求输入
func (x *XSS) sanitizeRequest(r *http.Request) {
	// 清理查询参数
	x.sanitizeForm(r)
}

// sanitizeForm 清理表单数据
func (x *XSS) sanitizeForm(r *http.Request) {
	// 清理 URL 参数
	for _, values := range r.URL.Query() {
		for i, value := range values {
			values[i] = Sanitize(value)
		}
	}
}

// responseWriter 包装 ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	encode bool
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if w.encode {
		// 对输出进行编码
		encoded := html.EscapeString(string(b))
		return w.ResponseWriter.Write([]byte(encoded))
	}
	return w.ResponseWriter.Write(b)
}

// Sanitize 清理输入
func Sanitize(input string) string {
	// 移除 script 标签
	scriptPattern := regexp.MustCompile(`<script[^>]*>.*?</script>`)
	input = scriptPattern.ReplaceAllString(input, "")

	// 移除事件处理器
	eventPattern := regexp.MustCompile(`on\w+\s*=`)
	input = eventPattern.ReplaceAllString(input, "")

	// 移除 javascript: 协议
	jsPattern := regexp.MustCompile(`(?i)javascript:`)
	input = jsPattern.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}

// Encode 编码输出
func Encode(input string) string {
	return html.EscapeString(input)
}

// EncodeURL 编码 URL
func EncodeURL(input string) string {
	return html.EscapeString(input)
}

// EncodeJS 编码 JavaScript
func EncodeJS(input string) string {
	// 简化实现
	replacer := strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\"", "\\\"",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	)
	return replacer.Replace(input)
}

// Default 默认配置
func Default() *XSS {
	return New(&Config{
		EnableInputSanitization:  true,
		EnableOutputEncoding:     true,
		EnableCSP:               true,
		CSPDirective:            "default-src 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval'; style-src 'self' 'unsafe-inline'",
		EnableXSSProtection:     true,
		EnableContentTypeNosniff: true,
	})
}
