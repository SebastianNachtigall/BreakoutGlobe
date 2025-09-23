package redis

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type PubSubTestSuite struct {
	suite.Suite
	client *redis.Client
	pubsub *PubSub
}

func (suite *PubSubTestSuite) SetupSuite() {
	// Skip integration tests in short mode
	if testing.Short() {
		suite.T().Skip("Skipping Redis integration test in short mode")
	}
	
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
	suite.pubsub = NewPubSub(client)
}

func (suite *PubSubTestSuite) SetupTest() {
	// Clean up Redis before each test
	ctx := context.Background()
	suite.client.FlushDB(ctx)
}

func (suite *PubSubTestSuite) TearDownSuite() {
	if suite.client != nil {
		suite.client.Close()
	}
}

func (suite *PubSubTestSuite) TestPublishAvatarMovement() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create avatar movement event
	event := AvatarMovementEvent{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     "map-789",
		Position: LatLng{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		Timestamp: time.Now(),
	}

	// Execute
	err := suite.pubsub.PublishAvatarMovement(ctx, event)

	// Assert
	suite.NoError(err)
}

func (suite *PubSubTestSuite) TestPublishPOICreated() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create POI created event
	event := POICreatedEvent{
		POIID:       "poi-123",
		MapID:       "map-456",
		Name:        "Test Meeting Room",
		Description: "A place for team meetings",
		Position: LatLng{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		CreatedBy:        "user-789",
		MaxParticipants:  10,
		CurrentCount:     0,
		Timestamp:        time.Now(),
	}

	// Execute
	err := suite.pubsub.PublishPOICreated(ctx, event)

	// Assert
	suite.NoError(err)
}

func (suite *PubSubTestSuite) TestPublishPOIJoined() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create POI joined event
	event := POIJoinedEvent{
		POIID:        "poi-123",
		MapID:        "map-456",
		UserID:       "user-789",
		SessionID:    "session-abc",
		CurrentCount: 3,
		Timestamp:    time.Now(),
	}

	// Execute
	err := suite.pubsub.PublishPOIJoined(ctx, event)

	// Assert
	suite.NoError(err)
}

func (suite *PubSubTestSuite) TestPublishPOILeft() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create POI left event
	event := POILeftEvent{
		POIID:        "poi-123",
		MapID:        "map-456",
		UserID:       "user-789",
		SessionID:    "session-abc",
		CurrentCount: 2,
		Timestamp:    time.Now(),
	}

	// Execute
	err := suite.pubsub.PublishPOILeft(ctx, event)

	// Assert
	suite.NoError(err)
}

func (suite *PubSubTestSuite) TestPublishPOIUpdated() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Create POI updated event
	event := POIUpdatedEvent{
		POIID:           "poi-123",
		MapID:           "map-456",
		Name:            "Updated Meeting Room",
		Description:     "An updated place for team meetings",
		MaxParticipants: 15,
		CurrentCount:    5,
		Timestamp:       time.Now(),
	}

	// Execute
	err := suite.pubsub.PublishPOIUpdated(ctx, event)

	// Assert
	suite.NoError(err)
}

func (suite *PubSubTestSuite) TestSubscribeToMapEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mapID := "map-123"
	eventsChan := make(chan Event, 10)
	errChan := make(chan error, 1)

	// Start subscription in goroutine
	go func() {
		err := suite.pubsub.SubscribeToMapEvents(ctx, mapID, eventsChan)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			errChan <- err
		}
	}()

	// Give subscription time to start
	time.Sleep(100 * time.Millisecond)

	// Publish an avatar movement event
	avatarEvent := AvatarMovementEvent{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     mapID,
		Position: LatLng{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		Timestamp: time.Now(),
	}

	err := suite.pubsub.PublishAvatarMovement(ctx, avatarEvent)
	suite.NoError(err)

	// Wait for event to be received
	select {
	case receivedEvent := <-eventsChan:
		suite.Equal(EventTypeAvatarMovement, receivedEvent.Type)
		
		// Unmarshal and verify the event data
		var receivedAvatarEvent AvatarMovementEvent
		err := json.Unmarshal(receivedEvent.Data, &receivedAvatarEvent)
		suite.NoError(err)
		suite.Equal(avatarEvent.SessionID, receivedAvatarEvent.SessionID)
		suite.Equal(avatarEvent.UserID, receivedAvatarEvent.UserID)
		suite.Equal(avatarEvent.MapID, receivedAvatarEvent.MapID)
		suite.Equal(avatarEvent.Position.Lat, receivedAvatarEvent.Position.Lat)
		suite.Equal(avatarEvent.Position.Lng, receivedAvatarEvent.Position.Lng)
	case err := <-errChan:
		suite.Fail("Subscription error: %v", err)
	case <-time.After(2 * time.Second):
		suite.Fail("Expected to receive avatar movement event")
	}
}

