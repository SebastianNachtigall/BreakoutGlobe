package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MiddlewareTestSuite contains tests for general middleware functionality
type MiddlewareTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *MiddlewareTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
}

func (suite *MiddlewareTestSuite) TestCORSMiddleware() {
	// Setup router with CORS middleware
	suite.router.Use(CORS())
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Test preflight request
	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	req.Header.Set("Access-Control-Request-Headers", "Content-Type")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert CORS headers
	suite.Equal(http.StatusNoContent, w.Code)
	suite.Equal("*", w.Header().Get("Access-Control-Allow-Origin"))
	suite.Contains(w.Header().Get("Access-Control-Allow-Methods"), "GET")
	suite.Contains(w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	suite.Equal("true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func (suite *MiddlewareTestSuite) TestCORSMiddleware_ActualRequest() {
	// Setup router with CORS middleware
	suite.router.Use(CORS())
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Test actual request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert response and CORS headers
	suite.Equal(http.StatusOK, w.Code)
	suite.Equal("*", w.Header().Get("Access-Control-Allow-Origin"))
	suite.Equal("true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func (suite *MiddlewareTestSuite) TestSecurityHeaders() {
	// Setup router with security headers middleware
	suite.router.Use(SecurityHeaders())
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert security headers
	suite.Equal(http.StatusOK, w.Code)
	suite.Equal("nosniff", w.Header().Get("X-Content-Type-Options"))
	suite.Equal("deny", w.Header().Get("X-Frame-Options"))
	suite.Equal("1; mode=block", w.Header().Get("X-XSS-Protection"))
	suite.Equal("no-referrer-when-downgrade", w.Header().Get("Referrer-Policy"))
	suite.Contains(w.Header().Get("Content-Security-Policy"), "default-src 'self'")
}

func (suite *MiddlewareTestSuite) TestRequestTimeout() {
	// Setup router with timeout middleware
	suite.router.Use(RequestTimeout())
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert successful response
	suite.Equal(http.StatusOK, w.Code)
}

func (suite *MiddlewareTestSuite) TestRateLimitHeaders() {
	// Setup router with rate limit headers middleware
	suite.router.Use(RateLimitHeaders())
	suite.router.GET("/test", func(c *gin.Context) {
		// Simulate setting rate limit info
		c.Set("rateLimitLimit", 100)
		c.Set("rateLimitRemaining", 99)
		c.Set("rateLimitReset", 1234567890)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert rate limit headers
	suite.Equal(http.StatusOK, w.Code)
	suite.Equal("100", w.Header().Get("X-RateLimit-Limit"))
	suite.Equal("99", w.Header().Get("X-RateLimit-Remaining"))
	suite.Equal("1234567890", w.Header().Get("X-RateLimit-Reset"))
}

func (suite *MiddlewareTestSuite) TestHealthCheck() {
	// Setup router with health check middleware
	suite.router.Use(HealthCheck("/health"))
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Test health check endpoint
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert health check response
	suite.Equal(http.StatusOK, w.Code)
	suite.Contains(w.Body.String(), `"status":"healthy"`)
	suite.Contains(w.Body.String(), `"timestamp":`)
}

func (suite *MiddlewareTestSuite) TestHealthCheck_NonHealthPath() {
	// Setup router with health check middleware
	suite.router.Use(HealthCheck("/health"))
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Test non-health endpoint
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	
	suite.router.ServeHTTP(w, req)
	
	// Assert normal response
	suite.Equal(http.StatusOK, w.Code)
	suite.Contains(w.Body.String(), `"message":"success"`)
}

func TestMiddlewareTestSuite(t *testing.T) {
	suite.Run(t, new(MiddlewareTestSuite))
}

// Test individual functions
func TestIsHealthCheckPath(t *testing.T) {
	tests := []struct {
		path     string
		expected bool
	}{
		{"/health", true},
		{"/health/", true},
		{"/healthz", true},
		{"/ping", true},
		{"/status", true},
		{"/api/health", false},
		{"/test", false},
		{"/", false},
	}
	
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			result := isHealthCheckPath(tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}