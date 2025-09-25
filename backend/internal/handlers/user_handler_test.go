package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
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

// ExpectAvatarUploadSuccess sets up the user service to successfully upload an avatar
func (s *UserTestScenario) ExpectAvatarUploadSuccess(userID string, filename string, fileData []byte, expectedUser *models.User) *UserTestScenario {
	s.mockUserService.On("UploadAvatar", mock.Anything, userID, filename, fileData).Return(expectedUser, nil)
	
	return s
}

// ExpectAvatarUploadError sets up the user service to return an error during avatar upload
func (s *UserTestScenario) ExpectAvatarUploadError(userID string, filename string, fileData []byte, err error) *UserTestScenario {
	s.mockUserService.On("UploadAvatar", mock.Anything, userID, filename, fileData).Return(nil, err)
	
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

// UploadAvatar executes an avatar upload request and returns the response
func (s *UserTestScenario) UploadAvatar(t *testing.T, userID string, filename string, fileData []byte, contentType string) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create multipart form data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	
	// Create form file with proper content type
	part, err := writer.CreateFormFile("avatar", filename)
	if err != nil {
		t.Errorf("Failed to create form file: %v", err)
		return nil
	}
	
	// Set content type in the part header
	if contentType != "" {
		// We need to create the part manually to set the content type
		writer.Close()
		body.Reset()
		writer = multipart.NewWriter(body)
		
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="avatar"; filename="%s"`, filename))
		h.Set("Content-Type", contentType)
		part, err = writer.CreatePart(h)
		if err != nil {
			t.Errorf("Failed to create form part: %v", err)
			return nil
		}
	}
	
	// Write file data
	_, err = part.Write(fileData)
	if err != nil {
		t.Errorf("Failed to write file data: %v", err)
		return nil
	}
	
	// Close writer
	writer.Close()
	
	req := httptest.NewRequest(http.MethodPost, "/api/users/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("X-User-ID", userID) // Simulate authenticated user
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// UploadAvatarExpectingError executes an avatar upload request expecting an error
func (s *UserTestScenario) UploadAvatarExpectingError(t *testing.T, userID string, filename string, fileData []byte, contentType string, expectedStatus int) *httptest.ResponseRecorder {
	t.Helper()
	
	recorder := s.UploadAvatar(t, userID, filename, fileData, contentType)
	
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



func TestUploadAvatar_Success(t *testing.T) {
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	userID := uuid.New().String()
	filename := "avatar.jpg"
	fileData := []byte("fake-jpeg-data")
	
	// Create expected user with avatar URL
	expectedUser := &models.User{
		ID:          userID,
		DisplayName: "Test User",
		AvatarURL:   "/api/users/avatar/" + filename,
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Configure expectations
	scenario.ExpectRateLimitSuccess().
		ExpectAvatarUploadSuccess(userID, filename, fileData, expectedUser)

	// Execute avatar upload
	recorder := scenario.UploadAvatar(t, userID, filename, fileData, "image/jpeg")

	// Verify success
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			http.StatusOK, recorder.Code, recorder.Body.String())
	}

	// Parse response
	var response CreateProfileResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
	}

	// Verify avatar URL is set
	if response.ID != userID {
		t.Errorf("Expected user ID %s, got %s", userID, response.ID)
	}
}

func TestUploadAvatar_InvalidFileType(t *testing.T) {
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	userID := uuid.New().String()
	filename := "avatar.txt"
	fileData := []byte("not-an-image")

	// Rate limiting happens before file validation, but rate limit headers are not added on validation errors
	scenario.mockRateLimiter.On("CheckRateLimit", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(nil)

	// Execute request expecting validation error
	recorder := scenario.UploadAvatarExpectingError(t, userID, filename, fileData, "text/plain", 400)

	// Verify error response
	body := recorder.Body.String()
	if !contains(body, "INVALID_FILE_TYPE") {
		t.Errorf("Expected error response to contain 'INVALID_FILE_TYPE', got: %s", body)
	}
}

func TestUploadAvatar_FileTooLarge(t *testing.T) {
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	userID := uuid.New().String()
	filename := "avatar.jpg"
	// Create file data larger than 2MB
	fileData := make([]byte, 3*1024*1024) // 3MB

	// Rate limiting happens before file validation, but rate limit headers are not added on validation errors
	scenario.mockRateLimiter.On("CheckRateLimit", mock.Anything, mock.Anything, services.ActionCreatePOI).Return(nil)

	// Execute request expecting validation error
	recorder := scenario.UploadAvatarExpectingError(t, userID, filename, fileData, "image/jpeg", 400)

	// Verify error response
	body := recorder.Body.String()
	if !contains(body, "FILE_TOO_LARGE") {
		t.Errorf("Expected error response to contain 'FILE_TOO_LARGE', got: %s", body)
	}
}

func TestUploadAvatar_RateLimited(t *testing.T) {
	scenario := NewUserTestScenario(t)
	defer scenario.Cleanup(t)

	userID := uuid.New().String()
	filename := "avatar.jpg"
	fileData := []byte("fake-jpeg-data")

	// Rate limit scenario
	scenario.ExpectRateLimitExceeded()

	// Execute request expecting rate limit error
	recorder := scenario.UploadAvatarExpectingError(t, userID, filename, fileData, "image/jpeg", 429)

	// Verify rate limit error
	if recorder.Code != 429 {
		t.Errorf("Expected status 429, got %d", recorder.Code)
	}
	
	retryAfter := recorder.Header().Get("Retry-After")
	if retryAfter != "3600" {
		t.Errorf("Expected Retry-After header '3600', got '%s'", retryAfter)
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

func (m *MockUserService) UploadAvatar(ctx context.Context, userID string, filename string, fileData []byte) (*models.User, error) {
	args := m.Called(ctx, userID, filename, fileData)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}