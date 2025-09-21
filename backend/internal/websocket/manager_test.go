package websocket

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ManagerTestSuite contains the test suite for WebSocket manager
type ManagerTestSuite struct {
	suite.Suite
	manager *Manager
}

func (suite *ManagerTestSuite) SetupTest() {
	suite.manager = NewManager()
}

func (suite *ManagerTestSuite) TearDownTest() {
	suite.manager.Shutdown()
}

func (suite *ManagerTestSuite) TestClientRegistration() {
	// Create mock client
	client := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	// Register client
	suite.manager.RegisterClient(client)
	
	// Wait for registration to complete
	time.Sleep(10 * time.Millisecond)
	
	// Verify client is registered
	suite.True(suite.manager.IsClientConnected("session-1"))
	suite.Equal(1, suite.manager.GetConnectedClients())
	suite.Equal(1, suite.manager.GetMapClients("map-1"))
}

func (suite *ManagerTestSuite) TestClientUnregistration() {
	// Create and register client
	client := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	suite.manager.RegisterClient(client)
	time.Sleep(10 * time.Millisecond)
	
	// Verify client is registered
	suite.True(suite.manager.IsClientConnected("session-1"))
	
	// Unregister client
	suite.manager.UnregisterClient(client)
	time.Sleep(10 * time.Millisecond)
	
	// Verify client is unregistered
	suite.False(suite.manager.IsClientConnected("session-1"))
	suite.Equal(0, suite.manager.GetConnectedClients())
	suite.Equal(0, suite.manager.GetMapClients("map-1"))
}

func (suite *ManagerTestSuite) TestMultipleClientsInSameMap() {
	// Create clients for the same map
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	// Register both clients
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)
	
	// Verify both clients are registered
	suite.True(suite.manager.IsClientConnected("session-1"))
	suite.True(suite.manager.IsClientConnected("session-2"))
	suite.Equal(2, suite.manager.GetConnectedClients())
	suite.Equal(2, suite.manager.GetMapClients("map-1"))
	
	// Verify map client sessions
	sessions := suite.manager.GetMapClientSessions("map-1")
	suite.Len(sessions, 2)
	suite.Contains(sessions, "session-1")
	suite.Contains(sessions, "session-2")
}

func (suite *ManagerTestSuite) TestMultipleClientsInDifferentMaps() {
	// Create clients for different maps
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-2",
		Send:      make(chan Message, 256),
	}
	
	// Register both clients
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)
	
	// Verify both clients are registered
	suite.Equal(2, suite.manager.GetConnectedClients())
	suite.Equal(1, suite.manager.GetMapClients("map-1"))
	suite.Equal(1, suite.manager.GetMapClients("map-2"))
	
	// Verify client maps
	maps := suite.manager.GetClientMaps()
	suite.Len(maps, 2)
	suite.Contains(maps, "map-1")
	suite.Contains(maps, "map-2")
}

func (suite *ManagerTestSuite) TestBroadcastToMap() {
	// Create clients for the same map
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	// Create client for different map
	client3 := &Client{
		SessionID: "session-3",
		UserID:    "user-3",
		MapID:     "map-2",
		Send:      make(chan Message, 256),
	}
	
	// Register all clients
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	suite.manager.RegisterClient(client3)
	time.Sleep(10 * time.Millisecond)
	
	// Broadcast message to map-1
	message := Message{
		Type: "test_message",
		Data: map[string]interface{}{
			"content": "Hello map-1!",
		},
		Timestamp: time.Now(),
	}
	
	err := suite.manager.BroadcastToMap("map-1", message)
	suite.NoError(err)
	
	// Verify clients in map-1 received the message
	select {
	case receivedMsg := <-client1.Send:
		suite.Equal("test_message", receivedMsg.Type)
		suite.Equal("Hello map-1!", receivedMsg.Data.(map[string]interface{})["content"])
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Client 1 did not receive message")
	}
	
	select {
	case receivedMsg := <-client2.Send:
		suite.Equal("test_message", receivedMsg.Type)
		suite.Equal("Hello map-1!", receivedMsg.Data.(map[string]interface{})["content"])
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Client 2 did not receive message")
	}
	
	// Verify client in map-2 did not receive the message
	select {
	case <-client3.Send:
		suite.Fail("Client 3 should not have received message")
	case <-time.After(50 * time.Millisecond):
		// Expected - client 3 should not receive the message
	}
}

