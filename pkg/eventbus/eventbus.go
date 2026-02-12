// Package eventbus 提供事件总线功能
package eventbus

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"time"
)

// Event 事件接口
type Event interface {
	// Name 事件名称
	Name() string
	// Data 事件数据
	Data() any
}

// BaseEvent 基础事件
type BaseEvent struct {
	EventName string
	Payload   any
	Timestamp time.Time
	Metadata  map[string]any
}

func (e *BaseEvent) Name() string {
	return e.EventName
}

func (e *BaseEvent) Data() any {
	return e.Payload
}

// NewEvent 创建事件
func NewEvent(name string, data any) *BaseEvent {
	return &BaseEvent{
		EventName: name,
		Payload:   data,
		Timestamp: time.Now(),
		Metadata:  make(map[string]any),
	}
}

// EventHandler 事件处理器
type EventHandler func(ctx context.Context, event Event) error

// EventSubscription 事件订阅
type EventSubscription struct {
	ID       string
	Topic    string
	Handler  EventHandler
	Filter   func(Event) bool
	Once     bool
	Async    bool
}

// EventBus 事件总线
type EventBus struct {
	mu            sync.RWMutex
	subscriptions map[string][]*EventSubscription
	handlers      map[string][]EventHandler
	queue         chan Event
	running       bool
	wg            sync.WaitGroup
	stopChan      chan struct{}
}

// NewEventBus 创建事件总线
func NewEventBus() *EventBus {
	return &EventBus{
		subscriptions: make(map[string][]*EventSubscription),
		handlers:      make(map[string][]EventHandler),
		queue:         make(chan Event, 1000),
		stopChan:      make(chan struct{}),
	}
}

// Subscribe 订阅事件
func (eb *EventBus) Subscribe(topic string, handler EventHandler) string {
	return eb.SubscribeWithOptions(topic, handler, nil, false, false)
}

// SubscribeWithOptions 订阅事件（带选项）
func (eb *EventBus) SubscribeWithOptions(topic string, handler EventHandler, filter func(Event) bool, once, async bool) string {
	sub := &EventSubscription{
		ID:      generateID(),
		Topic:   topic,
		Handler: handler,
		Filter:  filter,
		Once:    once,
		Async:   async,
	}

	eb.mu.Lock()
	eb.subscriptions[topic] = append(eb.subscriptions[topic], sub)
	eb.mu.Unlock()

	return sub.ID
}

// SubscribeOnce 订阅一次性事件
func (eb *EventBus) SubscribeOnce(topic string, handler EventHandler) string {
	return eb.SubscribeWithOptions(topic, handler, nil, true, false)
}

// SubscribeAsync 异步订阅事件
func (eb *EventBus) SubscribeAsync(topic string, handler EventHandler) string {
	return eb.SubscribeWithOptions(topic, handler, nil, false, true)
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(topic string, subscriptionID string) {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	subs, ok := eb.subscriptions[topic]
	if !ok {
		return
	}

	filtered := make([]*EventSubscription, 0, len(subs))
	for _, sub := range subs {
		if sub.ID != subscriptionID {
			filtered = append(filtered, sub)
		}
	}

	if len(filtered) > 0 {
		eb.subscriptions[topic] = filtered
	} else {
		delete(eb.subscriptions, topic)
	}
}

// Publish 发布事件（同步）
func (eb *EventBus) Publish(ctx context.Context, event Event) error {
	if event == nil {
		return fmt.Errorf("event cannot be nil")
	}

	topic := event.Name()

	eb.mu.RLock()
	subs, ok := eb.subscriptions[topic]
	eb.mu.RUnlock()

	if !ok || len(subs) == 0 {
		return nil
	}

	var errs []error
	var toRemove []string

	for _, sub := range subs {
		// 检查过滤器
		if sub.Filter != nil && !sub.Filter(event) {
			continue
		}

		// 执行处理器
		if sub.Async {
			// 异步执行
			eb.wg.Add(1)
			go func(s *EventSubscription) {
				defer eb.wg.Done()
				_ = s.Handler(ctx, event)
			}(sub)
		} else {
			// 同步执行
			if err := sub.Handler(ctx, event); err != nil {
				errs = append(errs, err)
			}
		}

		// 标记一次性订阅为待删除
		if sub.Once {
			toRemove = append(toRemove, sub.ID)
		}
	}

	// 移除一次性订阅
	for _, id := range toRemove {
		eb.Unsubscribe(topic, id)
	}

	if len(errs) > 0 {
		return fmt.Errorf("publish errors: %v", errs)
	}

	return nil
}

// PublishAsync 异步发布事件
func (eb *EventBus) PublishAsync(event Event) {
	go func() {
		ctx := context.Background()
		_ = eb.Publish(ctx, event)
	}()
}

// PublishPublish 发布事件（简单版本）
func (eb *EventBus) PublishPublish(topic string, data any) error {
	event := NewEvent(topic, data)
	ctx := context.Background()
	return eb.Publish(ctx, event)
}

// HasSubscribers 检查是否有订阅者
func (eb *EventBus) HasSubscribers(topic string) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs, ok := eb.subscriptions[topic]
	return ok && len(subs) > 0
}

