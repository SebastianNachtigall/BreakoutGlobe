package testdata

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"breakoutglobe/internal/websocket"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

// TestWebSocket provides WebSocket integration testing infrastructure
type TestWebSocket struct {
	t        TestingT
	server   *httptest.Server
	handler  *websocket.Handler
	clients  map[string]*TestWSClient
	mutex    sync.RWMutex
	upgrader ws.Upgrader
}

// TestWSClient represents a test WebSocket client connection
type TestWSClient struct {
	SessionID   string
	UserID      string
	MapID       string
	Conn        *ws.Conn
	Messages    chan websocket.Message
	Errors      chan error
	Connected   bool
	mutex       sync.RWMutex
	stopReading chan struct{}
}

// SetupWebSocket creates a WebSocket integration testing environment
func SetupWebSocket(t TestingT) *TestWebSocket {
	t.Helper()

	// Create mock services for testing
	sessionService := &MockSessionServiceForWS{}
	rateLimiter := &MockRateLimiterForWS{}
	userService := &MockUserServiceForWS{}
	poiService := &MockPOIServiceForWS{}

	// Create WebSocket handler (it creates its own manager internally)
	handler := websocket.NewHandler(sessionService, rateLimiter, userService, poiService)

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add WebSocket endpoint
	router.GET("/ws", handler.HandleWebSocket)

	// Create test server
	server := httptest.NewServer(router)

	// Create upgrader for test clients
	upgrader := ws.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}

	testWS := &TestWebSocket{
		t:        t,
		server:   server,
		handler:  handler,
		clients:  make(map[string]*TestWSClient),
		upgrader: upgrader,
	}

	// Register cleanup
	t.Cleanup(func() {
		testWS.Cleanup()
	})

	return testWS
}

// CreateClient creates a new WebSocket test client
func (tws *TestWebSocket) CreateClient(sessionID, userID, mapID string) *TestWSClient {
	tws.t.Helper()

	// Convert HTTP URL to WebSocket URL
	wsURL := strings.Replace(tws.server.URL, "http://", "ws://", 1) + "/ws"
	u, err := url.Parse(wsURL)
	if err != nil {
		tws.t.Errorf("Failed to parse WebSocket URL: %v", err)
		return nil
	}

	// Create headers with Authorization Bearer token
	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+sessionID)

	// Create WebSocket connection with proper authentication
	conn, _, err := ws.DefaultDialer.Dial(u.String(), headers)
	if err != nil {
		tws.t.Errorf("Failed to connect WebSocket: %v", err)
		return nil
	}

	client := &TestWSClient{
		SessionID:   sessionID,
		UserID:      userID,
		MapID:       mapID,
		Conn:        conn,
		Messages:    make(chan websocket.Message, 100),
		Errors:      make(chan error, 10),
		Connected:   true,
		stopReading: make(chan struct{}),
	}

	// Start message reading goroutine
	go client.readMessages()

	// Store client
	tws.mutex.Lock()
	tws.clients[sessionID] = client
	tws.mutex.Unlock()

	return client
}

// GetClient returns a test client by session ID
func (tws *TestWebSocket) GetClient(sessionID string) *TestWSClient {
	tws.mutex.RLock()
	defer tws.mutex.RUnlock()
	return tws.clients[sessionID]
}

// GetConnectedClients returns the number of connected clients (test implementation)
func (tws *TestWebSocket) GetConnectedClients() int {
	tws.mutex.RLock()
	defer tws.mutex.RUnlock()
	
	count := 0
	for _, client := range tws.clients {
		if client.IsConnected() {
			count++
		}
	}
	return count
}

// GetMapClients returns the number of clients connected to a specific map (test implementation)
func (tws *TestWebSocket) GetMapClients(mapID string) int {
	tws.mutex.RLock()
	defer tws.mutex.RUnlock()
	
	count := 0
	for _, client := range tws.clients {
		if client.IsConnected() && client.MapID == mapID {
			count++
		}
	}
	return count
}

