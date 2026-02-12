package performance

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"sync"
	"time"

	"gitee.com/com_818cloud/shode/pkg/trace"
)

// MetricsExporter 指标导出器
type MetricsExporter struct {
	collectors   []MetricsCollector
	exporters    []MetricsBackend
	registry     *MetricsRegistry
	interval     time.Duration
	enabled      bool
	mu           sync.RWMutex
}

// MetricsCollector 指标收集器接口
type MetricsCollector interface {
	Collect() *MetricCollection
	Name() string
}

// MetricsBackend 指标后端接口
type MetricsBackend interface {
	Export(ctx context.Context, collection *MetricCollection) error
	Name() string
}

// MetricCollection 指标集合
type MetricCollection struct {
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	Metrics     map[string]*Metric     `json:"metrics"`
	Source      string                 `json:"source"`
}

// Metric 指标
type Metric struct {
	Name      string        `json:"name"`
	Type      string        `json:"type"` // "gauge", "counter", "histogram", "summary"
	Value     float64       `json:"value"`
	Labels    map[string]string `json:"labels,omitempty"`
	Timestamp time.Time     `json:"timestamp"`
	Histogram *Histogram    `json:"histogram,omitempty"`
}

// Histogram 直方图
type Histogram struct {
	SampleCount int64              `json:"sample_count"`
	SampleSum   float64            `json:"sample_sum"`
	Buckets     map[string]int64   `json:"buckets"`
}

// MetricsRegistry 指标注册表
type MetricsRegistry struct {
	metrics    map[string]*MetricDefinition
	instances  map[string][]*Metric
	mu         sync.RWMutex
}

// MetricDefinition 指标定义
type MetricDefinition struct {
	Name        string
	Type        string
	Description string
	Labels      []string
	Help        string
}

// NewMetricsExporter 创建指标导出器
func NewMetricsExporter(interval time.Duration) *MetricsExporter {
	exporter := &MetricsExporter{
		collectors: make([]MetricsCollector, 0),
		exporters:  make([]MetricsBackend, 0),
		registry:   NewMetricsRegistry(),
		interval:   interval,
		enabled:    true,
	}

	// 注册默认收集器
	exporter.RegisterCollector(&RuntimeMetricsCollector{})
	exporter.RegisterCollector(&PerformanceMetricsCollector{})
	exporter.RegisterCollector(&BusinessMetricsCollector{})

	return exporter
}

// NewMetricsRegistry 创建指标注册表
func NewMetricsRegistry() *MetricsRegistry {
	return &MetricsRegistry{
		metrics:   make(map[string]*MetricDefinition),
		instances: make(map[string][]*Metric),
	}
}

// Start 启动指标导出
func (me *MetricsExporter) Start(ctx context.Context) error {
	if !me.enabled {
		return nil
	}

	// 启动定期收集和导出
	go me.run(ctx)

	return nil
}

// Stop 停止指标导出
func (me *MetricsExporter) Stop() {
	me.enabled = false
}

// run 运行收集和导出
func (me *MetricsExporter) run(ctx context.Context) {
	ticker := time.NewTicker(me.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			me.collectAndExport(ctx)
		case <-ctx.Done():
			return
		}
	}
}

// collectAndExport 收集并导出指标
func (me *MetricsExporter) collectAndExport(ctx context.Context) {
	// 收集所有收集器的指标
	collection := &MetricCollection{
		Timestamp: time.Now(),
		Labels:    make(map[string]string),
		Metrics:   make(map[string]*Metric),
		Source:    "shode_performance",
	}

	// 从各个收集器收集指标
	for _, collector := range me.collectors {
		collected := collector.Collect()
		// 合并指标
		for name, metric := range collected.Metrics {
			collection.Metrics[name] = metric
		}
	}

	// 添加注册表的指标
	me.registry.mu.RLock()
	for name, metrics := range me.registry.instances {
		if len(metrics) > 0 {
			collection.Metrics[name] = metrics[len(metrics)-1] // 使用最新的
		}
	}
	me.registry.mu.RUnlock()

	// 导出到各个后端
	for _, backend := range me.exporters {
		if err := backend.Export(ctx, collection); err != nil {
			// 记录错误但继续尝试其他后端
			fmt.Printf("Export to %s failed: %v\n", backend.Name(), err)
		}
	}
}

