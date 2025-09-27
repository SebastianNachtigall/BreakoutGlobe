package websocket

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockSessionService for testing
type MockSessionService struct {
	mock.Mock
}

func (m *MockSessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Session), args.Error(1)
}

func (m *MockSessionService) SessionHeartbeat(ctx context.Context, sessionID string) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

func (m *MockSessionService) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	args := m.Called(ctx, sessionID)
	return args.Error(0)
}

// MockRateLimiter for testing
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

// MockPOIService for testing
type MockPOIService struct {
	mock.Mock
}

func (m *MockPOIService) JoinPOI(ctx context.Context, poiID, userID string) error {
	args := m.Called(ctx, poiID, userID)
	return args.Error(0)
}

func (m *MockPOIService) LeavePOI(ctx context.Context, poiID, userID string) error {
	args := m.Called(ctx, poiID, userID)
	return args.Error(0)
}

// WebSocketHandlerTestSuite contains the test suite for WebSocket handler
type WebSocketHandlerTestSuite struct {
	suite.Suite
	mockSessionService *MockSessionService
	mockRateLimiter    *MockRateLimiter
	handler            *Handler
	server             *httptest.Server
	wsURL              string
}

func (suite *WebSocketHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	
	suite.mockSessionService = new(MockSessionService)
	suite.mockRateLimiter = new(MockRateLimiter)
	
	// Create handler with mock POI service
	mockPOIService := &MockPOIService{}
	suite.handler = NewHandler(suite.mockSessionService, suite.mockRateLimiter, nil, mockPOIService)
	
	// Setup test server
	router := gin.New()
	router.GET("/ws", suite.handler.HandleWebSocket)
	
	suite.server = httptest.NewServer(router)
	suite.wsURL = "ws" + strings.TrimPrefix(suite.server.URL, "http") + "/ws"
}

func (suite *WebSocketHandlerTestSuite) TearDownTest() {
	suite.server.Close()
	suite.mockSessionService.AssertExpectations(suite.T())
	suite.mockRateLimiter.AssertExpectations(suite.T())
}

