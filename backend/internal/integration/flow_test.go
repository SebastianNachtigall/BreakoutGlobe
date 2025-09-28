package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"breakoutglobe/internal/handlers"
	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"
	"breakoutglobe/internal/repository"
	"breakoutglobe/internal/services"
	"breakoutglobe/internal/testdata"
	"breakoutglobe/internal/websocket"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// FlowTestEnvironment provides a complete integration testing environment
type FlowTestEnvironment struct {
	t              testing.TB
	db             *testdata.TestDB
	redis          *testdata.TestRedis
	websocket      *testdata.TestWebSocket
	fixtures       *testdata.TestFixtures
	testData       *testdata.BasicTestData
	server         *httptest.Server
	router         *gin.Engine
	poiService     *services.POIService
	sessionService *services.SessionService
	poiHandler     *handlers.POIHandler
	sessionHandler *handlers.SessionHandler
	wsHandler      *websocket.Handler
}

// SetupFlowTest creates a complete integration testing environment
func SetupFlowTest(t testing.TB) *FlowTestEnvironment {
	t.Helper()

	if testing.Short() {
		t.Skip("Skipping flow integration test in short mode")
	}

	// Setup database
	testDB := testdata.Setup(t)

	// Setup Redis
	testRedis := testdata.SetupRedis(t)
	if testRedis == nil {
		t.Fatal("Failed to setup Redis for integration tests")
	}

	// Setup WebSocket
	testWS := testdata.SetupWebSocket(t)

	// Setup test fixtures and basic test data
	fixtures := testdata.NewTestFixtures(testDB)
	testData := fixtures.SetupBasicTestData()

	// Create repositories
	poiRepo := repository.NewPOIRepository(testDB.DB)
	sessionRepo := repository.NewSessionRepository(testDB.DB)
	userRepo := repository.NewUserRepository(testDB.DB)

	// Create Redis components
	poiParticipants := redis.NewPOIParticipants(testRedis.Client())
	sessionPresence := redis.NewSessionPresence(testRedis.Client())
	pubsub := redis.NewPubSub(testRedis.Client())

	// Create rate limiter (mock for testing)
	rateLimiter := &MockRateLimiter{}

	// Create services
	userService := services.NewUserService(userRepo)
	poiService := services.NewPOIService(poiRepo, poiParticipants, pubsub, userService)
	sessionService := services.NewSessionService(sessionRepo, sessionPresence, pubsub)

	// Create handlers
	poiHandler := handlers.NewPOIHandler(poiService, userService, rateLimiter)
	sessionHandler := handlers.NewSessionHandler(sessionService, rateLimiter)

	// Create WebSocket handler
	wsHandler := websocket.NewHandler(sessionService, rateLimiter, userService, poiService)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	poiHandler.RegisterRoutes(router)
	sessionHandler.RegisterRoutes(router)
	router.GET("/ws", wsHandler.HandleWebSocket)

	// Create test server
	server := httptest.NewServer(router)

	env := &FlowTestEnvironment{
		t:              t,
		db:             testDB,
		redis:          testRedis,
		websocket:      testWS,
		fixtures:       fixtures,
		testData:       testData,
		server:         server,
		router:         router,
		poiService:     poiService,
		sessionService: sessionService,
		poiHandler:     poiHandler,
		sessionHandler: sessionHandler,
		wsHandler:      wsHandler,
	}

	// Register cleanup
	t.Cleanup(func() {
		env.Cleanup()
	})

	return env
}

// Cleanup cleans up all test resources
func (env *FlowTestEnvironment) Cleanup() {
	if env.server != nil {
		env.server.Close()
	}
	// Other cleanup is handled by individual component cleanup
}

