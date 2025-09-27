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

// UserServiceInterface defines the interface for user operations
type UserServiceInterface interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
}

// PubSubInterface defines the interface for PubSub operations
type PubSubInterface interface {
	SubscribePOIEvents(ctx context.Context, callback func(eventType string, data interface{})) error
}

// Handler handles WebSocket connections and messages
type Handler struct {
	sessionService SessionServiceInterface
	rateLimiter    RateLimiterInterface
	userService    UserServiceInterface
	pubsub         PubSubInterface
	manager        *Manager
	upgrader       ws.Upgrader
	logger         *slog.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(sessionService SessionServiceInterface, rateLimiter RateLimiterInterface, userService UserServiceInterface) *Handler {
	return &Handler{
		sessionService: sessionService,
		rateLimiter:    rateLimiter,
		userService:    userService,
		pubsub:         nil, // Will be set via SetPubSub if needed
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

// SetPubSub sets the PubSub interface for real-time event broadcasting
func (h *Handler) SetPubSub(pubsub PubSubInterface) {
	h.pubsub = pubsub
	
	// Start listening for PubSub events
	if pubsub != nil {
		go h.listenForPubSubEvents()
	}
}

// HandleWebSocket handles WebSocket connection upgrades
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// Extract session ID from query parameter (preferred for WebSocket) or Authorization header
	sessionID := c.Query("sessionId")
	if sessionID == "" {
		// Fallback to Authorization header
		authHeader := c.GetHeader("Authorization")
		var err error
		sessionID, err = extractSessionID(authHeader)
		if err != nil {
			h.logger.Warn("WebSocket connection failed: missing sessionId")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing sessionId query parameter or Authorization header"})
			return
		}
	}
	
	h.logger.Info("WebSocket connection attempt", "sessionId", sessionID)
	
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
	
	// Automatically send initial users to the new client
	h.logger.Info("📋 Automatically sending initial users to new client", "sessionId", sessionID)
	h.handleRequestInitialUsers(c.Request.Context(), client, Message{Type: "request_initial_users"})
	
	// Try to get user profile for display name, avatar, and about me
	displayName := session.UserID
	var avatarURL *string
	var aboutMe *string
	
	if h.userService != nil {
		user, err := h.userService.GetUser(c.Request.Context(), session.UserID)
		if err == nil && user != nil {
			displayName = user.DisplayName
			avatarURL = user.AvatarURL
			aboutMe = user.AboutMe
		} else {
			h.logger.Debug("Could not get user profile for user_joined", 
				"userId", session.UserID, 
				"error", err)
			// Fallback to first 8 characters of UUID
			if len(session.UserID) > 8 {
				displayName = session.UserID[:8]
			}
		}
	} else {
		// Fallback to first 8 characters of UUID
		if len(session.UserID) > 8 {
			displayName = session.UserID[:8]
		}
	}
	
	// Convert relative avatar URL to absolute URL
	// TODO: Make base URL configurable instead of hardcoded localhost:8080
	var fullAvatarURL *string
	if avatarURL != nil && *avatarURL != "" {
		// Convert relative path to full URL
		fullURL := "http://localhost:8080" + *avatarURL
		fullAvatarURL = &fullURL
	}
	
	userJoinedMsg := Message{
		Type: "user_joined",
		Data: map[string]interface{}{
			"sessionId":   sessionID,
			"userId":      session.UserID,
			"displayName": displayName,
			"avatarURL":   fullAvatarURL,
			"aboutMe":     aboutMe,
			"position": map[string]float64{
				"lat": session.AvatarPos.Lat,
				"lng": session.AvatarPos.Lng,
			},
			"role": "user", // Default role
		},
		Timestamp: time.Now(),
	}
	
	mapClientCount := h.manager.GetMapClients(session.MapID)
	h.logger.Info("📡 Broadcasting user joined", 
		"sessionId", sessionID, 
		"userId", session.UserID, 
		"mapId", session.MapID,
		"mapClientCount", mapClientCount,
		"broadcastType", "user_joined")
	
	h.manager.BroadcastToMapExcept(session.MapID, sessionID, userJoinedMsg)
	
	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump(h)
}

// readPump handles reading messages from the WebSocket connection
func (c *Client) readPump(handler *Handler) {
	defer func() {
		// Broadcast user left to other clients in the same map
		userLeftMsg := Message{
			Type: "user_left",
			Data: map[string]interface{}{
				"sessionId": c.SessionID,
				"userId":    c.UserID,
			},
			Timestamp: time.Now(),
		}
		c.Manager.BroadcastToMapExcept(c.MapID, c.SessionID, userLeftMsg)
		
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
	case "request_initial_users":
		h.logger.Info("📋 Request initial users received", "sessionId", client.SessionID)
		h.handleRequestInitialUsers(ctx, client, msg)
	case "poi_join":
		h.handlePOIJoin(ctx, client, msg)
	case "poi_leave":
		h.handlePOILeave(ctx, client, msg)
	case "call_request":
		h.handleCallRequest(ctx, client, msg)
	case "call_accept":
		h.handleCallAccept(ctx, client, msg)
	case "call_reject":
		h.handleCallReject(ctx, client, msg)
	case "call_end":
		h.handleCallEnd(ctx, client, msg)
	case "webrtc_offer":
		h.handleWebRTCOffer(ctx, client, msg)
	case "webrtc_answer":
		h.handleWebRTCAnswer(ctx, client, msg)
	case "ice_candidate":
		h.handleICECandidate(ctx, client, msg)
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
	h.logger.Info("🏃 Avatar move request received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID, 
		"mapId", client.MapID)
	
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
	
	// Get current map clients for logging
	mapClientCount := h.manager.GetMapClients(client.MapID)
	h.logger.Info("📡 Broadcasting avatar movement", 
		"sessionId", client.SessionID, 
		"userId", client.UserID, 
		"mapId", client.MapID,
		"position", position,
		"mapClientCount", mapClientCount,
		"broadcastType", "avatar_moved")
	
	// Broadcast to all clients in the same map except the sender
	h.manager.BroadcastToMapExcept(client.MapID, client.SessionID, broadcastMsg)
	
	h.logger.Info("✅ Avatar position updated and broadcasted", 
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
		
	case "request_initial_users":
		// Request initial users message - no additional validation needed
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
		
	case "call_request":
		// Validate call request message
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		if _, ok := data["targetUserId"].(string); !ok {
			return errors.New("targetUserId is required for call_request")
		}
		
		if _, ok := data["callId"].(string); !ok {
			return errors.New("callId is required for call_request")
		}
		
		return nil
		
	case "call_accept", "call_reject":
		// Validate call accept/reject messages
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		if _, ok := data["callId"].(string); !ok {
			return errors.New("callId is required")
		}
		
		if _, ok := data["callerUserId"].(string); !ok {
			return errors.New("callerUserId is required")
		}
		
		return nil
		
	case "call_end":
		// Validate call end message
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		if _, ok := data["callId"].(string); !ok {
			return errors.New("callId is required for call_end")
		}
		
		if _, ok := data["otherUserId"].(string); !ok {
			return errors.New("otherUserId is required for call_end")
		}
		
		return nil
		
	case "webrtc_offer", "webrtc_answer":
		// Validate WebRTC offer/answer messages
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		if _, ok := data["callId"].(string); !ok {
			return errors.New("callId is required for WebRTC offer/answer")
		}
		
		if _, ok := data["targetUserId"].(string); !ok {
			return errors.New("targetUserId is required for WebRTC offer/answer")
		}
		
		if _, ok := data["sdp"]; !ok {
			return errors.New("sdp is required for WebRTC offer/answer")
		}
		
		return nil
		
	case "ice_candidate":
		// Validate ICE candidate messages
		data, ok := msg.Data.(map[string]interface{})
		if !ok {
			return errors.New("invalid data format")
		}
		
		if _, ok := data["callId"].(string); !ok {
			return errors.New("callId is required for ICE candidate")
		}
		
		if _, ok := data["targetUserId"].(string); !ok {
			return errors.New("targetUserId is required for ICE candidate")
		}
		
		if _, ok := data["candidate"]; !ok {
			return errors.New("candidate is required for ICE candidate")
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


// handleRequestInitialUsers sends the list of currently connected users to a new client
func (h *Handler) handleRequestInitialUsers(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("📋 Processing initial users request", 
		"sessionId", client.SessionID, 
		"mapId", client.MapID)
	
	// Get all sessions for the current map
	sessions := h.manager.GetMapClientSessions(client.MapID)
	
	var users []map[string]interface{}
	
	// For each session, get the user information
	for _, sessionID := range sessions {
		if sessionID == client.SessionID {
			continue // Skip the requesting client
		}
		
		// Get session info from the session service
		session, err := h.sessionService.GetSession(ctx, sessionID)
		if err != nil {
			h.logger.Warn("Failed to get session for initial users", 
				"sessionId", sessionID, 
				"error", err.Error())
			continue
		}
		
		if !session.IsActive {
			continue
		}
		
		// Try to get user profile for display name, avatar, and about me
		displayName := session.UserID
		var avatarURL *string
		var aboutMe *string
		
		if h.userService != nil {
			user, err := h.userService.GetUser(ctx, session.UserID)
			if err == nil && user != nil {
				displayName = user.DisplayName
				avatarURL = user.AvatarURL
				aboutMe = user.AboutMe
				h.logger.Info("📸 User profile found for initial users", 
					"userId", session.UserID, 
					"displayName", displayName,
					"hasAvatar", avatarURL != nil,
					"avatarURL", func() string {
						if avatarURL != nil {
							return *avatarURL
						}
						return "nil"
					}())
			} else {
				h.logger.Debug("Could not get user profile for display name", 
					"userId", session.UserID, 
					"error", err)
				// Fallback to first 8 characters of UUID
				if len(session.UserID) > 8 {
					displayName = session.UserID[:8]
				}
			}
		} else {
			// Fallback to first 8 characters of UUID
			if len(session.UserID) > 8 {
				displayName = session.UserID[:8]
			}
		}
		
		// Convert relative avatar URL to absolute URL
		// TODO: Make base URL configurable instead of hardcoded localhost:8080
		var fullAvatarURL *string
		if avatarURL != nil && *avatarURL != "" {
			// Convert relative path to full URL
			fullURL := "http://localhost:8080" + *avatarURL
			fullAvatarURL = &fullURL
		}
		
		userData := map[string]interface{}{
			"sessionId":   sessionID,
			"userId":      session.UserID,
			"displayName": displayName,
			"avatarURL":   fullAvatarURL,
			"aboutMe":     aboutMe,
			"position": map[string]float64{
				"lat": session.AvatarPos.Lat,
				"lng": session.AvatarPos.Lng,
			},
			"role": "user", // Default role
		}
		
		users = append(users, userData)
	}
	
	// Send initial users message
	initialUsersMsg := Message{
		Type: "initial_users",
		Data: map[string]interface{}{
			"users": users,
		},
		Timestamp: time.Now(),
	}
	
	select {
	case client.Send <- initialUsersMsg:
		h.logger.Info("Sent initial users to client", 
			"sessionId", client.SessionID, 
			"userCount", len(users))
	default:
		h.logger.Warn("Failed to send initial users to client", 
			"sessionId", client.SessionID)
	}}

// Video Call Handlers

// handleCallRequest processes incoming call requests
func (h *Handler) handleCallRequest(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("📞 Call request received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid call request data format")
		return
	}
	
	targetUserId, ok := data["targetUserId"].(string)
	if !ok || targetUserId == "" {
		h.sendErrorMessage(client, "Target user ID is required for call request")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for call request")
		return
	}
	
	// Get caller info
	callerInfo := map[string]interface{}{
		"userId":      client.UserID,
		"sessionId":   client.SessionID,
		"displayName": data["callerName"], // Optional caller name
	}
	
	// Create call request message for target user
	callRequestMsg := Message{
		Type: "call_request",
		Data: map[string]interface{}{
			"callId":     callId,
			"callerInfo": callerInfo,
		},
		Timestamp: time.Now(),
	}
	
	// Find target user and send call request
	h.manager.BroadcastToUser(targetUserId, callRequestMsg, client.SessionID)
	
	h.logger.Info("📞 Call request sent to target user", 
		"callId", callId,
		"caller", client.UserID,
		"target", targetUserId)
}

// handleCallAccept processes call accept messages
func (h *Handler) handleCallAccept(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("✅ Call accept received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid call accept data format")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for call accept")
		return
	}
	
	callerUserId, ok := data["callerUserId"].(string)
	if !ok || callerUserId == "" {
		h.sendErrorMessage(client, "Caller user ID is required for call accept")
		return
	}
	
	// Create call accept message for caller
	callAcceptMsg := Message{
		Type: "call_accept",
		Data: map[string]interface{}{
			"callId":   callId,
			"accepter": client.UserID,
		},
		Timestamp: time.Now(),
	}
	
	// Send accept message to caller
	h.manager.BroadcastToUser(callerUserId, callAcceptMsg, client.SessionID)
	
	// Broadcast call status update to all users on the map (both users are now in call)
	callStatusMsg := Message{
		Type: "user_call_status",
		Data: map[string]interface{}{
			"userId":   client.UserID,
			"isInCall": true,
		},
		Timestamp: time.Now(),
	}
	h.manager.BroadcastToMap(client.MapID, callStatusMsg)
	h.logger.Info("📡 Broadcasting call status", "userId", client.UserID, "isInCall", true, "mapId", client.MapID)
	
	callerStatusMsg := Message{
		Type: "user_call_status",
		Data: map[string]interface{}{
			"userId":   callerUserId,
			"isInCall": true,
		},
		Timestamp: time.Now(),
	}
	h.manager.BroadcastToMap(client.MapID, callerStatusMsg)
	h.logger.Info("📡 Broadcasting call status", "userId", callerUserId, "isInCall", true, "mapId", client.MapID)
	
	h.logger.Info("✅ Call accept sent to caller", 
		"callId", callId,
		"accepter", client.UserID,
		"caller", callerUserId)
}

// handleCallReject processes call reject messages
func (h *Handler) handleCallReject(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("❌ Call reject received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid call reject data format")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for call reject")
		return
	}
	
	callerUserId, ok := data["callerUserId"].(string)
	if !ok || callerUserId == "" {
		h.sendErrorMessage(client, "Caller user ID is required for call reject")
		return
	}
	
	// Create call reject message for caller
	callRejectMsg := Message{
		Type: "call_reject",
		Data: map[string]interface{}{
			"callId":   callId,
			"rejecter": client.UserID,
		},
		Timestamp: time.Now(),
	}
	
	// Send reject message to caller
	h.manager.BroadcastToUser(callerUserId, callRejectMsg, client.SessionID)
	
	// Broadcast call status update to all users on the map (both users are no longer in call)
	callStatusMsg := Message{
		Type: "user_call_status",
		Data: map[string]interface{}{
			"userId":   client.UserID,
			"isInCall": false,
		},
		Timestamp: time.Now(),
	}
	h.manager.BroadcastToMap(client.MapID, callStatusMsg)
	h.logger.Info("📡 Broadcasting call status", "userId", client.UserID, "isInCall", false, "mapId", client.MapID)
	
	callerStatusMsg := Message{
		Type: "user_call_status",
		Data: map[string]interface{}{
			"userId":   callerUserId,
			"isInCall": false,
		},
		Timestamp: time.Now(),
	}
	h.manager.BroadcastToMap(client.MapID, callerStatusMsg)
	h.logger.Info("📡 Broadcasting call status", "userId", callerUserId, "isInCall", false, "mapId", client.MapID)
	
	h.logger.Info("❌ Call reject sent to caller", 
		"callId", callId,
		"rejecter", client.UserID,
		"caller", callerUserId)
}

// handleCallEnd processes call end messages
func (h *Handler) handleCallEnd(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("📵 Call end received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid call end data format")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for call end")
		return
	}
	
	otherUserId, ok := data["otherUserId"].(string)
	if !ok || otherUserId == "" {
		h.sendErrorMessage(client, "Other user ID is required for call end")
		return
	}
	
	// Create call end message for other user
	callEndMsg := Message{
		Type: "call_end",
		Data: map[string]interface{}{
			"callId": callId,
			"ender":  client.UserID,
		},
		Timestamp: time.Now(),
	}
	
	// Send end message to other user
	h.manager.BroadcastToUser(otherUserId, callEndMsg, client.SessionID)
	
	// Broadcast call status update to all users on the map (both users are no longer in call)
	callStatusMsg := Message{
		Type: "user_call_status",
		Data: map[string]interface{}{
			"userId":   client.UserID,
			"isInCall": false,
		},
		Timestamp: time.Now(),
	}
	h.manager.BroadcastToMap(client.MapID, callStatusMsg)
	h.logger.Info("📡 Broadcasting call status", "userId", client.UserID, "isInCall", false, "mapId", client.MapID)
	
	otherStatusMsg := Message{
		Type: "user_call_status",
		Data: map[string]interface{}{
			"userId":   otherUserId,
			"isInCall": false,
		},
		Timestamp: time.Now(),
	}
	h.manager.BroadcastToMap(client.MapID, otherStatusMsg)
	h.logger.Info("📡 Broadcasting call status", "userId", otherUserId, "isInCall", false, "mapId", client.MapID)
	
	h.logger.Info("📵 Call end sent to other user", 
		"callId", callId,
		"ender", client.UserID,
		"other", otherUserId)
}

// sendErrorMessage sends an error message to a client
func (h *Handler) sendErrorMessage(client *Client, message string) {
	errorMsg := Message{
		Type: "error",
		Data: map[string]interface{}{
			"message": message,
		},
		Timestamp: time.Now(),
	}
	
	select {
	case client.Send <- errorMsg:
		h.logger.Warn("Error message sent to client", 
			"sessionId", client.SessionID, 
			"message", message)
	default:
		h.logger.Error("Failed to send error message to client", 
			"sessionId", client.SessionID, 
			"message", message)
	}
}

// WebRTC Signaling Handlers

// handleWebRTCOffer processes WebRTC offer messages
func (h *Handler) handleWebRTCOffer(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("📝 WebRTC offer received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid WebRTC offer data format")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for WebRTC offer")
		return
	}
	
	targetUserId, ok := data["targetUserId"].(string)
	if !ok || targetUserId == "" {
		h.sendErrorMessage(client, "Target user ID is required for WebRTC offer")
		return
	}
	
	sdp := data["sdp"]
	if sdp == nil {
		h.sendErrorMessage(client, "SDP is required for WebRTC offer")
		return
	}
	
	// Create WebRTC offer message for target user
	offerMsg := Message{
		Type: "webrtc_offer",
		Data: map[string]interface{}{
			"callId":     callId,
			"fromUserId": client.UserID,
			"sdp":        sdp,
		},
		Timestamp: time.Now(),
	}
	
	// Send offer to target user
	h.manager.BroadcastToUser(targetUserId, offerMsg, client.SessionID)
	
	h.logger.Info("📝 WebRTC offer sent to target user", 
		"callId", callId,
		"from", client.UserID,
		"to", targetUserId)
}

// handleWebRTCAnswer processes WebRTC answer messages
func (h *Handler) handleWebRTCAnswer(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("📋 WebRTC answer received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid WebRTC answer data format")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for WebRTC answer")
		return
	}
	
	targetUserId, ok := data["targetUserId"].(string)
	if !ok || targetUserId == "" {
		h.sendErrorMessage(client, "Target user ID is required for WebRTC answer")
		return
	}
	
	sdp := data["sdp"]
	if sdp == nil {
		h.sendErrorMessage(client, "SDP is required for WebRTC answer")
		return
	}
	
	// Create WebRTC answer message for target user
	answerMsg := Message{
		Type: "webrtc_answer",
		Data: map[string]interface{}{
			"callId":     callId,
			"fromUserId": client.UserID,
			"sdp":        sdp,
		},
		Timestamp: time.Now(),
	}
	
	// Send answer to target user
	h.manager.BroadcastToUser(targetUserId, answerMsg, client.SessionID)
	
	h.logger.Info("📋 WebRTC answer sent to target user", 
		"callId", callId,
		"from", client.UserID,
		"to", targetUserId)
}

