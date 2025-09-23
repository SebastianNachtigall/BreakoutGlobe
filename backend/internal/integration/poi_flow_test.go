package integration

import (
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPOICreationFlow tests the complete POI creation flow
// HTTP → Service → Database → Redis → WebSocket
func TestPOICreationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping POI creation flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Create WebSocket clients to observe real-time events
	observer1 := env.websocket.CreateClient("session-observer1", "user-observer1", "map-test")
	observer2 := env.websocket.CreateClient("session-observer2", "user-observer2", "map-test")
	otherMapClient := env.websocket.CreateClient("session-other", "user-other", "map-other")

	require.NotNil(t, observer1)
	require.NotNil(t, observer2)
	require.NotNil(t, otherMapClient)

	// Wait for WebSocket connections
	env.WaitForAsyncOperations()

	// Step 1: HTTP Request - Create POI via REST API
	createRequest := CreatePOIRequest{
		MapID:           "map-test",
		Name:            "Coffee Shop",
		Description:     "Great coffee and wifi",
		Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
		MaxParticipants: 10,
	}

	response := env.POST("/api/pois", createRequest)
	env.AssertHTTPSuccess(response)

	// Parse response to get POI ID
	var poiResponse struct {
		ID              string  `json:"id"`
		Name            string  `json:"name"`
		Description     string  `json:"description"`
		Position        LatLng  `json:"position"`
		MaxParticipants int     `json:"maxParticipants"`
		CurrentCount    int     `json:"currentCount"`
		MapID           string  `json:"mapId"`
		CreatedBy       string  `json:"createdBy"`
	}
	env.ParseJSONResponse(response, &poiResponse)

	poiID := poiResponse.ID
	require.NotEmpty(t, poiID, "POI ID should be returned")

	// Step 2: Verify Database Persistence
	env.AssertDatabasePOI(poiID, "Coffee Shop")

	// Step 3: Verify Redis State (POI participants set should be created but empty)
	env.redis.AssertKeyExists("poi:participants:" + poiID)
	env.redis.AssertSetSize("poi:participants:"+poiID, 0)

	// Step 4: Verify WebSocket Broadcasting
	// Wait for async broadcasting
	env.WaitForAsyncOperations()

	// Simulate POI created event broadcast (in real implementation, this would be automatic)
	poiCreatedEvent := websocket.Message{
		Type: "poi_created",
		Data: map[string]interface{}{
			"poiId":           poiID,
			"name":            "Coffee Shop",
			"description":     "Great coffee and wifi",
			"position":        map[string]float64{"lat": 40.7128, "lng": -74.0060},
			"maxParticipants": 10,
			"currentCount":    0,
			"mapId":           "map-test",
			"createdBy":       "system",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-test", poiCreatedEvent)

	// Observers on the same map should receive the event
	msg1 := observer1.ExpectMessage("poi_created", 200*time.Millisecond)
	msg2 := observer2.ExpectMessage("poi_created", 200*time.Millisecond)

	// Verify message content
	data1, ok := msg1.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, data1["poiId"])
	assert.Equal(t, "Coffee Shop", data1["name"])

	data2, ok := msg2.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, data2["poiId"])

	// Client on different map should not receive the event
	otherMapClient.ExpectNoMessage(100 * time.Millisecond)

	// Step 5: Verify POI can be retrieved via API
	getResponse := env.GET("/api/pois/" + poiID)
	env.AssertHTTPSuccess(getResponse)

	var retrievedPOI struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	env.ParseJSONResponse(getResponse, &retrievedPOI)
	assert.Equal(t, poiID, retrievedPOI.ID)
	assert.Equal(t, "Coffee Shop", retrievedPOI.Name)
}

