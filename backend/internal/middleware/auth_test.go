package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService for testing
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateJWT(token string) (*services.JWTClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*services.JWTClaims), args.Error(1)
}

// MockUserService for testing
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) GetUser(c *gin.Context, userID string) (*models.User, error) {
	args := m.Called(c, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// Helper to create test router with middleware
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

// Helper to create valid JWT claims
func createTestClaims(userID, email string, role models.UserRole) *services.JWTClaims {
	return &services.JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
	}
}

// TestRequireAuth_ValidToken tests middleware with valid token
func TestRequireAuth_ValidToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	// Setup expectations
	claims := createTestClaims("user-123", "test@example.com", models.UserRoleUser)
	mockAuthService.On("ValidateJWT", "valid-token").Return(claims, nil)
	
	// Setup route with middleware
	router.GET("/protected", RequireAuth(mockAuthService), func(c *gin.Context) {
		// Verify context values are set
		userID, _ := c.Get("userID")
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		
		assert.Equal(t, "user-123", userID)
		assert.Equal(t, "test@example.com", email)
		assert.Equal(t, models.UserRoleUser, role)
		
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})
	
	// Make request
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestRequireAuth_MissingToken tests middleware without token
func TestRequireAuth_MissingToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	router.GET("/protected", RequireAuth(mockAuthService), func(c *gin.Context) {
		t.Fatal("Handler should not be called")
	})
	
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_TOKEN")
}

// TestRequireAuth_InvalidToken tests middleware with invalid token
func TestRequireAuth_InvalidToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	// Setup expectations
	mockAuthService.On("ValidateJWT", "invalid-token").Return(nil, assert.AnError)
	
	router.GET("/protected", RequireAuth(mockAuthService), func(c *gin.Context) {
		t.Fatal("Handler should not be called")
	})
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TOKEN")
	mockAuthService.AssertExpectations(t)
}

// TestRequireAuth_ExpiredToken tests middleware with expired token
func TestRequireAuth_ExpiredToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	// Setup expectations - return error with "expired" in message
	mockAuthService.On("ValidateJWT", "expired-token").Return(nil, assert.AnError)
	
	router.GET("/protected", RequireAuth(mockAuthService), func(c *gin.Context) {
		t.Fatal("Handler should not be called")
	})
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestRequireAuth_MalformedHeader tests middleware with malformed auth header
func TestRequireAuth_MalformedHeader(t *testing.T) {
	testCases := []struct {
		name   string
		header string
	}{
		{"no Bearer prefix", "just-a-token"},
		{"wrong prefix", "Basic token"},
		{"only Bearer", "Bearer"},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockAuthService := &MockAuthService{}
			router := setupTestRouter()
			
			router.GET("/protected", RequireAuth(mockAuthService), func(c *gin.Context) {
				t.Fatal("Handler should not be called")
			})
			
			req := httptest.NewRequest("GET", "/protected", nil)
			req.Header.Set("Authorization", tc.header)
			w := httptest.NewRecorder()
			
			router.ServeHTTP(w, req)
			
			assert.Equal(t, http.StatusUnauthorized, w.Code)
		})
	}
}

// TestOptionalAuth_WithToken tests optional auth with valid token
func TestOptionalAuth_WithToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("user-123", "test@example.com", models.UserRoleUser)
	mockAuthService.On("ValidateJWT", "valid-token").Return(claims, nil)
	
	router.GET("/optional", OptionalAuth(mockAuthService), func(c *gin.Context) {
		userID, exists := c.Get("userID")
		assert.True(t, exists)
		assert.Equal(t, "user-123", userID)
		c.JSON(http.StatusOK, gin.H{"authenticated": true})
	})
	
	req := httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestOptionalAuth_WithoutToken tests optional auth without token