// handleICECandidate processes ICE candidate messages
func (h *Handler) handleICECandidate(ctx context.Context, client *Client, msg Message) {
	h.logger.Info("🧊 ICE candidate received", 
		"sessionId", client.SessionID, 
		"userId", client.UserID)
	
	data, ok := msg.Data.(map[string]interface{})
	if !ok {
		h.sendErrorMessage(client, "Invalid ICE candidate data format")
		return
	}
	
	callId, ok := data["callId"].(string)
	if !ok || callId == "" {
		h.sendErrorMessage(client, "Call ID is required for ICE candidate")
		return
	}
	
	targetUserId, ok := data["targetUserId"].(string)
	if !ok || targetUserId == "" {
		h.sendErrorMessage(client, "Target user ID is required for ICE candidate")
		return
	}
	
	candidate := data["candidate"]
	if candidate == nil {
		h.sendErrorMessage(client, "Candidate is required for ICE candidate")
		return
	}
	
	// Create ICE candidate message for target user
	candidateMsg := Message{
		Type: "ice_candidate",
		Data: map[string]interface{}{
			"callId":     callId,
			"fromUserId": client.UserID,
			"candidate":  candidate,
		},
		Timestamp: time.Now(),
	}
	
	// Send candidate to target user
	h.manager.BroadcastToUser(targetUserId, candidateMsg, client.SessionID)
	
	h.logger.Info("🧊 ICE candidate sent to target user", 
		"callId", callId,
		"from", client.UserID,
		"to", targetUserId)
}

