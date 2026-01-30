package middleware

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"
)

// LoggingMiddleware 请求日志中间件
type LoggingMiddleware struct {
	*BaseMiddleware
	config *LoggingConfig
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	// LogHeaders 是否记录请求头
	LogHeaders bool
	// LogBody 是否记录请求体
	LogBody bool
	// LogResponse 是否记录响应
	LogResponse bool
	// OutputWriter 日志输出函数
	OutputWriter func(format string, args ...interface{})
}

// DefaultLoggingConfig 默认日志配置
var DefaultLoggingConfig = &LoggingConfig{
	LogHeaders:  false,
	LogBody:     false,
	LogResponse: false,
	OutputWriter: func(format string, args ...interface{}) {
		fmt.Printf("[HTTP] "+format+"\n", args...)
	},
}

// NewLoggingMiddleware 创建日志中间件
func NewLoggingMiddleware(config *LoggingConfig) *LoggingMiddleware {
	if config == nil {
		config = DefaultLoggingConfig
	}

	if config.OutputWriter == nil {
		config.OutputWriter = DefaultLoggingConfig.OutputWriter
	}

	return &LoggingMiddleware{
		BaseMiddleware: NewBaseMiddleware("logging", 300, nil),
		config:        config,
	}
}

// Process 处理日志记录
func (lm *LoggingMiddleware) Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
	// 包装 ResponseWriter 以捕获状态码和响应大小
	wrapped := &responseWriter{ResponseWriter: w, status: 200}

	// 记录请求开始时间
	start := time.Now()

	// 记录请求信息
	lm.logRequest(r)

	// 执行后续中间件和处理器
	next(ctx, wrapped, r)

	// 记录响应信息
	lm.logResponse(r, wrapped, time.Since(start))

	return true
}

// logRequest 记录请求
func (lm *LoggingMiddleware) logRequest(r *http.Request) {
	now := time.Now().Format("2006-01-02 15:04:05")

	reqInfo := fmt.Sprintf("%s %s %s",
		now,
		r.Method,
		r.URL.Path,
	)

	if r.URL.RawQuery != "" {
		reqInfo += fmt.Sprintf("?%s", r.URL.RawQuery)
	}

	if lm.config.LogHeaders {
		reqInfo += fmt.Sprintf("\nHeaders: %v", r.Header)
	}

	if lm.config.LogBody && r.Method != "GET" && r.ContentLength > 0 {
		// 注意：读取 Body 后会消耗，只能记录长度
		reqInfo += fmt.Sprintf("\nBody Length: %d bytes", r.ContentLength)
	}

	lm.config.OutputWriter(reqInfo)
}

// logResponse 记录响应
func (lm *LoggingMiddleware) logResponse(r *http.Request, w *responseWriter, duration time.Duration) {
	now := time.Now().Format("2006-01-02 15:04:05")

	respInfo := fmt.Sprintf("%s | Status: %d | Duration: %v",
		now,
		w.status,
		duration.Round(time.Millisecond),
	)

	if lm.config.LogResponse && w.size > 0 {
		respInfo += fmt.Sprintf(" | Size: %d bytes", w.size)
	}

	lm.config.OutputWriter(respInfo)
}

// responseWriter 包装 http.ResponseWriter 以捕获响应信息
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

func (w *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}