func (suite *WebSocketHandlerTestSuite) TestWebSocketConnection_Success() {
	// Mock session validation
	session := &models.Session{
		ID:     "session-123",
		UserID: "user-456",
		MapID:  "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	
	// Connect to WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Should receive welcome message
	var msg Message
	err = conn.ReadJSON(&msg)
	suite.NoError(err)
	suite.Equal("welcome", msg.Type)
	suite.Equal("session-123", msg.Data.(map[string]interface{})["sessionId"])
}

func (suite *WebSocketHandlerTestSuite) TestWebSocketConnection_InvalidSession() {
	// Mock session validation failure
	suite.mockSessionService.On("GetSession", mock.Anything, "invalid-session").Return((*models.Session)(nil), assert.AnError)
	
	// Try to connect with invalid session
	header := http.Header{}
	header.Set("Authorization", "Bearer invalid-session")
	
	conn, resp, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	if conn != nil {
		conn.Close()
	}
	
	// Should fail with unauthorized
	suite.Error(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (suite *WebSocketHandlerTestSuite) TestWebSocketConnection_MissingAuth() {
	// Try to connect without authorization
	conn, resp, err := ws.DefaultDialer.Dial(suite.wsURL, nil)
	if conn != nil {
		conn.Close()
	}
	
	// Should fail with unauthorized
	suite.Error(err)
	suite.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func (suite *WebSocketHandlerTestSuite) TestAvatarMovement_Success() {
	// Setup connection
	session := &models.Session{
		ID:     "session-123",
		UserID: "user-456",
		MapID:  "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionUpdateAvatar).Return(nil)
	suite.mockSessionService.On("UpdateAvatarPosition", mock.Anything, "session-123").Return(nil)
	
	// Connect
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Send avatar movement
	moveMsg := Message{
		Type: "avatar_move",
		Data: map[string]interface{}{
			"position": map[string]float64{
				"lat": 40.7128,
				"lng": -74.0060,
			},
		},
	}
	
	err = conn.WriteJSON(moveMsg)
	suite.NoError(err)
	
	// Should receive acknowledgment
	var ackMsg Message
	err = conn.ReadJSON(&ackMsg)
	suite.NoError(err)
	suite.Equal("avatar_move_ack", ackMsg.Type)
}

func (suite *WebSocketHandlerTestSuite) TestAvatarMovement_RateLimited() {
	// Setup connection
	session := &models.Session{
		ID:     "session-123",
		UserID: "user-456",
		MapID:  "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	
	// Mock rate limit exceeded
	rateLimitErr := &services.RateLimitError{
		UserID:     "user-456",
		Action:     services.ActionUpdateAvatar,
		Limit:      10,
		Window:     time.Minute,
		RetryAfter: time.Minute,
	}
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionUpdateAvatar).Return(rateLimitErr)
	
	// Connect
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Send avatar movement
	moveMsg := Message{
		Type: "avatar_move",
		Data: map[string]interface{}{
			"position": map[string]float64{
				"lat": 40.7128,
				"lng": -74.0060,
			},
		},
	}
	
	err = conn.WriteJSON(moveMsg)
	suite.NoError(err)
	
	// Should receive rate limit error
	var errorMsg Message
	err = conn.ReadJSON(&errorMsg)
	suite.NoError(err)
	suite.Equal("error", errorMsg.Type)
	suite.Contains(errorMsg.Data.(map[string]interface{})["message"], "rate limit")
}

func (suite *WebSocketHandlerTestSuite) TestHeartbeat() {
	// Setup connection
	session := &models.Session{
		ID:     "session-123",
		UserID: "user-456",
		MapID:  "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	suite.mockSessionService.On("SessionHeartbeat", mock.Anything, "session-123").Return(nil)
	
	// Connect
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Send heartbeat
	heartbeatMsg := Message{
		Type: "heartbeat",
		Data: map[string]interface{}{},
	}
	
	err = conn.WriteJSON(heartbeatMsg)
	suite.NoError(err)
	
	// Should receive pong
	var pongMsg Message
	err = conn.ReadJSON(&pongMsg)
	suite.NoError(err)
	suite.Equal("pong", pongMsg.Type)
}

func (suite *WebSocketHandlerTestSuite) TestConnectionCleanup() {
	// Setup connection
	session := &models.Session{
		ID:     "session-123",
		UserID: "user-456",
		MapID:  "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	
	// Connect
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Verify client is registered
	suite.True(suite.handler.manager.IsClientConnected("session-123"))
	
	// Close connection
	conn.Close()
	
	// Wait for cleanup
	time.Sleep(100 * time.Millisecond)
	
	// Verify client is unregistered
	suite.False(suite.handler.manager.IsClientConnected("session-123"))
}

func (suite *WebSocketHandlerTestSuite) TestBroadcastToMap() {
	// Setup two connections for the same map
	session1 := &models.Session{
		ID:     "session-1",
		UserID: "user-1",
		MapID:  "map-789",
		IsActive: true,
	}
	session2 := &models.Session{
		ID:     "session-2",
		UserID: "user-2",
		MapID:  "map-789",
		IsActive: true,
	}
	
	suite.mockSessionService.On("GetSession", mock.Anything, "session-1").Return(session1, nil)
	suite.mockSessionService.On("GetSession", mock.Anything, "session-2").Return(session2, nil)
	
	// Connect first client
	header1 := http.Header{}
	header1.Set("Authorization", "Bearer session-1")
	conn1, _, err := ws.DefaultDialer.Dial(suite.wsURL, header1)
	suite.NoError(err)
	defer conn1.Close()
	
	// Connect second client
	header2 := http.Header{}
	header2.Set("Authorization", "Bearer session-2")
	conn2, _, err := ws.DefaultDialer.Dial(suite.wsURL, header2)
	suite.NoError(err)
	defer conn2.Close()
	
	// Read welcome messages
	var welcomeMsg1, welcomeMsg2 Message
	conn1.ReadJSON(&welcomeMsg1)
	conn2.ReadJSON(&welcomeMsg2)
	
	// Broadcast message to map
	broadcastMsg := Message{
		Type: "test_broadcast",
		Data: map[string]interface{}{
			"message": "Hello map!",
		},
	}
	
	err = suite.handler.manager.BroadcastToMap("map-789", broadcastMsg)
	suite.NoError(err)
	
	// Both clients should receive the broadcast
	var msg1, msg2 Message
	err = conn1.ReadJSON(&msg1)
	suite.NoError(err)
	suite.Equal("test_broadcast", msg1.Type)
	
	err = conn2.ReadJSON(&msg2)
	suite.NoError(err)
	suite.Equal("test_broadcast", msg2.Type)
}

func (suite *WebSocketHandlerTestSuite) TestPOIJoin() {
	// Setup connection
	session := &models.Session{
		ID:       "session-123",
		UserID:   "user-456",
		MapID:    "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionUpdateAvatar).Return(nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionUpdateAvatar).Return(nil)
	
	// Connect to WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Send POI join message
	joinMsg := Message{
		Type: "poi_join",
		Data: map[string]interface{}{
			"poiId": "poi-123",
		},
		Timestamp: time.Now(),
	}
	
	err = conn.WriteJSON(joinMsg)
	suite.NoError(err)
	
	// Should receive acknowledgment
	var ackMsg Message
	err = conn.ReadJSON(&ackMsg)
	suite.NoError(err)
	suite.Equal("poi_join_ack", ackMsg.Type)
	
	// Verify acknowledgment data
	ackData := ackMsg.Data.(map[string]interface{})
	suite.Equal("session-123", ackData["sessionId"])
	suite.Equal("poi-123", ackData["poiId"])
	suite.Equal(true, ackData["success"])
}

func (suite *WebSocketHandlerTestSuite) TestPOILeave() {
	// Setup connection
	session := &models.Session{
		ID:       "session-123",
		UserID:   "user-456",
		MapID:    "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionUpdateAvatar).Return(nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-456", services.ActionUpdateAvatar).Return(nil)
	
	// Connect to WebSocket
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Send POI leave message
	leaveMsg := Message{
		Type: "poi_leave",
		Data: map[string]interface{}{
			"poiId": "poi-123",
		},
		Timestamp: time.Now(),
	}
	
	err = conn.WriteJSON(leaveMsg)
	suite.NoError(err)
	
	// Should receive acknowledgment
	var ackMsg Message
	err = conn.ReadJSON(&ackMsg)
	suite.NoError(err)
	suite.Equal("poi_leave_ack", ackMsg.Type)
	
	// Verify acknowledgment data
	ackData := ackMsg.Data.(map[string]interface{})
	suite.Equal("session-123", ackData["sessionId"])
	suite.Equal("poi-123", ackData["poiId"])
	suite.Equal(true, ackData["success"])
}

func (suite *WebSocketHandlerTestSuite) TestPOIEventBroadcasting() {
	// Setup two connections for the same map
	session1 := &models.Session{
		ID:       "session-1",
		UserID:   "user-1",
		MapID:    "map-789",
		IsActive: true,
	}
	session2 := &models.Session{
		ID:       "session-2",
		UserID:   "user-2",
		MapID:    "map-789",
		IsActive: true,
	}
	
	suite.mockSessionService.On("GetSession", mock.Anything, "session-1").Return(session1, nil)
	suite.mockSessionService.On("GetSession", mock.Anything, "session-2").Return(session2, nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-1", services.ActionUpdateAvatar).Return(nil)
	suite.mockRateLimiter.On("CheckRateLimit", mock.Anything, "user-1", services.ActionUpdateAvatar).Return(nil)
	
	// Connect first client
	header1 := http.Header{}
	header1.Set("Authorization", "Bearer session-1")
	conn1, _, err := ws.DefaultDialer.Dial(suite.wsURL, header1)
	suite.NoError(err)
	defer conn1.Close()
	
	// Connect second client
	header2 := http.Header{}
	header2.Set("Authorization", "Bearer session-2")
	conn2, _, err := ws.DefaultDialer.Dial(suite.wsURL, header2)
	suite.NoError(err)
	defer conn2.Close()
	
	// Read welcome messages
	var welcomeMsg1, welcomeMsg2 Message
	conn1.ReadJSON(&welcomeMsg1)
	conn2.ReadJSON(&welcomeMsg2)
	
	// Client 1 joins a POI
	joinMsg := Message{
		Type: "poi_join",
		Data: map[string]interface{}{
			"poiId": "poi-123",
		},
		Timestamp: time.Now(),
	}
	
	err = conn1.WriteJSON(joinMsg)
	suite.NoError(err)
	
	// Client 1 should receive acknowledgment
	var ackMsg Message
	err = conn1.ReadJSON(&ackMsg)
	suite.NoError(err)
	suite.Equal("poi_join_ack", ackMsg.Type)
	
	// Client 2 should receive broadcast of the join event
	var broadcastMsg Message
	err = conn2.ReadJSON(&broadcastMsg)
	suite.NoError(err)
	suite.Equal("poi_joined", broadcastMsg.Type)
	
	// Verify broadcast data
	broadcastData := broadcastMsg.Data.(map[string]interface{})
	suite.Equal("session-1", broadcastData["sessionId"])
	suite.Equal("user-1", broadcastData["userId"])
	suite.Equal("poi-123", broadcastData["poiId"])
}

func (suite *WebSocketHandlerTestSuite) TestInvalidMessageFormat() {
	// Setup connection
	session := &models.Session{
		ID:     "session-123",
		UserID: "user-456",
		MapID:  "map-789",
		IsActive: true,
	}
	suite.mockSessionService.On("GetSession", mock.Anything, "session-123").Return(session, nil)
	
	// Connect
	header := http.Header{}
	header.Set("Authorization", "Bearer session-123")
	conn, _, err := ws.DefaultDialer.Dial(suite.wsURL, header)
	suite.NoError(err)
	defer conn.Close()
	
	// Read welcome message
	var welcomeMsg Message
	conn.ReadJSON(&welcomeMsg)
	
	// Send invalid JSON
	err = conn.WriteMessage(ws.TextMessage, []byte("invalid json"))
	suite.NoError(err)
	
	// Should receive error message or connection should close
	var errorMsg Message
	err = conn.ReadJSON(&errorMsg)
	if err == nil {
		suite.Equal("error", errorMsg.Type)
		if errorMsg.Data != nil {
			suite.Contains(errorMsg.Data.(map[string]interface{})["message"], "invalid message format")
		}
	} else {
		// Connection closed due to invalid message, which is also acceptable behavior
		suite.True(ws.IsCloseError(err, ws.CloseUnsupportedData, ws.CloseAbnormalClosure))
	}
}

func TestWebSocketHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(WebSocketHandlerTestSuite))
}

// Test helper functions
func TestExtractSessionID(t *testing.T) {
	tests := []struct {
		name        string
		header      string
		expectedID  string
		expectError bool
	}{
		{
			name:        "Valid Bearer token",
			header:      "Bearer session-123",
			expectedID:  "session-123",
			expectError: false,
		},
		{
			name:        "Missing Bearer prefix",
			header:      "session-123",
			expectedID:  "",
			expectError: true,
		},
		{
			name:        "Empty header",
			header:      "",
			expectedID:  "",
			expectError: true,
		},
		{
			name:        "Only Bearer",
			header:      "Bearer",
			expectedID:  "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sessionID, err := extractSessionID(tt.header)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, sessionID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, sessionID)
			}
		})
	}
}

func TestValidateMessage(t *testing.T) {
	tests := []struct {
		name        string
		message     Message
		expectError bool
	}{
		{
			name: "Valid avatar_move message",
			message: Message{
				Type: "avatar_move",
				Data: map[string]interface{}{
					"position": map[string]interface{}{
						"lat": 40.7128,
						"lng": -74.0060,
					},
				},
			},
			expectError: false,
		},
		{
			name: "Valid heartbeat message",
			message: Message{
				Type: "heartbeat",
				Data: map[string]interface{}{},
			},
			expectError: false,
		},
		{
			name: "Invalid message type",
			message: Message{
				Type: "invalid_type",
				Data: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "Missing position in avatar_move",
			message: Message{
				Type: "avatar_move",
				Data: map[string]interface{}{},
			},
			expectError: true,
		},
		{
			name: "Invalid position format",
			message: Message{
				Type: "avatar_move",
				Data: map[string]interface{}{
					"position": "invalid",
				},
			},
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMessage(tt.message)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}