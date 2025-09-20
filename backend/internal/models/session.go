package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Session represents a user session with avatar position and activity tracking
type Session struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID     string    `json:"userId" gorm:"index;type:varchar(36);not null"`
	AvatarPos  LatLng    `json:"avatarPosition" gorm:"embedded;embeddedPrefix:avatar_pos_"`
	CreatedAt  time.Time `json:"createdAt" gorm:"not null"`
	LastActive time.Time `json:"lastActive" gorm:"not null"`
	IsActive   bool      `json:"isActive" gorm:"default:true"`
}

// NewSession creates a new session with a generated ID and current timestamp
func NewSession(userID string, initialPos LatLng) (*Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if err := initialPos.Validate(); err != nil {
		return nil, fmt.Errorf("invalid initial position: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		AvatarPos:  initialPos,
		CreatedAt:  now,
		LastActive: now,
		IsActive:   true,
	}

	return session, nil
}

// Validate checks if the session has all required fields and valid data
func (s Session) Validate() error {
	if s.ID == "" {
		return fmt.Errorf("session ID is required")
	}

	if s.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	if err := s.AvatarPos.Validate(); err != nil {
		return err
	}

	if s.CreatedAt.IsZero() {
		return fmt.Errorf("created at is required")
	}

	if s.LastActive.IsZero() {
		return fmt.Errorf("last active is required")
	}

	return nil
}

// IsExpired checks if the session has been inactive for longer than the timeout duration
func (s Session) IsExpired(timeout time.Duration) bool {
	return time.Since(s.LastActive) >= timeout
}

// UpdateActivity updates the last active timestamp and sets the session as active
func (s *Session) UpdateActivity() {
	s.LastActive = time.Now()
	s.IsActive = true
}

// UpdateAvatarPosition updates the avatar position and activity timestamp
func (s *Session) UpdateAvatarPosition(newPos LatLng) error {
	if err := newPos.Validate(); err != nil {
		return err
	}

	s.AvatarPos = newPos
	s.UpdateActivity()
	return nil
}

// Deactivate marks the session as inactive
func (s *Session) Deactivate() {
	s.IsActive = false
}