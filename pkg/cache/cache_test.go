package cache

import (
	"testing"
	"time"
)

// TestNewCache 测试创建缓存
func TestNewCache(t *testing.T) {
	c := NewCache()
	if c == nil {
		t.Fatal("NewCache() returned nil")
	}
}

// TestSetAndGet 测试设置和获取
func TestSetAndGet(t *testing.T) {
	c := NewCache()
	
	c.Set("key1", "value1", 60)
	value, exists := c.Get("key1")
	
	if !exists {
		t.Error("Get() should return true for existing key")
	}
	if value != "value1" {
		t.Errorf("Get() = %v, want %v", value, "value1")
	}
}

// TestSetAndGet_NonExistent 测试获取不存在的键
func TestSetAndGet_NonExistent(t *testing.T) {
	c := NewCache()
	
	value, exists := c.Get("nonexistent")
	if exists {
		t.Error("Get() should return false for non-existent key")
	}
	if value != "" {
		t.Errorf("Get() should return empty string for non-existent key, got %v", value)
	}
}

// TestDelete 测试删除缓存
func TestDelete(t *testing.T) {
	c := NewCache()
	
	c.Set("key1", "value1", 60)
	c.Delete("key1")
	
	_, exists := c.Get("key1")
	if exists {
		t.Error("Get() should return false after Delete()")
	}
}

// TestClear 测试清空缓存
func TestClear(t *testing.T) {
	c := NewCache()
	
	c.Set("key1", "value1", 60)
	c.Set("key2", "value2", 60)
	c.Clear()
	
	_, exists1 := c.Get("key1")
	_, exists2 := c.Get("key2")
	
	if exists1 || exists2 {
		t.Error("Cache should be empty after Clear()")
	}
}

// TestExists 测试检查键是否存在
func TestExists(t *testing.T) {
	c := NewCache()
	
	if c.Exists("key1") {
		t.Error("Exists() should return false for non-existent key")
	}
	
	c.Set("key1", "value1", 60)
	if !c.Exists("key1") {
		t.Error("Exists() should return true for existing key")
	}
}

// TestTTL 测试缓存过期
func TestTTL(t *testing.T) {
	c := NewCache()
	
	// 设置 1 秒 TTL
	c.Set("key1", "value1", 1)
	
	// 立即检查，应该存在
	if !c.Exists("key1") {
		t.Error("key1 should exist immediately after Set()")
	}
	
	// 等待 2 秒
	time.Sleep(2 * time.Second)
	
	// 应该已过期
	if c.Exists("key1") {
		t.Error("key1 should be expired after TTL")
	}
}

// TestSetBatch 批量设置测试
func TestSetBatch(t *testing.T) {
	c := NewCache()
	
	items := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	
	c.SetBatch(items, 60)
	
	for key, expectedValue := range items {
		value, exists := c.Get(key)
		if !exists {
			t.Errorf("key %s should exist after SetBatch()", key)
		}
		if value != expectedValue {
			t.Errorf("key %s = %v, want %v", key, value, expectedValue)
		}
	}
}

// TestGetKeys 获取所有键测试
func TestGetKeys(t *testing.T) {
	c := NewCache()
	
	c.Set("key1", "value1", 60)
	c.Set("key2", "value2", 60)
	
	keys := c.GetKeys("key*")
	
	if len(keys) != 2 {
		t.Errorf("GetKeys() returned %d items, want 2", len(keys))
	}
}

// TestGetStats 测试缓存统计
func TestGetStats(t *testing.T) {
	c := NewCache()
	
	total, _ := c.GetStats()
	
	c.Set("key1", "value1", 60)
	
	total2, _ := c.GetStats()
	if total2 != total + 1 {
		t.Errorf("GetStats() total = %d, want %d", total2, total+1)
	}
}

// TestSetOverwrite 测试覆盖已存在的键
func TestSetOverwrite(t *testing.T) {
	c := NewCache()
	
	c.Set("key1", "value1", 60)
	c.Set("key1", "value2", 60)
	
	value, _ := c.Get("key1")
	if value != "value2" {
		t.Errorf("key1 = %v, want %v", value, "value2")
	}
}

// TestGetTTL 测试获取剩余 TTL
func TestGetTTL(t *testing.T) {
	c := NewCache()
	
	c.Set("key1", "value1", 60)
	
	ttl := c.GetTTL("key1")
	if ttl <= 0 || ttl > 60 {
		t.Errorf("GetTTL() = %d, want value between 1 and 60", ttl)
	}
}

// TestNoExpiration 测试永不过期的缓存
func TestNoExpiration(t *testing.T) {
	c := NewCache()
	
	// 设置 TTL 为 0，表示永不过期
	c.Set("key1", "value1", 0)
	
	// 等待一小段时间
	time.Sleep(100 * time.Millisecond)
	
	// 应该仍然存在
	if !c.Exists("key1") {
		t.Error("key1 should not expire when TTL is 0")
	}
}
