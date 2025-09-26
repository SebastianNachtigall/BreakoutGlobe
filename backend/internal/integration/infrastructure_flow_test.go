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

// TestInfrastructureFlow_DatabaseRedisWebSocketIntegration tests complete infrastructure integration
func TestInfrastructureFlow_DatabaseRedisWebSocketIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping infrastructure flow integration test in short mode")
	}

	// Setup all infrastructure components
	testDB := testdata.Setup(t)
	testRedis := testdata.SetupRedis(t)
	testWS := testdata.SetupWebSocket(t)

	// Create fixtures for valid foreign key relationships
	fixtures := testdata.NewTestFixtures(testDB)
	basicData := fixtures.SetupBasicTestData()

	ctx := context.Background()

	// Test 1: Database + Redis Integration
	t.Run("DatabaseRedisIntegration", func(t *testing.T) {
		t.Log("Starting Database + Redis integration test")
		// Create a session in database with valid foreign keys
		session := &models.Session{
			ID:        "test-session-1",
			UserID:    basicData.GetUser(0).ID, // Use valid user ID from fixtures
			MapID:     basicData.GetTestMap().ID, // Use valid map ID from fixtures
			AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedAt: time.Now(),
			LastActive: time.Now(),
			IsActive:  true,
		}

		// Save to database
		err := testDB.DB.Create(session).Error
		require.NoError(t, err)

		// Verify in database
		var count int64
		testDB.DB.Model(&models.Session{}).Where("id = ?", session.ID).Count(&count)
		assert.Equal(t, int64(1), count, "Session should exist in database")
		
		testDB.DB.Model(&models.Session{}).Where("user_id = ?", session.UserID).Count(&count)
		assert.Equal(t, int64(1), count, "Session with user_id should exist in database")

		// Set presence in Redis
		err = testRedis.Client().Set(ctx, "session:"+session.ID, "active", 30*time.Minute).Err()
		require.NoError(t, err)

		// Verify in Redis
		testRedis.AssertKeyExists("session:" + session.ID)

		// Create a POI in database
		poi := &models.POI{
			ID:              "test-poi-1",
			Name:            "Test Restaurant",
			Description:     "Great food",
			Position:        models.LatLng{Lat: 40.7505, Lng: -73.9934},
			CreatedBy:       session.UserID,
			MapID:           session.MapID,
			MaxParticipants: 5,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err = testDB.DB.Create(poi).Error
		require.NoError(t, err)

		// Verify POI in database
		var poiCount int64
		testDB.DB.Model(&models.POI{}).Where("id = ?", poi.ID).Count(&poiCount)
		assert.Equal(t, int64(1), poiCount, "POI should exist in database")

		// Add participant to POI in Redis
		err = testRedis.Client().SAdd(ctx, "poi:participants:"+poi.ID, session.ID).Err()
		require.NoError(t, err)

		// Verify participant in Redis
		isMember, err := testRedis.Client().SIsMember(ctx, "poi:participants:"+poi.ID, session.ID).Result()
		require.NoError(t, err)
		assert.True(t, isMember, "Session should be a participant of the POI")
	})

	// Test 2: WebSocket Integration
	t.Run("WebSocketIntegration", func(t *testing.T) {
		// Create WebSocket client
		client := testWS.CreateClient("ws-session-1", "ws-user-1", "ws-map-1")
		require.NotNil(t, client)

		// Wait for connection
		time.Sleep(50 * time.Millisecond)

		// Verify connection
		assert.True(t, client.IsConnected())
		assert.Equal(t, "ws-session-1", client.SessionID)
		assert.Equal(t, "ws-user-1", client.UserID)
		assert.Equal(t, "ws-map-1", client.MapID)

		// Send a test message
		testMessage := map[string]interface{}{
			"type": "test_message",
			"data": map[string]interface{}{
				"content": "Hello WebSocket",
			},
		}

		testMsg := websocket.Message{
			Type: "test_message",
			Data: testMessage["data"],
		}
		err := client.SendMessage(testMsg)
		require.NoError(t, err)

		// The message would be processed by the WebSocket handler
		// For this test, we just verify the infrastructure is working
	})

	// Test 3: Complete Flow Integration
	t.Run("CompleteFlowIntegration", func(t *testing.T) {
		// Scenario: User creates session, joins POI, moves avatar

		// Step 1: Create user session (Database + Redis)
		userSession := &models.Session{
			ID:        "flow-session-1",
			UserID:    basicData.GetUser(1).ID, // Use valid user ID from fixtures
			MapID:     basicData.GetTestMap().ID, // Use valid map ID from fixtures
			AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedAt: time.Now(),
			LastActive: time.Now(),
			IsActive:  true,
		}

		// Save session to database
		err := testDB.DB.Create(userSession).Error
		require.NoError(t, err)
		
		// Verify session in database
		var sessionCount int64
		testDB.DB.Model(&models.Session{}).Where("id = ?", userSession.ID).Count(&sessionCount)
		assert.Equal(t, int64(1), sessionCount, "Session should exist in database")

		// Set session presence in Redis
		err = testRedis.Client().Set(ctx, "session:"+userSession.ID, "active", 30*time.Minute).Err()
		require.NoError(t, err)
		testRedis.AssertKeyExists("session:" + userSession.ID)

		// Create WebSocket connection
		wsClient := testWS.CreateClient(userSession.ID, userSession.UserID, userSession.MapID)
		require.NotNil(t, wsClient)
		time.Sleep(50 * time.Millisecond)
		assert.True(t, wsClient.IsConnected())

		// Step 2: Create POI (Database)
		poi := &models.POI{
			ID:              "flow-poi-1",
			Name:            "Flow Restaurant",
			Description:     "Integration test POI",
			Position:        models.LatLng{Lat: 40.7505, Lng: -73.9934},
			CreatedBy:       userSession.UserID,
			MapID:           userSession.MapID,
			MaxParticipants: 10,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err = testDB.DB.Create(poi).Error
		require.NoError(t, err)
		
		// Verify POI in database
		var poiCount int64
		testDB.DB.Model(&models.POI{}).Where("id = ?", poi.ID).Count(&poiCount)
		assert.Equal(t, int64(1), poiCount, "POI should exist in database")

		// Step 3: User joins POI (Redis)
		err = testRedis.Client().SAdd(ctx, "poi:participants:"+poi.ID, userSession.ID).Err()
		require.NoError(t, err)

		// Verify participation
		isMember, err := testRedis.Client().SIsMember(ctx, "poi:participants:"+poi.ID, userSession.ID).Result()
		require.NoError(t, err)
		assert.True(t, isMember)

		// Step 4: Update avatar position (Database + Redis + WebSocket)
		newPosition := models.LatLng{Lat: 40.7589, Lng: -73.9851}

		// Update in database
		err = testDB.DB.Model(userSession).Update("avatar_pos_lat", newPosition.Lat).Error
		require.NoError(t, err)
		err = testDB.DB.Model(userSession).Update("avatar_pos_lng", newPosition.Lng).Error
		require.NoError(t, err)

		// Update presence in Redis
		err = testRedis.Client().HSet(ctx, "presence:"+userSession.ID, 
			"lat", newPosition.Lat, 
			"lng", newPosition.Lng,
			"lastUpdate", time.Now().Unix()).Err()
		require.NoError(t, err)

		// Broadcast via WebSocket (simulate)
		movementMessage := map[string]interface{}{
			"type": "avatar_movement",
			"data": map[string]interface{}{
				"sessionId": userSession.ID,
				"position":  newPosition,
			},
		}

		movementMsg := websocket.Message{
			Type: "avatar_movement",
			Data: movementMessage["data"],
		}
		err = wsClient.SendMessage(movementMsg)
		require.NoError(t, err)

		// Verify all systems are updated
		// Database
		var updatedSession models.Session
		err = testDB.DB.Where("id = ?", userSession.ID).First(&updatedSession).Error
		require.NoError(t, err)
		assert.Equal(t, newPosition.Lat, updatedSession.AvatarPos.Lat)
		assert.Equal(t, newPosition.Lng, updatedSession.AvatarPos.Lng)

		// Redis
		lat, err := testRedis.Client().HGet(ctx, "presence:"+userSession.ID, "lat").Float64()
		require.NoError(t, err)
		assert.Equal(t, newPosition.Lat, lat)

		// WebSocket (connection still active)
		assert.True(t, wsClient.IsConnected())

		t.Log("Complete infrastructure flow integration test passed!")
	})

	// Test 4: Multi-User Scenario
	t.Run("MultiUserScenario", func(t *testing.T) {
		const numUsers = 3
		sessions := make([]*models.Session, numUsers)
		clients := make([]*testdata.TestWSClient, numUsers)

		// Create multiple users
		for i := 0; i < numUsers; i++ {
			// Use existing users from fixtures, cycling through them
			userIndex := i % len(basicData.Users)
			session := &models.Session{
				ID:        fmt.Sprintf("multi-session-%d", i+1),
				UserID:    basicData.GetUser(userIndex).ID, // Use valid user ID from fixtures
				MapID:     basicData.GetTestMap().ID, // Use valid map ID from fixtures
				AvatarPos: models.LatLng{Lat: 40.7100 + float64(i)*0.001, Lng: -74.0050 + float64(i)*0.001},
				CreatedAt: time.Now(),
				LastActive: time.Now(),
				IsActive:  true,
			}

			// Save to database
			err := testDB.DB.Create(session).Error
			require.NoError(t, err)
			sessions[i] = session

			// Set Redis presence
			err = testRedis.Client().Set(ctx, "session:"+session.ID, "active", 30*time.Minute).Err()
			require.NoError(t, err)

			// Create WebSocket connection
			client := testWS.CreateClient(session.ID, session.UserID, session.MapID)
			require.NotNil(t, client)
			clients[i] = client
		}

		// Wait for all connections
		time.Sleep(100 * time.Millisecond)

		// Verify all users are properly set up
		for i := 0; i < numUsers; i++ {
			var sessionCount int64
			testDB.DB.Model(&models.Session{}).Where("id = ?", sessions[i].ID).Count(&sessionCount)
			assert.Equal(t, int64(1), sessionCount, "Session should exist in database")
			
			testRedis.AssertKeyExists("session:" + sessions[i].ID)
			assert.True(t, clients[i].IsConnected())
		}

		// Create a shared POI
		sharedPOI := &models.POI{
			ID:              "shared-poi",
			Name:            "Shared Meeting Point",
			Description:     "Multi-user test POI",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedBy:       sessions[0].UserID,
			MapID:           basicData.GetTestMap().ID, // Use valid map ID from fixtures
			MaxParticipants: 10,
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		err := testDB.DB.Create(sharedPOI).Error
		require.NoError(t, err)

		// All users join the POI
		for i := 0; i < numUsers; i++ {
			err = testRedis.Client().SAdd(ctx, "poi:participants:"+sharedPOI.ID, sessions[i].ID).Err()
			require.NoError(t, err)
		}

		// Verify all participants
		participantCount, err := testRedis.Client().SCard(ctx, "poi:participants:"+sharedPOI.ID).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(numUsers), participantCount)

		t.Log("Multi-user scenario integration test passed!")
	})
}

// TestInfrastructureFlow_ErrorHandling tests error handling across infrastructure layers
func TestInfrastructureFlow_ErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping infrastructure error handling test in short mode")
	}

	testDB := testdata.Setup(t)
	testRedis := testdata.SetupRedis(t)
	testWS := testdata.SetupWebSocket(t)

	// Create fixtures for valid foreign key relationships
	fixtures := testdata.NewTestFixtures(testDB)
	basicData := fixtures.SetupBasicTestData()

	ctx := context.Background()

	t.Run("DatabaseErrorHandling", func(t *testing.T) {
		// Try to create session with invalid foreign key
		invalidSession := &models.Session{
			ID:        "error-session-1",
			UserID:    "nonexistent-user", // Invalid user ID - should cause foreign key violation
			MapID:     basicData.GetTestMap().ID, // Valid map ID
			AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedAt: time.Now(),
			LastActive: time.Now(),
			IsActive:  true,
		}

		err := testDB.DB.Create(invalidSession).Error
		assert.Error(t, err, "Creating session with invalid user ID should fail")

		// Verify no data was persisted
		var count int64
		testDB.DB.Model(&models.Session{}).Where("user_id = ?", "nonexistent-user").Count(&count)
		assert.Equal(t, int64(0), count)
		
		// Test successful creation with valid foreign keys
		validSession := &models.Session{
			ID:        "error-session-2",
			UserID:    basicData.GetUser(0).ID, // Valid user ID from fixtures
			MapID:     basicData.GetTestMap().ID, // Valid map ID from fixtures
			AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
			CreatedAt: time.Now(),
			LastActive: time.Now(),
			IsActive:  true,
		}

		err = testDB.DB.Create(validSession).Error
		assert.NoError(t, err, "Creating session with valid foreign keys should succeed")
	})

	t.Run("RedisErrorHandling", func(t *testing.T) {
		// Try to access non-existent key
		result, err := testRedis.Client().Get(ctx, "non-existent-key").Result()
		assert.Error(t, err)
		assert.Empty(t, result)

		// This is expected behavior - Redis returns error for missing keys
		t.Log("Redis properly handles non-existent key access")
	})

	t.Run("WebSocketErrorHandling", func(t *testing.T) {
		// Create client and immediately close connection
		client := testWS.CreateClient("error-session", "error-user", "error-map")
		require.NotNil(t, client)

		// Close connection
		client.Close()

		// Try to send message to closed connection
		testMsg := websocket.Message{
			Type: "test",
			Data: nil,
		}
		err := client.SendMessage(testMsg)
		assert.Error(t, err, "Sending to closed WebSocket should fail")

		assert.False(t, client.IsConnected())
	})
}

