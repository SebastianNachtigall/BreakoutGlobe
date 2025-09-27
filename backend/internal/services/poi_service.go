package services

import (
	"context"
	"fmt"
	"mime/multipart"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/redis"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// POIServiceInterface defines the interface for POI management operations
type POIServiceInterface interface {
	// CreatePOI creates a new POI with duplicate location checking
	CreatePOI(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int) (*models.POI, error)
	
	// CreatePOIWithImage creates a new POI with optional image upload
	CreatePOIWithImage(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int, imageFile *multipart.FileHeader) (*models.POI, error)
	
	// GetPOI retrieves a POI by ID
	GetPOI(ctx context.Context, poiID string) (*models.POI, error)
	
	// GetPOIsForMap retrieves all POIs for a specific map
	GetPOIsForMap(ctx context.Context, mapID string) ([]*models.POI, error)
	
	// GetPOIsInBounds retrieves POIs within specified geographic bounds
	GetPOIsInBounds(ctx context.Context, mapID string, bounds POIBounds) ([]*models.POI, error)
	
	// UpdatePOI updates POI information
	UpdatePOI(ctx context.Context, poiID string, updateData POIUpdateData) (*models.POI, error)
	
	// DeletePOI deletes a POI and removes all participants
	DeletePOI(ctx context.Context, poiID string) error
	
	// JoinPOI adds a user to a POI with capacity checking
	JoinPOI(ctx context.Context, poiID, userID string) error
	
	// LeavePOI removes a user from a POI
	LeavePOI(ctx context.Context, poiID, userID string) error
	
	// GetPOIParticipants retrieves all participants of a POI
	GetPOIParticipants(ctx context.Context, poiID string) ([]string, error)
	
	// GetPOIParticipantCount retrieves the current participant count
	GetPOIParticipantCount(ctx context.Context, poiID string) (int, error)
	
	// GetPOIParticipantsWithInfo retrieves all participants of a POI with their display names and avatars
	GetPOIParticipantsWithInfo(ctx context.Context, poiID string) ([]POIParticipantInfo, error)
	
	// GetUserPOIs retrieves all POIs a user is participating in
	GetUserPOIs(ctx context.Context, userID string) ([]string, error)
	
	// ValidatePOI validates that a POI exists
	ValidatePOI(ctx context.Context, poiID string) (*models.POI, error)
}

// POIRepositoryInterface defines the interface for POI repository operations
type POIRepositoryInterface interface {
	Create(ctx context.Context, poi *models.POI) error
	GetByID(ctx context.Context, id string) (*models.POI, error)
	GetByMapID(ctx context.Context, mapID string) ([]*models.POI, error)
	GetInBounds(ctx context.Context, mapID string, minLat, maxLat, minLng, maxLng float64) ([]*models.POI, error)
	Update(ctx context.Context, poi *models.POI) error
	Delete(ctx context.Context, id string) error
	CheckDuplicateLocation(ctx context.Context, mapID string, lat, lng float64, excludeID string) ([]*models.POI, error)
}

// POIParticipantsInterface defines the interface for POI participant operations
type POIParticipantsInterface interface {
	JoinPOI(ctx context.Context, poiID, userID string) error
	LeavePOI(ctx context.Context, poiID, userID string) error
	GetParticipants(ctx context.Context, poiID string) ([]string, error)
	GetParticipantCount(ctx context.Context, poiID string) (int, error)
	IsParticipant(ctx context.Context, poiID, userID string) (bool, error)
	CanJoinPOI(ctx context.Context, poiID string, maxParticipants int) (bool, error)
	RemoveAllParticipants(ctx context.Context, poiID string) error
	RemoveParticipantFromAllPOIs(ctx context.Context, userID string) error
	GetPOIsForParticipant(ctx context.Context, userID string) ([]string, error)
}

// ImageUploaderInterface defines the interface for image upload operations
type ImageUploaderInterface interface {
	UploadPOIImage(ctx context.Context, imageFile *multipart.FileHeader) (string, error)
}

// UserServiceInterface defines the interface for user operations needed by POI service
type UserServiceInterface interface {
	GetUser(ctx context.Context, userID string) (*models.User, error)
}



// POIService implements POI management operations
type POIService struct {
	poiRepo       POIRepositoryInterface
	participants  POIParticipantsInterface
	pubsub        PubSub
	imageUploader ImageUploaderInterface
	userService   UserServiceInterface
}

// POIBounds represents geographic bounds for POI queries
type POIBounds struct {
	MinLat float64 `json:"minLat"`
	MaxLat float64 `json:"maxLat"`
	MinLng float64 `json:"minLng"`
	MaxLng float64 `json:"maxLng"`
}

// POIUpdateData represents data for updating a POI
type POIUpdateData struct {
	Name            string `json:"name,omitempty"`
	Description     string `json:"description,omitempty"`
	MaxParticipants int    `json:"maxParticipants,omitempty"`
}

// POIParticipantInfo represents a POI participant with display information
type POIParticipantInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarUrl"`
}

