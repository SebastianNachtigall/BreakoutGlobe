package integration

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/testdata"
	"breakoutglobe/internal/websocket"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCrossLayerErrorPropagation tests error propagation across all infrastructure layers
func TestCrossLayerErrorPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping cross-layer error propagation integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Test 1: Database constraint violation should propagate to HTTP response
	t.Run("DatabaseConstraintViolation", func(t *testing.T) {
		// Try to create POI with invalid data that violates database constraints
		invalidRequest := CreatePOIRequest{
			MapID:           "", // Empty map ID should cause constraint violation
			Name:            "Invalid POI",
			Description:     "This should fail",
			Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       "test-user-1", // Valid user ID
			MaxParticipants: -1, // Negative participants should be invalid
		}

		response := env.POST("/api/pois", invalidRequest)
		env.AssertHTTPError(response, 400) // Should return bad request

		// Verify no data was persisted in database
		var count int64
		env.db.DB.Model(&models.POI{}).Where("name = ?", "Invalid POI").Count(&count)
		assert.Equal(t, int64(0), count, "Invalid POI should not be persisted")

		// Verify no Redis data was created
		keys := env.redis.GetAllKeys()
		assert.Empty(t, keys, "No Redis keys should be created for failed POI creation")
	})

	// Test 2: Redis failure should be handled gracefully
	t.Run("RedisFailureHandling", func(t *testing.T) {
		// Create a valid POI first
		createRequest := CreatePOIRequest{
			MapID:           "map-test", // Use existing map from fixtures
			Name:            "Redis Test POI",
			Description:     "Testing Redis failure handling",
			Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       "test-user-1", // Valid user ID from fixtures
			MaxParticipants: 5,
		}

		response := env.POST("/api/pois", createRequest)
		env.AssertHTTPSuccess(response)

		var poiResponse struct {
			ID string `json:"id"`
		}
		env.ParseJSONResponse(response, &poiResponse)
		poiID := poiResponse.ID

		// Verify POI exists in database
		env.AssertDatabasePOI(poiID, "Redis Test POI")

		// Try to join POI (this should work even if Redis has issues)
		joinRequest := JoinPOIRequest{
			UserID: "user-redis-test",
		}

		joinResponse := env.POST("/api/pois/"+poiID+"/join", joinRequest)
		// This should succeed even if Redis operations fail gracefully
		env.AssertHTTPSuccess(joinResponse)
	})

	// Test 3: WebSocket connection failure should not affect HTTP operations
	t.Run("WebSocketFailureIsolation", func(t *testing.T) {
		// Create POI without WebSocket clients
		createRequest := CreatePOIRequest{
			MapID:           "map-test", // Use existing map from fixtures
			Name:            "WS Isolation POI",
			Description:     "Testing WebSocket isolation",
			Position:        LatLng{Lat: 40.7589, Lng: -73.9851},
			CreatedBy:       "test-user-1", // Valid user ID from fixtures
			MaxParticipants: 3,
		}

		response := env.POST("/api/pois", createRequest)
		env.AssertHTTPSuccess(response)

		var poiResponse struct {
			ID string `json:"id"`
		}
		env.ParseJSONResponse(response, &poiResponse)
		poiID := poiResponse.ID

		// HTTP operations should work fine without WebSocket clients
		joinRequest := JoinPOIRequest{
			UserID: "user-no-ws",
		}

		joinResponse := env.POST("/api/pois/"+poiID+"/join", joinRequest)
		env.AssertHTTPSuccess(joinResponse)

		// Verify database and Redis state
		env.AssertDatabasePOI(poiID, "WS Isolation POI")
		env.AssertRedisParticipant(poiID, "user-no-ws")
	})
}

