package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session represents a user session with avatar position and activity tracking
type Session struct {
	ID         string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	UserID     string         `json:"userId" gorm:"index;type:varchar(36);not null"`
	User       *User          `json:"user,omitempty" gorm:"foreignKey:UserID;references:ID"`
	MapID      string         `json:"mapId" gorm:"index;type:varchar(36);not null"`
	Map        *Map           `json:"map,omitempty" gorm:"foreignKey:MapID;references:ID"`
	AvatarPos  LatLng         `json:"avatarPosition" gorm:"embedded;embeddedPrefix:avatar_pos_"`
	CreatedAt  time.Time      `json:"createdAt" gorm:"not null"`
	LastActive time.Time      `json:"lastActive" gorm:"not null"`
	IsActive   bool           `json:"isActive" gorm:"default:true"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete support
}

// NewSession creates a new session with a generated ID and current timestamp
func NewSession(userID, mapID string, initialPos LatLng) (*Session, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}

	if mapID == "" {
		return nil, fmt.Errorf("map ID is required")
	}

	if err := initialPos.Validate(); err != nil {
		return nil, fmt.Errorf("invalid initial position: %w", err)
	}

	now := time.Now()
	session := &Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		MapID:      mapID,
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

	if s.MapID == "" {
		return fmt.Errorf("map ID is required")
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

// BelongsToUser checks if the session belongs to the specified user
func (s *Session) BelongsToUser(userID string) bool {
	return s.UserID == userID
}

// BelongsToMap checks if the session belongs to the specified map
func (s *Session) BelongsToMap(mapID string) bool {
	return s.MapID == mapID
}

// CanBeAccessedBy checks if a user can access this session
func (s *Session) CanBeAccessedBy(user *User) bool {
	if user == nil {
		return false
	}
	
	// Users can access their own sessions
	if s.UserID == user.ID {
		return true
	}
	
	// Admins and superadmins can access any session
	return user.IsAdmin() || user.IsSuperAdmin()
}

// TableName returns the table name for GORM
func (Session) TableName() string {
	return "sessions"
}