// Default configuration values
const (
	MaxPOINameLength        = 100
	MaxPOIDescriptionLength = 500
)

// NewPOIService creates a new POIService instance
func NewPOIService(poiRepo POIRepositoryInterface, participants POIParticipantsInterface, pubsub PubSub, userService UserServiceInterface) *POIService {
	return &POIService{
		poiRepo:       poiRepo,
		participants:  participants,
		pubsub:        pubsub,
		imageUploader: nil, // No image uploader by default
		userService:   userService,
	}
}

// NewPOIServiceWithImageUploader creates a new POIService instance with image upload support
func NewPOIServiceWithImageUploader(poiRepo POIRepositoryInterface, participants POIParticipantsInterface, pubsub PubSub, imageUploader ImageUploaderInterface, userService UserServiceInterface) *POIService {
	return &POIService{
		poiRepo:       poiRepo,
		participants:  participants,
		pubsub:        pubsub,
		imageUploader: imageUploader,
		userService:   userService,
	}
}

// CreatePOI creates a new POI with duplicate location checking
func (s *POIService) CreatePOI(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int) (*models.POI, error) {
	// Validate input
	if err := s.validatePOIInput(mapID, name, createdBy, maxParticipants); err != nil {
		return nil, err
	}

	// Validate position
	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("invalid position: %w", err)
	}

	// Check for duplicate location
	duplicates, err := s.poiRepo.CheckDuplicateLocation(ctx, mapID, position.Lat, position.Lng, "")
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate location: %w", err)
	}
	if len(duplicates) > 0 {
		return nil, fmt.Errorf("POI already exists at this location (lat: %f, lng: %f)", position.Lat, position.Lng)
	}

	// Create new POI
	poi := &models.POI{
		ID:              uuid.New().String(),
		MapID:           mapID,
		Name:            name,
		Description:     description,
		Position:        position,
		CreatedBy:       createdBy,
		MaxParticipants: maxParticipants,
		CreatedAt:       time.Now(),
	}

	// Validate the POI
	if err := poi.Validate(); err != nil {
		return nil, fmt.Errorf("invalid POI data: %w", err)
	}

	// Save to database
	if err := s.poiRepo.Create(ctx, poi); err != nil {
		return nil, fmt.Errorf("failed to create POI in database: %w", err)
	}

	// Publish POI created event
	createdEvent := redis.POICreatedEvent{
		POIID:           poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		Position:        redis.LatLng{Lat: position.Lat, Lng: position.Lng},
		CreatedBy:       poi.CreatedBy,
		MaxParticipants: poi.MaxParticipants,
		Timestamp:       time.Now(),
	}

	if err := s.pubsub.PublishPOICreated(ctx, createdEvent); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to publish POI created event: %v\n", err)
	}

	return poi, nil
}