// RegisterCollector 注册收集器
func (me *MetricsExporter) RegisterCollector(collector MetricsCollector) {
	me.mu.Lock()
	defer me.mu.Unlock()

	me.collectors = append(me.collectors, collector)
}

// RegisterBackend 注册后端
func (me *MetricsExporter) RegisterBackend(backend MetricsBackend) {
	me.mu.Lock()
	defer me.mu.Unlock()

	me.exporters = append(me.exporters, backend)
}

// RuntimeMetricsCollector 运行时指标收集器
type RuntimeMetricsCollector struct {
	lastGCStats uint64
}

// Collect 收集指标
func (rmc *RuntimeMetricsCollector) Collect() *MetricCollection {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	collection := &MetricCollection{
		Timestamp: time.Now(),
		Metrics:   make(map[string]*Metric),
	}

	// 内存指标
	collection.Metrics["memory_heap_alloc"] = &Metric{
		Name:  "memory_heap_alloc",
		Type:  "gauge",
		Value: float64(m.HeapAlloc),
		Labels: map[string]string{"unit": "bytes"},
	}

	collection.Metrics["memory_heap_sys"] = &Metric{
		Name:  "memory_heap_sys",
		Type:  "gauge",
		Value: float64(m.HeapSys),
		Labels: map[string]string{"unit": "bytes"},
	}

	collection.Metrics["memory_stack_inuse"] = &Metric{
		Name:  "memory_stack_inuse",
		Type:  "gauge",
		Value: float64(m.StackInuse),
		Labels: map[string]string{"unit": "bytes"},
	}

	collection.Metrics["gc_count"] = &Metric{
		Name:  "gc_count",
		Type:  "counter",
		Value: float64(m.NumGC),
	}

	collection.Metrics["gc_pause_total"] = &Metric{
		Name:  "gc_pause_total",
		Type:  "gauge",
		Value: float64(m.PauseTotalNs) / 1e6, // 转换为毫秒
		Labels: map[string]string{"unit": "ms"},
	}

	// Goroutine指标
	collection.Metrics["goroutine_count"] = &Metric{
		Name:  "goroutine_count",
		Type:  "gauge",
		Value: float64(runtime.NumGoroutine()),
	}

	// CPU指标
	collection.Metrics["cpu_count"] = &Metric{
		Name:  "cpu_count",
		Type:  "gauge",
		Value: float64(runtime.NumCPU()),
	}

	// GC指标
	if uint64(m.NumGC) > rmc.lastGCStats {
		rmc.lastGCStats = uint64(m.NumGC)
		collection.Metrics["gc_last_pause"] = &Metric{
			Name:  "gc_last_pause",
			Type:  "gauge",
			Value: float64(m.PauseNs[(m.NumGC+255)%256]) / 1e6,
			Labels: map[string]string{"unit": "ms"},
		}
	}

	return collection
}

// Name 返回收集器名称
func (rmc *RuntimeMetricsCollector) Name() string {
	return "runtime_metrics"
}

// PerformanceMetricsCollector 性能指标收集器
type PerformanceMetricsCollector struct {
	metrics map[string]*PerformanceMetrics
	mu      sync.RWMutex
}

