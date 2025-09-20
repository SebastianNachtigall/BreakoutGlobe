package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"breakoutglobe/internal/models"

	"github.com/redis/go-redis/v9"
)

// SessionPresenceData represents the data stored in Redis for session presence
type SessionPresenceData struct {
	UserID         string           `json:"userId"`
	MapID          string           `json:"mapId"`
	AvatarPosition models.LatLng    `json:"avatarPosition"`
	LastActive     time.Time        `json:"lastActive"`
	CurrentPOI     *string          `json:"currentPOI,omitempty"`
}

// SessionPresence manages session presence data in Redis
type SessionPresence struct {
	client *redis.Client
}

// NewSessionPresence creates a new SessionPresence instance
func NewSessionPresence(client *redis.Client) *SessionPresence {
	return &SessionPresence{
		client: client,
	}
}

// SetSessionPresence stores session presence data in Redis with TTL
func (sp *SessionPresence) SetSessionPresence(ctx context.Context, sessionID string, data *SessionPresenceData, ttl time.Duration) error {
	key := sp.getSessionKey(sessionID)
	
	// Serialize data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}
	
	// Store in Redis with TTL
	err = sp.client.Set(ctx, key, jsonData, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set session presence: %w", err)
	}
	
	return nil
}

// GetSessionPresence retrieves session presence data from Redis
func (sp *SessionPresence) GetSessionPresence(ctx context.Context, sessionID string) (*SessionPresenceData, error) {
	key := sp.getSessionKey(sessionID)
	
	// Get data from Redis
	jsonData, err := sp.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	
	// Deserialize JSON data
	var data SessionPresenceData
	err = json.Unmarshal([]byte(jsonData), &data)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}
	
	return &data, nil
}

// UpdateSessionActivity updates the last active timestamp for a session
func (sp *SessionPresence) UpdateSessionActivity(ctx context.Context, sessionID string, lastActive time.Time) error {
	// Get current session data
	data, err := sp.GetSessionPresence(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session data: %w", err)
	}
	
	// Update last active time
	data.LastActive = lastActive
	
	// Get current TTL to preserve it
	key := sp.getSessionKey(sessionID)
	ttl, err := sp.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}
	
	// If TTL is -1 (no expiry) or -2 (key doesn't exist), set default
	if ttl < 0 {
		ttl = 30 * time.Minute
	}
	
	// Store updated data
	return sp.SetSessionPresence(ctx, sessionID, data, ttl)
}

// UpdateAvatarPosition updates the avatar position for a session
func (sp *SessionPresence) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	// Get current session data
	data, err := sp.GetSessionPresence(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session data: %w", err)
	}
	
	// Update avatar position and last active time
	data.AvatarPosition = position
	data.LastActive = time.Now()
	
	// Get current TTL to preserve it
	key := sp.getSessionKey(sessionID)
	ttl, err := sp.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}
	
	// If TTL is -1 (no expiry) or -2 (key doesn't exist), set default
	if ttl < 0 {
		ttl = 30 * time.Minute
	}
	
	// Store updated data
	return sp.SetSessionPresence(ctx, sessionID, data, ttl)
}

// SetCurrentPOI sets or clears the current POI for a session
func (sp *SessionPresence) SetCurrentPOI(ctx context.Context, sessionID string, poiID *string) error {
	// Get current session data
	data, err := sp.GetSessionPresence(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session data: %w", err)
	}
	
	// Update current POI and last active time
	data.CurrentPOI = poiID
	data.LastActive = time.Now()
	
	// Get current TTL to preserve it
	key := sp.getSessionKey(sessionID)
	ttl, err := sp.client.TTL(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to get TTL: %w", err)
	}
	
	// If TTL is -1 (no expiry) or -2 (key doesn't exist), set default
	if ttl < 0 {
		ttl = 30 * time.Minute
	}
	
	// Store updated data
	return sp.SetSessionPresence(ctx, sessionID, data, ttl)
}

// RemoveSessionPresence removes session presence data from Redis
func (sp *SessionPresence) RemoveSessionPresence(ctx context.Context, sessionID string) error {
	key := sp.getSessionKey(sessionID)
	
	err := sp.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove session presence: %w", err)
	}
	
	return nil
}

// GetActiveSessionsForMap retrieves all active sessions for a specific map
func (sp *SessionPresence) GetActiveSessionsForMap(ctx context.Context, mapID string) ([]*SessionPresenceData, error) {
	// Get all session keys
	pattern := "session:*"
	keys, err := sp.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get session keys: %w", err)
	}
	
	var activeSessions []*SessionPresenceData
	
	// Check each session
	for _, key := range keys {
		jsonData, err := sp.client.Get(ctx, key).Result()
		if err != nil {
			// Skip if key expired or doesn't exist
			continue
		}
		
		var data SessionPresenceData
		err = json.Unmarshal([]byte(jsonData), &data)
		if err != nil {
			// Skip malformed data
			continue
		}
		
		// Only include sessions for the specified map
		if data.MapID == mapID {
			activeSessions = append(activeSessions, &data)
		}
	}
	
	return activeSessions, nil
}

// SessionHeartbeat extends the TTL for a session and updates last active time
func (sp *SessionPresence) SessionHeartbeat(ctx context.Context, sessionID string, ttl time.Duration) error {
	key := sp.getSessionKey(sessionID)
	
	// Check if session exists
	exists, err := sp.client.Exists(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to check session existence: %w", err)
	}
	
	if exists == 0 {
		return fmt.Errorf("session not found")
	}
	
	// Update last active time
	err = sp.UpdateSessionActivity(ctx, sessionID, time.Now())
	if err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	
	// Extend TTL
	err = sp.client.Expire(ctx, key, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to extend session TTL: %w", err)
	}
	
	return nil
}

// CleanupExpiredSessions removes expired sessions for a specific map
// This is a manual cleanup that can be called periodically
func (sp *SessionPresence) CleanupExpiredSessions(ctx context.Context, mapID string) (int, error) {
	// Get all session keys
	pattern := "session:*"
	keys, err := sp.client.Keys(ctx, pattern).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get session keys: %w", err)
	}
	
	cleanedCount := 0
	
	// Check each session
	for _, key := range keys {
		// Check if key still exists (not expired)
		exists, err := sp.client.Exists(ctx, key).Result()
		if err != nil {
			continue
		}
		
		if exists == 0 {
			// Key has expired, count it as cleaned
			cleanedCount++
			continue
		}
		
		// Get session data to check if it belongs to the map
		jsonData, err := sp.client.Get(ctx, key).Result()
		if err != nil {
			// Key expired between exists check and get, count as cleaned
			cleanedCount++
			continue
		}
		
		var data SessionPresenceData
		err = json.Unmarshal([]byte(jsonData), &data)
		if err != nil {
			// Malformed data, remove it
			sp.client.Del(ctx, key)
			cleanedCount++
			continue
		}
		
		// Only process sessions for the specified map
		if data.MapID != mapID {
			continue
		}
		
		// Check if session is considered expired based on last active time
		// (This is additional cleanup beyond Redis TTL)
		if time.Since(data.LastActive) > 30*time.Minute {
			err = sp.client.Del(ctx, key).Err()
			if err == nil {
				cleanedCount++
			}
		}
	}
	
	return cleanedCount, nil
}

// getSessionKey generates the Redis key for a session
func (sp *SessionPresence) getSessionKey(sessionID string) string {
	return fmt.Sprintf("session:%s", sessionID)
}

// GetSessionIDFromKey extracts session ID from Redis key
func (sp *SessionPresence) GetSessionIDFromKey(key string) string {
	return strings.TrimPrefix(key, "session:")
}