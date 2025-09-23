package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	redislib "breakoutglobe/internal/redis"
	"breakoutglobe/internal/testdata"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRedisIntegration_POIParticipants tests POI participants functionality
func TestRedisIntegration_POIParticipants_AddRemove(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	// Setup test Redis
	testRedis := testdata.SetupRedis(t)
	participants := redislib.NewPOIParticipants(testRedis.Client())
	ctx := context.Background()

	// Test adding participant
	poiID := "poi-123"
	sessionID := "session-456"

	err := participants.JoinPOI(ctx, poiID, sessionID)
	require.NoError(t, err)

	// Verify participant was added to Redis set
	participantsKey := "poi:participants:" + poiID
	testRedis.AssertKeyExists(participantsKey)
	testRedis.AssertSetContains(participantsKey, sessionID)

	// Verify participant count
	count, err := participants.GetParticipantCount(ctx, poiID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	// Remove participant
	err = participants.LeavePOI(ctx, poiID, sessionID)
	require.NoError(t, err)

	// Verify participant was removed
	testRedis.AssertSetNotContains(participantsKey, sessionID)

	// Verify count is zero
	count, err = participants.GetParticipantCount(ctx, poiID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestRedisIntegration_SessionPresence_SetGet(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	// Setup test Redis
	testRedis := testdata.SetupRedis(t)
	presence := redislib.NewSessionPresence(testRedis.Client())
	ctx := context.Background()

	// Test setting session presence
	sessionID := "session-123"
	data := &redislib.SessionPresenceData{
		UserID: "user-123",
		MapID:  "map-456",
		AvatarPosition: models.LatLng{Lat: 40.7128, Lng: -74.0060},
		LastActive:     time.Now(),
	}

	err := presence.SetSessionPresence(ctx, sessionID, data, 5*time.Minute)
	require.NoError(t, err)

	// Verify presence was set in Redis
	presenceKey := "session:" + sessionID
	testRedis.AssertKeyExists(presenceKey)
}

// TestRedisIntegration_PubSub tests pub/sub functionality
func TestRedisIntegration_PubSub_PublishSubscribe(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	// Setup test Redis
	testRedis := testdata.SetupRedis(t)
	pubsub := redislib.NewPubSub(testRedis.Client())
	ctx := context.Background()

	mapID := "map-123"
	channel := "map:" + mapID

	// Subscribe to channel
	subscription := testRedis.Subscribe(channel)
	defer subscription.Close()

	// Wait for subscription to be established
	time.Sleep(10 * time.Millisecond)

	// Publish POI created event
	event := redislib.POICreatedEvent{
		POIID:           "poi-123",
		MapID:           mapID,
		Name:            "Test POI",
		Description:     "Test Description",
		Position:        redislib.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
		CurrentCount:    0,
		Timestamp:       time.Now(),
	}

	err := pubsub.PublishPOICreated(ctx, event)
	require.NoError(t, err)

	// Receive message
	msg, err := subscription.ReceiveTimeout(ctx, 100*time.Millisecond)
	require.NoError(t, err)

	switch m := msg.(type) {
	case *redis.Subscription:
		// This is the subscription confirmation, get the actual message
		msg, err = subscription.ReceiveTimeout(ctx, 100*time.Millisecond)
		require.NoError(t, err)
		actualMsg, ok := msg.(*redis.Message)
		require.True(t, ok)
		assert.Equal(t, channel, actualMsg.Channel)
		// Verify the message contains the event data
		assert.Contains(t, actualMsg.Payload, "poi-123")
	case *redis.Message:
		assert.Equal(t, channel, m.Channel)
		assert.Contains(t, m.Payload, "poi-123")
	default:
		t.Fatalf("Unexpected message type: %T", msg)
	}
}

func TestRedisIntegration_ConcurrentPOIParticipants(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	testRedis := testdata.SetupRedis(t)
	participants := redislib.NewPOIParticipants(testRedis.Client())
	ctx := context.Background()

	poiID := "concurrent-poi"
	const numSessions = 10

	// Concurrently add participants
	done := make(chan error, numSessions)
	for i := 0; i < numSessions; i++ {
		go func(index int) {
			sessionID := fmt.Sprintf("session-%d", index)
			done <- participants.JoinPOI(ctx, poiID, sessionID)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numSessions; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// Verify all participants were added
	count, err := participants.GetParticipantCount(ctx, poiID)
	require.NoError(t, err)
	assert.Equal(t, numSessions, count)

	// Verify Redis set has correct size
	participantsKey := "poi:participants:" + poiID
	testRedis.AssertSetSize(participantsKey, numSessions)
}