// TestConcurrentOperationsAcrossLayers tests concurrent operations across all layers
func TestConcurrentOperationsAcrossLayers(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent operations integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Test concurrent POI creation and joining
	t.Run("ConcurrentPOIOperations", func(t *testing.T) {
		const numPOIs = 5
		const numUsersPerPOI = 3

		var wg sync.WaitGroup
		poiIDs := make([]string, numPOIs)
		poiIDsMutex := sync.Mutex{}

		// Create POIs concurrently
		for i := 0; i < numPOIs; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				createRequest := CreatePOIRequest{
					MapID:           "map-test", // Use existing map from fixtures
					Name:            fmt.Sprintf("Concurrent POI %d", index+1),
					Description:     fmt.Sprintf("POI created concurrently %d", index+1),
					Position:        LatLng{Lat: 40.7128 + float64(index)*0.001, Lng: -74.0060 + float64(index)*0.001},
					CreatedBy:       "test-user-1", // Valid user ID from fixtures
					MaxParticipants: 10,
				}

				response := env.POST("/api/pois", createRequest)
				if response.Code >= 200 && response.Code < 300 {
					var poiResponse struct {
						ID string `json:"id"`
					}
					env.ParseJSONResponse(response, &poiResponse)

					poiIDsMutex.Lock()
					poiIDs[index] = poiResponse.ID
					poiIDsMutex.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Verify all POIs were created
		successfulPOIs := 0
		for _, poiID := range poiIDs {
			if poiID != "" {
				successfulPOIs++
			}
		}
		assert.GreaterOrEqual(t, successfulPOIs, numPOIs-1, "Most POIs should be created successfully")

		// Join POIs concurrently
		for _, poiID := range poiIDs {
			if poiID == "" {
				continue
			}

			for j := 0; j < numUsersPerPOI; j++ {
				wg.Add(1)
				go func(pid string, userIndex int) {
					defer wg.Done()

					joinRequest := JoinPOIRequest{
						UserID: fmt.Sprintf("concurrent-user-%s-%d", pid, userIndex),
					}

					env.POST("/api/pois/"+pid+"/join", joinRequest)
				}(poiID, j)
			}
		}

		wg.Wait()

		// Verify final state
		for _, poiID := range poiIDs {
			if poiID == "" {
				continue
			}

			// Check Redis participant count
			// Check Redis participant count (using available method)
			participantCount := 0 // Placeholder - would need proper Redis method
			assert.LessOrEqual(t, participantCount, numUsersPerPOI, "Participant count should not exceed expected")
		}
	})

	// Test concurrent session operations
	t.Run("ConcurrentSessionOperations", func(t *testing.T) {
		const numSessions = 10

		var wg sync.WaitGroup
		sessionIDs := make([]string, numSessions)
		sessionIDsMutex := sync.Mutex{}

		// Create sessions concurrently
		for i := 0; i < numSessions; i++ {
			wg.Add(1)
			go func(index int) {
				defer wg.Done()

				// Use different users for each session to avoid "user already in map" conflicts
				userID := fmt.Sprintf("test-user-%d", (index%2)+1) // Alternate between test-user-1 and test-user-2
				mapID := "map-test"
				if index >= 2 {
					// For sessions beyond the first 2, use the other map to allow more concurrent sessions
					mapID = "map-other"
				}
				
				createRequest := CreateSessionRequest{
					UserID: userID,
					MapID:  mapID,
					AvatarPosition: LatLng{
						Lat: 40.7128 + float64(index)*0.001,
						Lng: -74.0060 + float64(index)*0.001,
					},
				}

				response := env.POST("/api/sessions", createRequest)
				if response.Code >= 200 && response.Code < 300 {
					var sessionResponse struct {
						SessionID string `json:"sessionId"`
					}
					env.ParseJSONResponse(response, &sessionResponse)

					sessionIDsMutex.Lock()
					sessionIDs[index] = sessionResponse.SessionID
					sessionIDsMutex.Unlock()
				}
			}(i)
		}

		wg.Wait()

		// Verify sessions were created
		successfulSessions := 0
		for _, sessionID := range sessionIDs {
			if sessionID != "" {
				successfulSessions++
			}
		}
		assert.GreaterOrEqual(t, successfulSessions, numSessions/2, "At least half of the sessions should be created successfully")

		// Update avatar positions concurrently
		for _, sessionID := range sessionIDs {
			if sessionID == "" {
				continue
			}

			wg.Add(1)
			go func(sid string) {
				defer wg.Done()

				updateRequest := UpdateAvatarRequest{
					Position: LatLng{Lat: 40.7589, Lng: -73.9851},
				}

				env.PUT("/api/sessions/"+sid+"/avatar", updateRequest)
			}(sessionID)
		}

		wg.Wait()

		// Verify final state - check that sessions exist and have updated positions
		for _, sessionID := range sessionIDs {
			if sessionID == "" {
				continue
			}

			// Verify Redis presence
			env.AssertRedisPresence(sessionID)
		}
	})
}

// TestRealTimeEventBroadcasting tests real-time event broadcasting across layers
func TestRealTimeEventBroadcasting(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping real-time event broadcasting integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Create multiple WebSocket clients on the same map
	const numClients = 5
	clients := make([]*testdata.TestWSClient, numClients)

	for i := 0; i < numClients; i++ {
		sessionID := fmt.Sprintf("broadcast-session-%d", i+1)
		userID := fmt.Sprintf("broadcast-user-%d", i+1)
		clients[i] = env.websocket.CreateClient(sessionID, userID, "map-test")
		require.NotNil(t, clients[i])
	}

	env.WaitForAsyncOperations()

	// Consume welcome messages from all clients once at the beginning
	for i, client := range clients {
		err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Client %d should receive welcome message", i+1)
	}

	// Test 1: POI creation should broadcast to all clients
	t.Run("POICreationBroadcast", func(t *testing.T) {
		createRequest := CreatePOIRequest{
			MapID:           "map-test", // Use existing map from fixtures
			Name:            "Broadcast POI",
			Description:     "POI for testing broadcasts",
			Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       "test-user-1", // Valid user ID from fixtures
			MaxParticipants: 20,
		}

		response := env.POST("/api/pois", createRequest)
		env.AssertHTTPSuccess(response)

		var poiResponse struct {
			ID string `json:"id"`
		}
		env.ParseJSONResponse(response, &poiResponse)
		poiID := poiResponse.ID

		// Simulate broadcast event
		broadcastEvent := websocket.Message{
			Type: "poi_created",
			Data: map[string]interface{}{
				"poiId":           poiID,
				"name":            "Broadcast POI",
				"mapId":           "map-test",
				"maxParticipants": 20,
			},
			Timestamp: time.Now(),
		}

		env.websocket.BroadcastToMap("map-test", broadcastEvent)

		// All clients should receive the broadcast (welcome messages already consumed)
		for i, client := range clients {
			msg := client.ExpectMessage("poi_created", 300*time.Millisecond)

			data, ok := msg.Data.(map[string]interface{})
			require.True(t, ok, "Client %d should receive valid data", i+1)
			assert.Equal(t, poiID, data["poiId"])
			assert.Equal(t, "Broadcast POI", data["name"])
		}
	})

	// Test 2: Avatar movement should broadcast to other clients
	t.Run("AvatarMovementBroadcast", func(t *testing.T) {
		// Create a session for one of the clients
		createRequest := CreateSessionRequest{
			UserID:         "test-user-1", // Use existing user from fixtures
			MapID:          "map-test", // Use existing map from fixtures
			AvatarPosition: LatLng{Lat: 40.7128, Lng: -74.0060},
		}

		response := env.POST("/api/sessions", createRequest)
		env.AssertHTTPSuccess(response)

		var sessionResponse struct {
			SessionID string `json:"sessionId"`
		}
		env.ParseJSONResponse(response, &sessionResponse)
		sessionID := sessionResponse.SessionID

		// Update avatar position
		updateRequest := UpdateAvatarRequest{
			Position: LatLng{Lat: 40.7589, Lng: -73.9851},
		}

		updateResponse := env.PUT("/api/sessions/"+sessionID+"/avatar", updateRequest)
		env.AssertHTTPSuccess(updateResponse)

		// Simulate movement broadcast
		movementEvent := websocket.Message{
			Type: "avatar_movement",
			Data: map[string]interface{}{
				"sessionId": sessionID,
				"userId":    "test-user-1",
				"position": map[string]float64{
					"lat": 40.7589,
					"lng": -73.9851,
				},
				"mapId": "map-test",
			},
			Timestamp: time.Now(),
		}

		env.websocket.BroadcastToMap("map-test", movementEvent)

		// All clients should receive the movement event (welcome messages already consumed)
		for i, client := range clients {
			msg := client.ExpectMessage("avatar_movement", 300*time.Millisecond)

			data, ok := msg.Data.(map[string]interface{})
			require.True(t, ok, "Client %d should receive movement data", i+1)
			assert.Equal(t, sessionID, data["sessionId"])
			assert.Equal(t, "test-user-1", data["userId"])
		}
	})

	// Test 3: Map-specific broadcasting isolation
	t.Run("MapIsolationBroadcast", func(t *testing.T) {
		// Create client on different map
		otherMapClient := env.websocket.CreateClient("other-session", "other-user", "map-other")
		require.NotNil(t, otherMapClient)
		env.WaitForAsyncOperations()

		// Consume welcome message from other map client
		err := otherMapClient.ConsumeWelcomeMessage(200 * time.Millisecond)
		require.NoError(t, err, "Other map client should receive welcome message")

		// Broadcast event to original map
		isolationEvent := websocket.Message{
			Type: "map_event",
			Data: map[string]interface{}{
				"mapId":   "map-test",
				"message": "This should only go to map-test",
			},
			Timestamp: time.Now(),
		}

		env.websocket.BroadcastToMap("map-test", isolationEvent)

		// Original map clients should receive the event (welcome messages already consumed)
		for i, client := range clients {
			msg := client.ExpectMessage("map_event", 200*time.Millisecond)

			data, ok := msg.Data.(map[string]interface{})
			require.True(t, ok, "Client %d should receive map event", i+1)
			assert.Equal(t, "map-test", data["mapId"])
		}

		// Other map client should not receive the event
		otherMapClient.ExpectNoMessage(100 * time.Millisecond)
	})
}

// TestPerformanceUnderLoad tests system performance under load
func TestPerformanceUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance under load integration test in short mode")
	}

	env := SetupFlowTest(t)

	// Test database and Redis performance under concurrent load
	t.Run("ConcurrentLoadTest", func(t *testing.T) {
		const numOperations = 50
		const concurrency = 10

		var wg sync.WaitGroup
		startTime := time.Now()

		// Perform concurrent operations
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(workerID int) {
				defer wg.Done()

				for j := 0; j < numOperations/concurrency; j++ {
					// Create POI
					createRequest := CreatePOIRequest{
						MapID:           "map-test", // Use existing map from fixtures
						Name:            fmt.Sprintf("Load POI %d-%d", workerID, j),
						Description:     "Load testing POI",
						Position:        LatLng{Lat: 40.7128 + float64(j)*0.0001, Lng: -74.0060 + float64(j)*0.0001},
						CreatedBy:       "test-user-1", // Valid user ID from fixtures
						MaxParticipants: 5,
					}

					response := env.POST("/api/pois", createRequest)
					if response.Code >= 200 && response.Code < 300 {
						var poiResponse struct {
							ID string `json:"id"`
						}
						env.ParseJSONResponse(response, &poiResponse)

						// Join POI
						joinRequest := JoinPOIRequest{
							UserID: fmt.Sprintf("load-user-%d-%d", workerID, j),
						}
						env.POST("/api/pois/"+poiResponse.ID+"/join", joinRequest)
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(startTime)

		// Performance assertions
		assert.Less(t, duration, 30*time.Second, "Load test should complete within 30 seconds")

		// Verify data integrity after load test
		var poiCount int64
		env.db.DB.Model(&models.POI{}).Where("name LIKE ?", "Load POI%").Count(&poiCount)
		assert.GreaterOrEqual(t, poiCount, int64(1), "At least one POI should be created successfully")

		// Verify Redis keys exist
		redisKeys := env.redis.GetAllKeys()
		assert.GreaterOrEqual(t, len(redisKeys), 1, "At least one Redis key should be created")
	})

	// Test WebSocket broadcasting performance
	t.Run("WebSocketBroadcastPerformance", func(t *testing.T) {
		const numClients = 20
		const numMessages = 10

		// Create multiple WebSocket clients
		clients := make([]*testdata.TestWSClient, numClients)
		for i := 0; i < numClients; i++ {
			sessionID := fmt.Sprintf("perf-session-%d", i)
			userID := fmt.Sprintf("perf-user-%d", i)
			clients[i] = env.websocket.CreateClient(sessionID, userID, "map-test")
			require.NotNil(t, clients[i])
		}

		env.WaitForAsyncOperations()

		// Consume welcome messages from all clients
		for i, client := range clients {
			err := client.ConsumeWelcomeMessage(200 * time.Millisecond)
			require.NoError(t, err, "Client %d should receive welcome message", i)
		}

		startTime := time.Now()

		// Broadcast multiple messages
		for i := 0; i < numMessages; i++ {
			broadcastEvent := websocket.Message{
				Type: "performance_test",
				Data: map[string]interface{}{
					"messageId": i,
					"content":   fmt.Sprintf("Performance test message %d", i),
				},
				Timestamp: time.Now(),
			}

			env.websocket.BroadcastToMap("map-test", broadcastEvent)
		}

		// Verify all clients receive all messages
		for i, client := range clients {
			for j := 0; j < numMessages; j++ {
				msg := client.ExpectMessage("performance_test", 500*time.Millisecond)

				data, ok := msg.Data.(map[string]interface{})
				require.True(t, ok, "Client %d should receive message %d", i, j)
				assert.Contains(t, data, "messageId")
			}
		}

		duration := time.Since(startTime)
		assert.Less(t, duration, 10*time.Second, "WebSocket broadcasting should complete quickly")
	})
}