# Infrastructure Flow Integration Tests

## Overview

This document describes the comprehensive infrastructure flow integration tests that validate the complete end-to-end functionality across all infrastructure layers: Database, Redis, and WebSocket.

## Test Architecture

### Core Integration Test: `infrastructure_flow_test.go`

The main infrastructure flow integration test demonstrates complete system integration across all three infrastructure layers:

#### Test Structure

1. **TestInfrastructureFlow_DatabaseRedisWebSocketIntegration**
   - Tests complete integration of Database + Redis + WebSocket
   - Validates data persistence, caching, and real-time communication
   - Demonstrates end-to-end user scenarios

2. **TestInfrastructureFlow_ErrorHandling**
   - Tests error handling and recovery across all infrastructure layers
   - Validates graceful degradation when components fail
   - Ensures system resilience

3. **TestInfrastructureFlow_Performance**
   - Tests performance under concurrent load
   - Validates system behavior with multiple simultaneous operations
   - Measures throughput and response times

### Test Scenarios Covered

#### 1. Database + Redis Integration
```go
// Create session in database
session := &models.Session{...}
err := testDB.DB.Create(session).Error

// Set presence in Redis
err = testRedis.Client().Set(ctx, "session:"+session.ID, "active", 30*time.Minute).Err()

// Create POI in database
poi := &models.POI{...}
err = testDB.DB.Create(poi).Error

// Add participant to POI in Redis
err = testRedis.Client().SAdd(ctx, "poi:participants:"+poi.ID, session.ID).Err()
```

#### 2. WebSocket Integration
```go
// Create WebSocket client
client := testWS.CreateClient("session-1", "user-1", "map-1")

// Send test message
testMsg := websocket.Message{
    Type: "test_message",
    Data: testData,
}
err := client.SendMessage(testMsg)
```

#### 3. Complete Flow Integration
```go
// 1. Create user session (Database + Redis)
userSession := &models.Session{...}
err := testDB.DB.Create(userSession).Error
err = testRedis.Client().Set(ctx, "session:"+userSession.ID, "active", 30*time.Minute).Err()

// 2. Create WebSocket connection
wsClient := testWS.CreateClient(userSession.ID, userSession.UserID, userSession.MapID)

// 3. Create POI (Database)
poi := &models.POI{...}
err = testDB.DB.Create(poi).Error

// 4. User joins POI (Redis)
err = testRedis.Client().SAdd(ctx, "poi:participants:"+poi.ID, userSession.ID).Err()

// 5. Update avatar position (Database + Redis + WebSocket)
// Database update
err = testDB.DB.Model(userSession).Update("avatar_pos_lat", newPosition.Lat).Error

// Redis presence update
err = testRedis.Client().HSet(ctx, "presence:"+userSession.ID, 
    "lat", newPosition.Lat, 
    "lng", newPosition.Lng).Err()

// WebSocket broadcast
movementMsg := websocket.Message{
    Type: "avatar_movement",
    Data: movementData,
}
err = wsClient.SendMessage(movementMsg)
```

#### 4. Multi-User Scenarios
- Tests concurrent user sessions
- Validates shared POI interactions
- Ensures proper isolation between maps
- Tests real-time synchronization

#### 5. Error Handling
- Database constraint violations
- Redis connection failures
- WebSocket disconnections
- Graceful degradation testing

#### 6. Performance Testing
- Concurrent database operations
- Concurrent Redis operations
- Concurrent WebSocket connections
- Load testing with multiple users

## Infrastructure Components Tested

### 1. Database Integration (`testdata.Setup`)
- **Purpose**: Validates PostgreSQL database operations
- **Features**:
  - Session management (create, update, query)
  - POI management (create, query, relationships)
  - Transaction handling
  - Data integrity constraints
  - Concurrent access patterns

### 2. Redis Integration (`testdata.SetupRedis`)
- **Purpose**: Validates Redis caching and real-time data
- **Features**:
  - Session presence tracking
  - POI participant management
  - Real-time data updates
  - Key expiration handling
  - Set operations for participants

### 3. WebSocket Integration (`testdata.SetupWebSocket`)
- **Purpose**: Validates real-time communication
- **Features**:
  - Connection lifecycle management
  - Message broadcasting
  - Client isolation by map
  - Concurrent connection handling
  - Message ordering and delivery

## Test Execution

### Prerequisites
- PostgreSQL running on localhost:5432
- Redis running on localhost:6380
- TEST_INTEGRATION environment variable set

### Running Tests
```bash
# Run all infrastructure flow tests
TEST_INTEGRATION=1 go test ./internal/integration -run TestInfrastructureFlow -v

# Run specific integration test
TEST_INTEGRATION=1 go test ./internal/integration -run TestInfrastructureFlow_DatabaseRedisWebSocketIntegration -v

# Run with short mode (skips integration tests)
go test ./internal/integration -run TestInfrastructureFlow -short -v
```

### Expected Behavior
- **With Infrastructure**: Tests run and validate complete system integration
- **Without Infrastructure**: Tests fail gracefully with connection errors
- **Short Mode**: Tests are skipped to avoid infrastructure dependencies

## Integration Test Benefits

### 1. End-to-End Validation
- Validates complete user journeys across all systems
- Ensures data consistency between Database and Redis
- Confirms real-time updates via WebSocket

### 2. System Resilience
- Tests error propagation and handling
- Validates graceful degradation
- Ensures system stability under load

### 3. Performance Validation
- Measures system performance under realistic conditions
- Identifies bottlenecks in multi-layer operations
- Validates concurrent access patterns

### 4. Integration Confidence
- Provides confidence that all infrastructure components work together
- Validates deployment readiness
- Ensures production-like behavior

## Test Infrastructure Features

### Isolation
- Each test gets isolated database and Redis instances
- WebSocket connections are properly managed and cleaned up
- No test interference or data pollution

### Realistic Scenarios
- Tests mirror actual application usage patterns
- Validates real user workflows
- Tests edge cases and error conditions

### Comprehensive Coverage
- Database operations (CRUD, transactions, constraints)
- Redis operations (caching, sets, expiration)
- WebSocket operations (connections, messaging, broadcasting)
- Cross-layer data consistency
- Error handling and recovery
- Performance under load

## Future Enhancements

### 1. Service Layer Integration
- Add service layer integration tests
- Test business logic with infrastructure
- Validate rate limiting and business rules

### 2. Handler Layer Integration
- Add HTTP handler integration tests
- Test complete HTTP → Service → Infrastructure flows
- Validate API contracts with infrastructure

### 3. Advanced Scenarios
- Test complex multi-user interactions
- Add geographic boundary testing
- Test real-time event ordering
- Add chaos engineering scenarios

## Conclusion

The infrastructure flow integration tests provide comprehensive validation of the complete system architecture. They ensure that Database, Redis, and WebSocket components work together seamlessly to support the application's real-time, multi-user functionality.

These tests serve as:
- **Quality Gates**: Ensuring system integration before deployment
- **Documentation**: Demonstrating how components interact
- **Regression Prevention**: Catching integration issues early
- **Performance Baseline**: Establishing performance expectations

The test infrastructure is designed to be:
- **Reliable**: Consistent results across environments
- **Maintainable**: Easy to update and extend
- **Comprehensive**: Covering all critical integration paths
- **Realistic**: Mirroring production usage patterns