// BroadcastToMap broadcasts a message to all clients on a map (test implementation)
func (tws *TestWebSocket) BroadcastToMap(mapID string, message websocket.Message) {
	tws.mutex.RLock()
	defer tws.mutex.RUnlock()
	
	for _, client := range tws.clients {
		if client.IsConnected() && client.MapID == mapID {
			select {
			case client.Messages <- message:
			default:
				// Channel full, skip this client
			}
		}
	}
}

// Cleanup closes all connections and shuts down the test environment
func (tws *TestWebSocket) Cleanup() {
	// Close all test clients
	tws.mutex.Lock()
	for _, client := range tws.clients {
		client.Close()
	}
	tws.clients = make(map[string]*TestWSClient)
	tws.mutex.Unlock()

	// Note: We can't directly shutdown the handler's internal manager
	// The cleanup of test clients should be sufficient for testing

	// Close test server
	if tws.server != nil {
		tws.server.Close()
	}
}

// AssertClientConnected asserts that a client is connected
func (tws *TestWebSocket) AssertClientConnected(sessionID string) {
	tws.t.Helper()
	tws.mutex.RLock()
	client, exists := tws.clients[sessionID]
	tws.mutex.RUnlock()
	
	if !exists || !client.IsConnected() {
		tws.t.Errorf("Expected client %s to be connected, but it's not", sessionID)
	}
}

// AssertClientDisconnected asserts that a client is disconnected
func (tws *TestWebSocket) AssertClientDisconnected(sessionID string) {
	tws.t.Helper()
	tws.mutex.RLock()
	client, exists := tws.clients[sessionID]
	tws.mutex.RUnlock()
	
	if exists && client.IsConnected() {
		tws.t.Errorf("Expected client %s to be disconnected, but it's connected", sessionID)
	}
}

// AssertConnectedClientsCount asserts the total number of connected clients
func (tws *TestWebSocket) AssertConnectedClientsCount(expectedCount int) {
	tws.t.Helper()
	actualCount := tws.GetConnectedClients()
	if actualCount != expectedCount {
		tws.t.Errorf("Expected %d connected clients, got %d", expectedCount, actualCount)
	}
}

// AssertMapClientsCount asserts the number of clients connected to a specific map
func (tws *TestWebSocket) AssertMapClientsCount(mapID string, expectedCount int) {
	tws.t.Helper()
	actualCount := tws.GetMapClients(mapID)
	if actualCount != expectedCount {
		tws.t.Errorf("Expected %d clients on map %s, got %d", expectedCount, mapID, actualCount)
	}
}

// TestWSClient methods

// SendMessage sends a message to the WebSocket server
func (c *TestWSClient) SendMessage(message websocket.Message) error {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if !c.Connected {
		return fmt.Errorf("client is not connected")
	}

	return c.Conn.WriteJSON(message)
}

// ReceiveMessage waits for a message with timeout
func (c *TestWSClient) ReceiveMessage(timeout time.Duration) (websocket.Message, error) {
	select {
	case msg := <-c.Messages:
		return msg, nil
	case err := <-c.Errors:
		return websocket.Message{}, err
	case <-time.After(timeout):
		return websocket.Message{}, fmt.Errorf("timeout waiting for message")
	}
}

