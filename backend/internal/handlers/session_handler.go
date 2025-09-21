package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SessionServiceInterface defines the interface for session service operations
type SessionServiceInterface interface {
	CreateSession(ctx context.Context, userID, mapID string, avatarPosition models.LatLng) (*models.Session, error)
	GetSession(ctx context.Context, sessionID string) (*models.Session, error)
	UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error
	EndSession(ctx context.Context, sessionID string) error
	SessionHeartbeat(ctx context.Context, sessionID string) error
	GetActiveSessionsForMap(ctx context.Context, mapID string) ([]*models.Session, error)
	CleanupExpiredSessions(ctx context.Context) error
}

// SessionHandler handles HTTP requests for session operations
type SessionHandler struct {
	sessionService SessionServiceInterface
	rateLimiter    services.RateLimiterInterface
}

// NewSessionHandler creates a new SessionHandler instance
func NewSessionHandler(sessionService SessionServiceInterface, rateLimiter services.RateLimiterInterface) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
		rateLimiter:    rateLimiter,
	}
}

// RegisterRoutes registers session-related routes
func (h *SessionHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// Session management
		api.POST("/sessions", h.CreateSession)
		api.GET("/sessions/:sessionId", h.GetSession)
		api.PUT("/sessions/:sessionId/avatar", h.UpdateAvatarPosition)
		api.POST("/sessions/:sessionId/heartbeat", h.SessionHeartbeat)
		api.DELETE("/sessions/:sessionId", h.EndSession)
		
		// Map-related session queries
		api.GET("/maps/:mapId/sessions", h.GetActiveSessionsForMap)
	}
}

// Request/Response DTOs

// CreateSessionRequest represents the request body for creating a session
type CreateSessionRequest struct {
	UserID         string         `json:"userId" binding:"required"`
	MapID          string         `json:"mapId" binding:"required"`
	AvatarPosition models.LatLng  `json:"avatarPosition" binding:"required"`
}

// CreateSessionResponse represents the response for creating a session
type CreateSessionResponse struct {
	SessionID      string         `json:"sessionId"`
	UserID         string         `json:"userId"`
	MapID          string         `json:"mapId"`
	AvatarPosition models.LatLng  `json:"avatarPosition"`
	CreatedAt      time.Time      `json:"createdAt"`
	IsActive       bool           `json:"isActive"`
}

// GetSessionResponse represents the response for getting a session
type GetSessionResponse struct {
	SessionID      string         `json:"sessionId"`
	UserID         string         `json:"userId"`
	MapID          string         `json:"mapId"`
	AvatarPosition models.LatLng  `json:"avatarPosition"`
	CreatedAt      time.Time      `json:"createdAt"`
	LastActive     time.Time      `json:"lastActive"`
	IsActive       bool           `json:"isActive"`
}

// UpdateAvatarPositionRequest represents the request body for updating avatar position
type UpdateAvatarPositionRequest struct {
	Position models.LatLng `json:"position" binding:"required"`
}

// UpdateAvatarPositionResponse represents the response for updating avatar position
type UpdateAvatarPositionResponse struct {
	Success bool `json:"success"`
}

// SessionHeartbeatResponse represents the response for session heartbeat
type SessionHeartbeatResponse struct {
	Success bool `json:"success"`
}

// EndSessionResponse represents the response for ending a session
type EndSessionResponse struct {
	Success bool `json:"success"`
}

// SessionInfo represents session information in responses
type SessionInfo struct {
	SessionID      string         `json:"sessionId"`
	UserID         string         `json:"userId"`
	AvatarPosition models.LatLng  `json:"avatarPosition"`
	LastActive     time.Time      `json:"lastActive"`
	IsActive       bool           `json:"isActive"`
}

