package services

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ActionType represents different types of actions that can be rate limited
type ActionType string

const (
	ActionCreateSession ActionType = "create_session"
	ActionUpdateAvatar  ActionType = "update_avatar"
	ActionCreatePOI     ActionType = "create_poi"
	ActionJoinPOI       ActionType = "join_poi"
	ActionLeavePOI      ActionType = "leave_poi"
	ActionUpdatePOI     ActionType = "update_poi"
	ActionDeletePOI     ActionType = "delete_poi"
)

// RateLimit defines the limit configuration for an action
type RateLimit struct {
	Requests int           `json:"requests"` // Number of requests allowed
	Window   time.Duration `json:"window"`   // Time window for the limit
}

// RateLimiterConfig holds the configuration for the rate limiter
type RateLimiterConfig struct {
	DefaultLimits map[ActionType]RateLimit `json:"defaultLimits"`
	KeyPrefix     string                   `json:"keyPrefix"`
}

// RedisClientInterface defines the interface for Redis operations needed by rate limiter
type RedisClientInterface interface {
	ZAdd(ctx context.Context, key string, score float64, member string) error
	ZRemRangeByScore(ctx context.Context, key string, min, max string) error
	ZCard(ctx context.Context, key string) (int64, error)
	Expire(ctx context.Context, key string, expiration time.Duration) error
	Pipeline() PipelineInterface
}

// PipelineInterface defines the interface for Redis pipeline operations
type PipelineInterface interface {
	ZAdd(ctx context.Context, key string, score float64, member string)
	ZRemRangeByScore(ctx context.Context, key string, min, max string)
	ZCard(ctx context.Context, key string)
	Expire(ctx context.Context, key string, expiration time.Duration)
	Exec(ctx context.Context) ([]interface{}, error)
}

// RateLimiterInterface defines the interface for rate limiting operations
type RateLimiterInterface interface {
	// IsAllowed checks if a user is allowed to perform an action
	IsAllowed(ctx context.Context, userID string, action ActionType) (bool, error)
	
	// GetRemainingRequests returns the number of remaining requests for a user action
	GetRemainingRequests(ctx context.Context, userID string, action ActionType) (int, error)
	
	// GetWindowResetTime returns when the current rate limit window will reset
	GetWindowResetTime(ctx context.Context, userID string, action ActionType) (time.Time, error)
	
	// SetCustomLimit sets a custom rate limit for a specific user and action
	SetCustomLimit(userID string, action ActionType, limit RateLimit)
	
	// ClearUserLimits clears all rate limit data for a user
	ClearUserLimits(ctx context.Context, userID string) error
	
	// GetUserStats returns rate limiting statistics for a user
	GetUserStats(ctx context.Context, userID string) (*UserRateLimitStats, error)
	
	// CheckRateLimit is a helper function that checks rate limit and returns appropriate error
	CheckRateLimit(ctx context.Context, userID string, action ActionType) error
	
	// GetRateLimitHeaders returns HTTP headers for rate limiting information
	GetRateLimitHeaders(ctx context.Context, userID string, action ActionType) (map[string]string, error)
}

// RateLimiter implements sliding window rate limiting using Redis sorted sets
type RateLimiter struct {
	redis       RedisClientInterface
	config      RateLimiterConfig
	customLimits map[string]map[ActionType]RateLimit // userID -> action -> limit
	mutex       sync.RWMutex
}

// UserRateLimitStats contains rate limiting statistics for a user
type UserRateLimitStats struct {
	UserID      string                        `json:"userId"`
	ActionStats map[ActionType]ActionStats    `json:"actionStats"`
	GeneratedAt time.Time                     `json:"generatedAt"`
}

// ActionStats contains statistics for a specific action
type ActionStats struct {
	Action       ActionType    `json:"action"`
	CurrentCount int           `json:"currentCount"`
	Limit        int           `json:"limit"`
	WindowStart  time.Time     `json:"windowStart"`
	WindowEnd    time.Time     `json:"windowEnd"`
	Remaining    int           `json:"remaining"`
}

// NewRateLimiter creates a new rate limiter instance
func NewRateLimiter(redis RedisClientInterface, config RateLimiterConfig) *RateLimiter {
	return &RateLimiter{
		redis:        redis,
		config:       config,
		customLimits: make(map[string]map[ActionType]RateLimit),
	}
}

