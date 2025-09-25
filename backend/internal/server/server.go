package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"breakoutglobe/internal/config"
	"breakoutglobe/internal/handlers"
	"breakoutglobe/internal/models"
	"breakoutglobe/internal/repository"
	"breakoutglobe/internal/services"
)

type Server struct {
	config *config.Config
	router *gin.Engine
	db     *gorm.DB
	// Simple in-memory storage for POI participants (for testing)
	poiParticipants map[string]map[string]string // poiId -> sessionId -> username
}

func New(cfg *config.Config) *Server {
	gin.SetMode(cfg.GinMode)
	
	// Setup database connection
	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	
	// Auto-migrate models
	if err := db.AutoMigrate(&models.User{}); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}
	
	router := gin.Default()
	
	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))
	
	s := &Server{
		config: cfg,
		router: router,
		db:     db,
		poiParticipants: make(map[string]map[string]string),
	}
	
	s.setupRoutes()
	
	return s
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"service": "breakoutglobe-api",
		})
	})
	
	// API status
	api := s.router.Group("/api")
	{
		api.GET("/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "BreakoutGlobe API is running",
			})
		})
		
		// Simple session endpoints for testing
		api.POST("/sessions", s.createSession)
		api.GET("/sessions/:sessionId", s.getSession)
		api.PUT("/sessions/:sessionId/avatar", s.updateAvatarPosition)
		
		// Simple POI endpoints for testing
		api.GET("/pois", s.getPOIs)
		api.POST("/pois", s.createPOI)
		api.POST("/pois/:poiId/join", s.joinPOI)
		api.POST("/pois/:poiId/leave", s.leavePOI)
		
		// User profile endpoints with proper handlers
		s.setupUserRoutes(api)
	}
	
	// WebSocket endpoint (simple echo for now)
	s.router.GET("/ws", s.handleWebSocket)
}

func (s *Server) setupUserRoutes(api *gin.RouterGroup) {
	// Setup dependencies
	userRepo := repository.NewUserRepository(s.db)
	userService := services.NewUserService(userRepo)
	
	// For now, create a simple in-memory rate limiter (TODO: use Redis in production)
	rateLimiter := &SimpleRateLimiter{}
	
	userHandler := handlers.NewUserHandler(userService, rateLimiter)
	
	// Register user routes
	userHandler.RegisterRoutes(s.router)
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

// SimpleRateLimiter is a simple in-memory rate limiter for testing
type SimpleRateLimiter struct {
	mu sync.Mutex
	requests map[string][]time.Time
}

func (r *SimpleRateLimiter) IsAllowed(ctx context.Context, userID string, action services.ActionType) (bool, error) {
	allowed, _ := r.checkLimit(userID, action, false)
	return allowed, nil
}

func (r *SimpleRateLimiter) CheckRateLimit(ctx context.Context, userID string, action services.ActionType) error {
	allowed, err := r.checkLimit(userID, action, true)
	if err != nil {
		return err
	}
	if !allowed {
		return &services.RateLimitError{
			UserID:     userID,
			Action:     action,
			Limit:      100,
			RetryAfter: 3600, // 1 hour in seconds
		}
	}
	return nil
}

func (r *SimpleRateLimiter) checkLimit(userID string, action services.ActionType, addRequest bool) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.requests == nil {
		r.requests = make(map[string][]time.Time)
	}
	
	key := fmt.Sprintf("%s:%s", userID, action)
	now := time.Now()
	window := 1 * time.Hour // Simple 1-hour window
	limit := 100 // Simple limit of 100 requests per hour
	
	// Clean old requests
	var validRequests []time.Time
	for _, reqTime := range r.requests[key] {
		if now.Sub(reqTime) < window {
			validRequests = append(validRequests, reqTime)
		}
	}
	
	// Check if limit exceeded
	if len(validRequests) >= limit {
		return false, nil
	}
	
	// Add current request if requested
	if addRequest {
		validRequests = append(validRequests, now)
		r.requests[key] = validRequests
	}
	
	return true, nil
}

