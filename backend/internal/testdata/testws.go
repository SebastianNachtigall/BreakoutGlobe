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

	// Add query parameters for session info
	q := u.Query()
	q.Set("sessionId", sessionID)
	q.Set("userId", userID)
	q.Set("mapId", mapID)
	u.RawQuery = q.Encode()

	// Create WebSocket connection
	conn, _, err := ws.DefaultDialer.Dial(u.String(), nil)
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
	// Return a mock session for testing
	return &models.Session{
		ID:     sessionID,
		UserID: "user-" + sessionID,
		MapID:  "map-test",
	}, nil
}

func (m *MockSessionServiceForWS) SessionHeartbeat(ctx context.Context, sessionID string) error {
	return nil
}

func (m *MockSessionServiceForWS) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	return nil
}