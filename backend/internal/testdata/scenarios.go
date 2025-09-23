package testdata

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"breakoutglobe/internal/handlers"
	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"breakoutglobe/internal/websocket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// POITestScenario provides a fluent API for testing POI-related functionality
type POITestScenario struct {
	mockSetup *MockSetup
	userID    uuid.UUID
	mapID     uuid.UUID
	handler   *handlers.POIHandler
	router    *gin.Engine
}

// NewPOITestScenario creates a new POI test scenario with sensible defaults
func NewPOITestScenario() *POITestScenario {
	mockSetup := NewMockSetup()
	
	scenario := &POITestScenario{
		mockSetup: mockSetup,
		userID:    GenerateUUID(),
		mapID:     GenerateUUID(),
	}
	
	// Create handler with mocks
	scenario.handler = handlers.NewPOIHandler(
		mockSetup.POIService.Mock(),
		mockSetup.RateLimiter.Mock(),
	)
	
	// Setup router
	gin.SetMode(gin.TestMode)
	scenario.router = gin.New()
	scenario.handler.RegisterRoutes(scenario.router)
	
	return scenario
}

// WithUser sets a custom user ID for the scenario
func (s *POITestScenario) WithUser(userID uuid.UUID) *POITestScenario {
	s.userID = userID
	return s
}

// WithMap sets a custom map ID for the scenario
func (s *POITestScenario) WithMap(mapID uuid.UUID) *POITestScenario {
	s.mapID = mapID
	return s
}

// ExpectRateLimitSuccess sets up the rate limiter to allow the request for CREATE action
func (s *POITestScenario) ExpectRateLimitSuccess() *POITestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreatePOI).
		Returns()
	
	s.mockSetup.RateLimiter.ExpectGetRateLimitHeaders().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreatePOI).
		Returns(map[string]string{
			"X-RateLimit-Limit":     "5",
			"X-RateLimit-Remaining": "4",
		})
	
	return s
}

// ExpectJoinRateLimitSuccess sets up the rate limiter to allow JOIN requests
func (s *POITestScenario) ExpectJoinRateLimitSuccess() *POITestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionJoinPOI).
		Returns()
	
	return s
}

// ExpectJoinRateLimitSuccessWithHeaders sets up rate limiter for successful join (with headers)
func (s *POITestScenario) ExpectJoinRateLimitSuccessWithHeaders() *POITestScenario {
	s.ExpectJoinRateLimitSuccess()
	
	s.mockSetup.RateLimiter.ExpectGetRateLimitHeaders().
		WithUserID(s.userID.String()).
		WithAction(services.ActionJoinPOI).
		Returns(map[string]string{
			"X-RateLimit-Limit":     "20",
			"X-RateLimit-Remaining": "19",
		})
	
	return s
}

// ExpectRateLimitExceeded sets up the rate limiter to reject the request
func (s *POITestScenario) ExpectRateLimitExceeded() *POITestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreatePOI).
		ReturnsRateLimitExceeded()
	
	return s
}

// ExpectCreationSuccess sets up the POI service to successfully create a POI
func (s *POITestScenario) ExpectCreationSuccess(expectedPOI *models.POI) *POITestScenario {
	s.mockSetup.POIService.ExpectCreatePOI().
		WithMapID(s.mapID.String()).
		WithCreatedBy(s.userID.String()).
		Returns(expectedPOI)
	
	return s
}

// ExpectJoinSuccess sets up the POI service to successfully join a POI
func (s *POITestScenario) ExpectJoinSuccess() *POITestScenario {
	// Mock JoinPOI method
	s.mockSetup.POIService.Mock().On("JoinPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"), // poiID
		s.userID.String(),
	).Return(nil)
	
	return s
}

// ExpectCapacityExceeded sets up the POI service to return capacity exceeded error
func (s *POITestScenario) ExpectCapacityExceeded() *POITestScenario {
	s.mockSetup.POIService.Mock().On("JoinPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"),
		s.userID.String(),
	).Return(fmt.Errorf("POI capacity exceeded"))
	
	return s
}

// ExpectNotFound sets up the POI service to return not found error
func (s *POITestScenario) ExpectNotFound() *POITestScenario {
	s.mockSetup.POIService.Mock().On("JoinPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"),
		s.userID.String(),
	).Return(fmt.Errorf("POI not found"))
	
	return s
}

