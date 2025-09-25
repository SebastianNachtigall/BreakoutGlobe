package testdata

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	"breakoutglobe/internal/models"
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

	// Create mock session service for testing
	sessionService := &MockSessionServiceForWS{}

	// Create WebSocket handler (it creates its own manager internally)
	handler := websocket.NewHandler(sessionService, nil)

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
			
			// Handle specific test cases
			switch parts[1] {
			case "mover":
				mapID = "map-movement"
			case "observer":
				mapID = "map-movement"
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