// Collect 收集指标
func (pmc *PerformanceMetricsCollector) Collect() *MetricCollection {
	pmc.mu.RLock()
	defer pmc.mu.RUnlock()

	collection := &MetricCollection{
		Timestamp: time.Now(),
		Metrics:   make(map[string]*Metric),
	}

	for name, perf := range pmc.metrics {
		prefix := fmt.Sprintf("performance_%s_", name)

		collection.Metrics[prefix+"execution_time"] = &Metric{
			Name:  prefix + "execution_time",
			Type:  "gauge",
			Value: float64(perf.ExecutionTime.Microseconds()) / 1000, // 毫秒
			Labels: map[string]string{"unit": "ms"},
		}

		collection.Metrics[prefix+"memory_usage"] = &Metric{
			Name:  prefix + "memory_usage",
			Type:  "gauge",
			Value: float64(perf.MemoryUsage),
			Labels: map[string]string{"unit": "bytes"},
		}

		collection.Metrics[prefix+"cpu_usage"] = &Metric{
			Name:  prefix + "cpu_usage",
			Type:  "gauge",
			Value: perf.CPUUsage,
			Labels: map[string]string{"unit": "percent"},
		}

		collection.Metrics[prefix+"throughput"] = &Metric{
			Name:  prefix + "throughput",
			Type:  "gauge",
			Value: perf.Throughput,
			Labels: map[string]string{"unit": "ops_per_sec"},
		}

		collection.Metrics[prefix+"latency"] = &Metric{
			Name:  prefix + "latency",
			Type:  "gauge",
			Value: float64(perf.Latency.Microseconds()) / 1000,
			Labels: map[string]string{"unit": "ms"},
		}

		collection.Metrics[prefix+"error_rate"] = &Metric{
			Name:  prefix + "error_rate",
			Type:  "gauge",
			Value: perf.ErrorRate * 100,
			Labels: map[string]string{"unit": "percent"},
		}

		collection.Metrics[prefix+"cache_hit_rate"] = &Metric{
			Name:  prefix + "cache_hit_rate",
			Type:  "gauge",
			Value: perf.CacheHitRate * 100,
			Labels: map[string]string{"unit": "percent"},
		}
	}

	return collection
}

// Name 返回收集器名称
func (pmc *PerformanceMetricsCollector) Name() string {
	return "performance_metrics"
}

// SetMetrics 设置性能指标
func (pmc *PerformanceMetricsCollector) SetMetrics(name string, metrics *PerformanceMetrics) {
	pmc.mu.Lock()
	defer pmc.mu.Unlock()

	if pmc.metrics == nil {
		pmc.metrics = make(map[string]*PerformanceMetrics)
	}

	pmc.metrics[name] = metrics
}

// BusinessMetricsCollector 业务指标收集器
type BusinessMetricsCollector struct {
	counters   map[string]int64
	histograms map[string]*HistogramData
	mu         sync.RWMutex
}

// HistogramData 直方图数据
type HistogramData struct {
	Values []float64
	Buckets []float64
}

// Collect 收集指标
func (bmc *BusinessMetricsCollector) Collect() *MetricCollection {
	bmc.mu.RLock()
	defer bmc.mu.RUnlock()

	collection := &MetricCollection{
		Timestamp: time.Now(),
		Metrics:   make(map[string]*Metric),
	}

	// 导出计数器
	for name, value := range bmc.counters {
		collection.Metrics["business_"+name] = &Metric{
			Name:  "business_" + name,
			Type:  "counter",
			Value: float64(value),
		}
	}

	// 导出直方图
	for name, data := range bmc.histograms {
		histogram := &Histogram{
			SampleCount: int64(len(data.Values)),
			Buckets:     make(map[string]int64),
		}

		// 计算总和
		for _, v := range data.Values {
			histogram.SampleSum += v
		}

		// 填充桶
		sort.Float64s(data.Values)
		for _, bucket := range data.Buckets {
			count := int64(0)
			for _, v := range data.Values {
				if v <= bucket {
					count++
				}
			}
			histogram.Buckets[fmt.Sprintf("%.2f", bucket)] = count
		}

		collection.Metrics["business_"+name] = &Metric{
			Name:      "business_" + name,
			Type:      "histogram",
			Histogram: histogram,
		}
	}

	return collection
}

// Name 返回收集器名称
func (bmc *BusinessMetricsCollector) Name() string {
	return "business_metrics"
}

// IncrementCounter 递增计数器
func (bmc *BusinessMetricsCollector) IncrementCounter(name string, value int64) {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	if bmc.counters == nil {
		bmc.counters = make(map[string]int64)
	}

	bmc.counters[name] += value
}

