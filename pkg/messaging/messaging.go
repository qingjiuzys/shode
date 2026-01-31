// Package messaging 提供消息系统增强功能。
package messaging

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// MessagingEngine 消息系统增强引擎
type MessagingEngine struct {
	brokers      map[string]*MessageBroker
	persistence  *MessagePersistence
	consumers    map[string]*ConsumerGroup
	deadLetters  map[string]*DeadLetterQueue
	retries      map[string]*RetryPolicy
	eventSourcing *EventSourcingStore
	cqrs         *CQRSEngine
	mu           sync.RWMutex
}

// NewMessagingEngine 创建消息系统增强引擎
func NewMessagingEngine() *MessagingEngine {
	return &MessagingEngine{
		brokers:       make(map[string]*MessageBroker),
		persistence:   NewMessagePersistence(),
		consumers:     make(map[string]*ConsumerGroup),
		deadLetters:   make(map[string]*DeadLetterQueue),
		retries:       make(map[string]*RetryPolicy),
		eventSourcing: NewEventSourcingStore(),
		cqrs:          NewCQRSEngine(),
	}
}

// Publish 发布消息
func (me *MessagingEngine) Publish(ctx context.Context, topic string, message *Message) error {
	return me.persistence.Save(ctx, topic, message)
}

// Subscribe 订阅消息
func (me *MessagingEngine) Subscribe(ctx context.Context, topic, group string, handler MessageHandler) error {
	consumer := &ConsumerGroup{
		ID:      group,
		Topic:   topic,
		Handler: handler,
	}

	me.consumers[group] = consumer

	return nil
}

// CreateConsumerGroup 创建消费者组
func (me *MessagingEngine) CreateConsumerGroup(groupID string, topics []string) *ConsumerGroup {
	me.mu.Lock()
	defer me.mu.Unlock()

	group := &ConsumerGroup{
		ID:     groupID,
		Topics: topics,
	}

	me.consumers[groupID] = group

	return group
}

// MessageBroker 消息代理
type MessageBroker struct {
	Name     string         `json:"name"`
	Type     string         `json:"type"` // "kafka", "rabbitmq", "nats"
	Config   map[string]interface{} `json:"config"`
	Topics   map[string]*Topic `json:"topics"`
	mu       sync.RWMutex
}

// Topic 主题
type Topic struct {
	Name           string `json:"name"`
	Partitions     int    `json:"partitions"`
	ReplicationFactor int `json:"replication_factor"`
	Retention      time.Duration `json:"retention"`
}

// Message 消息
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Key       string                 `json:"key"`
	Value     interface{}            `json:"value"`
	Headers   map[string]string      `json:"headers"`
	Timestamp time.Time              `json:"timestamp"`
}

// MessageHandler 消息处理器
type MessageHandler func(ctx context.Context, msg *Message) error

// NewMessageBroker 创建消息代理
func NewMessageBroker(name, brokerType string) *MessageBroker {
	return &MessageBroker{
		Name:   name,
		Type:   brokerType,
		Config: make(map[string]interface{}),
		Topics: make(map[string]*Topic),
	}
}

// CreateTopic 创建主题
func (mb *MessageBroker) CreateTopic(name string, partitions, replication int) {
	mb.mu.Lock()
	defer mb.mu.Unlock()

	topic := &Topic{
		Name:             name,
		Partitions:       partitions,
		ReplicationFactor: replication,
		Retention:        7 * 24 * time.Hour,
	}

	mb.Topics[name] = topic
}

// MessagePersistence 消息持久化
type MessagePersistence struct {
	storage map[string][]*Message
	backend string // "memory", "disk", "database"
	mu      sync.RWMutex
}

// NewMessagePersistence 创建消息持久化
func NewMessagePersistence() *MessagePersistence {
	return &MessagePersistence{
		storage: make(map[string][]*Message),
		backend: "memory",
	}
}

