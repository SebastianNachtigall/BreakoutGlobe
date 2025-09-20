package services

import (
	"context"
	"fmt"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SessionRepository defines the interface for session data access
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

// SessionPresence defines the interface for session presence management
type SessionPresence interface {
	SetSessionPresence(ctx context.Context, sessionID string, data redis.SessionPresenceData, ttl time.Duration) error
	GetSessionPresence(ctx context.Context, sessionID string) (*redis.SessionPresenceData, error)
	UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error
	UpdateSessionActivity(ctx context.Context, sessionID string) error
	SessionHeartbeat(ctx context.Context, sessionID string, ttl time.Duration) error
	RemoveSessionPresence(ctx context.Context, sessionID string) error
	GetActiveSessionsForMap(ctx context.Context, mapID string) ([]*redis.SessionPresenceData, error)
	SetCurrentPOI(ctx context.Context, sessionID, poiID string) error
	CleanupExpiredSessions(ctx context.Context, maxAge time.Duration) (int, error)
}

// PubSub defines the interface for publishing real-time events
type PubSub interface {
	PublishAvatarMovement(ctx context.Context, event redis.AvatarMovementEvent) error
	PublishPOICreated(ctx context.Context, event redis.POICreatedEvent) error
	PublishPOIUpdated(ctx context.Context, event redis.POIUpdatedEvent) error
	PublishPOIJoined(ctx context.Context, event redis.POIJoinedEvent) error
	PublishPOILeft(ctx context.Context, event redis.POILeftEvent) error
}

// SessionService handles session management business logic
type SessionService struct {
	repo     SessionRepository
	presence SessionPresence
	pubsub   PubSub
}

// NewSessionService creates a new SessionService instance
func NewSessionService(repo SessionRepository, presence SessionPresence, pubsub PubSub) *SessionService {
	return &SessionService{
		repo:     repo,
		presence: presence,
		pubsub:   pubsub,
	}
}

// CreateSession creates a new user session for a map
func (s *SessionService) CreateSession(ctx context.Context, userID, mapID string, position models.LatLng) (*models.Session, error) {
	// Validate input
	if userID == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	if mapID == "" {
		return nil, fmt.Errorf("map ID is required")
	}
	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("invalid position: %w", err)
	}

	// Check if user already has an active session in this map
	existingSession, err := s.repo.GetByUserAndMap(userID, mapID)
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, fmt.Errorf("failed to check existing session: %w", err)
	}
	if existingSession != nil && existingSession.IsActive {
		return nil, fmt.Errorf("user already has an active session in this map")
	}

	// Create new session
	session := &models.Session{
		ID:         uuid.New().String(),
		UserID:     userID,
		MapID:      mapID,
		AvatarPos:  position,
		IsActive:   true,
		CreatedAt:  time.Now(),
		LastActive: time.Now(),
	}

	// Save to database
	if err := s.repo.Create(session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Set presence in Redis
	presenceData := redis.SessionPresenceData{
		UserID:         userID,
		MapID:          mapID,
		AvatarPosition: position,
		LastActive:     time.Now(),
		CurrentPOI:     nil, // No POI initially
	}

	if err := s.presence.SetSessionPresence(ctx, session.ID, presenceData, 30*time.Minute); err != nil {
		// Log error but don't fail the session creation
		// In a production system, you might want to use a proper logger here
		fmt.Printf("Warning: failed to set session presence: %v\n", err)
	}

	return session, nil
}

// GetSession retrieves a session by ID
func (s *SessionService) GetSession(ctx context.Context, sessionID string) (*models.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	session, err := s.repo.GetByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("session not found")
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

// UpdateAvatarPosition updates the avatar position for a session
func (s *SessionService) UpdateAvatarPosition(ctx context.Context, sessionID string, position models.LatLng) error {
	// Validate input
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}
	if err := position.Validate(); err != nil {
		return fmt.Errorf("invalid position: %w", err)
	}

	// Get session to verify it exists and get user/map info
	session, err := s.repo.GetByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	if !session.IsActive {
		return fmt.Errorf("session is not active")
	}

	// Update position in database
	if err := s.repo.UpdateAvatarPosition(sessionID, position); err != nil {
		return fmt.Errorf("failed to update avatar position: %w", err)
	}

	// Update position in Redis presence
	if err := s.presence.UpdateAvatarPosition(ctx, sessionID, position); err != nil {
		// Log error but don't fail the update
		fmt.Printf("Warning: failed to update avatar position in presence: %v\n", err)
	}

	// Publish avatar movement event
	event := redis.AvatarMovementEvent{
		SessionID: sessionID,
		UserID:    session.UserID,
		MapID:     session.MapID,
		Position:  redis.LatLng{Lat: position.Lat, Lng: position.Lng},
		Timestamp: time.Now(),
	}

	if err := s.pubsub.PublishAvatarMovement(ctx, event); err != nil {
		// Log error but don't fail the update
		fmt.Printf("Warning: failed to publish avatar movement event: %v\n", err)
	}

	return nil
}

