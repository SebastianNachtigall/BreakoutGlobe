package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	redislib "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"breakoutglobe/internal/config"
	"breakoutglobe/internal/database"
	"breakoutglobe/internal/handlers"
	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"
	"breakoutglobe/internal/repository"
	"breakoutglobe/internal/services"
	"breakoutglobe/internal/services/uploads"
	"breakoutglobe/internal/websocket"
)

type Server struct {
	config *config.Config
	router *gin.Engine
	db     *gorm.DB
	redis  *redislib.Client
	// Simple in-memory storage for POI participants (for testing)
	poiParticipants map[string]map[string]string // poiId -> sessionId -> username
	// POI service for WebSocket handler
	poiService *services.POIService
}

func New(cfg *config.Config) *Server {
	log.Printf("🚀 Creating new server with config: GinMode=%s, DatabaseURL=%s", cfg.GinMode, cfg.DatabaseURL)
	gin.SetMode(cfg.GinMode)
	
	var db *gorm.DB
	var redisClient *redislib.Client
	
	// Only connect to database and Redis if not in test mode
	if cfg.GinMode != "test" {
		log.Printf("🔗 Attempting to connect to database: %s", cfg.DatabaseURL)
		
		// Setup database connection and run migrations
		var err error
		db, err = database.Initialize(cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("❌ Failed to initialize database: %v", err)
		}
		
		log.Println("✅ Database connection established and migrations completed")
		
		// Setup Redis connection
		log.Printf("🔗 Attempting to connect to Redis: %s", cfg.RedisURL)
		redisConfig := redis.Config{
			Addr:     strings.TrimPrefix(cfg.RedisURL, "redis://"),
			Password: "",
			DB:       0,
		}
		redisClient = redis.NewClient(redisConfig)
		
		if err := redis.TestConnection(redisClient); err != nil {
			log.Fatalf("❌ Failed to connect to Redis: %v", err)
		}
		
		log.Println("✅ Redis connection established")
	} else {
		log.Println("⚠️ Running in test mode, skipping database and Redis connections")
	}
	
	router := gin.Default()
	
	// CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID"},
		AllowCredentials: true,
	}))
	
	s := &Server{
		config: cfg,
		router: router,
		db:     db,
		redis:  redisClient,
		poiParticipants: make(map[string]map[string]string),
	}
	
	s.setupRoutes()
	
	return s
}

func (s *Server) setupRoutes() {
	log.Println("🔧 Setting up routes...")
	
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
		
		// Setup POI routes with proper handlers
		s.setupPOIRoutes(api)
		
		// User profile endpoints with proper handlers
		log.Println("About to call setupUserRoutes")
		s.setupUserRoutes(api)
		log.Println("setupUserRoutes call completed")
		
		// Serve uploaded avatar files
		api.GET("/users/avatar/:filename", s.serveAvatar)
		
		// Serve uploaded POI images
		s.router.Static("/uploads", "./uploads")
	}
	
	// WebSocket handler setup removed during phantom debugging
}

func (s *Server) setupUserRoutes(api *gin.RouterGroup) {
	log.Printf("🔧 setupUserRoutes called, db is nil: %v", s.db == nil)
	
	// Only setup user routes if database is available (not in test mode)
	if s.db != nil {
		log.Println("📊 Database available, setting up user routes and WebSocket handler")
		
		// Setup dependencies
		userRepo := repository.NewUserRepository(s.db)
		userService := services.NewUserService(userRepo)
		
		// For now, create a simple in-memory rate limiter (TODO: use Redis in production)
		rateLimiter := &SimpleRateLimiter{}
		
		userHandler := handlers.NewUserHandler(userService, rateLimiter)
		
		// Register user routes
		userHandler.RegisterRoutes(s.router)
		
		// Setup WebSocket handler for multi-user functionality
		s.setupWebSocketHandler(userService, rateLimiter, s.poiService)
	} else {
		log.Println("⚠️ Database not available, skipping user routes and WebSocket handler setup")
	}
}

func (s *Server) setupPOIRoutes(api *gin.RouterGroup) {
	log.Printf("🔧 setupPOIRoutes called, db is nil: %v, redis is nil: %v", s.db == nil, s.redis == nil)
	
	// Only setup POI routes if database and Redis are available (not in test mode)
	if s.db != nil && s.redis != nil {
		log.Println("📊 Database and Redis available, setting up POI routes with proper handlers")
		
		// Setup dependencies
		poiRepo := repository.NewPOIRepository(s.db)
		poiParticipants := redis.NewPOIParticipants(s.redis)
		pubsub := redis.NewPubSub(s.redis)
		
		// Create user service for participant name resolution
		userRepo := repository.NewUserRepository(s.db)
		userService := services.NewUserService(userRepo)
		
		// Create image uploader
		uploadDir := filepath.Join(".", "uploads")
		baseURL := "http://localhost:8080" // TODO: Make this configurable
		imageUploader := uploads.NewImageUploader(uploadDir, baseURL)
		
		// Create POI service with image uploader and user service
		s.poiService = services.NewPOIServiceWithImageUploader(poiRepo, poiParticipants, pubsub, imageUploader, userService)
		
		// Create rate limiter (simple in-memory for now)
		rateLimiter := &SimpleRateLimiter{}
		
		// Create POI handler with user service for participant names
		poiHandler := handlers.NewPOIHandler(s.poiService, userService, rateLimiter)
		
		// Register POI routes
		poiHandler.RegisterRoutes(s.router)
		
		log.Println("✅ POI routes setup complete with database-backed handlers")
	} else {
		log.Println("⚠️ Database or Redis not available, using mock POI handlers")
		
		// Fallback to mock handlers for testing
		api.GET("/pois", s.getPOIs)
		api.POST("/pois", s.createPOI)
		api.POST("/pois/:poiId/join", s.joinPOI)
		api.POST("/pois/:poiId/leave", s.leavePOI)
	}
}

