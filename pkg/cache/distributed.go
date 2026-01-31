// Package cache 提供分布式缓存功能。
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// CacheItem 缓存项
type CacheItem struct {
	Key        string
	Value      interface{}
	Expiration time.Time
	TTL        time.Duration
	Metadata   map[string]interface{}
}

// IsExpired 检查是否过期
func (ci *CacheItem) IsExpired() bool {
	return !ci.Expiration.IsZero() && time.Now().After(ci.Expiration)
}

// DistributedCache 分布式缓存接口
type DistributedCache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Clear(ctx context.Context) error
	Keys(ctx context.Context, pattern string) ([]string, error)
	GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error)
}

// MemoryDistributedCache 内存分布式缓存实现
type MemoryDistributedCache struct {
	items    map[string]*CacheItem
	mu       sync.RWMutex
	onEvict  func(key string, value interface{})
	stopChan chan struct{}
}

// NewMemoryDistributedCache 创建内存分布式缓存
func NewMemoryDistributedCache() *MemoryDistributedCache {
	cache := &MemoryDistributedCache{
		items:    make(map[string]*CacheItem),
		stopChan: make(chan struct{}),
	}

	// 启动清理 goroutine
	go cache.cleanupLoop()

	return cache
}

// cleanupLoop 定期清理过期项
func (c *MemoryDistributedCache) cleanupLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

// cleanup 清理过期项
func (c *MemoryDistributedCache) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if item.IsExpired() {
			if c.onEvict != nil {
				c.onEvict(key, item.Value)
			}
			delete(c.items, key)
		}
	}
}

// Get 获取缓存
func (c *MemoryDistributedCache) Get(ctx context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	if item.IsExpired() {
		return fmt.Errorf("key expired: %s", key)
	}

	// 序列化转换
	data, err := json.Marshal(item.Value)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, dest)
}

// Set 设置缓存
func (c *MemoryDistributedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &CacheItem{
		Key:        key,
		Value:      value,
		Expiration: expiration,
		TTL:        ttl,
		Metadata:   make(map[string]interface{}),
	}

	return nil
}

// Delete 删除缓存
func (c *MemoryDistributedCache) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		delete(c.items, key)
		return nil
	}

	return fmt.Errorf("key not found: %s", key)
}

// Exists 检查键是否存在
func (c *MemoryDistributedCache) Exists(ctx context.Context, key string) (bool, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return false, nil
	}

	return !item.IsExpired(), nil
}

// Expire 设置过期时间
func (c *MemoryDistributedCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	if !exists {
		return fmt.Errorf("key not found: %s", key)
	}

	item.TTL = ttl
	item.Expiration = time.Now().Add(ttl)

	return nil
}

// TTL 获取剩余时间
func (c *MemoryDistributedCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return 0, fmt.Errorf("key not found: %s", key)
	}

	if item.Expiration.IsZero() {
		return -1, nil // 永不过期
	}

	remaining := time.Until(item.Expiration)
	if remaining < 0 {
		return 0, fmt.Errorf("key expired: %s", key)
	}

	return remaining, nil
}

// Clear 清空所有缓存
func (c *MemoryDistributedCache) Clear(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*CacheItem)
	return nil
}

// Keys 获取所有键
func (c *MemoryDistributedCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]string, 0)
	for key := range c.items {
		// 简单的模式匹配（实际应该用完整的 glob 匹配）
		if pattern == "*" || key == pattern {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

// GetSet 设置并返回旧值
func (c *MemoryDistributedCache) GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, exists := c.items[key]
	oldValue := interface{}(nil)
	if exists && !item.IsExpired() {
		oldValue = item.Value
	}

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	c.items[key] = &CacheItem{
		Key:        key,
		Value:      value,
		Expiration: expiration,
		TTL:        ttl,
		Metadata:   make(map[string]interface{}),
	}

	return oldValue, nil
}

// Stop 停止缓存
func (c *MemoryDistributedCache) Stop() {
	close(c.stopChan)
}

// SetOnEvict 设置淘汰回调
func (c *MemoryDistributedCache) SetOnEvict(fn func(key string, value interface{})) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onEvict = fn
}

// CacheStatistics 缓存统计
type CacheStatistics struct {
	Hits     int64
	Misses   int64
	Deletions int64
	Evictions int64
}

// StatisticsCache 支持统计的缓存
type StatisticsCache struct {
	cache       DistributedCache
	statistics  CacheStatistics
	mu          sync.RWMutex
}

// NewStatisticsCache 创建支持统计的缓存
func NewStatisticsCache(cache DistributedCache) *StatisticsCache {
	return &StatisticsCache{
		cache: cache,
	}
}

// Get 获取并记录统计
func (sc *StatisticsCache) Get(ctx context.Context, key string, dest interface{}) error {
	err := sc.cache.Get(ctx, key, dest)

	sc.mu.Lock()
	defer sc.mu.Unlock()

	if err != nil {
		sc.statistics.Misses++
	} else {
		sc.statistics.Hits++
	}

	return err
}

// Set 设置缓存
func (sc *StatisticsCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return sc.cache.Set(ctx, key, value, ttl)
}

// Delete 删除缓存
func (sc *StatisticsCache) Delete(ctx context.Context, key string) error {
	err := sc.cache.Delete(ctx, key)

	sc.mu.Lock()
	defer sc.mu.Unlock()
	if err == nil {
		sc.statistics.Deletions++
	}

	return err
}

// Exists 检查键是否存在
func (sc *StatisticsCache) Exists(ctx context.Context, key string) (bool, error) {
	return sc.cache.Exists(ctx, key)
}

// Expire 设置过期时间
func (sc *StatisticsCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return sc.cache.Expire(ctx, key, ttl)
}

// TTL 获取剩余时间
func (sc *StatisticsCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return sc.cache.TTL(ctx, key)
}

// Clear 清空缓存
func (sc *StatisticsCache) Clear(ctx context.Context) error {
	return sc.cache.Clear(ctx)
}

// Keys 获取所有键
func (sc *StatisticsCache) Keys(ctx context.Context, pattern string) ([]string, error) {
	return sc.cache.Keys(ctx, pattern)
}

// GetSet 设置并返回旧值
func (sc *StatisticsCache) GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error) {
	return sc.cache.GetSet(ctx, key, value, ttl)
}

// GetStatistics 获取统计信息
func (sc *StatisticsCache) GetStatistics() CacheStatistics {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	return sc.statistics
}

// ResetStatistics 重置统计
func (sc *StatisticsCache) ResetStatistics() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.statistics = CacheStatistics{}
}

// HitRate 计算命中率
func (sc *StatisticsCache) HitRate() float64 {
	sc.mu.RLock()
	defer sc.mu.RUnlock()

	total := sc.statistics.Hits + sc.statistics.Misses
	if total == 0 {
		return 0
	}

	return float64(sc.statistics.Hits) / float64(total)
}