// ReceiveMessages waits for multiple messages with timeout
func (c *TestWSClient) ReceiveMessages(count int, timeout time.Duration) ([]websocket.Message, error) {
	messages := make([]websocket.Message, 0, count)
	deadline := time.Now().Add(timeout)

	for i := 0; i < count; i++ {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			return messages, fmt.Errorf("timeout waiting for messages, got %d of %d", len(messages), count)
		}

		msg, err := c.ReceiveMessage(remaining)
		if err != nil {
			return messages, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// ExpectMessage waits for a message and asserts its properties
func (c *TestWSClient) ExpectMessage(messageType string, timeout time.Duration) websocket.Message {
	msg, err := c.ReceiveMessage(timeout)
	if err != nil {
		panic(fmt.Sprintf("Failed to receive message: %v", err))
	}

	if msg.Type != messageType {
		panic(fmt.Sprintf("Expected message type %s, got %s", messageType, msg.Type))
	}

	return msg
}

// ExpectNoMessage asserts that no message is received within the timeout
func (c *TestWSClient) ExpectNoMessage(timeout time.Duration) {
	select {
	case msg := <-c.Messages:
		panic(fmt.Sprintf("Expected no message, but received: %+v", msg))
	case <-time.After(timeout):
		// Expected - no message received
	}
}

// ConsumeWelcomeMessage consumes the initial welcome message sent by the server
// and updates the client properties with the actual values from the server
func (c *TestWSClient) ConsumeWelcomeMessage(timeout time.Duration) error {
	msg, err := c.ReceiveMessage(timeout)
	if err != nil {
		return fmt.Errorf("failed to receive welcome message: %v", err)
	}
	
	if msg.Type != "welcome" {
		return fmt.Errorf("expected welcome message, got %s", msg.Type)
	}
	
	// Update client properties from welcome message data
	if data, ok := msg.Data.(map[string]interface{}); ok {
		if userID, ok := data["userId"].(string); ok {
			c.mutex.Lock()
			c.UserID = userID
			c.mutex.Unlock()
		}
		if mapID, ok := data["mapId"].(string); ok {
			c.mutex.Lock()
			c.MapID = mapID
			c.mutex.Unlock()
		}
	}
	
	return nil
}

// Close closes the WebSocket connection
func (c *TestWSClient) Close() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.Connected {
		c.Connected = false
		close(c.stopReading)
		if c.Conn != nil {
			c.Conn.Close()
		}
	}
}

// IsConnected returns whether the client is connected
func (c *TestWSClient) IsConnected() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.Connected
}

// readMessages reads messages from the WebSocket connection
func (c *TestWSClient) readMessages() {
	defer func() {
		c.mutex.Lock()
		c.Connected = false
		c.mutex.Unlock()
	}()

	for {
		select {
		case <-c.stopReading:
			return
		default:
			var message websocket.Message
			err := c.Conn.ReadJSON(&message)
			if err != nil {
				if !ws.IsCloseError(err, ws.CloseGoingAway, ws.CloseNormalClosure) {
					select {
					case c.Errors <- err:
					case <-c.stopReading:
						return
					}
				}
				return
			}

			select {
			case c.Messages <- message:
			case <-c.stopReading:
				return
			}
		}
	}
}

// MockSessionServiceForWS provides a mock session service for WebSocket testing
type MockSessionServiceForWS struct{}