// CreatePOIWithImage creates a new POI with optional image upload
func (s *POIService) CreatePOIWithImage(ctx context.Context, mapID, name, description string, position models.LatLng, createdBy string, maxParticipants int, imageFile *multipart.FileHeader) (*models.POI, error) {
	// Validate input
	if err := s.validatePOIInput(mapID, name, createdBy, maxParticipants); err != nil {
		return nil, err
	}

	// Validate position
	if err := position.Validate(); err != nil {
		return nil, fmt.Errorf("invalid position: %w", err)
	}

	// Check for duplicate location
	duplicates, err := s.poiRepo.CheckDuplicateLocation(ctx, mapID, position.Lat, position.Lng, "")
	if err != nil {
		return nil, fmt.Errorf("failed to check duplicate location: %w", err)
	}
	if len(duplicates) > 0 {
		return nil, fmt.Errorf("POI already exists at this location (lat: %f, lng: %f)", position.Lat, position.Lng)
	}

	// Upload image if provided
	var imageURL string
	if imageFile != nil && s.imageUploader != nil {
		imageURL, err = s.imageUploader.UploadPOIImage(ctx, imageFile)
		if err != nil {
			return nil, fmt.Errorf("failed to upload POI image: %w", err)
		}
	}

	// Create new POI
	poi := &models.POI{
		ID:              uuid.New().String(),
		MapID:           mapID,
		Name:            name,
		Description:     description,
		Position:        position,
		CreatedBy:       createdBy,
		MaxParticipants: maxParticipants,
		ImageURL:        imageURL,
		CreatedAt:       time.Now(),
	}

	// Validate the POI
	if err := poi.Validate(); err != nil {
		return nil, fmt.Errorf("invalid POI data: %w", err)
	}

	// Save to database
	if err := s.poiRepo.Create(ctx, poi); err != nil {
		return nil, fmt.Errorf("failed to create POI in database: %w", err)
	}

	// Publish POI created event
	createdEvent := redis.POICreatedEvent{
		POIID:           poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		Position:        redis.LatLng{Lat: position.Lat, Lng: position.Lng},
		CreatedBy:       poi.CreatedBy,
		MaxParticipants: poi.MaxParticipants,
		Timestamp:       time.Now(),
	}

	if err := s.pubsub.PublishPOICreated(ctx, createdEvent); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to publish POI created event: %v\n", err)
	}

	return poi, nil
}

// GetPOI retrieves a POI by ID
func (s *POIService) GetPOI(ctx context.Context, poiID string) (*models.POI, error) {
	poi, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("POI not found: %s", poiID)
		}
		return nil, fmt.Errorf("failed to get POI: %w", err)
	}
	return poi, nil
}

// GetPOIsForMap retrieves all POIs for a specific map
func (s *POIService) GetPOIsForMap(ctx context.Context, mapID string) ([]*models.POI, error) {
	pois, err := s.poiRepo.GetByMapID(ctx, mapID)
	if err != nil {
		return nil, fmt.Errorf("failed to get POIs for map: %w", err)
	}
	return pois, nil
}

// GetPOIsInBounds retrieves POIs within specified geographic bounds
func (s *POIService) GetPOIsInBounds(ctx context.Context, mapID string, bounds POIBounds) ([]*models.POI, error) {
	// Validate bounds
	if err := s.validateBounds(bounds); err != nil {
		return nil, err
	}

	pois, err := s.poiRepo.GetInBounds(ctx, mapID, bounds.MinLat, bounds.MaxLat, bounds.MinLng, bounds.MaxLng)
	if err != nil {
		return nil, fmt.Errorf("failed to get POIs in bounds: %w", err)
	}
	return pois, nil
}

