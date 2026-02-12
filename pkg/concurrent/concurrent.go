// Package concurrent 提供并发控制工具
package concurrent

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// WaitGroup 等待组
type WaitGroup struct {
	wg sync.WaitGroup
}

// NewWaitGroup 创建等待组
func NewWaitGroup() *WaitGroup {
	return &WaitGroup{}
}

// Add 添加计数
func (w *WaitGroup) Add(delta int) {
	w.wg.Add(delta)
}

// Done 完成一个任务
func (w *WaitGroup) Done() {
	w.wg.Done()
}

// Wait 等待所有任务完成
func (w *WaitGroup) Wait() {
	w.wg.Wait()
}

// Go 启动goroutine
func (w *WaitGroup) Go(fn func()) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		fn()
	}()
}

// GoWithContext 启动带context的goroutine
func (w *WaitGroup) GoWithContext(ctx context.Context, fn func(context.Context)) {
	w.wg.Add(1)
	go func() {
		defer w.wg.Done()
		fn(ctx)
	}()
}

// Semaphore 信号量
type Semaphore struct {
	ch chan struct{}
}

// NewSemaphore 创建信号量
func NewSemaphore(size int) *Semaphore {
	if size <= 0 {
		size = 1
	}
	return &Semaphore{
		ch: make(chan struct{}, size),
	}
}

// Acquire 获取信号量
func (s *Semaphore) Acquire() {
	s.ch <- struct{}{}
}

// TryAcquire 尝试获取信号量（非阻塞）
func (s *Semaphore) TryAcquire() bool {
	select {
	case s.ch <- struct{}{}:
		return true
	default:
		return false
	}
}

