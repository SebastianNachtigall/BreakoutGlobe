package repository

import (
	"context"
	"fmt"

	"breakoutglobe/internal/database"
	"breakoutglobe/internal/interfaces"
	"breakoutglobe/internal/models"
	"gorm.io/gorm"
)

// userRepository implements UserRepositoryInterface
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) interfaces.UserRepositoryInterface {
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

// ClearAllUsers removes all users from the database - Development helper method
func (r *userRepository) ClearAllUsers(ctx context.Context) error {
	// Use a transaction to handle foreign key constraints
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to begin transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Delete in order to respect foreign key constraints:
	// 1. Sessions (references users)
	if err := tx.Exec("DELETE FROM sessions").Error; err != nil {
		return fmt.Errorf("failed to clear sessions: %w", err)
	}

	// 2. POIs (references users via created_by)
	if err := tx.Exec("DELETE FROM pois").Error; err != nil {
		return fmt.Errorf("failed to clear pois: %w", err)
	}

	// 3. Maps (references users via creator)
	if err := tx.Exec("DELETE FROM maps").Error; err != nil {
		return fmt.Errorf("failed to clear maps: %w", err)
	}

	// 4. Finally delete users
	if err := tx.Exec("DELETE FROM users").Error; err != nil {
		return fmt.Errorf("failed to clear users: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// After clearing all users, recreate system user and default map
	// Import the database package to access the migration functions
	if err := r.recreateSystemDefaults(ctx); err != nil {
		return fmt.Errorf("failed to recreate system defaults: %w", err)
	}

	return nil
}

// recreateSystemDefaults recreates the system user and default map after clearing all users
func (r *userRepository) recreateSystemDefaults(ctx context.Context) error {
	// Create system user
	if err := database.CreateSystemUserIfNotExists(r.db); err != nil {
		return fmt.Errorf("failed to create system user: %w", err)
	}

	// Create default map
	if err := database.CreateDefaultMapIfNotExists(r.db); err != nil {
		return fmt.Errorf("failed to create default map: %w", err)
	}

	return nil
}