// ExpectGetSuccess sets up the POI service to successfully retrieve a POI
func (s *POITestScenario) ExpectGetSuccess(expectedPOI *models.POI) *POITestScenario {
	s.mockSetup.POIService.ExpectGetPOI(expectedPOI.ID).Returns(expectedPOI)
	return s
}

// ExpectGetNotFound sets up the POI service to return not found for GET
func (s *POITestScenario) ExpectGetNotFound() *POITestScenario {
	s.mockSetup.POIService.Mock().On("GetPOI", 
		mock.Anything, 
		mock.AnythingOfType("string"),
	).Return(nil, gorm.ErrRecordNotFound)
	
	return s
}

// CreatePOI executes a POI creation request and returns the response
func (s *POITestScenario) CreatePOI(t TestingT, request CreatePOIRequest) *CreatePOIResponse {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
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
	var response CreatePOIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// CreatePOIExpectingError executes a POI creation request expecting an error
func (s *POITestScenario) CreatePOIExpectingError(t TestingT, request CreatePOIRequest) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/pois", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// JoinPOI executes a POI join request
func (s *POITestScenario) JoinPOI(t TestingT, poiID string, request JoinPOIRequest) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	url := fmt.Sprintf("/api/pois/%s/join", poiID)
	req := httptest.NewRequest(http.MethodPost, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// GetPOI executes a POI retrieval request and returns the response
func (s *POITestScenario) GetPOI(t TestingT, poiID string) *GetPOIResponse {
	t.Helper()
	
	url := fmt.Sprintf("/api/pois/%s", poiID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	// Verify success status
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			http.StatusOK, recorder.Code, recorder.Body.String())
		return nil
	}
	
	// Parse response
	var response GetPOIResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// GetPOIExpectingError executes a POI retrieval request expecting an error
func (s *POITestScenario) GetPOIExpectingError(t TestingT, poiID string) *httptest.ResponseRecorder {
	t.Helper()
	
	url := fmt.Sprintf("/api/pois/%s", poiID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// Request/Response types for POI scenarios

// CreatePOIRequest represents a POI creation request
type CreatePOIRequest struct {
	MapID           string        `json:"mapId"`
	Name            string        `json:"name"`
	Description     string        `json:"description"`
	Position        models.LatLng `json:"position"`
	CreatedBy       string        `json:"createdBy"`
	MaxParticipants int           `json:"maxParticipants"`
}

// CreatePOIResponse is defined in assertions.go

// JoinPOIRequest represents a POI join request
type JoinPOIRequest struct {
	UserID string `json:"userId"`
}

// GetPOIResponse represents a POI retrieval response
type GetPOIResponse struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Position    models.LatLng `json:"position"`
	CreatedBy   string        `json:"createdBy"`
}

// SessionTestScenario provides a fluent API for testing Session-related functionality
type SessionTestScenario struct {
	mockSetup *MockSetup
	userID    uuid.UUID
	mapID     uuid.UUID
	handler   *handlers.SessionHandler
	router    *gin.Engine
}

// NewSessionTestScenario creates a new Session test scenario with sensible defaults
func NewSessionTestScenario() *SessionTestScenario {
	mockSetup := NewMockSetup()
	
	scenario := &SessionTestScenario{
		mockSetup: mockSetup,
		userID:    GenerateUUID(),
		mapID:     GenerateUUID(),
	}
	
	// Create handler with mocks
	scenario.handler = handlers.NewSessionHandler(
		mockSetup.SessionService.Mock(),
		mockSetup.RateLimiter.Mock(),
	)
	
	// Setup router
	gin.SetMode(gin.TestMode)
	scenario.router = gin.New()
	scenario.handler.RegisterRoutes(scenario.router)
	
	return scenario
}

// WithUser sets a custom user ID for the scenario
func (s *SessionTestScenario) WithUser(userID uuid.UUID) *SessionTestScenario {
	s.userID = userID
	return s
}

// WithMap sets a custom map ID for the scenario
func (s *SessionTestScenario) WithMap(mapID uuid.UUID) *SessionTestScenario {
	s.mapID = mapID
	return s
}

// ExpectRateLimitSuccess sets up the rate limiter to allow CREATE SESSION requests
func (s *SessionTestScenario) ExpectRateLimitSuccess() *SessionTestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreateSession).
		Returns()
	
	s.mockSetup.RateLimiter.ExpectGetRateLimitHeaders().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreateSession).
		Returns(map[string]string{
			"X-RateLimit-Limit":     "10",
			"X-RateLimit-Remaining": "9",
		})
	
	return s
}

