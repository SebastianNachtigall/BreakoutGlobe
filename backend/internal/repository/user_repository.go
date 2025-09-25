package repository

import (
	"context"
	"fmt"

	"breakoutglobe/internal/models"
	"gorm.io/gorm"
)

// UserRepositoryInterface defines the interface for user data operations
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &userRepository{
		db: db,
	}
}

// Create creates a new user in the database
func (r *userRepository) Create(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Validate before creating
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *userRepository) GetByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates an existing user in the database
func (r *userRepository) Update(ctx context.Context, user *models.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	// Validate before updating
	if err := user.Validate(); err != nil {
		return fmt.Errorf("user validation failed: %w", err)
	}

	err := r.db.WithContext(ctx).Save(user).Error
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}