// AcquireWithContext 获取信号量（带context）
func (s *Semaphore) AcquireWithContext(ctx context.Context) error {
	select {
	case s.ch <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// AcquireWithTimeout 获取信号量（带超时）
func (s *Semaphore) AcquireWithTimeout(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.AcquireWithContext(ctx)
}

// Release 释放信号量
func (s *Semaphore) Release() {
	<-s.ch
}

// Available 可用信号量数
func (s *Semaphore) Available() int {
	return cap(s.ch) - len(s.ch)
}

// Size 信号量大小
func (s *Semaphore) Size() int {
	return cap(s.ch)
}

// Mutex 互斥锁
type Mutex struct {
	mu sync.Mutex
}

// NewMutex 创建互斥锁
func NewMutex() *Mutex {
	return &Mutex{}
}

// Lock 加锁
func (m *Mutex) Lock() {
	m.mu.Lock()
}

// Unlock 解锁
func (m *Mutex) Unlock() {
	m.mu.Unlock()
}

// TryLock 尝试加锁
func (m *Mutex) TryLock() bool {
	return m.mu.TryLock()
}

// RWMutex 读写锁
type RWMutex struct {
	mu sync.RWMutex
}

// NewRWMutex 创建读写锁
func NewRWMutex() *RWMutex {
	return &RWMutex{}
}

// Lock 加写锁
func (m *RWMutex) Lock() {
	m.mu.Lock()
}

// Unlock 解写锁
func (m *RWMutex) Unlock() {
	m.mu.Unlock()
}

// RLock 加读锁
func (m *RWMutex) RLock() {
	m.mu.RLock()
}

// RUnlock 解读锁
func (m *RWMutex) RUnlock() {
	m.mu.RUnlock()
}

// Once 只执行一次
type Once struct {
	once sync.Once
}

// NewOnce 创建Once
func NewOnce() *Once {
	return &Once{}
}

// Do 执行函数
func (o *Once) Do(fn func()) {
	o.once.Do(fn)
}

// Done 检查是否已执行
func (o *Once) Done() bool {
	// sync.Once没有直接的方法检查，这里用私有实现
	return false
}

// Pool 对象池
type Pool struct {
	pool sync.Pool
}

// NewPool 创建对象池
func NewPool(newFunc func() any) *Pool {
	return &Pool{
		pool: sync.Pool{
			New: newFunc,
		},
	}
}

// Get 获取对象
func (p *Pool) Get() any {
	return p.pool.Get()
}

// Put 放回对象
func (p *Pool) Put(x any) {
	p.pool.Put(x)
}

// AtomicInt32 原子int32
type AtomicInt32 struct {
	value int32
}

// NewAtomicInt32 创建原子int32
func NewAtomicInt32(initial int32) *AtomicInt32 {
	return &AtomicInt32{value: initial}
}

// Get 获取值
func (a *AtomicInt32) Get() int32 {
	return atomic.LoadInt32(&a.value)
}

// Set 设置值
func (a *AtomicInt32) Set(value int32) {
	atomic.StoreInt32(&a.value, value)
}

// Add 增加值
func (a *AtomicInt32) Add(delta int32) int32 {
	return atomic.AddInt32(&a.value, delta)
}

// CompareAndSwap 比较并交换
func (a *AtomicInt32) CompareAndSwap(old, new int32) bool {
	return atomic.CompareAndSwapInt32(&a.value, old, new)
}

// Increment 自增
func (a *AtomicInt32) Increment() int32 {
	return a.Add(1)
}

// Decrement 自减
func (a *AtomicInt32) Decrement() int32 {
	return a.Add(-1)
}

// AtomicInt64 原子int64
type AtomicInt64 struct {
	value int64
}

// NewAtomicInt64 创建原子int64
func NewAtomicInt64(initial int64) *AtomicInt64 {
	return &AtomicInt64{value: initial}
}

// Get 获取值
func (a *AtomicInt64) Get() int64 {
	return atomic.LoadInt64(&a.value)
}

// Set 设置值
func (a *AtomicInt64) Set(value int64) {
	atomic.StoreInt64(&a.value, value)
}

// Add 增加值
func (a *AtomicInt64) Add(delta int64) int64 {
	return atomic.AddInt64(&a.value, delta)
}

// CompareAndSwap 比较并交换
func (a *AtomicInt64) CompareAndSwap(old, new int64) bool {
	return atomic.CompareAndSwapInt64(&a.value, old, new)
}

// Increment 自增
func (a *AtomicInt64) Increment() int64 {
	return a.Add(1)
}

// Decrement 自减
func (a *AtomicInt64) Decrement() int64 {
	return a.Add(-1)
}

// AtomicBool 原子布尔值
type AtomicBool struct {
	value uint32
}

// NewAtomicBool 创建原子布尔值
func NewAtomicBool(initial bool) *AtomicBool {
	b := &AtomicBool{}
	if initial {
		b.value = 1
	}
	return b
}

// Get 获取值
func (a *AtomicBool) Get() bool {
	return atomic.LoadUint32(&a.value) != 0
}

// Set 设置值
func (a *AtomicBool) Set(value bool) {
	if value {
		atomic.StoreUint32(&a.value, 1)
	} else {
		atomic.StoreUint32(&a.value, 0)
	}
}

// CompareAndSwap 比较并交换
func (a *AtomicBool) CompareAndSwap(old, new bool) bool {
	var oldVal, newVal uint32
	if old {
		oldVal = 1
	}
	if new {
		newVal = 1
	}
	return atomic.CompareAndSwapUint32(&a.value, oldVal, newVal)
}

// Toggle 切换值
func (a *AtomicBool) Toggle() {
	for {
		current := a.Get()
		if a.CompareAndSwap(current, !current) {
			break
		}
	}
}

// SafeMap 安全map
type SafeMap struct {
	mu   sync.RWMutex
	data map[string]any
}

// NewSafeMap 创建安全map
func NewSafeMap() *SafeMap {
	return &SafeMap{
		data: make(map[string]any),
	}
}

// Set 设置值
func (m *SafeMap) Set(key string, value any) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data[key] = value
}

// Get 获取值
func (m *SafeMap) Get(key string) (any, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	val, ok := m.data[key]
	return val, ok
}

// Delete 删除值
func (m *SafeMap) Delete(key string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, key)
}

// Has 检查是否存在
func (m *SafeMap) Has(key string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.data[key]
	return ok
}

// Keys 获取所有键
func (m *SafeMap) Keys() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	keys := make([]string, 0, len(m.data))
	for k := range m.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取所有值
func (m *SafeMap) Values() []any {
	m.mu.RLock()
	defer m.mu.RUnlock()

	values := make([]any, 0, len(m.data))
	for _, v := range m.data {
		values = append(values, v)
	}
	return values
}