// ExpectRateLimitExceeded sets up the rate limiter to reject CREATE SESSION requests
func (s *SessionTestScenario) ExpectRateLimitExceeded() *SessionTestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionCreateSession).
		ReturnsRateLimitExceeded()
	
	return s
}

// ExpectUpdateRateLimitSuccess sets up the rate limiter to allow UPDATE AVATAR requests
func (s *SessionTestScenario) ExpectUpdateRateLimitSuccess() *SessionTestScenario {
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionUpdateAvatar).
		Returns()
	
	s.mockSetup.RateLimiter.ExpectGetRateLimitHeaders().
		WithUserID(s.userID.String()).
		WithAction(services.ActionUpdateAvatar).
		Returns(map[string]string{
			"X-RateLimit-Limit":     "60",
			"X-RateLimit-Remaining": "59",
		})
	
	return s
}

// ExpectCreationSuccess sets up the session service to successfully create a session
func (s *SessionTestScenario) ExpectCreationSuccess(expectedSession *models.Session) *SessionTestScenario {
	s.mockSetup.SessionService.ExpectCreateSession().
		WithUserID(s.userID.String()).
		WithMapID(s.mapID.String()).
		Returns(expectedSession)
	
	return s
}

// ExpectGetSuccess sets up the session service to successfully retrieve a session
func (s *SessionTestScenario) ExpectGetSuccess(expectedSession *models.Session) *SessionTestScenario {
	s.mockSetup.SessionService.ExpectGetSession(expectedSession.ID).Returns(expectedSession)
	return s
}

// ExpectGetNotFound sets up the session service to return not found for GET
func (s *SessionTestScenario) ExpectGetNotFound() *SessionTestScenario {
	s.mockSetup.SessionService.Mock().On("GetSession", 
		mock.Anything, 
		mock.AnythingOfType("string"),
	).Return(nil, gorm.ErrRecordNotFound)
	
	return s
}

// ExpectUpdatePositionSuccess sets up the session service to successfully update avatar position
func (s *SessionTestScenario) ExpectUpdatePositionSuccess() *SessionTestScenario {
	// First, the handler calls GetSession to get user ID for rate limiting
	session := NewSession().WithUser(s.userID).Build()
	s.mockSetup.SessionService.Mock().On("GetSession", 
		mock.Anything, 
		mock.AnythingOfType("string"), // sessionID
	).Return(session, nil)
	
	// Then it calls UpdateAvatarPosition
	s.mockSetup.SessionService.Mock().On("UpdateAvatarPosition", 
		mock.Anything, 
		mock.AnythingOfType("string"), // sessionID
		mock.AnythingOfType("models.LatLng"),
	).Return(nil)
	
	return s
}

// ExpectUpdatePositionNotFound sets up the session service to return not found for update
func (s *SessionTestScenario) ExpectUpdatePositionNotFound() *SessionTestScenario {
	// The handler calls GetSession first, which should return not found
	s.mockSetup.SessionService.Mock().On("GetSession", 
		mock.Anything, 
		mock.AnythingOfType("string"),
	).Return(nil, gorm.ErrRecordNotFound)
	
	return s
}

// ExpectHeartbeatSuccess sets up the session service to successfully handle heartbeat
func (s *SessionTestScenario) ExpectHeartbeatSuccess() *SessionTestScenario {
	s.mockSetup.SessionService.Mock().On("SessionHeartbeat", 
		mock.Anything, 
		mock.AnythingOfType("string"), // sessionID
	).Return(nil)
	
	return s
}

// ExpectHeartbeatNotFound sets up the session service to return not found for heartbeat
func (s *SessionTestScenario) ExpectHeartbeatNotFound() *SessionTestScenario {
	s.mockSetup.SessionService.Mock().On("SessionHeartbeat", 
		mock.Anything, 
		mock.AnythingOfType("string"),
	).Return(gorm.ErrRecordNotFound)
	
	return s
}

