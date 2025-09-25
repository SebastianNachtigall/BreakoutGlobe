package handlers

import (
	"context"
	"errors"
	"net/http"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/services"

	"github.com/gin-gonic/gin"
)

// UserServiceInterface defines the interface for user service operations
type UserServiceInterface interface {
	CreateGuestProfile(ctx context.Context, displayName string) (*models.User, error)
	UploadAvatar(ctx context.Context, userID string, filename string, fileData []byte) (*models.User, error)
}

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	userService UserServiceInterface
	rateLimiter services.RateLimiterInterface
}

// NewUserHandler creates a new UserHandler instance
func NewUserHandler(userService UserServiceInterface, rateLimiter services.RateLimiterInterface) *UserHandler {
	return &UserHandler{
		userService: userService,
		rateLimiter: rateLimiter,
	}
}

// RegisterRoutes registers user-related routes
func (h *UserHandler) RegisterRoutes(router *gin.Engine) {
	api := router.Group("/api")
	{
		// User profile management
		api.POST("/users/profile", h.CreateProfile)
		api.POST("/users/avatar", h.UploadAvatar)
	}
}

// Request/Response DTOs

// CreateProfileRequest represents the request body for creating a user profile
type CreateProfileRequest struct {
	DisplayName string `json:"displayName" binding:"required"`
	AccountType string `json:"accountType" binding:"required"`
	AboutMe     string `json:"aboutMe,omitempty"`
}

// CreateProfileResponse represents the response for creating a user profile
type CreateProfileResponse struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
	AccountType string `json:"accountType"`
	Role        string `json:"role"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   string `json:"createdAt"`
}

// CreateProfile handles POST /api/users/profile
func (h *UserHandler) CreateProfile(c *gin.Context) {
	var req CreateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}
	
	// Validate request
	if err := h.validateCreateProfileRequest(req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "VALIDATION_ERROR",
			Message: err.Error(),
		})
		return
	}
	
	// Check rate limit (using existing ActionCreatePOI for now)
	if err := h.rateLimiter.CheckRateLimit(c, "anonymous", services.ActionCreatePOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Create guest profile
	user, err := h.userService.CreateGuestProfile(c, req.DisplayName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create profile",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, user.ID, services.ActionCreatePOI)
	
	// Return response
	response := CreateProfileResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AccountType: string(user.AccountType),
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}
	
	c.JSON(http.StatusCreated, response)
}

// UploadAvatar handles POST /api/users/avatar
func (h *UserHandler) UploadAvatar(c *gin.Context) {
	// Get user ID from header (temporary - will be from auth middleware later)
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Code:    "UNAUTHORIZED",
			Message: "User ID required",
		})
		return
	}
	
	// Check rate limit
	if err := h.rateLimiter.CheckRateLimit(c, userID, services.ActionCreatePOI); err != nil {
		h.handleRateLimitError(c, err)
		return
	}
	
	// Parse multipart form
	err := c.Request.ParseMultipartForm(2 << 20) // 2MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_REQUEST",
			Message: "Failed to parse multipart form",
			Details: err.Error(),
		})
		return
	}
	
	// Get file from form
	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "MISSING_FILE",
			Message: "Avatar file is required",
			Details: err.Error(),
		})
		return
	}
	defer file.Close()
	
	// Validate file size (max 2MB)
	if header.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "FILE_TOO_LARGE",
			Message: "File size must be less than 2MB",
		})
		return
	}
	
	// Validate file type by checking file extension and content type
	contentType := header.Header.Get("Content-Type")
	validTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
	}
	
	if !validTypes[contentType] {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Code:    "INVALID_FILE_TYPE",
			Message: "Only JPEG and PNG files are allowed",
		})
		return
	}
	
	// Read file data
	fileData := make([]byte, header.Size)
	_, err = file.Read(fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "FILE_READ_ERROR",
			Message: "Failed to read file data",
			Details: err.Error(),
		})
		return
	}
	
	// Upload avatar via service
	user, err := h.userService.UploadAvatar(c, userID, header.Filename, fileData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Code:    "UPLOAD_FAILED",
			Message: "Failed to upload avatar",
			Details: err.Error(),
		})
		return
	}
	
	// Add rate limit headers
	h.addRateLimitHeaders(c, userID, services.ActionCreatePOI)
	
	// Return updated user profile
	response := CreateProfileResponse{
		ID:          user.ID,
		DisplayName: user.DisplayName,
		AccountType: string(user.AccountType),
		Role:        string(user.Role),
		IsActive:    user.IsActive,
		CreatedAt:   user.CreatedAt.Format(time.RFC3339),
	}
	
	c.JSON(http.StatusOK, response)
}

// Helper methods

// validateCreateProfileRequest validates the create profile request
func (h *UserHandler) validateCreateProfileRequest(req CreateProfileRequest) error {
	if req.DisplayName == "" {
		return errors.New("display name is required")
	}
	if len(req.DisplayName) < 3 {
		return errors.New("display name must be at least 3 characters")
	}
	if len(req.DisplayName) > 50 {
		return errors.New("display name must be less than 50 characters")
	}
	if req.AccountType != "guest" {
		return errors.New("only guest account type is supported")
	}
	return nil
}

// handleRateLimitError handles rate limit errors
func (h *UserHandler) handleRateLimitError(c *gin.Context, err error) {
	if _, ok := err.(*services.RateLimitError); ok {
		// Set Retry-After header
		c.Header("Retry-After", "3600") // Use fixed value for simplicity
		
		c.JSON(http.StatusTooManyRequests, ErrorResponse{
			Code:    "RATE_LIMIT_EXCEEDED",
			Message: "Rate limit exceeded",
			Details: err.Error(),
		})
		return
	}
	
	// Generic rate limit error
	c.JSON(http.StatusTooManyRequests, ErrorResponse{
		Code:    "RATE_LIMIT_EXCEEDED",
		Message: "Rate limit exceeded",
		Details: err.Error(),
	})
}

// addRateLimitHeaders adds rate limit headers to the response
func (h *UserHandler) addRateLimitHeaders(c *gin.Context, userID string, action services.ActionType) {
	headers, err := h.rateLimiter.GetRateLimitHeaders(c, userID, action)
	if err != nil {
		// Log error but don't fail the request
		return
	}
	
	for key, value := range headers {
		c.Header(key, value)
	}
}