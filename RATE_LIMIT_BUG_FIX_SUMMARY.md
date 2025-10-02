# Rate Limiting Bug Fix Summary

## Problem
Avatar movement rate limiting was broken in production, causing persistent rate limit errors even after waiting the expected reset time. Users would see:
- `retryAfter: 0.0000036` (essentially 0 seconds)
- Rate limits persisting indefinitely, even after waiting 60+ seconds

## Root Cause
The `GetWindowResetTime` method in both rate limiter implementations was incorrectly calculating when the sliding window would reset:

### 1. Redis-based RateLimiter (services/rate_limiter.go)
**Broken Logic:**
```go
func (rl *RateLimiter) GetWindowResetTime(...) (time.Time, error) {
    // WRONG: Always returns now + window duration
    resetTime := now.Add(limit.Window)
    return resetTime, nil
}
```

**Fixed Logic:**
```go
func (rl *RateLimiter) GetWindowResetTime(...) (time.Time, error) {
    // Get the oldest entry in the sliding window
    results, err := rl.redis.ZRangeWithScores(ctx, key, 0, 0)
    if len(results) == 0 {
        return time.Now(), nil // Empty window resets immediately
    }
    
    // Calculate when oldest entry expires: oldestTime + window
    oldestTime := time.Unix(0, int64(scoreFloat))
    resetTime := oldestTime.Add(limit.Window)
    return resetTime, nil
}
```

### 2. SimpleRateLimiter (server/server.go) - Used in Production
**Broken Logic:**
```go
func (r *SimpleRateLimiter) GetWindowResetTime(...) (time.Time, error) {
    return time.Now().Add(1 * time.Hour), nil // WRONG: Always 1 hour from now
}
```

**Fixed Logic:**
```go
func (r *SimpleRateLimiter) GetWindowResetTime(...) (time.Time, error) {
    // Find oldest request and calculate when it expires
    oldestRequest := requests[0]
    for _, reqTime := range requests {
        if reqTime.Before(oldestRequest) {
            oldestRequest = reqTime
        }
    }
    resetTime := oldestRequest.Add(window)
    return resetTime, nil
}
```

## Additional Fixes

### Improved Rate Limits for Avatar Movement
Updated SimpleRateLimiter to use appropriate limits per action:
- **Avatar Movement**: 60 requests per minute (1 per second) - was 100 per hour
- **Session Creation**: 10 requests per minute
- **POI Operations**: 30 requests per minute  
- **Profile Updates**: 5 requests per minute

### Added Missing Redis Interface Method
Added `ZRangeWithScores` method to `RedisClientInterface` and mock implementations.

## Files Modified
1. `backend/internal/services/rate_limiter.go` - Fixed Redis-based rate limiter
2. `backend/internal/server/server.go` - Fixed SimpleRateLimiter and improved rate limits
3. `backend/internal/services/rate_limiter_test.go` - Updated existing test
4. `backend/internal/services/rate_limiter_window_reset_test.go` - Added comprehensive tests

## Testing
- ✅ All existing tests pass
- ✅ New tests verify correct sliding window behavior
- ✅ Server compiles successfully
- ✅ Rate limits now reset properly based on oldest entry expiration

## Impact
- **Avatar movement** rate limits now work correctly
- **Proper retryAfter** times are returned (not microseconds)
- **Sliding window** algorithm works as intended
- **Production ready** with appropriate rate limits per action type

The fix ensures that rate limit windows reset when the oldest request in the window expires, not from the current time, which is the correct behavior for sliding window rate limiting.