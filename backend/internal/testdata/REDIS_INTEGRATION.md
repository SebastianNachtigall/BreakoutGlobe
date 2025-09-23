# Redis Integration Testing Infrastructure

This document describes the Redis integration testing infrastructure that provides isolated, reliable Redis testing capabilities for the application.

## Overview

The Redis integration testing infrastructure provides:

- **Isolated Redis Databases**: Each test gets its own Redis database number to ensure complete isolation
- **Automatic Cleanup**: Test data is automatically cleaned up after each test
- **Fluent Assertion API**: Easy-to-use assertion methods for Redis operations
- **Pub/Sub Testing**: Built-in support for testing Redis pub/sub functionality
- **Environment Configuration**: Configurable Redis connection settings via environment variables

## Quick Start

### Basic Usage

```go
func TestMyRedisFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }
    
    // Setup isolated Redis instance
    testRedis := testdata.SetupRedis(t)
    
    // Use Redis client directly
    client := testRedis.Client()
    err := client.Set(context.Background(), "key", "value", 0).Err()
    require.NoError(t, err)
    
    // Or use assertion helpers
    testRedis.AssertKeyExists("key")
}
```

### Testing Redis Components

```go
func TestPOIParticipants_Integration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }
    
    // Setup test Redis
    testRedis := testdata.SetupRedis(t)
    participants := redis.NewPOIParticipants(testRedis.Client())
    ctx := context.Background()
    
    // Test functionality
    err := participants.JoinPOI(ctx, "poi-123", "session-456")
    require.NoError(t, err)
    
    // Verify Redis state
    testRedis.AssertSetContains("poi:participants:poi-123", "session-456")
}
```

## Core Components

### TestRedis Structure

The `TestRedis` struct provides the main interface for Redis integration testing:

```go
type TestRedis struct {
    t      TestingT
    client *redis.Client
    dbNum  int
}
```

### Key Methods

#### Setup and Configuration

- `SetupRedis(t TestingT) *TestRedis` - Creates an isolated Redis instance for testing
- `Client() *redis.Client` - Returns the Redis client for direct operations
- `DBNum() int` - Returns the database number being used (for debugging)

#### Assertion Helpers

- `AssertKeyExists(key string)` - Asserts that a key exists
- `AssertKeyNotExists(key string)` - Asserts that a key does not exist
- `AssertSetContains(key, member)` - Asserts that a set contains a member
- `AssertSetNotContains(key, member)` - Asserts that a set does not contain a member
- `AssertSetSize(key, expectedSize)` - Asserts that a set has a specific size
- `AssertHashField(key, field, expectedValue)` - Asserts that a hash field has a specific value

#### Pub/Sub Support

- `Subscribe(channels ...string) *redis.PubSub` - Creates a subscription to channels

## Environment Configuration

The Redis integration infrastructure can be configured via environment variables:

- `TEST_REDIS_HOST` - Redis host (default: "localhost")
- `TEST_REDIS_PORT` - Redis port (default: "6380")
- `TEST_REDIS_PASSWORD` - Redis password (default: "")

## Database Isolation

Each test gets its own Redis database number (0-15) to ensure complete isolation:

- Database numbers are generated using timestamp-based selection
- Each test automatically flushes its database before and after execution
- Collisions are possible but rare with the timestamp-based approach

## Integration Test Patterns

### POI Participants Testing

```go
func TestRedisIntegration_POIParticipants_AddRemove(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }

    testRedis := testdata.SetupRedis(t)
    participants := redis.NewPOIParticipants(testRedis.Client())
    ctx := context.Background()

    // Test adding participant
    err := participants.JoinPOI(ctx, "poi-123", "session-456")
    require.NoError(t, err)

    // Verify Redis state
    testRedis.AssertKeyExists("poi:participants:poi-123")
    testRedis.AssertSetContains("poi:participants:poi-123", "session-456")

    // Test removing participant
    err = participants.LeavePOI(ctx, "poi-123", "session-456")
    require.NoError(t, err)

    // Verify cleanup
    testRedis.AssertSetNotContains("poi:participants:poi-123", "session-456")
}
```

### Session Presence Testing

```go
func TestRedisIntegration_SessionPresence_SetGet(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }

    testRedis := testdata.SetupRedis(t)
    presence := redis.NewSessionPresence(testRedis.Client())
    ctx := context.Background()

    // Test setting presence
    data := &redis.SessionPresenceData{
        UserID: "user-123",
        MapID:  "map-456",
        AvatarPosition: models.LatLng{Lat: 40.7128, Lng: -74.0060},
        LastActive:     time.Now(),
    }

    err := presence.SetSessionPresence(ctx, "session-123", data, 5*time.Minute)
    require.NoError(t, err)

    // Verify presence was set
    testRedis.AssertKeyExists("session:session-123")
}
```