func (suite *PubSubTestSuite) TestSubscribeToMapEvents_POIEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mapID := "map-456"
	eventsChan := make(chan Event, 10)
	errChan := make(chan error, 1)

	// Start subscription in goroutine
	go func() {
		err := suite.pubsub.SubscribeToMapEvents(ctx, mapID, eventsChan)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			errChan <- err
		}
	}()

	// Give subscription time to start
	time.Sleep(100 * time.Millisecond)

	// Publish POI events
	poiCreatedEvent := POICreatedEvent{
		POIID:       "poi-123",
		MapID:       mapID,
		Name:        "Test Room",
		Description: "Test Description",
		Position: LatLng{
			Lat: 41.0,
			Lng: -75.0,
		},
		CreatedBy:        "user-789",
		MaxParticipants:  10,
		CurrentCount:     0,
		Timestamp:        time.Now(),
	}

	err := suite.pubsub.PublishPOICreated(ctx, poiCreatedEvent)
	suite.NoError(err)

	poiJoinedEvent := POIJoinedEvent{
		POIID:        "poi-123",
		MapID:        mapID,
		UserID:       "user-abc",
		SessionID:    "session-def",
		CurrentCount: 1,
		Timestamp:    time.Now(),
	}

	err = suite.pubsub.PublishPOIJoined(ctx, poiJoinedEvent)
	suite.NoError(err)

	// Receive and verify events
	eventsReceived := 0
	for eventsReceived < 2 {
		select {
		case receivedEvent := <-eventsChan:
			eventsReceived++
			
			if receivedEvent.Type == EventTypePOICreated {
				var receivedPOIEvent POICreatedEvent
				err := json.Unmarshal(receivedEvent.Data, &receivedPOIEvent)
				suite.NoError(err)
				suite.Equal(poiCreatedEvent.POIID, receivedPOIEvent.POIID)
				suite.Equal(poiCreatedEvent.Name, receivedPOIEvent.Name)
			} else if receivedEvent.Type == EventTypePOIJoined {
				var receivedJoinEvent POIJoinedEvent
				err := json.Unmarshal(receivedEvent.Data, &receivedJoinEvent)
				suite.NoError(err)
				suite.Equal(poiJoinedEvent.POIID, receivedJoinEvent.POIID)
				suite.Equal(poiJoinedEvent.UserID, receivedJoinEvent.UserID)
			}
		case err := <-errChan:
			suite.Fail("Subscription error: %v", err)
		case <-time.After(2 * time.Second):
			suite.Fail("Expected to receive POI events")
		}
	}
}

func (suite *PubSubTestSuite) TestSubscribeToUserEvents() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID := "user-123"
	eventsChan := make(chan Event, 10)
	errChan := make(chan error, 1)

	// Start subscription in goroutine
	go func() {
		err := suite.pubsub.SubscribeToUserEvents(ctx, userID, eventsChan)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			errChan <- err
		}
	}()

	// Give subscription time to start
	time.Sleep(100 * time.Millisecond)

	// Publish a POI joined event for this user
	poiJoinedEvent := POIJoinedEvent{
		POIID:        "poi-456",
		MapID:        "map-789",
		UserID:       userID,
		SessionID:    "session-abc",
		CurrentCount: 2,
		Timestamp:    time.Now(),
	}

	err := suite.pubsub.PublishPOIJoined(ctx, poiJoinedEvent)
	suite.NoError(err)

	// Wait for event to be received
	select {
	case receivedEvent := <-eventsChan:
		suite.Equal(EventTypePOIJoined, receivedEvent.Type)
		
		var receivedJoinEvent POIJoinedEvent
		err := json.Unmarshal(receivedEvent.Data, &receivedJoinEvent)
		suite.NoError(err)
		suite.Equal(poiJoinedEvent.UserID, receivedJoinEvent.UserID)
		suite.Equal(poiJoinedEvent.POIID, receivedJoinEvent.POIID)
	case err := <-errChan:
		suite.Fail("Subscription error: %v", err)
	case <-time.After(2 * time.Second):
		suite.Fail("Expected to receive POI joined event for user")
	}
}

