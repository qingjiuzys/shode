// Package middleware 提供 HTTP 中间件框架。
//
// 中间件系统特点：
//   - 基于责任链模式
//   - 优先级排序执行（数字越小越先执行）
//   - 支持跳过条件
//   - 可中断执行链
//
// 内置中间件：
//   - Recovery (优先级 10): Panic 恢复
//   - RequestID (优先级 50): 请求追踪 ID
//   - CORS (优先级 100): 跨域资源共享
//   - RateLimit (优先级 200): 令牌桶限流
//   - Logging (优先级 300): 请求/响应日志
//
// 使用示例：
//
//	mgr := middleware.NewManager()
//	mgr.Use(middleware.NewRecoveryMiddleware())
//	mgr.Use(middleware.NewCORSMiddleware())
//
//	handler := mgr.HandleFunc(ctx, w, r, finalHandler)
//
// 自定义中间件：
//
//	custom := middleware.NewBaseMiddleware("custom", 150, func(ctx, w, r, next) bool {
//	    // 前置逻辑
//	    next(ctx, w, r)
//	    // 后置逻辑
//	    return true
//	})
package middleware

import (
	"context"
	"net/http"
	"sync"
)

// Middleware 中间件接口
// 中间件可以拦截和处理 HTTP 请求，在请求到达最终处理器之前执行逻辑
type Middleware interface {
	// Name 返回中间件名称
	Name() string

	// Priority 返回中间件优先级（数字越小越先执行）
	Priority() int

	// Process 处理请求
	// next: 调用下一个中间件或最终处理器
	// 返回: 是否继续执行后续中间件
	Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool

	// SkipCondition 判断是否跳过此中间件
	// 返回 true 时跳过此中间件的执行
	SkipCondition(r *http.Request) bool
}

// NextFunc 下一个中间件或最终处理器的函数类型
type NextFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request)

// Manager 中间件管理器
type Manager struct {
	middlewares []Middleware
	mu          sync.RWMutex
}

// NewManager 创建中间件管理器
func NewManager() *Manager {
	return &Manager{
		middlewares: make([]Middleware, 0),
	}
}

// Use 注册中间件
func (m *Manager) Use(mw Middleware) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 插入到正确位置（按优先级排序）
	m.middlewares = append(m.middlewares, mw)
	m.sortMiddlewares()

	return m
}

// Remove 移除中间件
func (m *Manager) Remove(name string) *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, mw := range m.middlewares {
		if mw.Name() == name {
			// 删除并保持顺序
			m.middlewares = append(m.middlewares[:i], m.middlewares[i+1:]...)
			break
		}
	}

	return m
}

// sortMiddlewares 按优先级排序中间件
func (m *Manager) sortMiddlewares() {
	// 使用冒泡排序（中间件数量通常不多）
	n := len(m.middlewares)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if m.middlewares[j].Priority() > m.middlewares[j+1].Priority() {
				m.middlewares[j], m.middlewares[j+1] = m.middlewares[j+1], m.middlewares[j]
			}
		}
	}
}

// Execute 执行中间件链
func (m *Manager) Execute(ctx context.Context, w http.ResponseWriter, r *http.Request, finalHandler http.HandlerFunc) {
	m.mu.RLock()
	middlewares := make([]Middleware, len(m.middlewares))
	copy(middlewares, m.middlewares)
	m.mu.RUnlock()

	// 创建执行链
	nextIndex := 0
	var next NextFunc
	next = func(ctx context.Context, w http.ResponseWriter, r *http.Request) {
		if nextIndex >= len(middlewares) {
			// 所有中间件执行完毕，调用最终处理器
			finalHandler(w, r)
			return
		}

		mw := middlewares[nextIndex]
		nextIndex++

		// 检查是否跳过此中间件
		if mw.SkipCondition(r) {
			next(ctx, w, r)
			return
		}

		// 执行中间件
		shouldContinue := mw.Process(ctx, w, r, next)
		if !shouldContinue {
			// 中间件中断了链的执行
			return
		}
	}

	// 开始执行中间件链
	next(ctx, w, r)
}

// GetMiddleware 获取指定名称的中间件
func (m *Manager) GetMiddleware(name string) (Middleware, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, mw := range m.middlewares {
		if mw.Name() == name {
			return mw, true
		}
	}

	return nil, false
}

// List 列出所有中间件
func (m *Manager) List() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, len(m.middlewares))
	for i, mw := range m.middlewares {
		names[i] = mw.Name()
	}

	return names
}

// Clear 清除所有中间件
func (m *Manager) Clear() *Manager {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.middlewares = make([]Middleware, 0)
	return m
}

// BaseMiddleware 基础中间件实现，方便用户自定义
type BaseMiddleware struct {
	name      string
	priority  int
	skipFunc  func(r *http.Request) bool
	processFn func(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool
}

// NewBaseMiddleware 创建基础中间件
func NewBaseMiddleware(name string, priority int, processFn func(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool) *BaseMiddleware {
	return &BaseMiddleware{
		name:      name,
		priority:  priority,
		skipFunc:  func(r *http.Request) bool { return false },
		processFn: processFn,
	}
}

// Name 返回中间件名称
func (bm *BaseMiddleware) Name() string {
	return bm.name
}

// Priority 返回优先级
func (bm *BaseMiddleware) Priority() int {
	return bm.priority
}

// Process 执行中间件逻辑
func (bm *BaseMiddleware) Process(ctx context.Context, w http.ResponseWriter, r *http.Request, next NextFunc) bool {
	return bm.processFn(ctx, w, r, next)
}

// SkipCondition 判断是否跳过
func (bm *BaseMiddleware) SkipCondition(r *http.Request) bool {
	if bm.skipFunc != nil {
		return bm.skipFunc(r)
	}
	return false
}

// SetSkipFunc 设置跳过条件函数
func (bm *BaseMiddleware) SetSkipFunc(fn func(r *http.Request) bool) {
	bm.skipFunc = fn
}

// Chain 链式调用辅助函数
func Chain(middlewares ...Middleware) *Manager {
	m := NewManager()
	for _, mw := range middlewares {
		m.Use(mw)
	}
	return m
}
