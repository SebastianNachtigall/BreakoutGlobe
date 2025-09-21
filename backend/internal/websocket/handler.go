package websocket

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	ws "github.com/gorilla/websocket"
)

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	SessionID string
	UserID    string
	MapID     string
	Conn      *ws.Conn
	Send      chan Message
	Manager   *Manager
}

// SessionServiceInterface defines the interface for session operations
type SessionServiceInterface interface {
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	SessionHeartbeat(ctx context.Context, sessionID string) error
	UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error
}

// RateLimiterInterface defines the interface for rate limiting
type RateLimiterInterface interface {
	CheckRateLimit(ctx context.Context, userID string, action services.ActionType) error
}

// Handler handles WebSocket connections and messages
type Handler struct {
	sessionService SessionServiceInterface
	rateLimiter    RateLimiterInterface
	manager        *Manager
	upgrader       ws.Upgrader
	logger         *slog.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(sessionService SessionServiceInterface, rateLimiter RateLimiterInterface) *Handler {
	return &Handler{
		sessionService: sessionService,
		rateLimiter:    rateLimiter,
		manager:        NewManager(),
		upgrader: ws.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		logger: slog.Default(),
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Extract session ID from Authorization header
	authHeader := c.GetHeader("Authorization")
	sessionID, err := extractSessionID(authHeader)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header"})
		return
	}
	
	// Validate session
	session, err := h.sessionService.GetSession(c.Request.Context(), sessionID)
	if err != nil {
		h.logger.Warn("WebSocket connection failed: invalid session", 
			"sessionId", sessionID, 
			"error", err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
		return
	}
	
	if !session.IsActive {
		h.logger.Warn("WebSocket connection failed: inactive session", 
			"sessionId", sessionID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session is not active"})
		return
	}
	
	// Upgrade connection
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("Failed to upgrade WebSocket connection", 
			"sessionId", sessionID, 
			"error", err.Error())
		return
	}
	
	// Create client
	client := &Client{
		SessionID: sessionID,
		UserID:    session.UserID,
		MapID:     session.MapID,
		Conn:      conn,
		Send:      make(chan Message, 256),
		Manager:   h.manager,
	}
	
	// Register client
	h.manager.RegisterClient(client)
	
	h.logger.Info("WebSocket client connected", 
		"sessionId", sessionID, 
		"userId", session.UserID, 
		"mapId", session.MapID)
	
	// Send welcome message
	welcomeMsg := Message{
		Type: "welcome",
		Data: map[string]interface{}{
			"sessionId": sessionID,
			"userId":    session.UserID,
			"mapId":     session.MapID,
		},
		Timestamp: time.Now(),
	}
	client.Send <- welcomeMsg
	
	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(h)
}

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump(handler *Handler) {
	defer func() {
		c.Manager.UnregisterClient(c)
		c.Conn.Close()
	}()
	
	// Set read deadline and pong handler
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		var msg Message
		err := c.Conn.ReadJSON(&msg)
		if err != nil {
			if ws.IsUnexpectedCloseError(err, ws.CloseGoingAway, ws.CloseAbnormalClosure) {
				handler.logger.Error("WebSocket read error", 
					"sessionId", c.SessionID, 
					"error", err.Error())
			}
			break
		}
		
		msg.Timestamp = time.Now()
		
		// Validate message
		if err := validateMessage(msg); err != nil {
			errorMsg := Message{
				Type: "error",
				Data: map[string]interface{}{
					"message": "Invalid message format: " + err.Error(),
				},
				Timestamp: time.Now(),
			}
			c.Send <- errorMsg
			continue
		}
		
		// Handle message
		handler.handleMessage(c, msg)
	}
}

// writePump handles writing messages to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(ws.CloseMessage, []byte{})
				return
			}
			
			if err := c.Conn.WriteJSON(message); err != nil {
				return
			}
			
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(ws.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming WebSocket messages
func (h *Handler) handleMessage(client *Client, msg Message) {
	ctx := context.Background()
	
	switch msg.Type {
	case "heartbeat":
		h.handleHeartbeat(ctx, client, msg)
	case "avatar_move":
		h.handleAvatarMove(ctx, client, msg)
	case "poi_join":
		h.handlePOIJoin(ctx, client, msg)
	case "poi_leave":
		h.handlePOILeave(ctx, client, msg)
	default:
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": fmt.Sprintf("Unknown message type: %s", msg.Type),
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
	}
}

// handleHeartbeat processes heartbeat messages
func (h *Handler) handleHeartbeat(ctx context.Context, client *Client, msg Message) {
	// Update session heartbeat
	if err := h.sessionService.SessionHeartbeat(ctx, client.SessionID); err != nil {
		h.logger.Error("Failed to update session heartbeat", 
			"sessionId", client.SessionID, 
			"error", err.Error())
		
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Failed to update session heartbeat",
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	// Send pong response
	pongMsg := Message{
		Type: "pong",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
		},
		Timestamp: time.Now(),
	}
	client.Send <- pongMsg
}

// handleAvatarMove processes avatar movement messages
func (h *Handler) handleAvatarMove(ctx context.Context, client *Client, msg Message) {
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(ctx, client.UserID, services.ActionUpdateAvatar); err != nil {
		if rateLimitErr, ok := err.(*services.RateLimitError); ok {
			errorMsg := Message{
				Type: "error",
				Data: map[string]interface{}{
					"code":       "RATE_LIMIT_EXCEEDED",
					"message":    "Avatar movement rate limit exceeded",
					"retryAfter": rateLimitErr.RetryAfter.Seconds(),
				},
				Timestamp: time.Now(),
			}
			client.Send <- errorMsg
			return
		}
		
		h.logger.Error("Rate limit check failed", 
			"sessionId", client.SessionID, 
			"error", err.Error())
		
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Rate limit check failed",
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	// Extract position from message
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Invalid message data format",
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	positionData, ok := data["position"].(map[string]interface{})
	if !ok {
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Invalid position data format",
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	lat, ok1 := positionData["lat"].(float64)
	lng, ok2 := positionData["lng"].(float64)
	if !ok1 || !ok2 {
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Invalid position coordinates",
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	position := models.LatLng{Lat: lat, Lng: lng}
	
	// Validate position
	if err := position.Validate(); err != nil {
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Invalid position: " + err.Error(),
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	// Update avatar position
	if err := h.sessionService.UpdateAvatarPosition(ctx, client.SessionID, position); err != nil {
		h.logger.Error("Failed to update avatar position", 
			"sessionId", client.SessionID, 
			"position", position, 
			"error", err.Error())
		
		errorMsg := Message{
			Type: "error",
			Data: map[string]interface{}{
				"message": "Failed to update avatar position",
			},
			Timestamp: time.Now(),
		}
		client.Send <- errorMsg
		return
	}
	
	// Send acknowledgment
	ackMsg := Message{
		Type: "avatar_move_ack",
		Data: map[string]interface{}{
			"sessionId": client.SessionID,
			"position":  position,
		},
		Timestamp: time.Now(),
	}
	client.Send <- ackMsg
	
	// Broadcast movement to other clients in the same map
	broadcastMsg := Message{
		Type: "avatar_moved",
		Data: map[string]interface{}{
			"sessionId": client.SessionID,
			"userId":    client.UserID,
			"position":  position,
		},
		Timestamp: time.Now(),
	}
	
	// Broadcast to all clients in the same map except the sender
	h.manager.BroadcastToMapExcept(client.MapID, client.SessionID, broadcastMsg)
	
	h.logger.Debug("Avatar position updated", 
		"sessionId", client.SessionID, 
		"userId", client.UserID, 
		"position", position)
}

// Helper functions

// extractSessionID extracts session ID from Authorization header
func extractSessionID(authHeader string) (string, error) {
	if authHeader == "" {
		return "", errors.New("missing authorization header")
	}
	
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", errors.New("invalid authorization header format")
	}
	
	sessionID := strings.TrimSpace(parts[1])
	if sessionID == "" {
		return "", errors.New("empty session ID")
	}
	
	return sessionID, nil
}

// validateMessage validates incoming WebSocket messages
func validateMessage(msg Message) error {
	if msg.Type == "" {
		return errors.New("message type is required")
	}
	
	switch msg.Type {
	case "heartbeat":
		// Heartbeat messages don't need additional validation
		return nil
		
	case "avatar_move":
		// Validate avatar movement message
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		position, ok := data["position"]
		if !ok {
			return errors.New("position is required for avatar_move")
		}
		
		positionMap, ok := position.(map[string]interface{})
		if !ok {
			return errors.New("position must be an object")
		}
		
		if _, ok := positionMap["lat"]; !ok {
			return errors.New("latitude is required")
		}
		
		if _, ok := positionMap["lng"]; !ok {
			return errors.New("longitude is required")
		}
		
		return nil
		
	case "poi_join", "poi_leave":
		// Validate POI messages
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		poiID, ok := data["poiId"].(string)
		if !ok || poiID == "" {
			return errors.New("poiId is required")
		}
		
		return nil
		
	default:
		return fmt.Errorf("unknown message type: %s", msg.Type)
	}
}

// handlePOIJoin handles POI join events
func (h *Handler) handlePOIJoin(ctx context.Context, client *Client, msg Message) {
	// Validate message
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid POI join message format", "sessionId", client.SessionID)
		return
	}
	
	poiID, ok := data["poiId"].(string)
	if !ok || poiID == "" {
		h.logger.Error("Missing or invalid POI ID in join message", "sessionId", client.SessionID)
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(ctx, client.UserID, services.ActionUpdateAvatar); err != nil {
		if rateLimitErr, ok := err.(*services.RateLimitError); ok {
			errorMsg := Message{
				Type: "error",
				Data: map[string]interface{}{
					"message": "Rate limit exceeded: " + rateLimitErr.Error(),
				},
				Timestamp: time.Now(),
			}
			client.Send <- errorMsg
		}
		h.logger.Warn("POI join rate limited", "sessionId", client.SessionID, "userId", client.UserID)
		return
	}
	
	// Send acknowledgment
	ackMsg := Message{
		Type: "poi_join_ack",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sessionId": client.SessionID,
			"poiId": poiID,
			"success": true,
		},
	}
	
	select {
	case client.Send <- ackMsg:
	default:
		h.logger.Warn("Failed to send POI join acknowledgment", "sessionId", client.SessionID)
	}
	
	// Broadcast POI join event to other clients in the same map
	broadcastMsg := Message{
		Type: "poi_joined",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sessionId": client.SessionID,
			"userId": client.UserID,
			"poiId": poiID,
		},
	}
	
	h.manager.BroadcastToMapExcept(client.MapID, client.SessionID, broadcastMsg)
	
	h.logger.Info("User joined POI", "sessionId", client.SessionID, "userId", client.UserID, "poiId", poiID)
}

// handlePOILeave handles POI leave events
func (h *Handler) handlePOILeave(ctx context.Context, client *Client, msg Message) {
	// Validate message
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid POI leave message format", "sessionId", client.SessionID)
		return
	}
	
	poiID, ok := data["poiId"].(string)
	if !ok || poiID == "" {
		h.logger.Error("Missing or invalid POI ID in leave message", "sessionId", client.SessionID)
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(ctx, client.UserID, services.ActionUpdateAvatar); err != nil {
		if rateLimitErr, ok := err.(*services.RateLimitError); ok {
			errorMsg := Message{
				Type: "error",
				Data: map[string]interface{}{
					"message": "Rate limit exceeded: " + rateLimitErr.Error(),
				},
				Timestamp: time.Now(),
			}
			client.Send <- errorMsg
		}
		h.logger.Warn("POI leave rate limited", "sessionId", client.SessionID, "userId", client.UserID)
		return
	}
	
	// Send acknowledgment
	ackMsg := Message{
		Type: "poi_leave_ack",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sessionId": client.SessionID,
			"poiId": poiID,
			"success": true,
		},
	}
	
	select {
	case client.Send <- ackMsg:
	default:
		h.logger.Warn("Failed to send POI leave acknowledgment", "sessionId", client.SessionID)
	}
	
	// Broadcast POI leave event to other clients in the same map
	broadcastMsg := Message{
		Type: "poi_left",
		Timestamp: time.Now(),
		Data: map[string]interface{}{
			"sessionId": client.SessionID,
			"userId": client.UserID,
			"poiId": poiID,
		},
	}
	
	h.manager.BroadcastToMapExcept(client.MapID, client.SessionID, broadcastMsg)
	
	h.logger.Info("User left POI", "sessionId", client.SessionID, "userId", client.UserID, "poiId", poiID)
}