// UpdatePOI updates POI information
func (s *POIService) UpdatePOI(ctx context.Context, poiID string, updateData POIUpdateData) (*models.POI, error) {
	// Get existing POI
	poi, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("POI not found: %s", poiID)
		}
		return nil, fmt.Errorf("failed to get POI: %w", err)
	}

	// Update fields if provided
	updated := false
	if updateData.Name != "" && updateData.Name != poi.Name {
		if len(updateData.Name) > MaxPOINameLength {
			return nil, fmt.Errorf("POI name too long (max %d characters)", MaxPOINameLength)
		}
		poi.Name = updateData.Name
		updated = true
	}

	if updateData.Description != poi.Description {
		if len(updateData.Description) > MaxPOIDescriptionLength {
			return nil, fmt.Errorf("POI description too long (max %d characters)", MaxPOIDescriptionLength)
		}
		poi.Description = updateData.Description
		updated = true
	}

	if updateData.MaxParticipants > 0 && updateData.MaxParticipants != poi.MaxParticipants {
		if updateData.MaxParticipants < 1 {
			return nil, fmt.Errorf("max participants must be at least 1")
		}
		poi.MaxParticipants = updateData.MaxParticipants
		updated = true
	}

	if !updated {
		return poi, nil // No changes needed
	}

	// Validate updated POI
	if err := poi.Validate(); err != nil {
		return nil, fmt.Errorf("invalid updated POI data: %w", err)
	}

	// Save to database
	if err := s.poiRepo.Update(ctx, poi); err != nil {
		return nil, fmt.Errorf("failed to update POI in database: %w", err)
	}

	// Publish POI updated event
	updatedEvent := redis.POIUpdatedEvent{
		POIID:           poi.ID,
		MapID:           poi.MapID,
		Name:            poi.Name,
		Description:     poi.Description,
		MaxParticipants: poi.MaxParticipants,
		Timestamp:       time.Now(),
	}

	if err := s.pubsub.PublishPOIUpdated(ctx, updatedEvent); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to publish POI updated event: %v\n", err)
	}

	return poi, nil
}

// DeletePOI deletes a POI and removes all participants
func (s *POIService) DeletePOI(ctx context.Context, poiID string) error {
	// Verify POI exists
	_, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("POI not found: %s", poiID)
		}
		return fmt.Errorf("failed to get POI: %w", err)
	}

	// Remove all participants first
	if err := s.participants.RemoveAllParticipants(ctx, poiID); err != nil {
		return fmt.Errorf("failed to remove POI participants: %w", err)
	}

	// Delete POI from database
	if err := s.poiRepo.Delete(ctx, poiID); err != nil {
		return fmt.Errorf("failed to delete POI from database: %w", err)
	}

	return nil
}

// JoinPOI adds a user to a POI with capacity checking
func (s *POIService) JoinPOI(ctx context.Context, poiID, userID string) error {
	// Get POI to verify it exists and get capacity info
	poi, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("POI not found: %s", poiID)
		}
		return fmt.Errorf("failed to get POI: %w", err)
	}

	// Check if user is already a participant
	isParticipant, err := s.participants.IsParticipant(ctx, poiID, userID)
	if err != nil {
		return fmt.Errorf("failed to check participant status: %w", err)
	}
	if isParticipant {
		return fmt.Errorf("user is already a participant in POI %s", poiID)
	}

	// Check if POI has capacity
	canJoin, err := s.participants.CanJoinPOI(ctx, poiID, poi.MaxParticipants)
	if err != nil {
		return fmt.Errorf("failed to check POI capacity: %w", err)
	}
	if !canJoin {
		return fmt.Errorf("POI is at maximum capacity (%d participants)", poi.MaxParticipants)
	}

	// Add user to POI
	if err := s.participants.JoinPOI(ctx, poiID, userID); err != nil {
		return fmt.Errorf("failed to join POI: %w", err)
	}
	
	// Update discussion timer based on new participant count
	if err := s.updateDiscussionTimer(ctx, poiID); err != nil {
		// Log error but don't fail the join operation
		// Discussion timer is not critical for POI functionality
	}

	// Get updated participant count
	currentCount, err := s.participants.GetParticipantCount(ctx, poiID)
	if err != nil {
		currentCount = 0 // Fallback to 0 if we can't get the count
	}

	// Get updated participant information for the event
	participantsInfo, err := s.GetPOIParticipantsWithInfo(ctx, poiID)
	if err != nil {
		// Log error but continue with basic event
		fmt.Printf("Warning: failed to get participant info for POI joined event: %v\n", err)
		participantsInfo = []POIParticipantInfo{}
	}

	// Convert to Redis participant format
	var redisParticipants []redis.POIParticipant
	for _, p := range participantsInfo {
		redisParticipants = append(redisParticipants, redis.POIParticipant{
			ID:        p.ID,
			Name:      p.Name,
			AvatarURL: p.AvatarURL,
		})
	}

	// Publish enhanced POI joined event with participant information
	joinedEventWithParticipants := redis.POIJoinedEventWithParticipants{
		POIID:        poiID,
		MapID:        poi.MapID,
		UserID:       userID,
		SessionID:    userID, // For now, use userID as sessionID
		CurrentCount: currentCount,
		Participants: redisParticipants,
		Timestamp:    time.Now(),
	}

	if err := s.pubsub.PublishPOIJoinedWithParticipants(ctx, joinedEventWithParticipants); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to publish POI joined event with participants: %v\n", err)
	}

	return nil
}