func (r *SimpleRateLimiter) GetRemainingRequests(ctx context.Context, userID string, action services.ActionType) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.requests == nil {
		return 100, nil
	}
	
	key := fmt.Sprintf("%s:%s", userID, action)
	now := time.Now()
	window := 1 * time.Hour
	limit := 100
	
	// Count valid requests
	count := 0
	for _, reqTime := range r.requests[key] {
		if now.Sub(reqTime) < window {
			count++
		}
	}
	
	remaining := limit - count
	if remaining < 0 {
		remaining = 0
	}
	return remaining, nil
}

func (r *SimpleRateLimiter) GetWindowResetTime(ctx context.Context, userID string, action services.ActionType) (time.Time, error) {
	return time.Now().Add(1 * time.Hour), nil
}

func (r *SimpleRateLimiter) SetCustomLimit(userID string, action services.ActionType, limit services.RateLimit) {
	// Simple implementation - ignore custom limits for now
}

func (r *SimpleRateLimiter) ClearUserLimits(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if r.requests == nil {
		return nil
	}
	
	// Remove all entries for this user
	for key := range r.requests {
		if len(key) > len(userID) && key[:len(userID)] == userID {
			delete(r.requests, key)
		}
	}
	
	return nil
}

func (r *SimpleRateLimiter) GetUserStats(ctx context.Context, userID string) (*services.UserRateLimitStats, error) {
	return &services.UserRateLimitStats{
		UserID:      userID,
		ActionStats: make(map[services.ActionType]services.ActionStats),
		GeneratedAt: time.Now(),
	}, nil
}

func (r *SimpleRateLimiter) GetRateLimitHeaders(ctx context.Context, userID string, action services.ActionType) (map[string]string, error) {
	remaining, _ := r.GetRemainingRequests(ctx, userID, action)
	return map[string]string{
		"X-RateLimit-Limit":     "100",
		"X-RateLimit-Remaining": fmt.Sprintf("%d", remaining),
	}, nil
}

// Simple handlers for testing integration

func (s *Server) createSession(c *gin.Context) {
	var req struct {
		UserID         string `json:"userId"`
		MapID          string `json:"mapId"`
		AvatarPosition struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"avatarPosition"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	sessionID := "session-" + req.UserID + "-" + req.MapID
	
	c.JSON(http.StatusCreated, gin.H{
		"sessionId":      sessionID,
		"userId":         req.UserID,
		"mapId":          req.MapID,
		"avatarPosition": req.AvatarPosition,
		"isActive":       true,
	})
}

func (s *Server) getSession(c *gin.Context) {
	sessionID := c.Param("sessionId")
	
	c.JSON(http.StatusOK, gin.H{
		"sessionId":      sessionID,
		"userId":         "test-user",
		"mapId":          "default-map",
		"avatarPosition": gin.H{"lat": 40.7128, "lng": -74.0060},
		"isActive":       true,
	})
}

func (s *Server) updateAvatarPosition(c *gin.Context) {
	sessionID := c.Param("sessionId")
	
	var req struct {
		Position struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"position"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"sessionId": sessionID,
		"position":  req.Position,
	})
}

func (s *Server) getPOIs(c *gin.Context) {
	mapID := c.Query("mapId")
	
	// Helper function to get participant info for a POI
	getParticipantInfo := func(poiID string) ([]gin.H, int) {
		participants := []gin.H{}
		if poiParticipants, exists := s.poiParticipants[poiID]; exists {
			for sessionID, username := range poiParticipants {
				participants = append(participants, gin.H{
					"id":   sessionID,
					"name": username,
				})
			}
		}
		return participants, len(participants)
	}
	
	// Get participant info for each POI
	poi1Participants, poi1Count := getParticipantInfo("poi-1")
	poi2Participants, poi2Count := getParticipantInfo("poi-2")
	
	// Return some mock POIs with real participant information
	pois := []gin.H{
		{
			"id":              "poi-1",
			"mapId":           mapID,
			"name":            "Meeting Room A",
			"description":     "A comfortable meeting room",
			"position":        gin.H{"lat": 40.7130, "lng": -74.0062},
			"createdBy":       "user-1",
			"maxParticipants": 10,
			"participantCount": poi1Count,
			"participants":     poi1Participants,
		},
		{
			"id":              "poi-2",
			"mapId":           mapID,
			"name":            "Coffee Corner",
			"description":     "Grab a coffee and chat",
			"position":        gin.H{"lat": 40.7125, "lng": -74.0058},
			"createdBy":       "user-2",
			"maxParticipants": 5,
			"participantCount": poi2Count,
			"participants":     poi2Participants,
		},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"mapId": mapID,
		"pois":  pois,
		"count": len(pois),
	})
}

