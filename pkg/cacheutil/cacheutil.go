// Package cacheutil 提供缓存工具
package cacheutil

import (
	"sync"
	"time"
)

// Cache 缓存接口
type Cache[K comparable, V any] interface {
	Get(key K) (V, bool)
	Set(key K, value V)
	Delete(key K)
	Clear()
	Len() int
	Keys() []K
	Values() []V
}

// SimpleCache 简单缓存
type SimpleCache[K comparable, V any] struct {
	mu    sync.RWMutex
	data  map[K]V
 maxSize int
}

// NewSimpleCache 创建简单缓存
func NewSimpleCache[K comparable, V any](maxSize int) *SimpleCache[K, V] {
	return &SimpleCache[K, V]{
		data:    make(map[K]V),
		maxSize: maxSize,
	}
}

// Get 获取值
func (c *SimpleCache[K, V]) Get(key K) (V, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.data[key]
	return value, exists
}

// Set 设置值
func (c *SimpleCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.data) >= c.maxSize {
		// 简单的驱逐策略：删除第一个
		for k := range c.data {
			delete(c.data, k)
			break
		}
	}

	c.data[key] = value
}

// Delete 删除值
func (c *SimpleCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

// Clear 清空缓存
func (c *SimpleCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[K]V)
}

// Len 获取长度
func (c *SimpleCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Keys 获取所有键
func (c *SimpleCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取所有值
func (c *SimpleCache[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.data))
	for _, v := range c.data {
		values = append(values, v)
	}
	return values
}

// TTLCache 带过期时间的缓存
type TTLCache[K comparable, V any] struct {
	mu       sync.RWMutex
	data     map[K]*ttlItem[V]
	maxSize  int
	ttl      time.Duration
	onEvict  func(K, V)
}

type ttlItem[V any] struct {
	value      V
	expiration time.Time
}

// NewTTLCache 创建TTL缓存
func NewTTLCache[K comparable, V any](maxSize int, ttl time.Duration) *TTLCache[K, V] {
	cache := &TTLCache[K, V]{
		data:    make(map[K]*ttlItem[V]),
		maxSize: maxSize,
		ttl:     ttl,
	}

	// 启动清理goroutine
	go cache.cleanup()

	return cache
}

// Get 获取值
func (c *TTLCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.data[key]
	if !exists {
		var zero V
		return zero, false
	}

	if time.Now().After(item.expiration) {
		delete(c.data, key)
		if c.onEvict != nil {
			c.onEvict(key, item.value)
		}
		var zero V
		return zero, false
	}

	return item.value, true
}

// Set 设置值
func (c *TTLCache[K, V]) Set(key K, value V) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL 设置值（指定TTL）
func (c *TTLCache[K, V]) SetWithTTL(key K, value V, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.maxSize > 0 && len(c.data) >= c.maxSize {
		c.evictOne()
	}

	c.data[key] = &ttlItem[V]{
		value:      value,
		expiration: time.Now().Add(ttl),
	}
}

// Delete 删除值
func (c *TTLCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if item, exists := c.data[key]; exists {
		delete(c.data, key)
		if c.onEvict != nil {
			c.onEvict(key, item.value)
		}
	}
}

// Clear 清空缓存
func (c *TTLCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.onEvict != nil {
		for k, item := range c.data {
			c.onEvict(k, item.value)
		}
	}

	c.data = make(map[K]*ttlItem[V])
}

// Len 获取长度
func (c *TTLCache[K, V]) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.data)
}

// Keys 获取所有键
func (c *TTLCache[K, V]) Keys() []K {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]K, 0, len(c.data))
	for k := range c.data {
		keys = append(keys, k)
	}
	return keys
}

// Values 获取所有值
func (c *TTLCache[K, V]) Values() []V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	values := make([]V, 0, len(c.data))
	for _, item := range c.data {
		values = append(values, item.value)
	}
	return values
}

// SetOnEvict 设置驱逐回调
func (c *TTLCache[K, V]) SetOnEvict(fn func(K, V)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = fn
}

// cleanup 定期清理过期项
func (c *TTLCache[K, V]) cleanup() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for k, item := range c.data {
			if now.After(item.expiration) {
				if c.onEvict != nil {
					c.onEvict(k, item.value)
				}
				delete(c.data, k)
			}
		}
		c.mu.Unlock()
	}
}

// evictOne 驱逐一个元素
func (c *TTLCache[K, V]) evictOne() {
	for k, item := range c.data {
		if c.onEvict != nil {
			c.onEvict(k, item.value)
		}
		delete(c.data, k)
		return
	}
}

// LoadingCache 支持懒加载的缓存
type LoadingCache[K comparable, V any] struct {
	cache  *SimpleCache[K, V]
	loader func(K) (V, error)
	mu     sync.Mutex
}

// NewLoadingCache 创建LoadingCache
func NewLoadingCache[K comparable, V any](maxSize int, loader func(K) (V, error)) *LoadingCache[K, V] {
	return &LoadingCache[K, V]{
		cache:  NewSimpleCache[K, V](maxSize),
		loader: loader,
	}
}

// Get 获取值（如果不存在则加载）
func (c *LoadingCache[K, V]) Get(key K) (V, error) {
	// 先尝试从缓存获取
	if value, exists := c.cache.Get(key); exists {
		return value, nil
	}

	// 使用锁防止重复加载
	c.mu.Lock()
	defer c.mu.Unlock()

	// 再次检查（可能其他goroutine已经加载）
	if value, exists := c.cache.Get(key); exists {
		return value, nil
	}

	// 加载值
	value, err := c.loader(key)
	if err != nil {
		var zero V
		return zero, err
	}

	c.cache.Set(key, value)
	return value, nil
}

