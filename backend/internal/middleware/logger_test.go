package middleware

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// LoggerTestSuite contains the test suite for logging middleware
type LoggerTestSuite struct {
	suite.Suite
	router    *gin.Engine
	logBuffer *bytes.Buffer
	logger    *slog.Logger
}

func (suite *LoggerTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Create a buffer to capture log output
	suite.logBuffer = &bytes.Buffer{}
	
	// Create a logger that writes to the buffer
	suite.logger = slog.New(slog.NewJSONHandler(suite.logBuffer, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	
	// Setup router with logging middleware
	suite.router = gin.New()
	suite.router.Use(RequestLogger(suite.logger))
}

func (suite *LoggerTestSuite) TestRequestLogger_BasicLogging() {
	// Setup test route
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "192.168.1.1")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	suite.NotEmpty(logOutput)
	
	// Should contain request log
	suite.Contains(logOutput, `"level":"INFO"`)
	suite.Contains(logOutput, `"msg":"HTTP Request"`)
	suite.Contains(logOutput, `"method":"GET"`)
	suite.Contains(logOutput, `"path":"/test"`)
	suite.Contains(logOutput, `"status":200`)
	suite.Contains(logOutput, `"user_agent":"test-agent"`)
	suite.Contains(logOutput, `"client_ip":"192.168.1.1"`)
	suite.Contains(logOutput, `"duration":`)
}

func (suite *LoggerTestSuite) TestRequestLogger_WithUserContext() {
	// Setup test route that sets user context
	suite.router.POST("/user-action", func(c *gin.Context) {
		// Simulate setting user context
		c.Set("userID", "user-123")
		c.Set("sessionID", "session-456")
		c.JSON(http.StatusCreated, gin.H{"success": true})
	})
	
	// Create request
	reqBody := `{"action": "test"}`
	req := httptest.NewRequest(http.MethodPost, "/user-action", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusCreated, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should contain user context
	suite.Contains(logOutput, `"user_id":"user-123"`)
	suite.Contains(logOutput, `"session_id":"session-456"`)
	suite.Contains(logOutput, `"method":"POST"`)
	suite.Contains(logOutput, `"status":201`)
	suite.Contains(logOutput, `"content_length":18`) // Length of request body
}

func (suite *LoggerTestSuite) TestRequestLogger_ErrorLogging() {
	// Setup test route that returns an error
	suite.router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "test error"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusBadRequest, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should log as WARNING for 4xx errors
	suite.Contains(logOutput, `"level":"WARN"`)
	suite.Contains(logOutput, `"status":400`)
}

func (suite *LoggerTestSuite) TestRequestLogger_ServerErrorLogging() {
	// Setup test route that returns server error
	suite.router.GET("/server-error", func(c *gin.Context) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server error"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/server-error", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusInternalServerError, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should log as ERROR for 5xx errors
	suite.Contains(logOutput, `"level":"ERROR"`)
	suite.Contains(logOutput, `"status":500`)
}

func (suite *LoggerTestSuite) TestRequestLogger_SlowRequestLogging() {
	// Setup test route that takes time
	suite.router.GET("/slow", func(c *gin.Context) {
		time.Sleep(100 * time.Millisecond)
		c.JSON(http.StatusOK, gin.H{"message": "slow response"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should contain duration information
	suite.Contains(logOutput, `"duration":`)
	
	// Parse the log to check duration
	var logEntry map[string]interface{}
	lines := strings.Split(strings.TrimSpace(logOutput), "\n")
	err := json.Unmarshal([]byte(lines[0]), &logEntry)
	suite.NoError(err)
	
	duration, ok := logEntry["duration"].(float64)
	suite.True(ok)
	suite.Greater(duration, 100.0) // Should be at least 100ms
}

func (suite *LoggerTestSuite) TestRequestLogger_QueryParameters() {
	// Setup test route
	suite.router.GET("/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"results": []string{}})
	})
	
	// Create request with query parameters
	req := httptest.NewRequest(http.MethodGet, "/search?q=test&limit=10&offset=0", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should contain query parameters
	suite.Contains(logOutput, `"query":"q=test&limit=10&offset=0"`)
}

func (suite *LoggerTestSuite) TestRequestLogger_SensitiveDataFiltering() {
	// Setup test route
	suite.router.POST("/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"token": "secret-token"})
	})
	
	// Create request with sensitive data
	reqBody := `{"password": "secret123", "email": "user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer secret-token")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should not contain sensitive data
	suite.NotContains(logOutput, "secret123")
	suite.NotContains(logOutput, "secret-token")
	
	// Should contain filtered indicators
	suite.Contains(logOutput, `"path":"/login"`)
}

func (suite *LoggerTestSuite) TestRequestLogger_RequestIDPropagation() {
	// Setup test route that uses request ID
	suite.router.GET("/request-id", func(c *gin.Context) {
		requestID := c.GetString("requestID")
		c.JSON(http.StatusOK, gin.H{"requestId": requestID})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/request-id", nil)
	req.Header.Set("X-Request-ID", "custom-request-id")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert response
	suite.Equal(http.StatusOK, w.Code)
	
	// Parse log output
	logOutput := suite.logBuffer.String()
	
	// Should contain request ID
	suite.Contains(logOutput, `"request_id":`)
}

func TestLoggerTestSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

// Test helper functions
func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name           string
		headers        map[string]string
		remoteAddr     string
		expectedIP     string
	}{
		{
			name:       "X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "192.168.1.1, 10.0.0.1"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.1",
		},
		{
			name:       "X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "192.168.1.2"},
			remoteAddr: "127.0.0.1:8080",
			expectedIP: "192.168.1.2",
		},
		{
			name:       "Remote address fallback",
			headers:    map[string]string{},
			remoteAddr: "192.168.1.3:8080",
			expectedIP: "192.168.1.3",
		},
		{
			name:       "IPv6 address",
			headers:    map[string]string{},
			remoteAddr: "[::1]:8080",
			expectedIP: "::1",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock request
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}
			
			// Create Gin context
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			ip := getClientIP(c)
			assert.Equal(t, tt.expectedIP, ip)
		})
	}
}

func TestFilterSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		headers  map[string]string
		expected map[string]string
	}{
		{
			name: "Filter authorization header",
			path: "/api/test",
			headers: map[string]string{
				"Authorization": "Bearer secret-token",
				"Content-Type":  "application/json",
			},
			expected: map[string]string{
				"Authorization": "[FILTERED]",
				"Content-Type":  "application/json",
			},
		},
		{
			name: "Filter cookie header",
			path: "/api/test",
			headers: map[string]string{
				"Cookie":       "session=secret; user=test",
				"Content-Type": "application/json",
			},
			expected: map[string]string{
				"Cookie":       "[FILTERED]",
				"Content-Type": "application/json",
			},
		},
		{
			name: "No filtering needed",
			path: "/api/test",
			headers: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "test-agent",
			},
			expected: map[string]string{
				"Content-Type": "application/json",
				"User-Agent":   "test-agent",
			},
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := filterSensitiveData(tt.path, tt.headers)
			assert.Equal(t, tt.expected, result)
		})
	}
}