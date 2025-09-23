package testdata

import (
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSetupWebSocket verifies that WebSocket test setup works correctly
func TestSetupWebSocket(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	// Setup WebSocket test environment
	testWS := SetupWebSocket(t)

	// Verify server is running
	assert.NotNil(t, testWS.server)
	assert.NotNil(t, testWS.handler)

	// Verify initial state
	assert.Equal(t, 0, testWS.GetConnectedClients())
}

// TestWebSocketClientConnection verifies client connection functionality
func TestWebSocketClientConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create a test client
	client := testWS.CreateClient("session-1", "user-1", "map-1")
	require.NotNil(t, client)

	// Wait for connection to be established
	time.Sleep(50 * time.Millisecond)

	// Verify client is connected
	assert.True(t, client.IsConnected())
	testWS.AssertClientConnected("session-1")
	testWS.AssertConnectedClientsCount(1)
	testWS.AssertMapClientsCount("map-1", 1)

	// Verify client properties
	assert.Equal(t, "session-1", client.SessionID)
	assert.Equal(t, "user-1", client.UserID)
	assert.Equal(t, "map-1", client.MapID)
}

// TestWebSocketClientDisconnection verifies client disconnection functionality
func TestWebSocketClientDisconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create and connect client
	client := testWS.CreateClient("session-1", "user-1", "map-1")
	require.NotNil(t, client)

	// Wait for connection
	time.Sleep(50 * time.Millisecond)
	testWS.AssertClientConnected("session-1")

	// Disconnect client
	client.Close()

	// Wait for disconnection to be processed
	time.Sleep(50 * time.Millisecond)

	// Verify client is disconnected
	assert.False(t, client.IsConnected())
	testWS.AssertClientDisconnected("session-1")
	testWS.AssertConnectedClientsCount(0)
	testWS.AssertMapClientsCount("map-1", 0)
}

// TestWebSocketMultipleClients verifies multiple client connections
func TestWebSocketMultipleClients(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create multiple clients
	clients := make([]*TestWSClient, 3)
	for i := 0; i < 3; i++ {
		sessionID := fmt.Sprintf("session-%d", i+1)
		userID := fmt.Sprintf("user-%d", i+1)
		clients[i] = testWS.CreateClient(sessionID, userID, "map-1")
		require.NotNil(t, clients[i])
	}

	// Wait for connections
	time.Sleep(100 * time.Millisecond)

	// Verify all clients are connected
	testWS.AssertConnectedClientsCount(3)
	testWS.AssertMapClientsCount("map-1", 3)

	for i, client := range clients {
		sessionID := fmt.Sprintf("session-%d", i+1)
		assert.True(t, client.IsConnected())
		testWS.AssertClientConnected(sessionID)
	}

	// Disconnect one client
	clients[1].Close()
	time.Sleep(50 * time.Millisecond)

	// Verify counts updated
	testWS.AssertConnectedClientsCount(2)
	testWS.AssertMapClientsCount("map-1", 2)
	testWS.AssertClientDisconnected("session-2")
}

// TestWebSocketMultipleMaps verifies clients on different maps
func TestWebSocketMultipleMaps(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create clients on different maps
	client1 := testWS.CreateClient("session-1", "user-1", "map-1")
	client2 := testWS.CreateClient("session-2", "user-2", "map-1")
	client3 := testWS.CreateClient("session-3", "user-3", "map-2")
	client4 := testWS.CreateClient("session-4", "user-4", "map-2")

	require.NotNil(t, client1)
	require.NotNil(t, client2)
	require.NotNil(t, client3)
	require.NotNil(t, client4)

	// Wait for connections
	time.Sleep(100 * time.Millisecond)

	// Verify total connections
	testWS.AssertConnectedClientsCount(4)

	// Verify map-specific connections
	testWS.AssertMapClientsCount("map-1", 2)
	testWS.AssertMapClientsCount("map-2", 2)
	testWS.AssertMapClientsCount("map-3", 0)
}

// TestWebSocketMessageSending verifies message sending functionality
func TestWebSocketMessageSending(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create client
	client := testWS.CreateClient("session-1", "user-1", "map-1")
	require.NotNil(t, client)

	// Wait for connection
	time.Sleep(50 * time.Millisecond)

	// Send a message from client
	message := websocket.Message{
		Type: "avatar_movement",
		Data: map[string]interface{}{
			"position": map[string]float64{
				"lat": 40.7128,
				"lng": -74.0060,
			},
		},
		Timestamp: time.Now(),
	}

	err := client.SendMessage(message)
	require.NoError(t, err)

	// Note: In a real implementation, we would verify the server processes the message
	// For now, we just verify the message was sent without error
}

