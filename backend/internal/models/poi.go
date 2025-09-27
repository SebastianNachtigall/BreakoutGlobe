package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POI represents a Point of Interest on the map where users can meet
type POI struct {
	ID              string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	MapID           string         `json:"mapId" gorm:"index;type:varchar(36);not null"`
	Map             *Map           `json:"map,omitempty" gorm:"foreignKey:MapID;references:ID"`
	Name            string         `json:"name" gorm:"not null;type:varchar(255)"`
	Description     string         `json:"description" gorm:"type:text"`
	Position        LatLng         `json:"position" gorm:"embedded;embeddedPrefix:position_"`
	CreatedBy       string         `json:"createdBy" gorm:"index;type:varchar(36);not null"`
	Creator         *User          `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	MaxParticipants int            `json:"maxParticipants" gorm:"default:10;not null"`
	ImageURL        string         `json:"imageUrl,omitempty" gorm:"type:varchar(500)"` // Optional POI image
	CreatedAt       time.Time      `json:"createdAt" gorm:"not null"`
	UpdatedAt       time.Time      `json:"updatedAt" gorm:"not null"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete support
}

// NewPOI creates a new POI with validation and generated ID
func NewPOI(mapID, name, description string, position LatLng, createdBy string) (*POI, error) {
	now := time.Now()
	poi := &POI{
		ID:              uuid.New().String(),
		MapID:           mapID,
		Name:            name,
		Description:     description,
		Position:        position,
		CreatedBy:       createdBy,
		MaxParticipants: 10, // Default value
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := poi.Validate(); err != nil {
		return nil, err
	}

	return poi, nil
}

// Validate checks if the POI has all required fields and valid data
func (p POI) Validate() error {
	if p.ID == "" {
		return fmt.Errorf("POI ID is required")
	}

	if p.MapID == "" {
		return fmt.Errorf("map ID is required")
	}

	if p.Name == "" {
		return fmt.Errorf("POI name is required")
	}

	if len(p.Name) > 255 {
		return fmt.Errorf("POI name must be 255 characters or less")
	}

	if err := p.Position.Validate(); err != nil {
		return err
	}

	if p.CreatedBy == "" {
		return fmt.Errorf("created by is required")
	}

	if p.MaxParticipants < 1 || p.MaxParticipants > 50 {
		return fmt.Errorf("max participants must be between 1 and 50")
	}

	if p.CreatedAt.IsZero() {
		return fmt.Errorf("created at is required")
	}

	if len(p.ImageURL) > 500 {
		return fmt.Errorf("image URL must be 500 characters or less")
	}

	return nil
}

// DistanceTo calculates the distance in kilometers from this POI to a given position
func (p POI) DistanceTo(position LatLng) float64 {
	return p.Position.DistanceTo(position)
}

// IsWithinRadius checks if this POI is within the specified radius (in km) from a center point
func (p POI) IsWithinRadius(center LatLng, radiusKm float64) bool {
	return p.DistanceTo(center) <= radiusKm
}

// IsOwnedBy checks if the POI is owned by the specified user
func (p POI) IsOwnedBy(userID string) bool {
	return p.CreatedBy == userID
}

// BelongsToMap checks if the POI belongs to the specified map
func (p POI) BelongsToMap(mapID string) bool {
	return p.MapID == mapID
}

// CanBeModifiedBy checks if a user can modify this POI
func (p POI) CanBeModifiedBy(user *User) bool {
	if user == nil {
		return false
	}
	
	// POI creator can always modify
	if p.CreatedBy == user.ID {
		return true
	}
	
	// Admins and superadmins can modify any POI
	return user.IsAdmin() || user.IsSuperAdmin()
}

// CanBeAccessedBy checks if a user can access this POI
func (p POI) CanBeAccessedBy(user *User) bool {
	// All users can view POIs (read access)
	// Write access is controlled by CanBeModifiedBy
	return user != nil
}

// Update updates the POI's updatedAt timestamp
func (p *POI) Update() {
	p.UpdatedAt = time.Now()
}

// TableName returns the table name for GORM
func (POI) TableName() string {
	return "pois"
}