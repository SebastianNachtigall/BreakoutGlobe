package services

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockRedisClient is a mock implementation of Redis client for rate limiting
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) ZAdd(ctx context.Context, key string, score float64, member string) error {
	args := m.Called(ctx, key, score, member)
	return args.Error(0)
}

func (m *MockRedisClient) ZRemRangeByScore(ctx context.Context, key string, min, max string) error {
	args := m.Called(ctx, key, min, max)
	return args.Error(0)
}

func (m *MockRedisClient) ZCard(ctx context.Context, key string) (int64, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	args := m.Called(ctx, key, expiration)
	return args.Error(0)
}

func (m *MockRedisClient) ZRangeWithScores(ctx context.Context, key string, start, stop int64) ([]interface{}, error) {
	args := m.Called(ctx, key, start, stop)
	return args.Get(0).([]interface{}), args.Error(1)
}

func (m *MockRedisClient) Pipeline() PipelineInterface {
	args := m.Called()
	return args.Get(0).(PipelineInterface)
}

// MockPipeline is a mock implementation of Redis pipeline
type MockPipeline struct {
	mock.Mock
}

func (m *MockPipeline) ZAdd(ctx context.Context, key string, score float64, member string) {
	m.Called(ctx, key, score, member)
}

func (m *MockPipeline) ZRemRangeByScore(ctx context.Context, key string, min, max string) {
	m.Called(ctx, key, min, max)
}

func (m *MockPipeline) ZCard(ctx context.Context, key string) {
	m.Called(ctx, key)
}

func (m *MockPipeline) Expire(ctx context.Context, key string, expiration time.Duration) {
	m.Called(ctx, key, expiration)
}

func (m *MockPipeline) Exec(ctx context.Context) ([]interface{}, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]interface{}), args.Error(1)
}

// RateLimiterTestSuite contains the test suite for RateLimiter
type RateLimiterTestSuite struct {
	suite.Suite
	mockRedis   *MockRedisClient
	rateLimiter *RateLimiter
}

func (suite *RateLimiterTestSuite) SetupTest() {
	suite.mockRedis = new(MockRedisClient)
	
	// Default rate limit configurations
	config := RateLimiterConfig{
		DefaultLimits: map[ActionType]RateLimit{
			ActionCreateSession:    {Requests: 10, Window: time.Minute},
			ActionUpdateAvatar:     {Requests: 60, Window: time.Minute},
			ActionCreatePOI:        {Requests: 5, Window: time.Minute},
			ActionJoinPOI:          {Requests: 20, Window: time.Minute},
			ActionLeavePOI:         {Requests: 20, Window: time.Minute},
			ActionUpdatePOI:        {Requests: 10, Window: time.Minute},
			ActionDeletePOI:        {Requests: 5, Window: time.Minute},
		},
		KeyPrefix: "rate_limit:",
	}
	
	suite.rateLimiter = NewRateLimiter(suite.mockRedis, config)
}

func (suite *RateLimiterTestSuite) TearDownTest() {
	suite.mockRedis.AssertExpectations(suite.T())
}

func (suite *RateLimiterTestSuite) TestIsAllowed_WithinLimit() {
	ctx := context.Background()
	userID := "user-123"
	action := ActionCreateSession
	
	// Mock expectations for sliding window check
	key := "rate_limit:user-123:create_session"
	
	// Mock pipeline operations
	mockPipeline := new(MockPipeline)
	suite.mockRedis.On("Pipeline").Return(PipelineInterface(mockPipeline))
	
	// Pipeline operations
	mockPipeline.On("ZRemRangeByScore", ctx, key, "0", mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZAdd", ctx, key, mock.AnythingOfType("float64"), mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZCard", ctx, key).Return()
	mockPipeline.On("Expire", ctx, key, time.Minute*2).Return()
	
	// Pipeline execution returns count within limit
	mockPipeline.On("Exec", ctx).Return([]interface{}{nil, nil, int64(5), nil}, nil)
	
	// Execute
	allowed, err := suite.rateLimiter.IsAllowed(ctx, userID, action)
	
	// Assert
	suite.NoError(err)
	suite.True(allowed)
}

