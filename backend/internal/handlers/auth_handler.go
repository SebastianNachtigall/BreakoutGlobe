package handlers

import (
	"context"
	"net/http"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
)

// AuthServiceInterface defines the interface for authentication operations
type AuthServiceInterface interface {
	GenerateJWT(userID, email string, role models.UserRole) (string, time.Time, error)
	ValidateJWT(token string) (*services.JWTClaims, error)
}

// AuthUserServiceInterface defines the interface for user operations needed by auth
type AuthUserServiceInterface interface {
	CreateFullAccount(ctx context.Context, email, password, displayName, aboutMe string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUser(ctx context.Context, userID string) (*models.User, error)
	VerifyPassword(ctx context.Context, userID, password string) error
}

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	authService AuthServiceInterface
	userService AuthUserServiceInterface
	rateLimiter services.RateLimiterInterface
}

// NewAuthHandler creates a new AuthHandler instance
func NewAuthHandler(authService AuthServiceInterface, userService AuthUserServiceInterface, rateLimiter services.RateLimiterInterface) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		rateLimiter: rateLimiter,
	}
}

// Request/Response DTOs

// SignupRequest represents the request body for user signup
type SignupRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName" binding:"required,min=3,max=50"`
	AboutMe     string `json:"aboutMe,omitempty"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response for successful authentication
type AuthResponse struct {
	Token     string       `json:"token"`
	ExpiresAt string       `json:"expiresAt"`
	User      UserResponse `json:"user"`
}

// UserResponse represents user data in auth responses
type UserResponse struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"displayName"`
	AccountType string `json:"accountType"`
	Role        string `json:"role"`
	AvatarURL   string `json:"avatarUrl,omitempty"`
	AboutMe     string `json:"aboutMe,omitempty"`
	CreatedAt   string `json:"createdAt"`
}

// Signup handles POST /api/auth/signup
func (h *AuthHandler) Signup(c *gin.Context) {
	var req SignupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, "signup:"+req.Email, services.ActionCreatePOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}

	// Create full account
	user, err := h.userService.CreateFullAccount(c, req.Email, req.Password, req.DisplayName, req.AboutMe)
	if err != nil {
		// Check for specific error types
		if containsString(err.Error(), "email already in use") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Code:    "EMAIL_IN_USE",
				Message: "Email address is already registered",
			})
			return
		}
		if containsString(err.Error(), "password") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "INVALID_PASSWORD",
				Message: err.Error(),
			})
			return
		}
		if containsString(err.Error(), "display name") || containsString(err.Error(), "validation") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "SIGNUP_FAILED",
			Message: "Failed to create account",
			Details: err.Error(),
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.authService.GenerateJWT(user.ID, *user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate authentication token",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusCreated, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      mapUserToResponse(user),
	})
}

// Login handles POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, "login:"+req.Email, services.ActionCreatePOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}

	// Get user by email
	user, err := h.userService.GetUserByEmail(c, req.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "INVALID_CREDENTIALS",
			Message: "Invalid email or password",
		})
		return
	}

	// Verify password
	if err := h.userService.VerifyPassword(c, user.ID, req.Password); err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "INVALID_CREDENTIALS",
			Message: "Invalid email or password",
		})
		return
	}

	// Generate JWT token
	token, expiresAt, err := h.authService.GenerateJWT(user.ID, *user.Email, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "TOKEN_GENERATION_FAILED",
			Message: "Failed to generate authentication token",
		})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt.Format(time.RFC3339),
		User:      mapUserToResponse(user),
	})
}

// Logout handles POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWT is stateless, so logout is handled client-side by removing the token
	// This endpoint exists for consistency and future enhancements (e.g., token blacklist)
	c.JSON(http.StatusOK, gin.H{
		"message": "Logged out successfully",
	})
}

// GetCurrentUser handles GET /api/auth/me
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "Authentication required",
		})
		return
	}

	// Get user from service
	user, err := h.userService.GetUser(c, userID.(string))
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Code:    "USER_NOT_FOUND",
			Message: "User not found",
		})
		return
	}

	// Return user data
	c.JSON(http.StatusOK, mapUserToResponse(user))
}

// Helper methods

// mapUserToResponse converts a User model to UserResponse
func mapUserToResponse(user *models.User) UserResponse {
	response := UserResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AccountType: string(user.AccountType),
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}

	if user.Email != nil {
		response.Email = *user.Email
	}
	if user.AvatarURL != nil {
		response.AvatarURL = *user.AvatarURL
	}
	if user.AboutMe != nil {
		response.AboutMe = *user.AboutMe
	}

	return response
}

// containsString checks if a string contains a substring (case-insensitive)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || 
		(len(s) > 0 && len(substr) > 0 && findSubstringInString(s, substr)))
}

func findSubstringInString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// handleRateLimitError handles rate limit errors
func (h *AuthHandler) handleRateLimitError(c *gin.Context, err error) {
	if rateLimitErr, ok := err.(*services.RateLimitError); ok {
		c.Header("Retry-After", "3600")
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: "Too many requests. Please try again later.",
			Details: rateLimitErr.Error(),
		})
		return
	}

	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Too many requests. Please try again later.",
	})
}
