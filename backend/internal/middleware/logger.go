package middleware

import (
	"log/slog"
	"net"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger returns a middleware that logs HTTP requests with structured logging
func RequestLogger(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		
		// Generate or extract request ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = c.GetString("requestID")
			if requestID == "" {
				requestID = "unknown"
			}
		}
		c.Set("requestID", requestID)
		
		// Extract request information
		method := c.Request.Method
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		userAgent := c.Request.UserAgent()
		clientIP := getClientIP(c)
		contentLength := c.Request.ContentLength
		
		// Process request
		c.Next()
		
		// Calculate duration
		duration := time.Since(start)
		
		// Extract response information
		status := c.Writer.Status()
		responseSize := c.Writer.Size()
		
		// Extract user context if available
		userID := c.GetString("userID")
		sessionID := c.GetString("sessionID")
		
		// Determine log level based on status code
		var logLevel slog.Level
		switch {
		case status >= 500:
			logLevel = slog.LevelError
		case status >= 400:
			logLevel = slog.LevelWarn
		default:
			logLevel = slog.LevelInfo
		}
		
		// Build log attributes
		attrs := []slog.Attr{
			slog.String("request_id", requestID),
			slog.String("method", method),
			slog.String("path", path),
			slog.Int("status", status),
			slog.Float64("duration", float64(duration.Nanoseconds())/1e6), // Duration in milliseconds
			slog.String("client_ip", clientIP),
			slog.String("user_agent", userAgent),
			slog.Int64("content_length", contentLength),
			slog.Int("response_size", responseSize),
		}
		
		// Add query parameters if present
		if query != "" {
			attrs = append(attrs, slog.String("query", query))
		}
		
		// Add user context if available
		if userID != "" {
			attrs = append(attrs, slog.String("user_id", userID))
		}
		if sessionID != "" {
			attrs = append(attrs, slog.String("session_id", sessionID))
		}
		
		// Add error information if present
		if len(c.Errors) > 0 {
			attrs = append(attrs, slog.String("error", c.Errors.String()))
		}
		
		// Add filtered headers for debugging (excluding sensitive data)
		headers := make(map[string]string)
		for key, values := range c.Request.Header {
			if len(values) > 0 {
				headers[key] = values[0]
			}
		}
		filteredHeaders := filterSensitiveData(path, headers)
		if len(filteredHeaders) > 0 {
			// Convert to a more compact representation for logging
			var headerPairs []string
			for key, value := range filteredHeaders {
				if !isSensitiveHeader(key) {
					headerPairs = append(headerPairs, key+":"+value)
				}
			}
			if len(headerPairs) > 0 {
				attrs = append(attrs, slog.String("headers", strings.Join(headerPairs, ",")))
			}
		}
		
		// Log the request
		logger.LogAttrs(c.Request.Context(), logLevel, "HTTP Request", attrs...)
	}
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check X-Forwarded-For header first
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check X-Real-IP header
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}
	return ip
}

// filterSensitiveData removes or masks sensitive information from headers
func filterSensitiveData(path string, headers map[string]string) map[string]string {
	filtered := make(map[string]string)
	
	for key, value := range headers {
		if isSensitiveHeader(key) {
			filtered[key] = "[FILTERED]"
		} else {
			filtered[key] = value
		}
	}
	
	return filtered
}

// isSensitiveHeader checks if a header contains sensitive information
func isSensitiveHeader(header string) bool {
	sensitiveHeaders := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
		"x-access-token",
		"x-csrf-token",
		"x-session-token",
	}
	
	headerLower := strings.ToLower(header)
	for _, sensitive := range sensitiveHeaders {
		if headerLower == sensitive {
			return true
		}
	}
	
	return false
}

// StructuredLogger creates a structured logger for the application
func StructuredLogger() *slog.Logger {
	opts := &slog.HandlerOptions{
		Level: slog.LevelInfo,
		AddSource: true,
	}
	
	handler := slog.NewJSONHandler(gin.DefaultWriter, opts)
	return slog.New(handler)
}

// LoggerConfig holds configuration for the logger middleware
type LoggerConfig struct {
	Logger        *slog.Logger
	SkipPaths     []string
	SkipUserAgent []string
}

// RequestLoggerWithConfig returns a middleware with custom configuration
func RequestLoggerWithConfig(config LoggerConfig) gin.HandlerFunc {
	logger := config.Logger
	if logger == nil {
		logger = StructuredLogger()
	}
	
	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}
	
	skipUserAgents := make(map[string]bool)
	for _, ua := range config.SkipUserAgent {
		skipUserAgents[ua] = true
	}
	
	return func(c *gin.Context) {
		// Skip logging for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}
		
		// Skip logging for certain user agents (e.g., health checks)
		if skipUserAgents[c.Request.UserAgent()] {
			c.Next()
			return
		}
		
		// Use the main logger middleware
		RequestLogger(logger)(c)
	}
}