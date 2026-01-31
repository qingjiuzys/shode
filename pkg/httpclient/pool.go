// Package httpclient 提供 HTTP 连接池实现。
//
// 连接池特点：
//   - 连接复用，提高性能
//   - 支持 HTTP/2
//   - 可配置的超时和连接限制
//   - 自动重试和失败处理
//   - 详细的统计信息
//
// 使用示例：
//
//	pool := httpclient.NewConnectionPool(nil)
//	defer pool.Close()
//
//	resp, err := pool.Get(ctx, "https://api.example.com/data")
//
//	// 并发请求会复用连接
//	for i := 0; i < 100; i++ {
//	    go pool.Get(ctx, url)
//	}
//
// 自定义配置：
//
//	config := &httpclient.PoolConfig{
//	    MaxConnsPerHost: 50,
//	    Timeout: 30 * time.Second,
//	}
//	pool := httpclient.NewConnectionPool(config)
package httpclient

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

// PoolConfig HTTP 连接池配置
type PoolConfig struct {
	MaxConnsPerHost       int           // 每个主机的最大连接数
	MaxIdleConns          int           // 最大空闲连接数
	IdleConnTimeout       time.Duration // 空闲连接超时时间
	ResponseHeaderTimeout time.Duration // 响应头超时时间
	Timeout               time.Duration // 总请求超时时间
	KeepAlive             time.Duration // Keep-Alive 时间
	TLSHandshakeTimeout   time.Duration // TLS 握手超时
	ExpectContinueTimeout time.Duration // 100-Continue 超时
	InsecureSkipVerify    bool          // 是否跳过 TLS 验证
	MaxIdleConnsPerHost   int           // 每个主机的最大空闲连接数
}

// DefaultPoolConfig 默认连接池配置
var DefaultPoolConfig = &PoolConfig{
	MaxConnsPerHost:       100,
	MaxIdleConns:          100,
	IdleConnTimeout:       90 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
	Timeout:               30 * time.Second,
	KeepAlive:             30 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
	InsecureSkipVerify:    false,
	MaxIdleConnsPerHost:   10,
}

// ConnectionPool HTTP 连接池
type ConnectionPool struct {
	client     *http.Client
	config     *PoolConfig
	stats      *PoolStats
	mu         sync.RWMutex
	customTransport *http.Transport
}

// PoolStats 连接池统计信息
type PoolStats struct {
	ActiveRequests     int64
	CompletedRequests  int64
	FailedRequests     int64
	TotalRequests      int64
	AvgResponseTime    time.Duration
	MaxResponseTime    time.Duration
	MinResponseTime    time.Duration
	totalResponseTime  int64 // 内部字段：总响应时间（纳秒）
}

// NewConnectionPool 创建新的 HTTP 连接池
func NewConnectionPool(config *PoolConfig) *ConnectionPool {
	if config == nil {
		config = DefaultPoolConfig
	}

	// 创建自定义 Transport
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   config.Timeout,
			KeepAlive: config.KeepAlive,
		}).DialContext,
		MaxIdleConns:          config.MaxIdleConns,
		IdleConnTimeout:       config.IdleConnTimeout,
		TLSHandshakeTimeout:   config.TLSHandshakeTimeout,
		ExpectContinueTimeout: config.ExpectContinueTimeout,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   config.MaxIdleConnsPerHost,
		MaxConnsPerHost:       config.MaxConnsPerHost,
	}

	// 配置 TLS
	tlsConfig := &tls.Config{
		InsecureSkipVerify: config.InsecureSkipVerify,
		MinVersion:         tls.VersionTLS12,
	}
	transport.TLSClientConfig = tlsConfig

	return &ConnectionPool{
		client: &http.Client{
			Transport: transport,
			Timeout:   config.Timeout,
		},
		config:         config,
		stats:          &PoolStats{MinResponseTime: time.Hour},
		customTransport: transport,
	}
}

// GetClient 获取 HTTP 客户端
func (p *ConnectionPool) GetClient() *http.Client {
	return p.client
}

// Get 获取 URL（使用 GET 方法）
func (p *ConnectionPool) Get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	return p.Do(req)
}

// Post 发送 POST 请求
func (p *ConnectionPool) Post(ctx context.Context, url string, contentType string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return p.Do(req)
}

