package services

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"breakoutglobe/internal/interfaces"
	"breakoutglobe/internal/models"
	"breakoutglobe/internal/storage"
)

// UpdateProfileRequest represents a request to update user profile
type UpdateProfileRequest struct {
	DisplayName *string `json:"displayName,omitempty"`
	AboutMe     *string `json:"aboutMe,omitempty"`
}

// UserService handles user-related business logic
type UserService struct {
	userRepo    interfaces.UserRepositoryInterface
	fileStorage storage.FileStorage
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo interfaces.UserRepositoryInterface, fileStorage storage.FileStorage) *UserService {
	return &UserService{
		userRepo:    userRepo,
		fileStorage: fileStorage,
	}
}

// CreateGuestProfile creates a new guest user profile
func (s *UserService) CreateGuestProfile(ctx context.Context, displayName string) (*models.User, error) {
	// Create new guest user
	user, err := models.NewGuestUser(displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create guest user: %w", err)
	}

	// Validate user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Save to repository
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	return user, nil
}

// CreateGuestProfileWithAboutMe creates a new guest user profile with aboutMe field
func (s *UserService) CreateGuestProfileWithAboutMe(ctx context.Context, displayName, aboutMe string) (*models.User, error) {
	fmt.Printf("🏗️ UserService: CreateGuestProfileWithAboutMe called with displayName='%s', aboutMe='%s'\n", displayName, aboutMe)
	
	// Create new guest user
	user, err := models.NewGuestUser(displayName)
	if err != nil {
		return nil, fmt.Errorf("failed to create guest user: %w", err)
	}

	// Always set aboutMe field (even if empty) to ensure consistent behavior
	user.AboutMe = &aboutMe
	fmt.Printf("📝 UserService: Set user.AboutMe to '%v' (pointer to '%s')\n", user.AboutMe, aboutMe)

	// Validate user
	if err := user.Validate(); err != nil {
		return nil, fmt.Errorf("user validation failed: %w", err)
	}

	// Save to repository
	fmt.Printf("💾 UserService: Saving user to repository...\n")
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}

	fmt.Printf("✅ UserService: User saved successfully, returning user with AboutMe='%v'\n", user.AboutMe)
	return user, nil
}

// GetUser retrieves a user by ID
func (s *UserService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}
	
	return user, nil
}

// UploadAvatar uploads an avatar image for a user
func (s *UserService) UploadAvatar(ctx context.Context, userID string, filename string, fileData []byte) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate unique file key
	fileKey := s.fileStorage.GenerateUniqueKey("avatars", userID, filename)
	
	// Determine content type based on file extension
	contentType := getContentTypeFromFilename(filename)
	
	// Upload file to storage
	avatarURL, err := s.fileStorage.UploadFile(ctx, fileKey, fileData, contentType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload avatar: %w", err)
	}
	
	// Delete old avatar if exists
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		// Extract old file key from URL and delete
		if oldKey := extractFileKeyFromURL(*user.AvatarURL); oldKey != "" {
			s.fileStorage.DeleteFile(ctx, oldKey)
		}
	}
	
	// Update user's avatar URL
	user.AvatarURL = &avatarURL
	user.UpdatedAt = time.Now()
	
	// Save updated user
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Clean up uploaded file if database update fails
		s.fileStorage.DeleteFile(ctx, fileKey)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// UpdateProfile updates a user's profile information
func (s *UserService) UpdateProfile(ctx context.Context, userID string, req *UpdateProfileRequest) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	// Apply updates based on account type
	if user.AccountType == models.AccountTypeGuest {
		// Guest profiles can only update AboutMe
		if req.AboutMe != nil {
			// Validate AboutMe length
			if len(*req.AboutMe) > 1000 {
				return nil, fmt.Errorf("aboutMe too long: maximum 1000 characters")
			}
			user.AboutMe = req.AboutMe
		}
		// Ignore DisplayName updates for guest profiles
	} else {
		// Full profiles can update both DisplayName and AboutMe
		if req.DisplayName != nil {
			// Validate DisplayName
			if err := models.ValidateDisplayName(*req.DisplayName); err != nil {
				return nil, fmt.Errorf("invalid display name: %w", err)
			}
			user.DisplayName = *req.DisplayName
		}
		
		if req.AboutMe != nil {
			// Validate AboutMe length
			if len(*req.AboutMe) > 1000 {
				return nil, fmt.Errorf("aboutMe too long: maximum 1000 characters")
			}
			user.AboutMe = req.AboutMe
		}
	}
	
	// Update timestamp
	user.UpdatedAt = time.Now()
	
	// Save updated user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}
	
	return user, nil
}

// getContentTypeFromFilename determines content type from file extension
func getContentTypeFromFilename(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "image/jpeg" // Default fallback
	}
}

// extractFileKeyFromURL extracts the file key from a full URL
func extractFileKeyFromURL(url string) string {
	// Extract the path after /uploads/
	if idx := strings.Index(url, "/uploads/"); idx != -1 {
		return url[idx+9:] // Skip "/uploads/"
	}
	return ""
}

// ClearAllUsers removes all users from the database - Development helper method
func (s *UserService) ClearAllUsers(ctx context.Context) error {
	return s.userRepo.ClearAllUsers(ctx)
}