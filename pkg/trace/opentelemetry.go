// Package trace 提供分布式追踪功能。
package trace

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// OpenTelemetry OpenTelemetry 接口
type OpenTelemetry interface {
	Tracer(name string) Tracer
	Shutdown(ctx context.Context) error
}

// Tracer 追踪器
type Tracer interface {
	Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span)
}

// Span 跨度
type Span interface {
	End(opts ...SpanEndOption)
	AddEvent(name string, opts ...EventOption)
	SetAttributes(attributes map[string]interface{})
	RecordError(err error)
	SetStatus(code int, message string)
	SpanContext() SpanContext
}

// SpanContext 跨度上下文
type SpanContext struct {
	TraceID string
	SpanID  string
}

// SpanOption 跨度选项
type SpanOption func(*SpanConfig)

// SpanConfig 跨度配置
type SpanConfig struct {
	Attributes  map[string]interface{}
	Links       []Link
	StartTime   time.Time
	Kind        SpanKind
}

// Link 链接
type Link struct {
	Context SpanContext
	Attributes map[string]interface{}
}

// SpanKind 跨度类型
type SpanKind int

const (
	SpanKindInternal SpanKind = iota
	SpanKindServer
	SpanKindClient
	SpanKindProducer
	SpanKindConsumer
)

// SpanEndOption 跨度结束选项
type SpanEndOption func(*SpanEndConfig)

// SpanEndConfig 跨度结束配置
type SpanEndConfig struct {
	EndTime time.Time
}

// EventOption 事件选项
type EventOption func(*EventConfig)

// EventConfig 事件配置
type EventConfig struct {
	Attributes map[string]interface{}
	Timestamp  time.Time
}

// MemoryTracer 内存追踪器
type MemoryTracer struct {
	name    string
	spans   []*RecordedSpan
	mu      sync.RWMutex
	exporter SpanExporter
}

// RecordedSpan 记录的跨度
type RecordedSpan struct {
	Name       string
	SpanContext SpanContext
	Parent     SpanContext
	StartTime  time.Time
	EndTime    time.Time
	Attributes map[string]interface{}
	Events     []*Event
	Status     *Status
	Kind       SpanKind
}

// Event 事件
type Event struct {
	Name       string
	Timestamp  time.Time
	Attributes map[string]interface{}
}

// Status 状态
type Status struct {
	Code    int
	Message string
}

// NewMemoryTracer 创建内存追踪器
func NewMemoryTracer(name string) *MemoryTracer {
	return &MemoryTracer{
		name:  name,
		spans: make([]*RecordedSpan, 0),
	}
}

// Start 开始跨度
func (mt *MemoryTracer) Start(ctx context.Context, spanName string, opts ...SpanOption) (context.Context, Span) {
	config := &SpanConfig{
		Attributes: make(map[string]interface{}),
		StartTime:  time.Now(),
		Kind:       SpanKindInternal,
	}

	for _, opt := range opts {
		opt(config)
	}

	// 获取父级上下文
	var parentContext SpanContext
	if parentSpan := SpanFromContext(ctx); parentSpan != nil {
		parentContext = parentSpan.SpanContext()
	}

	// 创建跨度
	span := &memorySpan{
		tracer:      mt,
		name:        spanName,
		spanContext: SpanContext{
			TraceID: generateTraceID(),
			SpanID:  generateSpanID(),
		},
		parentContext: parentContext,
		startTime:     config.StartTime,
		attributes:    config.Attributes,
		events:        make([]*Event, 0),
		kind:          config.Kind,
	}

	// 记录跨度
	mt.mu.Lock()
	mt.spans = append(mt.spans, &RecordedSpan{
		Name:        spanName,
		SpanContext: span.spanContext,
		Parent:      parentContext,
		StartTime:   config.StartTime,
		Attributes:  config.Attributes,
		Events:      make([]*Event, 0),
		Kind:        config.Kind,
	})
	mt.mu.Unlock()

	// 将跨度添加到上下文
	ctx = ContextWithSpan(ctx, span)

	return ctx, span
}

// memorySpan 内存跨度
type memorySpan struct {
	tracer        *MemoryTracer
	name          string
	spanContext   SpanContext
	parentContext SpanContext
	startTime     time.Time
	endTime       time.Time
	attributes    map[string]interface{}
	events        []*Event
	status        *Status
	kind          SpanKind
	ended         bool
	mu            sync.Mutex
}

// End 结束跨度
func (ms *memorySpan) End(opts ...SpanOption) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.ended {
		return
	}

	ms.endTime = time.Now()
	ms.ended = true

	// 更新记录
	if ms.tracer.exporter != nil {
		ms.tracer.exporter.Export(ms)
	}
}

