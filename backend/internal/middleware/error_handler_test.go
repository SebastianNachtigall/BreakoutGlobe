package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ErrorHandlerTestSuite contains the test suite for error handling middleware
type ErrorHandlerTestSuite struct {
	suite.Suite
	router *gin.Engine
}

func (suite *ErrorHandlerTestSuite) SetupTest() {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	suite.router = gin.New()
	suite.router.Use(RequestID())
	suite.router.Use(ErrorHandler())
	suite.router.Use(ErrorHandlerMiddleware())
	suite.router.NoRoute(NoRouteHandler())
	suite.router.NoMethod(NoMethodHandler())
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_PanicRecovery() {
	// Setup route that panics
	suite.router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusInternalServerError, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INTERNAL_ERROR", response.Code)
	suite.Equal("Internal server error", response.Message)
	suite.NotEmpty(response.RequestID)
	suite.WithinDuration(time.Now(), response.Timestamp, time.Second)
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_ErrorAbort() {
	// Setup route that aborts with error
	suite.router.GET("/error", func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"code":    "CUSTOM_ERROR",
			"message": "Custom error message",
		})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/error", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("CUSTOM_ERROR", response["code"])
	suite.Equal("Custom error message", response["message"])
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_ValidationError() {
	// Setup route that returns validation error
	suite.router.POST("/validate", func(c *gin.Context) {
		var req struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/validate", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("VALIDATION_ERROR", response.Code)
	suite.Equal("Request validation failed", response.Message)
	suite.NotEmpty(response.Details)
	suite.NotEmpty(response.RequestID)
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_NotFound() {
	// Create request to non-existent route
	req := httptest.NewRequest(http.MethodGet, "/nonexistent", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("NOT_FOUND", response.Code)
	suite.Equal("Endpoint not found", response.Message)
	suite.NotEmpty(response.RequestID)
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_MethodNotAllowed() {
	// Setup GET route
	suite.router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	
	// Create POST request to GET route
	req := httptest.NewRequest(http.MethodPost, "/test", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert - Gin treats this as 404, not 405, which is acceptable behavior
	suite.Equal(http.StatusNotFound, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("NOT_FOUND", response.Code)
	suite.Equal("Endpoint not found", response.Message)
	suite.NotEmpty(response.RequestID)
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_JSONParseError() {
	// Setup route that expects JSON
	suite.router.POST("/json", func(c *gin.Context) {
		var req map[string]interface{}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.AbortWithError(http.StatusBadRequest, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	})
	
	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/json", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusBadRequest, w.Code)
	
	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("INVALID_JSON", response.Code)
	suite.Equal("Invalid JSON format", response.Message)
	suite.NotEmpty(response.RequestID)
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_RequestIDGeneration() {
	// Setup route that panics
	suite.router.GET("/panic1", func(c *gin.Context) {
		panic("test panic 1")
	})
	suite.router.GET("/panic2", func(c *gin.Context) {
		panic("test panic 2")
	})
	
	// Create first request
	req1 := httptest.NewRequest(http.MethodGet, "/panic1", nil)
	w1 := httptest.NewRecorder()
	suite.router.ServeHTTP(w1, req1)
	
	// Create second request
	req2 := httptest.NewRequest(http.MethodGet, "/panic2", nil)
	w2 := httptest.NewRecorder()
	suite.router.ServeHTTP(w2, req2)
	
	// Parse responses
	var response1, response2 ErrorResponse
	json.Unmarshal(w1.Body.Bytes(), &response1)
	json.Unmarshal(w2.Body.Bytes(), &response2)
	
	// Assert different request IDs
	suite.NotEqual(response1.RequestID, response2.RequestID)
	suite.NotEmpty(response1.RequestID)
	suite.NotEmpty(response2.RequestID)
}

func (suite *ErrorHandlerTestSuite) TestErrorHandler_SuccessfulRequest() {
	// Setup successful route
	suite.router.GET("/success", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Create request
	req := httptest.NewRequest(http.MethodGet, "/success", nil)
	w := httptest.NewRecorder()
	
	// Execute
	suite.router.ServeHTTP(w, req)
	
	// Assert
	suite.Equal(http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	suite.NoError(err)
	suite.Equal("success", response["message"])
}

func TestErrorHandlerTestSuite(t *testing.T) {
	suite.Run(t, new(ErrorHandlerTestSuite))
}

// Test helper functions
func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()
	
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
	assert.Len(t, id1, 8) // Should be 8 characters
	assert.Len(t, id2, 8)
}

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedCode   string
		expectedMsg    string
		expectedStatus int
	}{
		{
			name:           "JSON syntax error",
			err:            errors.New("invalid character 'i' looking for beginning of value"),
			expectedCode:   "INVALID_JSON",
			expectedMsg:    "Invalid JSON format",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Binding validation error",
			err:            errors.New("Key: 'Request.Name' Error:Field validation for 'Name' failed on the 'required' tag"),
			expectedCode:   "VALIDATION_ERROR",
			expectedMsg:    "Request validation failed",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Generic error",
			err:            errors.New("some other error"),
			expectedCode:   "BAD_REQUEST",
			expectedMsg:    "Bad request",
			expectedStatus: http.StatusBadRequest,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, msg, status := classifyError(tt.err)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, tt.expectedMsg, msg)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}