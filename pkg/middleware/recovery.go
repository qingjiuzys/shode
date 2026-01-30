package middleware

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
)

// RecoveryMiddleware 错误恢复中间件
type RecoveryMiddleware struct {
	*BaseMiddleware
	// StackTrace 是否输出堆栈跟踪
	StackTrace bool
	// ErrorHandler 自定义错误处理函数
	ErrorHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request, err interface{})
}

// NewRecoveryMiddleware 创建恢复中间件
func NewRecoveryMiddleware(errorHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request, err interface{})) *RecoveryMiddleware {
	return &RecoveryMiddleware{
		BaseMiddleware: NewBaseMiddleware("recovery", 10, nil), // 最高优先级，最先执行
		StackTrace:    true,
		ErrorHandler:  errorHandler,
	}
}

// Process 处理 panic 恢复
func (rm *RecoveryMiddleware) Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) (result bool) {
	// 使用 defer 捕获 panic
	defer func() {
		if err := recover(); err != nil {
			rm.handleError(ctx, w, r, err)
			result = false
		}
	}()

	// 执行后续中间件
	next(ctx, w, r)
	return true
}

// handleError 处理错误
func (rm *RecoveryMiddleware) handleError(ctx context.Context, w http.ResponseWriter, r *http.Request, err interface{}) {
	// 设置响应头
	w.Header().Set("Content-Type", "application/json")

	// 如果有自定义错误处理器，使用它
	if rm.ErrorHandler != nil {
		rm.ErrorHandler(ctx, w, r, err)
		return
	}

	// 默认错误处理
	w.WriteHeader(http.StatusInternalServerError)

	if rm.StackTrace {
		// 输出堆栈跟踪
		stack := debug.Stack()
		fmt.Fprintf(w, `{
			"error": "Internal Server Error",
			"message": "%v",
			"stack": "%s"
		}`, err, stack)
	} else {
		fmt.Fprintf(w, `{
			"error": "Internal Server Error",
			"message": "Something went wrong"
		}`)
	}
}
