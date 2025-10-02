package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUserHandler_ClearAllUsers_Success(t *testing.T) {
	// Setup
	mockUserService := &MockUserService{}
	mockRateLimiter := &MockRateLimiter{}
	
	handler := NewUserHandler(mockUserService, mockRateLimiter)
	
	// Configure mock expectations
	mockUserService.On("ClearAllUsers", mock.Anything).Return(nil)
	
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/users/dev/clear-all", handler.ClearAllUsers)
	
	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/users/dev/clear-all", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	// Verify mock expectations
	mockUserService.AssertExpectations(t)
}

func TestUserHandler_ClearAllUsers_ServiceError(t *testing.T) {
	// Setup
	mockUserService := &MockUserService{}
	mockRateLimiter := &MockRateLimiter{}
	
	handler := NewUserHandler(mockUserService, mockRateLimiter)
	
	// Configure mock expectations - service returns error
	mockUserService.On("ClearAllUsers", mock.Anything).Return(assert.AnError)
	
	// Setup Gin router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.DELETE("/api/users/dev/clear-all", handler.ClearAllUsers)
	
	// Create request
	req := httptest.NewRequest(http.MethodDelete, "/api/users/dev/clear-all", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	// Verify mock expectations
	mockUserService.AssertExpectations(t)
}