func (suite *PubSubTestSuite) TestEventFiltering() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mapID := "map-123"
	otherMapID := "map-456"
	eventsChan := make(chan Event, 10)
	errChan := make(chan error, 1)

	// Start subscription for map-123 only
	go func() {
		err := suite.pubsub.SubscribeToMapEvents(ctx, mapID, eventsChan)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			errChan <- err
		}
	}()

	// Give subscription time to start
	time.Sleep(100 * time.Millisecond)

	// Publish event for the subscribed map
	avatarEvent1 := AvatarMovementEvent{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     mapID, // This should be received
		Position: LatLng{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		Timestamp: time.Now(),
	}

	// Publish event for a different map
	avatarEvent2 := AvatarMovementEvent{
		SessionID: "session-789",
		UserID:    "user-abc",
		MapID:     otherMapID, // This should NOT be received
		Position: LatLng{
			Lat: 41.0,
			Lng: -75.0,
		},
		Timestamp: time.Now(),
	}

	err := suite.pubsub.PublishAvatarMovement(ctx, avatarEvent1)
	suite.NoError(err)
	
	err = suite.pubsub.PublishAvatarMovement(ctx, avatarEvent2)
	suite.NoError(err)

	// Should only receive the event for the subscribed map
	select {
	case receivedEvent := <-eventsChan:
		suite.Equal(EventTypeAvatarMovement, receivedEvent.Type)
		
		var receivedAvatarEvent AvatarMovementEvent
		err := json.Unmarshal(receivedEvent.Data, &receivedAvatarEvent)
		suite.NoError(err)
		suite.Equal(mapID, receivedAvatarEvent.MapID) // Should be the subscribed map
		suite.Equal(avatarEvent1.SessionID, receivedAvatarEvent.SessionID)
	case err := <-errChan:
		suite.Fail("Subscription error: %v", err)
	case <-time.After(1 * time.Second):
		suite.Fail("Expected to receive avatar movement event for subscribed map")
	}

	// Should not receive any more events (the other map event should be filtered out)
	select {
	case <-eventsChan:
		suite.Fail("Should not receive events for other maps")
	case <-time.After(500 * time.Millisecond):
		// This is expected - no more events should be received
	}
}

func (suite *PubSubTestSuite) TestMessageSerialization() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Test complex event serialization
	event := POICreatedEvent{
		POIID:       "poi-123",
		MapID:       "map-456",
		Name:        "Test Meeting Room with Special Characters: àáâãäå",
		Description: "A place for team meetings\nwith multiple lines\tand tabs",
		Position: LatLng{
			Lat: 40.712800123456789, // High precision
			Lng: -74.006000987654321,
		},
		CreatedBy:        "user-789",
		MaxParticipants:  10,
		CurrentCount:     0,
		Timestamp:        time.Now().UTC(), // Use UTC for consistency
	}

	// Publish the event
	err := suite.pubsub.PublishPOICreated(ctx, event)
	suite.NoError(err)

	// Test that we can serialize and deserialize the event correctly
	eventData, err := json.Marshal(event)
	suite.NoError(err)

	var deserializedEvent POICreatedEvent
	err = json.Unmarshal(eventData, &deserializedEvent)
	suite.NoError(err)

	// Verify all fields are preserved
	suite.Equal(event.POIID, deserializedEvent.POIID)
	suite.Equal(event.MapID, deserializedEvent.MapID)
	suite.Equal(event.Name, deserializedEvent.Name)
	suite.Equal(event.Description, deserializedEvent.Description)
	suite.Equal(event.Position.Lat, deserializedEvent.Position.Lat)
	suite.Equal(event.Position.Lng, deserializedEvent.Position.Lng)
	suite.Equal(event.CreatedBy, deserializedEvent.CreatedBy)
	suite.Equal(event.MaxParticipants, deserializedEvent.MaxParticipants)
	suite.Equal(event.CurrentCount, deserializedEvent.CurrentCount)
	// Note: Time comparison might have slight precision differences, so we check within a reasonable range
	suite.WithinDuration(event.Timestamp, deserializedEvent.Timestamp, time.Millisecond)
}

func (suite *PubSubTestSuite) TestMultipleSubscribers() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	mapID := "map-123"
	eventsChan1 := make(chan Event, 10)
	eventsChan2 := make(chan Event, 10)
	errChan := make(chan error, 2)

	// Start two subscriptions for the same map
	go func() {
		err := suite.pubsub.SubscribeToMapEvents(ctx, mapID, eventsChan1)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			errChan <- err
		}
	}()

	go func() {
		err := suite.pubsub.SubscribeToMapEvents(ctx, mapID, eventsChan2)
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			errChan <- err
		}
	}()

	// Give subscriptions time to start
	time.Sleep(200 * time.Millisecond)

	// Publish an event
	avatarEvent := AvatarMovementEvent{
		SessionID: "session-123",
		UserID:    "user-456",
		MapID:     mapID,
		Position: LatLng{
			Lat: 40.7128,
			Lng: -74.0060,
		},
		Timestamp: time.Now(),
	}

	err := suite.pubsub.PublishAvatarMovement(ctx, avatarEvent)
	suite.NoError(err)

	// Both subscribers should receive the event
	receivedCount := 0
	for receivedCount < 2 {
		select {
		case event1 := <-eventsChan1:
			suite.Equal(EventTypeAvatarMovement, event1.Type)
			receivedCount++
		case event2 := <-eventsChan2:
			suite.Equal(EventTypeAvatarMovement, event2.Type)
			receivedCount++
		case err := <-errChan:
			suite.Fail("Subscription error: %v", err)
		case <-time.After(2 * time.Second):
			suite.Fail("Expected both subscribers to receive the event")
		}
	}
}

func TestPubSubTestSuite(t *testing.T) {
	suite.Run(t, new(PubSubTestSuite))
}