// TestInfrastructureFlow_Performance tests performance across infrastructure layers
func TestInfrastructureFlow_Performance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping infrastructure performance test in short mode")
	}

	testDB := testdata.Setup(t)
	testRedis := testdata.SetupRedis(t)
	testWS := testdata.SetupWebSocket(t)

	// Create fixtures and additional users for performance testing
	fixtures := testdata.NewTestFixtures(testDB)
	basicData := fixtures.SetupBasicTestData()
	
	// Create additional users for performance testing
	const numOperations = 50
	for i := 0; i < numOperations; i++ {
		user := &models.User{
			ID:          fmt.Sprintf("perf-user-%d", i),
			DisplayName: fmt.Sprintf("Performance User %d", i),
			AccountType: models.AccountTypeGuest,
			Role:        models.UserRoleUser,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		err := testDB.DB.Create(user).Error
		require.NoError(t, err, "Should create performance test user %d", i)
	}

	ctx := context.Background()

	t.Run("ConcurrentOperations", func(t *testing.T) {
		// Test concurrent database operations
		t.Run("DatabaseConcurrency", func(t *testing.T) {
			done := make(chan error, numOperations)

			for i := 0; i < numOperations; i++ {
				go func(index int) {
					session := &models.Session{
						ID:        fmt.Sprintf("perf-session-%d", index),
						UserID:    fmt.Sprintf("perf-user-%d", index), // Use created performance users
						MapID:     basicData.GetTestMap().ID, // Use valid map from fixtures
						AvatarPos: models.LatLng{Lat: 40.7128, Lng: -74.0060},
						CreatedAt: time.Now(),
						LastActive: time.Now(),
						IsActive:  true,
					}

					done <- testDB.DB.Create(session).Error
				}(i)
			}

			// Wait for all operations
			successCount := 0
			for i := 0; i < numOperations; i++ {
				if err := <-done; err == nil {
					successCount++
				}
			}

			assert.GreaterOrEqual(t, successCount, numOperations-5, "Most database operations should succeed")
		})

		// Test concurrent Redis operations
		t.Run("RedisConcurrency", func(t *testing.T) {
			done := make(chan error, numOperations)

			for i := 0; i < numOperations; i++ {
				go func(index int) {
					key := fmt.Sprintf("perf-key-%d", index)
					done <- testRedis.Client().Set(ctx, key, "value", time.Minute).Err()
				}(i)
			}

			// Wait for all operations
			successCount := 0
			for i := 0; i < numOperations; i++ {
				if err := <-done; err == nil {
					successCount++
				}
			}

			assert.GreaterOrEqual(t, successCount, numOperations-2, "Most Redis operations should succeed")
		})

		// Test concurrent WebSocket connections
		t.Run("WebSocketConcurrency", func(t *testing.T) {
			clients := make([]*testdata.TestWSClient, numOperations)

			// Create connections concurrently
			for i := 0; i < numOperations; i++ {
				client := testWS.CreateClient(
					fmt.Sprintf("perf-ws-session-%d", i),
					fmt.Sprintf("perf-ws-user-%d", i),
					"perf-ws-map",
				)
				clients[i] = client
			}

			// Wait for connections
			time.Sleep(200 * time.Millisecond)

			// Count successful connections
			connectedCount := 0
			for _, client := range clients {
				if client != nil && client.IsConnected() {
					connectedCount++
				}
			}

			assert.GreaterOrEqual(t, connectedCount, numOperations-5, "Most WebSocket connections should succeed")

			// Clean up
			for _, client := range clients {
				if client != nil {
					client.Close()
				}
			}
		})
	})
}