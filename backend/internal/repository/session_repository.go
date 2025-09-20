package repository

import (
	"context"
	"fmt"
	"time"

	"breakoutglobe/internal/database"
	"breakoutglobe/internal/models"
	"gorm.io/gorm"
)

// SessionRepository defines the interface for session data operations
type SessionRepository interface {
	Create(ctx context.Context, session *models.Session) (*models.Session, error)
	GetByID(ctx context.Context, id string) (*models.Session, error)
	GetByUserAndMap(ctx context.Context, userID, mapID string) (*models.Session, error)
	GetActiveByMap(ctx context.Context, mapID string) ([]*models.Session, error)
	Update(ctx context.Context, session *models.Session) (*models.Session, error)
	UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) (*models.Session, error)
	Delete(ctx context.Context, id string) error
	ExpireOldSessions(ctx context.Context, timeout time.Duration) (int64, error)
}

// sessionRepository implements SessionRepository interface
type sessionRepository struct {
	db *database.DB
}

// NewSessionRepository creates a new session repository
func NewSessionRepository(db *database.DB) SessionRepository {
	return &sessionRepository{
		db: db,
	}
}

// Create creates a new session in the database
func (r *sessionRepository) Create(ctx context.Context, session *models.Session) (*models.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	// Check if user already has an active session in this map
	var existingSession models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND map_id = ? AND is_active = ?", session.UserID, session.MapID, true).
		First(&existingSession).Error

	if err == nil {
		return nil, fmt.Errorf("user already has an active session in this map")
	} else if err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing session: %w", err)
	}

	// Validate the session
	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Create the session
	if err := r.db.WithContext(ctx).Create(session).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// GetByID retrieves a session by its ID
func (r *sessionRepository) GetByID(ctx context.Context, id string) (*models.Session, error) {
	if id == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	var session models.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetByUserAndMap retrieves a session by user ID and map ID
func (r *sessionRepository) GetByUserAndMap(ctx context.Context, userID, mapID string) (*models.Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if mapID == "" {
		return nil, fmt.Errorf("map ID cannot be empty")
	}

	var session models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND map_id = ? AND is_active = ?", userID, mapID, true).
		First(&session).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return &session, nil
}

// GetActiveByMap retrieves all active sessions for a specific map
func (r *sessionRepository) GetActiveByMap(ctx context.Context, mapID string) ([]*models.Session, error) {
	if mapID == "" {
		return nil, fmt.Errorf("map ID cannot be empty")
	}

	var sessions []*models.Session
	err := r.db.WithContext(ctx).
		Where("map_id = ? AND is_active = ?", mapID, true).
		Order("last_active DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(ctx context.Context, session *models.Session) (*models.Session, error) {
	if session == nil {
		return nil, fmt.Errorf("session cannot be nil")
	}

	// Validate the session
	if err := session.Validate(); err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	// Update the session
	err := r.db.WithContext(ctx).Save(session).Error
	if err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	return session, nil
}

// UpdateAvatarPosition updates the avatar position for a session
func (r *sessionRepository) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) (*models.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID cannot be empty")
	}

	// Validate position
	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("invalid position: %w", err)
	}

	// Get the session first
	session, err := r.GetByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Update avatar position and activity
	err = session.UpdateAvatarPosition(position)
	if err != nil {
		return nil, fmt.Errorf("failed to update avatar position: %w", err)
	}

	// Save the updated session
	return r.Update(ctx, session)
}

// Delete deletes a session by ID
func (r *sessionRepository) Delete(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("session ID cannot be empty")
	}

	result := r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

// ExpireOldSessions marks sessions as inactive if they haven't been active within the timeout period
func (r *sessionRepository) ExpireOldSessions(ctx context.Context, timeout time.Duration) (int64, error) {
	cutoffTime := time.Now().Add(-timeout)

	result := r.db.WithContext(ctx).
		Model(&models.Session{}).
		Where("last_active < ? AND is_active = ?", cutoffTime, true).
		Update("is_active", false)

	if result.Error != nil {
		return 0, fmt.Errorf("failed to expire old sessions: %w", result.Error)
	}

	return result.RowsAffected, nil
}