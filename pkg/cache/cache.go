// Package cache 提供线程安全的内存缓存实现。
//
// 缓存特点：
//   - 支持键值对存储
//   - 支持过期时间 (TTL)
//   - 自动清理过期条目
//   - 线程安全（使用读写锁）
//   - 支持通配符删除
//
// 使用示例：
//
//	cache := cache.NewCache()
//	cache.Set("key", "value", 60) // 存储 60 秒
//	value, found := cache.Get("key")
//
// 缓存会在后台定期清理过期条目，无需手动管理。
package cache

import (
	"strings"
	"sync"
	"time"
)

// CacheEntry represents a cache entry with value and expiration
type CacheEntry struct {
	Value     string
	ExpiresAt time.Time
}

// Cache provides thread-safe in-memory caching
type Cache struct {
	entries map[string]*CacheEntry
	mu      sync.RWMutex
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	c := &Cache{
		entries: make(map[string]*CacheEntry),
	}
	// Start cleanup goroutine
	go c.cleanupExpired()
	return c
}

// Set stores a value in the cache with optional TTL
// If ttlSeconds is 0 or negative, the entry never expires
func (c *Cache) Set(key, value string, ttlSeconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &CacheEntry{
		Value: value,
	}

	if ttlSeconds > 0 {
		entry.ExpiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	} else {
		// Never expires
		entry.ExpiresAt = time.Time{}
	}

	c.entries[key] = entry
}

// Get retrieves a value from the cache
// Returns the value and true if found and not expired, false otherwise
func (c *Cache) Get(key string) (string, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return "", false
	}

	// Check if expired
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		// Entry expired, but we don't delete it here (cleanup goroutine will)
		return "", false
	}

	return entry.Value, true
}

// Delete removes a key from the cache
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, key)
}

// Clear removes all entries from the cache
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = make(map[string]*CacheEntry)
}

// Exists checks if a key exists and is not expired
func (c *Cache) Exists(key string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return false
	}

	// Check if expired
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		return false
	}

	return true
}

// GetTTL returns the remaining TTL in seconds for a key
// Returns -1 if the key doesn't exist or never expires
func (c *Cache) GetTTL(key string) int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.entries[key]
	if !exists {
		return -1
	}

	// If never expires
	if entry.ExpiresAt.IsZero() {
		return -1
	}

	remaining := time.Until(entry.ExpiresAt)
	if remaining <= 0 {
		return 0 // Expired
	}

	return int(remaining.Seconds())
}

// SetBatch sets multiple key-value pairs at once
// ttlSeconds applies to all entries
func (c *Cache) SetBatch(keyValues map[string]string, ttlSeconds int) {
	c.mu.Lock()
	defer c.mu.Unlock()

	expiresAt := time.Time{}
	if ttlSeconds > 0 {
		expiresAt = time.Now().Add(time.Duration(ttlSeconds) * time.Second)
	}

	for key, value := range keyValues {
		c.entries[key] = &CacheEntry{
			Value:     value,
			ExpiresAt: expiresAt,
		}
	}
}

// GetKeys returns all keys matching a pattern
// Pattern supports simple wildcard: * matches any characters
func (c *Cache) GetKeys(pattern string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var keys []string
	now := time.Now()

	for key, entry := range c.entries {
		// Skip expired entries
		if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
			continue
		}

		// Simple pattern matching
		if matchesPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys
}

// matchesPattern performs simple wildcard pattern matching
func matchesPattern(str, pattern string) bool {
	if pattern == "*" {
		return true
	}

	// Simple prefix/suffix matching
	if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
		// Contains pattern
		substr := pattern[1 : len(pattern)-1]
		return strings.Contains(str, substr)
	} else if strings.HasPrefix(pattern, "*") {
		// Suffix pattern
		suffix := pattern[1:]
		return strings.HasSuffix(str, suffix)
	} else if strings.HasSuffix(pattern, "*") {
		// Prefix pattern
		prefix := pattern[:len(pattern)-1]
		return strings.HasPrefix(str, prefix)
	}

	// Exact match
	return str == pattern
}

// cleanupExpired periodically removes expired entries
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		now := time.Now()
		for key, entry := range c.entries {
			if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
				delete(c.entries, key)
			}
		}
		c.mu.Unlock()
	}
}

// GetStats returns cache statistics
func (c *Cache) GetStats() (total, expired int) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	total = len(c.entries)

	for _, entry := range c.entries {
		if !entry.ExpiresAt.IsZero() && now.After(entry.ExpiresAt) {
			expired++
		}
	}

	return total, expired
}