// IsAllowed checks if a user is allowed to perform an action using sliding window algorithm
func (rl *RateLimiter) IsAllowed(ctx context.Context, userID string, action ActionType) (bool, error) {
	limit := rl.getLimit(userID, action)
	key := rl.getKey(userID, action)
	now := time.Now()
	windowStart := now.Add(-limit.Window)
	
	// Use pipeline for atomic operations
	pipe := rl.redis.Pipeline()
	
	// Remove expired entries
	pipe.ZRemRangeByScore(ctx, key, "0", strconv.FormatInt(windowStart.UnixNano(), 10))
	
	// Add current request
	requestID := uuid.New().String()
	pipe.ZAdd(ctx, key, float64(now.UnixNano()), requestID)
	
	// Count current requests
	pipe.ZCard(ctx, key)
	
	// Set expiration for cleanup
	pipe.Expire(ctx, key, limit.Window*2)
	
	// Execute pipeline
	results, err := pipe.Exec(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to execute rate limit check: %w", err)
	}
	
	// Get count from results (3rd operation - ZCard)
	if len(results) < 3 {
		return false, fmt.Errorf("unexpected pipeline results length: %d", len(results))
	}
	
	count, ok := results[2].(int64)
	if !ok {
		return false, fmt.Errorf("unexpected count type: %T", results[2])
	}
	
	return count <= int64(limit.Requests), nil
}

// GetRemainingRequests returns the number of remaining requests for a user action
func (rl *RateLimiter) GetRemainingRequests(ctx context.Context, userID string, action ActionType) (int, error) {
	limit := rl.getLimit(userID, action)
	key := rl.getKey(userID, action)
	
	count, err := rl.redis.ZCard(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}
	
	remaining := limit.Requests - int(count)
	if remaining < 0 {
		remaining = 0
	}
	
	return remaining, nil
}

// GetWindowResetTime returns when the current rate limit window will reset
func (rl *RateLimiter) GetWindowResetTime(ctx context.Context, userID string, action ActionType) (time.Time, error) {
	limit := rl.getLimit(userID, action)
	now := time.Now()
	
	// The window resets after the window duration from now
	resetTime := now.Add(limit.Window)
	
	return resetTime, nil
}

// SetCustomLimit sets a custom rate limit for a specific user and action
func (rl *RateLimiter) SetCustomLimit(userID string, action ActionType, limit RateLimit) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()
	
	if rl.customLimits[userID] == nil {
		rl.customLimits[userID] = make(map[ActionType]RateLimit)
	}
	
	rl.customLimits[userID][action] = limit
}

// ClearUserLimits clears all rate limit data for a user
func (rl *RateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	// Clear custom limits
	rl.mutex.Lock()
	delete(rl.customLimits, userID)
	rl.mutex.Unlock()
	
	// Clear Redis data for all actions
	pipe := rl.redis.Pipeline()
	
	for action := range rl.config.DefaultLimits {
		key := rl.getKey(userID, action)
		pipe.ZRemRangeByScore(ctx, key, "0", "+inf")
	}
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("failed to clear user limits: %w", err)
	}
	
	return nil
}

// GetUserStats returns rate limiting statistics for a user
func (rl *RateLimiter) GetUserStats(ctx context.Context, userID string) (*UserRateLimitStats, error) {
	stats := &UserRateLimitStats{
		UserID:      userID,
		ActionStats: make(map[ActionType]ActionStats),
		GeneratedAt: time.Now(),
	}
	
	for action := range rl.config.DefaultLimits {
		limit := rl.getLimit(userID, action)
		key := rl.getKey(userID, action)
		
		count, err := rl.redis.ZCard(ctx, key)
		if err != nil {
			return nil, fmt.Errorf("failed to get count for action %s: %w", action, err)
		}
		
		now := time.Now()
		windowStart := now.Add(-limit.Window)
		windowEnd := now.Add(limit.Window)
		remaining := limit.Requests - int(count)
		if remaining < 0 {
			remaining = 0
		}
		
		stats.ActionStats[action] = ActionStats{
			Action:       action,
			CurrentCount: int(count),
			Limit:        limit.Requests,
			WindowStart:  windowStart,
			WindowEnd:    windowEnd,
			Remaining:    remaining,
		}
	}
	
	return stats, nil
}

// getLimit returns the rate limit for a user and action (custom or default)
func (rl *RateLimiter) getLimit(userID string, action ActionType) RateLimit {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()
	
	// Check for custom limit first
	if userLimits, exists := rl.customLimits[userID]; exists {
		if customLimit, exists := userLimits[action]; exists {
			return customLimit
		}
	}
	
	// Return default limit
	if defaultLimit, exists := rl.config.DefaultLimits[action]; exists {
		return defaultLimit
	}
	
	// Fallback to a conservative default
	return RateLimit{Requests: 10, Window: time.Minute}
}

