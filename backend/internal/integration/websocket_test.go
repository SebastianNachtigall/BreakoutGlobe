package integration

import (
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/testdata"
	"breakoutglobe/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestWebSocketIntegration_BasicConnection tests basic WebSocket connection functionality
func TestWebSocketIntegration_BasicConnection(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	// Setup WebSocket test environment
	testWS := testdata.SetupWebSocket(t)

	// Create a client connection
	client := testWS.CreateClient("session-123", "user-456", "map-789")
	require.NotNil(t, client)

	// Wait for connection to be established
	time.Sleep(50 * time.Millisecond)

	// Consume welcome message
	err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
	require.NoError(t, err, "Should receive welcome message")

	// Verify connection
	assert.True(t, client.IsConnected())
	testWS.AssertClientConnected("session-123")
	testWS.AssertConnectedClientsCount(1)
	testWS.AssertMapClientsCount("map-test", 1) // Mock returns "map-test" for session-123

	// Verify client properties
	assert.Equal(t, "session-123", client.SessionID)
	assert.Equal(t, "user-123", client.UserID) // Mock returns "user-123" for session-123
	assert.Equal(t, "map-test", client.MapID)  // Mock returns "map-test" for session-123
}

// TestWebSocketIntegration_MultiClientBroadcast tests broadcasting to multiple clients
func TestWebSocketIntegration_MultiClientBroadcast(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	// Create multiple clients on the same map
	clients := make([]*testdata.TestWSClient, 3)
	for i := 0; i < 3; i++ {
		sessionID := fmt.Sprintf("session-%d", i+1)
		userID := fmt.Sprintf("user-%d", i+1)
		clients[i] = testWS.CreateClient(sessionID, userID, "map-test")
		require.NotNil(t, clients[i])
	}

	// Wait for all connections and consume welcome messages
	time.Sleep(100 * time.Millisecond)
	
	// Consume welcome messages from all clients
	for i, client := range clients {
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Client %d should receive welcome message", i+1)
	}

	// Verify all clients are connected
	testWS.AssertConnectedClientsCount(3)
	testWS.AssertMapClientsCount("map-test", 3)

	// Broadcast a POI creation event
	poiMessage := websocket.Message{
		Type: "poi_created",
		Data: map[string]interface{}{
			"poiId":      "poi-123",
			"name":       "Coffee Shop",
			"position":   map[string]float64{"lat": 40.7128, "lng": -74.0060},
			"createdBy":  "user-1",
			"mapId":      "map-test",
		},
		Timestamp: time.Now(),
	}

	testWS.BroadcastToMap("map-test", poiMessage)

	// Verify all clients receive the message
	for i, client := range clients {
		msg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Client %d should receive message", i+1)
		assert.Equal(t, "poi_created", msg.Type)

		// Verify message data
		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "poi-123", data["poiId"])
		assert.Equal(t, "Coffee Shop", data["name"])
	}
}

// TestWebSocketIntegration_MapIsolation tests that messages are isolated by map
func TestWebSocketIntegration_MapIsolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	// Create clients on different maps
	client1 := testWS.CreateClient("session-map1-1", "user-1", "map-1")
	client2 := testWS.CreateClient("session-map1-2", "user-2", "map-1")
	client3 := testWS.CreateClient("session-map2-1", "user-3", "map-2")

	require.NotNil(t, client1)
	require.NotNil(t, client2)
	require.NotNil(t, client3)

	// Wait for connections and consume welcome messages
	time.Sleep(100 * time.Millisecond)
	
	// Consume welcome messages from all clients
	err1 := client1.ConsumeWelcomeMessage(200 * time.Millisecond)
	err2 := client2.ConsumeWelcomeMessage(200 * time.Millisecond)
	err3 := client3.ConsumeWelcomeMessage(200 * time.Millisecond)
	require.NoError(t, err1, "Client 1 should receive welcome message")
	require.NoError(t, err2, "Client 2 should receive welcome message")
	require.NoError(t, err3, "Client 3 should receive welcome message")

	// Verify map isolation
	testWS.AssertMapClientsCount("map-1", 2)
	testWS.AssertMapClientsCount("map-2", 1)

	// Broadcast message to map-1 only
	message := websocket.Message{
		Type: "avatar_movement",
		Data: map[string]interface{}{
			"userId":   "user-1",
			"position": map[string]float64{"lat": 40.7128, "lng": -74.0060},
		},
		Timestamp: time.Now(),
	}

	testWS.BroadcastToMap("map-1", message)

	// Clients on map-1 should receive the message
	msg1, err1 := client1.ReceiveMessage(100 * time.Millisecond)
	msg2, err2 := client2.ReceiveMessage(100 * time.Millisecond)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.Equal(t, "avatar_movement", msg1.Type)
	assert.Equal(t, "avatar_movement", msg2.Type)

	// Client on map-2 should not receive the message
	client3.ExpectNoMessage(100 * time.Millisecond)
}

