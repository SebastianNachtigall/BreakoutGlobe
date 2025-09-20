package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"breakoutglobe/internal/models"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type SessionPresenceTestSuite struct {
	suite.Suite
	client   *redis.Client
	presence *SessionPresence
}

func (suite *SessionPresenceTestSuite) SetupSuite() {
	// Connect to test Redis instance
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       1, // Use DB 1 for tests to avoid conflicts
	})

	// Test connection
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	suite.Require().NoError(err, "Redis connection failed - make sure Redis is running")

	suite.client = client
	suite.presence = NewSessionPresence(client)
}

func (suite *SessionPresenceTestSuite) SetupTest() {
	// Clean up Redis before each test
	ctx := context.Background()
	suite.client.FlushDB(ctx)
}

func (suite *SessionPresenceTestSuite) TearDownSuite() {
	if suite.client != nil {
		suite.client.Close()
	}
}

func (suite *SessionPresenceTestSuite) TestSetSessionPresence() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now(),
		CurrentPOI:      nil,
	}

	// Execute
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 30*time.Minute)

	// Assert
	suite.NoError(err)

	// Verify data was stored
	key := "session:session-123"
	result, err := suite.client.Get(ctx, key).Result()
	suite.NoError(err)

	var stored SessionPresenceData
	err = json.Unmarshal([]byte(result), &stored)
	suite.NoError(err)
	
	suite.Equal(sessionData.UserID, stored.UserID)
	suite.Equal(sessionData.MapID, stored.MapID)
	suite.Equal(sessionData.AvatarPosition.Lat, stored.AvatarPosition.Lat)
	suite.Equal(sessionData.AvatarPosition.Lng, stored.AvatarPosition.Lng)

	// Verify TTL was set
	ttl, err := suite.client.TTL(ctx, key).Result()
	suite.NoError(err)
	suite.True(ttl > 25*time.Minute) // Should be close to 30 minutes
}

func (suite *SessionPresenceTestSuite) TestGetSessionPresence() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now(),
		CurrentPOI:      nil,
	}

	// Set up test data
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 30*time.Minute)
	suite.Require().NoError(err)

	// Execute
	result, err := suite.presence.GetSessionPresence(ctx, "session-123")

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(sessionData.UserID, result.UserID)
	suite.Equal(sessionData.MapID, result.MapID)
	suite.Equal(sessionData.AvatarPosition.Lat, result.AvatarPosition.Lat)
	suite.Equal(sessionData.AvatarPosition.Lng, result.AvatarPosition.Lng)
}

func (suite *SessionPresenceTestSuite) TestGetSessionPresence_NotFound() {
	ctx := context.Background()

	// Execute
	result, err := suite.presence.GetSessionPresence(ctx, "non-existent-session")

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(redis.Nil, err)
}

func (suite *SessionPresenceTestSuite) TestUpdateSessionActivity() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now().Add(-5 * time.Minute),
		CurrentPOI:      nil,
	}

	// Set up test data
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 30*time.Minute)
	suite.Require().NoError(err)

	// Execute - update activity
	newActivity := time.Now()
	err = suite.presence.UpdateSessionActivity(ctx, "session-123", newActivity)

	// Assert
	suite.NoError(err)

	// Verify the last active time was updated
	result, err := suite.presence.GetSessionPresence(ctx, "session-123")
	suite.NoError(err)
	suite.True(result.LastActive.After(sessionData.LastActive))
	suite.WithinDuration(newActivity, result.LastActive, time.Second)
}

func (suite *SessionPresenceTestSuite) TestUpdateAvatarPosition() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now(),
		CurrentPOI:      nil,
	}

	// Set up test data
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 30*time.Minute)
	suite.Require().NoError(err)

	// Execute - update avatar position
	newPosition := models.LatLng{Lat: 41.0000, Lng: -75.0000}
	err = suite.presence.UpdateAvatarPosition(ctx, "session-123", newPosition)

	// Assert
	suite.NoError(err)

	// Verify the position was updated
	result, err := suite.presence.GetSessionPresence(ctx, "session-123")
	suite.NoError(err)
	suite.Equal(newPosition.Lat, result.AvatarPosition.Lat)
	suite.Equal(newPosition.Lng, result.AvatarPosition.Lng)
	
	// Verify last active was also updated
	suite.True(result.LastActive.After(sessionData.LastActive))
}

func (suite *SessionPresenceTestSuite) TestSetCurrentPOI() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now(),
		CurrentPOI:      nil,
	}

	// Set up test data
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 30*time.Minute)
	suite.Require().NoError(err)

	// Execute - set current POI
	poiID := "poi-789"
	err = suite.presence.SetCurrentPOI(ctx, "session-123", &poiID)

	// Assert
	suite.NoError(err)

	// Verify the POI was set
	result, err := suite.presence.GetSessionPresence(ctx, "session-123")
	suite.NoError(err)
	suite.NotNil(result.CurrentPOI)
	suite.Equal(poiID, *result.CurrentPOI)

	// Test clearing POI
	err = suite.presence.SetCurrentPOI(ctx, "session-123", nil)
	suite.NoError(err)

	result, err = suite.presence.GetSessionPresence(ctx, "session-123")
	suite.NoError(err)
	suite.Nil(result.CurrentPOI)
}

