// Package stream 提供流处理功能。
package stream

import (
	"context"
	"sync"
	"time"
)

// Stream 流接口
type Stream interface {
	Next(ctx context.Context) (interface{}, error)
	Close() error
}

// Processor 处理器
type Processor interface {
	Process(ctx context.Context, data interface{}) (interface{}, error)
}

// Source 数据源
type Source interface {
	Generate(ctx context.Context) (<-chan interface{}, error)
}

// Sink 数据汇
type Sink interface {
	Consume(ctx context.Context, dataCh <-chan interface{}) error
}

// MemoryStream 内存流
type MemoryStream struct {
	data   []interface{}
	index  int
	mu     sync.Mutex
	closed bool
}

// NewMemoryStream 创建内存流
func NewMemoryStream(data ...interface{}) *MemoryStream {
	return &MemoryStream{
		data:   data,
		index:  0,
	}
}

// Next 获取下一个元素
func (ms *MemoryStream) Next(ctx context.Context) (interface{}, error) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	if ms.closed {
		return nil, nil // 流结束
	}

	if ms.index >= len(ms.data) {
		ms.closed = true
		return nil, nil
	}

	item := ms.data[ms.index]
	ms.index++
	return item, nil
}

// Close 关闭流
func (ms *MemoryStream) Close() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.closed = true
	return nil
}

// ChannelStream 通道流
type ChannelStream struct {
	ch     <-chan interface{}
	closed bool
	mu     sync.Mutex
}

// NewChannelStream 创建通道流
func NewChannelStream(ch <-chan interface{}) *ChannelStream {
	return &ChannelStream{ch: ch}
}

// Next 获取下一个元素
func (cs *ChannelStream) Next(ctx context.Context) (interface{}, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case data, ok := <-cs.ch:
		if !ok {
			cs.mu.Lock()
			cs.closed = true
			cs.mu.Unlock()
			return nil, nil
		}
		return data, nil
	}
}

// Close 关闭流
func (cs *ChannelStream) Close() error {
	return nil
}

// Map 映射转换
type Map struct {
	upstream Stream
	mapper   func(interface{}) (interface{}, error)
}

// NewMap 创建映射
func NewMap(upstream Stream, mapper func(interface{}) (interface{}, error)) *Map {
	return &Map{
		upstream: upstream,
		mapper:   mapper,
	}
}

// Next 获取下一个元素
func (m *Map) Next(ctx context.Context) (interface{}, error) {
	data, err := m.upstream.Next(ctx)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}

	return m.mapper(data)
}

// Close 关闭流
func (m *Map) Close() error {
	return m.upstream.Close()
}

// Filter 过滤
type Filter struct {
	upstream  Stream
	predicate func(interface{}) bool
}

// NewFilter 创建过滤
func NewFilter(upstream Stream, predicate func(interface{}) bool) *Filter {
	return &Filter{
		upstream:  upstream,
		predicate: predicate,
	}
}

// Next 获取下一个元素
func (f *Filter) Next(ctx context.Context) (interface{}, error) {
	for {
		data, err := f.upstream.Next(ctx)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, nil
		}

		if f.predicate(data) {
			return data, nil
		}
	}
}

// Close 关闭流
func (f *Filter) Close() error {
	return f.upstream.Close()
}

// FlatMap 扁平映射
type FlatMap struct {
	upstream    Stream
	mapper      func(interface{}) Stream
	current     Stream
}

// NewFlatMap 创建扁平映射
func NewFlatMap(upstream Stream, mapper func(interface{}) Stream) *FlatMap {
	return &FlatMap{
		upstream: upstream,
		mapper:   mapper,
	}
}

// Next 获取下一个元素
func (fm *FlatMap) Next(ctx context.Context) (interface{}, error) {
	for {
		// 如果有当前流，从中获取元素
		if fm.current != nil {
			data, err := fm.current.Next(ctx)
			if err != nil {
				return nil, err
			}
			if data != nil {
				return data, nil
			}
			// 当前流结束，关闭并继续
			fm.current.Close()
			fm.current = nil
		}

		// 从上游获取元素并映射为流
		upstreamData, err := fm.upstream.Next(ctx)
		if err != nil {
			return nil, err
		}
		if upstreamData == nil {
			return nil, nil
		}

		fm.current = fm.mapper(upstreamData)
	}
}

// Close 关闭流
func (fm *FlatMap) Close() error {
	if fm.current != nil {
		fm.current.Close()
	}
	return fm.upstream.Close()
}

// Window 窗口
type Window struct {
	size     int
	advance  int
	upstream Stream
	buffer   []interface{}
	mu       sync.Mutex
}

