// Package web 提供 Web 服务器功能
package web

import (
	"fmt"
	"net/http"
)

// Server HTTP 服务器
type Server struct {
	addr    string
	handler http.Handler
}

// NewServer 创建服务器
func NewServer(addr string) *Server {
	return &Server{
		addr: addr,
	}
}

// SetHandler 设置处理器
func (s *Server) SetHandler(h http.Handler) {
	s.handler = h
}

// Start 启动服务器
func (s *Server) Start() error {
	if s.handler == nil {
		return fmt.Errorf("no handler set")
	}
	return http.ListenAndServe(s.addr, s.handler)
}