// TestWebSocketBroadcasting verifies message broadcasting functionality
func TestWebSocketBroadcasting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create multiple clients on the same map
	client1 := testWS.CreateClient("session-1", "user-1", "map-1")
	client2 := testWS.CreateClient("session-2", "user-2", "map-1")
	client3 := testWS.CreateClient("session-3", "user-3", "map-2") // Different map

	require.NotNil(t, client1)
	require.NotNil(t, client2)
	require.NotNil(t, client3)

	// Wait for connections
	time.Sleep(100 * time.Millisecond)

	// Broadcast message to map-1
	broadcastMessage := websocket.Message{
		Type: "poi_created",
		Data: map[string]interface{}{
			"poiId": "poi-123",
			"name":  "Test POI",
		},
		Timestamp: time.Now(),
	}

	testWS.BroadcastToMap("map-1", broadcastMessage)

	// Wait for message delivery
	time.Sleep(50 * time.Millisecond)

	// Verify clients on map-1 received the message
	msg1, err1 := client1.ReceiveMessage(100 * time.Millisecond)
	msg2, err2 := client2.ReceiveMessage(100 * time.Millisecond)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "poi_created", msg1.Type)
	assert.Equal(t, "poi_created", msg2.Type)

	// Verify client on map-2 did not receive the message
	client3.ExpectNoMessage(100 * time.Millisecond)
}

// TestWebSocketMessageReceiving verifies message receiving functionality
func TestWebSocketMessageReceiving(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create client
	client := testWS.CreateClient("session-1", "user-1", "map-1")
	require.NotNil(t, client)

	// Wait for connection
	time.Sleep(50 * time.Millisecond)

	// Send a broadcast message
	message := websocket.Message{
		Type: "test_message",
		Data: map[string]interface{}{
			"content": "Hello, WebSocket!",
		},
		Timestamp: time.Now(),
	}

	testWS.BroadcastToMap("map-1", message)

	// Receive and verify message
	receivedMsg := client.ExpectMessage("test_message", 100*time.Millisecond)
	assert.Equal(t, "test_message", receivedMsg.Type)

	// Verify message data
	data, ok := receivedMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "Hello, WebSocket!", data["content"])
}

// TestWebSocketConcurrentConnections verifies concurrent connection handling
func TestWebSocketConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	const numClients = 10
	clients := make([]*TestWSClient, numClients)
	done := make(chan bool, numClients)

	// Create clients concurrently
	for i := 0; i < numClients; i++ {
		go func(index int) {
			sessionID := fmt.Sprintf("session-%d", index)
			userID := fmt.Sprintf("user-%d", index)
			clients[index] = testWS.CreateClient(sessionID, userID, "map-1")
			done <- true
		}(i)
	}

	// Wait for all clients to connect
	for i := 0; i < numClients; i++ {
		<-done
	}

	// Wait for connections to be processed
	time.Sleep(200 * time.Millisecond)

	// Verify all clients are connected
	testWS.AssertConnectedClientsCount(numClients)
	testWS.AssertMapClientsCount("map-1", numClients)

	// Verify each client individually
	for i, client := range clients {
		require.NotNil(t, client, "Client %d should not be nil", i)
		assert.True(t, client.IsConnected(), "Client %d should be connected", i)
	}
}

// TestWebSocketReconnection verifies reconnection handling
func TestWebSocketReconnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := SetupWebSocket(t)

	// Create initial client
	client1 := testWS.CreateClient("session-1", "user-1", "map-1")
	require.NotNil(t, client1)

	// Wait for connection
	time.Sleep(50 * time.Millisecond)
	testWS.AssertClientConnected("session-1")

	// Disconnect client
	client1.Close()
	time.Sleep(50 * time.Millisecond)
	testWS.AssertClientDisconnected("session-1")

	// Reconnect with same session ID (simulating reconnection)
	client2 := testWS.CreateClient("session-1", "user-1", "map-1")
	require.NotNil(t, client2)

	// Wait for reconnection
	time.Sleep(50 * time.Millisecond)
	testWS.AssertClientConnected("session-1")
	testWS.AssertConnectedClientsCount(1)
}

// BenchmarkWebSocketConnections benchmarks WebSocket connection performance
func BenchmarkWebSocketConnections(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping WebSocket integration benchmark in short mode")
	}

	testWS := SetupWebSocket(b)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		userID := fmt.Sprintf("user-%d", i)
		client := testWS.CreateClient(sessionID, userID, "map-1")
		if client == nil {
			b.Fatalf("Failed to create client %d", i)
		}
		client.Close()
	}
}

// BenchmarkWebSocketBroadcasting benchmarks message broadcasting performance
func BenchmarkWebSocketBroadcasting(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping WebSocket integration benchmark in short mode")
	}

	testWS := SetupWebSocket(b)

	// Create multiple clients
	const numClients = 100
	clients := make([]*TestWSClient, numClients)
	for i := 0; i < numClients; i++ {
		sessionID := fmt.Sprintf("session-%d", i)
		userID := fmt.Sprintf("user-%d", i)
		clients[i] = testWS.CreateClient(sessionID, userID, "map-1")
		if clients[i] == nil {
			b.Fatalf("Failed to create client %d", i)
		}
	}

	// Wait for connections
	time.Sleep(500 * time.Millisecond)

	message := websocket.Message{
		Type: "benchmark_message",
		Data: map[string]interface{}{
			"content": "Benchmark message",
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testWS.BroadcastToMap("map-1", message)
	}
}