// GetActiveSessionsResponse represents the response for getting active sessions
type GetActiveSessionsResponse struct {
	MapID    string        `json:"mapId"`
	Sessions []SessionInfo `json:"sessions"`
	Count    int           `json:"count"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// CreateSession handles POST /api/sessions
func (h *SessionHandler) CreateSession(c *gin.Context) {
	var req CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Validate request
	if err := h.validateCreateSessionRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: err.Error(),
		})
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, req.UserID, services.ActionCreateSession); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Create session
	session, err := h.sessionService.CreateSession(c, req.UserID, req.MapID, req.AvatarPosition)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "MAP_NOT_FOUND",
				Message: "Map not found",
			})
			return
		}
		
		if isUserAlreadyInMapError(err) {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    "USER_ALREADY_IN_MAP",
				Message: "User already has an active session in this map",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create session",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, req.UserID, services.ActionCreateSession)
	
	// Return response
	response := CreateSessionResponse{
		SessionID:      session.ID,
		UserID:         session.UserID,
		MapID:          session.MapID,
		AvatarPosition: session.AvatarPos,
		CreatedAt:      session.CreatedAt,
		IsActive:       session.IsActive,
	}
	
	c.JSON(http.StatusCreated, response)
}

// GetSession handles GET /api/sessions/:sessionId
func (h *SessionHandler) GetSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Session ID is required",
		})
		return
	}
	
	// Get session
	session, err := h.sessionService.GetSession(c, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get session",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	response := GetSessionResponse{
		SessionID:      session.ID,
		UserID:         session.UserID,
		MapID:          session.MapID,
		AvatarPosition: session.AvatarPos,
		CreatedAt:      session.CreatedAt,
		LastActive:     session.LastActive,
		IsActive:       session.IsActive,
	}
	
	c.JSON(http.StatusOK, response)
}

// UpdateAvatarPosition handles PUT /api/sessions/:sessionId/avatar
func (h *SessionHandler) UpdateAvatarPosition(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Session ID is required",
		})
		return
	}
	
	var req UpdateAvatarPositionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Validate position
	if err := req.Position.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid position",
			Details: err.Error(),
		})
		return
	}
	
	// Get session to extract user ID for rate limiting
	session, err := h.sessionService.GetSession(c, sessionID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get session",
			Details: err.Error(),
		})
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, session.UserID, services.ActionUpdateAvatar); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Update avatar position
	if err := h.sessionService.UpdateAvatarPosition(c, sessionID, req.Position); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update avatar position",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, session.UserID, services.ActionUpdateAvatar)
	
	// Return response
	c.JSON(http.StatusOK, UpdateAvatarPositionResponse{
		Success: true,
	})
}

// SessionHeartbeat handles POST /api/sessions/:sessionId/heartbeat
func (h *SessionHandler) SessionHeartbeat(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Session ID is required",
		})
		return
	}
	
	// Update session heartbeat
	if err := h.sessionService.SessionHeartbeat(c, sessionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update session heartbeat",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	c.JSON(http.StatusOK, SessionHeartbeatResponse{
		Success: true,
	})
}

// EndSession handles DELETE /api/sessions/:sessionId
func (h *SessionHandler) EndSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Session ID is required",
		})
		return
	}
	
	// End session
	if err := h.sessionService.EndSession(c, sessionID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Code:    "SESSION_NOT_FOUND",
				Message: "Session not found",
			})
			return
		}
		
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to end session",
			Details: err.Error(),
		})
		return
	}
	
	// Return response
	c.JSON(http.StatusOK, EndSessionResponse{
		Success: true,
	})
}

// GetActiveSessionsForMap handles GET /api/maps/:mapId/sessions
func (h *SessionHandler) GetActiveSessionsForMap(c *gin.Context) {
	mapID := c.Param("mapId")
	if mapID == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Map ID is required",
		})
		return
	}
	
	// Get active sessions
	sessions, err := h.sessionService.GetActiveSessionsForMap(c, mapID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get active sessions",
			Details: err.Error(),
		})
		return
	}
	
	// Convert to response format
	sessionInfos := make([]SessionInfo, len(sessions))
	for i, session := range sessions {
		sessionInfos[i] = SessionInfo{
			SessionID:      session.ID,
			UserID:         session.UserID,
			AvatarPosition: session.AvatarPos,
			LastActive:     session.LastActive,
			IsActive:       session.IsActive,
		}
	}
	
	// Return response
	response := GetActiveSessionsResponse{
		MapID:    mapID,
		Sessions: sessionInfos,
		Count:    len(sessionInfos),
	}
	
	c.JSON(http.StatusOK, response)
}

// Helper methods

// validateCreateSessionRequest validates the create session request
func (h *SessionHandler) validateCreateSessionRequest(req CreateSessionRequest) error {
	if req.UserID == "" {
		return errors.New("user ID is required")
	}
	if req.MapID == "" {
		return errors.New("map ID is required")
	}
	if err := req.AvatarPosition.Validate(); err != nil {
		return errors.New("invalid avatar position: " + err.Error())
	}
	return nil
}

// handleRateLimitError handles rate limit errors
func (h *SessionHandler) handleRateLimitError(c *gin.Context, err error) {
	if rateLimitErr, ok := err.(*services.RateLimitError); ok {
		// Set Retry-After header
		c.Header("Retry-After", strconv.Itoa(int(rateLimitErr.RetryAfter.Seconds())))
		
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: "Rate limit exceeded",
			Details: err.Error(),
		})
		return
	}
	
	// Generic rate limit error
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Rate limit exceeded",
		Details: err.Error(),
	})
}

// addRateLimitHeaders adds rate limit headers to the response
func (h *SessionHandler) addRateLimitHeaders(c *gin.Context, userID string, action services.ActionType) {
	headers, err := h.rateLimiter.GetRateLimitHeaders(c, userID, action)
	if err != nil {
		// Log error but don't fail the request
		return
	}
	
	for key, value := range headers {
		c.Header(key, value)
	}
}

// isUserAlreadyInMapError checks if the error indicates user already in map
func isUserAlreadyInMapError(err error) bool {
	// This would depend on how the session service reports this error
	// For now, we'll check the error message
	return err != nil && (err.Error() == "user already has an active session in this map" ||
		err.Error() == "duplicate session")
}