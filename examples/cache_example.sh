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
