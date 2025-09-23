# WebSocket Integration Testing Infrastructure

This document describes the WebSocket integration testing infrastructure that provides comprehensive testing capabilities for real-time WebSocket functionality.

## Overview

The WebSocket integration testing infrastructure provides:

- **Real HTTP Server**: Actual HTTP server with WebSocket upgrade handling
- **Multi-Client Support**: Multiple concurrent WebSocket client connections
- **Message Broadcasting**: Real message broadcasting and delivery testing
- **Connection Lifecycle**: Complete connection setup, management, and cleanup
- **Map Isolation**: Testing message isolation between different maps
- **Concurrent Testing**: Support for concurrent connection and message testing
- **Fluent API**: Easy-to-use assertion and expectation methods

## Quick Start

### Basic WebSocket Testing

```go
func TestMyWebSocketFeature(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    // Setup WebSocket test environment
    testWS := testdata.SetupWebSocket(t)
    
    // Create a client connection
    client := testWS.CreateClient("session-1", "user-1", "map-1")
    require.NotNil(t, client)
    
    // Wait for connection
    time.Sleep(50 * time.Millisecond)
    
    // Verify connection
    assert.True(t, client.IsConnected())
    testWS.AssertClientConnected("session-1")
}
```

### Multi-Client Broadcasting

```go
func TestWebSocketBroadcasting(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Create multiple clients
    client1 := testWS.CreateClient("session-1", "user-1", "map-1")
    client2 := testWS.CreateClient("session-2", "user-2", "map-1")
    
    // Wait for connections
    time.Sleep(100 * time.Millisecond)
    
    // Broadcast message
    message := websocket.Message{
        Type: "poi_created",
        Data: map[string]interface{}{
            "poiId": "poi-123",
            "name":  "Coffee Shop",
        },
        Timestamp: time.Now(),
    }
    
    testWS.BroadcastToMap("map-1", message)
    
    // Verify both clients receive the message
    msg1 := client1.ExpectMessage("poi_created", 100*time.Millisecond)
    msg2 := client2.ExpectMessage("poi_created", 100*time.Millisecond)
    
    assert.Equal(t, "poi_created", msg1.Type)
    assert.Equal(t, "poi_created", msg2.Type)
}
```

## Core Components

### TestWebSocket Structure

The `TestWebSocket` struct provides the main interface for WebSocket integration testing:

```go
type TestWebSocket struct {
    t        TestingT
    server   *httptest.Server
    handler  *websocket.Handler
    clients  map[string]*TestWSClient
    mutex    sync.RWMutex
    upgrader ws.Upgrader
}
```

### TestWSClient Structure

The `TestWSClient` represents a WebSocket client connection for testing:

```go
type TestWSClient struct {
    SessionID   string
    UserID      string
    MapID       string
    Conn        *ws.Conn
    Messages    chan websocket.Message
    Errors      chan error
    Connected   bool
    mutex       sync.RWMutex
    stopReading chan struct{}
}
```

## Key Methods

### Setup and Management

- `SetupWebSocket(t TestingT) *TestWebSocket` - Creates WebSocket test environment
- `CreateClient(sessionID, userID, mapID string) *TestWSClient` - Creates test client
- `GetClient(sessionID string) *TestWSClient` - Retrieves existing client
- `Cleanup()` - Cleans up all connections and resources

### Connection Management

- `GetConnectedClients() int` - Returns total connected client count
- `GetMapClients(mapID string) int` - Returns clients connected to specific map
- `BroadcastToMap(mapID string, message websocket.Message)` - Broadcasts to map

### Assertions

- `AssertClientConnected(sessionID string)` - Asserts client is connected
- `AssertClientDisconnected(sessionID string)` - Asserts client is disconnected
- `AssertConnectedClientsCount(expectedCount int)` - Asserts total client count
- `AssertMapClientsCount(mapID string, expectedCount int)` - Asserts map client count

### Client Methods

- `SendMessage(message websocket.Message) error` - Sends message to server
- `ReceiveMessage(timeout time.Duration) (websocket.Message, error)` - Receives message
- `ReceiveMessages(count int, timeout time.Duration) ([]websocket.Message, error)` - Receives multiple messages
- `ExpectMessage(messageType string, timeout time.Duration) websocket.Message` - Expects specific message type
- `ExpectNoMessage(timeout time.Duration)` - Asserts no message is received
- `Close()` - Closes the WebSocket connection
- `IsConnected() bool` - Returns connection status

## Testing Patterns

### Basic Connection Testing

```go
func TestWebSocketConnection(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Test connection
    client := testWS.CreateClient("session-1", "user-1", "map-1")
    require.NotNil(t, client)
    
    time.Sleep(50 * time.Millisecond)
    
    // Verify connection properties
    assert.True(t, client.IsConnected())
    assert.Equal(t, "session-1", client.SessionID)
    assert.Equal(t, "user-1", client.UserID)
    assert.Equal(t, "map-1", client.MapID)
    
    // Test disconnection
    client.Close()
    time.Sleep(50 * time.Millisecond)
    assert.False(t, client.IsConnected())
}
```