// RecordValue 记录值到直方图
func (bmc *BusinessMetricsCollector) RecordValue(name string, value float64, buckets []float64) {
	bmc.mu.Lock()
	defer bmc.mu.Unlock()

	if bmc.histograms == nil {
		bmc.histograms = make(map[string]*HistogramData)
	}

	data, exists := bmc.histograms[name]
	if !exists {
		data = &HistogramData{
			Values: make([]float64, 0),
			Buckets: buckets,
		}
		bmc.histograms[name] = data
	}

	data.Values = append(data.Values, value)
}

// PrometheusBackend Prometheus后端
type PrometheusBackend struct {
	address     string
	namespace   string
	httpClient  *http.Client
}

// NewPrometheusBackend 创建Prometheus后端
func NewPrometheusBackend(address, namespace string) *PrometheusBackend {
	return &PrometheusBackend{
		address:   address,
		namespace: namespace,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// Export 导出到Prometheus
func (pb *PrometheusBackend) Export(ctx context.Context, collection *MetricCollection) error {
	// 将指标转换为Prometheus格式
	_ = pb.convertToPrometheusFormat(collection)

	// 发送到Prometheus Pushgateway
	req, err := http.NewRequestWithContext(ctx, "POST", pb.address+"/metrics/job/shode", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "text/plain")
	// TODO: 实际发送数据

	_, err = pb.httpClient.Do(req)
	return err
}

// convertToPrometheusFormat 转换为Prometheus格式
func (pb *PrometheusBackend) convertToPrometheusFormat(collection *MetricCollection) string {
	// 简化实现
	return ""
}

// Name 返回后端名称
func (pb *PrometheusBackend) Name() string {
	return "prometheus"
}

// OpenTelemetryBackend OpenTelemetry后端
type OpenTelemetryBackend struct {
	tracer     trace.Tracer
	serviceName string
}

// NewOpenTelemetryBackend 创建OpenTelemetry后端
func NewOpenTelemetryBackend(serviceName string, tracer trace.Tracer) *OpenTelemetryBackend {
	return &OpenTelemetryBackend{
		tracer:      tracer,
		serviceName: serviceName,
	}
}

// Export 导出到OpenTelemetry
func (otb *OpenTelemetryBackend) Export(ctx context.Context, collection *MetricCollection) error {
	// 创建指标span
	_, span := otb.tracer.Start(ctx, "metrics_export")
	defer span.End()

	// 添加指标作为span属性
	attrs := make(map[string]interface{})
	for name, metric := range collection.Metrics {
		attrs[name] = metric.Value
	}
	span.SetAttributes(attrs)

	return nil
}

// Name 返回后端名称
func (otb *OpenTelemetryBackend) Name() string {
	return "opentelemetry"
}

// JSONBackend JSON文件后端
type JSONBackend struct {
	filePath   string
	httpServer *http.Server
	mu         sync.RWMutex
}

// NewJSONBackend 创建JSON后端
func NewJSONBackend(filePath string, port int) *JSONBackend {
	jb := &JSONBackend{
		filePath: filePath,
	}

	// 启动HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", jb.handleMetricsRequest)
	mux.HandleFunc("/health", jb.handleHealthRequest)

	jb.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go jb.httpServer.ListenAndServe()

	return jb
}

// Export 导出到JSON文件
func (jb *JSONBackend) Export(ctx context.Context, collection *MetricCollection) error {
	jb.mu.Lock()
	defer jb.mu.Unlock()

	_, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return err
	}

	// 写入文件（简化实现）
	_ = jb.filePath
	return nil
}

// handleMetricsRequest 处理指标请求
func (jb *JSONBackend) handleMetricsRequest(w http.ResponseWriter, r *http.Request) {
	jb.mu.RLock()
	defer jb.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\"status\":\"ok\"}"))
}

// handleHealthRequest 处理健康检查请求
func (jb *JSONBackend) handleHealthRequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

// Name 返回后端名称
func (jb *JSONBackend) Name() string {
	return "json"
}