// Save 保存消息
func (mp *MessagePersistence) Save(ctx context.Context, topic string, message *Message) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	message.ID = generateMessageID()
	message.Timestamp = time.Now()

	mp.storage[topic] = append(mp.storage[topic], message)

	return nil
}

// Load 加载消息
func (mp *MessagePersistence) Load(ctx context.Context, topic string, offset int) ([]*Message, error) {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	messages, exists := mp.storage[topic]
	if !exists {
		return nil, fmt.Errorf("topic not found: %s", topic)
	}

	if offset >= len(messages) {
		return []*Message{}, nil
	}

	return messages[offset:], nil
}

// ConsumerGroup 消费者组
type ConsumerGroup struct {
	ID       string              `json:"id"`
	Topics   []string            `json:"topics"`
	Consumers []*Consumer         `json:"consumers"`
	Handler  MessageHandler       `json:"-"`
	Offset   map[string]int       `json:"offset"`
	Status   string              `json:"status"`
	mu       sync.RWMutex
}

// Consumer 消费者
type Consumer struct {
	ID       string `json:"id"`
	Group    string `json:"group"`
	Topic    string `json:"topic"`
	Partition int  `json:"partition"`
}

// NewConsumerGroup 创建消费者组
func NewConsumerGroup(id string, topics []string) *ConsumerGroup {
	return &ConsumerGroup{
		ID:       id,
		Topics:   topics,
		Consumers: make([]*Consumer, 0),
		Offset:   make(map[string]int),
		Status:   "active",
	}
}

// AddConsumer 添加消费者
func (cg *ConsumerGroup) AddConsumer(consumer *Consumer) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	cg.Consumers = append(cg.Consumers, consumer)
}

// Consume 消费消息
func (cg *ConsumerGroup) Consume(ctx context.Context) error {
	cg.mu.RLock()
	defer cg.mu.RUnlock()

	for _, topic := range cg.Topics {
		offset := cg.Offset[topic]

		// 简化实现
		_ = offset

		// 调用处理器
		if cg.Handler != nil {
			message := &Message{
				ID:   generateMessageID(),
				Topic: topic,
			}
			_ = cg.Handler(ctx, message)
		}

		cg.Offset[topic]++
	}

	return nil
}

// Commit 提交偏移量
func (cg *ConsumerGroup) Commit(topic string, offset int) {
	cg.mu.Lock()
	defer cg.mu.Unlock()

	cg.Offset[topic] = offset
}

// DeadLetterQueue 死信队列
type DeadLetterQueue struct {
	Name     string     `json:"name"`
	Messages []*Message `json:"messages"`
	MaxSize  int        `json:"max_size"`
	Retention time.Duration `json:"retention"`
	mu       sync.RWMutex
}

// NewDeadLetterQueue 创建死信队列
func NewDeadLetterQueue(name string, maxSize int) *DeadLetterQueue {
	return &DeadLetterQueue{
		Name:     name,
		Messages: make([]*Message, 0),
		MaxSize:  maxSize,
		Retention: 30 * 24 * time.Hour,
	}
}

// Enqueue 入队
func (dlq *DeadLetterQueue) Enqueue(message *Message) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if len(dlq.Messages) >= dlq.MaxSize {
		return fmt.Errorf("dead letter queue full")
	}

	dlq.Messages = append(dlq.Messages, message)

	return nil
}

// Dequeue 出队
func (dlq *DeadLetterQueue) Dequeue() (*Message, error) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	if len(dlq.Messages) == 0 {
		return nil, fmt.Errorf("dead letter queue empty")
	}

	message := dlq.Messages[0]
	dlq.Messages = dlq.Messages[1:]

	return message, nil
}

// RetryPolicy 重试策略
type RetryPolicy struct {
	Name          string        `json:"name"`
	MaxRetries    int           `json:"max_retries"`
	Backoff       time.Duration `json:"backoff"`
	MaxBackoff    time.Duration `json:"max_backoff"`
	RetryableErrors []string    `json:"retryable_errors"`
}