// CreatePOIRequest represents a POI creation request
type CreatePOIRequest struct {
	MapID           string  `json:"mapId"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Position        LatLng  `json:"position"`
	CreatedBy       string  `json:"createdBy"`
	MaxParticipants int     `json:"maxParticipants"`
}

// LatLng represents a geographic coordinate
type LatLng struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CreateSessionRequest represents a session creation request
type CreateSessionRequest struct {
	UserID         string `json:"userId"`
	MapID          string `json:"mapId"`
	AvatarPosition LatLng `json:"avatarPosition"`
}

// UpdateAvatarRequest represents an avatar position update request
type UpdateAvatarRequest struct {
	Position LatLng `json:"position"`
}

// JoinPOIRequest represents a POI join request
type JoinPOIRequest struct {
	UserID string `json:"userId"`
}

// MockRateLimiter provides a mock rate limiter for testing
type MockRateLimiter struct{}

func (m *MockRateLimiter) IsAllowed(ctx context.Context, userID string, action services.ActionType) (bool, error) {
	return true, nil // Allow all actions in tests
}

func (m *MockRateLimiter) GetRemainingRequests(ctx context.Context, userID string, action services.ActionType) (int, error) {
	return 1000, nil // High limit for tests
}

func (m *MockRateLimiter) GetWindowResetTime(ctx context.Context, userID string, action services.ActionType) (time.Time, error) {
	return time.Now().Add(time.Hour), nil
}

func (m *MockRateLimiter) SetCustomLimit(userID string, action services.ActionType, limit services.RateLimit) {
	// No-op for tests
}

func (m *MockRateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	return nil // No-op for tests
}

func (m *MockRateLimiter) GetUserStats(ctx context.Context, userID string) (*services.UserRateLimitStats, error) {
	return &services.UserRateLimitStats{}, nil
}

func (m *MockRateLimiter) CheckRateLimit(ctx context.Context, userID string, action services.ActionType) error {
	return nil // Allow all actions in tests
}

func (m *MockRateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action services.ActionType) (map[string]string, error) {
	return map[string]string{}, nil
}

// Helper methods for making HTTP requests

// POST makes a POST request to the test server
func (env *FlowTestEnvironment) POST(path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// GET makes a GET request to the test server
func (env *FlowTestEnvironment) GET(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	
	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// PUT makes a PUT request to the test server
func (env *FlowTestEnvironment) PUT(path string, body interface{}) *httptest.ResponseRecorder {
	jsonBody, _ := json.Marshal(body)
	req := httptest.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// DELETE makes a DELETE request to the test server
func (env *FlowTestEnvironment) DELETE(path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	
	recorder := httptest.NewRecorder()
	env.router.ServeHTTP(recorder, req)
	return recorder
}

// AssertHTTPSuccess asserts that an HTTP response is successful
func (env *FlowTestEnvironment) AssertHTTPSuccess(recorder *httptest.ResponseRecorder) {
	env.t.Helper()
	if recorder.Code < 200 || recorder.Code >= 300 {
		env.t.Errorf("Expected successful HTTP response, got %d: %s", recorder.Code, recorder.Body.String())
	}
}

// AssertHTTPError asserts that an HTTP response is an error
func (env *FlowTestEnvironment) AssertHTTPError(recorder *httptest.ResponseRecorder, expectedCode int) {
	env.t.Helper()
	if recorder.Code != expectedCode {
		env.t.Errorf("Expected HTTP error %d, got %d: %s", expectedCode, recorder.Code, recorder.Body.String())
	}
}

// ParseJSONResponse parses a JSON response into the provided interface
func (env *FlowTestEnvironment) ParseJSONResponse(recorder *httptest.ResponseRecorder, target interface{}) {
	env.t.Helper()
	err := json.Unmarshal(recorder.Body.Bytes(), target)
	if err != nil {
		env.t.Errorf("Failed to parse JSON response: %v", err)
	}
}

// WaitForAsyncOperations waits for async operations to complete
func (env *FlowTestEnvironment) WaitForAsyncOperations() {
	time.Sleep(100 * time.Millisecond)
}

// AssertDatabasePOI asserts that a POI exists in the database with expected properties
func (env *FlowTestEnvironment) AssertDatabasePOI(poiID string, expectedName string) {
	env.t.Helper()
	
	var poi models.POI
	err := env.db.DB.Where("id = ?", poiID).First(&poi).Error
	require.NoError(env.t, err, "POI should exist in database")
	assert.Equal(env.t, expectedName, poi.Name)
}

// AssertDatabaseSession asserts that a session exists in the database
func (env *FlowTestEnvironment) AssertDatabaseSession(sessionID string, expectedUserID string) {
	env.t.Helper()
	
	var session models.Session
	err := env.db.DB.Where("id = ?", sessionID).First(&session).Error
	require.NoError(env.t, err, "Session should exist in database")
	assert.Equal(env.t, expectedUserID, session.UserID)
}

// AssertRedisParticipant asserts that a user is a participant in Redis
func (env *FlowTestEnvironment) AssertRedisParticipant(poiID, userID string) {
	env.t.Helper()
	env.redis.AssertSetContains("poi:participants:"+poiID, userID)
}

// AssertRedisPresence asserts that a session has presence in Redis
func (env *FlowTestEnvironment) AssertRedisPresence(sessionID string) {
	env.t.Helper()
	env.redis.AssertKeyExists("session:" + sessionID)
}