func (suite *SessionPresenceTestSuite) TestRemoveSessionPresence() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now(),
		CurrentPOI:      nil,
	}

	// Set up test data
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 30*time.Minute)
	suite.Require().NoError(err)

	// Verify it exists
	_, err = suite.presence.GetSessionPresence(ctx, "session-123")
	suite.NoError(err)

	// Execute - remove session
	err = suite.presence.RemoveSessionPresence(ctx, "session-123")

	// Assert
	suite.NoError(err)

	// Verify it's gone
	_, err = suite.presence.GetSessionPresence(ctx, "session-123")
	suite.Error(err)
	suite.Equal(redis.Nil, err)
}

func (suite *SessionPresenceTestSuite) TestGetActiveSessionsForMap() {
	ctx := context.Background()
	
	// Create multiple sessions for the same map
	sessions := []struct {
		sessionID string
		userID    string
	}{
		{"session-1", "user-1"},
		{"session-2", "user-2"},
		{"session-3", "user-3"},
	}

	mapID := "map-456"
	for _, s := range sessions {
		sessionData := &SessionPresenceData{
			UserID:          s.userID,
			MapID:           mapID,
			AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
			LastActive:      time.Now(),
			CurrentPOI:      nil,
		}
		err := suite.presence.SetSessionPresence(ctx, s.sessionID, sessionData, 30*time.Minute)
		suite.Require().NoError(err)
	}

	// Add a session for a different map
	otherMapData := &SessionPresenceData{
		UserID:          "user-other",
		MapID:           "map-other",
		AvatarPosition:  models.LatLng{Lat: 41.0000, Lng: -75.0000},
		LastActive:      time.Now(),
		CurrentPOI:      nil,
	}
	err := suite.presence.SetSessionPresence(ctx, "session-other", otherMapData, 30*time.Minute)
	suite.Require().NoError(err)

	// Execute
	activeSessions, err := suite.presence.GetActiveSessionsForMap(ctx, mapID)

	// Assert
	suite.NoError(err)
	suite.Len(activeSessions, 3)

	// Verify all sessions belong to the correct map
	userIDs := make([]string, len(activeSessions))
	for i, session := range activeSessions {
		suite.Equal(mapID, session.MapID)
		userIDs[i] = session.UserID
	}

	suite.Contains(userIDs, "user-1")
	suite.Contains(userIDs, "user-2")
	suite.Contains(userIDs, "user-3")
	suite.NotContains(userIDs, "user-other")
}

func (suite *SessionPresenceTestSuite) TestSessionHeartbeat() {
	ctx := context.Background()
	
	sessionData := &SessionPresenceData{
		UserID:          "user-123",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now().Add(-10 * time.Minute),
		CurrentPOI:      nil,
	}

	// Set up test data with short TTL
	err := suite.presence.SetSessionPresence(ctx, "session-123", sessionData, 1*time.Minute)
	suite.Require().NoError(err)

	// Execute heartbeat
	err = suite.presence.SessionHeartbeat(ctx, "session-123", 5*time.Minute)

	// Assert
	suite.NoError(err)

	// Verify TTL was extended
	key := "session:session-123"
	ttl, err := suite.client.TTL(ctx, key).Result()
	suite.NoError(err)
	suite.True(ttl > 4*time.Minute) // Should be close to 5 minutes

	// Verify last active was updated
	result, err := suite.presence.GetSessionPresence(ctx, "session-123")
	suite.NoError(err)
	suite.True(result.LastActive.After(sessionData.LastActive))
}

func (suite *SessionPresenceTestSuite) TestCleanupExpiredSessions() {
	ctx := context.Background()
	
	// Create sessions with different last active times
	activeSession := &SessionPresenceData{
		UserID:          "user-active",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:      time.Now(), // Recent activity
		CurrentPOI:      nil,
	}

	expiredSession := &SessionPresenceData{
		UserID:          "user-expired",
		MapID:           "map-456",
		AvatarPosition:  models.LatLng{Lat: 41.0000, Lng: -75.0000},
		LastActive:      time.Now().Add(-35 * time.Minute), // Older than 30 min
		CurrentPOI:      nil,
	}

	// Set both sessions with normal TTL
	err := suite.presence.SetSessionPresence(ctx, "session-active", activeSession, 30*time.Minute)
	suite.Require().NoError(err)

	err = suite.presence.SetSessionPresence(ctx, "session-expired", expiredSession, 30*time.Minute)
	suite.Require().NoError(err)

	// Verify both sessions exist before cleanup
	_, err = suite.presence.GetSessionPresence(ctx, "session-active")
	suite.NoError(err)
	_, err = suite.presence.GetSessionPresence(ctx, "session-expired")
	suite.NoError(err)

	// Execute cleanup - this should remove sessions with LastActive > 30 minutes ago
	cleanedCount, err := suite.presence.CleanupExpiredSessions(ctx, "map-456")

	// Assert
	suite.NoError(err)
	suite.Equal(1, cleanedCount) // Should have cleaned up 1 expired session

	// Verify active session still exists
	_, err = suite.presence.GetSessionPresence(ctx, "session-active")
	suite.NoError(err)

	// Verify expired session is gone
	_, err = suite.presence.GetSessionPresence(ctx, "session-expired")
	suite.Error(err)
	suite.Equal(redis.Nil, err)
}

func TestSessionPresenceTestSuite(t *testing.T) {
	suite.Run(t, new(SessionPresenceTestSuite))
}