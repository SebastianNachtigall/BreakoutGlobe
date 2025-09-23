package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// POIParticipants manages POI participant tracking using Redis sets
type POIParticipants struct {
	client *redis.Client
}

// NewPOIParticipants creates a new POI participants manager
func NewPOIParticipants(client *redis.Client) *POIParticipants {
	return &POIParticipants{
		client: client,
	}
}

// JoinPOI adds a session to a POI's participant set
func (pp *POIParticipants) JoinPOI(ctx context.Context, poiID, sessionID string) error {
	key := pp.getPOIParticipantsKey(poiID)
	
	err := pp.client.SAdd(ctx, key, sessionID).Err()
	if err != nil {
		return fmt.Errorf("failed to add participant to POI: %w", err)
	}
	
	return nil
}

// LeavePOI removes a session from a POI's participant set
func (pp *POIParticipants) LeavePOI(ctx context.Context, poiID, sessionID string) error {
	key := pp.getPOIParticipantsKey(poiID)
	
	err := pp.client.SRem(ctx, key, sessionID).Err()
	if err != nil {
		return fmt.Errorf("failed to remove participant from POI: %w", err)
	}
	
	return nil
}

// GetParticipantCount returns the number of participants in a POI
func (pp *POIParticipants) GetParticipantCount(ctx context.Context, poiID string) (int, error) {
	key := pp.getPOIParticipantsKey(poiID)
	
	count, err := pp.client.SCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get participant count: %w", err)
	}
	
	return int(count), nil
}

// GetParticipants returns all participant session IDs for a POI
func (pp *POIParticipants) GetParticipants(ctx context.Context, poiID string) ([]string, error) {
	key := pp.getPOIParticipantsKey(poiID)
	
	participants, err := pp.client.SMembers(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get participants: %w", err)
	}
	
	return participants, nil
}

// IsParticipant checks if a session is a participant in a POI
func (pp *POIParticipants) IsParticipant(ctx context.Context, poiID, sessionID string) (bool, error) {
	key := pp.getPOIParticipantsKey(poiID)
	
	isMember, err := pp.client.SIsMember(ctx, key, sessionID).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check participant membership: %w", err)
	}
	
	return isMember, nil
}

// CanJoinPOI checks if a POI has capacity for another participant
func (pp *POIParticipants) CanJoinPOI(ctx context.Context, poiID string, maxParticipants int) (bool, error) {
	count, err := pp.GetParticipantCount(ctx, poiID)
	if err != nil {
		return false, err
	}
	
	return count < maxParticipants, nil
}

// JoinPOIWithCapacityCheck adds a session to a POI only if there's capacity
func (pp *POIParticipants) JoinPOIWithCapacityCheck(ctx context.Context, poiID, sessionID string, maxParticipants int) error {
	key := pp.getPOIParticipantsKey(poiID)
	
	// Use a Lua script to atomically check capacity and add participant
	script := `
		local key = KEYS[1]
		local sessionID = ARGV[1]
		local maxParticipants = tonumber(ARGV[2])
		
		-- Check if already a member
		if redis.call('SISMEMBER', key, sessionID) == 1 then
			return 0  -- Already a member, no change
		end
		
		-- Check current count
		local currentCount = redis.call('SCARD', key)
		if currentCount >= maxParticipants then
			return -1  -- At capacity
		end
		
		-- Add participant
		redis.call('SADD', key, sessionID)
		return 1  -- Successfully added
	`
	
	result, err := pp.client.Eval(ctx, script, []string{key}, sessionID, maxParticipants).Result()
	if err != nil {
		return fmt.Errorf("failed to execute join with capacity check: %w", err)
	}
	
	switch result.(int64) {
	case -1:
		return fmt.Errorf("POI is at capacity (max: %d participants)", maxParticipants)
	case 0, 1:
		return nil // Success (already member or newly added)
	default:
		return fmt.Errorf("unexpected result from capacity check script: %v", result)
	}
}

// RemoveAllParticipants removes all participants from a POI
func (pp *POIParticipants) RemoveAllParticipants(ctx context.Context, poiID string) error {
	key := pp.getPOIParticipantsKey(poiID)
	
	// Delete the entire set
	err := pp.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to remove all participants: %w", err)
	}
	
	return nil
}

// RemoveParticipantFromAllPOIs removes a session from all POIs they're participating in
func (pp *POIParticipants) RemoveParticipantFromAllPOIs(ctx context.Context, sessionID string) error {
	// Get all POI participant keys
	pattern := "poi:participants:*"
	keys, err := pp.client.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get POI keys: %w", err)
	}
	
	if len(keys) == 0 {
		return nil // No POIs exist
	}
	
	// Use a pipeline for efficiency
	pipe := pp.client.Pipeline()
	
	// Add SREM commands for each POI
	for _, key := range keys {
		pipe.SRem(ctx, key, sessionID)
	}
	
	// Execute pipeline
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to execute removal pipeline: %w", err)
	}
	
	return nil
}

// GetPOIsForParticipant returns all POI IDs that a session is participating in
func (pp *POIParticipants) GetPOIsForParticipant(ctx context.Context, sessionID string) ([]string, error) {
	// Get all POI participant keys
	pattern := "poi:participants:*"
	keys, err := pp.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get POI keys: %w", err)
	}
	
	if len(keys) == 0 {
		return []string{}, nil // No POIs exist
	}
	
	// Use a pipeline to check membership in all POIs
	pipe := pp.client.Pipeline()
	
	// Add SISMEMBER commands for each POI
	for _, key := range keys {
		pipe.SIsMember(ctx, key, sessionID)
	}
	
	// Execute pipeline
	results, err := pipe.Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to execute membership check pipeline: %w", err)
	}
	
	// Collect POI IDs where the session is a member
	var participatingPOIs []string
	for i, result := range results {
		if cmd, ok := result.(*redis.BoolCmd); ok {
			if isMember, err := cmd.Result(); err == nil && isMember {
				// Extract POI ID from key
				poiID := pp.extractPOIIDFromKey(keys[i])
				if poiID != "" {
					participatingPOIs = append(participatingPOIs, poiID)
				}
			}
		}
	}
	
	return participatingPOIs, nil
}

// getPOIParticipantsKey generates the Redis key for POI participants
func (pp *POIParticipants) getPOIParticipantsKey(poiID string) string {
	return fmt.Sprintf("poi:participants:%s", poiID)
}

// extractPOIIDFromKey extracts the POI ID from a Redis key
func (pp *POIParticipants) extractPOIIDFromKey(key string) string {
	prefix := "poi:participants:"
	if len(key) > len(prefix) && key[:len(prefix)] == prefix {
		return key[len(prefix):]
	}
	return ""
}