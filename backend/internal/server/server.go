package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
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
		redisConfig, err := parseRedisURL(cfg.RedisURL)
		if err != nil {
			log.Fatalf("❌ Failed to parse Redis URL: %v", err)
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
		AllowOrigins:     []string{
			"http://localhost:3000",                                    // Local development
			"https://frontend-production-0050.up.railway.app",         // Railway production
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-User-ID"},
		AllowCredentials: true,
	}))
	
	s := &Server{
		config: cfg,
		router: router,
		db:     db,
		redis:  redisClient,
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
		
		// Setup session routes with proper handlers
		s.setupSessionRoutes(api)
		
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

func (s *Server) setupSessionRoutes(api *gin.RouterGroup) {
	log.Printf("🔧 setupSessionRoutes called, db is nil: %v, redis is nil: %v", s.db == nil, s.redis == nil)
	
	// Only setup session routes if database and Redis are available (not in test mode)
	if s.db != nil && s.redis != nil {
		log.Println("📊 Database and Redis available, setting up session routes with proper handlers")
		
		// Setup dependencies
		sessionRepo := repository.NewSessionRepository(s.db)
		sessionPresence := redis.NewSessionPresence(s.redis)
		pubsub := redis.NewPubSub(s.redis) // Add the missing pubsub parameter
		sessionService := services.NewSessionService(sessionRepo, sessionPresence, pubsub)
		
		// Create rate limiter (simple in-memory for now)
		rateLimiter := &SimpleRateLimiter{}
		
		// Create session handler
		sessionHandler := handlers.NewSessionHandler(sessionService, rateLimiter)
		
		// Register session routes (this includes the heartbeat endpoint)
		sessionHandler.RegisterRoutes(s.router)
		
		log.Println("✅ Session routes setup complete with proper handlers including heartbeat")
	} else {
		log.Println("⚠️ Database or Redis not available, session endpoints not available in test mode")
		// No fallback handlers - proper service-backed handlers only
	}
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
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:8080" // Fallback for local development
		}
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
		log.Println("⚠️ Database or Redis not available, POI endpoints not available in test mode")
		// No fallback handlers - proper service-backed handlers only
	}
}

func (s *Server) setupWebSocketHandler(userService *services.UserService, rateLimiter services.RateLimiterInterface, poiService *services.POIService) {
	log.Println("🔧 Setting up WebSocket handler...")
	
	// Use the proper session service with database validation
	sessionRepo := repository.NewSessionRepository(s.db)
	sessionPresence := redis.NewSessionPresence(s.redis)
	pubsub := redis.NewPubSub(s.redis)
	sessionService := services.NewSessionService(sessionRepo, sessionPresence, pubsub)
	
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

// Mock handlers removed - using proper service-backed handlers only

// POI mock handlers removed - using proper service-backed handlers only



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

// User profile mock handlers removed - using proper service-backed handlers only

// SimpleSessionService mock removed - using proper SessionService only

// parseRedisURL parses a Redis URL and returns a Redis config
// Supports formats like: redis://user:password@host:port/db
func parseRedisURL(redisURL string) (redis.Config, error) {
	config := redis.Config{
		Password: "",
		DB:       0,
	}
	
	// Handle simple localhost case
	if redisURL == "redis://localhost:6379" {
		config.Addr = "localhost:6379"
		return config, nil
	}
	
	// Parse the URL
	u, err := url.Parse(redisURL)
	if err != nil {
		return config, fmt.Errorf("invalid Redis URL: %w", err)
	}
	
	// Extract host and port
	config.Addr = u.Host
	
	// Extract password if present
	if u.User != nil {
		if password, ok := u.User.Password(); ok {
			config.Password = password
		}
	}
	
	return config, nil
}