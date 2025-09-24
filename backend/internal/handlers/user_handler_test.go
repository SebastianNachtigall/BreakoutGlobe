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
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// UserTestScenario provides a fluent API for testing User-related functionality
// This follows established patterns but avoids import cycles
type UserTestScenario struct {
	mockUserService *MockUserService
	mockRateLimiter *MockRateLimiter
	userID          string
	handler         *UserHandler
	router          *gin.Engine
}

// NewUserTestScenario creates a new User test scenario with sensible defaults
func NewUserTestScenario(t *testing.T) *UserTestScenario {
	mockUserService := &MockUserService{}
	mockRateLimiter := &MockRateLimiter{}
	
	scenario := &UserTestScenario{
		mockUserService: mockUserService,
		mockRateLimiter: mockRateLimiter,
		userID:          uuid.New().String(),
	}
	
	// Create handler with mocks
	scenario.handler = NewUserHandler(
		mockUserService,
		mockRateLimiter,
	)
	
	// Setup router
	gin.SetMode(gin.TestMode)
	scenario.router = gin.New()
	scenario.handler.RegisterRoutes(scenario.router)
	
	return scenario
}

// WithUser sets a custom user ID for the scenario
func (s *UserTestScenario) WithUser(userID string) *UserTestScenario {
	s.userID = userID
	return s
}

// ExpectRateLimitSuccess sets up the rate limiter to allow CREATE PROFILE requests
func (s *UserTestScenario) ExpectRateLimitSuccess() *UserTestScenario {
	s.mockRateLimiter.On("CheckRateLimit", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(nil)
	s.mockRateLimiter.On("GetRateLimitHeaders", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(map[string]string{
		"X-RateLimit-Limit":     "5",
		"X-RateLimit-Remaining": "4",
	}, nil)
	
	return s
}

// ExpectRateLimitExceeded sets up the rate limiter to reject CREATE PROFILE requests
func (s *UserTestScenario) ExpectRateLimitExceeded() *UserTestScenario {
	rateLimitErr := &services.RateLimitError{
		UserID:     s.userID,
		Action:     services.ActionCreatePOI,
		Limit:      5,
		RetryAfter: 3600,
	}
	s.mockRateLimiter.On("CheckRateLimit", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(rateLimitErr)
	
	return s
}

// ExpectGuestProfileCreationSuccess sets up the user service to successfully create a guest profile
func (s *UserTestScenario) ExpectGuestProfileCreationSuccess(displayName string, expectedUser *models.User) *UserTestScenario {
	s.mockUserService.On("CreateGuestProfile", mock.Anything, displayName).Return(expectedUser, nil)
	
	return s
}

// ExpectGuestProfileCreationError sets up the user service to return an error during guest profile creation
func (s *UserTestScenario) ExpectGuestProfileCreationError(displayName string, err error) *UserTestScenario {
	s.mockUserService.On("CreateGuestProfile", mock.Anything, displayName).Return(nil, err)
	
	return s
}

// CreateGuestProfile executes a guest profile creation request and returns the response
func (s *UserTestScenario) CreateGuestProfile(t *testing.T, displayName string) *CreateProfileResponse {
	t.Helper()
	
	request := CreateProfileRequest{
		DisplayName: displayName,
		AccountType: "guest",
	}
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	// Verify success status
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			http.StatusCreated, recorder.Code, recorder.Body.String())
		return nil
	}
	
	// Parse response
	var response CreateProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// CreateGuestProfileExpectingError executes a guest profile creation request expecting an error
func (s *UserTestScenario) CreateGuestProfileExpectingError(t *testing.T, displayName string, expectedStatus int) *httptest.ResponseRecorder {
	t.Helper()
	
	request := CreateProfileRequest{
		DisplayName: displayName,
		AccountType: "guest",
	}
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	// Verify expected error status
	if recorder.Code != expectedStatus {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			expectedStatus, recorder.Code, recorder.Body.String())
	}
	
	return recorder
}

// Cleanup cleans up test resources
func (s *UserTestScenario) Cleanup(t *testing.T) {
	s.mockUserService.AssertExpectations(t)
	s.mockRateLimiter.AssertExpectations(t)
}

// TestUserHandler demonstrates the established test infrastructure patterns

func TestCreateGuestProfile_Success(t *testing.T) {
	// Setup using established infrastructure patterns - 3 lines vs 15+ in old version
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	// Create expected user using builder pattern
	expectedUser := &models.User{
		ID:          uuid.New().String(),
		DisplayName: "Test User",
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Configure expectations - fluent and readable
	scenario.ExpectRateLimitSuccess().
		ExpectGuestProfileCreationSuccess("Test User", expectedUser)

	// Execute and verify - business intent is clear
	response := scenario.CreateGuestProfile(t, "Test User")

	// Use assertions focused on business logic
	if response.DisplayName != "Test User" {
		t.Errorf("Expected display name 'Test User', got %s", response.DisplayName)
	}
	if response.AccountType != "guest" {
		t.Errorf("Expected account type 'guest', got %s", response.AccountType)
	}
	if response.Role != "user" {
		t.Errorf("Expected role 'user', got %s", response.Role)
	}
	if !response.IsActive {
		t.Errorf("Expected user to be active")
	}
	if response.ID == "" {
		t.Errorf("Expected user ID to be set")
	}
}

func TestCreateGuestProfile_ValidationError(t *testing.T) {
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	// Validation happens before rate limiting and service calls, so no expectations needed

	// Execute request expecting validation error
	recorder := scenario.CreateGuestProfileExpectingError(t, "AB", 400)

	// Verify error response
	if recorder.Code != 400 {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}
	
	// Check error response contains validation error code
	body := recorder.Body.String()
	if !contains(body, "VALIDATION_ERROR") {
		t.Errorf("Expected error response to contain 'VALIDATION_ERROR', got: %s", body)
	}
}

func TestCreateGuestProfile_RateLimited(t *testing.T) {
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	// Rate limit scenario is self-documenting
	scenario.ExpectRateLimitExceeded()

	// Execute request expecting rate limit error
	recorder := scenario.CreateGuestProfileExpectingError(t, "Test User", 429)

	// Error handling is automatic and consistent
	if recorder.Code != 429 {
		t.Errorf("Expected status 429, got %d", recorder.Code)
	}
	
	// Check for rate limit headers
	retryAfter := recorder.Header().Get("Retry-After")
	if retryAfter != "3600" {
		t.Errorf("Expected Retry-After header '3600', got '%s'", retryAfter)
	}
	
	// Check error response contains rate limit error code
	body := recorder.Body.String()
	if !contains(body, "RATE_LIMIT_EXCEEDED") {
		t.Errorf("Expected error response to contain 'RATE_LIMIT_EXCEEDED', got: %s", body)
	}
}



// Note: CreateProfileRequest and CreateProfileResponse are defined in user_handler.go

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		len(substr) == 0 || 
		indexOfSubstring(s, substr) >= 0)
}

func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

// Note: MockRateLimiter is defined in session_handler_test.go

// MockUserService for testing - follows established patterns but avoids import cycles
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

func (m *MockUserService) UpdateUser(ctx context.Context, userID string, updateData map[string]interface{}) (*models.User, error) {
	args := m.Called(ctx, userID, updateData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}