package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock implementations for testing

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) GenerateJWT(userID, email string, role models.UserRole) (string, time.Time, error) {
	args := m.Called(userID, email, role)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockAuthService) ValidateJWT(token string) (*services.JWTClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.JWTClaims), args.Error(1)
}

type MockAuthUserService struct {
	mock.Mock
}

func (m *MockAuthUserService) CreateFullAccount(ctx context.Context, email, password, displayName, aboutMe string) (*models.User, error) {
	args := m.Called(ctx, email, password, displayName, aboutMe)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthUserService) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthUserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockAuthUserService) VerifyPassword(ctx context.Context, userID, password string) error {
	args := m.Called(ctx, userID, password)
	return args.Error(0)
}

type MockAuthRateLimiter struct {
	mock.Mock
}

func (m *MockAuthRateLimiter) CheckRateLimit(ctx context.Context, key string, action services.ActionType) error {
	args := m.Called(ctx, key, action)
	return args.Error(0)
}

func (m *MockAuthRateLimiter) IsAllowed(ctx context.Context, userID string, action services.ActionType) (bool, error) {
	args := m.Called(ctx, userID, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action services.ActionType) (map[string]string, error) {
	args := m.Called(ctx, userID, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockAuthRateLimiter) GetRemainingRequests(ctx context.Context, userID string, action services.ActionType) (int, error) {
	args := m.Called(ctx, userID, action)
	return args.Int(0), args.Error(1)
}

func (m *MockAuthRateLimiter) GetWindowResetTime(ctx context.Context, userID string, action services.ActionType) (time.Time, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockAuthRateLimiter) SetCustomLimit(userID string, action services.ActionType, limit services.RateLimit) {
	m.Called(userID, action, limit)
}

func (m *MockAuthRateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockAuthRateLimiter) GetUserStats(ctx context.Context, userID string) (*services.UserRateLimitStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserRateLimitStats), args.Error(1)
}

// Helper functions

func setupAuthTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func createTestUser(id, email, displayName string, accountType models.AccountType, role models.UserRole) *models.User {
	emailPtr := &email
	return &models.User{
		ID:          id,
		Email:       emailPtr,
		DisplayName: displayName,
		AccountType: accountType,
		Role:        role,
		CreatedAt:   time.Now(),
	}
}

// Signup Tests

func TestSignup_Success(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/signup", handler.Signup)
	
	// Setup test data
	signupReq := SignupRequest{
		Email:       "test@example.com",
		Password:    "Password123!",
		DisplayName: "Test User",
		AboutMe:     "Hello world",
	}
	
	user := createTestUser("user-123", "test@example.com", "Test User", models.AccountTypeFull, models.UserRoleUser)
	expiresAt := time.Now().Add(24 * time.Hour)
	
	// Setup expectations
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "signup:test@example.com", services.ActionCreatePOI).Return(nil)
	mockUserService.On("CreateFullAccount", mock.Anything, "test@example.com", "Password123!", "Test User", "Hello world").Return(user, nil)
	mockAuthService.On("GenerateJWT", "user-123", "test@example.com", models.UserRoleUser).Return("test-token", expiresAt, nil)
	
	// Make request
	body, _ := json.Marshal(signupReq)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	// Assertions
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "test-token", response.Token)
	assert.Equal(t, "user-123", response.User.ID)
	assert.Equal(t, "test@example.com", response.User.Email)
	assert.Equal(t, "Test User", response.User.DisplayName)
	
	mockRateLimiter.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestSignup_InvalidRequest(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/signup", handler.Signup)
	
	testCases := []struct {
		name        string
		requestBody string
		expectedMsg string
	}{
		{
			name:        "missing email",
			requestBody: `{"password":"Password123!","displayName":"Test"}`,
			expectedMsg: "INVALID_REQUEST",
		},
		{
			name:        "invalid email format",
			requestBody: `{"email":"not-an-email","password":"Password123!","displayName":"Test"}`,
			expectedMsg: "INVALID_REQUEST",
		},
		{
			name:        "password too short",
			requestBody: `{"email":"test@example.com","password":"short","displayName":"Test"}`,
			expectedMsg: "INVALID_REQUEST",
		},
		{
			name:        "missing display name",
			requestBody: `{"email":"test@example.com","password":"Password123!"}`,
			expectedMsg: "INVALID_REQUEST",
		},
		{
			name:        "display name too short",
			requestBody: `{"email":"test@example.com","password":"Password123!","displayName":"AB"}`,
			expectedMsg: "INVALID_REQUEST",
		},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBufferString(tc.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), tc.expectedMsg)
		})
	}
}

