package services

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"
)

// MockRateLimiter is a mock implementation of RateLimiterInterface for testing
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) IsAllowed(ctx context.Context, userID string, action ActionType) (bool, error) {
	args := m.Called(ctx, userID, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockRateLimiter) GetRemainingRequests(ctx context.Context, userID string, action ActionType) (int, error) {
	args := m.Called(ctx, userID, action)
	return args.Int(0), args.Error(1)
}

func (m *MockRateLimiter) GetWindowResetTime(ctx context.Context, userID string, action ActionType) (time.Time, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockRateLimiter) SetCustomLimit(userID string, action ActionType, limit RateLimit) {
	m.Called(userID, action, limit)
}

func (m *MockRateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRateLimiter) GetUserStats(ctx context.Context, userID string) (*UserRateLimitStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserRateLimitStats), args.Error(1)
}

func (m *MockRateLimiter) CheckRateLimit(ctx context.Context, userID string, action ActionType) error {
	args := m.Called(ctx, userID, action)
	return args.Error(0)
}

func (m *MockRateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action ActionType) (map[string]string, error) {
	args := m.Called(ctx, userID, action)
	if len(args) < 2 {
		return nil, nil
	}
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

// MockUserService is a mock implementation of UserServiceInterface for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateGuestProfile(ctx context.Context, displayName string) (*models.User, error) {
	args := m.Called(ctx, displayName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) CreateGuestProfileWithAboutMe(ctx context.Context, displayName, aboutMe string) (*models.User, error) {
	args := m.Called(ctx, displayName, aboutMe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UploadAvatar(ctx context.Context, userID string, filename string, fileData []byte) (*models.User, error) {
	args := m.Called(ctx, userID, filename, fileData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*models.User, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}// MockPu
// MockPubSub is a mock implementation of PubSub for testing
type MockPubSub struct {
	mock.Mock
}

func (m *MockPubSub) PublishPOICreated(ctx context.Context, event redis.POICreatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOIUpdated(ctx context.Context, event redis.POIUpdatedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOIJoined(ctx context.Context, event redis.POIJoinedEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOILeft(ctx context.Context, event redis.POILeftEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishAvatarMovement(ctx context.Context, event redis.AvatarMovementEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOIJoinedWithParticipants(ctx context.Context, event redis.POIJoinedEventWithParticipants) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockPubSub) PublishPOILeftWithParticipants(ctx context.Context, event redis.POILeftEventWithParticipants) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}