func (suite *RateLimiterTestSuite) TestIsAllowed_ExceedsLimit() {
	ctx := context.Background()
	userID := "user-123"
	action := ActionCreateSession
	
	// Mock expectations for sliding window check
	key := "rate_limit:user-123:create_session"
	
	// Mock pipeline operations
	mockPipeline := new(MockPipeline)
	suite.mockRedis.On("Pipeline").Return(PipelineInterface(mockPipeline))
	
	// Pipeline operations
	mockPipeline.On("ZRemRangeByScore", ctx, key, "0", mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZAdd", ctx, key, mock.AnythingOfType("float64"), mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZCard", ctx, key).Return()
	mockPipeline.On("Expire", ctx, key, time.Minute*2).Return()
	
	// Pipeline execution returns count exceeding limit (10)
	mockPipeline.On("Exec", ctx).Return([]interface{}{nil, nil, int64(11), nil}, nil)
	
	// Execute
	allowed, err := suite.rateLimiter.IsAllowed(ctx, userID, action)
	
	// Assert
	suite.NoError(err)
	suite.False(allowed)
}

func (suite *RateLimiterTestSuite) TestIsAllowed_CustomLimit() {
	ctx := context.Background()
	userID := "premium-user-456"
	action := ActionCreatePOI
	
	// Set custom limit for this user
	customLimit := RateLimit{Requests: 20, Window: time.Minute}
	suite.rateLimiter.SetCustomLimit(userID, action, customLimit)
	
	key := "rate_limit:premium-user-456:create_poi"
	
	// Mock pipeline operations
	mockPipeline := new(MockPipeline)
	suite.mockRedis.On("Pipeline").Return(PipelineInterface(mockPipeline))
	
	// Pipeline operations
	mockPipeline.On("ZRemRangeByScore", ctx, key, "0", mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZAdd", ctx, key, mock.AnythingOfType("float64"), mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZCard", ctx, key).Return()
	mockPipeline.On("Expire", ctx, key, time.Minute*2).Return()
	
	// Pipeline execution returns count within custom limit
	mockPipeline.On("Exec", ctx).Return([]interface{}{nil, nil, int64(15), nil}, nil)
	
	// Execute
	allowed, err := suite.rateLimiter.IsAllowed(ctx, userID, action)
	
	// Assert
	suite.NoError(err)
	suite.True(allowed) // 15 < 20 (custom limit)
}

func (suite *RateLimiterTestSuite) TestIsAllowed_DifferentActions() {
	ctx := context.Background()
	userID := "user-123"
	
	// Test different actions have different limits
	testCases := []struct {
		action        ActionType
		expectedLimit int
	}{
		{ActionCreateSession, 10},
		{ActionUpdateAvatar, 60},
		{ActionCreatePOI, 5},
		{ActionJoinPOI, 20},
		{ActionLeavePOI, 20},
		{ActionUpdatePOI, 10},
		{ActionDeletePOI, 5},
	}
	
	for _, tc := range testCases {
		suite.Run(string(tc.action), func() {
			// Create a fresh mock for each sub-test
			mockRedis := new(MockRedisClient)
			mockPipeline := new(MockPipeline)
			
			// Create a new rate limiter instance for this test
			config := RateLimiterConfig{
				DefaultLimits: map[ActionType]RateLimit{
					ActionCreateSession:    {Requests: 10, Window: time.Minute},
					ActionUpdateAvatar:     {Requests: 60, Window: time.Minute},
					ActionCreatePOI:        {Requests: 5, Window: time.Minute},
					ActionJoinPOI:          {Requests: 20, Window: time.Minute},
					ActionLeavePOI:         {Requests: 20, Window: time.Minute},
					ActionUpdatePOI:        {Requests: 10, Window: time.Minute},
					ActionDeletePOI:        {Requests: 5, Window: time.Minute},
				},
				KeyPrefix: "rate_limit:",
			}
			rateLimiter := NewRateLimiter(mockRedis, config)
			
			key := rateLimiter.getKey(userID, tc.action)
			
			// Mock pipeline operations
			mockRedis.On("Pipeline").Return(PipelineInterface(mockPipeline))
			
			mockPipeline.On("ZRemRangeByScore", ctx, key, "0", mock.AnythingOfType("string")).Return()
			mockPipeline.On("ZAdd", ctx, key, mock.AnythingOfType("float64"), mock.AnythingOfType("string")).Return()
			mockPipeline.On("ZCard", ctx, key).Return()
			mockPipeline.On("Expire", ctx, key, time.Minute*2).Return()
			
			// Return count just under the limit
			mockPipeline.On("Exec", ctx).Return([]interface{}{nil, nil, int64(tc.expectedLimit - 1), nil}, nil)
			
			// Execute
			allowed, err := rateLimiter.IsAllowed(ctx, userID, tc.action)
			
			// Assert
			suite.NoError(err)
			suite.True(allowed)
			
			// Verify expectations
			mockRedis.AssertExpectations(suite.T())
			mockPipeline.AssertExpectations(suite.T())
		})
	}
}

