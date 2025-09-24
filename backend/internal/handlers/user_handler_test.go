package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// TestUserHandler_CreateProfile tests user profile creation using TDD and scenario-based testing
func TestUserHandler_CreateProfile(t *testing.T) {
	t.Run("creates guest profile successfully", func(t *testing.T) {
		// RED PHASE: This test will fail because UserHandler doesn't exist yet
		scenario := newUserHandlerScenario(t)
		defer scenario.cleanup()
		
		// Use expectProfileCreationSuccess() for guest profile creation workflow
		scenario.expectRateLimitSuccess().
			expectProfileCreationSuccess()
		
		profile := scenario.createProfile(CreateProfileRequest{
			DisplayName: "Test User",
			AccountType: "guest",
		})
		
		// Use fluent assertions focused on business logic
		assert.Equal(t, "Test User", profile.DisplayName)
		assert.Equal(t, "guest", profile.AccountType)
		assert.Equal(t, "user", profile.Role)
		assert.True(t, profile.IsActive)
		assert.NotEmpty(t, profile.ID)
	})

	t.Run("validates display name length", func(t *testing.T) {
		scenario := newUserHandlerScenario(t)
		defer scenario.cleanup()
		
		// Test display name validation (3-50 characters)
		req := httptest.NewRequest(http.MethodPost, "/api/users/profile", 
			bytes.NewBuffer([]byte(`{"displayName":"AB","accountType":"guest"}`)))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		
		scenario.router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		
		var errorResponse ErrorResponse
		json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		assert.Equal(t, "VALIDATION_ERROR", errorResponse.Code)
		assert.Contains(t, errorResponse.Message, "display name must be at least 3 characters")
	})

	t.Run("handles rate limiting", func(t *testing.T) {
		scenario := newUserHandlerScenario(t)
		defer scenario.cleanup()
		
		// Setup rate limit error
		rateLimitErr := &services.RateLimitError{
			UserID:     "test-user",
			Action:     services.ActionCreatePOI, // Use existing action for now
			Limit:      5,
			RetryAfter: 3600,
		}
		scenario.mockRateLimiter.On("CheckRateLimit", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(rateLimitErr)
		
		req := httptest.NewRequest(http.MethodPost, "/api/users/profile", 
			bytes.NewBuffer([]byte(`{"displayName":"Test User","accountType":"guest"}`)))
		req.Header.Set("Content-Type", "application/json")
		recorder := httptest.NewRecorder()
		
		scenario.router.ServeHTTP(recorder, req)
		
		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		assert.Equal(t, "3600", recorder.Header().Get("Retry-After"))
		
		var errorResponse ErrorResponse
		json.Unmarshal(recorder.Body.Bytes(), &errorResponse)
		assert.Equal(t, "RATE_LIMIT_EXCEEDED", errorResponse.Code)
	})
}

// userHandlerScenario provides scenario-based testing for UserHandler
type userHandlerScenario struct {
	t               *testing.T
	mockUserService *MockUserService // This interface doesn't exist yet - will fail
	mockRateLimiter *MockRateLimiter
	handler         *UserHandler // This doesn't exist yet - will fail
	router          *gin.Engine
}

// newUserHandlerScenario creates a new UserHandler test scenario
func newUserHandlerScenario(t *testing.T) *userHandlerScenario {
	gin.SetMode(gin.TestMode)
	
	// This will fail because UserHandler and MockUserService don't exist yet
	mockUserService := new(MockUserService)
	mockRateLimiter := new(MockRateLimiter)
	handler := NewUserHandler(mockUserService, mockRateLimiter) // This function doesn't exist yet
	
	router := gin.New()
	handler.RegisterRoutes(router) // This method doesn't exist yet
	
	return &userHandlerScenario{
		t:               t,
		mockUserService: mockUserService,
		mockRateLimiter: mockRateLimiter,
		handler:         handler,
		router:          router,
	}
}

// cleanup cleans up test resources
func (s *userHandlerScenario) cleanup() {
	s.mockUserService.AssertExpectations(s.t)
	s.mockRateLimiter.AssertExpectations(s.t)
}

// expectRateLimitSuccess sets up expectations for successful rate limiting
func (s *userHandlerScenario) expectRateLimitSuccess() *userHandlerScenario {
	s.mockRateLimiter.On("CheckRateLimit", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(nil)
	s.mockRateLimiter.On("GetRateLimitHeaders", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(map[string]string{
		"X-RateLimit-Limit":     "5",
		"X-RateLimit-Remaining": "4",
	}, nil)
	return s
}

// expectProfileCreationSuccess sets up expectations for successful profile creation
func (s *userHandlerScenario) expectProfileCreationSuccess() *userHandlerScenario {
	expectedUser := &models.User{
		ID:          "user-123",
		DisplayName: "Test User",
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
	}
	
	s.mockUserService.On("CreateGuestProfile", mock.Anything, "Test User").Return(expectedUser, nil)
	return s
}

// createProfile executes profile creation request
func (s *userHandlerScenario) createProfile(request CreateProfileRequest) *CreateProfileResponse {
	body, _ := json.Marshal(request)
	req := httptest.NewRequest(http.MethodPost, "/api/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	s.router.ServeHTTP(recorder, req)
	
	var response CreateProfileResponse
	json.Unmarshal(recorder.Body.Bytes(), &response)
	return &response
}

// Note: CreateProfileRequest and CreateProfileResponse are now defined in user_handler.go

// MockUserService for testing - doesn't exist yet
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateGuestProfile(ctx context.Context, displayName string) (*models.User, error) {
	args := m.Called(ctx, displayName)
	return args.Get(0).(*models.User), args.Error(1)
}

// Note: MockRateLimiter is already defined in session_handler_test.go