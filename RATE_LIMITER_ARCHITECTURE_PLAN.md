# Rate Limiter Architecture Plan

## Current State ✅

**Fixed Issues:**
- ✅ **Single Shared Instance**: All handlers now use one shared rate limiter instance
- ✅ **Consistent Rate Limits**: Avatar movement now allows 60 requests/minute (1/second)
- ✅ **Proper Reset Time Calculation**: Fixed sliding window reset time calculation
- ✅ **Memory Efficient**: No duplicate rate limiter instances

**Current Implementation:**
- **Production**: Uses `SimpleRateLimiter` (in-memory, single server instance)
- **Tests**: Use mocks (`MockRateLimiter`)
- **Location**: `backend/internal/server/server.go` - single shared instance

## Rate Limits by Action Type

| Action | Limit | Window | Use Case |
|--------|-------|--------|----------|
| `ActionUpdateAvatar` | 60 requests | 1 minute | Avatar movement (1/second) |
| `ActionCreateSession` | 10 requests | 1 minute | Session creation |
| `ActionCreatePOI` | 30 requests | 1 minute | POI creation |
| `ActionJoinPOI` | 30 requests | 1 minute | Joining POIs |
| `ActionLeavePOI` | 30 requests | 1 minute | Leaving POIs |
| `ActionUpdatePOI` | 30 requests | 1 minute | POI updates |
| `ActionDeletePOI` | 30 requests | 1 minute | POI deletion |
| `ActionUpdateProfile` | 5 requests | 1 minute | Profile updates |
| Default | 100 requests | 1 hour | Other actions |

## Future Migration Plan 🚀

### Phase 1: Production Redis Migration (Recommended)
**Goal**: Replace SimpleRateLimiter with Redis-based RateLimiter for production

**Benefits:**
- **Persistence**: Rate limits survive server restarts
- **Scalability**: Works across multiple server instances
- **Advanced Features**: Custom limits per user, detailed statistics
- **Better Performance**: Optimized sliding window with Redis sorted sets

**Implementation:**
```go
// In server.go New() function
var rateLimiter services.RateLimiterInterface
if cfg.GinMode != "test" && redisClient != nil {
    // Use Redis-based rate limiter in production
    config := services.GetDefaultRateLimiterConfig()
    rateLimiter = services.NewRateLimiter(redisClient, config)
    log.Println("✅ Using Redis-based rate limiter")
} else {
    // Use simple rate limiter for tests/development
    rateLimiter = &SimpleRateLimiter{}
    log.Println("⚠️ Using simple in-memory rate limiter")
}
```

### Phase 2: Enhanced Features (Optional)
- **User-specific limits**: VIP users get higher limits
- **Dynamic limits**: Adjust limits based on server load
- **Rate limit analytics**: Track usage patterns
- **Graceful degradation**: Fallback to SimpleRateLimiter if Redis fails

## Implementation Notes

### Redis-based RateLimiter Features
- ✅ **Sliding Window**: Proper time-based rate limiting
- ✅ **Atomic Operations**: Redis pipelines for consistency
- ✅ **Configurable**: Different limits per action type
- ✅ **Statistics**: User rate limit stats and remaining requests
- ✅ **Custom Limits**: Per-user custom rate limits
- ✅ **Proper Reset Times**: Based on oldest entry expiration

### SimpleRateLimiter Features
- ✅ **In-Memory**: Fast, no external dependencies
- ✅ **Thread-Safe**: Mutex-protected operations
- ✅ **Action-Specific**: Different limits per action type
- ✅ **Proper Reset Times**: Fixed sliding window calculation
- ❌ **No Persistence**: Resets on server restart
- ❌ **Single Instance**: Doesn't work with multiple servers

## Migration Checklist

When ready to migrate to Redis-based rate limiter:

- [ ] Update server initialization to conditionally use Redis rate limiter
- [ ] Test Redis rate limiter in staging environment
- [ ] Monitor performance impact
- [ ] Verify rate limits work correctly across server restarts
- [ ] Update documentation
- [ ] Remove SimpleRateLimiter (optional, can keep as fallback)

## Testing Strategy

- **Unit Tests**: Mock-based testing for handlers
- **Integration Tests**: Test with real Redis instance
- **Load Tests**: Verify rate limiting under high load
- **Failover Tests**: Test fallback behavior when Redis is unavailable