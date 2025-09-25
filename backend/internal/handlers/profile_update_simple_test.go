package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
)

func TestUpdateProfile_Basic(t *testing.T) {
	// Arrange
	mockUserService := &services.MockUserService{}
	mockRateLimiter := &services.MockRateLimiter{}
	
	handler := NewUserHandler(mockUserService, mockRateLimiter)
	
	router := gin.New()
	handler.RegisterRoutes(router)
	
	userID := "test-user-123"
	
	// Create test user
	updatedUser := &models.User{
		ID:          userID,
		DisplayName: "Test User",
		AccountType: models.AccountTypeGuest,
		Role:        models.UserRoleUser,
		AboutMe:     stringPtr("Updated about me"),
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	
	// Setup mocks
	mockRateLimiter.On("CheckRateLimit", mock.Anything, userID, services.ActionUpdateProfile).Return(nil)
	mockUserService.On("UpdateProfile", mock.Anything, userID, mock.AnythingOfType("*services.UpdateProfileRequest")).Return(updatedUser, nil)
	mockRateLimiter.On("GetRateLimitHeaders", mock.Anything, userID, services.ActionUpdateProfile).Return(map[string]string{
		"X-RateLimit-Limit":     "60",
		"X-RateLimit-Remaining": "59",
		"X-RateLimit-Reset":     "1640995200",
	}, nil)
	
	// Prepare request
	updateData := map[string]interface{}{
		"aboutMe": "Updated about me",
	}
	body, err := json.Marshal(updateData)
	assert.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)
	
	recorder := httptest.NewRecorder()
	
	// Act
	router.ServeHTTP(recorder, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, recorder.Code)
	
	var response map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, userID, response["id"])
	assert.Equal(t, "Test User", response["displayName"])
	assert.Equal(t, "Updated about me", response["aboutMe"])
	
	// Verify mocks
	mockUserService.AssertExpectations(t)
	mockRateLimiter.AssertExpectations(t)
}