func (m *MockSessionServiceForWS) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	// Return a mock session for testing with proper fields
	// Extract expected userID and mapID from sessionID for consistent testing
	var userID, mapID string
	
	// Parse session ID to extract user and map info for testing
	if strings.HasPrefix(sessionID, "session-") {
		// For session IDs like "session-123", "session-mover", etc.
		parts := strings.Split(sessionID, "-")
		if len(parts) >= 2 {
			userID = "user-" + parts[1]
			mapID = "map-test" // Default map for most tests
			
			// Handle specific test cases - be more specific about which tests need special map assignments
			switch sessionID {
			// WebSocket movement test specific sessions
			case "session-mover":
				mapID = "map-movement"
			case "session-observer":
				// This is used by both websocket_test.go (expects map-movement) and session_flow_test.go (expects map-movement-test)
				// For now, default to map-movement for the websocket test
				mapID = "map-movement"
			// Session flow test specific sessions  
			case "session-movement-observer":
				mapID = "map-movement-test"
			// POI flow test observers should use map-test (default)
			case "session-observer1", "session-observer2", "session-poi-observer":
				mapID = "map-test"
			// Cross-map isolation test
			case "session-other":
				mapID = "map-other"
			case "other":
				mapID = "map-other" // For cross-map isolation testing
			case "creator":
				mapID = "map-poi"
			case "p1", "p2":
				mapID = "map-poi"
			case "ordering":
				mapID = "map-ordering"
			case "lifecycle":
				mapID = "map-lifecycle"
			case "1", "2", "3":
				mapID = "map-test" // For MultiClientBroadcast test
			// New test cases for message validation
			case "welcome":
				if len(parts) >= 3 && parts[2] == "test" {
					userID = "user-welcome-test"
					mapID = "map-welcome-test"
				}
			case "type":
				if len(parts) >= 3 && parts[2] == "test" {
					userID = "user-type-test"
					mapID = "map-type-test"
				}
			case "data":
				if len(parts) >= 3 && parts[2] == "test" {
					userID = "user-data-test"
					mapID = "map-data-test"
				}
			case "invalid":
				if len(parts) >= 3 && parts[2] == "test" {
					userID = "user-invalid-test"
					mapID = "map-invalid-test"
				}
			case "sequence":
				if len(parts) >= 3 && parts[2] == "test" {
					userID = "user-sequence-test"
					mapID = "map-sequence-test"
				}
			}
			
			// Handle broadcast session IDs (broadcast-session-1, broadcast-session-2, etc.)
			if strings.HasPrefix(sessionID, "broadcast-session-") {
				parts := strings.Split(sessionID, "-")
				if len(parts) >= 3 {
					userID = "user-broadcast-session-" + parts[2]
					mapID = "map-test"
				}
			}
			
			// Handle perf session IDs (perf-session-1, perf-session-2, etc.)
			if strings.HasPrefix(sessionID, "perf-session-") {
				parts := strings.Split(sessionID, "-")
				if len(parts) >= 3 {
					userID = "user-perf-session-" + parts[2]
					mapID = "map-test"
				}
			}
			
			// Handle MapIsolation test with compound session IDs
			if len(parts) >= 3 {
				compound := parts[1] + "-" + parts[2]
				switch compound {
				case "map1-1", "map1-2":
					mapID = "map-1"
					userID = "user-" + compound
				case "map2-1":
					mapID = "map-2"
					userID = "user-" + compound
				}
			}
		}
	} else {
		// For other session ID formats
		userID = "user-" + sessionID
		mapID = "map-test"
		
		// Handle special cases for cross-layer tests
		if sessionID == "other-session" {
			userID = "user-other-session"
			mapID = "map-other"
		}
		if sessionID == "session-other" {
			userID = "user-other"
			mapID = "map-other"
		}
	}
	
	return &models.Session{
		ID:       sessionID,
		UserID:   userID,
		MapID:    mapID,
		IsActive: true, // Important: session must be active for WebSocket connection
	}, nil
}

func (m *MockSessionServiceForWS) SessionHeartbeat(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockSessionServiceForWS) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	return nil
}

// MockRateLimiterForWS provides a mock rate limiter for WebSocket testing
type MockRateLimiterForWS struct{}

func (m *MockRateLimiterForWS) IsAllowed(ctx context.Context, userID string, action services.ActionType) (bool, error) {
	return true, nil
}

func (m *MockRateLimiterForWS) CheckRateLimit(ctx context.Context, userID string, action services.ActionType) error {
	return nil
}

func (m *MockRateLimiterForWS) GetRemainingRequests(ctx context.Context, userID string, action services.ActionType) (int, error) {
	return 100, nil
}

func (m *MockRateLimiterForWS) GetWindowResetTime(ctx context.Context, userID string, action services.ActionType) (time.Time, error) {
	return time.Now().Add(time.Hour), nil
}

func (m *MockRateLimiterForWS) SetCustomLimit(userID string, action services.ActionType, limit services.RateLimit) {
	// No-op for testing
}

func (m *MockRateLimiterForWS) ClearUserLimits(ctx context.Context, userID string) error {
	return nil
}

func (m *MockRateLimiterForWS) GetUserStats(ctx context.Context, userID string) (*services.UserRateLimitStats, error) {
	return &services.UserRateLimitStats{}, nil
}