// LeavePOI removes a user from a POI
func (s *POIService) LeavePOI(ctx context.Context, poiID, userID string) error {
	// Get POI to verify it exists
	poi, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("POI not found: %s", poiID)
		}
		return fmt.Errorf("failed to get POI: %w", err)
	}

	// Check if user is a participant
	isParticipant, err := s.participants.IsParticipant(ctx, poiID, userID)
	if err != nil {
		return fmt.Errorf("failed to check participant status: %w", err)
	}
	if !isParticipant {
		return fmt.Errorf("user is not a participant in POI %s", poiID)
	}

	// Remove user from POI
	if err := s.participants.LeavePOI(ctx, poiID, userID); err != nil {
		return fmt.Errorf("failed to leave POI: %w", err)
	}
	
	// Update discussion timer based on new participant count
	if err := s.updateDiscussionTimer(ctx, poiID); err != nil {
		// Log error but don't fail the leave operation
		// Discussion timer is not critical for POI functionality
	}

	// Get updated participant count
	currentCount, err := s.participants.GetParticipantCount(ctx, poiID)
	if err != nil {
		currentCount = 0 // Fallback to 0 if we can't get the count
	}

	// Get updated participant information for the event (after user left)
	participantsInfo, err := s.GetPOIParticipantsWithInfo(ctx, poiID)
	if err != nil {
		// Log error but continue with basic event
		fmt.Printf("Warning: failed to get participant info for POI left event: %v\n", err)
		participantsInfo = []POIParticipantInfo{}
	}

	// Convert to Redis participant format
	var redisParticipants []redis.POIParticipant
	for _, p := range participantsInfo {
		redisParticipants = append(redisParticipants, redis.POIParticipant{
			ID:        p.ID,
			Name:      p.Name,
			AvatarURL: p.AvatarURL,
		})
	}

	// Publish enhanced POI left event with participant information
	leftEventWithParticipants := redis.POILeftEventWithParticipants{
		POIID:        poiID,
		MapID:        poi.MapID,
		UserID:       userID,
		SessionID:    userID, // For now, use userID as sessionID
		CurrentCount: currentCount,
		Participants: redisParticipants,
		Timestamp:    time.Now(),
	}

	if err := s.pubsub.PublishPOILeftWithParticipants(ctx, leftEventWithParticipants); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to publish POI left event with participants: %v\n", err)
	}

	return nil
}

// GetPOIParticipants retrieves all participants of a POI
func (s *POIService) GetPOIParticipants(ctx context.Context, poiID string) ([]string, error) {
	participants, err := s.participants.GetParticipants(ctx, poiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get POI participants: %w", err)
	}
	return participants, nil
}

// GetPOIParticipantCount retrieves the current participant count
func (s *POIService) GetPOIParticipantCount(ctx context.Context, poiID string) (int, error) {
	count, err := s.participants.GetParticipantCount(ctx, poiID)
	if err != nil {
		return 0, fmt.Errorf("failed to get POI participant count: %w", err)
	}
	return count, nil
}

// GetPOIParticipantsWithInfo retrieves all participants of a POI with their display names and avatars
func (s *POIService) GetPOIParticipantsWithInfo(ctx context.Context, poiID string) ([]POIParticipantInfo, error) {
	// Get participant user IDs
	participantIDs, err := s.participants.GetParticipants(ctx, poiID)
	if err != nil {
		return nil, fmt.Errorf("failed to get POI participants: %w", err)
	}

	// Get user information for each participant
	var participantsInfo []POIParticipantInfo
	for _, userID := range participantIDs {
		participantInfo := POIParticipantInfo{
			ID:        userID,
			Name:      userID, // Fallback to userID
			AvatarURL: "",     // No avatar by default
		}

		// Try to get user details if user service is available
		if s.userService != nil {
			user, err := s.userService.GetUser(ctx, userID)
			if err == nil && user != nil {
				participantInfo.Name = user.DisplayName
				if user.AvatarURL != nil {
					participantInfo.AvatarURL = *user.AvatarURL
				}
			}
			// If user service fails, we continue with fallback values
		}

		participantsInfo = append(participantsInfo, participantInfo)
	}

	return participantsInfo, nil
}

