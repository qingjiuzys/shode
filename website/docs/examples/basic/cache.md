# 缓存系统示例

## 简介

这个示例展示如何使用 Shode 的缓存系统，包括设置、获取、删除缓存等操作，以及 TTL（生存时间）的使用。

## 代码

```shode
#!/usr/bin/env shode

# Cache Example
# Demonstrates cache operations with TTL

Println "Cache Example"

# Set cache with TTL
SetCache "key1" "value1" 60
SetCache "key2" "value2" 120
SetCache "key3" "value3" 0  # No expiration

# Get cache
value1 = GetCache "key1"
Println "key1: " + value1

# Check if key exists
exists = CacheExists "key1"
Println "key1 exists: " + exists

# Get TTL
ttl = GetCacheTTL "key1"
Println "key1 TTL: " + ttl + " seconds"

# Get all keys matching pattern
keys = GetCacheKeys "key*"
Println "Keys matching 'key*': " + keys

# Delete cache
DeleteCache "key1"
Println "Deleted key1"

# Clear all cache
ClearCache
Println "Cleared all cache"
```

## 运行方式

```bash
shode run examples/cache_example.sh
```

## 预期输出

```
Cache Example
key1: value1
key1 exists: true
key1 TTL: 60 seconds
Keys matching 'key*': key1,key2,key3
Deleted key1
Cleared all cache
```

## 功能说明

### TTL（生存时间）

- `SetCache "key" "value" 60` - 设置缓存，60秒后过期
- `SetCache "key" "value" 0` - 设置缓存，永不过期
- `GetCacheTTL "key"` - 获取剩余 TTL

### 模式匹配

```shode
# 获取所有匹配模式的键
keys = GetCacheKeys "user:*"
# 返回所有以 "user:" 开头的键
```

### 批量操作

```shode
# 批量设置缓存
keyValues = {"key1": "value1", "key2": "value2", "key3": "value3"}
SetCacheBatch keyValues 300
```

## 使用场景

- **API 响应缓存**: 缓存 API 查询结果，减少数据库访问
- **会话存储**: 存储用户会话信息
- **计算结果缓存**: 缓存耗时计算的结果
- **限流计数**: 使用缓存实现限流功能

## 相关文档

- [用户指南 - 缓存系统](../../guides/user-guide.md#7-缓存系统)
- [API 参考 - 缓存函数](../../api/stdlib.md#缓存函数)