// GetOr 获取值或返回默认值
func (c *LoadingCache[K, V]) GetOr(key K, defaultValue V) V {
	value, err := c.Get(key)
	if err != nil {
		return defaultValue
	}
	return value
}

// Set 设置值
func (c *LoadingCache[K, V]) Set(key K, value V) {
	c.cache.Set(key, value)
}

// Delete 删除值
func (c *LoadingCache[K, V]) Delete(key K) {
	c.cache.Delete(key)
}

// Clear 清空缓存
func (c *LoadingCache[K, V]) Clear() {
	c.cache.Clear()
}

// Len 获取长度
func (c *LoadingCache[K, V]) Len() int {
	return c.cache.Len()
}

// Keys 获取所有键
func (c *LoadingCache[K, V]) Keys() []K {
	return c.cache.Keys()
}

// Values 获取所有值
func (c *LoadingCache[K, V]) Values() []V {
	return c.cache.Values()
}

// Stats 缓存统计
type Stats struct {
	Hits     int64
	Misses   int64
	Evictions int64
}

// CachedCache 带统计的缓存
type CachedCache[K comparable, V any] struct {
	cache   Cache[K, V]
	stats   Stats
	statsMu sync.RWMutex
}

// NewCachedCache 创建带统计的缓存
func NewCachedCache[K comparable, V any](cache Cache[K, V]) *CachedCache[K, V] {
	return &CachedCache[K, V]{
		cache: cache,
	}
}

// Get 获取值
func (c *CachedCache[K, V]) Get(key K) (V, bool) {
	value, exists := c.cache.Get(key)

	c.statsMu.Lock()
	if exists {
		c.stats.Hits++
	} else {
		c.stats.Misses++
	}
	c.statsMu.Unlock()

	return value, exists
}

// Set 设置值
func (c *CachedCache[K, V]) Set(key K, value V) {
	c.cache.Set(key, value)
}

// Delete 删除值
func (c *CachedCache[K, V]) Delete(key K) {
	c.cache.Delete(key)
}

// Clear 清空缓存
func (c *CachedCache[K, V]) Clear() {
	c.cache.Clear()
}

// Len 获取长度
func (c *CachedCache[K, V]) Len() int {
	return c.cache.Len()
}

// Keys 获取所有键
func (c *CachedCache[K, V]) Keys() []K {
	return c.cache.Keys()
}

// Values 获取所有值
func (c *CachedCache[K, V]) Values() []V {
	return c.cache.Values()
}

// Stats 获取统计信息
func (c *CachedCache[K, V]) Stats() Stats {
	c.statsMu.RLock()
	defer c.statsMu.RUnlock()
	return c.stats
}

// HitRate 获取命中率
func (c *CachedCache[K, V]) HitRate() float64 {
	stats := c.Stats()
	total := stats.Hits + stats.Misses
	if total == 0 {
		return 0
	}
	return float64(stats.Hits) / float64(total)
}

// SizeCache 固定大小的缓存（FIFO）
type SizeCache[K comparable, V any] struct {
	mu     sync.Mutex
	data   map[K]V
	keys   []K
	maxSize int
}

// NewSizeCache 创建固定大小缓存
func NewSizeCache[K comparable, V any](maxSize int) *SizeCache[K, V] {
	return &SizeCache[K, V]{
		data:    make(map[K]V),
		keys:    make([]K, 0, maxSize),
		maxSize: maxSize,
	}
}

// Get 获取值
func (c *SizeCache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	value, exists := c.data[key]
	return value, exists
}

// Set 设置值
func (c *SizeCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 如果已存在，更新值
	if _, exists := c.data[key]; exists {
		c.data[key] = value
		return
	}

	// 如果达到最大容量，删除最老的
	if len(c.keys) >= c.maxSize {
	 oldest := c.keys[0]
		delete(c.data, oldest)
		c.keys = c.keys[1:]
	}

	// 添加新值
	c.data[key] = value
	c.keys = append(c.keys, key)
}

// Delete 删除值
func (c *SizeCache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.data[key]; exists {
		delete(c.data, key)
		for i, k := range c.keys {
			if k == key {
				c.keys = append(c.keys[:i], c.keys[i+1:]...)
				break
			}
		}
	}
}

// Clear 清空缓存
func (c *SizeCache[K, V]) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = make(map[K]V)
	c.keys = make([]K, 0, c.maxSize)
}

// Len 获取长度
func (c *SizeCache[K, V]) Len() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.data)
}

// Keys 获取所有键
func (c *SizeCache[K, V]) Keys() []K {
	c.mu.Lock()
	defer c.mu.Unlock()

	keys := make([]K, len(c.keys))
	copy(keys, c.keys)
	return keys
}

// Values 获取所有值
func (c *SizeCache[K, V]) Values() []V {
	c.mu.Lock()
	defer c.mu.Unlock()

	values := make([]V, 0, len(c.keys))
	for _, k := range c.keys {
		values = append(values, c.data[k])
	}
	return values
}

// Memoize 函数记忆化
func Memoize[K comparable, V any](fn func(K) V) func(K) V {
	cache := NewSimpleCache[K, V](0)

	return func(key K) V {
		if value, exists := cache.Get(key); exists {
			return value
		}

		value := fn(key)
		cache.Set(key, value)
		return value
	}
}

// MemoizeWithError 带错误的函数记忆化
func MemoizeWithError[K comparable, V any](fn func(K) (V, error)) func(K) (V, error) {
	cache := NewSimpleCache[K, V](0)

	return func(key K) (V, error) {
		if value, exists := cache.Get(key); exists {
			return value, nil
		}

		value, err := fn(key)
		if err != nil {
			var zero V
			return zero, err
		}

		cache.Set(key, value)
		return value, nil
	}
}