func (suite *RateLimiterTestSuite) TestGetRemainingRequests() {
	ctx := context.Background()
	userID := "user-123"
	action := ActionCreateSession
	
	key := "rate_limit:user-123:create_session"
	
	// Mock ZCard to return current count
	suite.mockRedis.On("ZCard", ctx, key).Return(int64(3), nil)
	
	// Execute
	remaining, err := suite.rateLimiter.GetRemainingRequests(ctx, userID, action)
	
	// Assert
	suite.NoError(err)
	suite.Equal(7, remaining) // 10 (limit) - 3 (current) = 7
}

func (suite *RateLimiterTestSuite) TestGetRemainingRequests_ExceedsLimit() {
	ctx := context.Background()
	userID := "user-123"
	action := ActionCreateSession
	
	key := "rate_limit:user-123:create_session"
	
	// Mock ZCard to return count exceeding limit
	suite.mockRedis.On("ZCard", ctx, key).Return(int64(15), nil)
	
	// Execute
	remaining, err := suite.rateLimiter.GetRemainingRequests(ctx, userID, action)
	
	// Assert
	suite.NoError(err)
	suite.Equal(0, remaining) // Should return 0 when exceeded
}

func (suite *RateLimiterTestSuite) TestGetWindowResetTime() {
	ctx := context.Background()
	userID := "user-123"
	action := ActionCreateSession
	
	// Mock empty window (no entries)
	suite.mockRedis.On("ZRangeWithScores", ctx, "rate_limit:user-123:create_session", int64(0), int64(0)).Return([]interface{}{}, nil)
	
	// Execute
	resetTime, err := suite.rateLimiter.GetWindowResetTime(ctx, userID, action)
	
	// Assert
	suite.NoError(err)
	suite.True(resetTime.Before(time.Now().Add(time.Second)), "Empty window should reset immediately")
}

func (suite *RateLimiterTestSuite) TestClearUserLimits() {
	ctx := context.Background()
	userID := "user-123"
	
	// Mock expectations for clearing all action keys for user
	expectedKeys := []string{
		"rate_limit:user-123:create_session",
		"rate_limit:user-123:update_avatar",
		"rate_limit:user-123:create_poi",
		"rate_limit:user-123:join_poi",
		"rate_limit:user-123:leave_poi",
		"rate_limit:user-123:update_poi",
		"rate_limit:user-123:delete_poi",
	}
	
	// Mock pipeline operations for clearing
	mockPipeline := new(MockPipeline)
	suite.mockRedis.On("Pipeline").Return(PipelineInterface(mockPipeline))
	
	for _, key := range expectedKeys {
		mockPipeline.On("ZRemRangeByScore", ctx, key, "0", "+inf").Return()
	}
	
	mockPipeline.On("Exec", ctx).Return(make([]interface{}, len(expectedKeys)), nil)
	
	// Execute
	err := suite.rateLimiter.ClearUserLimits(ctx, userID)
	
	// Assert
	suite.NoError(err)
}

