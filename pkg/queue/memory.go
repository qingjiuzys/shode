// Package queue 提供消息队列功能。
package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Message 消息
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Body      []byte                 `json:"body"`
	Headers   map[string]string      `json:"headers"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
	Retries   int                    `json:"retries"`
	MaxRetries int                   `json:"max_retries"`
}

// Handler 消息处理器
type Handler func(ctx context.Context, msg *Message) error

// Queue 队列接口
type Queue interface {
	Publish(ctx context.Context, topic string, body []byte) error
	Subscribe(ctx context.Context, topic string, handler Handler) error
	Consume(ctx context.Context, topic string, handler Handler) error
	Ack(ctx context.Context, msgID string) error
	Nack(ctx context.Context, msgID string) error
	Length(ctx context.Context, topic string) (int, error)
	Close() error
}

// MemoryQueue 内存队列
type MemoryQueue struct {
	topics     map[string][]*Message
	handlers   map[string][]Handler
	mu         sync.RWMutex
	processing map[string]*Message
	stopChans  map[string]chan struct{}
	wg         sync.WaitGroup
}

// NewMemoryQueue 创建内存队列
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		topics:     make(map[string][]*Message),
		handlers:   make(map[string][]Handler),
		processing: make(map[string]*Message),
		stopChans:  make(map[string]chan struct{}),
	}
}

// Publish 发布消息
func (mq *MemoryQueue) Publish(ctx context.Context, topic string, body []byte) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	msg := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Body:      body,
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
		Retries:   0,
		MaxRetries: 3,
	}

	mq.topics[topic] = append(mq.topics[topic], msg)

	return nil
}

// Subscribe 订阅主题
func (mq *MemoryQueue) Subscribe(ctx context.Context, topic string, handler Handler) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.handlers[topic] = append(mq.handlers[topic], handler)

	// 如果还没有停止通道，创建一个
	if _, exists := mq.stopChans[topic]; !exists {
		stopChan := make(chan struct{})
		mq.stopChans[topic] = stopChan

		// 启动消费 goroutine
		mq.wg.Add(1)
		go mq.consumeLoop(topic, stopChan)
	}

	return nil
}

// Consume 消费消息
func (mq *MemoryQueue) Consume(ctx context.Context, topic string, handler Handler) error {
	return mq.Subscribe(ctx, topic, handler)
}

// consumeLoop 消费循环
func (mq *MemoryQueue) consumeLoop(topic string, stopChan chan struct{}) {
	defer mq.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			mq.processMessages(topic)
		}
	}
}

// processMessages 处理消息
func (mq *MemoryQueue) processMessages(topic string) {
	mq.mu.Lock()

	// 获取主题的消息
	messages := mq.topics[topic]
	if len(messages) == 0 {
		mq.mu.Unlock()
		return
	}

	// 取出第一条消息
	msg := messages[0]
	mq.topics[topic] = messages[1:]

	// 记录为处理中
	mq.processing[msg.ID] = msg

	handlers := mq.handlers[topic]
	mq.mu.Unlock()

	// 处理消息
	for _, handler := range handlers {
		ctx := context.Background()
		if err := handler(ctx, msg); err != nil {
			// 处理失败，重试
			msg.Retries++
			if msg.Retries < msg.MaxRetries {
				mq.mu.Lock()
				mq.topics[topic] = append([]*Message{msg}, mq.topics[topic]...)
				mq.mu.Unlock()
			}
			break
		}
	}

	// 处理完成，从 processing 中移除
	mq.mu.Lock()
	delete(mq.processing, msg.ID)
	mq.mu.Unlock()
}

// Ack 确认消息
func (mq *MemoryQueue) Ack(ctx context.Context, msgID string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	delete(mq.processing, msgID)
	return nil
}

// Nack 拒绝消息
func (mq *MemoryQueue) Nack(ctx context.Context, msgID string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	// 从 processing 移动回队列
	if msg, exists := mq.processing[msgID]; exists {
		msg.Retries++
		if msg.Retries < msg.MaxRetries {
			mq.topics[msg.Topic] = append([]*Message{msg}, mq.topics[msg.Topic]...)
		}
		delete(mq.processing, msgID)
	}

	return nil
}

// Length 获取队列长度
func (mq *MemoryQueue) Length(ctx context.Context, topic string) (int, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	return len(mq.topics[topic]), nil
}

// Close 关闭队列
func (mq *MemoryQueue) Close() error {
	// 停止所有消费 goroutine
	mq.mu.Lock()
	for topic, stopChan := range mq.stopChans {
		close(stopChan)
		delete(mq.stopChans, topic)
	}
	mq.mu.Unlock()

	// 等待所有 goroutine 完成
	mq.wg.Wait()

	return nil
}

// generateMessageID 生成消息 ID
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}

// DeadLetterQueue 死信队列
type DeadLetterQueue struct {
	queue     *MemoryQueue
	maxSize   int
	mu        sync.RWMutex
}

// NewDeadLetterQueue 创建死信队列
func NewDeadLetterQueue(maxSize int) *DeadLetterQueue {
	return &DeadLetterQueue{
		queue:   NewMemoryQueue(),
		maxSize: maxSize,
	}
}

// Add 添加到死信队列
func (dlq *DeadLetterQueue) Add(msg *Message) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// 检查大小
	length, _ := dlq.queue.Length(context.Background(), "dead_letter")
	if length >= dlq.maxSize {
		return fmt.Errorf("dead letter queue is full")
	}

	// 序列化消息
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	return dlq.queue.Publish(context.Background(), "dead_letter", data)
}

// Get 获取死信消息
func (dlq *DeadLetterQueue) Get() (*Message, error) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	length, _ := dlq.queue.Length(context.Background(), "dead_letter")
	if length == 0 {
		return nil, fmt.Errorf("dead letter queue is empty")
	}

	// 获取第一条消息
	messages := dlq.queue.topics["dead_letter"]
	if len(messages) == 0 {
		return nil, fmt.Errorf("no messages in dead letter queue")
	}

	msg := messages[0]
	dlq.queue.topics["dead_letter"] = messages[1:]

	return msg, nil
}

// Retry 重试死信消息
func (dlq *DeadLetterQueue) Retry(msg *Message, originalTopic string) error {
	// 重置重试次数
	msg.Retries = 0

	// 发布到原始主题
	data, err := json.Marshal(msg.Body)
	if err != nil {
		return err
	}

	return dlq.queue.Publish(context.Background(), originalTopic, data)
}

// PriorityQueue 优先级队列
type PriorityQueue struct {
	queues map[string][]*Message // priority -> messages
	mu     sync.RWMutex
}

// Priority 优先级
type Priority int

const (
	PriorityLow    Priority = 0
	PriorityMedium Priority = 1
	PriorityHigh   Priority = 2
)

// NewPriorityQueue 创建优先级队列
func NewPriorityQueue() *PriorityQueue {
	return &PriorityQueue{
		queues: make(map[string][]*Message),
	}
}

// Publish 发布消息
func (pq *PriorityQueue) Publish(ctx context.Context, topic string, body []byte, priority Priority) error {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	msg := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Body:      body,
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	priorityKey := fmt.Sprintf("%s_%d", topic, priority)
	pq.queues[priorityKey] = append(pq.queues[priorityKey], msg)

	return nil
}

// Consume 消费消息（按优先级）
func (pq *PriorityQueue) Consume(ctx context.Context, topic string) (*Message, error) {
	pq.mu.Lock()
	defer pq.mu.Unlock()

	// 按优先级顺序查找
	priorities := []Priority{PriorityHigh, PriorityMedium, PriorityLow}
	for _, priority := range priorities {
		priorityKey := fmt.Sprintf("%s_%d", topic, priority)
		messages := pq.queues[priorityKey]

		if len(messages) > 0 {
			msg := messages[0]
			pq.queues[priorityKey] = messages[1:]
			return msg, nil
		}
	}

	return nil, fmt.Errorf("no messages available")
}

// DelayQueue 延迟队列
type DelayQueue struct {
	messages map[string]*DelayedMessage
	mu       sync.RWMutex
}

// DelayedMessage 延迟消息
type DelayedMessage struct {
	Message      *Message
	ExecuteAfter time.Time
}

// NewDelayQueue 创建延迟队列
func NewDelayQueue() *DelayQueue {
	dq := &DelayQueue{
		messages: make(map[string]*DelayedMessage),
	}

	// 启动处理循环
	go dq.processLoop()

	return dq
}

// Publish 发布延迟消息
func (dq *DelayQueue) Publish(ctx context.Context, topic string, body []byte, delay time.Duration) error {
	dq.mu.Lock()
	defer dq.mu.Unlock()

	msg := &Message{
		ID:        generateMessageID(),
		Topic:     topic,
		Body:      body,
		Headers:   make(map[string]string),
		Metadata:  make(map[string]interface{}),
		Timestamp: time.Now(),
	}

	delayed := &DelayedMessage{
		Message:      msg,
		ExecuteAfter: time.Now().Add(delay),
	}

	dq.messages[msg.ID] = delayed

	return nil
}

// processLoop 处理循环
func (dq *DelayQueue) processLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		dq.process()
	}
}

// process 处理到期的消息
func (dq *DelayQueue) process() {
	dq.mu.Lock()
	defer dq.mu.Unlock()

	now := time.Now()
	ready := make([]*Message, 0)

	for id, delayed := range dq.messages {
		if now.After(delayed.ExecuteAfter) {
			ready = append(ready, delayed.Message)
			delete(dq.messages, id)
		}
	}

	// 处理到期的消息
	for _, msg := range ready {
		// 这里应该调用 handler 处理消息
		// 简化实现，只打印
		fmt.Printf("Processing delayed message: %s\n", msg.ID)
	}
}