### Message Broadcasting Testing

```go
func TestWebSocketBroadcasting(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Create clients on same map
    client1 := testWS.CreateClient("session-1", "user-1", "map-test")
    client2 := testWS.CreateClient("session-2", "user-2", "map-test")
    
    // Create client on different map
    client3 := testWS.CreateClient("session-3", "user-3", "map-other")
    
    time.Sleep(100 * time.Millisecond)
    
    // Broadcast to specific map
    message := websocket.Message{
        Type: "test_broadcast",
        Data: map[string]interface{}{
            "content": "Hello map-test!",
        },
        Timestamp: time.Now(),
    }
    
    testWS.BroadcastToMap("map-test", message)
    
    // Clients on map-test should receive message
    msg1 := client1.ExpectMessage("test_broadcast", 100*time.Millisecond)
    msg2 := client2.ExpectMessage("test_broadcast", 100*time.Millisecond)
    
    // Client on different map should not receive message
    client3.ExpectNoMessage(100 * time.Millisecond)
}
```

### Avatar Movement Testing

```go
func TestWebSocketAvatarMovement(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Create mover and observer
    mover := testWS.CreateClient("session-mover", "user-mover", "map-1")
    observer := testWS.CreateClient("session-observer", "user-observer", "map-1")
    
    time.Sleep(100 * time.Millisecond)
    
    // Send movement message
    movementMsg := websocket.Message{
        Type: "avatar_movement",
        Data: map[string]interface{}{
            "userId": "user-mover",
            "position": map[string]float64{
                "lat": 40.7128,
                "lng": -74.0060,
            },
        },
        Timestamp: time.Now(),
    }
    
    err := mover.SendMessage(movementMsg)
    require.NoError(t, err)
    
    // Simulate server broadcasting the movement
    testWS.BroadcastToMap("map-1", movementMsg)
    
    // Observer should receive movement
    receivedMsg := observer.ExpectMessage("avatar_movement", 200*time.Millisecond)
    
    data, ok := receivedMsg.Data.(map[string]interface{})
    require.True(t, ok)
    assert.Equal(t, "user-mover", data["userId"])
}
```

### POI Event Testing

```go
func TestWebSocketPOIEvents(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Create multiple clients
    clients := make([]*testdata.TestWSClient, 3)
    for i := 0; i < 3; i++ {
        sessionID := fmt.Sprintf("session-%d", i+1)
        userID := fmt.Sprintf("user-%d", i+1)
        clients[i] = testWS.CreateClient(sessionID, userID, "map-poi")
    }
    
    time.Sleep(100 * time.Millisecond)
    
    // Test POI creation event
    poiCreatedMsg := websocket.Message{
        Type: "poi_created",
        Data: map[string]interface{}{
            "poiId":      "poi-123",
            "name":       "Coffee Shop",
            "position":   map[string]float64{"lat": 40.7128, "lng": -74.0060},
            "createdBy":  "user-1",
        },
        Timestamp: time.Now(),
    }
    
    testWS.BroadcastToMap("map-poi", poiCreatedMsg)
    
    // All clients should receive the event
    for i, client := range clients {
        msg := client.ExpectMessage("poi_created", 200*time.Millisecond)
        
        data, ok := msg.Data.(map[string]interface{})
        require.True(t, ok, "Client %d should receive valid data", i+1)
        assert.Equal(t, "poi-123", data["poiId"])
        assert.Equal(t, "Coffee Shop", data["name"])
    }
}
```

### Concurrent Connection Testing

```go
func TestWebSocketConcurrentConnections(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    const numClients = 10
    clients := make([]*testdata.TestWSClient, numClients)
    done := make(chan bool, numClients)
    
    // Create clients concurrently
    for i := 0; i < numClients; i++ {
        go func(index int) {
            sessionID := fmt.Sprintf("session-%d", index)
            userID := fmt.Sprintf("user-%d", index)
            clients[index] = testWS.CreateClient(sessionID, userID, "map-concurrent")
            done <- true
        }(i)
    }
    
    // Wait for all connections
    for i := 0; i < numClients; i++ {
        <-done
    }
    
    time.Sleep(200 * time.Millisecond)
    
    // Verify connections
    testWS.AssertConnectedClientsCount(numClients)
    testWS.AssertMapClientsCount("map-concurrent", numClients)
    
    // Verify each client
    for i, client := range clients {
        require.NotNil(t, client, "Client %d should not be nil", i)
        assert.True(t, client.IsConnected(), "Client %d should be connected", i)
    }
}
```

### Connection Lifecycle Testing

```go
func TestWebSocketConnectionLifecycle(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Test initial connection
    client := testWS.CreateClient("session-lifecycle", "user-lifecycle", "map-lifecycle")
    require.NotNil(t, client)
    
    time.Sleep(50 * time.Millisecond)
    
    // Verify connection
    testWS.AssertClientConnected("session-lifecycle")
    testWS.AssertConnectedClientsCount(1)
    testWS.AssertMapClientsCount("map-lifecycle", 1)
    
    // Test message sending
    testMsg := websocket.Message{
        Type: "heartbeat",
        Data: map[string]interface{}{
            "timestamp": time.Now().Unix(),
        },
        Timestamp: time.Now(),
    }
    
    err := client.SendMessage(testMsg)
    require.NoError(t, err)
    
    // Test disconnection
    client.Close()
    time.Sleep(100 * time.Millisecond)
    
    // Verify disconnection
    assert.False(t, client.IsConnected())
    testWS.AssertClientDisconnected("session-lifecycle")
    testWS.AssertConnectedClientsCount(0)
    testWS.AssertMapClientsCount("map-lifecycle", 0)
}
```