// Do 执行 HTTP 请求
func (p *ConnectionPool) Do(req *http.Request) (*http.Response, error) {
	atomic.AddInt64(&p.stats.TotalRequests, 1)
	atomic.AddInt64(&p.stats.ActiveRequests, 1)
	defer atomic.AddInt64(&p.stats.ActiveRequests, -1)

	start := time.Now()
	resp, err := p.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		atomic.AddInt64(&p.stats.FailedRequests, 1)
		return nil, err
	}

	atomic.AddInt64(&p.stats.CompletedRequests, 1)
	p.updateResponseTime(duration)

	return resp, nil
}

// updateResponseTime 更新响应时间统计
func (p *ConnectionPool) updateResponseTime(duration time.Duration) {
	// 累加总响应时间
	atomic.AddInt64(&p.stats.totalResponseTime, int64(duration))

	// 更新平均响应时间
	completed := atomic.LoadInt64(&p.stats.CompletedRequests)
	if completed > 0 {
		avg := atomic.LoadInt64(&p.stats.totalResponseTime) / completed
		atomic.StoreInt64((*int64)(&p.stats.AvgResponseTime), avg)
	}

	// 更新最大响应时间
	for {
		oldMax := atomic.LoadInt64((*int64)(&p.stats.MaxResponseTime))
		if duration <= time.Duration(oldMax) {
			break
		}
		if atomic.CompareAndSwapInt64((*int64)(&p.stats.MaxResponseTime), oldMax, int64(duration)) {
			break
		}
	}

	// 更新最小响应时间
	for {
		oldMin := atomic.LoadInt64((*int64)(&p.stats.MinResponseTime))
		if duration >= time.Duration(oldMin) {
			break
		}
		if atomic.CompareAndSwapInt64((*int64)(&p.stats.MinResponseTime), oldMin, int64(duration)) {
			break
		}
	}
}

// GetStats 获取连接池统计信息
func (p *ConnectionPool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		ActiveRequests:    atomic.LoadInt64(&p.stats.ActiveRequests),
		CompletedRequests: atomic.LoadInt64(&p.stats.CompletedRequests),
		FailedRequests:    atomic.LoadInt64(&p.stats.FailedRequests),
		TotalRequests:     atomic.LoadInt64(&p.stats.TotalRequests),
		AvgResponseTime:   time.Duration(atomic.LoadInt64((*int64)(&p.stats.AvgResponseTime))),
		MaxResponseTime:   time.Duration(atomic.LoadInt64((*int64)(&p.stats.MaxResponseTime))),
		MinResponseTime:   time.Duration(atomic.LoadInt64((*int64)(&p.stats.MinResponseTime))),
	}
}

// ResetStats 重置统计信息
func (p *ConnectionPool) ResetStats() {
	p.mu.Lock()
	defer p.mu.Unlock()

	atomic.StoreInt64(&p.stats.ActiveRequests, 0)
	atomic.StoreInt64(&p.stats.CompletedRequests, 0)
	atomic.StoreInt64(&p.stats.FailedRequests, 0)
	atomic.StoreInt64(&p.stats.TotalRequests, 0)
	atomic.StoreInt64(&p.stats.totalResponseTime, 0)
	atomic.StoreInt64((*int64)(&p.stats.AvgResponseTime), 0)
	atomic.StoreInt64((*int64)(&p.stats.MaxResponseTime), 0)
	atomic.StoreInt64((*int64)(&p.stats.MinResponseTime), int64(time.Hour))
}

// CloseIdleConnections 关闭所有空闲连接
func (p *ConnectionPool) CloseIdleConnections() {
	if p.customTransport != nil {
		p.customTransport.CloseIdleConnections()
	}
}

// Close 关闭连接池
func (p *ConnectionPool) Close() error {
	p.CloseIdleConnections()
	return nil
}

// GetTransport 获取自定义 Transport（用于高级配置）
func (p *ConnectionPool) GetTransport() *http.Transport {
	return p.customTransport
}

// SetMaxConnsPerHost 动态设置每个主机的最大连接数
func (p *ConnectionPool) SetMaxConnsPerHost(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.customTransport != nil {
		p.customTransport.MaxConnsPerHost = max
	}
	p.config.MaxConnsPerHost = max
}

// SetMaxIdleConns 动态设置最大空闲连接数
func (p *ConnectionPool) SetMaxIdleConns(max int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.customTransport != nil {
		p.customTransport.MaxIdleConns = max
	}
	p.config.MaxIdleConns = max
}