// RegisterMetric 注册指标
func (mr *MetricsRegistry) RegisterMetric(definition *MetricDefinition) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	mr.metrics[definition.Name] = definition
}

// RecordGauge 记录仪表值
func (mr *MetricsRegistry) RecordGauge(name string, value float64, labels map[string]string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	metric := &Metric{
		Name:      name,
		Type:      "gauge",
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	if _, exists := mr.instances[name]; !exists {
		mr.instances[name] = make([]*Metric, 0)
	}

	mr.instances[name] = append(mr.instances[name], metric)
}

// IncrementCounter 递增计数器
func (mr *MetricsRegistry) IncrementCounter(name string, value float64, labels map[string]string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, exists := mr.instances[name]; !exists {
		mr.instances[name] = make([]*Metric, 0)
	}

	// 获取当前值
	currentValue := 0.0
	if len(mr.instances[name]) > 0 {
		currentValue = mr.instances[name][len(mr.instances[name])-1].Value
	}

	metric := &Metric{
		Name:      name,
		Type:      "counter",
		Value:     currentValue + value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	mr.instances[name] = append(mr.instances[name], metric)
}

// RecordHistogram 记录直方图
func (mr *MetricsRegistry) RecordHistogram(name string, value float64, labels map[string]string) {
	mr.mu.Lock()
	defer mr.mu.Unlock()

	if _, exists := mr.instances[name]; !exists {
		mr.instances[name] = make([]*Metric, 0)
	}

	// 获取或创建直方图
	var histogram *Histogram
	if len(mr.instances[name]) > 0 {
		lastMetric := mr.instances[name][len(mr.instances[name])-1]
		if lastMetric.Histogram != nil {
			histogram = lastMetric.Histogram
		}
	}

	if histogram == nil {
		histogram = &Histogram{
			Buckets: make(map[string]int64),
		}
	}

	histogram.SampleCount++
	histogram.SampleSum += value

	metric := &Metric{
		Name:      name,
		Type:      "histogram",
		Histogram: histogram,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	mr.instances[name] = append(mr.instances[name], metric)
}

// GetMetrics 获取所有指标
func (mr *MetricsRegistry) GetMetrics() map[string]*MetricDefinition {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	definitions := make(map[string]*MetricDefinition, len(mr.metrics))
	for k, v := range mr.metrics {
		definitions[k] = v
	}

	return definitions
}

// GetMetricInstances 获取指标实例
func (mr *MetricsRegistry) GetMetricInstances(name string) []*Metric {
	mr.mu.RLock()
	defer mr.mu.RUnlock()

	instances, exists := mr.instances[name]
	if !exists {
		return []*Metric{}
	}

	result := make([]*Metric, len(instances))
	copy(result, instances)
	return result
}

// ExportMetricsJSON 导出指标为JSON
func (me *MetricsExporter) ExportMetricsJSON() (string, error) {
	collection := &MetricCollection{
		Timestamp: time.Now(),
		Labels:    make(map[string]string),
		Metrics:   make(map[string]*Metric),
		Source:    "shode_performance",
	}

	// 从各个收集器收集指标
	for _, collector := range me.collectors {
		collected := collector.Collect()
		for name, metric := range collected.Metrics {
			collection.Metrics[name] = metric
		}
	}

	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetMetricsSummary 获取指标摘要
func (me *MetricsExporter) GetMetricsSummary() map[string]interface{} {
	summary := make(map[string]interface{})

	summary["collectors"] = len(me.collectors)
	summary["exporters"] = len(me.exporters)
	summary["enabled"] = me.enabled
	summary["interval"] = me.interval.String()

	collectorNames := make([]string, len(me.collectors))
	for i, c := range me.collectors {
		collectorNames[i] = c.Name()
	}

	exporterNames := make([]string, len(me.exporters))
	for i, e := range me.exporters {
		exporterNames[i] = e.Name()
	}

	summary["collector_names"] = collectorNames
	summary["exporter_names"] = exporterNames

	return summary
}
