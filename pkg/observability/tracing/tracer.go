// Package tracing 分布式追踪系统
package tracing

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Config 追踪配置
type Config struct {
	ServiceName string
	Endpoint    string
	Sampler     float64
	Headers     map[string]string
	Environment string
}

// Tracer 追踪器
type Tracer struct {
	serviceName string
	spans       []*Span
	mu          sync.Mutex
}

// Span 追踪 span
type Span struct {
	traceID   string
	spanID    string
	parentID  string
	name      string
	startTime time.Time
	duration  time.Duration
	attrs     map[string]interface{}
	events    []Event
	status    string
}

// Event 事件
type Event struct {
	Time   time.Time
	Name   string
	Attrs  map[string]interface{}
}

// InitTracer 初始化追踪器
func InitTracer(config Config) (*Tracer, error) {
	return &Tracer{
		serviceName: config.ServiceName,
		spans:       make([]*Span, 0),
	}, nil
}

// Close 关闭追踪器
func (t *Tracer) Close() error {
	return nil
}

// StartSpan 启动 span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...SpanOption) (context.Context, *Span) {
	t.mu.Lock()
	defer t.mu.Unlock()

	span := &Span{
		traceID:   generateID(),
		spanID:    generateID(),
		name:      name,
		startTime: time.Now(),
		attrs:     make(map[string]interface{}),
		events:    make([]Event, 0),
		status:    "ok",
	}

	// 应用选项
	for _, opt := range opts {
		opt(span)
	}

	t.spans = append(t.spans, span)
	return context.WithValue(ctx, "span", span), span
}

// SpanOption Span 选项
type SpanOption func(*Span)

// WithAttributes 添加属性
func WithAttributes(attrs map[string]interface{}) SpanOption {
	return func(s *Span) {
		for k, v := range attrs {
			s.attrs[k] = v
		}
	}
}

// End 结束 span
func (s *Span) End() {
	s.duration = time.Since(s.startTime)
}

// SetAttributes 设置属性
func (s *Span) SetAttributes(attrs map[string]interface{}) {
	for k, v := range attrs {
		s.attrs[k] = v
	}
}

// AddEvent 添加事件
func (s *Span) AddEvent(name string, attrs map[string]interface{}) {
	s.events = append(s.events, Event{
		Time:  time.Now(),
		Name:  name,
		Attrs: attrs,
	})
}

// RecordError 记录错误
func (s *Span) RecordError(err error) {
	s.status = "error"
	s.attrs["error"] = err.Error()
}

// SetStatus 设置状态
func (s *Span) SetStatus(status string) {
	s.status = status
}

// HTTPMiddleware HTTP 中间件
func (t *Tracer) HTTPMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// 创建 span
		ctx, span := t.StartSpan(ctx, r.URL.Path,
			WithAttributes(map[string]interface{}{
				"http.method": r.Method,
				"http.url":    r.URL.String(),
			}),
		)
		defer span.End()

		// 更新请求上下文
		r = r.WithContext(ctx)

		// 包装 ResponseWriter
		wrapped := &responseWriter{ResponseWriter: w, status: 200}

		// 调用下一个处理器
		next.ServeHTTP(wrapped, r)

		// 设置 span 属性
		span.SetAttributes(map[string]interface{}{
			"http.status_code": wrapped.status,
		})

		// 设置 span 状态
		if wrapped.status >= 400 {
			span.SetStatus("error")
		}
	})
}

// responseWriter 包装 ResponseWriter
type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

// WrapHandler 包装 HTTP 处理器
func (t *Tracer) WrapHandler(handler http.HandlerFunc) http.Handler {
	return t.HTTPMiddleware(handler)
}

// HTTPClient HTTP 客户端
type HTTPClient struct {
	client *http.Client
	tracer *Tracer
}

// NewHTTPClient 创建 HTTP 客户端
func NewHTTPClient(tracer *Tracer) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{},
		tracer: tracer,
	}
}

