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

// Test scenario for profile update functionality
type ProfileUpdateTestScenario struct {
	handler         *UserHandler
	router          *gin.Engine
	mockUserService *services.MockUserService
	mockRateLimiter *services.MockRateLimiter
}

func newProfileUpdateScenario(t *testing.T) *ProfileUpdateTestScenario {
	t.Helper()
	
	mockUserService := &services.MockUserService{}
	mockRateLimiter := &services.MockRateLimiter{}
	
	handler := NewUserHandler(mockUserService, mockRateLimiter)
	
	router := gin.New()
	handler.RegisterRoutes(router)
	
	return &ProfileUpdateTestScenario{
		handler:         handler,
		router:          router,
		mockUserService: mockUserService,
		mockRateLimiter: mockRateLimiter,
	}
}

func (s *ProfileUpdateTestScenario) cleanup(t *testing.T) {
	s.mockUserService.AssertExpectations(t)
	s.mockRateLimiter.AssertExpectations(t)
}

// expectProfileUpdateAuthorization sets up expectations for authorized profile updates
func (s *ProfileUpdateTestScenario) expectProfileUpdateAuthorization(userID string, updates *services.UpdateProfileRequest, updatedUser *models.User) *ProfileUpdateTestScenario {
	s.mockRateLimiter.On("CheckRateLimit", mock.Anything, userID, services.ActionUpdateProfile).Return(nil)
	s.mockUserService.On("UpdateProfile", mock.Anything, userID, updates).Return(updatedUser, nil)
	s.mockRateLimiter.On("GetRateLimitHeaders", mock.Anything, userID, services.ActionUpdateProfile).Return(map[string]string{
		"X-RateLimit-Limit":     "60",
		"X-RateLimit-Remaining": "59",
		"X-RateLimit-Reset":     "1640995200",
	})
	return s
}

// expectGuestProfileUpdateRestrictions sets up expectations for guest profile limitations
func (s *ProfileUpdateTestScenario) expectGuestProfileUpdateRestrictions(userID string, fullRequest *services.UpdateProfileRequest, user *models.User) *ProfileUpdateTestScenario {
	s.mockRateLimiter.On("CheckRateLimit", mock.Anything, userID, services.ActionUpdateProfile).Return(nil)
	
	// The handler passes the full request to the service, and the service does the filtering
	// So we expect the service to receive the full request (including displayName)
	s.mockUserService.On("UpdateProfile", mock.Anything, userID, fullRequest).Return(user, nil)
	
	s.mockRateLimiter.On("GetRateLimitHeaders", mock.Anything, userID, services.ActionUpdateProfile).Return(map[string]string{
		"X-RateLimit-Limit":     "60",
		"X-RateLimit-Remaining": "59",
		"X-RateLimit-Reset":     "1640995200",
	})
	return s
}

// UpdateProfile executes a profile update request
func (s *ProfileUpdateTestScenario) UpdateProfile(t *testing.T, userID string, updates map[string]interface{}) *httptest.ResponseRecorder {
	t.Helper()
	
	body, err := json.Marshal(updates)
	assert.NoError(t, err)
	
	req := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userID)
	
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, req)
	
	return recorder
}