// getKey generates the Redis key for a user and action
func (rl *RateLimiter) getKey(userID string, action ActionType) string {
	return fmt.Sprintf("%s%s:%s", rl.config.KeyPrefix, userID, string(action))
}

// GetDefaultConfig returns a default rate limiter configuration
func GetDefaultRateLimiterConfig() RateLimiterConfig {
	return RateLimiterConfig{
		DefaultLimits: map[ActionType]RateLimit{
			ActionCreateSession: {Requests: 10, Window: time.Minute},     // 10 sessions per minute
			ActionUpdateAvatar:  {Requests: 60, Window: time.Minute},     // 60 avatar updates per minute (1 per second)
			ActionCreatePOI:     {Requests: 5, Window: time.Minute},      // 5 POI creations per minute
			ActionJoinPOI:       {Requests: 20, Window: time.Minute},     // 20 POI joins per minute
			ActionLeavePOI:      {Requests: 20, Window: time.Minute},     // 20 POI leaves per minute
			ActionUpdatePOI:     {Requests: 10, Window: time.Minute},     // 10 POI updates per minute
			ActionDeletePOI:     {Requests: 5, Window: time.Minute},      // 5 POI deletions per minute
		},
		KeyPrefix: "rate_limit:",
	}
}

// ValidateConfig validates a rate limiter configuration
func ValidateConfig(config RateLimiterConfig) error {
	if config.KeyPrefix == "" {
		return fmt.Errorf("key prefix cannot be empty")
	}
	
	if len(config.DefaultLimits) == 0 {
		return fmt.Errorf("default limits cannot be empty")
	}
	
	for action, limit := range config.DefaultLimits {
		if limit.Requests <= 0 {
			return fmt.Errorf("requests must be positive for action %s", action)
		}
		if limit.Window <= 0 {
			return fmt.Errorf("window must be positive for action %s", action)
		}
		if limit.Window < time.Second {
			return fmt.Errorf("window must be at least 1 second for action %s", action)
		}
	}
	
	return nil
}

// RateLimitError represents a rate limit exceeded error
type RateLimitError struct {
	UserID      string        `json:"userId"`
	Action      ActionType    `json:"action"`
	Limit       int           `json:"limit"`
	Window      time.Duration `json:"window"`
	ResetTime   time.Time     `json:"resetTime"`
	RetryAfter  time.Duration `json:"retryAfter"`
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded for user %s action %s: %d requests per %v (retry after %v)",
		e.UserID, e.Action, e.Limit, e.Window, e.RetryAfter)
}

// NewRateLimitError creates a new rate limit error
func NewRateLimitError(userID string, action ActionType, limit RateLimit, resetTime time.Time) *RateLimitError {
	retryAfter := time.Until(resetTime)
	if retryAfter < 0 {
		retryAfter = 0
	}
	
	return &RateLimitError{
		UserID:     userID,
		Action:     action,
		Limit:      limit.Requests,
		Window:     limit.Window,
		ResetTime:  resetTime,
		RetryAfter: retryAfter,
	}
}

// CheckRateLimit is a helper function that checks rate limit and returns appropriate error
func (rl *RateLimiter) CheckRateLimit(ctx context.Context, userID string, action ActionType) error {
	allowed, err := rl.IsAllowed(ctx, userID, action)
	if err != nil {
		return fmt.Errorf("rate limit check failed: %w", err)
	}
	
	if !allowed {
		limit := rl.getLimit(userID, action)
		resetTime, _ := rl.GetWindowResetTime(ctx, userID, action)
		return NewRateLimitError(userID, action, limit, resetTime)
	}
	
	return nil
}

// GetRateLimitHeaders returns HTTP headers for rate limiting information
func (rl *RateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action ActionType) (map[string]string, error) {
	limit := rl.getLimit(userID, action)
	remaining, err := rl.GetRemainingRequests(ctx, userID, action)
	if err != nil {
		return nil, err
	}
	
	resetTime, err := rl.GetWindowResetTime(ctx, userID, action)
	if err != nil {
		return nil, err
	}
	
	headers := map[string]string{
		"X-RateLimit-Limit":     strconv.Itoa(limit.Requests),
		"X-RateLimit-Remaining": strconv.Itoa(remaining),
		"X-RateLimit-Reset":     strconv.FormatInt(resetTime.Unix(), 10),
		"X-RateLimit-Window":    limit.Window.String(),
	}
	
	return headers, nil
}