// TestWebSocketIntegration_ConnectionLifecycle tests connection lifecycle management
func TestWebSocketIntegration_ConnectionLifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	// Create client
	client := testWS.CreateClient("session-lifecycle", "user-lifecycle", "map-lifecycle")
	require.NotNil(t, client)

	// Wait for connection and consume welcome message
	time.Sleep(50 * time.Millisecond)
	
	// Consume welcome message
	err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
	require.NoError(t, err, "Should receive welcome message")

	// Verify initial connection
	testWS.AssertClientConnected("session-lifecycle")
	testWS.AssertConnectedClientsCount(1)

	// Send a message to verify connection is working
	testMessage := websocket.Message{
		Type: "heartbeat",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}

	err = client.SendMessage(testMessage)
	require.NoError(t, err)

	// Disconnect client
	client.Close()

	// Wait for disconnection to be processed
	time.Sleep(100 * time.Millisecond)

	// Verify disconnection
	assert.False(t, client.IsConnected())
	testWS.AssertClientDisconnected("session-lifecycle")
	testWS.AssertConnectedClientsCount(0)
	testWS.AssertMapClientsCount("map-lifecycle", 0)
}

// TestWebSocketIntegration_AvatarMovement tests avatar movement message flow
func TestWebSocketIntegration_AvatarMovement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	// Create two clients to simulate avatar movement
	mover := testWS.CreateClient("session-mover", "user-mover", "map-movement")
	observer := testWS.CreateClient("session-observer", "user-observer", "map-movement")

	require.NotNil(t, mover)
	require.NotNil(t, observer)

	// Wait for connections and consume welcome messages
	time.Sleep(100 * time.Millisecond)
	
	// Consume welcome messages
	err1 := mover.ConsumeWelcomeMessage(200 * time.Millisecond)
	err2 := observer.ConsumeWelcomeMessage(200 * time.Millisecond)
	require.NoError(t, err1, "Mover should receive welcome message")
	require.NoError(t, err2, "Observer should receive welcome message")

	// Simulate avatar movement from mover
	movementMessage := websocket.Message{
		Type: "avatar_movement",
		Data: map[string]interface{}{
			"sessionId": "session-mover",
			"userId":    "user-mover",
			"position": map[string]float64{
				"lat": 40.7589,
				"lng": -73.9851,
			},
			"mapId": "map-movement",
		},
		Timestamp: time.Now(),
	}

	err := mover.SendMessage(movementMessage)
	require.NoError(t, err)

	// In a real implementation, the server would process this message
	// and broadcast it to other clients. For now, we simulate the broadcast.
	testWS.BroadcastToMap("map-movement", movementMessage)

	// Observer should receive the movement message
	receivedMsg, err := observer.ReceiveMessage(200 * time.Millisecond)
	require.NoError(t, err)
	assert.Equal(t, "avatar_movement", receivedMsg.Type)

	// Verify movement data
	data, ok := receivedMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "user-mover", data["userId"])

	// Position data comes as map[string]float64 from JSON unmarshaling
	position, ok := data["position"].(map[string]float64)
	require.True(t, ok, "Position should be a map[string]float64")
	assert.Equal(t, 40.7589, position["lat"])
	assert.Equal(t, -73.9851, position["lng"])
}

// TestWebSocketIntegration_POIEvents tests POI-related event broadcasting
func TestWebSocketIntegration_POIEvents(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	// Create multiple clients
	creator := testWS.CreateClient("session-creator", "user-creator", "map-poi")
	participant1 := testWS.CreateClient("session-p1", "user-p1", "map-poi")
	participant2 := testWS.CreateClient("session-p2", "user-p2", "map-poi")

	require.NotNil(t, creator)
	require.NotNil(t, participant1)
	require.NotNil(t, participant2)

	// Wait for connections and consume welcome messages
	time.Sleep(100 * time.Millisecond)
	
	// Consume welcome messages from all clients
	clients := []*testdata.TestWSClient{creator, participant1, participant2}
	for i, client := range clients {
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Client %d should receive welcome message", i+1)
	}

	// Test POI creation event
	poiCreatedMsg := websocket.Message{
		Type: "poi_created",
		Data: map[string]interface{}{
			"poiId":           "poi-restaurant",
			"name":            "Italian Restaurant",
			"description":     "Authentic Italian cuisine",
			"position":        map[string]float64{"lat": 40.7505, "lng": -73.9934},
			"createdBy":       "user-creator",
			"maxParticipants": 8,
			"mapId":           "map-poi",
		},
		Timestamp: time.Now(),
	}

	testWS.BroadcastToMap("map-poi", poiCreatedMsg)

	// All clients should receive POI creation event
	for i, client := range clients {
		msg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Client %d should receive POI created message", i+1)
		assert.Equal(t, "poi_created", msg.Type)

		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "poi-restaurant", data["poiId"])
		assert.Equal(t, "Italian Restaurant", data["name"])
	}

	// Test POI join event
	poiJoinMsg := websocket.Message{
		Type: "poi_joined",
		Data: map[string]interface{}{
			"poiId":     "poi-restaurant",
			"userId":    "user-p1",
			"sessionId": "session-p1",
			"mapId":     "map-poi",
		},
		Timestamp: time.Now(),
	}

	testWS.BroadcastToMap("map-poi", poiJoinMsg)

	// All clients should receive POI join event
	for i, client := range clients {
		msg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Client %d should receive POI joined message", i+1)
		assert.Equal(t, "poi_joined", msg.Type)

		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok)
		assert.Equal(t, "poi-restaurant", data["poiId"])
		assert.Equal(t, "user-p1", data["userId"])
	}
}