func TestProfileUpdateFunctionality(t *testing.T) {
	t.Run("should update guest profile with AboutMe field", func(t *testing.T) {
		// Arrange
		scenario := newProfileUpdateScenario(t)
		defer scenario.cleanup(t)
		
		userID := "guest-user-123"
		
		updates := &services.UpdateProfileRequest{
			AboutMe: stringPtr("Updated about me text"),
		}
		
		updatedUser := &models.User{
			ID:          userID,
			DisplayName: "Guest User",
			AccountType: models.AccountTypeGuest,
			Role:        models.UserRoleUser,
			AboutMe:     stringPtr("Updated about me text"),
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Configure expectations - expectProfileUpdateAuthorization()
		scenario.expectProfileUpdateAuthorization(userID, updates, updatedUser)
		
		// Act
		updateData := map[string]interface{}{
			"aboutMe": "Updated about me text",
		}
		recorder := scenario.UpdateProfile(t, userID, updateData)
		
		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, userID, response["id"])
		assert.Equal(t, "Guest User", response["displayName"])
		assert.Equal(t, "Updated about me text", response["aboutMe"])
		assert.Equal(t, "guest", response["accountType"])
	})

	t.Run("should update full profile with display name and about me", func(t *testing.T) {
		// Arrange
		scenario := newProfileUpdateScenario(t)
		defer scenario.cleanup(t)
		
		userID := "full-user-123"
		
		updates := &services.UpdateProfileRequest{
			DisplayName: stringPtr("Updated Full User"),
			AboutMe:     stringPtr("Updated about me for full user"),
		}
		
		updatedUser := &models.User{
			ID:          userID,
			DisplayName: "Updated Full User",
			AccountType: models.AccountTypeFull,
			Role:        models.UserRoleUser,
			AboutMe:     stringPtr("Updated about me for full user"),
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Configure expectations - expectProfileUpdateAuthorization()
		scenario.expectProfileUpdateAuthorization(userID, updates, updatedUser)
		
		// Act
		updateData := map[string]interface{}{
			"displayName": "Updated Full User",
			"aboutMe":     "Updated about me for full user",
		}
		recorder := scenario.UpdateProfile(t, userID, updateData)
		
		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		assert.Equal(t, userID, response["id"])
		assert.Equal(t, "Updated Full User", response["displayName"])
		assert.Equal(t, "Updated about me for full user", response["aboutMe"])
		assert.Equal(t, "full", response["accountType"])
	})

	t.Run("should restrict guest profile updates to AboutMe only", func(t *testing.T) {
		// Arrange
		scenario := newProfileUpdateScenario(t)
		defer scenario.cleanup(t)
		
		userID := "guest-user-123"
		
		// Guest tries to update display name, but only AboutMe should be updated
		// The handler will pass the full request to the service
		fullRequest := &services.UpdateProfileRequest{
			DisplayName: stringPtr("Attempted Display Name Change"),
			AboutMe:     stringPtr("Updated about me only"),
		}
		
		updatedUser := &models.User{
			ID:          userID,
			DisplayName: "Guest User", // Display name unchanged
			AccountType: models.AccountTypeGuest,
			Role:        models.UserRoleUser,
			AboutMe:     stringPtr("Updated about me only"),
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		
		// Configure expectations - expectGuestProfileUpdateRestrictions()
		scenario.expectGuestProfileUpdateRestrictions(userID, fullRequest, updatedUser)
		
		// Act - try to update both display name and about me
		updateData := map[string]interface{}{
			"displayName": "Attempted Display Name Change",
			"aboutMe":     "Updated about me only",
		}
		recorder := scenario.UpdateProfile(t, userID, updateData)
		
		// Assert
		assert.Equal(t, http.StatusOK, recorder.Code)
		
		var response map[string]interface{}
		err := json.Unmarshal(recorder.Body.Bytes(), &response)
		assert.NoError(t, err)
		
		// Display name should remain unchanged for guest profiles
		assert.Equal(t, "Guest User", response["displayName"])
		assert.Equal(t, "Updated about me only", response["aboutMe"])
	})

	t.Run("should return 401 for missing user ID", func(t *testing.T) {
		// Arrange
		scenario := newProfileUpdateScenario(t)
		defer scenario.cleanup(t)
		
		// Act - no X-User-ID header
		updateData := map[string]interface{}{
			"aboutMe": "Some update",
		}
		
		body, err := json.Marshal(updateData)
		assert.NoError(t, err)
		
		req := httptest.NewRequest(http.MethodPut, "/api/users/profile", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		// No X-User-ID header
		
		recorder := httptest.NewRecorder()
		scenario.router.ServeHTTP(recorder, req)
		
		// Assert
		assert.Equal(t, http.StatusUnauthorized, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "User ID required")
	})

	t.Run("should validate AboutMe field length", func(t *testing.T) {
		// Arrange
		scenario := newProfileUpdateScenario(t)
		defer scenario.cleanup(t)
		
		userID := "user-123"
		
		// Set up rate limiting expectations (validation happens after rate limiting)
		scenario.mockRateLimiter.On("CheckRateLimit", mock.Anything, userID, services.ActionUpdateProfile).Return(nil)
		
		// Act - try to update with too long AboutMe
		longAboutMe := make([]byte, 1001) // Assuming 1000 char limit
		for i := range longAboutMe {
			longAboutMe[i] = 'a'
		}
		
		updateData := map[string]interface{}{
			"aboutMe": string(longAboutMe),
		}
		recorder := scenario.UpdateProfile(t, userID, updateData)
		
		// Assert
		assert.Equal(t, http.StatusBadRequest, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "aboutMe too long")
	})

	t.Run("should handle rate limiting", func(t *testing.T) {
		// Arrange
		scenario := newProfileUpdateScenario(t)
		defer scenario.cleanup(t)
		
		userID := "user-123"
		
		// Configure rate limit exceeded
		resetTime := time.Now().Add(time.Hour)
		scenario.mockRateLimiter.On("CheckRateLimit", mock.Anything, userID, services.ActionUpdateProfile).
			Return(&services.RateLimitError{
				UserID:     userID,
				Action:     services.ActionUpdateProfile,
				Limit:      60,
				Window:     time.Minute,
				ResetTime:  resetTime,
				RetryAfter: time.Until(resetTime),
			})
		
		// Act
		updateData := map[string]interface{}{
			"aboutMe": "Some update",
		}
		recorder := scenario.UpdateProfile(t, userID, updateData)
		
		// Assert
		assert.Equal(t, http.StatusTooManyRequests, recorder.Code)
		assert.Contains(t, recorder.Body.String(), "rate limit exceeded")
	})
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}