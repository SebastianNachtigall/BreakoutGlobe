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
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MockSessionService is a mock implementation of SessionServiceInterface
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) CreateSession(ctx context.Context, userID, mapID string, avatarPosition models.LatLng) (*models.Session, error) {
	args := m.Called(ctx, userID, mapID, avatarPosition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	args := m.Called(ctx, sessionID, position)
	return args.Error(0)
}

func (m *MockSessionService) SessionHeartbeat(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionService) EndSession(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionService) GetActiveSessionsForMap(ctx context.Context, mapID string) ([]*models.Session, error) {
	args := m.Called(ctx, mapID)
	return args.Get(0).([]*models.Session), args.Error(1)
}

func (m *MockSessionService) CleanupExpiredSessions(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

// MockRateLimiter is a mock implementation of RateLimiterInterface
type MockRateLimiter struct {
	mock.Mock
}

func (m *MockRateLimiter) IsAllowed(ctx context.Context, userID string, action services.ActionType) (bool, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockRateLimiter) GetRemainingRequests(ctx context.Context, userID string, action services.ActionType) (int, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockRateLimiter) GetWindowResetTime(ctx context.Context, userID string, action services.ActionType) (time.Time, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(time.Time), args.Error(1)
}

func (m *MockRateLimiter) SetCustomLimit(userID string, action services.ActionType, limit services.RateLimit) {
	m.Called(userID, action, limit)
}

func (m *MockRateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockRateLimiter) GetUserStats(ctx context.Context, userID string) (*services.UserRateLimitStats, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.UserRateLimitStats), args.Error(1)
}

func (m *MockRateLimiter) CheckRateLimit(ctx context.Context, userID string, action services.ActionType) error {
	args := m.Called(ctx, userID, action)
	return args.Error(0)
}

func (m *MockRateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action services.ActionType) (map[string]string, error) {
	args := m.Called(ctx, userID, action)
	return args.Get(0).(map[string]string), args.Error(1)
}

// SessionHandlerTestSuite contains the test suite for SessionHandler
type SessionHandlerTestSuite struct {
	suite.Suite
	mockSessionService *MockSessionService
	mockRateLimiter    *MockRateLimiter
	handler            *SessionHandler
	router             *gin.Engine
}

func (suite *SessionHandlerTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	suite.mockSessionService = new(MockSessionService)
	suite.mockRateLimiter = new(MockRateLimiter)
	suite.handler = NewSessionHandler(suite.mockSessionService, suite.mockRateLimiter)
	
	// Setup router
	suite.router = gin.New()
	suite.handler.RegisterRoutes(suite.router)
}

func (suite *SessionHandlerTestSuite) TearDownTest() {
	suite.mockSessionService.AssertExpectations(suite.T())
	suite.mockRateLimiter.AssertExpectations(suite.T())
}

func (suite *SessionHandlerTestSuite) TestCreateSession() {
	// Test data
	reqBody := CreateSessionRequest{
		UserID:         "user-123",
		MapID:          "map-456",
		AvatarPosition: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	
	expectedSession := &models.Session{
		ID:         "session-789",
		UserID:     reqBody.UserID,
		MapID:      reqBody.MapID,
		AvatarPos:  reqBody.AvatarPosition,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionCreateSession).Return(nil)
	suite.mockSessionService.On("CreateSession", mock.AnythingOfType("*gin.Context"), reqBody.UserID, reqBody.MapID, reqBody.AvatarPosition).Return(expectedSession, nil)
	suite.mockRateLimiter.On("GetRateLimitHeaders", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionCreateSession).Return(map[string]string{
		"X-RateLimit-Limit":     "10",
		"X-RateLimit-Remaining": "9",
	}, nil)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusCreated, w.Code)
	
	var response CreateSessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(expectedSession.ID, response.SessionID)
	suite.Equal(expectedSession.UserID, response.UserID)
	suite.Equal(expectedSession.MapID, response.MapID)
	suite.Equal(expectedSession.AvatarPos.Lat, response.AvatarPosition.Lat)
	suite.Equal(expectedSession.AvatarPos.Lng, response.AvatarPosition.Lng)
	
	// Check rate limit headers
	suite.Equal("10", w.Header().Get("X-RateLimit-Limit"))
	suite.Equal("9", w.Header().Get("X-RateLimit-Remaining"))
}

func (suite *SessionHandlerTestSuite) TestCreateSession_RateLimited() {
	reqBody := CreateSessionRequest{
		UserID:         "user-123",
		MapID:          "map-456",
		AvatarPosition: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	
	// Mock rate limit exceeded
	rateLimitErr := &services.RateLimitError{
		UserID:     reqBody.UserID,
		Action:     services.ActionCreateSession,
		Limit:      10,
		Window:     time.Minute,
		RetryAfter: time.Minute,
	}
	
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionCreateSession).Return(rateLimitErr)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusTooManyRequests, w.Code)
	suite.Equal("60", w.Header().Get("Retry-After"))
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("RATE_LIMIT_EXCEEDED", response.Code)
	suite.Contains(response.Message, "Rate limit exceeded")
}

func (suite *SessionHandlerTestSuite) TestCreateSession_InvalidJSON() {
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INVALID_REQUEST", response.Code)
}