// CreateSession executes a session creation request and returns the response
func (s *SessionTestScenario) CreateSession(t TestingT, request CreateSessionRequest) *CreateSessionResponse {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
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
	var response CreateSessionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// CreateSessionExpectingError executes a session creation request expecting an error
func (s *SessionTestScenario) CreateSessionExpectingError(t TestingT, request CreateSessionRequest) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	req := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// GetSession executes a session retrieval request and returns the response
func (s *SessionTestScenario) GetSession(t TestingT, sessionID string) *GetSessionResponse {
	t.Helper()
	
	url := fmt.Sprintf("/api/sessions/%s", sessionID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	// Verify success status
	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d. Response: %s", 
			http.StatusOK, recorder.Code, recorder.Body.String())
		return nil
	}
	
	// Parse response
	var response GetSessionResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v. Body: %s", err, recorder.Body.String())
		return nil
	}
	
	return &response
}

// GetSessionExpectingError executes a session retrieval request expecting an error
func (s *SessionTestScenario) GetSessionExpectingError(t TestingT, sessionID string) *httptest.ResponseRecorder {
	t.Helper()
	
	url := fmt.Sprintf("/api/sessions/%s", sessionID)
	req := httptest.NewRequest(http.MethodGet, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// UpdateAvatarPosition executes an avatar position update request
func (s *SessionTestScenario) UpdateAvatarPosition(t TestingT, sessionID string, request UpdateAvatarPositionRequest) *httptest.ResponseRecorder {
	t.Helper()
	
	// Create HTTP request
	body, err := json.Marshal(request)
	if err != nil {
		t.Errorf("Failed to marshal request: %v", err)
		return nil
	}
	
	url := fmt.Sprintf("/api/sessions/%s/avatar", sessionID)
	req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// SessionHeartbeat executes a session heartbeat request
func (s *SessionTestScenario) SessionHeartbeat(t TestingT, sessionID string) *httptest.ResponseRecorder {
	t.Helper()
	
	url := fmt.Sprintf("/api/sessions/%s/heartbeat", sessionID)
	req := httptest.NewRequest(http.MethodPost, url, nil)
	recorder := httptest.NewRecorder()
	
	// Execute request
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

// Request/Response types for Session scenarios

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	UserID         string        `json:"userId"`
	MapID          string        `json:"mapId"`
	AvatarPosition models.LatLng `json:"avatarPosition"`
}

// CreateSessionResponse represents a session creation response
type CreateSessionResponse struct {
	SessionID      string        `json:"sessionId"`
	UserID         string        `json:"userId"`
	MapID          string        `json:"mapId"`
	AvatarPosition models.LatLng `json:"avatarPosition"`
	IsActive       bool          `json:"isActive"`
}

// GetSessionResponse represents a session retrieval response
type GetSessionResponse struct {
	SessionID      string        `json:"sessionId"`
	UserID         string        `json:"userId"`
	MapID          string        `json:"mapId"`
	AvatarPosition models.LatLng `json:"avatarPosition"`
	IsActive       bool          `json:"isActive"`
	CreatedAt      string        `json:"createdAt"`
	LastActive     string        `json:"lastActive"`
}

// UpdateAvatarPositionRequest represents an avatar position update request
type UpdateAvatarPositionRequest struct {
	Position models.LatLng `json:"position"`
}

// WebSocketTestScenario provides a fluent API for testing WebSocket functionality
type WebSocketTestScenario struct {
	mockSetup   *MockSetup
	sessionID   string
	userID      uuid.UUID
	mapID       uuid.UUID
	handler     *websocket.Handler
	server      *httptest.Server
	wsURL       string
}

// NewWebSocketTestScenario creates a new WebSocket test scenario with sensible defaults
func NewWebSocketTestScenario() *WebSocketTestScenario {
	mockSetup := NewMockSetup()
	
	scenario := &WebSocketTestScenario{
		mockSetup: mockSetup,
		sessionID: GenerateUUID().String(),
		userID:    GenerateUUID(),
		mapID:     GenerateUUID(),
	}
	
	// Create WebSocket handler with mocks
	scenario.handler = websocket.NewHandler(
		mockSetup.SessionService.Mock(),
		mockSetup.RateLimiter.Mock(),
	)
	
	// Setup HTTP server with WebSocket endpoint
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/ws", scenario.handler.HandleWebSocket)
	
	scenario.server = httptest.NewServer(router)
	scenario.wsURL = "ws" + strings.TrimPrefix(scenario.server.URL, "http") + "/ws"
	
	return scenario
}

// WithSession sets a custom session ID for the scenario
func (s *WebSocketTestScenario) WithSession(sessionID string) *WebSocketTestScenario {
	s.sessionID = sessionID
	return s
}

// WithUser sets a custom user ID for the scenario
func (s *WebSocketTestScenario) WithUser(userID uuid.UUID) *WebSocketTestScenario {
	s.userID = userID
	return s
}

// WithMap sets a custom map ID for the scenario
func (s *WebSocketTestScenario) WithMap(mapID uuid.UUID) *WebSocketTestScenario {
	s.mapID = mapID
	return s
}

// ExpectConnectionSuccess sets up the session service to allow WebSocket connection
func (s *WebSocketTestScenario) ExpectConnectionSuccess() *WebSocketTestScenario {
	session := NewSession().
		WithID(s.sessionID).
		WithUser(s.userID).
		WithMap(s.mapID).
		WithActive(true).
		Build()
	
	s.mockSetup.SessionService.Mock().On("GetSession", 
		mock.Anything, 
		s.sessionID,
	).Return(session, nil)
	
	return s
}

// ExpectConnectionFailure sets up the session service to reject WebSocket connection
func (s *WebSocketTestScenario) ExpectConnectionFailure() *WebSocketTestScenario {
	s.mockSetup.SessionService.Mock().On("GetSession", 
		mock.Anything, 
		s.sessionID,
	).Return(nil, fmt.Errorf("session not found"))
	
	return s
}

// ExpectHeartbeatSuccess sets up the session service to handle heartbeat successfully
func (s *WebSocketTestScenario) ExpectHeartbeatSuccess() *WebSocketTestScenario {
	s.mockSetup.SessionService.Mock().On("SessionHeartbeat", 
		mock.Anything, 
		s.sessionID,
	).Return(nil)
	
	return s
}

// ExpectHeartbeatError sets up the session service to return error for heartbeat
func (s *WebSocketTestScenario) ExpectHeartbeatError() *WebSocketTestScenario {
	s.mockSetup.SessionService.Mock().On("SessionHeartbeat", 
		mock.Anything, 
		s.sessionID,
	).Return(fmt.Errorf("heartbeat failed"))
	
	return s
}

// ExpectAvatarMoveSuccess sets up services for successful avatar movement
func (s *WebSocketTestScenario) ExpectAvatarMoveSuccess() *WebSocketTestScenario {
	// Rate limiter allows the movement
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionUpdateAvatar).
		Returns()
	
	// Session service updates position successfully
	s.mockSetup.SessionService.Mock().On("UpdateAvatarPosition", 
		mock.Anything, 
		s.sessionID,
		mock.AnythingOfType("models.LatLng"),
	).Return(nil)
	
	return s
}

// ExpectAvatarMoveRateLimited sets up rate limiter to reject avatar movement
func (s *WebSocketTestScenario) ExpectAvatarMoveRateLimited() *WebSocketTestScenario {
	// Rate limiter rejects the movement
	s.mockSetup.RateLimiter.ExpectCheckRateLimit().
		WithUserID(s.userID.String()).
		WithAction(services.ActionUpdateAvatar).
		ReturnsRateLimitExceeded()
	
	return s
}

// Connect establishes a WebSocket connection and returns the connection and welcome message
func (s *WebSocketTestScenario) Connect(t TestingT) (*ws.Conn, *WebSocketMessage) {
	t.Helper()
	
	// Setup authorization header
	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", s.sessionID))
	
	// Connect to WebSocket
	conn, _, err := ws.DefaultDialer.Dial(s.wsURL, header)
	if err != nil {
		t.Errorf("Failed to connect to WebSocket: %v", err)
		return nil, nil
	}
	
	// Read welcome message
	var welcomeMsg WebSocketMessage
	if err := conn.ReadJSON(&welcomeMsg); err != nil {
		conn.Close()
		t.Errorf("Failed to read welcome message: %v", err)
		return nil, nil
	}
	
	return conn, &welcomeMsg
}

// ConnectExpectingError attempts to connect expecting a failure
func (s *WebSocketTestScenario) ConnectExpectingError(t TestingT) error {
	t.Helper()
	
	// Setup authorization header
	header := http.Header{}
	header.Set("Authorization", fmt.Sprintf("Bearer %s", s.sessionID))
	
	// Try to connect to WebSocket
	conn, resp, err := ws.DefaultDialer.Dial(s.wsURL, header)
	if conn != nil {
		conn.Close()
	}
	
	if err == nil {
		t.Errorf("Expected connection to fail, but it succeeded")
		return fmt.Errorf("connection unexpectedly succeeded")
	}
	
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
	
	return err
}

// ConnectWithoutAuth attempts to connect without authorization header
func (s *WebSocketTestScenario) ConnectWithoutAuth(t TestingT) error {
	t.Helper()
	
	// Try to connect without authorization
	conn, resp, err := ws.DefaultDialer.Dial(s.wsURL, nil)
	if conn != nil {
		conn.Close()
	}
	
	if err == nil {
		t.Errorf("Expected connection to fail, but it succeeded")
		return fmt.Errorf("connection unexpectedly succeeded")
	}
	
	if resp != nil && resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 Unauthorized, got %d", resp.StatusCode)
	}
	
	return err
}

// SendHeartbeat sends a heartbeat message and verifies no error response
func (s *WebSocketTestScenario) SendHeartbeat(t TestingT, conn *ws.Conn) {
	t.Helper()
	
	heartbeatMsg := WebSocketMessage{
		Type: "heartbeat",
		Data: map[string]interface{}{},
	}
	
	if err := conn.WriteJSON(heartbeatMsg); err != nil {
		t.Errorf("Failed to send heartbeat: %v", err)
		return
	}
	
	// Give some time for processing
	time.Sleep(100 * time.Millisecond)
}

// SendHeartbeatExpectingError sends a heartbeat and expects an error response
func (s *WebSocketTestScenario) SendHeartbeatExpectingError(t TestingT, conn *ws.Conn) *WebSocketMessage {
	t.Helper()
	
	heartbeatMsg := WebSocketMessage{
		Type: "heartbeat",
		Data: map[string]interface{}{},
	}
	
	if err := conn.WriteJSON(heartbeatMsg); err != nil {
		t.Errorf("Failed to send heartbeat: %v", err)
		return nil
	}
	
	// Read error response
	var errorMsg WebSocketMessage
	if err := conn.ReadJSON(&errorMsg); err != nil {
		t.Errorf("Failed to read error response: %v", err)
		return nil
	}
	
	return &errorMsg
}

// SendAvatarMove sends an avatar movement message
func (s *WebSocketTestScenario) SendAvatarMove(t TestingT, conn *ws.Conn, position models.LatLng) {
	t.Helper()
	
	moveMsg := WebSocketMessage{
		Type: "avatar_move",
		Data: map[string]interface{}{
			"position": map[string]interface{}{
				"lat": position.Lat,
				"lng": position.Lng,
			},
		},
	}
	
	if err := conn.WriteJSON(moveMsg); err != nil {
		t.Errorf("Failed to send avatar move: %v", err)
		return
	}
	
	// Give some time for processing
	time.Sleep(100 * time.Millisecond)
}

// SendAvatarMoveExpectingError sends avatar movement and expects an error response
func (s *WebSocketTestScenario) SendAvatarMoveExpectingError(t TestingT, conn *ws.Conn, position models.LatLng) *WebSocketMessage {
	t.Helper()
	
	moveMsg := WebSocketMessage{
		Type: "avatar_move",
		Data: map[string]interface{}{
			"position": map[string]interface{}{
				"lat": position.Lat,
				"lng": position.Lng,
			},
		},
	}
	
	if err := conn.WriteJSON(moveMsg); err != nil {
		t.Errorf("Failed to send avatar move: %v", err)
		return nil
	}
	
	// Read error response
	var errorMsg WebSocketMessage
	if err := conn.ReadJSON(&errorMsg); err != nil {
		t.Errorf("Failed to read error response: %v", err)
		return nil
	}
	
	return &errorMsg
}

// ReceiveMessage reads a message from the WebSocket connection
func (s *WebSocketTestScenario) ReceiveMessage(t TestingT, conn *ws.Conn) *WebSocketMessage {
	t.Helper()
	
	var msg WebSocketMessage
	if err := conn.ReadJSON(&msg); err != nil {
		t.Errorf("Failed to receive message: %v", err)
		return nil
	}
	
	return &msg
}

// Cleanup closes the test server
func (s *WebSocketTestScenario) Cleanup() {
	if s.server != nil {
		s.server.Close()
	}
}

// WebSocketMessage represents a WebSocket message for testing
type WebSocketMessage struct {
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp,omitempty"`
}