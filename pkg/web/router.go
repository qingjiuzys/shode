// Package web 提供 Web 路由功能
package web

import (
	"net/http"
	"sync"
)

// Router HTTP 路由器
type Router struct {
	routes     map[string]map[string]http.HandlerFunc
	middleware []Middleware
	mu         sync.RWMutex
}

// NewRouter 创建路由器
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]map[string]http.HandlerFunc),
	}
}

// Get 注册 GET 路由
func (r *Router) Get(path string, handler http.HandlerFunc) {
	r.addRoute("GET", path, handler)
}

// Post 注册 POST 路由
func (r *Router) Post(path string, handler http.HandlerFunc) {
	r.addRoute("POST", path, handler)
}

// Put 注册 PUT 路由
func (r *Router) Put(path string, handler http.HandlerFunc) {
	r.addRoute("PUT", path, handler)
}

// Delete 注册 DELETE 路由
func (r *Router) Delete(path string, handler http.HandlerFunc) {
	r.addRoute("DELETE", path, handler)
}

// Patch 注册 PATCH 路由
func (r *Router) Patch(path string, handler http.HandlerFunc) {
	r.addRoute("PATCH", path, handler)
}

// addRoute 添加路由
func (r *Router) addRoute(method, path string, handler http.HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.routes[method] == nil {
		r.routes[method] = make(map[string]http.HandlerFunc)
	}
	r.routes[method][path] = handler
}

// Use 添加中间件
func (r *Router) Use(middleware Middleware) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, middleware)
}

// ServeHTTP 实现 http.Handler 接口
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mu.RLock()
	handlers, methodExists := r.routes[req.Method]
	if !methodExists {
		http.NotFound(w, req)
		r.mu.RUnlock()
		return
	}

	handler, pathExists := handlers[req.URL.Path]
	if !pathExists {
		http.NotFound(w, req)
		r.mu.RUnlock()
		return
	}
	r.mu.RUnlock()

	// 应用中间件
	var h http.Handler = http.HandlerFunc(handler)
	for i := len(r.middleware) - 1; i >= 0; i-- {
		h = r.middleware[i](h)
	}

	h.ServeHTTP(w, req)
}

// PathParam 获取路径参数（简化实现）
func PathParam(r *http.Request, key string) string {
	// 简化实现，实际应该从路径中提取参数
	return ""
}
