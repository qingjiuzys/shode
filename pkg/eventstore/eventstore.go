// Package eventstore 提供事件溯源功能。
package eventstore

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Aggregate string                 `json:"aggregate"`
	Version   int                    `json:"version"`
	Data      interface{}            `json:"data"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// EventStore 事件存储接口
type EventStore interface {
	Save(ctx context.Context, events []*Event) error
	Load(ctx context.Context, aggregateID string, fromVersion int) ([]*Event, error)
	GetEvents(ctx context.Context, filter EventFilter) ([]*Event, error)
	Subscribe(ctx context.Context, aggregateID string) (<-chan *Event, error)
}

// EventFilter 事件过滤器
type EventFilter struct {
	Aggregate  string
	EventType  string
	FromTime   time.Time
	ToTime     time.Time
	Limit      int
}

// MemoryEventStore 内存事件存储
type MemoryEventStore struct {
	events    map[string][]*Event // aggregateID -> events
	subscribers map[string][]chan *Event
	mu        sync.RWMutex
}

// NewMemoryEventStore 创建内存事件存储
func NewMemoryEventStore() *MemoryEventStore {
	return &MemoryEventStore{
		events:      make(map[string][]*Event),
		subscribers: make(map[string][]chan *Event),
	}
}

// Save 保存事件
func (es *MemoryEventStore) Save(ctx context.Context, events []*Event) error {
	es.mu.Lock()
	defer es.mu.Unlock()

	for _, event := range events {
		if _, exists := es.events[event.Aggregate]; !exists {
			es.events[event.Aggregate] = make([]*Event, 0)
		}

		// 检查版本冲突
		lastVersion := 0
		if len(es.events[event.Aggregate]) > 0 {
			lastVersion = es.events[event.Aggregate][len(es.events[event.Aggregate])-1].Version
		}

		if event.Version != lastVersion+1 {
			return fmt.Errorf("version conflict for aggregate %s: expected %d, got %d",
				event.Aggregate, lastVersion+1, event.Version)
		}

		es.events[event.Aggregate] = append(es.events[event.Aggregate], event)

		// 通知订阅者
		es.notifySubscribers(event)
	}

	return nil
}

// Load 加载事件
func (es *MemoryEventStore) Load(ctx context.Context, aggregateID string, fromVersion int) ([]*Event, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	events, exists := es.events[aggregateID]
	if !exists {
		return []*Event{}, nil
	}

	// 过滤版本
	result := make([]*Event, 0)
	for _, event := range events {
		if event.Version >= fromVersion {
			result = append(result, event)
		}
	}

	return result, nil
}

// GetEvents 获取事件
func (es *MemoryEventStore) GetEvents(ctx context.Context, filter EventFilter) ([]*Event, error) {
	es.mu.RLock()
	defer es.mu.RUnlock()

	result := make([]*Event, 0)

	for _, events := range es.events {
		for _, event := range events {
			if es.matchFilter(event, filter) {
				result = append(result, event)
			}
		}
	}

	return result, nil
}

// matchFilter 匹配过滤器
func (es *MemoryEventStore) matchFilter(event *Event, filter EventFilter) bool {
	if filter.Aggregate != "" && event.Aggregate != filter.Aggregate {
		return false
	}

	if filter.EventType != "" && event.Type != filter.EventType {
		return false
	}

	if !filter.FromTime.IsZero() && event.Timestamp.Before(filter.FromTime) {
		return false
	}

	if !filter.ToTime.IsZero() && event.Timestamp.After(filter.ToTime) {
		return false
	}

	return true
}

// Subscribe 订阅事件
func (es *MemoryEventStore) Subscribe(ctx context.Context, aggregateID string) (<-chan *Event, error) {
	es.mu.Lock()
	defer es.mu.Unlock()

	ch := make(chan *Event, 100)
	if _, exists := es.subscribers[aggregateID]; !exists {
		es.subscribers[aggregateID] = make([]chan *Event, 0)
	}
	es.subscribers[aggregateID] = append(es.subscribers[aggregateID], ch)

	return ch, nil
}

// notifySubscribers 通知订阅者
func (es *MemoryEventStore) notifySubscribers(event *Event) {
	if subscribers, exists := es.subscribers[event.Aggregate]; exists {
		for _, ch := range subscribers {
			select {
			case ch <- event:
			default:
				// 通道满，跳过
			}
		}
	}
}

// Aggregate 聚合根接口
type Aggregate interface {
	ID() string
	Version() int
	Apply(event *Event) error
}

// AggregateRepository 聚合仓储
type AggregateRepository struct {
	eventStore EventStore
	aggregates map[string]Aggregate
	mu         sync.RWMutex
}

// NewAggregateRepository 创建聚合仓储
func NewAggregateRepository(eventStore EventStore) *AggregateRepository {
	return &AggregateRepository{
		eventStore: eventStore,
		aggregates: make(map[string]Aggregate),
	}
}

// Save 保存聚合
func (ar *AggregateRepository) Save(ctx context.Context, aggregate Aggregate) error {
	// 获取待保存的事件
	events := ar.getUncommittedEvents(aggregate)

	// 保存事件
	if err := ar.eventStore.Save(ctx, events); err != nil {
		return err
	}

	// 更新聚合版本
	ar.mu.Lock()
	ar.aggregates[aggregate.ID()] = aggregate
	ar.mu.Unlock()

	return nil
}

// Load 加载聚合
func (ar *AggregateRepository) Load(ctx context.Context, aggregateID string, factory func() Aggregate) (Aggregate, error) {
	ar.mu.RLock()
	if aggregate, exists := ar.aggregates[aggregateID]; exists {
		ar.mu.RUnlock()
		return aggregate, nil
	}
	ar.mu.RUnlock()

	// 从事件存储加载
	events, err := ar.eventStore.Load(ctx, aggregateID, 0)
	if err != nil {
		return nil, err
	}

	// 创建聚合实例
	aggregate := factory()

	// 重放事件
	for _, event := range events {
		if err := aggregate.Apply(event); err != nil {
			return nil, err
		}
	}

	ar.mu.Lock()
	ar.aggregates[aggregateID] = aggregate
	ar.mu.Unlock()

	return aggregate, nil
}

// getUncommittedEvents 获取未提交事件
func (ar *AggregateRepository) getUncommittedEvents(aggregate Aggregate) []*Event {
	// 简化实现
	return []*Event{}
}

// Snapshot 快照
type Snapshot struct {
	AggregateID string
	Version     int
	Data        interface{}
	Timestamp   time.Time
}

// SnapshotStore 快照存储
type SnapshotStore struct {
	snapshots map[string][]*Snapshot
	mu        sync.RWMutex
}

// NewSnapshotStore 创建快照存储
func NewSnapshotStore() *SnapshotStore {
	return &SnapshotStore{
		snapshots: make(map[string][]*Snapshot),
	}
}

// Save 保存快照
func (ss *SnapshotStore) Save(ctx context.Context, snapshot *Snapshot) error {
	ss.mu.Lock()
	defer ss.mu.Unlock()

	if _, exists := ss.snapshots[snapshot.AggregateID]; !exists {
		ss.snapshots[snapshot.AggregateID] = make([]*Snapshot, 0)
	}

	ss.snapshots[snapshot.AggregateID] = append(ss.snapshots[snapshot.AggregateID], snapshot)
	return nil
}

// Get 获取快照
func (ss *SnapshotStore) Get(ctx context.Context, aggregateID string, version int) (*Snapshot, error) {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	snapshots, exists := ss.snapshots[aggregateID]
	if !exists {
		return nil, fmt.Errorf("no snapshots for aggregate: %s", aggregateID)
	}

	// 查找最新的快照
	var latest *Snapshot
	for _, snapshot := range snapshots {
		if snapshot.Version <= version {
			if latest == nil || snapshot.Version > latest.Version {
				latest = snapshot
			}
		}
	}

	return latest, nil
}

// SnapshotStrategy 快照策略
type SnapshotStrategy struct {
	interval int // 每隔 N 个事件创建快照
}

// NewSnapshotStrategy 创建快照策略
func NewSnapshotStrategy(interval int) *SnapshotStrategy {
	return &SnapshotStrategy{interval: interval}
}

// ShouldSnapshot 判断是否应该创建快照
func (ss *SnapshotStrategy) ShouldSnapshot(aggregate Aggregate) bool {
	return aggregate.Version()%ss.interval == 0
}

// EventReplayer 事件重放器
type EventReplayer struct {
	eventStore    EventStore
	snapshotStore *SnapshotStore
}

// NewEventReplayer 创建事件重放器
func NewEventReplayer(eventStore EventStore, snapshotStore *SnapshotStore) *EventReplayer {
	return &EventReplayer{
		eventStore:    eventStore,
		snapshotStore: snapshotStore,
	}
}

// Replay 重放事件
func (er *EventReplayer) Replay(ctx context.Context, aggregateID string, factory func() Aggregate) (Aggregate, error) {
	// 尝试加载快照
	snapshot, err := er.snapshotStore.Get(ctx, aggregateID, 1<<31-1)
	if err == nil && snapshot != nil {
		aggregate := factory()

		// 从快照恢复
		if err := er.restoreFromSnapshot(aggregate, snapshot); err != nil {
			return nil, err
		}

		// 重放后续事件
		events, err := er.eventStore.Load(ctx, aggregateID, snapshot.Version+1)
		if err != nil {
			return nil, err
		}

		for _, event := range events {
			if err := aggregate.Apply(event); err != nil {
				return nil, err
			}
		}

		return aggregate, nil
	}

	// 没有快照，重放所有事件
	aggregate := factory()
	events, err := er.eventStore.Load(ctx, aggregateID, 0)
	if err != nil {
		return nil, err
	}

	for _, event := range events {
		if err := aggregate.Apply(event); err != nil {
			return nil, err
		}
	}

	return aggregate, nil
}

// restoreFromSnapshot 从快照恢复
func (er *EventReplayer) restoreFromSnapshot(aggregate Aggregate, snapshot *Snapshot) error {
	// 简化实现
	data, err := json.Marshal(snapshot.Data)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, aggregate)
}

// EventProjector 事件投影器
type EventProjector struct {
	name      string
	handlers  map[string]func(*Event) error
	eventChan chan *Event
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
}

// NewEventProjector 创建事件投影器
func NewEventProjector(name string) *EventProjector {
	ctx, cancel := context.WithCancel(context.Background())

	return &EventProjector{
		name:      name,
		handlers:  make(map[string]func(*Event) error),
		eventChan: make(chan *Event, 1000),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// RegisterHandler 注册处理器
func (ep *EventProjector) RegisterHandler(eventType string, handler func(*Event) error) {
	ep.handlers[eventType] = handler
}

// Start 启动投影器
func (ep *EventProjector) Start() {
	ep.wg.Add(1)
	go ep.run()
}

// Stop 停止投影器
func (ep *EventProjector) Stop() {
	ep.cancel()
	ep.wg.Wait()
}

// run 运行投影器
func (ep *EventProjector) run() {
	defer ep.wg.Done()

	for {
		select {
		case <-ep.ctx.Done():
			return
		case event := <-ep.eventChan:
			ep.handleEvent(event)
		}
	}
}

// handleEvent 处理事件
func (ep *EventProjector) handleEvent(event *Event) error {
	handler, exists := ep.handlers[event.Type]
	if !exists {
		return nil
	}

	return handler(event)
}

// Project 投影事件
func (ep *EventProjector) Project(event *Event) error {
	select {
	case ep.eventChan <- event:
		return nil
	case <-ep.ctx.Done():
		return fmt.Errorf("projector stopped")
	}
}

// Saga 长事务
type Saga struct {
	ID         string
	Steps      []*SagaStep
	CurrentStep int
	Status     string
}

// SagaStep Saga 步骤
type SagaStep struct {
	Name      string
	Execute   func(ctx context.Context) error
	Compensate func(ctx context.Context) error
}

// NewSaga 创建 Saga
func NewSaga(id string) *Saga {
	return &Saga{
		ID:     id,
		Steps:  make([]*SagaStep, 0),
		Status: "pending",
	}
}

// AddStep 添加步骤
func (s *Saga) AddStep(name string, execute, compensate func(ctx context.Context) error) {
	s.Steps = append(s.Steps, &SagaStep{
		Name:      name,
		Execute:   execute,
		Compensate: compensate,
	})
}

// Execute 执行 Saga
func (s *Saga) Execute(ctx context.Context) error {
	s.Status = "running"

	for i, step := range s.Steps {
		s.CurrentStep = i

		if err := step.Execute(ctx); err != nil {
			s.Status = "failed"
			// 执行补偿
			s.compensate(ctx, i)
			return err
		}
	}

	s.Status = "completed"
	return nil
}

// compensate 补偿
func (s *Saga) compensate(ctx context.Context, failedStep int) {
	for i := failedStep - 1; i >= 0; i-- {
		step := s.Steps[i]
		if step.Compensate != nil {
			step.Compensate(ctx)
		}
	}
}
