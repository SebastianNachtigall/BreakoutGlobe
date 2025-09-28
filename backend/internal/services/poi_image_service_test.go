package services

import (
	"context"
	"mime/multipart"
	"testing"

	"breakoutglobe/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// POI Image Service Test Scenario
type poiImageServiceScenario struct {
	t               *testing.T
	mockRepo        *MockPOIRepository
	mockParts       *MockPOIParticipants
	mockPubsub      *MockPubSub
	mockUploader    *MockImageUploader
	mockUserService *MockUserService
	service         *POIService
}

func newPOIImageServiceScenario(t *testing.T) *poiImageServiceScenario {
	mockRepo := new(MockPOIRepository)
	mockParts := new(MockPOIParticipants)
	mockPubsub := new(MockPubSub)
	mockUploader := new(MockImageUploader)
	mockUserService := new(MockUserService)
	
	service := NewPOIServiceWithImageUploader(mockRepo, mockParts, mockPubsub, mockUploader, mockUserService)
	
	return &poiImageServiceScenario{
		t:               t,
		mockRepo:        mockRepo,
		mockParts:       mockParts,
		mockPubsub:      mockPubsub,
		mockUploader:    mockUploader,
		mockUserService: mockUserService,
		service:         service,
	}
}

func (s *poiImageServiceScenario) expectNoDuplicateLocation() *poiImageServiceScenario {
	s.mockRepo.On("CheckDuplicateLocation", mock.Anything, "map-123", 40.7128, -74.0060, "").Return([]*models.POI{}, nil)
	return s
}

func (s *poiImageServiceScenario) expectImageUploadSuccess() *poiImageServiceScenario {
	s.mockUploader.On("UploadPOIImage", mock.Anything, mock.AnythingOfType("*multipart.FileHeader")).Return("https://example.com/uploads/poi-image.jpg", nil)
	return s
}

func (s *poiImageServiceScenario) expectPOICreationSuccess() *poiImageServiceScenario {
	s.mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*models.POI")).Return(nil)
	return s
}

func (s *poiImageServiceScenario) expectEventPublishing() *poiImageServiceScenario {
	s.mockPubsub.On("PublishPOICreated", mock.Anything, mock.AnythingOfType("redis.POICreatedEvent")).Return(nil)
	return s
}

func (s *poiImageServiceScenario) createMockImageFile() *multipart.FileHeader {
	return &multipart.FileHeader{
		Filename: "test-image.jpg",
		Size:     1024,
		Header:   map[string][]string{"Content-Type": {"image/jpeg"}},
	}
}

func (s *poiImageServiceScenario) cleanup() {
	s.mockRepo.AssertExpectations(s.t)
	s.mockParts.AssertExpectations(s.t)
	s.mockPubsub.AssertExpectations(s.t)
	s.mockUploader.AssertExpectations(s.t)
	s.mockUserService.AssertExpectations(s.t)
}

func TestCreatePOIWithImage_Success(t *testing.T) {
	scenario := newPOIImageServiceScenario(t)
	defer scenario.cleanup()

	scenario.expectNoDuplicateLocation().
		expectImageUploadSuccess().
		expectPOICreationSuccess().
		expectEventPublishing()

	imageFile := scenario.createMockImageFile()
	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}

	poi, err := scenario.service.CreatePOIWithImage(
		context.Background(),
		"map-123",
		"Coffee Shop",
		"Great place to meet",
		position,
		"user-123",
		15,
		imageFile,
	)

	assert.NoError(t, err)
	assert.NotNil(t, poi)
	assert.Equal(t, "Coffee Shop", poi.Name)
	assert.Equal(t, "https://example.com/uploads/poi-image.jpg", poi.ImageURL)
}

func TestCreatePOIWithImage_WithoutImage_Success(t *testing.T) {
	scenario := newPOIImageServiceScenario(t)
	defer scenario.cleanup()

	scenario.expectNoDuplicateLocation().
		expectPOICreationSuccess().
		expectEventPublishing()

	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}

	poi, err := scenario.service.CreatePOIWithImage(
		context.Background(),
		"map-123",
		"Coffee Shop",
		"Great place to meet",
		position,
		"user-123",
		15,
		nil, // No image file
	)

	assert.NoError(t, err)
	assert.NotNil(t, poi)
	assert.Equal(t, "Coffee Shop", poi.Name)
	assert.Empty(t, poi.ImageURL)
}

// Mock interfaces for testing

type MockImageUploader struct {
	mock.Mock
}

func (m *MockImageUploader) UploadPOIImage(ctx context.Context, imageFile *multipart.FileHeader) (string, error) {
	args := m.Called(ctx, imageFile)
	return args.String(0), args.Error(1)
}

// Mock POI Repository
type MockPOIRepository struct {
	mock.Mock
}

func (m *MockPOIRepository) Create(ctx context.Context, poi *models.POI) error {
	args := m.Called(ctx, poi)
	return args.Error(0)
}

func (m *MockPOIRepository) GetByID(ctx context.Context, id string) (*models.POI, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.POI), args.Error(1)
}

func (m *MockPOIRepository) GetByMapID(ctx context.Context, mapID string) ([]*models.POI, error) {
	args := m.Called(ctx, mapID)
	return args.Get(0).([]*models.POI), args.Error(1)
}

func (m *MockPOIRepository) GetInBounds(ctx context.Context, mapID string, minLat, maxLat, minLng, maxLng float64) ([]*models.POI, error) {
	args := m.Called(ctx, mapID, minLat, maxLat, minLng, maxLng)
	return args.Get(0).([]*models.POI), args.Error(1)
}

func (m *MockPOIRepository) Update(ctx context.Context, poi *models.POI) error {
	args := m.Called(ctx, poi)
	return args.Error(0)
}

func (m *MockPOIRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPOIRepository) CheckDuplicateLocation(ctx context.Context, mapID string, lat, lng float64, excludeID string) ([]*models.POI, error) {
	args := m.Called(ctx, mapID, lat, lng, excludeID)
	return args.Get(0).([]*models.POI), args.Error(1)
}

// Mock POI Participants
type MockPOIParticipants struct {
	mock.Mock
}

func (m *MockPOIParticipants) JoinPOI(ctx context.Context, poiID, userID string) error {
	args := m.Called(ctx, poiID, userID)
	return args.Error(0)
}

func (m *MockPOIParticipants) LeavePOI(ctx context.Context, poiID, userID string) error {
	args := m.Called(ctx, poiID, userID)
	return args.Error(0)
}

func (m *MockPOIParticipants) GetParticipants(ctx context.Context, poiID string) ([]string, error) {
	args := m.Called(ctx, poiID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPOIParticipants) GetParticipantCount(ctx context.Context, poiID string) (int, error) {
	args := m.Called(ctx, poiID)
	return args.Int(0), args.Error(1)
}

func (m *MockPOIParticipants) IsParticipant(ctx context.Context, poiID, userID string) (bool, error) {
	args := m.Called(ctx, poiID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockPOIParticipants) CanJoinPOI(ctx context.Context, poiID string, maxParticipants int) (bool, error) {
	args := m.Called(ctx, poiID, maxParticipants)
	return args.Bool(0), args.Error(1)
}

func (m *MockPOIParticipants) RemoveAllParticipants(ctx context.Context, poiID string) error {
	args := m.Called(ctx, poiID)
	return args.Error(0)
}

func (m *MockPOIParticipants) RemoveParticipantFromAllPOIs(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockPOIParticipants) GetPOIsForParticipant(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]string), args.Error(1)
}

// MockPubSub is now defined in mocks.go