// NewWindow 创建窗口
func NewWindow(upstream Stream, size, advance int) *Window {
	return &Window{
		size:     size,
		advance:  advance,
		upstream: upstream,
		buffer:   make([]interface{}, 0),
	}
}

// Next 获取下一个窗口
func (w *Window) Next(ctx context.Context) ([]interface{}, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for len(w.buffer) < w.size {
		data, err := w.upstream.Next(ctx)
		if err != nil {
			return nil, err
		}
		if data == nil {
			if len(w.buffer) == 0 {
				return nil, nil
			}
			break
		}
		w.buffer = append(w.buffer, data)
	}

	result := make([]interface{}, len(w.buffer))
	copy(result, w.buffer)

	// 滑动窗口
	advance := w.advance
	if advance > len(w.buffer) {
		advance = len(w.buffer)
	}
	w.buffer = w.buffer[advance:]

	return result, nil
}

// Close 关闭流
func (w *Window) Close() error {
	return w.upstream.Close()
}

// SlidingWindow 滑动窗口
type SlidingWindow struct {
	size     time.Duration
	slide    time.Duration
	upstream Stream
	buffer   []time.Time
	data     []interface{}
	mu       sync.Mutex
}

// NewSlidingWindow 创建滑动窗口
func NewSlidingWindow(upstream Stream, size, slide time.Duration) *SlidingWindow {
	return &SlidingWindow{
		size:     size,
		slide:    slide,
		upstream: upstream,
		buffer:   make([]time.Time, 0),
		data:     make([]interface{}, 0),
	}
}

// Watermark 水位线
type Watermark struct {
	timestamp time.Time
}

// NewWatermark 创建水位线
func NewWatermark(timestamp time.Time) *Watermark {
	return &Watermark{timestamp: timestamp}
}

// Timestamp 时间戳
func (w *Watermark) Timestamp() time.Time {
	return w.timestamp
}

// Aggregate 聚合
type Aggregate struct {
	upstream    Stream
	window      Window
	aggregator  func([]interface{}) interface{}
}

// NewAggregate 创建聚合
func NewAggregate(upstream Stream, window Window, aggregator func([]interface{}) interface{}) *Aggregate {
	return &Aggregate{
		upstream:   upstream,
		window:     window,
		aggregator: aggregator,
	}
}

// Join 连接
type Join struct {
	left       Stream
	right      Stream
	leftKey    func(interface{}) string
	rightKey   func(interface{}) string
	window     time.Duration
	mu         sync.Mutex
	rightCache map[string][]interface{}
	leftCache  map[string][]interface{}
}

// NewJoin 创建连接
func NewJoin(left, right Stream, leftKey, rightKey func(interface{}) string, window time.Duration) *Join {
	return &Join{
		left:       left,
		right:      right,
		leftKey:    leftKey,
		rightKey:   rightKey,
		window:     window,
		rightCache: make(map[string][]interface{}),
		leftCache:  make(map[string][]interface{}),
	}
}

// Reduce 归约
type Reduce struct {
	upstream  Stream
	initial   interface{}
	reducer   func(acc, value interface{}) interface{}
	result    interface{}
}

// NewReduce 创建归约
func NewReduce(upstream Stream, initial interface{}, reducer func(acc, value interface{}) interface{}) *Reduce {
	return &Reduce{
		upstream: upstream,
		initial:   initial,
		reducer:   reducer,
		result:    initial,
	}
}

// Next 获取下一个元素
func (r *Reduce) Next(ctx context.Context) (interface{}, error) {
	for {
		data, err := r.upstream.Next(ctx)
		if err != nil {
			return nil, err
		}
		if data == nil {
			result := r.result
			r.result = r.initial
			return result, nil
		}

		r.result = r.reducer(r.result, data)
	}
}

// Close 关闭流
func (r *Reduce) Close() error {
	return r.upstream.Close()
}

// GroupBy 分组
type GroupBy struct {
	upstream Stream
	keyFunc  func(interface{}) string
	groups   map[string][]interface{}
	mu       sync.Mutex
}

// NewGroupBy 创建分组
func NewGroupBy(upstream Stream, keyFunc func(interface{}) string) *GroupBy {
	return &GroupBy{
		upstream: upstream,
		keyFunc:  keyFunc,
		groups:   make(map[string][]interface{}),
	}
}

// Count 计数
type Count struct {
	upstream Stream
	count    int64
}

// NewCount 创建计数
func NewCount(upstream Stream) *Count {
	return &Count{upstream: upstream}
}

