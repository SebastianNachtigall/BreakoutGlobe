package integration

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"
	"breakoutglobe/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPOIService_JoinPOI_WithParticipantInfo(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("Skipping integration test")
	}

	// Setup test environment
	testEnv := setupTestEnvironment(t)
	defer testEnv.cleanup()

	// Create test users with avatar information
	user1 := &models.User{
		ID:          generateTestID(),
		DisplayName: "Alice",
		AvatarURL:   stringPtr("/uploads/alice.jpg"),
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	user2 := &models.User{
		ID:          generateTestID(),
		DisplayName: "Bob",
		AvatarURL:   stringPtr("/uploads/bob.jpg"),
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	user3 := &models.User{
		ID:          generateTestID(),
		DisplayName: "Charlie",
		AvatarURL:   nil, // No avatar
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create users in database
	require.NoError(t, testEnv.db.Create(user1).Error)
	require.NoError(t, testEnv.db.Create(user2).Error)
	require.NoError(t, testEnv.db.Create(user3).Error)

	// Create a test map
	testMap := &models.Map{
		ID:        generateTestID(),
		Name:      "Test Map",
		CreatedBy: user1.ID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, testEnv.db.Create(testMap).Error)

	// Create a POI
	poi, err := testEnv.poiService.CreatePOI(
		context.Background(),
		testMap.ID,
		"Coffee Shop",
		"Great coffee",
		models.LatLng{Lat: 40.7128, Lng: -74.0060},
		user1.ID,
		10,
	)
	require.NoError(t, err)

	// Set up event capture
	events := make([]redis.Event, 0)
	eventsChan := make(chan redis.Event, 10)
	
	// Start capturing events in background
	go func() {
		for event := range eventsChan {
			events = append(events, event)
		}
	}()

	// Subscribe to map events
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	go func() {
		testEnv.pubsub.SubscribeToMapEvents(ctx, testMap.ID, eventsChan)
	}()

	// Give subscription time to establish
	time.Sleep(100 * time.Millisecond)

	// Join POI with users
	err = testEnv.poiService.JoinPOI(context.Background(), poi.ID, user1.ID)
	require.NoError(t, err)

	err = testEnv.poiService.JoinPOI(context.Background(), poi.ID, user2.ID)
	require.NoError(t, err)

	err = testEnv.poiService.JoinPOI(context.Background(), poi.ID, user3.ID)
	require.NoError(t, err)

	// Wait for events to be processed
	time.Sleep(200 * time.Millisecond)
	close(eventsChan)

	// Verify POI joined events were published
	joinEvents := filterPOIJoinedEvents(events)
	require.Len(t, joinEvents, 3, "Should have 3 POI joined events")

	// For now, just verify the basic events are published
	// We'll enhance this once we implement the participant info
	for _, event := range joinEvents {
		assert.Equal(t, poi.ID, event.POIID)
		assert.Equal(t, testMap.ID, event.MapID)
		assert.Contains(t, []string{user1.ID, user2.ID, user3.ID}, event.UserID)
	}
}

func TestPOIService_LeavePOI_WithParticipantInfo(t *testing.T) {
	if !isIntegrationTest() {
		t.Skip("Skipping integration test")
	}

	// Setup test environment
	testEnv := setupTestEnvironment(t)
	defer testEnv.cleanup()

	// Create test users
	user1 := &models.User{
		ID:          generateTestID(),
		DisplayName: "Alice",
		AvatarURL:   stringPtr("/uploads/alice.jpg"),
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	user2 := &models.User{
		ID:          generateTestID(),
		DisplayName: "Bob",
		AvatarURL:   stringPtr("/uploads/bob.jpg"),
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Create users in database
	require.NoError(t, testEnv.db.Create(user1).Error)
	require.NoError(t, testEnv.db.Create(user2).Error)

	// Create a test map
	testMap := &models.Map{
		ID:        generateTestID(),
		Name:      "Test Map",
		CreatedBy: user1.ID,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	require.NoError(t, testEnv.db.Create(testMap).Error)

	// Create POI
	poi, err := testEnv.poiService.CreatePOI(
		context.Background(),
		testMap.ID,
		"Coffee Shop",
		"Great coffee",
		models.LatLng{Lat: 40.7128, Lng: -74.0060},
		user1.ID,
		10,
	)
	require.NoError(t, err)

	// Join both users
	err = testEnv.poiService.JoinPOI(context.Background(), poi.ID, user1.ID)
	require.NoError(t, err)
	
	err = testEnv.poiService.JoinPOI(context.Background(), poi.ID, user2.ID)
	require.NoError(t, err)

	// Set up event capture for leave event
	events := make([]redis.Event, 0)
	eventsChan := make(chan redis.Event, 10)
	
	go func() {
		for event := range eventsChan {
			events = append(events, event)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	go func() {
		testEnv.pubsub.SubscribeToMapEvents(ctx, testMap.ID, eventsChan)
	}()

	time.Sleep(100 * time.Millisecond)

	// User1 leaves
	err = testEnv.poiService.LeavePOI(context.Background(), poi.ID, user1.ID)
	require.NoError(t, err)

	// Wait for events
	time.Sleep(200 * time.Millisecond)
	close(eventsChan)

	// Verify POI left event was published
	leftEvents := filterPOILeftEvents(events)
	require.Len(t, leftEvents, 1, "Should have 1 POI left event")

	leftEvent := leftEvents[0]
	assert.Equal(t, user1.ID, leftEvent.UserID)
	assert.Equal(t, poi.ID, leftEvent.POIID)
	assert.Equal(t, 1, leftEvent.CurrentCount) // Only user2 should remain
}

// Helper functions to filter events
func filterPOIJoinedEvents(events []redis.Event) []redis.POIJoinedEvent {
	var joinEvents []redis.POIJoinedEvent
	for _, event := range events {
		if event.Type == redis.EventTypePOIJoined {
			var joinEvent redis.POIJoinedEvent
			if err := json.Unmarshal(event.Data, &joinEvent); err == nil {
				joinEvents = append(joinEvents, joinEvent)
			}
		}
	}
	return joinEvents
}

func filterPOILeftEvents(events []redis.Event) []redis.POILeftEvent {
	var leftEvents []redis.POILeftEvent
	for _, event := range events {
		if event.Type == redis.EventTypePOILeft {
			var leftEvent redis.POILeftEvent
			if err := json.Unmarshal(event.Data, &leftEvent); err == nil {
				leftEvents = append(leftEvents, leftEvent)
			}
		}
	}
	return leftEvents
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}