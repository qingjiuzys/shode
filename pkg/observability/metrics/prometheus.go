// Package metrics Prometheus 指标收集
package metrics

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Registry 指标注册器
type Registry struct {
	metrics map[string]interface{}
	mu      sync.RWMutex
}

// NewRegistry 创建指标注册器
func NewRegistry() *Registry {
	return &Registry{
		metrics: make(map[string]interface{}),
	}
}

// MustRegister 注册指标
func (r *Registry) MustRegister(cols ...interface{}) {
	r.mu.Lock()
	defer r.mu.Unlock()
}

// ServeMetrics 启动指标端点
func (r *Registry) ServeMetrics(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "# Shode Metrics\n")
	})

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	fmt.Printf("📊 Metrics endpoint: http://%s/metrics\n", addr)
	return server.ListenAndServe()
}
// CounterVec 计数器向量
type CounterVec struct {
	counters map[string]*Counter
	mu       sync.RWMutex
}

// NewCounterVec 创建计数器向量
func NewCounterVec() *CounterVec {
	return &CounterVec{
		counters: make(map[string]*Counter),
	}
}

// Inc 增加计数
func (cv *CounterVec) Inc(labels ...string) {
	key := fmt.Sprint(labels)
	cv.mu.Lock()
	defer cv.mu.Unlock()

	if _, ok := cv.counters[key]; !ok {
		cv.counters[key] = &Counter{}
	}
	cv.counters[key].Inc()
}

// Counter 计数器
type Counter struct {
	value float64
	mu    sync.RWMutex
}

// Inc 增加计数
func (c *Counter) Inc() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value++
}

// Add 添加值
func (c *Counter) Add(val float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.value += val
}

// Get 获取值
func (c *Counter) Get() float64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.value
}

// HistogramVec 直方图向量
type HistogramVec struct {
	histograms map[string]*Histogram
	mu         sync.RWMutex
	buckets    []float64
}

// NewHistogramVec 创建直方图向量
func NewHistogramVec(buckets []float64) *HistogramVec {
	return &HistogramVec{
		histograms: make(map[string]*Histogram),
		buckets:    buckets,
	}
}

// Observe 观察值
func (hv *HistogramVec) Observe(value float64, labels ...string) {
	key := fmt.Sprint(labels)
	hv.mu.Lock()
	defer hv.mu.Unlock()

	if _, ok := hv.histograms[key]; !ok {
		hv.histograms[key] = NewHistogram(hv.buckets)
	}
	hv.histograms[key].Observe(value)
}

// Histogram 直方图
type Histogram struct {
	buckets []float64
	counts  map[string]uint64
	sum     float64
	count   uint64
	mu      sync.RWMutex
}

// NewHistogram 创建直方图
func NewHistogram(buckets []float64) *Histogram {
	return &Histogram{
		buckets: buckets,
		counts:  make(map[string]uint64),
	}
}

// Observe 观察值
func (h *Histogram) Observe(value float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.sum += value
	h.count++

	for _, bucket := range h.buckets {
		if value <= bucket {
			key := fmt.Sprintf("le:%.2f", bucket)
			h.counts[key]++
		}
	}
}

// GaugeVec 仪表向量
type GaugeVec struct {
	gauges map[string]*Gauge
	mu     sync.RWMutex
}

// NewGaugeVec 创建仪表向量
func NewGaugeVec() *GaugeVec {
	return &GaugeVec{
		gauges: make(map[string]*Gauge),
	}
}

// Set 设置值
func (gv *GaugeVec) Set(value float64, labels ...string) {
	key := fmt.Sprint(labels)
	gv.mu.Lock()
	defer gv.mu.Unlock()

	if _, ok := gv.gauges[key]; !ok {
		gv.gauges[key] = &Gauge{}
	}
	gv.gauges[key].Set(value)
}

// Inc 增加计数
func (gv *GaugeVec) Inc(labels ...string) {
	key := fmt.Sprint(labels)
	gv.mu.Lock()
	defer gv.mu.Unlock()

	if _, ok := gv.gauges[key]; !ok {
		gv.gauges[key] = &Gauge{}
	}
	gv.gauges[key].Inc()
}

// Dec 减少计数
func (gv *GaugeVec) Dec(labels ...string) {
	key := fmt.Sprint(labels)
	gv.mu.Lock()
	defer gv.mu.Unlock()

	if _, ok := gv.gauges[key]; !ok {
		gv.gauges[key] = &Gauge{}
	}
	gv.gauges[key].Dec()
}