func (s *Server) setupWebSocketHandler(userService *services.UserService, rateLimiter services.RateLimiterInterface, poiService *services.POIService) {
	log.Println("🔧 Setting up WebSocket handler...")
	
	// Create a simple session service adapter for the WebSocket handler
	sessionService := &SimpleSessionService{
		server:    s,
		positions: make(map[string]models.LatLng),
	}
	
	// Create WebSocket handler
	wsHandler := websocket.NewHandler(sessionService, rateLimiter, userService, poiService)
	
	// Set up PubSub integration if Redis is available
	if s.redis != nil {
		pubsub := redis.NewPubSub(s.redis)
		wsHandler.SetPubSub(pubsub)
		log.Println("✅ WebSocket handler PubSub integration enabled")
	} else {
		log.Println("⚠️ Redis not available, WebSocket handler will not receive real-time POI events")
	}
	
	// Register the WebSocket handler
	s.router.GET("/ws", wsHandler.HandleWebSocket)
	
	log.Println("✅ WebSocket handler setup complete - using proper multi-user handler")
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



// serveAvatar serves uploaded avatar files with comprehensive security validation
// Implements secure file serving with:
// - Path traversal attack prevention
// - File type validation (only images)
// - File size limits
// - Proper caching headers
// - MIME type detection
func (s *Server) serveAvatar(c *gin.Context) {
	filename := c.Param("filename")
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "filename is required"})
		return
	}
	
	// Security validation: prevent path traversal attacks
	// Check for various path traversal patterns
	if strings.Contains(filename, "..") || 
	   strings.Contains(filename, "/") || 
	   strings.Contains(filename, `\`) ||
	   strings.Contains(filename, "%2e%2e") || // URL encoded ..
	   strings.Contains(filename, "%2f") ||    // URL encoded /
	   strings.Contains(filename, "%5c") {     // URL encoded \
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
		return
	}
	
	// Validate file extension (only allow image files)
	ext := strings.ToLower(filepath.Ext(filename))
	allowedExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	if !allowedExtensions[ext] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file type"})
		return
	}
	
	// Construct file path (safe after validation)
	filePath := filepath.Join("uploads", "avatars", filename)
	
	// Check if file exists
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "avatar not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "file access error"})
		return
	}
	
	// Check file size (max 2MB for serving)
	if fileInfo.Size() > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		return
	}
	
	// Set proper cache headers for avatar files
	c.Header("Cache-Control", "public, max-age=3600") // Cache for 1 hour
	c.Header("ETag", fmt.Sprintf(`"%d-%d"`, fileInfo.Size(), fileInfo.ModTime().Unix()))
	
	// Set content type based on file extension
	switch ext {
	case ".jpg", ".jpeg":
		c.Header("Content-Type", "image/jpeg")
	case ".png":
		c.Header("Content-Type", "image/png")
	}
	
	// Serve the file
	c.File(filePath)
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

// SimpleSessionService is an adapter that makes the server's session functions work with WebSocket handler
type SimpleSessionService struct {
	server *Server
	// In-memory storage for session positions (in production, this would be in Redis/DB)
	positions map[string]models.LatLng
	mutex     sync.RWMutex
}

func (s *SimpleSessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	// Parse the session ID to extract user ID and map ID
	// Session ID format: "session-{userID}-{mapID}"
	// Note: userID is a UUID with hyphens, so we need to be careful with parsing
	
	if !strings.HasPrefix(sessionID, "session-") {
		return nil, fmt.Errorf("invalid session ID format: must start with 'session-'")
	}
	
	// Remove "session-" prefix
	remainder := sessionID[8:] // len("session-") = 8
	
	// Find the last occurrence of "-default-map" to extract the mapID
	mapSuffix := "-default-map"
	mapIndex := strings.LastIndex(remainder, mapSuffix)
	if mapIndex == -1 {
		return nil, fmt.Errorf("invalid session ID format: must end with '-default-map'")
	}
	
	userID := remainder[:mapIndex]
	mapID := remainder[mapIndex+1:] // Skip the "-" before "default-map"
	
	// Get stored position or use default
	s.mutex.RLock()
	position, exists := s.positions[sessionID]
	s.mutex.RUnlock()
	
	if !exists {
		position = models.LatLng{Lat: 40.7128, Lng: -74.0060} // Default position
	}
	
	// Create a mock session for WebSocket handler
	session := &models.Session{
		ID:       sessionID,
		UserID:   userID,
		MapID:    mapID,
		AvatarPos: position,
		IsActive: true,
	}
	
	return session, nil
}

func (s *SimpleSessionService) SessionHeartbeat(ctx context.Context, sessionID string) error {
	// For now, just return success
	return nil
}

func (s *SimpleSessionService) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	// Store the position in memory
	s.mutex.Lock()
	s.positions[sessionID] = position
	s.mutex.Unlock()
	
	return nil
}