func TestSignup_EmailAlreadyInUse(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/signup", handler.Signup)
	
	signupReq := SignupRequest{
		Email:       "existing@example.com",
		Password:    "Password123!",
		DisplayName: "Test User",
	}
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "signup:existing@example.com", services.ActionCreatePOI).Return(nil)
	mockUserService.On("CreateFullAccount", mock.Anything, "existing@example.com", "Password123!", "Test User", "").
		Return(nil, errors.New("email already in use"))
	
	body, _ := json.Marshal(signupReq)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "EMAIL_IN_USE")
	
	mockRateLimiter.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestSignup_WeakPassword(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/signup", handler.Signup)
	
	signupReq := SignupRequest{
		Email:       "test@example.com",
		Password:    "weakpass",
		DisplayName: "Test User",
	}
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "signup:test@example.com", services.ActionCreatePOI).Return(nil)
	mockUserService.On("CreateFullAccount", mock.Anything, "test@example.com", "weakpass", "Test User", "").
		Return(nil, errors.New("password does not meet requirements"))
	
	body, _ := json.Marshal(signupReq)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_PASSWORD")
	
	mockRateLimiter.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestSignup_RateLimited(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/signup", handler.Signup)
	
	signupReq := SignupRequest{
		Email:       "test@example.com",
		Password:    "Password123!",
		DisplayName: "Test User",
	}
	
	rateLimitErr := &services.RateLimitError{
		UserID:     "test@example.com",
		Action:     services.ActionCreatePOI,
		Limit:      100,
		RetryAfter: 3600,
	}
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "signup:test@example.com", services.ActionCreatePOI).Return(rateLimitErr)
	
	body, _ := json.Marshal(signupReq)
	req := httptest.NewRequest("POST", "/auth/signup", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	assert.Contains(t, w.Body.String(), "RATE_LIMIT_EXCEEDED")
	
	mockRateLimiter.AssertExpectations(t)
}

// Login Tests

func TestLogin_Success(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)
	
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}
	
	user := createTestUser("user-123", "test@example.com", "Test User", models.AccountTypeFull, models.UserRoleUser)
	expiresAt := time.Now().Add(24 * time.Hour)
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "login:test@example.com", services.ActionCreatePOI).Return(nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockUserService.On("VerifyPassword", mock.Anything, "user-123", "Password123!").Return(nil)
	mockAuthService.On("GenerateJWT", "user-123", "test@example.com", models.UserRoleUser).Return("test-token", expiresAt, nil)
	
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response AuthResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "test-token", response.Token)
	assert.Equal(t, "user-123", response.User.ID)
	
	mockRateLimiter.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials_UserNotFound(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)
	
	loginReq := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "Password123!",
	}
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "login:nonexistent@example.com", services.ActionCreatePOI).Return(nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "nonexistent@example.com").Return(nil, errors.New("user not found"))
	
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_CREDENTIALS")
	
	mockRateLimiter.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestLogin_InvalidCredentials_WrongPassword(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)
	
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "WrongPassword123!",
	}
	
	user := createTestUser("user-123", "test@example.com", "Test User", models.AccountTypeFull, models.UserRoleUser)
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "login:test@example.com", services.ActionCreatePOI).Return(nil)
	mockUserService.On("GetUserByEmail", mock.Anything, "test@example.com").Return(user, nil)
	mockUserService.On("VerifyPassword", mock.Anything, "user-123", "WrongPassword123!").Return(errors.New("invalid password"))
	
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_CREDENTIALS")
	
	mockRateLimiter.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

func TestLogin_RateLimited(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/login", handler.Login)
	
	loginReq := LoginRequest{
		Email:    "test@example.com",
		Password: "Password123!",
	}
	
	rateLimitErr := &services.RateLimitError{
		UserID:     "test@example.com",
		Action:     services.ActionCreatePOI,
		Limit:      100,
		RetryAfter: 3600,
	}
	
	mockRateLimiter.On("CheckRateLimit", mock.Anything, "login:test@example.com", services.ActionCreatePOI).Return(rateLimitErr)
	
	body, _ := json.Marshal(loginReq)
	req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
	
	mockRateLimiter.AssertExpectations(t)
}

// Logout Tests

func TestLogout_Success(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.POST("/auth/logout", handler.Logout)
	
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Logged out successfully")
}

// GetCurrentUser Tests

func TestGetCurrentUser_Success(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	
	// Middleware to set userID in context (simulating auth middleware)
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("userID", "user-123")
		handler.GetCurrentUser(c)
	})
	
	user := createTestUser("user-123", "test@example.com", "Test User", models.AccountTypeFull, models.UserRoleUser)
	
	mockUserService.On("GetUser", mock.Anything, "user-123").Return(user, nil)
	
	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response UserResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	
	assert.Equal(t, "user-123", response.ID)
	assert.Equal(t, "test@example.com", response.Email)
	assert.Equal(t, "Test User", response.DisplayName)
	
	mockUserService.AssertExpectations(t)
}

func TestGetCurrentUser_Unauthorized(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	router.GET("/auth/me", handler.GetCurrentUser)
	
	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "UNAUTHORIZED")
}

func TestGetCurrentUser_UserNotFound(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockAuthUserService{}
	mockRateLimiter := &MockAuthRateLimiter{}
	
	handler := NewAuthHandler(mockAuthService, mockUserService, mockRateLimiter)
	router := setupAuthTestRouter()
	
	router.GET("/auth/me", func(c *gin.Context) {
		c.Set("userID", "nonexistent-user")
		handler.GetCurrentUser(c)
	})
	
	mockUserService.On("GetUser", mock.Anything, "nonexistent-user").Return(nil, errors.New("user not found"))
	
	req := httptest.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "USER_NOT_FOUND")
	
	mockUserService.AssertExpectations(t)
}