// SessionHeartbeat updates the session's last active time
func (s *SessionService) SessionHeartbeat(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	// Get session to verify it exists
	session, err := s.repo.GetByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	if !session.IsActive {
		return fmt.Errorf("session is not active")
	}

	// Update last active time
	session.LastActive = time.Now()
	if err := s.repo.Update(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Update presence heartbeat
	if err := s.presence.SessionHeartbeat(ctx, sessionID, 30*time.Minute); err != nil {
		// Log error but don't fail the heartbeat
		fmt.Printf("Warning: failed to update session heartbeat in presence: %v\n", err)
	}

	return nil
}

// EndSession marks a session as inactive and removes presence
func (s *SessionService) EndSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	// Get session to verify it exists
	session, err := s.repo.GetByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	// Mark session as inactive
	session.IsActive = false
	session.LastActive = time.Now()
	if err := s.repo.Update(session); err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	// Remove presence from Redis
	if err := s.presence.RemoveSessionPresence(ctx, sessionID); err != nil {
		// Log error but don't fail the session end
		fmt.Printf("Warning: failed to remove session presence: %v\n", err)
	}

	return nil
}

// GetActiveSessionsForMap retrieves all active sessions for a map
func (s *SessionService) GetActiveSessionsForMap(ctx context.Context, mapID string) ([]*models.Session, error) {
	if mapID == "" {
		return nil, fmt.Errorf("map ID is required")
	}

	sessions, err := s.repo.GetActiveByMap(mapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions: %w", err)
	}

	return sessions, nil
}

// GetActiveSessionsFromPresence retrieves active sessions from Redis presence
func (s *SessionService) GetActiveSessionsFromPresence(ctx context.Context, mapID string) ([]*redis.SessionPresenceData, error) {
	if mapID == "" {
		return nil, fmt.Errorf("map ID is required")
	}

	sessions, err := s.presence.GetActiveSessionsForMap(ctx, mapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active sessions from presence: %w", err)
	}

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions from both database and presence
func (s *SessionService) CleanupExpiredSessions(ctx context.Context) (int, error) {
	// Cleanup database sessions (mark as inactive)
	if err := s.repo.ExpireOldSessions(30 * time.Minute); err != nil {
		return 0, fmt.Errorf("failed to cleanup database sessions: %w", err)
	}

	// Cleanup presence sessions
	cleanedCount, err := s.presence.CleanupExpiredSessions(ctx, 30*time.Minute)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup presence sessions: %w", err)
	}

	return cleanedCount, nil
}

// SetCurrentPOI sets the current POI for a session
func (s *SessionService) SetCurrentPOI(ctx context.Context, sessionID, poiID string) error {
	if sessionID == "" {
		return fmt.Errorf("session ID is required")
	}

	// Verify session exists and is active
	session, err := s.repo.GetByID(sessionID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	if !session.IsActive {
		return fmt.Errorf("session is not active")
	}

	// Update current POI in presence
	if err := s.presence.SetCurrentPOI(ctx, sessionID, poiID); err != nil {
		return fmt.Errorf("failed to set current POI: %w", err)
	}

	return nil
}

// GetSessionPresence retrieves session presence data from Redis
func (s *SessionService) GetSessionPresence(ctx context.Context, sessionID string) (*redis.SessionPresenceData, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("session ID is required")
	}

	presence, err := s.presence.GetSessionPresence(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session presence: %w", err)
	}

	return presence, nil
}

// ValidateSession checks if a session is valid and active
func (s *SessionService) ValidateSession(ctx context.Context, sessionID string) (*models.Session, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	if !session.IsActive {
		return nil, fmt.Errorf("session is not active")
	}

	// Check if session has expired (30 minutes of inactivity)
	if time.Since(session.LastActive) > 30*time.Minute {
		// Mark session as expired
		if err := s.EndSession(ctx, sessionID); err != nil {
			fmt.Printf("Warning: failed to end expired session: %v\n", err)
		}
		return nil, fmt.Errorf("session has expired")
	}

	return session, nil
}