// NewRetryPolicy 创建重试策略
func NewRetryPolicy(name string, maxRetries int) *RetryPolicy {
	return &RetryPolicy{
		Name:          name,
		MaxRetries:    maxRetries,
		Backoff:       1 * time.Second,
		MaxBackoff:    60 * time.Second,
		RetryableErrors: make([]string, 0),
	}
}

// ShouldRetry 是否应该重试
func (rp *RetryPolicy) ShouldRetry(attempt int, err error) bool {
	if attempt >= rp.MaxRetries {
		return false
	}

	// 简化实现，总是重试
	return true
}

// GetBackoff 获取退避时间
func (rp *RetryPolicy) GetBackoff(attempt int) time.Duration {
	backoff := rp.Backoff * time.Duration(1<<uint(attempt))
	if backoff > rp.MaxBackoff {
		return rp.MaxBackoff
	}
	return backoff
}

// EventSourcingStore 事件溯源存储
type EventSourcingStore struct {
	events    map[string][]*Event
	snapshots map[string]*Snapshot
	aggregates map[string]*Aggregate
	mu        sync.RWMutex
}

// Event 事件
type Event struct {
	ID        string                 `json:"id"`
	Aggregate string                 `json:"aggregate"`
	Type      string                 `json:"type"`
	Version   int                    `json:"version"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

// Snapshot 快照
type Snapshot struct {
	ID        string                 `json:"id"`
	Aggregate string                 `json:"aggregate"`
	Version   int                    `json:"version"`
	State     map[string]interface{} `json:"state"`
	Timestamp time.Time              `json:"timestamp"`
}

// Aggregate 聚合根
type Aggregate struct {
	ID      string                 `json:"id"`
	Type    string                 `json:"type"`
	Version int                    `json:"version"`
	State   map[string]interface{} `json:"state"`
}

// NewEventSourcingStore 创建事件溯源存储
func NewEventSourcingStore() *EventSourcingStore {
	return &EventSourcingStore{
		events:    make(map[string][]*Event),
		snapshots: make(map[string]*Snapshot),
		aggregates: make(map[string]*Aggregate),
	}
}

// SaveEvent 保存事件
func (ess *EventSourcingStore) SaveEvent(ctx context.Context, event *Event) error {
	ess.mu.Lock()
	defer ess.mu.Unlock()

	event.ID = generateEventID()
	event.Timestamp = time.Now()

	ess.events[event.Aggregate] = append(ess.events[event.Aggregate], event)

	// 更新聚合根
	aggregate := ess.aggregates[event.Aggregate]
	if aggregate != nil {
		aggregate.Version = event.Version
	}

	return nil
}

// GetEvents 获取事件
func (ess *EventSourcingStore) GetEvents(ctx context.Context, aggregateID string) ([]*Event, error) {
	ess.mu.RLock()
	defer ess.mu.RUnlock()

	events, exists := ess.events[aggregateID]
	if !exists {
		return nil, fmt.Errorf("aggregate not found: %s", aggregateID)
	}

	return events, nil
}

// SaveSnapshot 保存快照
func (ess *EventSourcingStore) SaveSnapshot(ctx context.Context, snapshot *Snapshot) error {
	ess.mu.Lock()
	defer ess.mu.Unlock()

	snapshot.ID = generateSnapshotID()
	snapshot.Timestamp = time.Now()

	ess.snapshots[snapshot.Aggregate] = snapshot

	return nil
}

// GetSnapshot 获取快照
func (ess *EventSourcingStore) GetSnapshot(ctx context.Context, aggregateID string) (*Snapshot, error) {
	ess.mu.RLock()
	defer ess.mu.RUnlock()

	snapshot, exists := ess.snapshots[aggregateID]
	if !exists {
		return nil, fmt.Errorf("snapshot not found: %s", aggregateID)
	}

	return snapshot, nil
}

// RebuildAggregate 重建聚合
func (ess *EventSourcingStore) RebuildAggregate(ctx context.Context, aggregateID string) (*Aggregate, error) {
	ess.mu.RLock()
	defer ess.mu.RUnlock()

	// 先尝试从快照恢复
	snapshot, hasSnapshot := ess.snapshots[aggregateID]

	aggregate := &Aggregate{
		ID:     aggregateID,
		State:  make(map[string]interface{}),
	}

	if hasSnapshot {
		aggregate.State = snapshot.State
		aggregate.Version = snapshot.Version
	}

	// 重放事件
	events, exists := ess.events[aggregateID]
	if !exists {
		return aggregate, nil
	}

	for _, event := range events {
		if snapshot != nil && event.Version <= snapshot.Version {
			continue
		}

		// 应用事件到状态
		ess.applyEvent(aggregate, event)
	}

	return aggregate, nil
}

// applyEvent 应用事件
func (ess *EventSourcingStore) applyEvent(aggregate *Aggregate, event *Event) {
	// 简化实现
	aggregate.State["last_event"] = event.Type
	aggregate.Version = event.Version
}

// CQRSEngine CQRS引擎
type CQRSEngine struct {
	readModels  map[string]*ReadModel
	writeModels map[string]*WriteModel
	projections map[string]*Projection
	mu          sync.RWMutex
}

// ReadModel 读模型
type ReadModel struct {
	Name   string                 `json:"name"`
	Data   map[string]interface{} `json:"data"`
	Version int                    `json:"version"`
}

// WriteModel 写模型
type WriteModel struct {
	Name    string                 `json:"name"`
	State   map[string]interface{} `json:"state"`
	Version int                    `json:"version"`
}

// Projection 投影
type Projection struct {
	Name      string       `json:"name"`
	Source    string       `json:"source"`
	Handler   ProjectionHandler `json:"-"`
	LastError error        `json:"last_error"`
}

// ProjectionHandler 投影处理器
type ProjectionHandler func(ctx context.Context, event *Event) error

// NewCQRSEngine 创建CQRS引擎
func NewCQRSEngine() *CQRSEngine {
	return &CQRSEngine{
		readModels:   make(map[string]*ReadModel),
		writeModels:  make(map[string]*WriteModel),
		projections:  make(map[string]*Projection),
	}
}

// CreateCommand 创建命令
func (ce *CQRSEngine) CreateCommand(ctx context.Context, modelName string, command map[string]interface{}) error {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	model, exists := ce.writeModels[modelName]
	if !exists {
		model = &WriteModel{
			Name:  modelName,
			State: make(map[string]interface{}),
		}
		ce.writeModels[modelName] = model
	}

	// 应用命令
	for k, v := range command {
		model.State[k] = v
	}
	model.Version++

	return nil
}

// Query 查询
func (ce *CQRSEngine) Query(ctx context.Context, modelName string) (map[string]interface{}, error) {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	model, exists := ce.readModels[modelName]
	if !exists {
		return nil, fmt.Errorf("read model not found: %s", modelName)
	}

	return model.Data, nil
}

// UpdateProjection 更新投影
func (ce *CQRSEngine) UpdateProjection(ctx context.Context, event *Event) error {
	ce.mu.RLock()
	defer ce.mu.RUnlock()

	for _, projection := range ce.projections {
		if projection.Source == event.Aggregate {
			if err := projection.Handler(ctx, event); err != nil {
				projection.LastError = err
				return err
			}
		}
	}

	return nil
}

// RegisterProjection 注册投影
func (ce *CQRSEngine) RegisterProjection(name, source string, handler ProjectionHandler) {
	ce.mu.Lock()
	defer ce.mu.Unlock()

	projection := &Projection{
		Name:    name,
		Source:  source,
		Handler: handler,
	}

	ce.projections[name] = projection
}

// generateMessageID 生成消息 ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// generateEventID 生成事件 ID
func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// generateSnapshotID 生成快照 ID
func generateSnapshotID() string {
	return fmt.Sprintf("snapshot_%d", time.Now().UnixNano())
}
