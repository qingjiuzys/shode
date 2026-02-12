// Package metrics 提供健康检查和追踪功能。
package metrics

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"
)

// HealthChecker 健康检查器
type HealthChecker struct {
	checks map[string]CheckFunc
	mu     sync.RWMutex
}

// CheckFunc 检查函数
type CheckFunc func() (interface{}, error)

// CheckResult 检查结果
type CheckResult struct {
	Status    string                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Data      map[string]interface{} `json:"data,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// HealthStatus 健康状态
type HealthStatus string

const (
	StatusPassing HealthStatus = "passing"
	StatusWarn   HealthStatus = "warning"
	StatusFailing HealthStatus = "failing"
)

// NewHealthChecker 创建健康检查器
func NewHealthChecker() *HealthChecker {
	return &HealthChecker{
		checks: make(map[string]CheckFunc),
	}
}

// RegisterCheck 注册健康检查
func (hc *HealthChecker) RegisterCheck(name string, check CheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks[name] = check
}

// UnregisterCheck 注销健康检查
func (hc *HealthChecker) UnregisterCheck(name string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	delete(hc.checks, name)
}

// Check 执行所有健康检查
func (hc *HealthChecker) Check() map[string]CheckResult {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	results := make(map[string]CheckResult)
	var overallStatus HealthStatus = StatusPassing

	for name, check := range hc.checks {
		result := CheckResult{
			Timestamp: time.Now().Format(time.RFC3339),
		}

		data, err := check()
		if err != nil {
			result.Status = string(StatusFailing)
			result.Message = err.Error()
			overallStatus = StatusFailing
		} else {
			result.Status = string(StatusPassing)
			result.Data = map[string]interface{}{"data": data}
		}

		results[name] = result
	}

	// 添加整体状态
	results["overall"] = CheckResult{
		Status:    string(overallStatus),
		Timestamp: time.Now().Format(time.RFC3339),
	}

	return results
}

// HTTPHandler 健康检查 HTTP 处理器
func (hc *HealthChecker) HTTPHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		results := hc.Check()

		w.Header().Set("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		encoder.Encode(results)
	}
}

// Tracer 分布式追踪器
type Tracer struct {
	serviceName string
	traces      map[string]*Span
	mu          sync.RWMutex
}

// Span 追踪跨度
type Span struct {
	TraceID   string    `json:"trace_id"`
	SpanID    string    `json:"span_id"`
	ParentID  string    `json:"parent_id,omitempty"`
	Name      string    `json:"name"`
	Start     time.Time `json:"start"`
	Duration  int64     `json:"duration_ms"`
	Tags      map[string]string `json:"tags,omitempty"`
	Logs      []LogEntry `json:"logs,omitempty"`
	Status     string    `json:"status"`
}

// LogEntry 日志条目
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
}

// NewTracer 创建追踪器
func NewTracer(serviceName string) *Tracer {
	return &Tracer{
		serviceName: serviceName,
		traces:      make(map[string]*Span),
	}
}

// StartSpan 开始跨度
func (t *Tracer) StartSpan(traceID, parentID, name string) *Span {
	span := &Span{
		TraceID:  traceID,
		SpanID:   generateSpanID(),
		ParentID:  parentID,
		Name:     name,
		Start:    time.Now(),
		Tags:     make(map[string]string),
		Status:   "started",
		Logs:     make([]LogEntry, 0),
	}

	t.mu.Lock()
	t.traces[span.SpanID] = span
	t.mu.Unlock()

	return span
}

// Finish 完成跨度
func (s *Span) Finish() {
	s.Duration = time.Since(s.Start).Milliseconds()
	s.Status = "completed"
}

// Tag 添加标签
func (s *Span) Tag(key, value string) {
	s.Tags[key] = value
}

// Log 添加日志
func (s *Span) Log(level, message string) {
	s.Logs = append(s.Logs, LogEntry{
		Timestamp: time.Now().Format(time.RFC3339),
		Level:     level,
		Message:   message,
	})
}

// generateSpanID 生成跨度ID
func generateSpanID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// HealthMetrics 健康指标收集器
type HealthMetrics struct {
	StartTime time.Time
}

// NewHealthMetrics 创建健康指标收集器
func NewHealthMetrics() *HealthMetrics {
	return &HealthMetrics{
		StartTime: time.Now(),
	}
}

// GetUptime 获取运行时间
func (m *HealthMetrics) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}

// GetMemoryStats 获取内存统计
func (m *HealthMetrics) GetMemoryStats() runtime.MemStats {
	return runtime.MemStats{}
}

// GetGoroutines 获取 goroutine 数量
func (m *HealthMetrics) GetGoroutines() int {
	return runtime.NumGoroutine()
}

// GetStats 获取所有统计信息
func (m *HealthMetrics) GetStats() map[string]interface{} {
	memStats := m.GetMemoryStats()

	return map[string]interface{}{
		"uptime_seconds":      m.GetUptime().Seconds(),
		"goroutines":          m.GetGoroutines(),
		"memory_alloc":        memStats.Alloc,
		"memory_total_alloc":  memStats.TotalAlloc,
		"memory_sys":          memStats.Sys,
		"memory_heap_alloc":   memStats.HeapAlloc,
		"memory_stack_inuse":  memStats.StackInuse,
		"num_gc":              memStats.NumGC,
	}
}

// ExportPrometheus 导出到 Prometheus 格式
func (m *HealthMetrics) ExportPrometheus() string {
	stats := m.GetStats()

	var output strings.Builder
	output.WriteString(fmt.Sprintf("# HELP app_uptime_seconds Application uptime in seconds\n"))
	output.WriteString(fmt.Sprintf("# TYPE app_uptime_seconds gauge\n"))
	output.WriteString(fmt.Sprintf("app_uptime_seconds %v\n", stats["uptime_seconds"]))

	output.WriteString(fmt.Sprintf("\n# HELP app_goroutines Current number of goroutines\n"))
	output.WriteString(fmt.Sprintf("# TYPE app_goroutines gauge\n"))
	output.WriteString(fmt.Sprintf("app_goroutines %v\n", stats["goroutines"]))

	return output.String()
}

// DefaultMetricsCollector 默认指标收集器
var DefaultMetricsCollector = NewHealthMetrics()