func (suite *ManagerTestSuite) TestBroadcastToMapExcept() {
	// Create clients for the same map
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	// Register both clients
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)
	
	// Broadcast message to map-1 except session-1
	message := Message{
		Type: "test_message",
		Data: map[string]interface{}{
			"content": "Hello others!",
		},
		Timestamp: time.Now(),
	}
	
	err := suite.manager.BroadcastToMapExcept("map-1", "session-1", message)
	suite.NoError(err)
	
	// Verify client1 (excluded) did not receive the message
	select {
	case <-client1.Send:
		suite.Fail("Client 1 should not have received message")
	case <-time.After(50 * time.Millisecond):
		// Expected - client 1 should not receive the message
	}
	
	// Verify client2 received the message
	select {
	case receivedMsg := <-client2.Send:
		suite.Equal("test_message", receivedMsg.Type)
		suite.Equal("Hello others!", receivedMsg.Data.(map[string]interface{})["content"])
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Client 2 did not receive message")
	}
}

func (suite *ManagerTestSuite) TestBroadcastToAll() {
	// Create clients for different maps
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-2",
		Send:      make(chan Message, 256),
	}
	
	// Register both clients
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)
	
	// Broadcast message to all
	message := Message{
		Type: "global_message",
		Data: map[string]interface{}{
			"content": "Hello everyone!",
		},
		Timestamp: time.Now(),
	}
	
	err := suite.manager.BroadcastToAll(message)
	suite.NoError(err)
	
	// Verify both clients received the message
	select {
	case receivedMsg := <-client1.Send:
		suite.Equal("global_message", receivedMsg.Type)
		suite.Equal("Hello everyone!", receivedMsg.Data.(map[string]interface{})["content"])
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Client 1 did not receive message")
	}
	
	select {
	case receivedMsg := <-client2.Send:
		suite.Equal("global_message", receivedMsg.Type)
		suite.Equal("Hello everyone!", receivedMsg.Data.(map[string]interface{})["content"])
	case <-time.After(100 * time.Millisecond):
		suite.Fail("Client 2 did not receive message")
	}
}

func (suite *ManagerTestSuite) TestBroadcastToNonExistentMap() {
	// Try to broadcast to a map with no clients
	message := Message{
		Type: "test_message",
		Data: map[string]interface{}{
			"content": "Hello nobody!",
		},
		Timestamp: time.Now(),
	}
	
	err := suite.manager.BroadcastToMap("non-existent-map", message)
	suite.NoError(err) // Should not error, just no recipients
}

func (suite *ManagerTestSuite) TestGetMapClientSessions() {
	// Create clients for the same map
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
	}
	
	// Register both clients
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)
	
	// Get sessions for map-1
	sessions := suite.manager.GetMapClientSessions("map-1")
	suite.Len(sessions, 2)
	suite.Contains(sessions, "session-1")
	suite.Contains(sessions, "session-2")
	
	// Get sessions for non-existent map
	emptySessions := suite.manager.GetMapClientSessions("non-existent-map")
	suite.Len(emptySessions, 0)
}

func (suite *ManagerTestSuite) TestShutdown() {
	// Create and register clients (without real websocket connections)
	client1 := &Client{
		SessionID: "session-1",
		UserID:    "user-1",
		MapID:     "map-1",
		Send:      make(chan Message, 256),
		Conn:      nil, // No real connection for this test
	}
	
	client2 := &Client{
		SessionID: "session-2",
		UserID:    "user-2",
		MapID:     "map-2",
		Send:      make(chan Message, 256),
		Conn:      nil, // No real connection for this test
	}
	
	suite.manager.RegisterClient(client1)
	suite.manager.RegisterClient(client2)
	time.Sleep(10 * time.Millisecond)
	
	// Verify clients are registered
	suite.Equal(2, suite.manager.GetConnectedClients())
	
	// Manually close channels before shutdown to avoid panic
	close(client1.Send)
	close(client2.Send)
	
	// Shutdown manager (this will try to close already closed channels, but that's ok)
	suite.manager.Shutdown()
	
	// Verify all clients are removed
	suite.Equal(0, suite.manager.GetConnectedClients())
	suite.False(suite.manager.IsClientConnected("session-1"))
	suite.False(suite.manager.IsClientConnected("session-2"))
}

func TestManagerTestSuite(t *testing.T) {
	suite.Run(t, new(ManagerTestSuite))
}

// Test individual functions
func TestNewManager(t *testing.T) {
	manager := NewManager()
	assert.NotNil(t, manager)
	assert.NotNil(t, manager.clients)
	assert.NotNil(t, manager.mapClients)
	assert.NotNil(t, manager.register)
	assert.NotNil(t, manager.unregister)
	assert.NotNil(t, manager.broadcast)
	
	// Cleanup
	manager.Shutdown()
}