// listenForPubSubEvents listens for Redis PubSub events and broadcasts them to WebSocket clients
func (h *Handler) listenForPubSubEvents() {
	if h.pubsub == nil {
		return
	}
	
	h.logger.Info("🔊 Starting PubSub event listener for WebSocket broadcasting")
	
	ctx := context.Background()
	err := h.pubsub.SubscribePOIEvents(ctx, func(eventType string, data interface{}) {
		h.handlePubSubEvent(eventType, data)
	})
	
	if err != nil {
		h.logger.Error("❌ Failed to subscribe to PubSub events", "error", err)
	}
}

// handlePubSubEvent processes PubSub events and broadcasts them to appropriate WebSocket clients
func (h *Handler) handlePubSubEvent(eventType string, data interface{}) {
	h.logger.Info("📢 Received PubSub event", "type", eventType, "data", data)
	
	switch eventType {
	case "poi_created":
		h.handlePOICreatedEvent(data)
	case "poi_joined":
		h.handlePOIJoinedEvent(data)
	case "poi_left":
		h.handlePOILeftEvent(data)
	case "poi_updated":
		h.handlePOIUpdatedEvent(data)
	default:
		h.logger.Warn("❓ Unknown PubSub event type", "type", eventType)
	}
}

// handlePOICreatedEvent broadcasts POI creation to all clients on the same map
func (h *Handler) handlePOICreatedEvent(data interface{}) {
	poiData, ok := data.(map[string]interface{})
	if !ok {
		h.logger.Error("❌ Invalid POI created event data", "data", data)
		return
	}
	
	mapID, ok := poiData["mapId"].(string)
	if !ok {
		h.logger.Error("❌ Missing mapId in POI created event", "data", data)
		return
	}
	
	// Create WebSocket message
	message := Message{
		Type:      "poi_created",
		Data:      poiData,
		Timestamp: time.Now(),
	}
	
	// Broadcast to all clients on the same map
	h.manager.BroadcastToMap(mapID, message)
	
	h.logger.Info("📢 Broadcasted POI created event", "mapId", mapID, "poiId", poiData["poiId"])
}