func (s *Server) createPOI(c *gin.Context) {
	var req struct {
		MapID           string `json:"mapId"`
		Name            string `json:"name"`
		Description     string `json:"description"`
		Position        struct {
			Lat float64 `json:"lat"`
			Lng float64 `json:"lng"`
		} `json:"position"`
		CreatedBy       string `json:"createdBy"`
		MaxParticipants int    `json:"maxParticipants"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	poiID := "poi-" + req.Name + "-" + req.CreatedBy
	
	c.JSON(http.StatusCreated, gin.H{
		"id":              poiID,
		"mapId":           req.MapID,
		"name":            req.Name,
		"description":     req.Description,
		"position":        req.Position,
		"createdBy":       req.CreatedBy,
		"maxParticipants": req.MaxParticipants,
		"participantCount": 0,
	})
}

func (s *Server) joinPOI(c *gin.Context) {
	poiID := c.Param("poiId")
	
	var req struct {
		SessionID string `json:"sessionId"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Initialize POI participants map if it doesn't exist
	if s.poiParticipants[poiID] == nil {
		s.poiParticipants[poiID] = make(map[string]string)
	}
	
	// Add participant with a generated username
	username := "User-" + req.SessionID
	s.poiParticipants[poiID][req.SessionID] = username
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"poiId":   poiID,
		"userId":  req.SessionID,
	})
}

func (s *Server) leavePOI(c *gin.Context) {
	poiID := c.Param("poiId")
	
	var req struct {
		SessionID string `json:"sessionId"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Remove participant if they exist
	if s.poiParticipants[poiID] != nil {
		delete(s.poiParticipants[poiID], req.SessionID)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"poiId":   poiID,
		"userId":  req.SessionID,
	})
}

func (s *Server) handleWebSocket(c *gin.Context) {
	// Simple WebSocket echo server for testing
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true // Allow all origins for testing
		},
	}
	
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade connection"})
		return
	}
	defer conn.Close()
	
	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			break
		}
		
		// Echo the message back
		if err := conn.WriteMessage(messageType, message); err != nil {
			break
		}
	}
}

// Simple user profile handlers for testing

func (s *Server) getUserProfile(c *gin.Context) {
	// For testing, return 404 to trigger profile creation
	c.JSON(http.StatusNotFound, gin.H{
		"error": "Profile not found",
	})
}

func (s *Server) createUserProfile(c *gin.Context) {
	var req struct {
		DisplayName string `json:"displayName"`
		AboutMe     string `json:"aboutMe"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Generate a simple user profile response
	profile := gin.H{
		"id":            "user-" + req.DisplayName,
		"displayName":   req.DisplayName,
		"aboutMe":       req.AboutMe,
		"accountType":   "guest",
		"role":          "user",
		"isActive":      true,
		"emailVerified": false,
		"createdAt":     "2024-01-01T00:00:00Z",
	}
	
	c.JSON(http.StatusCreated, profile)
}

func (s *Server) updateUserProfile(c *gin.Context) {
	var req struct {
		DisplayName string `json:"displayName"`
		AboutMe     string `json:"aboutMe"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Return updated profile
	profile := gin.H{
		"id":            "user-" + req.DisplayName,
		"displayName":   req.DisplayName,
		"aboutMe":       req.AboutMe,
		"accountType":   "guest",
		"role":          "user",
		"isActive":      true,
		"emailVerified": false,
		"createdAt":     "2024-01-01T00:00:00Z",
	}
	
	c.JSON(http.StatusOK, profile)
}

func (s *Server) uploadAvatar(c *gin.Context) {
	// For testing, just return success without actually handling the file
	profile := gin.H{
		"id":            "user-test",
		"displayName":   "Test User",
		"avatarURL":     "https://via.placeholder.com/128",
		"accountType":   "guest",
		"role":          "user",
		"isActive":      true,
		"emailVerified": false,
		"createdAt":     "2024-01-01T00:00:00Z",
	}
	
	c.JSON(http.StatusOK, profile)
}