func TestOptionalAuth_WithoutToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	router.GET("/optional", OptionalAuth(mockAuthService), func(c *gin.Context) {
		_, exists := c.Get("userID")
		assert.False(t, exists, "userID should not be set without token")
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
	})
	
	req := httptest.NewRequest("GET", "/optional", nil)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"authenticated":false`)
}

// TestOptionalAuth_InvalidToken tests optional auth with invalid token (should continue)
func TestOptionalAuth_InvalidToken(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	mockAuthService.On("ValidateJWT", "invalid-token").Return(nil, assert.AnError)
	
	router.GET("/optional", OptionalAuth(mockAuthService), func(c *gin.Context) {
		_, exists := c.Get("userID")
		assert.False(t, exists, "userID should not be set with invalid token")
		c.JSON(http.StatusOK, gin.H{"authenticated": false})
	})
	
	req := httptest.NewRequest("GET", "/optional", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestRequireFullAccount_FullAccount tests middleware with full account
func TestRequireFullAccount_FullAccount(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	router := setupTestRouter()
	
	claims := createTestClaims("user-123", "test@example.com", models.UserRoleUser)
	user := &models.User{
		ID:          "user-123",
		AccountType: models.AccountTypeFull,
	}
	
	mockAuthService.On("ValidateJWT", "valid-token").Return(claims, nil)
	mockUserService.On("GetUser", mock.Anything, "user-123").Return(user, nil)
	
	handlerCalled := false
	// Proper middleware chaining: RequireAuth → RequireFullAccount → handler
	router.GET("/full-only", 
		RequireAuth(mockAuthService),
		RequireFullAccount(mockUserService),
		func(c *gin.Context) {
			handlerCalled = true
			// Verify auth context is set
			userID, _ := c.Get("userID")
			assert.Equal(t, "user-123", userID)
			// Verify account type is set
			accountType, exists := c.Get("accountType")
			assert.True(t, exists, "accountType should be set")
			assert.Equal(t, models.AccountTypeFull, accountType)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	
	req := httptest.NewRequest("GET", "/full-only", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerCalled, "Handler should have been called")
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// TestRequireFullAccount_GuestAccount tests middleware with guest account
func TestRequireFullAccount_GuestAccount(t *testing.T) {
	mockAuthService := &MockAuthService{}
	mockUserService := &MockUserService{}
	router := setupTestRouter()
	
	claims := createTestClaims("user-123", "test@example.com", models.UserRoleUser)
	user := &models.User{
		ID:          "user-123",
		AccountType: models.AccountTypeGuest,
	}
	
	mockAuthService.On("ValidateJWT", "valid-token").Return(claims, nil)
	mockUserService.On("GetUser", mock.Anything, "user-123").Return(user, nil)
	
	handlerCalled := false
	// Proper middleware chaining
	router.GET("/full-only",
		RequireAuth(mockAuthService),
		RequireFullAccount(mockUserService),
		func(c *gin.Context) {
			handlerCalled = true
			t.Fatal("Handler should not be called for guest account")
		})
	
	req := httptest.NewRequest("GET", "/full-only", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "FULL_ACCOUNT_REQUIRED")
	assert.False(t, handlerCalled, "Handler should not have been called")
	mockAuthService.AssertExpectations(t)
	mockUserService.AssertExpectations(t)
}

// TestRequireAdmin_AdminUser tests middleware with admin user
func TestRequireAdmin_AdminUser(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("admin-123", "admin@example.com", models.UserRoleAdmin)
	mockAuthService.On("ValidateJWT", "admin-token").Return(claims, nil)
	
	// Proper middleware chaining
	router.GET("/admin-only",
		RequireAuth(mockAuthService),
		RequireAdmin(),
		func(c *gin.Context) {
			role, _ := c.Get("role")
			assert.Equal(t, models.UserRoleAdmin, role)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestRequireAdmin_SuperAdminUser tests middleware with superadmin user
func TestRequireAdmin_SuperAdminUser(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("superadmin-123", "superadmin@example.com", models.UserRoleSuperAdmin)
	mockAuthService.On("ValidateJWT", "superadmin-token").Return(claims, nil)
	
	// Proper middleware chaining
	router.GET("/admin-only",
		RequireAuth(mockAuthService),
		RequireAdmin(),
		func(c *gin.Context) {
			role, _ := c.Get("role")
			assert.Equal(t, models.UserRoleSuperAdmin, role)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer superadmin-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestRequireAdmin_RegularUser tests middleware with regular user
func TestRequireAdmin_RegularUser(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("user-123", "user@example.com", models.UserRoleUser)
	mockAuthService.On("ValidateJWT", "user-token").Return(claims, nil)
	
	handlerCalled := false
	// Proper middleware chaining
	router.GET("/admin-only",
		RequireAuth(mockAuthService),
		RequireAdmin(),
		func(c *gin.Context) {
			handlerCalled = true
			t.Fatal("Handler should not be called for regular user")
		})
	
	req := httptest.NewRequest("GET", "/admin-only", nil)
	req.Header.Set("Authorization", "Bearer user-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "ADMIN_REQUIRED")
	assert.False(t, handlerCalled, "Handler should not have been called")
	mockAuthService.AssertExpectations(t)
}

// TestRequireSuperAdmin_SuperAdmin tests middleware with superadmin
func TestRequireSuperAdmin_SuperAdmin(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("superadmin-123", "superadmin@example.com", models.UserRoleSuperAdmin)
	mockAuthService.On("ValidateJWT", "superadmin-token").Return(claims, nil)
	
	// Proper middleware chaining
	router.GET("/superadmin-only",
		RequireAuth(mockAuthService),
		RequireSuperAdmin(),
		func(c *gin.Context) {
			role, _ := c.Get("role")
			assert.Equal(t, models.UserRoleSuperAdmin, role)
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	
	req := httptest.NewRequest("GET", "/superadmin-only", nil)
	req.Header.Set("Authorization", "Bearer superadmin-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestRequireSuperAdmin_AdminUser tests middleware with admin user (should fail)
func TestRequireSuperAdmin_AdminUser(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("admin-123", "admin@example.com", models.UserRoleAdmin)
	mockAuthService.On("ValidateJWT", "admin-token").Return(claims, nil)
	
	handlerCalled := false
	// Proper middleware chaining
	router.GET("/superadmin-only",
		RequireAuth(mockAuthService),
		RequireSuperAdmin(),
		func(c *gin.Context) {
			handlerCalled = true
			t.Fatal("Handler should not be called for admin user")
		})
	
	req := httptest.NewRequest("GET", "/superadmin-only", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "SUPERADMIN_REQUIRED")
	assert.False(t, handlerCalled, "Handler should not have been called")
	mockAuthService.AssertExpectations(t)
}

// TestMiddlewareChaining tests multiple middleware in sequence
func TestMiddlewareChaining(t *testing.T) {
	mockAuthService := &MockAuthService{}
	router := setupTestRouter()
	
	claims := createTestClaims("admin-123", "admin@example.com", models.UserRoleAdmin)
	mockAuthService.On("ValidateJWT", "admin-token").Return(claims, nil)
	
	// Proper middleware chaining: RequireAuth → RequireAdmin → handler
	router.GET("/admin-endpoint",
		RequireAuth(mockAuthService),
		RequireAdmin(),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "success"})
		})
	
	req := httptest.NewRequest("GET", "/admin-endpoint", nil)
	req.Header.Set("Authorization", "Bearer admin-token")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockAuthService.AssertExpectations(t)
}

// TestAuthMiddleware_Integration tests realistic auth flow
func TestAuthMiddleware_Integration(t *testing.T) {
	// Create real auth service for integration test
	authService := services.NewAuthService("test-secret-key-for-integration-test", 1*time.Hour)
	router := setupTestRouter()
	
	// Generate real JWT token
	token, _, err := authService.GenerateJWT("user-123", "test@example.com", models.UserRoleUser)
	require.NoError(t, err)
	
	router.GET("/protected", RequireAuth(authService), func(c *gin.Context) {
		userID, _ := c.Get("userID")
		email, _ := c.Get("email")
		role, _ := c.Get("role")
		
		c.JSON(http.StatusOK, gin.H{
			"userID": userID,
			"email":  email,
			"role":   role,
		})
	})
	
	req := httptest.NewRequest("GET", "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "user-123")
	assert.Contains(t, w.Body.String(), "test@example.com")
}