// SubscriberCount 订阅者数量
func (eb *EventBus) SubscriberCount(topic string) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	subs, ok := eb.subscriptions[topic]
	if !ok {
		return 0
	}
	return len(subs)
}

// Clear 清除所有订阅
func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	eb.subscriptions = make(map[string][]*EventSubscription)
	eb.handlers = make(map[string][]EventHandler)
}

// Topics 获取所有主题
func (eb *EventBus) Topics() []string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	topics := make([]string, 0, len(eb.subscriptions))
	for topic := range eb.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

// Start 启动事件总线
func (eb *EventBus) Start() {
	eb.mu.Lock()
	if eb.running {
		eb.mu.Unlock()
		return
	}
	eb.running = true
	eb.mu.Unlock()

	eb.wg.Add(1)
	go eb.processQueue()
}

// Stop 停止事件总线
func (eb *EventBus) Stop() {
	eb.mu.Lock()
	if !eb.running {
		eb.mu.Unlock()
		return
	}
	eb.running = false
	eb.mu.Unlock()

	close(eb.stopChan)
	eb.wg.Wait()
	eb.stopChan = make(chan struct{})
}

// processQueue 处理事件队列
func (eb *EventBus) processQueue() {
	defer eb.wg.Done()

	for {
		select {
		case event := <-eb.queue:
			ctx := context.Background()
			_ = eb.Publish(ctx, event)
		case <-eb.stopChan:
			return
		}
	}
}

// generateID 生成订阅ID
func generateID() string {
	return fmt.Sprintf("sub-%d", time.Now().UnixNano())
}

// EventMatcher 事件匹配器
type EventMatcher interface {
	Match(event Event) bool
}

// TypeMatcher 类型匹配器
type TypeMatcher struct {
	eventType reflect.Type
}

func NewTypeMatcher(eventType any) *TypeMatcher {
	return &TypeMatcher{
		eventType: reflect.TypeOf(eventType),
	}
}

func (tm *TypeMatcher) Match(event Event) bool {
	if event == nil {
		return false
	}
	data := event.Data()
	if data == nil {
		return false
	}
	return reflect.TypeOf(data) == tm.eventType
}

// NameMatcher 名称匹配器
type NameMatcher struct {
	pattern string
}

func NewNameMatcher(pattern string) *NameMatcher {
	return &NameMatcher{pattern: pattern}
}

func (nm *NameMatcher) Match(event Event) bool {
	if event == nil {
		return false
	}
	return event.Name() == nm.pattern
}

// PrefixMatcher 前缀匹配器
type PrefixMatcher struct {
	prefix string
}

func NewPrefixMatcher(prefix string) *PrefixMatcher {
	return &PrefixMatcher{prefix: prefix}
}

func (pm *PrefixMatcher) Match(event Event) bool {
	if event == nil {
		return false
	}
	name := event.Name()
	return len(name) >= len(pm.prefix) && name[:len(pm.prefix)] == pm.prefix
}

// CompositeEventBus 复合事件总线
type CompositeEventBus struct {
	buses []*EventBus
}

func NewCompositeEventBus(buses ...*EventBus) *CompositeEventBus {
	return &CompositeEventBus{
		buses: buses,
	}
}