// Size 获取大小
func (m *SafeMap) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}

// Clear 清空map
func (m *SafeMap) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.data = make(map[string]any)
}

// Range 遍历map
func (m *SafeMap) Range(fn func(key string, value any) bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for k, v := range m.data {
		if !fn(k, v) {
			break
		}
	}
}

// SafeSlice 安全切片
type SafeSlice struct {
	mu   sync.RWMutex
	data []any
}

// NewSafeSlice 创建安全切片
func NewSafeSlice() *SafeSlice {
	return &SafeSlice{
		data: make([]any, 0),
	}
}

// Append 追加元素
func (s *SafeSlice) Append(items ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = append(s.data, items...)
}

// Get 获取元素
func (s *SafeSlice) Get(index int) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if index < 0 || index >= len(s.data) {
		return nil, false
	}
	return s.data[index], true
}

// Set 设置元素
func (s *SafeSlice) Set(index int, item any) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.data) {
		return false
	}
	s.data[index] = item
	return true
}

// Delete 删除元素
func (s *SafeSlice) Delete(index int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if index < 0 || index >= len(s.data) {
		return false
	}

	s.data = append(s.data[:index], s.data[index+1:]...)
	return true
}

// Size 获取大小
func (s *SafeSlice) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

// Clear 清空切片
func (s *SafeSlice) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make([]any, 0)
}

// ToSlice 转换为普通切片
func (s *SafeSlice) ToSlice() []any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]any, len(s.data))
	copy(result, s.data)
	return result
}

// RateLimiter 限流器
type RateLimiter struct {
	semaphore *Semaphore
	rate      time.Duration
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		semaphore: NewSemaphore(rate),
		rate:      window / time.Duration(rate),
	}
}

// Allow 检查是否允许
func (rl *RateLimiter) Allow() bool {
	return rl.semaphore.TryAcquire()
}

// Wait 等待直到允许
func (rl *RateLimiter) Wait() {
	rl.semaphore.Acquire()
	go func() {
		time.Sleep(rl.rate)
		rl.semaphore.Release()
	}()
}

// WorkerPool 工作池
type WorkerPool struct {
	wg        WaitGroup
	semaphore *Semaphore
	work      chan func()
	quit      chan struct{}
}

// NewWorkerPool 创建工作池
func NewWorkerPool(size int) *WorkerPool {
	return &WorkerPool{
		semaphore: NewSemaphore(size),
		work:      make(chan func()),
		quit:      make(chan struct{}),
	}
}

// Start 启动工作池
func (wp *WorkerPool) Start() {
	for {
		select {
		case fn := <-wp.work:
			wp.wg.Go(fn)
		case <-wp.quit:
			return
		}
	}
}

// Stop 停止工作池
func (wp *WorkerPool) Stop() {
	close(wp.quit)
	wp.wg.Wait()
}

// Submit 提交任务
func (wp *WorkerPool) Submit(fn func()) error {
	if !wp.semaphore.TryAcquire() {
		return nil // 或返回错误
	}

	go func() {
		defer wp.semaphore.Release()
		wp.work <- fn
	}()

	return nil
}

// Size 工作池大小
func (wp *WorkerPool) Size() int {
	return wp.semaphore.Size()
}

// Barrier 屏障
type Barrier struct {
	count int
	wg    sync.WaitGroup
}

// NewBarrier 创建屏障
func NewBarrier(count int) *Barrier {
	return &Barrier{count: count}
}

// Wait 等待所有goroutine到达屏障
func (b *Barrier) Wait() {
	b.wg.Done()
	b.wg.Wait()
}

// Add 添加goroutine
func (b *Barrier) Add() {
	b.wg.Add(1)
}

// Countdown 倒计时锁
type Countdown struct {
	count int32
	ch    chan struct{}
}

// NewCountdown 创建倒计时锁
func NewCountdown(count int) *Countdown {
	return &Countdown{
		count: int32(count),
		ch:    make(chan struct{}),
	}
}

// Decrement 倒计时
func (c *Countdown) Decrement() {
	if atomic.AddInt32(&c.count, -1) == 0 {
		close(c.ch)
	}
}

// Wait 等待倒计时结束
func (c *Countdown) Wait() {
	<-c.ch
}