// handlePOIJoinedEvent broadcasts POI join to all clients on the same map
func (h *Handler) handlePOIJoinedEvent(data interface{}) {
	poiData, ok := data.(map[string]interface{})
	if !ok {
		h.logger.Error("❌ Invalid POI joined event data", "data", data)
		return
	}
	
	mapID, ok := poiData["mapId"].(string)
	if !ok {
		h.logger.Error("❌ Missing mapId in POI joined event", "data", data)
		return
	}
	
	// Create WebSocket message
	message := Message{
		Type:      "poi_joined",
		Data:      poiData,
		Timestamp: time.Now(),
	}
	
	// Broadcast to all clients on the same map
	h.manager.BroadcastToMap(mapID, message)
	
	h.logger.Info("📢 Broadcasted POI joined event", "mapId", mapID, "poiId", poiData["poiId"], "userId", poiData["userId"])
}

// handlePOILeftEvent broadcasts POI leave to all clients on the same map
func (h *Handler) handlePOILeftEvent(data interface{}) {
	poiData, ok := data.(map[string]interface{})
	if !ok {
		h.logger.Error("❌ Invalid POI left event data", "data", data)
		return
	}
	
	mapID, ok := poiData["mapId"].(string)
	if !ok {
		h.logger.Error("❌ Missing mapId in POI left event", "data", data)
		return
	}
	
	// Create WebSocket message
	message := Message{
		Type:      "poi_left",
		Data:      poiData,
		Timestamp: time.Now(),
	}
	
	// Broadcast to all clients on the same map
	h.manager.BroadcastToMap(mapID, message)
	
	h.logger.Info("📢 Broadcasted POI left event", "mapId", mapID, "poiId", poiData["poiId"], "userId", poiData["userId"])
}

// handlePOIUpdatedEvent broadcasts POI updates to all clients on the same map
func (h *Handler) handlePOIUpdatedEvent(data interface{}) {
	poiData, ok := data.(map[string]interface{})
	if !ok {
		h.logger.Error("❌ Invalid POI updated event data", "data", data)
		return
	}
	
	mapID, ok := poiData["mapId"].(string)
	if !ok {
		h.logger.Error("❌ Missing mapId in POI updated event", "data", data)
		return
	}
	
	// Create WebSocket message
	message := Message{
		Type:      "poi_updated",
		Data:      poiData,
		Timestamp: time.Now(),
	}
	
	// Broadcast to all clients on the same map
	h.manager.BroadcastToMap(mapID, message)
	
	h.logger.Info("📢 Broadcasted POI updated event", "mapId", mapID, "poiId", poiData["poiId"])
}