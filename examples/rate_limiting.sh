#!/usr/bin/env shode

# API Rate Limiting Example
# Demonstrates rate limiting using cache

Println "=== API Rate Limiting Demo ==="

# Simulate rate limiting for a user
userId = "user123"
rateLimitKey = "ratelimit:" + userId
maxRequests = 5
windowSeconds = 60

Println "Rate limit: " + maxRequests + " requests per " + windowSeconds + " seconds"
Println "User ID: " + userId
Println ""

# Simulate multiple API requests
for i in 1 2 3 4 5 6 7
do
    Println "Request #" + i + ":"
    
    # Check current request count
    currentCountStr = GetCache rateLimitKey
    currentCount = 0
    if currentCountStr != "" {
        # Parse count (simplified)
        currentCount = currentCountStr
    }
    
    Println "  Current count: " + currentCount
    
    # Check if rate limit exceeded
    if currentCount >= maxRequests
    then
        Println "  Rate limit exceeded! Request blocked"
        break
    fi
    
    # Increment request count
    newCount = currentCount + 1
    SetCache rateLimitKey newCount windowSeconds
    Println "  Request allowed. New count: " + newCount
    
    # Small delay
    sleep 0.1
done

# Check final count
Println ""
Println "Final rate limit count:"
finalCount = GetCache rateLimitKey
Println "Count: " + finalCount

# Get TTL
ttl = GetCacheTTL rateLimitKey
Println "TTL remaining: " + ttl + " seconds"

Println ""
Println "=== Rate Limiting Demo Complete ==="
