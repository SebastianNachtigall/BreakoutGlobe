package services

import (
	"context"
	"testing"

	"breakoutglobe/internal/models"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MockPOIRepository is a mock implementation of repository.POIRepositoryInterface
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

// MockPOIParticipants is a mock implementation of redis.POIParticipants
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
	return args.Get(0).(int), args.Error(1)
}

func (m *MockPOIParticipants) IsParticipant(ctx context.Context, poiID, userID string) (bool, error) {
	args := m.Called(ctx, poiID, userID)
	return args.Get(0).(bool), args.Error(1)
}

func (m *MockPOIParticipants) CanJoinPOI(ctx context.Context, poiID string, maxParticipants int) (bool, error) {
	args := m.Called(ctx, poiID, maxParticipants)
	return args.Get(0).(bool), args.Error(1)
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

// POIServiceTestSuite contains the test suite for POIService
type POIServiceTestSuite struct {
	suite.Suite
	mockRepo         *MockPOIRepository
	mockParticipants *MockPOIParticipants
	mockPubSub       *MockPubSub
	service          *POIService
}

func (suite *POIServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockPOIRepository)
	suite.mockParticipants = new(MockPOIParticipants)
	suite.mockPubSub = new(MockPubSub)
	suite.service = NewPOIService(suite.mockRepo, suite.mockParticipants, suite.mockPubSub)
}

func (suite *POIServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockParticipants.AssertExpectations(suite.T())
	suite.mockPubSub.AssertExpectations(suite.T())
}

func (suite *POIServiceTestSuite) TestCreatePOI() {
	ctx := context.Background()
	mapID := "map-123"
	name := "Meeting Room"
	description := "A place for team meetings"
	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}
	createdBy := "user-123"
	maxParticipants := 10

	// Mock expectations
	suite.mockRepo.On("CheckDuplicateLocation", ctx, mapID, position.Lat, position.Lng, "").Return([]*models.POI{}, nil)
	suite.mockRepo.On("Create", ctx, mock.AnythingOfType("*models.POI")).Return(nil)
	suite.mockPubSub.On("PublishPOICreated", ctx, mock.AnythingOfType("redis.POICreatedEvent")).Return(nil)

	// Execute
	poi, err := suite.service.CreatePOI(ctx, mapID, name, description, position, createdBy, maxParticipants)

	// Assert
	suite.NoError(err)
	suite.NotNil(poi)
	suite.Equal(mapID, poi.MapID)
	suite.Equal(name, poi.Name)
	suite.Equal(description, poi.Description)
	suite.Equal(position.Lat, poi.Position.Lat)
	suite.Equal(position.Lng, poi.Position.Lng)
	suite.Equal(createdBy, poi.CreatedBy)
	suite.Equal(maxParticipants, poi.MaxParticipants)
	suite.NotEmpty(poi.ID)
}

func (suite *POIServiceTestSuite) TestCreatePOI_DuplicateLocation() {
	ctx := context.Background()
	mapID := "map-123"
	name := "Meeting Room"
	description := "A place for team meetings"
	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}
	createdBy := "user-123"
	maxParticipants := 10

	// Existing POI at same location
	existingPOI := &models.POI{
		ID:     "existing-poi-id",
		MapID:  mapID,
		Name:   "Existing POI",
		Position: position,
	}

	// Mock expectations
	suite.mockRepo.On("CheckDuplicateLocation", ctx, mapID, position.Lat, position.Lng, "").Return([]*models.POI{existingPOI}, nil)

	// Execute
	poi, err := suite.service.CreatePOI(ctx, mapID, name, description, position, createdBy, maxParticipants)

	// Assert
	suite.Error(err)
	suite.Nil(poi)
	suite.Contains(err.Error(), "POI already exists at this location")
}

func (suite *POIServiceTestSuite) TestGetPOI() {
	ctx := context.Background()
	poiID := "poi-123"

	// Expected POI
	expectedPOI := &models.POI{
		ID:     poiID,
		MapID:  "map-456",
		Name:   "Test POI",
	}

	// Mock expectations
	suite.mockRepo.On("GetByID", ctx, poiID).Return(expectedPOI, nil)

	// Execute
	poi, err := suite.service.GetPOI(ctx, poiID)

	// Assert
	suite.NoError(err)
	suite.Equal(expectedPOI, poi)
}

func (suite *POIServiceTestSuite) TestGetPOI_NotFound() {
	ctx := context.Background()
	poiID := "non-existent-poi"

	// Mock expectations
	suite.mockRepo.On("GetByID", ctx, poiID).Return((*models.POI)(nil), gorm.ErrRecordNotFound)

	// Execute
	poi, err := suite.service.GetPOI(ctx, poiID)

	// Assert
	suite.Error(err)
	suite.Nil(poi)
}

func (suite *POIServiceTestSuite) TestJoinPOI() {
	ctx := context.Background()
	poiID := "poi-123"
	userID := "user-456"

	// Existing POI
	existingPOI := &models.POI{
		ID:     poiID,
		MapID:  "map-789",
		Name:   "Test POI",
		MaxParticipants: 10,
	}

	// Mock expectations
	suite.mockRepo.On("GetByID", ctx, poiID).Return(existingPOI, nil)
	suite.mockParticipants.On("IsParticipant", ctx, poiID, userID).Return(false, nil)
	suite.mockParticipants.On("CanJoinPOI", ctx, poiID, existingPOI.MaxParticipants).Return(true, nil)
	suite.mockParticipants.On("JoinPOI", ctx, poiID, userID).Return(nil)
	suite.mockPubSub.On("PublishPOIJoined", ctx, mock.AnythingOfType("redis.POIJoinedEvent")).Return(nil)

	// Execute
	err := suite.service.JoinPOI(ctx, poiID, userID)

	// Assert
	suite.NoError(err)
}

func (suite *POIServiceTestSuite) TestJoinPOI_AlreadyParticipant() {
	ctx := context.Background()
	poiID := "poi-123"
	userID := "user-456"

	// Existing POI
	existingPOI := &models.POI{
		ID:     poiID,
		MapID:  "map-789",
		Name:   "Test POI",
		MaxParticipants: 10,
	}

	// Mock expectations
	suite.mockRepo.On("GetByID", ctx, poiID).Return(existingPOI, nil)
	suite.mockParticipants.On("IsParticipant", ctx, poiID, userID).Return(true, nil)

	// Execute
	err := suite.service.JoinPOI(ctx, poiID, userID)

	// Assert
	suite.Error(err)
	suite.Contains(err.Error(), "user is already a participant")
}

func (suite *POIServiceTestSuite) TestLeavePOI() {
	ctx := context.Background()
	poiID := "poi-123"
	userID := "user-456"

	// Existing POI
	existingPOI := &models.POI{
		ID:     poiID,
		MapID:  "map-789",
		Name:   "Test POI",
	}

	// Mock expectations
	suite.mockRepo.On("GetByID", ctx, poiID).Return(existingPOI, nil)
	suite.mockParticipants.On("IsParticipant", ctx, poiID, userID).Return(true, nil)
	suite.mockParticipants.On("LeavePOI", ctx, poiID, userID).Return(nil)
	suite.mockPubSub.On("PublishPOILeft", ctx, mock.AnythingOfType("redis.POILeftEvent")).Return(nil)

	// Execute
	err := suite.service.LeavePOI(ctx, poiID, userID)

	// Assert
	suite.NoError(err)
}

func TestPOIServiceTestSuite(t *testing.T) {
	suite.Run(t, new(POIServiceTestSuite))
}