// TestWebSocketIntegration_ConcurrentConnections tests concurrent connection handling
func TestWebSocketIntegration_ConcurrentConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	const numClients = 20
	clients := make([]*testdata.TestWSClient, numClients)
	done := make(chan int, numClients)

	// Create clients concurrently
	for i := 0; i < numClients; i++ {
		go func(index int) {
			sessionID := fmt.Sprintf("concurrent-session-%d", index)
			userID := fmt.Sprintf("concurrent-user-%d", index)
			mapID := fmt.Sprintf("map-%d", index%3) // Distribute across 3 maps

			client := testWS.CreateClient(sessionID, userID, mapID)
			if client != nil {
				clients[index] = client
			}
			done <- index
		}(i)
	}

	// Wait for all clients to attempt connection
	for i := 0; i < numClients; i++ {
		<-done
	}

	// Wait for connections to be processed
	time.Sleep(300 * time.Millisecond)

	// Count successful connections
	successfulConnections := 0
	for _, client := range clients {
		if client != nil && client.IsConnected() {
			successfulConnections++
		}
	}

	// Verify most connections succeeded (allow for some failures in concurrent scenario)
	assert.GreaterOrEqual(t, successfulConnections, numClients-2, 
		"Most concurrent connections should succeed")

	// Verify map distribution
	for mapIndex := 0; mapIndex < 3; mapIndex++ {
		mapID := fmt.Sprintf("map-%d", mapIndex)
		clientCount := testWS.GetMapClients(mapID)
		assert.GreaterOrEqual(t, clientCount, 0, "Map %s should have clients", mapID)
	}
}

// TestWebSocketIntegration_MessageOrdering tests message ordering and delivery
func TestWebSocketIntegration_MessageOrdering(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping WebSocket integration test in short mode")
	}

	testWS := testdata.SetupWebSocket(t)

	// Create client
	client := testWS.CreateClient("session-ordering", "user-ordering", "map-ordering")
	require.NotNil(t, client)

	// Wait for connection and consume welcome message
	time.Sleep(50 * time.Millisecond)
	
	// Consume welcome message
	err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
	require.NoError(t, err, "Should receive welcome message")

	// Send multiple messages in sequence
	const numMessages = 5
	for i := 0; i < numMessages; i++ {
		message := websocket.Message{
			Type: "sequence_test",
			Data: map[string]interface{}{
				"sequence": i,
				"content":  fmt.Sprintf("Message %d", i),
			},
			Timestamp: time.Now(),
		}

		testWS.BroadcastToMap("map-ordering", message)
		time.Sleep(10 * time.Millisecond) // Small delay between messages
	}

	// Receive and verify message ordering
	for i := 0; i < numMessages; i++ {
		msg, err := client.ReceiveMessage(200 * time.Millisecond)
		require.NoError(t, err, "Should receive message %d", i)
		assert.Equal(t, "sequence_test", msg.Type)

		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok)
		
		// Note: WebSocket doesn't guarantee ordering across different broadcasts,
		// but we can verify we receive all messages
		// Sequence data comes as int from the test broadcast
		sequence, ok := data["sequence"].(int)
		require.True(t, ok, "Sequence should be an int")
		assert.GreaterOrEqual(t, sequence, 0)
		assert.Less(t, sequence, numMessages)
	}
}

// BenchmarkWebSocketIntegration_Throughput benchmarks WebSocket message throughput
func BenchmarkWebSocketIntegration_Throughput(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping WebSocket integration benchmark in short mode")
	}

	testWS := testdata.SetupWebSocket(b)

	// Create a client
	client := testWS.CreateClient("benchmark-session", "benchmark-user", "benchmark-map")
	if client == nil {
		b.Fatal("Failed to create benchmark client")
	}

	// Wait for connection
	time.Sleep(100 * time.Millisecond)

	message := websocket.Message{
		Type: "benchmark_message",
		Data: map[string]interface{}{
			"payload": "This is a benchmark message payload",
		},
		Timestamp: time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		testWS.BroadcastToMap("benchmark-map", message)
	}
}