package testdata

import (
	"time"

	"breakoutglobe/internal/models"
	"github.com/google/uuid"
)

// POIBuilder provides a fluent API for creating test POI instances
type POIBuilder struct {
	poi *models.POI
}

// NewPOI creates a new POI builder with sensible defaults
func NewPOI() *POIBuilder {
	return &POIBuilder{
		poi: &models.POI{
			ID:              GenerateUUID().String(),
			MapID:           GenerateUUID().String(),
			Name:            "Test POI",
			Description:     "Test Description",
			Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060}, // NYC coordinates
			CreatedBy:       GenerateUUID().String(),
			MaxParticipants: 10,
			CreatedAt:       time.Now(),
		},
	}
}

// WithID sets a custom ID for the POI
func (b *POIBuilder) WithID(id string) *POIBuilder {
	b.poi.ID = id
	return b
}

// WithName sets a custom name for the POI
func (b *POIBuilder) WithName(name string) *POIBuilder {
	b.poi.Name = name
	return b
}

// WithDescription sets a custom description for the POI
func (b *POIBuilder) WithDescription(description string) *POIBuilder {
	b.poi.Description = description
	return b
}

// WithPosition sets a custom position for the POI
func (b *POIBuilder) WithPosition(position models.LatLng) *POIBuilder {
	b.poi.Position = position
	return b
}

// WithCreator sets the creator ID for the POI
func (b *POIBuilder) WithCreator(creatorID uuid.UUID) *POIBuilder {
	b.poi.CreatedBy = creatorID.String()
	return b
}

// WithMap sets the map ID for the POI
func (b *POIBuilder) WithMap(mapID uuid.UUID) *POIBuilder {
	b.poi.MapID = mapID.String()
	return b
}

// WithMaxParticipants sets the maximum participants for the POI
func (b *POIBuilder) WithMaxParticipants(max int) *POIBuilder {
	b.poi.MaxParticipants = max
	return b
}

// WithCreatedAt sets a custom creation time for the POI
func (b *POIBuilder) WithCreatedAt(createdAt time.Time) *POIBuilder {
	b.poi.CreatedAt = createdAt
	return b
}

// Build returns the constructed POI
func (b *POIBuilder) Build() *models.POI {
	return b.poi
}

// SessionBuilder provides a fluent API for creating test Session instances
type SessionBuilder struct {
	session *models.Session
}

// NewSession creates a new Session builder with sensible defaults
func NewSession() *SessionBuilder {
	now := time.Now()
	return &SessionBuilder{
		session: &models.Session{
			ID:         GenerateUUID().String(),
			UserID:     GenerateUUID().String(),
			MapID:      GenerateUUID().String(),
			AvatarPos:  models.LatLng{Lat: 40.7128, Lng: -74.0060}, // NYC coordinates
			CreatedAt:  now,
			LastActive: now,
			IsActive:   true,
		},
	}
}

// WithID sets a custom ID for the session
func (b *SessionBuilder) WithID(id string) *SessionBuilder {
	b.session.ID = id
	return b
}

// WithUser sets the user ID for the session
func (b *SessionBuilder) WithUser(userID uuid.UUID) *SessionBuilder {
	b.session.UserID = userID.String()
	return b
}

// WithMap sets the map ID for the session
func (b *SessionBuilder) WithMap(mapID uuid.UUID) *SessionBuilder {
	b.session.MapID = mapID.String()
	return b
}

// WithPosition sets the avatar position for the session
func (b *SessionBuilder) WithPosition(position models.LatLng) *SessionBuilder {
	b.session.AvatarPos = position
	return b
}

// WithActive sets the active status for the session
func (b *SessionBuilder) WithActive(active bool) *SessionBuilder {
	b.session.IsActive = active
	return b
}

// WithCreatedAt sets a custom creation time for the session
func (b *SessionBuilder) WithCreatedAt(createdAt time.Time) *SessionBuilder {
	b.session.CreatedAt = createdAt
	return b
}

// WithLastActive sets a custom last active time for the session
func (b *SessionBuilder) WithLastActive(lastActive time.Time) *SessionBuilder {
	b.session.LastActive = lastActive
	return b
}

// Build returns the constructed Session
func (b *SessionBuilder) Build() *models.Session {
	return b.session
}

// MapBuilder provides a fluent API for creating test Map instances
type MapBuilder struct {
	mapData *models.Map
}

// NewMap creates a new Map builder with sensible defaults
func NewMap() *MapBuilder {
	now := time.Now()
	return &MapBuilder{
		mapData: &models.Map{
			ID:          GenerateUUID().String(),
			Name:        "Test Map",
			Description: "Test map description",
			CreatedBy:   GenerateUUID().String(),
			IsActive:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// WithID sets a custom ID for the map
func (b *MapBuilder) WithID(id string) *MapBuilder {
	b.mapData.ID = id
	return b
}

// WithName sets a custom name for the map
func (b *MapBuilder) WithName(name string) *MapBuilder {
	b.mapData.Name = name
	return b
}

// WithDescription sets a custom description for the map
func (b *MapBuilder) WithDescription(description string) *MapBuilder {
	b.mapData.Description = description
	return b
}

// WithCreator sets the creator ID for the map
func (b *MapBuilder) WithCreator(creatorID uuid.UUID) *MapBuilder {
	b.mapData.CreatedBy = creatorID.String()
	return b
}

// WithActive sets the active status for the map
func (b *MapBuilder) WithActive(active bool) *MapBuilder {
	b.mapData.IsActive = active
	return b
}

// WithCreatedAt sets a custom creation time for the map
func (b *MapBuilder) WithCreatedAt(createdAt time.Time) *MapBuilder {
	b.mapData.CreatedAt = createdAt
	return b
}

// WithUpdatedAt sets a custom update time for the map
func (b *MapBuilder) WithUpdatedAt(updatedAt time.Time) *MapBuilder {
	b.mapData.UpdatedAt = updatedAt
	return b
}

// Build returns the constructed Map
func (b *MapBuilder) Build() *models.Map {
	return b.mapData
}

// UUID utility functions

// GenerateUUID creates a new UUID
func GenerateUUID() uuid.UUID {
	return uuid.New()
}

// ParseUUID parses a UUID string and panics if invalid (for test convenience)
func ParseUUID(s string) uuid.UUID {
	return uuid.MustParse(s)
}