// Gauge 仪表
type Gauge struct {
	value float64
	mu    sync.RWMutex
}

// Set 设置值
func (g *Gauge) Set(value float64) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value = value
}

// Inc 增加计数
func (g *Gauge) Inc() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value++
}

// Dec 减少计数
func (g *Gauge) Dec() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.value--
}

// Get 获取值
func (g *Gauge) Get() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.value
}

// HTTPMetrics HTTP 指标
type HTTPMetrics struct {
	requestsTotal    *CounterVec
	requestDuration  *HistogramVec
	requestSize      *HistogramVec
	responseSize     *HistogramVec
	requestsInFlight *GaugeVec
}

// NewHTTPMetrics 创建 HTTP 指标
func NewHTTPMetrics(namespace string) *HTTPMetrics {
	buckets := []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}
	return &HTTPMetrics{
		requestsTotal:    NewCounterVec(),
		requestDuration:  NewHistogramVec(buckets),
		requestSize:      NewHistogramVec([]float64{100, 1000, 10000, 100000, 1000000}),
		responseSize:     NewHistogramVec([]float64{100, 1000, 10000, 100000, 1000000}),
		requestsInFlight: NewGaugeVec(),
	}
}

// RecordRequest 记录请求
func (m *HTTPMetrics) RecordRequest(method, endpoint, status string, duration time.Duration, reqSize, resSize int) {
	m.requestsTotal.Inc(method, endpoint, status)
	m.requestDuration.Observe(duration.Seconds(), method, endpoint)
	m.requestSize.Observe(float64(reqSize), method, endpoint)
	m.responseSize.Observe(float64(resSize), method, endpoint)
}

// IncInFlight 增加处理中的请求数
func (m *HTTPMetrics) IncInFlight(method, endpoint string) {
	m.requestsInFlight.Inc(method, endpoint)
}

// DecInFlight 减少处理中的请求数
func (m *HTTPMetrics) DecInFlight(method, endpoint string) {
	m.requestsInFlight.Dec(method, endpoint)
}

// DBMetrics 数据库指标
type DBMetrics struct {
	connectionsActive *GaugeVec
	connectionsIdle   *GaugeVec
	queryDuration     *HistogramVec
	queryTotal        *CounterVec
	connectionErrors  *CounterVec
}

// NewDBMetrics 创建数据库指标
func NewDBMetrics(namespace string) *DBMetrics {
	queryBuckets := []float64{0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2}
	return &DBMetrics{
		connectionsActive: NewGaugeVec(),
		connectionsIdle:   NewGaugeVec(),
		queryDuration:     NewHistogramVec(queryBuckets),
		queryTotal:        NewCounterVec(),
		connectionErrors:  NewCounterVec(),
	}
}

// RecordQuery 记录查询
func (m *DBMetrics) RecordQuery(database, operation string, duration time.Duration, success bool) {
	status := "success"
	if !success {
		status = "error"
	}

	m.queryTotal.Inc(database, operation, status)
	m.queryDuration.Observe(duration.Seconds(), database, operation)
}

// UpdateConnections 更新连接数
func (m *DBMetrics) UpdateConnections(database string, active, idle int) {
	m.connectionsActive.Set(float64(active), database)
	m.connectionsIdle.Set(float64(idle), database)
}

// CacheMetrics 缓存指标
type CacheMetrics struct {
	hits        *CounterVec
	misses      *CounterVec
	setTotal    *CounterVec
	deleteTotal *CounterVec
	duration    *HistogramVec
	evictions   *CounterVec
}

// NewCacheMetrics 创建缓存指标
func NewCacheMetrics(namespace string) *CacheMetrics {
	cacheBuckets := []float64{0.0001, 0.0002, 0.0005, 0.001, 0.002, 0.005, 0.01, 0.02, 0.05, 0.1}
	return &CacheMetrics{
		hits:        NewCounterVec(),
		misses:      NewCounterVec(),
		setTotal:    NewCounterVec(),
		deleteTotal: NewCounterVec(),
		duration:    NewHistogramVec(cacheBuckets),
		evictions:   NewCounterVec(),
	}
}

// RecordHit 记录缓存命中
func (m *CacheMetrics) RecordHit(cache, cacheType string) {
	m.hits.Inc(cache, cacheType)
}

// RecordMiss 记录缓存未命中
func (m *CacheMetrics) RecordMiss(cache, cacheType string) {
	m.misses.Inc(cache, cacheType)
}