// GetUserPOIs retrieves all POIs a user is participating in
func (s *POIService) GetUserPOIs(ctx context.Context, userID string) ([]string, error) {
	poiIDs, err := s.participants.GetPOIsForParticipant(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user POIs: %w", err)
	}
	return poiIDs, nil
}

// ValidatePOI validates that a POI exists
func (s *POIService) ValidatePOI(ctx context.Context, poiID string) (*models.POI, error) {
	poi, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("POI not found: %s", poiID)
		}
		return nil, fmt.Errorf("failed to validate POI: %w", err)
	}
	return poi, nil
}

// Helper methods

// validatePOIInput validates basic POI input parameters
func (s *POIService) validatePOIInput(mapID, name, createdBy string, maxParticipants int) error {
	if mapID == "" {
		return fmt.Errorf("map ID is required")
	}
	if name == "" {
		return fmt.Errorf("POI name is required")
	}
	if len(name) > MaxPOINameLength {
		return fmt.Errorf("POI name too long (max %d characters)", MaxPOINameLength)
	}
	if createdBy == "" {
		return fmt.Errorf("created by is required")
	}
	if maxParticipants < 1 {
		return fmt.Errorf("max participants must be at least 1")
	}
	return nil
}

// validateBounds validates geographic bounds
func (s *POIService) validateBounds(bounds POIBounds) error {
	if bounds.MinLat >= bounds.MaxLat {
		return fmt.Errorf("invalid latitude bounds: min (%f) must be less than max (%f)", bounds.MinLat, bounds.MaxLat)
	}
	if bounds.MinLng >= bounds.MaxLng {
		return fmt.Errorf("invalid longitude bounds: min (%f) must be less than max (%f)", bounds.MinLng, bounds.MaxLng)
	}
	if bounds.MinLat < -90 || bounds.MaxLat > 90 {
		return fmt.Errorf("latitude bounds must be between -90 and 90")
	}
	if bounds.MinLng < -180 || bounds.MaxLng > 180 {
		return fmt.Errorf("longitude bounds must be between -180 and 180")
	}
	return nil
}

// updateDiscussionTimer updates the discussion timer state based on participant count
// Backend only tracks when 2+ users are present - frontend calculates duration
func (s *POIService) updateDiscussionTimer(ctx context.Context, poiID string) error {
	// Get current POI
	poi, err := s.poiRepo.GetByID(ctx, poiID)
	if err != nil {
		return fmt.Errorf("failed to get POI: %w", err)
	}
	
	// Get current participant count
	participantCount, err := s.participants.GetParticipantCount(ctx, poiID)
	if err != nil {
		return fmt.Errorf("failed to get participant count: %w", err)
	}
	
	now := time.Now()
	
	// Determine if discussion should be active (2+ participants)
	shouldBeActive := participantCount >= 2
	needsUpdate := false
	
	if shouldBeActive && !poi.IsDiscussionActive {
		// Start discussion timer - set timestamp when 2+ users are present
		poi.DiscussionStartTime = &now
		poi.IsDiscussionActive = true
		needsUpdate = true
		
	} else if !shouldBeActive && poi.IsDiscussionActive {
		// Stop discussion timer - clear timestamp when users drop below 2
		poi.IsDiscussionActive = false
		poi.DiscussionStartTime = nil
		needsUpdate = true
	}
	
	// Only update POI in database if there was a state change
	if needsUpdate {
		poi.UpdatedAt = now
		if err := s.poiRepo.Update(ctx, poi); err != nil {
			return fmt.Errorf("failed to update POI discussion timer: %w", err)
		}
	}
	
	return nil
}