### Pub/Sub Testing

```go
func TestRedisIntegration_PubSub_PublishSubscribe(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }

    testRedis := testdata.SetupRedis(t)
    pubsub := redis.NewPubSub(testRedis.Client())
    ctx := context.Background()

    // Subscribe to channel
    subscription := testRedis.Subscribe("map:123")
    defer subscription.Close()

    // Wait for subscription
    time.Sleep(10 * time.Millisecond)

    // Publish event
    event := redis.POICreatedEvent{
        POIID:     "poi-123",
        MapID:     "123",
        Name:      "Test POI",
        Position:  redis.LatLng{Lat: 40.7128, Lng: -74.0060},
        Timestamp: time.Now(),
    }

    err := pubsub.PublishPOICreated(ctx, event)
    require.NoError(t, err)

    // Receive and verify message
    msg, err := subscription.ReceiveTimeout(ctx, 100*time.Millisecond)
    require.NoError(t, err)
    // ... verify message content
}
```

## Concurrent Testing

The infrastructure supports concurrent access testing:

```go
func TestRedisIntegration_ConcurrentAccess(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping Redis integration test in short mode")
    }

    testRedis := testdata.SetupRedis(t)
    participants := redis.NewPOIParticipants(testRedis.Client())
    ctx := context.Background()

    const numSessions = 10
    done := make(chan error, numSessions)

    // Concurrently add participants
    for i := 0; i < numSessions; i++ {
        go func(index int) {
            sessionID := fmt.Sprintf("session-%d", index)
            done <- participants.JoinPOI(ctx, "poi-123", sessionID)
        }(i)
    }

    // Wait for completion
    for i := 0; i < numSessions; i++ {
        err := <-done
        assert.NoError(t, err)
    }

    // Verify final state
    testRedis.AssertSetSize("poi:participants:poi-123", numSessions)
}
```

## Best Practices

### Test Structure

1. **Always check for short mode**: Skip Redis integration tests in short mode
2. **Use descriptive test names**: Include "Integration" in test names
3. **Clean setup**: Use `testdata.SetupRedis(t)` for each test
4. **Proper cleanup**: Cleanup is automatic, but defer subscription closes

### Assertions

1. **Use fluent assertions**: Prefer `testRedis.AssertKeyExists()` over manual checks
2. **Test Redis state**: Verify that Redis contains expected data structures
3. **Test isolation**: Ensure tests don't interfere with each other

### Error Handling

1. **Check all errors**: Use `require.NoError(t, err)` for critical operations
2. **Meaningful messages**: Provide context in assertion failures
3. **Timeout handling**: Use appropriate timeouts for pub/sub operations

## File Organization

```
backend/internal/
├── testdata/
│   ├── testredis.go           # Redis testing infrastructure
│   ├── testredis_test.go      # Infrastructure tests
│   └── REDIS_INTEGRATION.md   # This documentation
├── integration/
│   └── redis_test.go          # Integration tests using Redis infrastructure
└── redis/
    ├── *.go                   # Redis components
    └── *_test.go              # Unit tests (skip in short mode)
```

## Troubleshooting

### Connection Issues

If tests fail with connection errors:

1. Check that Redis is running on the configured host/port
2. Verify environment variables are set correctly
3. Ensure Redis allows multiple database connections

### Database Conflicts

If tests interfere with each other:

1. Verify each test uses `testdata.SetupRedis(t)`
2. Check that cleanup is working properly
3. Consider using different Redis instances for parallel test runs

### Performance Issues

For slow tests:

1. Use appropriate timeouts for pub/sub operations
2. Consider reducing test data size for benchmarks
3. Ensure Redis is running locally for development

## Integration with CI/CD

The Redis integration tests are designed to work in CI/CD environments:

- Tests are skipped in short mode (`go test -short`)
- Environment variables can configure Redis connection
- Automatic cleanup prevents test pollution
- Isolated databases prevent conflicts

Example CI configuration:

```yaml
env:
  TEST_REDIS_HOST: redis
  TEST_REDIS_PORT: 6379
  
services:
  redis:
    image: redis:7-alpine
    ports:
      - 6379:6379
```

This infrastructure provides a solid foundation for testing Redis-based functionality with proper isolation, cleanup, and ease of use.