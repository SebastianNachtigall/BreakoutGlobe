package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// POI represents a Point of Interest on the map where users can meet
type POI struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name            string    `json:"name" gorm:"not null;type:varchar(255)"`
	Description     string    `json:"description" gorm:"type:text"`
	Position        LatLng    `json:"position" gorm:"embedded;embeddedPrefix:position_"`
	CreatedBy       string    `json:"createdBy" gorm:"index;type:varchar(36);not null"`
	MaxParticipants int       `json:"maxParticipants" gorm:"default:10;not null"`
	CreatedAt       time.Time `json:"createdAt" gorm:"not null"`
}

// NewPOI creates a new POI with validation and generated ID
func NewPOI(name, description string, position LatLng, createdBy string) (*POI, error) {
	poi := &POI{
		ID:              uuid.New().String(),
		Name:            name,
		Description:     description,
		Position:        position,
		CreatedBy:       createdBy,
		MaxParticipants: 10, // Default value
		CreatedAt:       time.Now(),
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