// RecordSet 记录缓存设置
func (m *CacheMetrics) RecordSet(cache string, duration time.Duration) {
	m.setTotal.Inc(cache)
	m.duration.Observe(duration.Seconds(), cache, "set")
}

// RecordDelete 记录缓存删除
func (m *CacheMetrics) RecordDelete(cache string) {
	m.deleteTotal.Inc(cache)
}

// RecordEviction 记录缓存驱逐
func (m *CacheMetrics) RecordEviction(cache string) {
	m.evictions.Inc(cache)
}

// Middleware HTTP 中间件
type Middleware struct {
	metrics *HTTPMetrics
}

// NewMiddleware 创建中间件
func NewMiddleware(metrics *HTTPMetrics) *Middleware {
	return &Middleware{metrics: metrics}
}

// Wrap 包装 HTTP 处理器
func (m *Middleware) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// 包装 ResponseWriter 以获取状态码和大小
		wrapped := &responseWriter{ResponseWriter: w, status: 200}

		// 增加处理中的请求数
		m.metrics.IncInFlight(r.Method, r.URL.Path)
		defer m.metrics.DecInFlight(r.Method, r.URL.Path)

		// 调用下一个处理器
		next.ServeHTTP(wrapped, r)

		// 记录指标
		duration := time.Since(start)
		m.metrics.RecordRequest(
			r.Method,
			r.URL.Path,
			fmt.Sprintf("%d", wrapped.status),
			duration,
			int(r.ContentLength),
			wrapped.size,
		)
	})
}

// responseWriter 包装 ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// CustomCounter 自定义计数器
type CustomCounter struct {
	counter *CounterVec
}

// NewCustomCounter 创建自定义计数器
func NewCustomCounter(name, help string, labels []string) *CustomCounter {
	return &CustomCounter{
		counter: NewCounterVec(),
	}
}

// Inc 增加计数
func (c *CustomCounter) Inc(labelValues ...string) {
	c.counter.Inc(labelValues...)
}

// Add 增加指定值
func (c *CustomCounter) Add(value float64, labelValues ...string) {
	c.counter.Inc(labelValues...)
}

// CustomGauge 自定义仪表
type CustomGauge struct {
	gauge *GaugeVec
}

// NewCustomGauge 创建自定义仪表
func NewCustomGauge(name, help string, labels []string) *CustomGauge {
	return &CustomGauge{
		gauge: NewGaugeVec(),
	}
}

// Set 设置值
func (g *CustomGauge) Set(value float64, labelValues ...string) {
	g.gauge.Set(value, labelValues...)
}

// Inc 增加计数
func (g *CustomGauge) Inc(labelValues ...string) {
	g.gauge.Inc(labelValues...)
}

// Dec 减少计数
func (g *CustomGauge) Dec(labelValues ...string) {
	g.gauge.Dec(labelValues...)
}

// CustomHistogram 自定义直方图
type CustomHistogram struct {
	histogram *HistogramVec
}

// NewCustomHistogram 创建自定义直方图
func NewCustomHistogram(name, help string, labels []string, buckets []float64) *CustomHistogram {
	return &CustomHistogram{
		histogram: NewHistogramVec(buckets),
	}
}

// Observe 观察值
func (h *CustomHistogram) Observe(value float64, labelValues ...string) {
	h.histogram.Observe(value, labelValues...)
}

// Timer 计时器
type Timer struct {
	start    time.Time
	histogram *CustomHistogram
	labels   []string
}

// NewTimer 创建计时器
func NewTimer(histogram *CustomHistogram, labels ...string) *Timer {
	return &Timer{
		start:    time.Now(),
		histogram: histogram,
		labels:   labels,
	}
}

// ObserveDuration 观察持续时间
func (t *Timer) ObserveDuration() {
	duration := time.Since(t.start)
	t.histogram.Observe(duration.Seconds(), t.labels...)
}

// ContextTimer 上下文计时器
type ContextTimer struct {
	histogram *CustomHistogram
}

// NewContextTimer 创建上下文计时器
func NewContextTimer(histogram *CustomHistogram) *ContextTimer {
	return &ContextTimer{histogram: histogram}
}

// Start 启动计时器
func (ct *ContextTimer) Start(ctx context.Context, labels ...string) func() {
	start := time.Now()
	return func() {
		duration := time.Since(start)
		ct.histogram.Observe(duration.Seconds(), labels...)
	}
}