// AddEvent 添加事件
func (ms *memorySpan) AddEvent(name string, opts ...EventOption) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	config := &EventConfig{
		Timestamp:  time.Now(),
		Attributes: make(map[string]interface{}),
	}

	for _, opt := range opts {
		opt(config)
	}

	event := &Event{
		Name:       name,
		Timestamp:  config.Timestamp,
		Attributes: config.Attributes,
	}

	ms.events = append(ms.events, event)
}

// SetAttributes 设置属性
func (ms *memorySpan) SetAttributes(attributes map[string]interface{}) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	for k, v := range attributes {
		ms.attributes[k] = v
	}
}

// RecordError 记录错误
func (ms *memorySpan) RecordError(err error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.status = &Status{
		Code:    2,
		Message: err.Error(),
	}

	ms.AddEvent("error", WithAttributes(map[string]interface{}{
		"error.message": err.Error(),
		"error.type":    fmt.Sprintf("%T", err),
	}))
}

// SetStatus 设置状态
func (ms *memorySpan) SetStatus(code int, message string) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	ms.status = &Status{
		Code:    code,
		Message: message,
	}
}

// SpanContext 获取跨度上下文
func (ms *memorySpan) SpanContext() SpanContext {
	return ms.spanContext
}

// SpanExporter 跨度导出器
type SpanExporter interface {
	Export(span Span)
	Shutdown(ctx context.Context) error
}

// SetExporter 设置导出器
func (mt *MemoryTracer) SetExporter(exporter SpanExporter) {
	mt.mu.Lock()
	defer mt.mu.Unlock()
	mt.exporter = exporter
}

// GetSpans 获取所有跨度
func (mt *MemoryTracer) GetSpans() []*RecordedSpan {
	mt.mu.RLock()
	defer mt.mu.RUnlock()

	return mt.spans
}

// contextKey 上下文键类型
type contextKey int

const (
	spanKey contextKey = iota
)

// ContextWithSpan 创建带跨度的上下文
func ContextWithSpan(ctx context.Context, span Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}

// SpanFromContext 从上下文获取跨度
func SpanFromContext(ctx context.Context) Span {
	if span, ok := ctx.Value(spanKey).(Span); ok {
		return span
	}
	return nil
}

// generateTraceID 生成 Trace ID
func generateTraceID() string {
	return fmt.Sprintf("trace_%d", time.Now().UnixNano())
}

// generateSpanID 生成 Span ID
func generateSpanID() string {
	return fmt.Sprintf("span_%d", time.Now().UnixNano())
}

// WithAttributes 创建属性选项
func WithAttributes(attributes map[string]interface{}) SpanOption {
	return func(config *SpanConfig) {
		for k, v := range attributes {
			config.Attributes[k] = v
		}
	}
}

// WithStartTime 创建开始时间选项
func WithStartTime(startTime time.Time) SpanOption {
	return func(config *SpanConfig) {
		config.StartTime = startTime
	}
}

// WithKind 创建类型选项
func WithKind(kind SpanKind) SpanOption {
	return func(config *SpanConfig) {
		config.Kind = kind
	}
}

// Propagator 传播器
type Propagator interface {
	Inject(ctx context.Context, carrier interface{}) error
	Extract(ctx context.Context, carrier interface{}) (SpanContext, error)
}

// TraceContextPropagator Trace Context 传播器
type TraceContextPropagator struct{}

// NewTraceContextPropagator 创建 Trace Context 传播器
func NewTraceContextPropagator() *TraceContextPropagator {
	return &TraceContextPropagator{}
}

// Inject 注入
func (tcp *TraceContextPropagator) Inject(ctx context.Context, carrier interface{}) error {
	span := SpanFromContext(ctx)
	if span == nil {
		return nil
	}

	// 简化实现，实际应该写入 carrier
	return nil
}

// Extract 提取
func (tcp *TraceContextPropagator) Extract(ctx context.Context, carrier interface{}) (SpanContext, error) {
	// 简化实现，实际应该从 carrier 读取
	return SpanContext{}, nil
}

// TextMapCarrier 文本映射载体
type TextMapCarrier struct {
	maps map[string]string
}

// NewTextMapCarrier 创建文本映射载体
func NewTextMapCarrier(maps map[string]string) *TextMapCarrier {
	return &TextMapCarrier{maps: maps}
}

// Get 获取值
func (tmc *TextMapCarrier) Get(key string) string {
	return tmc.maps[key]
}

// Set 设置值
func (tmc *TextMapCarrier) Set(key, value string) {
	tmc.maps[key] = value
}

// Keys 获取所有键
func (tmc *TextMapCarrier) Keys() []string {
	keys := make([]string, 0, len(tmc.maps))
	for k := range tmc.maps {
		keys = append(keys, k)
	}
	return keys
}

