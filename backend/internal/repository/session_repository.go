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
	Create(session *models.Session) error
	GetByID(id string) (*models.Session, error)
	GetByUserAndMap(userID, mapID string) (*models.Session, error)
	Update(session *models.Session) error
	UpdateAvatarPosition(sessionID string, position models.LatLng) error
	Delete(id string) error
	GetActiveByMap(mapID string) ([]*models.Session, error)
	ExpireOldSessions(timeout time.Duration) error
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
func (r *sessionRepository) Create(session *models.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	ctx := context.Background()

	// Check if user already has an active session in this map
	var existingSession models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND map_id = ? AND is_active = ?", session.UserID, session.MapID, true).
		First(&existingSession).Error

	if err == nil {
		return fmt.Errorf("user already has an active session in this map")
	} else if err != gorm.ErrRecordNotFound {
		return fmt.Errorf("failed to check existing session: %w", err)
	}

	// Generate ID if not set
	if session.ID == "" {
		newSession, err := models.NewSession(session.UserID, session.MapID, session.AvatarPos)
		if err != nil {
			return fmt.Errorf("failed to create session: %w", err)
		}
		session.ID = newSession.ID
		session.CreatedAt = newSession.CreatedAt
		session.LastActive = newSession.LastActive
	}

	// Validate before creating
	if err := session.Validate(); err != nil {
		return fmt.Errorf("session validation failed: %w", err)
	}

	err = r.db.WithContext(ctx).Create(session).Error
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetByID retrieves a session by its ID
func (r *sessionRepository) GetByID(id string) (*models.Session, error) {
	ctx := context.Background()
	var session models.Session
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetByUserAndMap retrieves a session by user ID and map ID
func (r *sessionRepository) GetByUserAndMap(userID, mapID string) (*models.Session, error) {
	ctx := context.Background()
	var session models.Session
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND map_id = ? AND is_active = ?", userID, mapID, true).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// Update updates an existing session
func (r *sessionRepository) Update(session *models.Session) error {
	if session == nil {
		return fmt.Errorf("session cannot be nil")
	}

	ctx := context.Background()

	// Validate before updating
	if err := session.Validate(); err != nil {
		return fmt.Errorf("session validation failed: %w", err)
	}

	err := r.db.WithContext(ctx).Save(session).Error
	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	return nil
}

// UpdateAvatarPosition updates the avatar position for a session
func (r *sessionRepository) UpdateAvatarPosition(sessionID string, position models.LatLng) error {
	ctx := context.Background()

	// Validate position
	if err := position.Validate(); err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("id = ?", sessionID).
		Updates(map[string]interface{}{
			"avatar_pos_lat": position.Lat,
			"avatar_pos_lng": position.Lng,
			"last_active":    time.Now(),
		})

	if result.Error != nil {
		return fmt.Errorf("failed to update avatar position: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// Delete removes a session from the database
func (r *sessionRepository) Delete(id string) error {
	ctx := context.Background()
	result := r.db.WithContext(ctx).Delete(&models.Session{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete session: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// GetActiveByMap retrieves all active sessions for a specific map
func (r *sessionRepository) GetActiveByMap(mapID string) ([]*models.Session, error) {
	ctx := context.Background()
	var sessions []*models.Session

	err := r.db.WithContext(ctx).
		Where("map_id = ? AND is_active = ?", mapID, true).
		Order("last_active DESC").
		Find(&sessions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions for map %s: %w", mapID, err)
	}

	return sessions, nil
}

// ExpireOldSessions marks old sessions as inactive
func (r *sessionRepository) ExpireOldSessions(timeout time.Duration) error {
	ctx := context.Background()
	cutoffTime := time.Now().Add(-timeout)

	result := r.db.WithContext(ctx).Model(&models.Session{}).
		Where("last_active < ? AND is_active = ?", cutoffTime, true).
		Update("is_active", false)

	if result.Error != nil {
		return fmt.Errorf("failed to expire old sessions: %w", result.Error)
	}

	return nil
}