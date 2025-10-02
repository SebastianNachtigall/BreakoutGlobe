package interfaces

import (
	"context"

	"breakoutglobe/internal/models"
)

// UserRepositoryInterface defines the interface for user data operations
type UserRepositoryInterface interface {
	Create(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	ClearAllUsers(ctx context.Context) error
}