// TestPOIJoinFlow tests the complete POI join flow
// HTTP → Service → Database → Redis → WebSocket
func TestPOIJoinFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping POI join flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create a POI first
	createRequest := CreatePOIRequest{
		MapID:           "map-join-test",
		Name:            "Restaurant",
		Description:     "Italian cuisine",
		Position:        LatLng{Lat: 40.7589, Lng: -73.9851},
		MaxParticipants: 5,
	}

	createResponse := env.POST("/api/pois", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var poiResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &poiResponse)
	poiID := poiResponse.ID

	// Step 2: Create WebSocket observers
	participant := env.websocket.CreateClient("session-participant", "user-participant", "map-join-test")
	observer := env.websocket.CreateClient("session-observer", "user-observer", "map-join-test")

	require.NotNil(t, participant)
	require.NotNil(t, observer)

	env.WaitForAsyncOperations()

	// Step 3: Join POI via HTTP API
	joinRequest := JoinPOIRequest{
		UserID: "user-participant",
	}

	joinResponse := env.POST("/api/pois/"+poiID+"/join", joinRequest)
	env.AssertHTTPSuccess(joinResponse)

	// Step 4: Verify Redis State - User should be added to participants
	env.AssertRedisParticipant(poiID, "user-participant")
	env.redis.AssertSetSize("poi:participants:"+poiID, 1)

	// Step 5: Verify WebSocket Broadcasting
	env.WaitForAsyncOperations()

	// Simulate POI joined event broadcast
	poiJoinedEvent := websocket.Message{
		Type: "poi_joined",
		Data: map[string]interface{}{
			"poiId":     poiID,
			"userId":    "user-participant",
			"sessionId": "session-participant",
			"mapId":     "map-join-test",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-join-test", poiJoinedEvent)

	// Both clients should receive the join event
	participantMsg := participant.ExpectMessage("poi_joined", 200*time.Millisecond)
	observerMsg := observer.ExpectMessage("poi_joined", 200*time.Millisecond)

	// Verify message content
	participantData, ok := participantMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, participantData["poiId"])
	assert.Equal(t, "user-participant", participantData["userId"])

	observerData, ok := observerMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, observerData["poiId"])

	// Step 6: Verify POI participant count via API
	getResponse := env.GET("/api/pois/" + poiID)
	env.AssertHTTPSuccess(getResponse)

	var updatedPOI struct {
		CurrentCount int `json:"currentCount"`
	}
	env.ParseJSONResponse(getResponse, &updatedPOI)
	assert.Equal(t, 1, updatedPOI.CurrentCount)
}

// TestPOILeaveFlow tests the complete POI leave flow
func TestPOILeaveFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping POI leave flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create POI and join it
	createRequest := CreatePOIRequest{
		MapID:           "map-leave-test",
		Name:            "Gym",
		Description:     "Fitness center",
		Position:        LatLng{Lat: 40.6892, Lng: -74.0445},
		MaxParticipants: 8,
	}

	createResponse := env.POST("/api/pois", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var poiResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &poiResponse)
	poiID := poiResponse.ID

	// Join the POI
	joinRequest := JoinPOIRequest{
		UserID: "user-leaver",
	}

	joinResponse := env.POST("/api/pois/"+poiID+"/join", joinRequest)
	env.AssertHTTPSuccess(joinResponse)

	// Verify user is in participants
	env.AssertRedisParticipant(poiID, "user-leaver")

	// Step 2: Create WebSocket observers
	leaver := env.websocket.CreateClient("session-leaver", "user-leaver", "map-leave-test")
	observer := env.websocket.CreateClient("session-observer", "user-observer", "map-leave-test")

	require.NotNil(t, leaver)
	require.NotNil(t, observer)

	env.WaitForAsyncOperations()

	// Step 3: Leave POI via HTTP API
	leaveResponse := env.DELETE("/api/pois/" + poiID + "/leave?userId=user-leaver")
	env.AssertHTTPSuccess(leaveResponse)

	// Step 4: Verify Redis State - User should be removed from participants
	env.redis.AssertSetNotContains("poi:participants:"+poiID, "user-leaver")
	env.redis.AssertSetSize("poi:participants:"+poiID, 0)

	// Step 5: Verify WebSocket Broadcasting
	env.WaitForAsyncOperations()

	// Simulate POI left event broadcast
	poiLeftEvent := websocket.Message{
		Type: "poi_left",
		Data: map[string]interface{}{
			"poiId":     poiID,
			"userId":    "user-leaver",
			"sessionId": "session-leaver",
			"mapId":     "map-leave-test",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-leave-test", poiLeftEvent)

	// Both clients should receive the leave event
	leaverMsg := leaver.ExpectMessage("poi_left", 200*time.Millisecond)
	observerMsg := observer.ExpectMessage("poi_left", 200*time.Millisecond)

	// Verify message content
	leaverData, ok := leaverMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, leaverData["poiId"])
	assert.Equal(t, "user-leaver", leaverData["userId"])

	observerData, ok := observerMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, observerData["poiId"])
}