func (suite *RateLimiterTestSuite) TestGetUserStats() {
	ctx := context.Background()
	userID := "user-123"
	
	// Mock ZCard calls for all actions
	actions := []ActionType{
		ActionCreateSession, ActionUpdateAvatar, ActionCreatePOI,
		ActionJoinPOI, ActionLeavePOI, ActionUpdatePOI, ActionDeletePOI,
	}
	
	expectedCounts := map[ActionType]int64{
		ActionCreateSession: 3,
		ActionUpdateAvatar:  25,
		ActionCreatePOI:     1,
		ActionJoinPOI:       5,
		ActionLeavePOI:      2,
		ActionUpdatePOI:     0,
		ActionDeletePOI:     0,
	}
	
	for _, action := range actions {
		key := suite.rateLimiter.getKey(userID, action)
		suite.mockRedis.On("ZCard", ctx, key).Return(expectedCounts[action], nil)
	}
	
	// Execute
	stats, err := suite.rateLimiter.GetUserStats(ctx, userID)
	
	// Assert
	suite.NoError(err)
	suite.NotNil(stats)
	suite.Equal(userID, stats.UserID)
	suite.Equal(len(actions), len(stats.ActionStats))
	
	for _, action := range actions {
		actionStats, exists := stats.ActionStats[action]
		suite.True(exists)
		suite.Equal(int(expectedCounts[action]), actionStats.CurrentCount)
		suite.True(actionStats.Limit > 0)
		suite.True(actionStats.WindowEnd.After(time.Now()))
	}
}

func (suite *RateLimiterTestSuite) TestIsAllowed_RedisError() {
	ctx := context.Background()
	userID := "user-123"
	action := ActionCreateSession
	
	// Mock pipeline operations with error
	mockPipeline := new(MockPipeline)
	suite.mockRedis.On("Pipeline").Return(PipelineInterface(mockPipeline))
	
	mockPipeline.On("ZRemRangeByScore", ctx, mock.AnythingOfType("string"), "0", mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZAdd", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("float64"), mock.AnythingOfType("string")).Return()
	mockPipeline.On("ZCard", ctx, mock.AnythingOfType("string")).Return()
	mockPipeline.On("Expire", ctx, mock.AnythingOfType("string"), time.Minute*2).Return()
	
	// Pipeline execution returns error
	mockPipeline.On("Exec", ctx).Return(nil, assert.AnError)
	
	// Execute
	allowed, err := suite.rateLimiter.IsAllowed(ctx, userID, action)
	
	// Assert
	suite.Error(err)
	suite.False(allowed) // Should deny on error for safety
}

func (suite *RateLimiterTestSuite) TestGetKey() {
	userID := "user-123"
	action := ActionCreateSession
	
	key := suite.rateLimiter.getKey(userID, action)
	
	suite.Equal("rate_limit:user-123:create_session", key)
}

func (suite *RateLimiterTestSuite) TestActionTypeString() {
	testCases := []struct {
		action   ActionType
		expected string
	}{
		{ActionCreateSession, "create_session"},
		{ActionUpdateAvatar, "update_avatar"},
		{ActionCreatePOI, "create_poi"},
		{ActionJoinPOI, "join_poi"},
		{ActionLeavePOI, "leave_poi"},
		{ActionUpdatePOI, "update_poi"},
		{ActionDeletePOI, "delete_poi"},
	}
	
	for _, tc := range testCases {
		suite.Equal(tc.expected, string(tc.action))
	}
}

func TestRateLimiterTestSuite(t *testing.T) {
	suite.Run(t, new(RateLimiterTestSuite))
}