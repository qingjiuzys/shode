// Package cache 提供多级缓存功能。
package cache

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// L2CacheConfig L2 缓存配置
type L2CacheConfig struct {
	L1Size       int
	L1TTL        time.Duration
	L2TTL        time.Duration
	L2Cache      DistributedCache
	EnableL1     bool
	EnableL2     bool
	UpdateL2OnL1 bool
}

// DefaultL2CacheConfig 默认 L2 缓存配置
var DefaultL2CacheConfig = L2CacheConfig{
	L1Size:       1000,
	L1TTL:        5 * time.Minute,
	L2TTL:        1 * time.Hour,
	EnableL1:     true,
	EnableL2:     true,
	UpdateL2OnL1: true,
}

// L1Cache L1 缓存（本地缓存）
type L1Cache struct {
	items map[string]*CacheItem
	mu    sync.RWMutex
	size  int
}

// NewL1Cache 创建 L1 缓存
func NewL1Cache(size int) *L1Cache {
	return &L1Cache{
		items: make(map[string]*CacheItem),
		size:  size,
	}
}

// Get 获取缓存
func (l1 *L1Cache) Get(key string) (interface{}, bool) {
	l1.mu.RLock()
	defer l1.mu.RUnlock()

	item, exists := l1.items[key]
	if !exists || item.IsExpired() {
		return nil, false
	}

	return item.Value, true
}

// Set 设置缓存
func (l1 *L1Cache) Set(key string, value interface{}, ttl time.Duration) {
	l1.mu.Lock()
	defer l1.mu.Unlock()

	// 如果超过大小，淘汰最旧的
	if len(l1.items) >= l1.size {
		l1.evictOldest()
	}

	var expiration time.Time
	if ttl > 0 {
		expiration = time.Now().Add(ttl)
	}

	l1.items[key] = &CacheItem{
		Key:        key,
		Value:      value,
		Expiration: expiration,
		TTL:        ttl,
		Metadata:   make(map[string]interface{}),
	}
}

// Delete 删除缓存
func (l1 *L1Cache) Delete(key string) {
	l1.mu.Lock()
	defer l1.mu.Unlock()

	delete(l1.items, key)
}

// Clear 清空缓存
func (l1 *L1Cache) Clear() {
	l1.mu.Lock()
	defer l1.mu.Unlock()

	l1.items = make(map[string]*CacheItem)
}

// evictOldest 淘汰最旧的项
func (l1 *L1Cache) evictOldest() {
	var oldestKey string
	var oldestTime time.Time

	for key, item := range l1.items {
		if oldestKey == "" || item.Expiration.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.Expiration
		}
	}

	if oldestKey != "" {
		delete(l1.items, oldestKey)
	}
}

// L2Cache 两级缓存
type L2Cache struct {
	config  L2CacheConfig
	l1      *L1Cache
	l2      DistributedCache
	mu      sync.RWMutex
}

// NewL2Cache 创建两级缓存
func NewL2Cache(config L2CacheConfig) *L2Cache {
	if config.L1Size == 0 {
		config.L1Size = DefaultL2CacheConfig.L1Size
	}

	return &L2Cache{
		config: config,
		l1:     NewL1Cache(config.L1Size),
		l2:     config.L2Cache,
	}
}

// Get 获取缓存（先查 L1，再查 L2）
func (l2 *L2Cache) Get(ctx context.Context, key string, dest interface{}) error {
	// 先查 L1
	if l2.config.EnableL1 {
		if value, found := l2.l1.Get(key); found {
			// 将值复制到 dest
			// 简化实现，实际应该用序列化
			return l2.assignValue(dest, value)
		}
	}

	// 再查 L2
	if l2.config.EnableL2 {
		var value interface{}
		if err := l2.l2.Get(ctx, key, &value); err == nil {
			// 回写 L1
			if l2.config.EnableL1 && l2.config.UpdateL2OnL1 {
				l2.l1.Set(key, value, l2.config.L1TTL)
			}
			return l2.assignValue(dest, value)
		}
	}

	return fmt.Errorf("key not found: %s", key)
}

// Set 设置缓存（同时写入 L1 和 L2）
func (l2 *L2Cache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// 写入 L1
	if l2.config.EnableL1 {
		l1TTL := l2.config.L1TTL
		if ttl > 0 && ttl < l1TTL {
			l1TTL = ttl
		}
		l2.l1.Set(key, value, l1TTL)
	}

	// 写入 L2
	if l2.config.EnableL2 {
		l2TTL := l2.config.L2TTL
		if ttl > 0 && ttl < l2TTL {
			l2TTL = ttl
		}
		return l2.l2.Set(ctx, key, value, l2TTL)
	}

	return nil
}

// Delete 删除缓存（同时删除 L1 和 L2）
func (l2 *L2Cache) Delete(ctx context.Context, key string) error {
	// 删除 L1
	if l2.config.EnableL1 {
		l2.l1.Delete(key)
	}

	// 删除 L2
	if l2.config.EnableL2 {
		return l2.l2.Delete(ctx, key)
	}

	return nil
}

// Exists 检查键是否存在
func (l2 *L2Cache) Exists(ctx context.Context, key string) (bool, error) {
	// 先检查 L1
	if l2.config.EnableL1 {
		if _, found := l2.l1.Get(key); found {
			return true, nil
		}
	}

	// 再检查 L2
	if l2.config.EnableL2 {
		return l2.l2.Exists(ctx, key)
	}

	return false, nil
}

