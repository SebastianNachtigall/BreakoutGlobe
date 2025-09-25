package services

import (
	"context"
	"time"

	"github.com/stretchr/testify/mock"

	"breakoutglobe/internal/models"
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
}