### Map Isolation Testing

```go
func TestWebSocketMapIsolation(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping WebSocket integration test in short mode")
    }
    
    testWS := testdata.SetupWebSocket(t)
    
    // Create clients on different maps
    client1 := testWS.CreateClient("session-1", "user-1", "map-1")
    client2 := testWS.CreateClient("session-2", "user-2", "map-1")
    client3 := testWS.CreateClient("session-3", "user-3", "map-2")
    
    time.Sleep(100 * time.Millisecond)
    
    // Verify map distribution
    testWS.AssertMapClientsCount("map-1", 2)
    testWS.AssertMapClientsCount("map-2", 1)
    
    // Broadcast to map-1 only
    message := websocket.Message{
        Type: "map_specific_event",
        Data: map[string]interface{}{
            "mapId": "map-1",
            "event": "something happened",
        },
        Timestamp: time.Now(),
    }
    
    testWS.BroadcastToMap("map-1", message)
    
    // Clients on map-1 should receive message
    msg1 := client1.ExpectMessage("map_specific_event", 100*time.Millisecond)
    msg2 := client2.ExpectMessage("map_specific_event", 100*time.Millisecond)
    
    // Client on map-2 should not receive message
    client3.ExpectNoMessage(100 * time.Millisecond)
}
```

## Best Practices

### Test Structure

1. **Always check for short mode**: Skip WebSocket integration tests in short mode
2. **Use descriptive test names**: Include "Integration" in test names for clarity
3. **Wait for connections**: Allow time for WebSocket connections to establish
4. **Clean resource management**: Cleanup is automatic, but close clients explicitly when testing disconnection

### Message Testing

1. **Use ExpectMessage for known messages**: Prefer `ExpectMessage()` over manual `ReceiveMessage()`
2. **Test message isolation**: Verify messages are only received by intended clients
3. **Verify message content**: Check both message type and data payload
4. **Handle timing**: Use appropriate timeouts for message delivery

### Connection Testing

1. **Test connection properties**: Verify SessionID, UserID, and MapID are correct
2. **Test lifecycle events**: Cover connection, messaging, and disconnection
3. **Test concurrent scenarios**: Verify behavior under concurrent load
4. **Test error conditions**: Include connection failures and recovery

### Performance Testing

1. **Benchmark critical paths**: Test connection establishment and message broadcasting
2. **Test with realistic loads**: Use appropriate numbers of concurrent clients
3. **Monitor resource usage**: Ensure tests don't leak connections or goroutines

## Integration with CI/CD

The WebSocket integration tests are designed to work in CI/CD environments:

- Tests are skipped in short mode (`go test -short`)
- No external dependencies required (uses in-memory test server)
- Automatic cleanup prevents resource leaks
- Concurrent tests are designed to be stable

Example CI configuration:

```yaml
test:
  script:
    - go test ./... -short  # Skips WebSocket integration tests
    - go test ./internal/integration -run WebSocket  # Run WebSocket tests specifically
```

## Troubleshooting

### Connection Issues

If WebSocket connections fail:

1. Check that the test server is properly started
2. Verify WebSocket upgrade is working correctly
3. Ensure proper timing with `time.Sleep()` calls
4. Check for port conflicts in concurrent tests

### Message Delivery Issues

If messages aren't being received:

1. Verify clients are connected to the correct map
2. Check message broadcasting is working
3. Ensure proper timing between send and receive
4. Verify message channels aren't full or blocked

### Timing Issues

For timing-related test failures:

1. Increase timeout values for slower environments
2. Add appropriate delays for connection establishment
3. Use `time.Sleep()` judiciously to allow async operations
4. Consider using retry logic for flaky operations

### Resource Leaks

To prevent resource leaks:

1. Always use `testdata.SetupWebSocket(t)` for automatic cleanup
2. Close clients explicitly when testing disconnection
3. Avoid creating too many concurrent connections in tests
4. Monitor goroutine counts in benchmarks

## File Organization

```
backend/internal/
├── testdata/
│   ├── testws.go              # WebSocket testing infrastructure
│   ├── testws_test.go         # Infrastructure tests
│   └── WEBSOCKET_INTEGRATION.md  # This documentation
├── integration/
│   └── websocket_test.go      # Integration tests using WebSocket infrastructure
└── websocket/
    ├── *.go                   # WebSocket implementation
    └── *_test.go              # Unit tests
```

This WebSocket integration testing infrastructure provides comprehensive testing capabilities for real-time WebSocket functionality, enabling reliable testing of connection management, message broadcasting, and multi-client scenarios.