// Expire 设置过期时间
func (l2 *L2Cache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	// 更新 L1
	if l2.config.EnableL1 {
		if value, found := l2.l1.Get(key); found {
			l2.l1.Set(key, value, ttl)
		}
	}

	// 更新 L2
	if l2.config.EnableL2 {
		return l2.l2.Expire(ctx, key, ttl)
	}

	return nil
}

// TTL 获取剩余时间
func (l2 *L2Cache) TTL(ctx context.Context, key string) (time.Duration, error) {
	// 优先返回 L1 的 TTL
	if l2.config.EnableL1 {
		// L1 的 TTL 不容易获取，这里简化为查询 L2
	}

	if l2.config.EnableL2 {
		return l2.l2.TTL(ctx, key)
	}

	return 0, fmt.Errorf("cache not available")
}

// Clear 清空所有缓存
func (l2 *L2Cache) Clear(ctx context.Context) error {
	// 清空 L1
	if l2.config.EnableL1 {
		l2.l1.Clear()
	}

	// 清空 L2
	if l2.config.EnableL2 {
		return l2.l2.Clear(ctx)
	}

	return nil
}

// Keys 获取所有键
func (l2 *L2Cache) Keys(ctx context.Context, pattern string) ([]string, error) {
	// 返回 L2 的键（L2 是完整的）
	if l2.config.EnableL2 {
		return l2.l2.Keys(ctx, pattern)
	}

	return []string{}, nil
}

// GetSet 设置并返回旧值
func (l2 *L2Cache) GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error) {
	// 先从 L1 获取旧值
	var oldValue interface{}
	var found bool

	if l2.config.EnableL1 {
		oldValue, found = l2.l1.Get(key)
	}

	// 如果 L1 没有，从 L2 获取
	if !found && l2.config.EnableL2 {
		var dest interface{}
		if err := l2.l2.Get(ctx, key, &dest); err == nil {
			oldValue = dest
			found = true
		}
	}

	// 设置新值
	if err := l2.Set(ctx, key, value, ttl); err != nil {
		return nil, err
	}

	if found {
		return oldValue, nil
	}

	return nil, nil
}

// assignValue 赋值（简化实现）
func (l2 *L2Cache) assignValue(dest, value interface{}) error {
	// 简化实现，实际应该用反射或序列化
	// 这里假设 dest 是 *interface{} 类型
	if ptr, ok := dest.(*interface{}); ok {
		*ptr = value
		return nil
	}

	return fmt.Errorf("unsupported dest type")
}

// CacheBreakdown 缓存穿透防护
type CacheBreakdown struct {
	cache      DistributedCache
	nullValues map[string]bool
	mu         sync.RWMutex
}

// NewCacheBreakdown 创建缓存穿透防护
func NewCacheBreakdown(cache DistributedCache) *CacheBreakdown {
	return &CacheBreakdown{
		cache:      cache,
		nullValues: make(map[string]bool),
	}
}

// Get 获取缓存（带穿透防护）
func (cb *CacheBreakdown) Get(ctx context.Context, key string, dest interface{}) error {
	// 检查是否为 null 值
	cb.mu.RLock()
	_, isNull := cb.nullValues[key]
	cb.mu.RUnlock()

	if isNull {
		return fmt.Errorf("key is null: %s", key)
	}

	// 正常获取
	return cb.cache.Get(ctx, key, dest)
}

// Set 设置缓存
func (cb *CacheBreakdown) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	if value == nil {
		// 记录 null 值
		cb.mu.Lock()
		cb.nullValues[key] = true
		cb.mu.Unlock()

		// 短暂缓存 null 值
		return cb.cache.Set(ctx, key, "", 30*time.Second)
	}

	return cb.cache.Set(ctx, key, value, ttl)
}

// Delete 删除缓存
func (cb *CacheBreakdown) Delete(ctx context.Context, key string) error {
	cb.mu.Lock()
	delete(cb.nullValues, key)
	cb.mu.Unlock()

	return cb.cache.Delete(ctx, key)
}

// Exists 检查键是否存在
func (cb *CacheBreakdown) Exists(ctx context.Context, key string) (bool, error) {
	return cb.cache.Exists(ctx, key)
}

// Expire 设置过期时间
func (cb *CacheBreakdown) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return cb.cache.Expire(ctx, key, ttl)
}

// TTL 获取剩余时间
func (cb *CacheBreakdown) TTL(ctx context.Context, key string) (time.Duration, error) {
	return cb.cache.TTL(ctx, key)
}

// Clear 清空缓存
func (cb *CacheBreakdown) Clear(ctx context.Context) error {
	cb.mu.Lock()
	cb.nullValues = make(map[string]bool)
	cb.mu.Unlock()

	return cb.cache.Clear(ctx)
}

// Keys 获取所有键
func (cb *CacheBreakdown) Keys(ctx context.Context, pattern string) ([]string, error) {
	return cb.cache.Keys(ctx, pattern)
}

// GetSet 设置并返回旧值
func (cb *CacheBreakdown) GetSet(ctx context.Context, key string, value interface{}, ttl time.Duration) (interface{}, error) {
	return cb.cache.GetSet(ctx, key, value, ttl)
}