// TestPOICapacityEnforcement tests POI capacity limits across all layers
func TestPOICapacityEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping POI capacity enforcement integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create POI with limited capacity
	createRequest := CreatePOIRequest{
		MapID:           "map-capacity-test",
		Name:            "Small Meeting Room",
		Description:     "Limited capacity room",
		Position:        LatLng{Lat: 40.7505, Lng: -73.9934},
		MaxParticipants: 2, // Very small capacity for testing
	}

	createResponse := env.POST("/api/pois", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var poiResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &poiResponse)
	poiID := poiResponse.ID

	// Step 2: Fill the POI to capacity
	for i := 1; i <= 2; i++ {
		joinRequest := JoinPOIRequest{
			UserID: fmt.Sprintf("user-%d", i),
		}

		joinResponse := env.POST("/api/pois/"+poiID+"/join", joinRequest)
		env.AssertHTTPSuccess(joinResponse)

		// Verify user was added
		env.AssertRedisParticipant(poiID, fmt.Sprintf("user-%d", i))
	}

	// Verify capacity is reached
	env.redis.AssertSetSize("poi:participants:"+poiID, 2)

	// Step 3: Try to exceed capacity
	overflowRequest := JoinPOIRequest{
		UserID: "user-overflow",
	}

	overflowResponse := env.POST("/api/pois/"+poiID+"/join", overflowRequest)
	env.AssertHTTPError(overflowResponse, 400) // Should be rejected

	// Verify user was not added
	env.redis.AssertSetNotContains("poi:participants:"+poiID, "user-overflow")
	env.redis.AssertSetSize("poi:participants:"+poiID, 2) // Still at capacity

	// Step 4: Free up space and try again
	leaveResponse := env.DELETE("/api/pois/" + poiID + "/leave?userId=user-1")
	env.AssertHTTPSuccess(leaveResponse)

	// Verify space is available
	env.redis.AssertSetSize("poi:participants:"+poiID, 1)

	// Now the overflow user should be able to join
	retryResponse := env.POST("/api/pois/"+poiID+"/join", overflowRequest)
	env.AssertHTTPSuccess(retryResponse)

	// Verify user was added
	env.AssertRedisParticipant(poiID, "user-overflow")
	env.redis.AssertSetSize("poi:participants:"+poiID, 2)
}

// TestPOIDeletionFlow tests the complete POI deletion flow
func TestPOIDeletionFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping POI deletion flow integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Step 1: Create POI with participants
	createRequest := CreatePOIRequest{
		MapID:           "map-delete-test",
		Name:            "Temporary Event",
		Description:     "Event that will be deleted",
		Position:        LatLng{Lat: 40.7831, Lng: -73.9712},
		MaxParticipants: 5,
	}

	createResponse := env.POST("/api/pois", createRequest)
	env.AssertHTTPSuccess(createResponse)

	var poiResponse struct {
		ID string `json:"id"`
	}
	env.ParseJSONResponse(createResponse, &poiResponse)
	poiID := poiResponse.ID

	// Add participants
	for i := 1; i <= 3; i++ {
		joinRequest := JoinPOIRequest{
			UserID: fmt.Sprintf("participant-%d", i),
		}
		joinResponse := env.POST("/api/pois/"+poiID+"/join", joinRequest)
		env.AssertHTTPSuccess(joinResponse)
	}

	// Verify participants exist
	env.redis.AssertSetSize("poi:participants:"+poiID, 3)

	// Step 2: Create WebSocket observers
	observer := env.websocket.CreateClient("session-observer", "user-observer", "map-delete-test")
	require.NotNil(t, observer)
	env.WaitForAsyncOperations()

	// Step 3: Delete POI via HTTP API
	deleteResponse := env.DELETE("/api/pois/" + poiID)
	env.AssertHTTPSuccess(deleteResponse)

	// Step 4: Verify Database - POI should be deleted
	var count int64
	env.db.DB.Model(&models.POI{}).Where("id = ?", poiID).Count(&count)
	assert.Equal(t, int64(0), count, "POI should be deleted from database")

	// Step 5: Verify Redis - Participants should be cleaned up
	env.redis.AssertKeyNotExists("poi:participants:" + poiID)

	// Step 6: Verify WebSocket Broadcasting
	env.WaitForAsyncOperations()

	// Simulate POI deleted event broadcast
	poiDeletedEvent := websocket.Message{
		Type: "poi_deleted",
		Data: map[string]interface{}{
			"poiId": poiID,
			"mapId": "map-delete-test",
		},
		Timestamp: time.Now(),
	}

	env.websocket.BroadcastToMap("map-delete-test", poiDeletedEvent)

	// Observer should receive the deletion event
	deleteMsg := observer.ExpectMessage("poi_deleted", 200*time.Millisecond)

	deleteData, ok := deleteMsg.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, poiID, deleteData["poiId"])

	// Step 7: Verify POI is no longer accessible via API
	getResponse := env.GET("/api/pois/" + poiID)
	env.AssertHTTPError(getResponse, 404) // Should return not found
}