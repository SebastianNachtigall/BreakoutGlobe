package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetWindowResetTime_EmptyWindow(t *testing.T) {
	// Setup
	mockRedis := new(MockRedisClient)
	config := RateLimiterConfig{
		DefaultLimits: map[ActionType]RateLimit{
			ActionUpdateAvatar: {Requests: 3, Window: time.Minute},
		},
		KeyPrefix: "test:",
	}

	rateLimiter := NewRateLimiter(mockRedis, config)
	userID := "test-user"
	action := ActionUpdateAvatar
	ctx := context.Background()

	// Mock empty window (no entries)
	mockRedis.On("ZRangeWithScores", ctx, "test:test-user:update_avatar", int64(0), int64(0)).Return([]interface{}{}, nil)

	// Test: No requests yet - should reset immediately
	resetTime, err := rateLimiter.GetWindowResetTime(ctx, userID, action)
	assert.NoError(t, err)
	assert.True(t, resetTime.Before(time.Now().Add(time.Second)), "Empty window should reset immediately")

	mockRedis.AssertExpectations(t)
}

func TestGetWindowResetTime_WithEntries(t *testing.T) {
	// Setup
	mockRedis := new(MockRedisClient)
	config := RateLimiterConfig{
		DefaultLimits: map[ActionType]RateLimit{
			ActionUpdateAvatar: {Requests: 3, Window: time.Minute},
		},
		KeyPrefix: "test:",
	}

	rateLimiter := NewRateLimiter(mockRedis, config)
	userID := "test-user"
	action := ActionUpdateAvatar
	ctx := context.Background()

	// Mock window with entries - oldest entry from 30 seconds ago
	oldestTime := time.Now().Add(-30 * time.Second)
	mockResult := []interface{}{
		map[string]interface{}{
			"Member": "request-1",
			"Score":  float64(oldestTime.UnixNano()),
		},
	}
	
	mockRedis.On("ZRangeWithScores", ctx, "test:test-user:update_avatar", int64(0), int64(0)).Return(mockResult, nil)

	// Test: Should return oldest entry + window duration
	resetTime, err := rateLimiter.GetWindowResetTime(ctx, userID, action)
	assert.NoError(t, err)
	
	expectedResetTime := oldestTime.Add(time.Minute)
	timeDiff := resetTime.Sub(expectedResetTime).Abs()
	assert.True(t, timeDiff < time.Second, 
		"Reset time should be oldest entry + window. Expected: %v, Got: %v, Diff: %v", 
		expectedResetTime, resetTime, timeDiff)

	mockRedis.AssertExpectations(t)
}