func (m *MockRateLimiterForWS) GetRateLimitHeaders(ctx context.Context, userID string, action services.ActionType) (map[string]string, error) {
	return map[string]string{}, nil
}

// MockUserServiceForWS provides a mock user service for WebSocket testing
type MockUserServiceForWS struct{}

func (m *MockUserServiceForWS) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return &models.User{
		ID:          userID,
		DisplayName: "Test User",
	}, nil
}

func (m *MockUserServiceForWS) CreateGuestProfile(ctx context.Context, displayName string) (*models.User, error) {
	return &models.User{
		ID:          "user-" + displayName,
		DisplayName: displayName,
	}, nil
}

func (m *MockUserServiceForWS) CreateGuestProfileWithAboutMe(ctx context.Context, displayName, aboutMe string) (*models.User, error) {
	return &models.User{
		ID:          "user-" + displayName,
		DisplayName: displayName,
		AboutMe:     &aboutMe,
	}, nil
}

func (m *MockUserServiceForWS) UploadAvatar(ctx context.Context, userID string, filename string, fileData []byte) (*models.User, error) {
	return &models.User{
		ID:          userID,
		DisplayName: "Test User",
	}, nil
}

func (m *MockUserServiceForWS) UpdateProfile(ctx context.Context, userID string, req *services.UpdateProfileRequest) (*models.User, error) {
	return &models.User{
		ID:          userID,
		DisplayName: "Test User",
	}, nil
}

// MockPOIServiceForWS provides a mock POI service for WebSocket testing
type MockPOIServiceForWS struct{}

func (m *MockPOIServiceForWS) CreatePOI(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int) (*models.POI, error) {
	return &models.POI{}, nil
}

func (m *MockPOIServiceForWS) CreatePOIWithImage(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int, imageFile *multipart.FileHeader) (*models.POI, error) {
	return &models.POI{}, nil
}

func (m *MockPOIServiceForWS) GetPOI(ctx context.Context, poiID string) (*models.POI, error) {
	return &models.POI{}, nil
}

func (m *MockPOIServiceForWS) GetPOIsForMap(ctx context.Context, mapID string) ([]*models.POI, error) {
	return []*models.POI{}, nil
}

func (m *MockPOIServiceForWS) GetPOIsInBounds(ctx context.Context, mapID string, bounds services.POIBounds) ([]*models.POI, error) {
	return []*models.POI{}, nil
}

func (m *MockPOIServiceForWS) UpdatePOI(ctx context.Context, poiID string, updateData services.POIUpdateData) (*models.POI, error) {
	return &models.POI{}, nil
}

func (m *MockPOIServiceForWS) DeletePOI(ctx context.Context, poiID string) error {
	return nil
}

func (m *MockPOIServiceForWS) JoinPOI(ctx context.Context, poiID, userID string) error {
	return nil
}

func (m *MockPOIServiceForWS) LeavePOI(ctx context.Context, poiID, userID string) error {
	return nil
}

func (m *MockPOIServiceForWS) GetPOIParticipants(ctx context.Context, poiID string) ([]string, error) {
	return []string{}, nil
}

func (m *MockPOIServiceForWS) GetPOIParticipantCount(ctx context.Context, poiID string) (int, error) {
	return 0, nil
}

func (m *MockPOIServiceForWS) GetPOIParticipantsWithInfo(ctx context.Context, poiID string) ([]services.POIParticipantInfo, error) {
	return []services.POIParticipantInfo{}, nil
}

func (m *MockPOIServiceForWS) GetUserPOIs(ctx context.Context, userID string) ([]string, error) {
	return []string{}, nil
}

func (m *MockPOIServiceForWS) ValidatePOI(ctx context.Context, poiID string) (*models.POI, error) {
	return &models.POI{}, nil
}

func (m *MockPOIServiceForWS) ClearAllPOIs(ctx context.Context, mapID string) error {
	return nil
}