// Next 获取下一个计数
func (c *Count) Next(ctx context.Context) (int64, error) {
	for {
		_, err := c.upstream.Next(ctx)
		if err != nil {
			return 0, err
		}
		if _, ok := <-ctx.Done(); ok {
			return c.count, nil
		}
		c.count++
	}
}

// Close 关闭流
func (c *Count) Close() error {
	return c.upstream.Close()
}

// Throttle 节流
type Throttle struct {
	upstream Stream
	interval time.Duration
	lastSend time.Time
}

// NewThrottle 创建节流
func NewThrottle(upstream Stream, interval time.Duration) *Throttle {
	return &Throttle{
		upstream: upstream,
		interval: interval,
	}
}

// Debounce 防抖
type Debounce struct {
	upstream  Stream
	delay     time.Duration
	timer     *time.Timer
	mu        sync.Mutex
	data      interface{}
	hasData   bool
}

// NewDebounce 创建防抖
func NewDebounce(upstream Stream, delay time.Duration) *Debounce {
	return &Debounce{
		upstream: upstream,
		delay:    delay,
	}
}

// Merge 合并
type Merge struct {
	streams []Stream
	indices []int
	mu      sync.Mutex
}

// NewMerge 创建合并
func NewMerge(streams ...Stream) *Merge {
	return &Merge{
		streams: streams,
		indices: make([]int, len(streams)),
	}
}

// Split 分流
type Split struct {
	upstream  Stream
	predicates []func(interface{}) int
	outputs    []Stream
}

// NewSplit 创建分流
func NewSplit(upstream Stream, predicates []func(interface{}) int) *Split {
	outputs := make([]Stream, len(predicates))
	for i := range outputs {
		outputs[i] = &filteredStream{
			upstream:  upstream,
			predicate: func(data interface{}) bool {
				idx := predicates[i](data)
				return idx == i
			},
		}
	}

	return &Split{
		upstream:   upstream,
		predicates: predicates,
		outputs:    outputs,
	}
}

// filteredStream 过滤流
type filteredStream struct {
	upstream  Stream
	predicate func(interface{}) bool
}

// Next 获取下一个元素
func (fs *filteredStream) Next(ctx context.Context) (interface{}, error) {
	for {
		data, err := fs.upstream.Next(ctx)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, nil
		}

		if fs.predicate(data) {
			return data, nil
		}
	}
}

// Close 关闭流
func (fs *filteredStream) Close() error {
	return fs.upstream.Close()
}

// Batch 批处理
type Batch struct {
	upstream Stream
	size     int
	timeout  time.Duration
	buffer   []interface{}
	timer    *time.Timer
	mu       sync.Mutex
}

// NewBatch 创建批处理
func NewBatch(upstream Stream, size int, timeout time.Duration) *Batch {
	return &Batch{
		upstream: upstream,
		size:     size,
		timeout:  timeout,
		buffer:   make([]interface{}, 0),
	}
}

// Partition 分区
type Partition struct {
	upstream  Stream
	partitions int
	keyFunc   func(interface{}) int
	outputs   []Stream
}

// NewPartition 创建分区
func NewPartition(upstream Stream, partitions int, keyFunc func(interface{}) int) *Partition {
	outputs := make([]Stream, partitions)
	for i := range outputs {
		outputs[i] = &partitionStream{
			upstream: upstream,
			partition: i,
			keyFunc:   keyFunc,
		}
	}

	return &Partition{
		upstream:  upstream,
		partitions: partitions,
		keyFunc:   keyFunc,
		outputs:   outputs,
	}
}

// partitionStream 分区流
type partitionStream struct {
	upstream  Stream
	partition int
	keyFunc   func(interface{}) int
}

// Next 获取下一个元素
func (ps *partitionStream) Next(ctx context.Context) (interface{}, error) {
	for {
		data, err := ps.upstream.Next(ctx)
		if err != nil {
			return nil, err
		}
		if data == nil {
			return nil, nil
		}

		if ps.keyFunc(data) == ps.partition {
			return data, nil
		}
	}
}

// Close 关闭流
func (ps *partitionStream) Close() error {
	return ps.upstream.Close()
}

// Union 联合
type Union struct {
	streams []Stream
}

// NewUnion 创建联合
func NewUnion(streams ...Stream) *Union {
	return &Union{streams: streams}
}

// Intersect 交集
type Intersect struct {
	streams []Stream
}

// NewIntersect 创建交集
func NewIntersect(streams ...Stream) *Intersect {
	return &Intersect{streams: streams}
}

// Difference 差集
type Difference struct {
	left  Stream
	right Stream
}

// NewDifference 创建差集
func NewDifference(left, right Stream) *Difference {
	return &Difference{left: left, right: right}
}
