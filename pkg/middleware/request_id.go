package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

// RequestIDMiddleware Request ID 中间件
type RequestIDMiddleware struct {
	*BaseMiddleware
	// HeaderName Request ID 请求头名称
	HeaderName string
	// Generator ID 生成函数
	Generator func() string
}

// NewRequestIDMiddleware 创建 Request ID 中间件
func NewRequestIDMiddleware(headerName string) *RequestIDMiddleware {
	if headerName == "" {
		headerName = "X-Request-ID"
	}

	return &RequestIDMiddleware{
		BaseMiddleware: NewBaseMiddleware("request_id", 50, nil),
		HeaderName:   headerName,
		Generator: func() string {
			// 使用 UUID 生成唯一 ID
			return uuid.New().String()
		},
	}
}

// Process 添加 Request ID
func (rim *RequestIDMiddleware) Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
	// 尝试从请求头获取 Request ID
	requestID := r.Header.Get(rim.HeaderName)
	if requestID == "" {
		// 如果没有，生成新的 Request ID
		requestID = rim.Generator()
		// 设置到请求头
		w.Header().Set(rim.HeaderName, requestID)
	}

	// 设置到响应头（用于追踪）
	w.Header().Set(rim.HeaderName, requestID)

	// 将 Request ID 存入 context
	ctx = context.WithValue(ctx, "request_id", requestID)

	// 执行后续中间件
	next(ctx, w, r)
	return true
}

// GetRequestID 从 context 获取 Request ID
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value("request_id").(string); ok {
		return id
	}
	return "unknown"
}

// GetRequestIDInt 从 context 获取 Request ID (数字形式，如果适用)
func GetRequestIDInt(ctx context.Context) int64 {
	idStr := GetRequestID(ctx)
	// 如果是数字字符串，转换为 int64
	if num, err := strconv.ParseInt(idStr, 10, 64); err == nil {
		return num
	}
	return 0
}
