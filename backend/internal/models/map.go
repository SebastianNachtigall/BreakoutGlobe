package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Map represents a map instance that contains isolated sessions and POIs
type Map struct {
	ID          string         `json:"id" gorm:"primaryKey;type:varchar(36)"`
	Name        string         `json:"name" gorm:"not null;type:varchar(255)"`
	Description string         `json:"description" gorm:"type:text"`
	CreatedBy   string         `json:"createdBy" gorm:"index;type:varchar(36);not null"`
	Creator     *User          `json:"creator,omitempty" gorm:"foreignKey:CreatedBy;references:ID"`
	IsActive    bool           `json:"isActive" gorm:"default:true"`
	CreatedAt   time.Time      `json:"createdAt" gorm:"not null"`
	UpdatedAt   time.Time      `json:"updatedAt" gorm:"not null"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"` // Soft delete support
}

// NewMap creates a new map instance with validation and generated ID
func NewMap(name, description, createdBy string) (*Map, error) {
	now := time.Now()
	mapData := &Map{
		ID:          uuid.New().String(),
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		IsActive:    true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := mapData.Validate(); err != nil {
		return nil, err
	}

	return mapData, nil
}

// Validate checks if the map has all required fields and valid data
func (m Map) Validate() error {
	if m.ID == "" {
		return fmt.Errorf("map ID is required")
	}

	if m.Name == "" {
		return fmt.Errorf("map name is required")
	}

	if len(m.Name) > 255 {
		return fmt.Errorf("map name must be 255 characters or less")
	}

	if m.CreatedBy == "" {
		return fmt.Errorf("created by is required")
	}

	if m.CreatedAt.IsZero() {
		return fmt.Errorf("created at is required")
	}

	return nil
}

// Deactivate marks the map as inactive
func (m *Map) Deactivate() {
	m.IsActive = false
	m.UpdatedAt = time.Now()
}

// Activate marks the map as active
func (m *Map) Activate() {
	m.IsActive = true
	m.UpdatedAt = time.Now()
}

// CanBeAccessedBy checks if a user can access this map
func (m *Map) CanBeAccessedBy(userID string) bool {
	// All active maps can be accessed by any user
	// Future enhancement: implement role-based access control
	return m.IsActive
}

// CanBeModifiedBy checks if a user can modify this map
func (m *Map) CanBeModifiedBy(user *User) bool {
	if user == nil {
		return false
	}
	
	// Map creator can always modify
	if m.CreatedBy == user.ID {
		return true
	}
	
	// Admins and superadmins can modify any map
	return user.IsAdmin() || user.IsSuperAdmin()
}

// IsOwnedBy checks if the map is owned by the specified user
func (m *Map) IsOwnedBy(userID string) bool {
	return m.CreatedBy == userID
}

// TableName returns the table name for GORM
func (Map) TableName() string {
	return "maps"
}