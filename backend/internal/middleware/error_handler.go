package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Code      string    `json:"code"`
	Message   string    `json:"message"`
	Details   string    `json:"details,omitempty"`
	RequestID string    `json:"requestId"`
	Timestamp time.Time `json:"timestamp"`
}

// ErrorHandler returns a middleware that handles errors and panics
func ErrorHandler() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		requestID := c.GetString("requestID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// Handle panic
		if recovered != nil {
			c.Header("Content-Type", "application/json")
			c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorResponse{
				Code:      "INTERNAL_ERROR",
				Message:   "Internal server error",
				RequestID: requestID,
				Timestamp: time.Now(),
			})
			return
		}
	})
}

// ErrorHandlerMiddleware returns a middleware that handles errors after request processing
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		requestID := c.GetString("requestID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// Handle errors that occurred during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			code, message, status := classifyError(err.Err)
			
			c.Header("Content-Type", "application/json")
			c.AbortWithStatusJSON(status, ErrorResponse{
				Code:      code,
				Message:   message,
				Details:   err.Error(),
				RequestID: requestID,
				Timestamp: time.Now(),
			})
			return
		}
	}
}

// NoRouteHandler handles 404 errors
func NoRouteHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:      "NOT_FOUND",
			Message:   "Endpoint not found",
			RequestID: requestID,
			Timestamp: time.Now(),
		})
	}
}

// NoMethodHandler handles 405 errors
func NoMethodHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := generateRequestID()
		c.JSON(http.StatusMethodNotAllowed, ErrorResponse{
			Code:      "METHOD_NOT_ALLOWED",
			Message:   "Method not allowed",
			RequestID: requestID,
			Timestamp: time.Now(),
		})
	}
}

// RequestID returns a middleware that adds a request ID to the context
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request ID is already provided
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		
		// Set request ID in context and response header
		c.Set("requestID", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	bytes := make([]byte, 4)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// classifyError classifies an error and returns appropriate code, message, and status
func classifyError(err error) (code, message string, status int) {
	if err == nil {
		return "UNKNOWN_ERROR", "Unknown error", http.StatusInternalServerError
	}
	
	errStr := err.Error()
	
	// JSON parsing errors
	if strings.Contains(errStr, "invalid character") ||
		strings.Contains(errStr, "unexpected end of JSON input") ||
		strings.Contains(errStr, "cannot unmarshal") {
		return "INVALID_JSON", "Invalid JSON format", http.StatusBadRequest
	}
	
	// Validation errors (Gin binding)
	if strings.Contains(errStr, "Field validation") ||
		strings.Contains(errStr, "required") ||
		strings.Contains(errStr, "Error:Field") {
		return "VALIDATION_ERROR", "Request validation failed", http.StatusBadRequest
	}
	
	// Content type errors
	if strings.Contains(errStr, "content-type") ||
		strings.Contains(errStr, "Content-Type") {
		return "INVALID_CONTENT_TYPE", "Invalid content type", http.StatusBadRequest
	}
	
	// Request size errors
	if strings.Contains(errStr, "request body too large") ||
		strings.Contains(errStr, "http: request body too large") {
		return "REQUEST_TOO_LARGE", "Request body too large", http.StatusRequestEntityTooLarge
	}
	
	// Timeout errors
	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline exceeded") {
		return "REQUEST_TIMEOUT", "Request timeout", http.StatusRequestTimeout
	}
	
	// Default to bad request for client errors
	return "BAD_REQUEST", "Bad request", http.StatusBadRequest
}