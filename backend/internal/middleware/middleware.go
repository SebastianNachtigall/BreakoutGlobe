package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set CORS headers
		c.Header("Access-Control-Allow-Origin", "*") // In production, be more specific
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-RateLimit-Limit, X-RateLimit-Remaining, X-RateLimit-Reset")
		
		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		c.Next()
	}
}

// SecurityHeaders returns a middleware that adds security headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")
		
		// Prevent clickjacking
		c.Header("X-Frame-Options", "deny")
		
		// Enable XSS protection
		c.Header("X-XSS-Protection", "1; mode=block")
		
		// Control referrer information
		c.Header("Referrer-Policy", "no-referrer-when-downgrade")
		
		// Content Security Policy
		csp := "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' ws: wss:; " +
			"frame-ancestors 'none'"
		c.Header("Content-Security-Policy", csp)
		
		c.Next()
	}
}

// RequestTimeout returns a middleware that adds request timeout
func RequestTimeout() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set a timeout for the request context
		ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
		defer cancel()
		
		// Replace the request context
		c.Request = c.Request.WithContext(ctx)
		
		c.Next()
	}
}

// RateLimitHeaders returns a middleware that adds rate limit headers
func RateLimitHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// Add rate limit headers if they were set by rate limiting middleware
		if limit, exists := c.Get("rateLimitLimit"); exists {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%v", limit))
		}
		
		if remaining, exists := c.Get("rateLimitRemaining"); exists {
			c.Header("X-RateLimit-Remaining", fmt.Sprintf("%v", remaining))
		}
		
		if reset, exists := c.Get("rateLimitReset"); exists {
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%v", reset))
		}
	}
}

// HealthCheck returns a middleware that handles health check requests
func HealthCheck(path string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == path || isHealthCheckPath(c.Request.URL.Path) {
			c.JSON(http.StatusOK, gin.H{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
				"service":   "breakout-globe-api",
			})
			c.Abort()
			return
		}
		
		c.Next()
	}
}

// isHealthCheckPath checks if the path is a common health check path
func isHealthCheckPath(path string) bool {
	healthPaths := []string{
		"/health",
		"/health/",
		"/healthz",
		"/ping",
		"/status",
	}
	
	for _, healthPath := range healthPaths {
		if path == healthPath {
			return true
		}
	}
	
	return false
}



// ContentTypeJSON returns a middleware that sets JSON content type for API responses
func ContentTypeJSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only set for API routes
		if strings.HasPrefix(c.Request.URL.Path, "/api/") {
			c.Header("Content-Type", "application/json; charset=utf-8")
		}
		
		c.Next()
	}
}

// RequestSizeLimit returns a middleware that limits request body size
func RequestSizeLimit(maxSize int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > maxSize {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, ErrorResponse{
				Code:      "REQUEST_TOO_LARGE",
				Message:   "Request body too large",
				RequestID: c.GetString("requestID"),
				Timestamp: time.Now(),
			})
			return
		}
		
		// Limit the request body reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		
		c.Next()
	}
}

// APIVersionHeader returns a middleware that adds API version header
func APIVersionHeader(version string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-API-Version", version)
		c.Next()
	}
}

// MiddlewareConfig holds configuration for all middleware
type MiddlewareConfig struct {
	EnableCORS           bool
	EnableSecurityHeaders bool
	EnableRequestTimeout bool
	EnableHealthCheck    bool
	HealthCheckPath      string
	RequestSizeLimit     int64
	APIVersion           string
	SkipLoggingPaths     []string
}

// DefaultMiddlewareConfig returns default middleware configuration
func DefaultMiddlewareConfig() MiddlewareConfig {
	return MiddlewareConfig{
		EnableCORS:           true,
		EnableSecurityHeaders: true,
		EnableRequestTimeout: true,
		EnableHealthCheck:    true,
		HealthCheckPath:      "/health",
		RequestSizeLimit:     10 * 1024 * 1024, // 10MB
		APIVersion:           "v1",
		SkipLoggingPaths:     []string{"/health", "/healthz", "/ping"},
	}
}

// SetupMiddleware configures all middleware for the Gin engine
func SetupMiddleware(router *gin.Engine, config MiddlewareConfig, logger *slog.Logger) {
	// Request ID middleware (should be first)
	router.Use(RequestID())
	
	// Request logging middleware
	if logger != nil {
		loggerConfig := LoggerConfig{
			Logger:    logger,
			SkipPaths: config.SkipLoggingPaths,
		}
		router.Use(RequestLoggerWithConfig(loggerConfig))
	}
	
	// Error handling middleware
	router.Use(ErrorHandler())
	
	// Health check middleware (before other middleware to avoid unnecessary processing)
	if config.EnableHealthCheck {
		router.Use(HealthCheck(config.HealthCheckPath))
	}
	
	// CORS middleware
	if config.EnableCORS {
		router.Use(CORS())
	}
	
	// Security headers middleware
	if config.EnableSecurityHeaders {
		router.Use(SecurityHeaders())
	}
	
	// Request timeout middleware
	if config.EnableRequestTimeout {
		router.Use(RequestTimeout())
	}
	
	// Request size limit middleware
	if config.RequestSizeLimit > 0 {
		router.Use(RequestSizeLimit(config.RequestSizeLimit))
	}
	
	// Content type middleware
	router.Use(ContentTypeJSON())
	
	// API version header middleware
	if config.APIVersion != "" {
		router.Use(APIVersionHeader(config.APIVersion))
	}
	
	// Rate limit headers middleware
	router.Use(RateLimitHeaders())
	
	// Setup 404 and 405 handlers
	router.NoRoute(NoRouteHandler())
	router.NoMethod(NoMethodHandler())
}