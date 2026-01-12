package engine

import (
	"hash/fnv"
	"sync"
	"time"
)

// CommandCache caches command execution results
type CommandCache struct {
	cache     map[uint64]*cacheEntry
	maxSize   int
	mu        sync.RWMutex
	hitCount  int64
	missCount int64
}

// cacheEntry represents a cached command result
type cacheEntry struct {
	result    *CommandResult
	timestamp time.Time
	expires   time.Time
}

// NewCommandCache creates a new command cache
func NewCommandCache(maxSize int) *CommandCache {
	return &CommandCache{
		cache:   make(map[uint64]*cacheEntry),
		maxSize: maxSize,
	}
}

// Get retrieves a command result from cache
// Optimized: Reduce lock contention by checking expiration before incrementing counters
func (cc *CommandCache) Get(cmd string, args []string) (*CommandResult, bool) {
	key := cc.generateKey(cmd, args)
	now := time.Now()

	cc.mu.RLock()
	entry, exists := cc.cache[key]
	cc.mu.RUnlock()

	if !exists {
		cc.mu.Lock()
		cc.missCount++
		cc.mu.Unlock()
		return nil, false
	}

	// Check if entry is expired
	if now.After(entry.expires) {
		cc.mu.Lock()
		// Double-check after acquiring write lock
		if entry, stillExists := cc.cache[key]; stillExists && now.After(entry.expires) {
			delete(cc.cache, key)
		}
		cc.missCount++
		cc.mu.Unlock()
		return nil, false
	}

	cc.mu.Lock()
	cc.hitCount++
	cc.mu.Unlock()
	return entry.result, true
}

// Put stores a command result in cache
func (cc *CommandCache) Put(cmd string, args []string, result *CommandResult) {
	key := cc.generateKey(cmd, args)

	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Evict if cache is full
	if len(cc.cache) >= cc.maxSize {
		cc.evictOldest()
	}

	// Determine expiration time based on command type
	expires := time.Now().Add(cc.getExpirationDuration(cmd))

	cc.cache[key] = &cacheEntry{
		result:    result,
		timestamp: time.Now(),
		expires:   expires,
	}
}

// Clear removes all entries from cache
func (cc *CommandCache) Clear() {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	cc.cache = make(map[uint64]*cacheEntry)
	cc.hitCount = 0
	cc.missCount = 0
}

// Stats returns cache statistics
func (cc *CommandCache) Stats() (hits, misses int64, size int) {
	cc.mu.RLock()
	defer cc.mu.RUnlock()

	return cc.hitCount, cc.missCount, len(cc.cache)
}

// generateKey generates a hash key for the command
func (cc *CommandCache) generateKey(cmd string, args []string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(cmd))

	for _, arg := range args {
		h.Write([]byte{0}) // separator
		h.Write([]byte(arg))
	}

	return h.Sum64()
}

// evictOldest evicts the oldest entry from cache
// Optimized: Use single pass with early exit for better performance
func (cc *CommandCache) evictOldest() {
	if len(cc.cache) == 0 {
		return
	}

	var oldestKey uint64
	var oldestTime time.Time
	first := true

	for key, entry := range cc.cache {
		if first || entry.timestamp.Before(oldestTime) {
			oldestTime = entry.timestamp
			oldestKey = key
			first = false
		}
	}

	if oldestKey != 0 {
		delete(cc.cache, oldestKey)
	}
}

// getExpirationDuration returns the cache expiration duration for a command
func (cc *CommandCache) getExpirationDuration(cmd string) time.Duration {
	// Different expiration times based on command type
	switch cmd {
	case "ls", "pwd", "whoami", "date":
		return 30 * time.Second // Short expiration for frequently changing commands
	case "echo", "cat", "grep":
		return 2 * time.Minute // Medium expiration for text processing
	case "find", "stat", "file":
		return 5 * time.Minute // Longer expiration for file system commands
	default:
		return 1 * time.Minute // Default expiration
	}
}

// Invalidate invalidates a specific command from cache
func (cc *CommandCache) Invalidate(cmd string, args []string) {
	key := cc.generateKey(cmd, args)

	cc.mu.Lock()
	delete(cc.cache, key)
	cc.mu.Unlock()
}

// InvalidateAllWithPrefix invalidates all commands with a given prefix
func (cc *CommandCache) InvalidateAllWithPrefix(prefix string) {
	cc.mu.Lock()
	defer cc.mu.Unlock()

	// Note: This is a simplified implementation
	// In real implementation, we would need to store the original command
	// or use a different data structure for prefix-based invalidation
	// For now, we'll clear the entire cache if prefix-based invalidation is needed
	if prefix != "" {
		cc.cache = make(map[uint64]*cacheEntry)
	}
}
