package services

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/repository"
)

// UserService handles user-related business logic
type UserService struct {
	userRepo repository.UserRepositoryInterface
}

// NewUserService creates a new UserService instance
func NewUserService(userRepo repository.UserRepositoryInterface) *UserService {
	return &UserService{
		userRepo: userRepo,
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

// UploadAvatar uploads an avatar image for a user
func (s *UserService) UploadAvatar(ctx context.Context, userID string, filename string, fileData []byte) (*models.User, error) {
	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Generate unique filename
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s_%d%s", userID, time.Now().Unix(), ext)
	
	// Create avatars directory if it doesn't exist
	avatarDir := "uploads/avatars"
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create avatar directory: %w", err)
	}
	
	// Save file to disk
	filePath := filepath.Join(avatarDir, uniqueFilename)
	if err := os.WriteFile(filePath, fileData, 0644); err != nil {
		return nil, fmt.Errorf("failed to save avatar file: %w", err)
	}
	
	// Update user's avatar URL
	user.AvatarURL = fmt.Sprintf("/api/users/avatar/%s", uniqueFilename)
	user.UpdatedAt = time.Now()
	
	// Save updated user
	if err := s.userRepo.Update(ctx, user); err != nil {
		// Clean up file if database update fails
		os.Remove(filePath)
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}