func (suite *SessionHandlerTestSuite) TestCreateSession_ValidationError() {
	reqBody := CreateSessionRequest{
		UserID:         "", // Invalid: empty user ID
		MapID:          "map-456",
		AvatarPosition: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INVALID_REQUEST", response.Code)
}

func (suite *SessionHandlerTestSuite) TestCreateSession_ServiceError() {
	reqBody := CreateSessionRequest{
		UserID:         "user-123",
		MapID:          "map-456",
		AvatarPosition: models.LatLng{Lat: 40.7128, Lng: -74.0060},
	}
	
	// Mock expectations
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), reqBody.UserID, services.ActionCreateSession).Return(nil)
	suite.mockSessionService.On("CreateSession", mock.AnythingOfType("*gin.Context"), reqBody.UserID, reqBody.MapID, reqBody.AvatarPosition).Return((*models.Session)(nil), errors.New("service error"))
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusInternalServerError, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INTERNAL_ERROR", response.Code)
}

func (suite *SessionHandlerTestSuite) TestGetSession() {
	sessionID := "session-789"
	expectedSession := &models.Session{
		ID:         sessionID,
		UserID:     "user-123",
		MapID:      "map-456",
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
	}
	
	// Mock expectations
	suite.mockSessionService.On("GetSession", mock.AnythingOfType("*gin.Context"), sessionID).Return(expectedSession, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response GetSessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(expectedSession.ID, response.SessionID)
	suite.Equal(expectedSession.UserID, response.UserID)
	suite.Equal(expectedSession.MapID, response.MapID)
	suite.Equal(expectedSession.AvatarPos.Lat, response.AvatarPosition.Lat)
	suite.Equal(expectedSession.AvatarPos.Lng, response.AvatarPosition.Lng)
}

func (suite *SessionHandlerTestSuite) TestGetSession_NotFound() {
	sessionID := "non-existent-session"
	
	// Mock expectations
	suite.mockSessionService.On("GetSession", mock.AnythingOfType("*gin.Context"), sessionID).Return((*models.Session)(nil), gorm.ErrRecordNotFound)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("SESSION_NOT_FOUND", response.Code)
}

func (suite *SessionHandlerTestSuite) TestUpdateAvatarPosition() {
	sessionID := "session-789"
	reqBody := UpdateAvatarPositionRequest{
		Position: models.LatLng{Lat: 41.0, Lng: -75.0},
	}
	
	existingSession := &models.Session{
		ID:         sessionID,
		UserID:     "user-123",
		MapID:      "map-456",
		AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
		IsActive:   true,
	}
	
	// Mock expectations
	suite.mockSessionService.On("GetSession", mock.AnythingOfType("*gin.Context"), sessionID).Return(existingSession, nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.AnythingOfType("*gin.Context"), existingSession.UserID, services.ActionUpdateAvatar).Return(nil)
	suite.mockSessionService.On("UpdateAvatarPosition", mock.AnythingOfType("*gin.Context"), sessionID, reqBody.Position).Return(nil)
	suite.mockRateLimiter.On("GetRateLimitHeaders", mock.AnythingOfType("*gin.Context"), existingSession.UserID, services.ActionUpdateAvatar).Return(map[string]string{
		"X-RateLimit-Limit":     "60",
		"X-RateLimit-Remaining": "59",
	}, nil)
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/sessions/"+sessionID+"/avatar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response UpdateAvatarPositionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
	
	// Check rate limit headers
	suite.Equal("60", w.Header().Get("X-RateLimit-Limit"))
	suite.Equal("59", w.Header().Get("X-RateLimit-Remaining"))
}

func (suite *SessionHandlerTestSuite) TestUpdateAvatarPosition_InvalidPosition() {
	sessionID := "session-789"
	reqBody := UpdateAvatarPositionRequest{
		Position: models.LatLng{Lat: 91.0, Lng: -75.0}, // Invalid latitude
	}
	
	// Create request
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPut, "/api/sessions/"+sessionID+"/avatar", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("VALIDATION_ERROR", response.Code)
}

func (suite *SessionHandlerTestSuite) TestSessionHeartbeat() {
	sessionID := "session-789"
	
	// Mock expectations
	suite.mockSessionService.On("SessionHeartbeat", mock.AnythingOfType("*gin.Context"), sessionID).Return(nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodPost, "/api/sessions/"+sessionID+"/heartbeat", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response SessionHeartbeatResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
}

func (suite *SessionHandlerTestSuite) TestEndSession() {
	sessionID := "session-789"
	
	// Mock expectations
	suite.mockSessionService.On("EndSession", mock.AnythingOfType("*gin.Context"), sessionID).Return(nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/"+sessionID, nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response EndSessionResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.True(response.Success)
}

func (suite *SessionHandlerTestSuite) TestGetActiveSessionsForMap() {
	mapID := "map-456"
	expectedSessions := []*models.Session{
		{
			ID:         "session-1",
			UserID:     "user-1",
			MapID:      mapID,
			AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060},
			IsActive:   true,
		},
		{
			ID:         "session-2",
			UserID:     "user-2",
			MapID:      mapID,
			AvatarPos:  models.LatLng{Lat: 41.0, Lng: -75.0},
			IsActive:   true,
		},
	}
	
	// Mock expectations
	suite.mockSessionService.On("GetActiveSessionsForMap", mock.AnythingOfType("*gin.Context"), mapID).Return(expectedSessions, nil)
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/maps/"+mapID+"/sessions", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response GetActiveSessionsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal(len(expectedSessions), len(response.Sessions))
	suite.Equal(expectedSessions[0].ID, response.Sessions[0].SessionID)
	suite.Equal(expectedSessions[1].ID, response.Sessions[1].SessionID)
}

func TestSessionHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(SessionHandlerTestSuite))
}