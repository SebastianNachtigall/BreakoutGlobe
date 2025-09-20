package repository

import (
	"context"
	"fmt"

	"breakoutglobe/internal/database"
	"breakoutglobe/internal/models"

	"gorm.io/gorm"
)

// POIRepository defines the interface for POI data operations
type POIRepository struct {
	db *database.DB
}

// NewPOIRepository creates a new POI repository instance
func NewPOIRepository(db *database.DB) *POIRepository {
	return &POIRepository{db: db}
}

// Create creates a new POI in the database
func (r *POIRepository) Create(ctx context.Context, poi *models.POI) (*models.POI, error) {
	// Check for existing POIs too close to this location (within 100 meters)
	const minDistanceKm = 0.1 // 100 meters
	
	var existingPOIs []models.POI
	err := r.db.WithContext(ctx).
		Where("map_id = ?", poi.MapID).
		Find(&existingPOIs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check existing POIs: %w", err)
	}
	
	// Check distance to existing POIs
	for _, existing := range existingPOIs {
		if existing.DistanceTo(poi.Position) < minDistanceKm {
			return nil, fmt.Errorf("POI too close to existing POI '%s' (minimum distance: %.0fm)", 
				existing.Name, minDistanceKm*1000)
		}
	}
	
	// Generate ID if not set
	if poi.ID == "" {
		newPOI, err := models.NewPOI(poi.MapID, poi.Name, poi.Description, poi.Position, poi.CreatedBy)
		if err != nil {
			return nil, fmt.Errorf("failed to create POI: %w", err)
		}
		poi = newPOI
	}
	
	// Validate before creating
	if err := poi.Validate(); err != nil {
		return nil, fmt.Errorf("POI validation failed: %w", err)
	}
	
	err = r.db.WithContext(ctx).Create(poi).Error
	if err != nil {
		return nil, fmt.Errorf("failed to create POI: %w", err)
	}
	
	return poi, nil
}

// GetByID retrieves a POI by its ID
func (r *POIRepository) GetByID(ctx context.Context, id string) (*models.POI, error) {
	var poi models.POI
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&poi).Error
	if err != nil {
		return nil, err
	}
	return &poi, nil
}

// GetByMapID retrieves all POIs for a specific map
func (r *POIRepository) GetByMapID(ctx context.Context, mapID string) ([]*models.POI, error) {
	var pois []*models.POI
	err := r.db.WithContext(ctx).
		Where("map_id = ?", mapID).
		Order("created_at DESC").
		Find(&pois).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get POIs for map %s: %w", mapID, err)
	}
	return pois, nil
}

// GetInBounds retrieves all POIs within the specified geographic bounds for a map
func (r *POIRepository) GetInBounds(ctx context.Context, mapID string, bounds models.Bounds) ([]*models.POI, error) {
	var pois []*models.POI
	
	query := r.db.WithContext(ctx).Where("map_id = ?", mapID)
	
	// Handle longitude wrapping around 180/-180
	if bounds.West > bounds.East {
		// Bounds cross the international date line
		query = query.Where(
			"(position_lat BETWEEN ? AND ?) AND (position_lng >= ? OR position_lng <= ?)",
			bounds.South, bounds.North, bounds.West, bounds.East,
		)
	} else {
		// Normal bounds
		query = query.Where(
			"(position_lat BETWEEN ? AND ?) AND (position_lng BETWEEN ? AND ?)",
			bounds.South, bounds.North, bounds.West, bounds.East,
		)
	}
	
	err := query.Order("created_at DESC").Find(&pois).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get POIs in bounds for map %s: %w", mapID, err)
	}
	
	return pois, nil
}

// Update updates an existing POI
func (r *POIRepository) Update(ctx context.Context, poi *models.POI) error {
	// Validate before updating
	if err := poi.Validate(); err != nil {
		return fmt.Errorf("POI validation failed: %w", err)
	}
	
	err := r.db.WithContext(ctx).Save(poi).Error
	if err != nil {
		return fmt.Errorf("failed to update POI: %w", err)
	}
	
	return nil
}

// Delete removes a POI from the database
func (r *POIRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.POI{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete POI: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	
	return nil
}

// GetNearby retrieves POIs within a specified radius from a center point
func (r *POIRepository) GetNearby(ctx context.Context, mapID string, center models.LatLng, radiusKm float64) ([]*models.POI, error) {
	// For simplicity, we'll use a bounding box approximation first, then filter by actual distance
	// 1 degree latitude ≈ 111 km
	latDelta := radiusKm / 111.0
	lngDelta := radiusKm / (111.0 * 0.7) // Rough approximation for longitude at mid-latitudes
	
	bounds := models.Bounds{
		North: center.Lat + latDelta,
		South: center.Lat - latDelta,
		East:  center.Lng + lngDelta,
		West:  center.Lng - lngDelta,
	}
	
	// Get POIs in the bounding box
	pois, err := r.GetInBounds(ctx, mapID, bounds)
	if err != nil {
		return nil, err
	}
	
	// Filter by actual distance
	var nearbyPOIs []*models.POI
	for _, poi := range pois {
		if poi.IsWithinRadius(center, radiusKm) {
			nearbyPOIs = append(nearbyPOIs, poi)
		}
	}
	
	return nearbyPOIs, nil
}