// TraceConfiguration 追踪配置
type TraceConfiguration struct {
	ServiceName  string
	Sampler      Sampler
	Exporter     SpanExporter
	Propagators  []Propagator
}

// Sampler 采样器
type Sampler interface {
	ShouldSample(traceID string) bool
}

// ProbabilitySampler 概率采样器
type ProbabilitySampler struct {
	probability float64
}

// NewProbabilitySampler 创建概率采样器
func NewProbabilitySampler(probability float64) *ProbabilitySampler {
	return &ProbabilitySampler{probability: probability}
}

// ShouldSample 判断是否采样
func (ps *ProbabilitySampler) ShouldSample(traceID string) bool {
	// 简化实现，总是采样
	return true
}

// AlwaysSampler 总是采样器
type AlwaysSampler struct{}

// NewAlwaysSampler 创建总是采样器
func NewAlwaysSampler() *AlwaysSampler {
	return &AlwaysSampler{}
}

// ShouldSample 总是采样
func (as *AlwaysSampler) ShouldSample(traceID string) bool {
	return true
}

// NeverSampler 从不采样器
type NeverSampler struct{}

// NewNeverSampler 创建从不采样器
func NewNeverSampler() *NeverSampler {
	return &NeverSampler{}
}

// ShouldSample 从不采样
func (ns *NeverSampler) ShouldSample(traceID string) bool {
	return false
}

// TraceIdGenerator Trace ID 生成器
type TraceIdGenerator interface {
	NewTraceID() string
	NewSpanID() string
}

// DefaultTraceIdGenerator 默认 Trace ID 生成器
type DefaultTraceIdGenerator struct{}

// NewDefaultTraceIdGenerator 创建默认 Trace ID 生成器
func NewDefaultTraceIdGenerator() *DefaultTraceIdGenerator {
	return &DefaultTraceIdGenerator{}
}

// NewTraceID 生成 Trace ID
func (dtig *DefaultTraceIdGenerator) NewTraceID() string {
	return generateTraceID()
}

// NewSpanID 生成 Span ID
func (dtig *DefaultTraceIdGenerator) NewSpanID() string {
	return generateSpanID()
}

// Baggage 传播的 baggage
type Baggage struct {
	members map[string]string
	mu      sync.RWMutex
}

// NewBaggage 创建 baggage
func NewBaggage() *Baggage {
	return &Baggage{
		members: make(map[string]string),
	}
}

// SetMember 设置成员
func (b *Baggage) SetMember(key, value string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.members[key] = value
}

// GetMember 获取成员
func (b *Baggage) GetMember(key string) (string, bool) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	value, exists := b.members[key]
	return value, exists
}

// DeleteMember 删除成员
func (b *Baggage) DeleteMember(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.members, key)
}

// Members 获取所有成员
func (b *Baggage) Members() map[string]string {
	b.mu.RLock()
	defer b.mu.RUnlock()

	members := make(map[string]string)
	for k, v := range b.members {
		members[k] = v
	}
	return members
}

// ContextWithBaggage 创建带 baggage 的上下文
func ContextWithBaggage(ctx context.Context, baggage *Baggage) context.Context {
	return context.WithValue(ctx, "baggage", baggage)
}

// BaggageFromContext 从上下文获取 baggage
func BaggageFromContext(ctx context.Context) *Baggage {
	if baggage, ok := ctx.Value("baggage").(*Baggage); ok {
		return baggage
	}
	return nil
}

// PerformanceProfiler 性能分析器
type PerformanceProfiler struct {
	tracer Tracer
}

// NewPerformanceProfiler 创建性能分析器
func NewPerformanceProfiler(tracer Tracer) *PerformanceProfiler {
	return &PerformanceProfiler{tracer: tracer}
}

// ProfileOperation 分析操作
func (pp *PerformanceProfiler) ProfileOperation(ctx context.Context, operationName string, fn func() error) error {
	ctx, span := pp.tracer.Start(ctx, operationName)
	defer span.End()

	err := fn()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(2, err.Error())
	} else {
		span.SetStatus(1, "OK")
	}

	return err
}

// ProfileOperationWithResult 分析操作（带结果）
func (pp *PerformanceProfiler) ProfileOperationWithResult(ctx context.Context, operationName string, fn func() (interface{}, error)) (interface{}, error) {
	ctx, span := pp.tracer.Start(ctx, operationName)
	defer span.End()

	result, err := fn()
	if err != nil {
		span.RecordError(err)
		span.SetStatus(2, err.Error())
		return nil, err
	}

	span.SetStatus(1, "OK")
	span.SetAttributes(map[string]interface{}{
		"result.type": fmt.Sprintf("%T", result),
	})

	return result, nil
}