// Do 执行 HTTP 请求
func (c *HTTPClient) Do(r *http.Request) (*http.Response, error) {
	ctx := r.Context()

	// 创建 span
	_, span := c.tracer.StartSpan(ctx, "HTTP "+r.Method,
		WithAttributes(map[string]interface{}{
			"http.method": r.Method,
			"http.url":    r.URL.String(),
		}),
	)
	defer span.End()

	// 执行请求
	resp, err := c.client.Do(r)
	if err != nil {
		span.RecordError(err)
		return nil, err
	}

	// 设置 span 属性
	span.SetAttributes(map[string]interface{}{
		"http.status_code": resp.StatusCode,
	})

	if resp.StatusCode >= 400 {
		span.SetStatus("error")
	}

	return resp, nil
}

// SpanBuilder Span 构建器
type SpanBuilder struct {
	tracer    *Tracer
	spanName  string
	ctx       context.Context
	span      *Span
	attrs     map[string]interface{}
}

// NewSpanBuilder 创建 Span 构建器
func NewSpanBuilder(tracer *Tracer, spanName string) *SpanBuilder {
	ctx, span := tracer.StartSpan(context.Background(), spanName)

	return &SpanBuilder{
		tracer:   tracer,
		spanName: spanName,
		ctx:      ctx,
		span:     span,
	}
}

// WithContext 使用上下文
func (b *SpanBuilder) WithContext(ctx context.Context) *SpanBuilder {
	b.ctx, b.span = b.tracer.StartSpan(ctx, b.spanName)
	return b
}

// WithAttributes 添加属性
func (b *SpanBuilder) WithAttributes(attrs map[string]interface{}) *SpanBuilder {
	b.attrs = attrs
	b.span.SetAttributes(attrs)
	return b
}

// WithError 添加错误
func (b *SpanBuilder) WithError(err error) *SpanBuilder {
	b.span.RecordError(err)
	return b
}

// WithEvent 添加事件
func (b *SpanBuilder) WithEvent(name string, attrs map[string]interface{}) *SpanBuilder {
	b.span.AddEvent(name, attrs)
	return b
}

// Build 完成 span 构建
func (b *SpanBuilder) Build() (context.Context, *Span) {
	return b.ctx, b.span
}

// End 结束 span
func (b *SpanBuilder) End() {
	b.span.End()
}

// Context 返回上下文
func (b *SpanBuilder) Context() context.Context {
	return b.ctx
}

// Span 返回 span
func (b *SpanBuilder) Span() *Span {
	return b.span
}

// GetSpanFromContext 从上下文获取 span
func GetSpanFromContext(ctx context.Context) *Span {
	if span, ok := ctx.Value("span").(*Span); ok {
		return span
	}
	return nil
}

// AddEventToContext 向上下文的 span 添加事件
func AddEventToContext(ctx context.Context, name string, attrs map[string]interface{}) {
	span := GetSpanFromContext(ctx)
	if span != nil {
		span.AddEvent(name, attrs)
	}
}

// SetErrorToContext 设置上下文的 span 错误
func SetErrorToContext(ctx context.Context, err error) {
	span := GetSpanFromContext(ctx)
	if span != nil {
		span.RecordError(err)
	}
}

// GetTraceID 获取追踪 ID
func GetTraceID(ctx context.Context) string {
	span := GetSpanFromContext(ctx)
	if span != nil {
		return span.traceID
	}
	return ""
}

// GetSpanID 获取 Span ID
func GetSpanID(ctx context.Context) string {
	span := GetSpanFromContext(ctx)
	if span != nil {
		return span.spanID
	}
	return ""
}

// generateID 生成唯一 ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Baggage 行李
type Baggage struct {
	baggage map[string]string
}

// NewBaggage 创建行李
func NewBaggage() *Baggage {
	return &Baggage{
		baggage: make(map[string]string),
	}
}

// Set 设置键值
func (b *Baggage) Set(key, value string) {
	b.baggage[key] = value
}

// Get 获取值
func (b *Baggage) Get(key string) (string, bool) {
	val, ok := b.baggage[key]
	return val, ok
}

// Delete 删除键
func (b *Baggage) Delete(key string) {
	delete(b.baggage, key)
}

// ToContext 将行李添加到上下文
func (b *Baggage) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, "baggage", b)
}

// FromContext 从上下文获取行李
func FromContext(ctx context.Context) *Baggage {
	if baggage, ok := ctx.Value("baggage").(*Baggage); ok {
		return baggage
	}
	return NewBaggage()
}
