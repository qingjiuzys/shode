// Package metrics 提供 Prometheus 指标收集和导出功能。
package metrics

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// MetricType 指标类型
type MetricType int

const (
	Counter MetricType = iota
	Gauge
	Histogram
	Summary
)

// Metric 指标接口
type Metric interface {
	Name() string
	Type() MetricType
	Help() string
	Collect() string
}

// Counter 计数器
type Counter struct {
	name   string
	help   string
	value  uint64
	labels map[string]string
}

// NewCounter 创建计数器
func NewCounter(name, help string) *Counter {
	return &Counter{
		name:   name,
		help:   help,
		labels: make(map[string]string),
	}
}

// Inc 增加计数
func (c *Counter) Inc() {
	atomic.AddUint64(&c.value, 1)
}

// Add 增加指定值
func (c *Counter) Add(delta uint64) {
	atomic.AddUint64(&c.value, delta)
}

// Get 获取当前值
func (c *Counter) Get() uint64 {
	return atomic.LoadUint64(&c.value)
}

// Name 返回指标名称
func (c *Counter) Name() string {
	return c.name
}

// Type 返回指标类型
func (c *Counter) Type() MetricType {
	return Counter
}

// Help 返回帮助信息
func (c *Counter) Help() string {
	return c.help
}

// Collect 收集指标数据
func (c *Counter) Collect() string {
	return fmt.Sprintf("%s %d", c.name, c.Get())
}

// Gauge 仪表
type Gauge struct {
	name   string
	help   string
	value  int64
	labels map[string]string
}

// NewGauge 创建仪表
func NewGauge(name, help string) *Gauge {
	return &Gauge{
		name:   name,
		help:   help,
		labels: make(map[string]string),
	}
}

// Set 设置值
func (g *Gauge) Set(value int64) {
	atomic.StoreInt64(&g.value, value)
}

// Inc 增加
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec 减少
func (g *Gauge) Dec() {
	g.Add(-1)
}

// Add 增加指定值
func (g *Gauge) Add(delta int64) {
	atomic.AddInt64(&g.value, delta)
}

// Get 获取当前值
func (g *Gauge) Get() int64 {
	return atomic.LoadInt64(&g.value)
}

// Name 返回指标名称
func (g *Gauge) Name() string {
	return g.name
}

// Type 返回指标类型
func (g *Gauge) Type() MetricType {
	return Gauge
}

// Help 返回帮助信息
func (g *Gauge) Help() string {
	return g.help
}

// Collect 收集指标数据
func (g *Gauge) Collect() string {
	return fmt.Sprintf("%s %d", g.name, g.Get())
}

// Histogram 直方图
type Histogram struct {
	name   string
	help   string
	buckets []float64
	sum    uint64
	count  uint64
	labels map[string]string
}

// NewHistogram 创建直方图
func NewHistogram(name, help string, buckets []float64) *Histogram {
	return &Histogram{
		name:    name,
		help:    help,
		buckets: buckets,
		labels:  make(map[string]string),
	}
}

// Observe 观察值
func (h *Histogram) Observe(value float64) {
	atomic.AddUint64(&h.sum, uint64(value))
	atomic.AddUint64(&h.count, 1)

	// TODO: 记录到对应的 bucket
}

// Name 返回指标名称
func (h *Histogram) Name() string {
	return h.name
}

// Type 返回指标类型
func (h *Histogram) Type() MetricType {
	return Histogram
}

// Help 返回帮助信息
func (h *Histogram) Help() string {
	return h.help
}

// Collect 收集指标数据
func (h *Histogram) Collect() string {
	return fmt.Sprintf("%s_sum %d %s_count %d", h.name, h.sum, h.name, h.count)
}

// Registry 指标注册表
type Registry struct {
	metrics map[string]Metric
	mu      sync.RWMutex
}

// NewRegistry 创建注册表
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]Metric),
	}
}

// Register 注册指标
func (r *Registry) Register(metric Metric) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metrics[metric.Name()] = metric
}

// Unregister 注销指标
func (r *Registry) Unregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.metrics, name)
}

// Get 获取指标
func (r *Registry) Get(name string) (Metric, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	metric, exists := r.metrics[name]
	return metric, exists
}

// Counter 获取或创建计数器
func (r *Registry) Counter(name, help string) *Counter {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m, exists := r.metrics[name]; exists {
		if c, ok := m.(*Counter); ok {
			return c
		}
	}

	c := NewCounter(name, help)
	r.metrics[name] = c
	return c
}

// Gauge 获取或创建仪表
func (r *Registry) Gauge(name, help string) *Gauge {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m, exists := r.metrics[name]; exists {
		if g, ok := m.(*Gauge); ok {
			return g
		}
	}

	g := NewGauge(name, help)
	r.metrics[name] = g
	return g
}

// Histogram 获取或创建直方图
func (r *Registry) Histogram(name, help string, buckets []float64) *Histogram {
	r.mu.Lock()
	defer r.mu.Unlock()

	if m, exists := r.metrics[name]; exists {
		if h, ok := m.(*Histogram); ok {
			return h
		}
	}

	h := NewHistogram(name, help, buckets)
	r.metrics[name] = h
	return h
}

// GetAll 获取所有指标
func (r *Registry) GetAll() []Metric {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Metric, 0, len(r.metrics))
	for _, m := range r.metrics {
		result = append(result, m)
	}
	return result
}

// HTTPHandler 指标 HTTP 处理器
type HTTPHandler struct {
	registry *Registry
}

// NewHTTPHandler 创建 HTTP 处理器
func NewHTTPHandler(registry *Registry) *HTTPHandler {
	return &HTTPHandler{
		registry: registry,
	}
}

// ServeHTTP 处理 HTTP 请求
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 支持 Prometheus 文本格式
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")

	metrics := h.registry.GetAll()
	for _, metric := range metrics {
		w.Write([]byte(metric.Collect() + "\n"))
	}
}

// DefaultMetrics 默认指标
var DefaultMetrics = struct {
	HTTPRequestsTotal    *Counter
	HTTPRequestDuration   *Histogram
	ActiveConnections    *Gauge
	MemoryUsage          *Gauge
}{
	HTTPRequestsTotal:   NewCounter("http_requests_total", "Total HTTP requests"),
	HTTPRequestDuration:  NewHistogram("http_request_duration_seconds", "HTTP request duration", []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}),
	ActiveConnections:   NewGauge("active_connections", "Active connections"),
	MemoryUsage:         NewGauge("memory_usage_bytes", "Memory usage in bytes"),
}

// DefaultRegistry 默认注册表
var DefaultRegistry = NewRegistry()

func init() {
	DefaultRegistry.Register(DefaultMetrics.HTTPRequestsTotal)
	DefaultRegistry.Register(DefaultMetrics.HTTPRequestDuration)
	DefaultRegistry.Register(DefaultMetrics.ActiveConnections)
	DefaultMetrics.MemoryUsage
}