func (ceb *CompositeEventBus) Publish(ctx context.Context, event Event) error {
	var lastErr error
	for _, bus := range ceb.buses {
		if err := bus.Publish(ctx, event); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

func (ceb *CompositeEventBus) Subscribe(topic string, handler EventHandler) string {
	// 只订阅第一个事件总线
	if len(ceb.buses) == 0 {
		return ""
	}
	return ceb.buses[0].Subscribe(topic, handler)
}

// BufferedEventBus 带缓冲的事件总线
type BufferedEventBus struct {
	*EventBus
	bufferSize int
	timeout    time.Duration
}

func NewBufferedEventBus(bufferSize int, timeout time.Duration) *BufferedEventBus {
	return &BufferedEventBus{
		EventBus:   NewEventBus(),
		bufferSize: bufferSize,
		timeout:    timeout,
	}
}

func (beb *BufferedEventBus) PublishBatch(ctx context.Context, events []Event) error {
	for _, event := range events {
		if err := beb.Publish(ctx, event); err != nil {
			return err
		}
	}
	return nil
}

// EventRecorder 事件记录器
type EventRecorder struct {
	mu      sync.RWMutex
	events  []Event
	maxSize int
}

func NewEventRecorder(maxSize int) *EventRecorder {
	if maxSize <= 0 {
		maxSize = 1000
	}
	return &EventRecorder{
		events:  make([]Event, 0, maxSize),
		maxSize: maxSize,
	}
}

func (er *EventRecorder) Record(event Event) {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.events = append(er.events, event)

	// 限制大小
	if len(er.events) > er.maxSize {
		er.events = er.events[1:]
	}
}

func (er *EventRecorder) GetEvents() []Event {
	er.mu.RLock()
	defer er.mu.RUnlock()

	events := make([]Event, len(er.events))
	copy(events, er.events)
	return events
}

func (er *EventRecorder) Clear() {
	er.mu.Lock()
	defer er.mu.Unlock()

	er.events = make([]Event, 0, er.maxSize)
}

func (er *EventRecorder) Count() int {
	er.mu.RLock()
	defer er.mu.RUnlock()

	return len(er.events)
}

// EventHandlerFunc 适配函数类型为EventHandler
type EventHandlerFunc func(ctx context.Context, event Event) error

func (hf EventHandlerFunc) Handle(ctx context.Context, event Event) error {
	return hf(ctx, event)
}

// Middleware 事件中间件
type Middleware func(EventHandler) EventHandler

// ChainMiddleware 链式中间件
func ChainMiddleware(middlewares ...Middleware) Middleware {
	return func(handler EventHandler) EventHandler {
		for i := len(middlewares) - 1; i >= 0; i-- {
			handler = middlewares[i](handler)
		}
		return handler
	}
}

// LoggingMiddleware 日志中间件
func LoggingMiddleware(logger func(topic string, event Event)) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			if event != nil && logger != nil {
				logger(event.Name(), event)
			}
			return next(ctx, event)
		}
	}
}

// RecoveryMiddleware 恢复中间件
func RecoveryMiddleware() Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("panic recovered: %v", r)
				}
			}()
			return next(ctx, event)
		}
	}
}

// TimeoutMiddleware 超时中间件
func TimeoutMiddleware(timeout time.Duration) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			ctx, cancel := context.WithTimeout(ctx, timeout)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				done <- next(ctx, event)
			}()

			select {
			case err := <-done:
				return err
			case <-ctx.Done():
				return fmt.Errorf("event handler timeout")
			}
		}
	}
}

// RetryMiddleware 重试中间件
func RetryMiddleware(maxRetries int, delay time.Duration) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			var lastErr error
			for i := 0; i <= maxRetries; i++ {
				if err := next(ctx, event); err == nil {
					return nil
				} else {
					lastErr = err
					if i < maxRetries {
						time.Sleep(delay)
					}
				}
			}
			return fmt.Errorf("retry failed after %d attempts: %w", maxRetries+1, lastErr)
		}
	}
}

// FilterMiddleware 过滤中间件
func FilterMiddleware(filter func(Event) bool) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			if filter != nil && !filter(event) {
				return nil // 跳过
			}
			return next(ctx, event)
		}
	}
}

// TransformMiddleware 转换中间件
func TransformMiddleware(transformer func(Event) Event) Middleware {
	return func(next EventHandler) EventHandler {
		return func(ctx context.Context, event Event) error {
			if transformer != nil {
				event = transformer(event)
			}
			return next(ctx, event)
		}
	}
}
