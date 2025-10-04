package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// EventType represents the type of real-time event
type EventType string

const (
	EventTypeAvatarMovement EventType = "avatar_movement"
	EventTypePOICreated     EventType = "poi_created"
	EventTypePOIUpdated     EventType = "poi_updated"
	EventTypePOIJoined      EventType = "poi_joined"
	EventTypePOILeft        EventType = "poi_left"
)

// LatLng represents a geographic coordinate
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// Event represents a generic real-time event
type Event struct {
	Type      EventType       `json:"type"`
	Data      json.RawMessage `json:"data"`
	Timestamp time.Time       `json:"timestamp"`
}

// AvatarMovementEvent represents an avatar position update
type AvatarMovementEvent struct {
	SessionID string    `json:"sessionId"`
	UserID    string    `json:"userId"`
	MapID     string    `json:"mapId"`
	Position  LatLng    `json:"position"`
	Timestamp time.Time `json:"timestamp"`
}

// POICreatedEvent represents a new POI being created
type POICreatedEvent struct {
	POIID           string    `json:"poiId"`
	MapID           string    `json:"mapId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	Position        LatLng    `json:"position"`
	CreatedBy       string    `json:"createdBy"`
	MaxParticipants int       `json:"maxParticipants"`
	ImageURL        string    `json:"imageUrl,omitempty"`
	ThumbnailURL    string    `json:"thumbnailUrl,omitempty"`
	CurrentCount    int       `json:"currentCount"`
	Timestamp       time.Time `json:"timestamp"`
}

// POIUpdatedEvent represents a POI being updated
type POIUpdatedEvent struct {
	POIID           string    `json:"poiId"`
	MapID           string    `json:"mapId"`
	Name            string    `json:"name"`
	Description     string    `json:"description"`
	MaxParticipants int       `json:"maxParticipants"`
	CurrentCount    int       `json:"currentCount"`
	Timestamp       time.Time `json:"timestamp"`
}

// POIParticipant represents a participant in a POI with avatar information
type POIParticipant struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// POIJoinedEvent represents a user joining a POI
type POIJoinedEvent struct {
	POIID        string    `json:"poiId"`
	MapID        string    `json:"mapId"`
	UserID       string    `json:"userId"`
	SessionID    string    `json:"sessionId"`
	CurrentCount int       `json:"currentCount"`
	Timestamp    time.Time `json:"timestamp"`
}

// POIJoinedEventWithParticipants represents a user joining a POI with participant information
type POIJoinedEventWithParticipants struct {
	POIID        string           `json:"poiId"`
	MapID        string           `json:"mapId"`
	UserID       string           `json:"userId"`
	SessionID    string           `json:"sessionId"`
	CurrentCount int              `json:"currentCount"`
	Participants []POIParticipant `json:"participants"`
	JoiningUser  POIParticipant   `json:"joiningUser"` // Info about the user who is joining
	Timestamp    time.Time        `json:"timestamp"`
}

// POILeftEvent represents a user leaving a POI
type POILeftEvent struct {
	POIID        string    `json:"poiId"`
	MapID        string    `json:"mapId"`
	UserID       string    `json:"userId"`
	SessionID    string    `json:"sessionId"`
	CurrentCount int       `json:"currentCount"`
	Timestamp    time.Time `json:"timestamp"`
}

// POILeftEventWithParticipants represents a user leaving a POI with participant information
type POILeftEventWithParticipants struct {
	POIID        string           `json:"poiId"`
	MapID        string           `json:"mapId"`
	UserID       string           `json:"userId"`
	SessionID    string           `json:"sessionId"`
	CurrentCount int              `json:"currentCount"`
	Participants []POIParticipant `json:"participants"`
	Timestamp    time.Time        `json:"timestamp"`
}

// PubSub manages Redis pub/sub operations for real-time events
type PubSub struct {
	client *redis.Client
}

// NewPubSub creates a new PubSub instance
func NewPubSub(client *redis.Client) *PubSub {
	return &PubSub{
		client: client,
	}
}

// PublishAvatarMovement publishes an avatar movement event
func (ps *PubSub) PublishAvatarMovement(ctx context.Context, event AvatarMovementEvent) error {
	return ps.publishEvent(ctx, EventTypeAvatarMovement, event, event.MapID, event.UserID)
}

// PublishPOICreated publishes a POI created event
func (ps *PubSub) PublishPOICreated(ctx context.Context, event POICreatedEvent) error {
	return ps.publishEvent(ctx, EventTypePOICreated, event, event.MapID, "")
}

// PublishPOIUpdated publishes a POI updated event
func (ps *PubSub) PublishPOIUpdated(ctx context.Context, event POIUpdatedEvent) error {
	return ps.publishEvent(ctx, EventTypePOIUpdated, event, event.MapID, "")
}

// PublishPOIJoined publishes a POI joined event
func (ps *PubSub) PublishPOIJoined(ctx context.Context, event POIJoinedEvent) error {
	return ps.publishEvent(ctx, EventTypePOIJoined, event, event.MapID, event.UserID)
}

// PublishPOIJoinedWithParticipants publishes a POI joined event with participant information
func (ps *PubSub) PublishPOIJoinedWithParticipants(ctx context.Context, event POIJoinedEventWithParticipants) error {
	return ps.publishEvent(ctx, EventTypePOIJoined, event, event.MapID, event.UserID)
}

// PublishPOILeft publishes a POI left event
func (ps *PubSub) PublishPOILeft(ctx context.Context, event POILeftEvent) error {
	return ps.publishEvent(ctx, EventTypePOILeft, event, event.MapID, event.UserID)
}

// PublishPOILeftWithParticipants publishes a POI left event with participant information
func (ps *PubSub) PublishPOILeftWithParticipants(ctx context.Context, event POILeftEventWithParticipants) error {
	return ps.publishEvent(ctx, EventTypePOILeft, event, event.MapID, event.UserID)
}

// publishEvent is a generic method to publish events to appropriate channels
func (ps *PubSub) publishEvent(ctx context.Context, eventType EventType, eventData interface{}, mapID, userID string) error {
	// Serialize event data
	data, err := json.Marshal(eventData)
	if err != nil {
		return fmt.Errorf("failed to marshal event data: %w", err)
	}

	// Create the event wrapper
	event := Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	}

	// Serialize the complete event
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to map-specific channel
	mapChannel := ps.getMapChannel(mapID)
	err = ps.client.Publish(ctx, mapChannel, eventJSON).Err()
	if err != nil {
		return fmt.Errorf("failed to publish to map channel %s: %w", mapChannel, err)
	}

	// If there's a user ID, also publish to user-specific channel
	if userID != "" {
		userChannel := ps.getUserChannel(userID)
		err = ps.client.Publish(ctx, userChannel, eventJSON).Err()
		if err != nil {
			return fmt.Errorf("failed to publish to user channel %s: %w", userChannel, err)
		}
	}

	return nil
}

// SubscribeToMapEvents subscribes to all events for a specific map
func (ps *PubSub) SubscribeToMapEvents(ctx context.Context, mapID string, eventsChan chan<- Event) error {
	channel := ps.getMapChannel(mapID)
	return ps.subscribeToChannel(ctx, channel, eventsChan)
}

// SubscribeToUserEvents subscribes to all events for a specific user
func (ps *PubSub) SubscribeToUserEvents(ctx context.Context, userID string, eventsChan chan<- Event) error {
	channel := ps.getUserChannel(userID)
	return ps.subscribeToChannel(ctx, channel, eventsChan)
}

// subscribeToChannel subscribes to a Redis channel and forwards messages to the events channel
func (ps *PubSub) subscribeToChannel(ctx context.Context, channel string, eventsChan chan<- Event) error {
	// Create a new pubsub instance for this subscription
	pubsub := ps.client.Subscribe(ctx, channel)
	defer pubsub.Close()

	// Get the channel for receiving messages
	msgChan := pubsub.Channel()

	// Process messages until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgChan:
			if !ok {
				return fmt.Errorf("subscription channel closed")
			}

			// Parse the event
			var event Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				// Log error but continue processing other messages
				continue
			}

			// Send event to the events channel
			select {
			case eventsChan <- event:
				// Event sent successfully
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Channel is full, skip this event to avoid blocking
				// In a production system, you might want to log this
			}
		}
	}
}

// SubscribeToMultipleChannels subscribes to multiple channels simultaneously
func (ps *PubSub) SubscribeToMultipleChannels(ctx context.Context, channels []string, eventsChan chan<- Event) error {
	if len(channels) == 0 {
		return fmt.Errorf("no channels provided")
	}

	// Create a new pubsub instance for this subscription
	pubsub := ps.client.Subscribe(ctx, channels...)
	defer pubsub.Close()

	// Get the channel for receiving messages
	msgChan := pubsub.Channel()

	// Process messages until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgChan:
			if !ok {
				return fmt.Errorf("subscription channel closed")
			}

			// Parse the event
			var event Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				// Log error but continue processing other messages
				continue
			}

			// Send event to the events channel
			select {
			case eventsChan <- event:
				// Event sent successfully
			case <-ctx.Done():
				return ctx.Err()
			default:
				// Channel is full, skip this event to avoid blocking
			}
		}
	}
}

// GetMapChannel returns the Redis channel name for a map
func (ps *PubSub) GetMapChannel(mapID string) string {
	return ps.getMapChannel(mapID)
}

// GetUserChannel returns the Redis channel name for a user
func (ps *PubSub) GetUserChannel(userID string) string {
	return ps.getUserChannel(userID)
}

// getMapChannel generates the Redis channel name for map events
func (ps *PubSub) getMapChannel(mapID string) string {
	return fmt.Sprintf("map:%s:events", mapID)
}

// getUserChannel generates the Redis channel name for user events
func (ps *PubSub) getUserChannel(userID string) string {
	return fmt.Sprintf("user:%s:events", userID)
}

// GetActiveChannels returns a list of channels that currently have subscribers
func (ps *PubSub) GetActiveChannels(ctx context.Context) ([]string, error) {
	// Use PUBSUB CHANNELS to get active channels
	channels, err := ps.client.PubSubChannels(ctx, "*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get active channels: %w", err)
	}
	return channels, nil
}

// GetChannelSubscriberCount returns the number of subscribers for a channel
func (ps *PubSub) GetChannelSubscriberCount(ctx context.Context, channel string) (int64, error) {
	counts, err := ps.client.PubSubNumSub(ctx, channel).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get subscriber count: %w", err)
	}
	
	if count, exists := counts[channel]; exists {
		return count, nil
	}
	return 0, nil
}

// FilterEventsByType filters events by their type
func FilterEventsByType(events []Event, eventType EventType) []Event {
	var filtered []Event
	for _, event := range events {
		if event.Type == eventType {
			filtered = append(filtered, event)
		}
	}
	return filtered
}

// ParseAvatarMovementEvent parses an Event into an AvatarMovementEvent
func ParseAvatarMovementEvent(event Event) (*AvatarMovementEvent, error) {
	if event.Type != EventTypeAvatarMovement {
		return nil, fmt.Errorf("event is not an avatar movement event")
	}

	var avatarEvent AvatarMovementEvent
	err := json.Unmarshal(event.Data, &avatarEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse avatar movement event: %w", err)
	}

	return &avatarEvent, nil
}

// ParsePOICreatedEvent parses an Event into a POICreatedEvent
func ParsePOICreatedEvent(event Event) (*POICreatedEvent, error) {
	if event.Type != EventTypePOICreated {
		return nil, fmt.Errorf("event is not a POI created event")
	}

	var poiEvent POICreatedEvent
	err := json.Unmarshal(event.Data, &poiEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POI created event: %w", err)
	}

	return &poiEvent, nil
}

// ParsePOIJoinedEvent parses an Event into a POIJoinedEvent
func ParsePOIJoinedEvent(event Event) (*POIJoinedEvent, error) {
	if event.Type != EventTypePOIJoined {
		return nil, fmt.Errorf("event is not a POI joined event")
	}

	var joinEvent POIJoinedEvent
	err := json.Unmarshal(event.Data, &joinEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POI joined event: %w", err)
	}

	return &joinEvent, nil
}

// ParsePOILeftEvent parses an Event into a POILeftEvent
func ParsePOILeftEvent(event Event) (*POILeftEvent, error) {
	if event.Type != EventTypePOILeft {
		return nil, fmt.Errorf("event is not a POI left event")
	}

	var leftEvent POILeftEvent
	err := json.Unmarshal(event.Data, &leftEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POI left event: %w", err)
	}

	return &leftEvent, nil
}

// ParsePOIUpdatedEvent parses an Event into a POIUpdatedEvent
func ParsePOIUpdatedEvent(event Event) (*POIUpdatedEvent, error) {
	if event.Type != EventTypePOIUpdated {
		return nil, fmt.Errorf("event is not a POI updated event")
	}

	var updatedEvent POIUpdatedEvent
	err := json.Unmarshal(event.Data, &updatedEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse POI updated event: %w", err)
	}

	return &updatedEvent, nil
}

// SubscribePOIEvents subscribes to all POI-related events across all maps and calls the callback for each event
func (ps *PubSub) SubscribePOIEvents(ctx context.Context, callback func(eventType string, data interface{})) error {
	// Subscribe to all map channels using a pattern
	// In Redis, we can use PSUBSCRIBE to subscribe to patterns
	pubsub := ps.client.PSubscribe(ctx, "map:*:events")
	defer pubsub.Close()

	// Get the channel for receiving messages
	msgChan := pubsub.Channel()

	// Process messages until context is cancelled
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgChan:
			if !ok {
				return fmt.Errorf("subscription channel closed")
			}

			// Parse the event
			var event Event
			err := json.Unmarshal([]byte(msg.Payload), &event)
			if err != nil {
				// Log error but continue processing other messages
				continue
			}

			// Only process POI-related events
			if event.Type == EventTypePOICreated || 
			   event.Type == EventTypePOIJoined || 
			   event.Type == EventTypePOILeft || 
			   event.Type == EventTypePOIUpdated {
				
				// Parse the event data based on type
				var eventData interface{}
				switch event.Type {
				case EventTypePOICreated:
					var poiEvent POICreatedEvent
					if err := json.Unmarshal(event.Data, &poiEvent); err == nil {
						eventData = map[string]interface{}{
							"poiId":           poiEvent.POIID,
							"mapId":           poiEvent.MapID,
							"name":            poiEvent.Name,
							"description":     poiEvent.Description,
							"position":        poiEvent.Position,
							"createdBy":       poiEvent.CreatedBy,
							"maxParticipants": poiEvent.MaxParticipants,
							"currentCount":    poiEvent.CurrentCount,
							"timestamp":       poiEvent.Timestamp,
						}
					}
				case EventTypePOIJoined:
					var joinEvent POIJoinedEvent
					if err := json.Unmarshal(event.Data, &joinEvent); err == nil {
						eventData = map[string]interface{}{
							"poiId":        joinEvent.POIID,
							"mapId":        joinEvent.MapID,
							"userId":       joinEvent.UserID,
							"sessionId":    joinEvent.SessionID,
							"currentCount": joinEvent.CurrentCount,
							"timestamp":    joinEvent.Timestamp,
						}
					}
				case EventTypePOILeft:
					var leftEvent POILeftEvent
					if err := json.Unmarshal(event.Data, &leftEvent); err == nil {
						eventData = map[string]interface{}{
							"poiId":        leftEvent.POIID,
							"mapId":        leftEvent.MapID,
							"userId":       leftEvent.UserID,
							"sessionId":    leftEvent.SessionID,
							"currentCount": leftEvent.CurrentCount,
							"timestamp":    leftEvent.Timestamp,
						}
					}
				case EventTypePOIUpdated:
					var updatedEvent POIUpdatedEvent
					if err := json.Unmarshal(event.Data, &updatedEvent); err == nil {
						eventData = map[string]interface{}{
							"poiId":           updatedEvent.POIID,
							"mapId":           updatedEvent.MapID,
							"name":            updatedEvent.Name,
							"description":     updatedEvent.Description,
							"maxParticipants": updatedEvent.MaxParticipants,
							"currentCount":    updatedEvent.CurrentCount,
							"timestamp":       updatedEvent.Timestamp,
						}
					}
				}

				// Call the callback with the parsed event
				if eventData != nil {
					callback(string(event.Type), eventData)
				}
			}
		}
	}
}