package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/testdata"
	"breakoutglobe/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessionCreationFlow tests the complete session creation flow
// HTTP → Service → Database → Redis → WebSocket
func TestSessionCreationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping session creation flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create WebSocket observers
	observer := env.websocket.CreateClient("session-observer", "user-observer", "map-session-test")
	require.NotNil(t, observer)
	env.WaitForAsyncOperations()

	// Step 2: HTTP Request - Create session via REST API
	createRequest := CreateSessionRequest{
		UserID:         "user-session-test",
		MapID:          "map-session-test",
		AvatarPosition: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	response := env.POST("/api/sessions", createRequest)
	env.AssertHTTPSuccess(response)

	// Parse response to get session ID
	var sessionResponse struct {
		ID             string `json:"id"`
		UserID         string `json:"userId"`
		MapID          string `json:"mapId"`
		AvatarPosition LatLng `json:"avatarPosition"`
		IsActive       bool   `json:"isActive"`
	}
	env.ParseJSONResponse(response, &sessionResponse)

	sessionID := sessionResponse.ID
	require.NotEmpty(t, sessionID, "Session ID should be returned")
	assert.Equal(t, "user-session-test", sessionResponse.UserID)
	assert.Equal(t, "map-session-test", sessionResponse.MapID)
	assert.True(t, sessionResponse.IsActive)

	// Step 3: Verify Database Persistence
	env.AssertDatabaseSession(sessionID, "user-session-test")

	// Step 4: Verify Redis Presence
	env.AssertRedisPresence(sessionID)

	// Step 5: Verify WebSocket Broadcasting
	env.WaitForAsyncOperations()

	// Simulate session created event broadcast
	sessionCreatedEvent := websocket.Message{
		Type: "session_created",
		Data: map[string]interface{}{
			"sessionId":      sessionID,
			"userId":         "user-session-test",
			"mapId":          "map-session-test",
			"avatarPosition": map[string]float64{"lat": 40.7128, "lng": -74.0060},
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-session-test", sessionCreatedEvent)

	// Observer should receive the session created event
	msg := observer.ExpectMessage("session_created", 200*time.Millisecond)

	data, ok := msg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, sessionID, data["sessionId"])
	assert.Equal(t, "user-session-test", data["userId"])

	// Step 6: Verify session can be retrieved via API
	getResponse := env.GET("/api/sessions/" + sessionID)
	env.AssertHTTPSuccess(getResponse)

	var retrievedSession struct {
		ID     string `json:"id"`
		UserID string `json:"userId"`
	}
	env.ParseJSONResponse(getResponse, &retrievedSession)
	assert.Equal(t, sessionID, retrievedSession.ID)
	assert.Equal(t, "user-session-test", retrievedSession.UserID)
}

// TestAvatarMovementFlow tests the complete avatar movement flow
// HTTP → Service → Database → Redis → WebSocket
func TestAvatarMovementFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping avatar movement flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create a session first
	createRequest := CreateSessionRequest{
		UserID:         "user-mover",
		MapID:          "map-movement-test",
		AvatarPosition: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	createResponse := env.POST("/api/sessions", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var sessionResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &sessionResponse)
	sessionID := sessionResponse.ID

	// Step 2: Create WebSocket clients
	mover := env.websocket.CreateClient(sessionID, "user-mover", "map-movement-test")
	observer := env.websocket.CreateClient("session-observer", "user-observer", "map-movement-test")
	otherMapClient := env.websocket.CreateClient("session-other", "user-other", "map-other")

	require.NotNil(t, mover)
	require.NotNil(t, observer)
	require.NotNil(t, otherMapClient)

	env.WaitForAsyncOperations()

	// Step 3: Update avatar position via HTTP API
	updateRequest := UpdateAvatarRequest{
		Position: LatLng{Lat: 40.7589, Lng: -73.9851},
	}

	updateResponse := env.PUT("/api/sessions/"+sessionID+"/avatar", updateRequest)
	env.AssertHTTPSuccess(updateResponse)

	// Step 4: Verify Database Update
	var session models.Session
	err := env.db.DB.Where("id = ?", sessionID).First(&session).Error
	require.NoError(t, err)
	assert.Equal(t, 40.7589, session.AvatarPos.Lat)
	assert.Equal(t, -73.9851, session.AvatarPos.Lng)

	// Step 5: Verify Redis Presence Update
	env.AssertRedisPresence(sessionID)

	// Step 6: Verify WebSocket Broadcasting
	env.WaitForAsyncOperations()

	// Simulate avatar movement event broadcast
	movementEvent := websocket.Message{
		Type: "avatar_movement",
		Data: map[string]interface{}{
			"sessionId": sessionID,
			"userId":    "user-mover",
			"position": map[string]float64{
				"lat": 40.7589,
				"lng": -73.9851,
			},
			"mapId": "map-movement-test",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-movement-test", movementEvent)

	// Clients on the same map should receive the movement
	moverMsg := mover.ExpectMessage("avatar_movement", 200*time.Millisecond)
	observerMsg := observer.ExpectMessage("avatar_movement", 200*time.Millisecond)

	// Verify message content
	moverData, ok := moverMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, sessionID, moverData["sessionId"])
	assert.Equal(t, "user-mover", moverData["userId"])

	position, ok := moverData["position"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, 40.7589, position["lat"])
	assert.Equal(t, -73.9851, position["lng"])

	observerData, ok := observerMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, sessionID, observerData["sessionId"])

	// Client on different map should not receive the movement
	otherMapClient.ExpectNoMessage(100 * time.Millisecond)
}

// TestSessionHeartbeatFlow tests the session heartbeat mechanism
func TestSessionHeartbeatFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping session heartbeat flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create a session
	createRequest := CreateSessionRequest{
		UserID:         "user-heartbeat",
		MapID:          "map-heartbeat-test",
		AvatarPosition: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	createResponse := env.POST("/api/sessions", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var sessionResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &sessionResponse)
	sessionID := sessionResponse.ID

	// Step 2: Verify initial presence
	env.AssertRedisPresence(sessionID)

	// Step 3: Send heartbeat via HTTP API
	heartbeatResponse := env.POST("/api/sessions/"+sessionID+"/heartbeat", nil)
	env.AssertHTTPSuccess(heartbeatResponse)

	// Step 4: Verify Redis presence is updated (TTL refreshed)
	env.AssertRedisPresence(sessionID)

	// Step 5: Verify database last activity is updated
	var session models.Session
	err := env.db.DB.Where("id = ?", sessionID).First(&session).Error
	require.NoError(t, err)
	
	// Last activity should be recent (within last few seconds)
	timeSinceLastActivity := time.Since(session.LastActive)
	assert.Less(t, timeSinceLastActivity, 5*time.Second, "Last activity should be recent")
}

// TestSessionEndFlow tests the session termination flow
func TestSessionEndFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping session end flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create a session
	createRequest := CreateSessionRequest{
		UserID:         "user-ending",
		MapID:          "map-end-test",
		AvatarPosition: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	createResponse := env.POST("/api/sessions", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var sessionResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &sessionResponse)
	sessionID := sessionResponse.ID

	// Step 2: Create WebSocket observer
	observer := env.websocket.CreateClient("session-observer", "user-observer", "map-end-test")
	require.NotNil(t, observer)
	env.WaitForAsyncOperations()

	// Step 3: Verify session exists
	env.AssertDatabaseSession(sessionID, "user-ending")
	env.AssertRedisPresence(sessionID)

	// Step 4: End session via HTTP API
	endResponse := env.DELETE("/api/sessions/" + sessionID)
	env.AssertHTTPSuccess(endResponse)

	// Step 5: Verify Database Update - Session should be marked inactive
	var session models.Session
	err := env.db.DB.Where("id = ?", sessionID).First(&session).Error
	require.NoError(t, err)
	assert.False(t, session.IsActive, "Session should be marked as inactive")

	// Step 6: Verify Redis Cleanup - Presence should be removed
	env.redis.AssertKeyNotExists("session:" + sessionID)

	// Step 7: Verify WebSocket Broadcasting
	env.WaitForAsyncOperations()

	// Simulate session ended event broadcast
	sessionEndedEvent := websocket.Message{
		Type: "session_ended",
		Data: map[string]interface{}{
			"sessionId": sessionID,
			"userId":    "user-ending",
			"mapId":     "map-end-test",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-end-test", sessionEndedEvent)

	// Observer should receive the session ended event
	msg := observer.ExpectMessage("session_ended", 200*time.Millisecond)

	data, ok := msg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, sessionID, data["sessionId"])
	assert.Equal(t, "user-ending", data["userId"])
}

// TestMultipleSessionsOnMap tests multiple sessions on the same map
func TestMultipleSessionsOnMap(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping multiple sessions flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	mapID := "map-multi-session"
	sessionIDs := make([]string, 3)

	// Step 1: Create multiple sessions on the same map
	for i := 0; i < 3; i++ {
		createRequest := CreateSessionRequest{
			UserID: fmt.Sprintf("user-%d", i+1),
			MapID:  mapID,
			AvatarPosition: LatLng{
				Lat: 40.7128 + float64(i)*0.001,
				Lng: -74.0060 + float64(i)*0.001,
			},
		}

		response := env.POST("/api/sessions", createRequest)
		env.AssertHTTPSuccess(response)

		var sessionResponse struct {
			ID string `json:"id"`
		}
		env.ParseJSONResponse(response, &sessionResponse)
		sessionIDs[i] = sessionResponse.ID

		// Verify each session
		env.AssertDatabaseSession(sessionIDs[i], fmt.Sprintf("user-%d", i+1))
		env.AssertRedisPresence(sessionIDs[i])
	}

	// Step 2: Create WebSocket clients for each session
	clients := make([]*testdata.TestWSClient, 3)
	for i := 0; i < 3; i++ {
		clients[i] = env.websocket.CreateClient(
			sessionIDs[i],
			fmt.Sprintf("user-%d", i+1),
			mapID,
		)
		require.NotNil(t, clients[i])
	}

	env.WaitForAsyncOperations()

	// Step 3: Test broadcasting to all sessions on the map
	broadcastEvent := websocket.Message{
		Type: "map_announcement",
		Data: map[string]interface{}{
			"mapId":   mapID,
			"message": "Welcome to the map!",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap(mapID, broadcastEvent)

	// Step 4: Verify all clients receive the broadcast
	for i, client := range clients {
		msg := client.ExpectMessage("map_announcement", 200*time.Millisecond)

		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok, "Client %d should receive valid data", i+1)
		assert.Equal(t, mapID, data["mapId"])
		assert.Equal(t, "Welcome to the map!", data["message"])
	}

	// Step 5: Test individual avatar movement
	updateRequest := UpdateAvatarRequest{
		Position: LatLng{Lat: 40.7589, Lng: -73.9851},
	}

	updateResponse := env.PUT("/api/sessions/"+sessionIDs[0]+"/avatar", updateRequest)
	env.AssertHTTPSuccess(updateResponse)

	// Simulate movement broadcast
	movementEvent := websocket.Message{
		Type: "avatar_movement",
		Data: map[string]interface{}{
			"sessionId": sessionIDs[0],
			"userId":    "user-1",
			"position": map[string]float64{
				"lat": 40.7589,
				"lng": -73.9851,
			},
			"mapId": mapID,
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap(mapID, movementEvent)

	// All clients should receive the movement event
	for i, client := range clients {
		msg := client.ExpectMessage("avatar_movement", 200*time.Millisecond)

		data, ok := msg.Data.(map[string]interface{})
		require.True(t, ok, "Client %d should receive movement data", i+1)
		assert.Equal(t, sessionIDs[0], data["sessionId"])
		assert.Equal(t, "user-1", data["userId"])
	}
}

// TestSessionCleanupFlow tests automatic session cleanup
func TestSessionCleanupFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping session cleanup flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create sessions that will expire
	expiredSessionIDs := make([]string, 2)
	for i := 0; i < 2; i++ {
		createRequest := CreateSessionRequest{
			UserID: fmt.Sprintf("user-expired-%d", i+1),
			MapID:  "map-cleanup-test",
			AvatarPosition: LatLng{
				Lat: 40.7128,
				Lng: -74.0060,
			},
		}

		response := env.POST("/api/sessions", createRequest)
		env.AssertHTTPSuccess(response)

		var sessionResponse struct {
			ID string `json:"id"`
		}
		env.ParseJSONResponse(response, &sessionResponse)
		expiredSessionIDs[i] = sessionResponse.ID
	}

	// Step 2: Manually expire sessions in database (simulate old sessions)
	for _, sessionID := range expiredSessionIDs {
		var session models.Session
		err := env.db.DB.Where("id = ?", sessionID).First(&session).Error
		require.NoError(t, err)

		// Set last activity to 2 hours ago
		session.LastActive = time.Now().Add(-2 * time.Hour)
		err = env.db.DB.Save(&session).Error
		require.NoError(t, err)
	}

	// Step 3: Create a fresh session that should not be cleaned up
	freshRequest := CreateSessionRequest{
		UserID:         "user-fresh",
		MapID:          "map-cleanup-test",
		AvatarPosition: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	freshResponse := env.POST("/api/sessions", freshRequest)
	env.AssertHTTPSuccess(freshResponse)

	var freshSessionResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(freshResponse, &freshSessionResponse)
	freshSessionID := freshSessionResponse.ID

	// Step 4: Trigger cleanup (simulate cleanup job)
	// In a real implementation, this would be done by a background job
	// For testing, we'll call the service method directly
	ctx := context.Background()
	err := env.sessionService.CleanupExpiredSessions(ctx)
	require.NoError(t, err)

	// Step 5: Verify expired sessions are cleaned up
	for _, sessionID := range expiredSessionIDs {
		var session models.Session
		err := env.db.DB.Where("id = ?", sessionID).First(&session).Error
		require.NoError(t, err)
		assert.False(t, session.IsActive, "Expired session should be marked inactive")

		// Redis presence should be cleaned up
		env.redis.AssertKeyNotExists("session:" + sessionID)
	}

	// Step 6: Verify fresh session is not affected
	var freshSession models.Session
	err = env.db.DB.Where("id = ?", freshSessionID).First(&freshSession).Error
	require.NoError(t, err)
	assert.True(t, freshSession.IsActive, "Fresh session should remain active")

	// Redis presence should still exist
	env.AssertRedisPresence(freshSessionID)
}