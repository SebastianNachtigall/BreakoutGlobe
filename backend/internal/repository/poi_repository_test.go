package repository

import (
	"context"
	"testing"
	"time"

	"breakoutglobe/internal/database"
	"breakoutglobe/internal/models"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

type POIRepositoryTestSuite struct {
	suite.Suite
	db   *database.DB
	repo *POIRepository
}

func (suite *POIRepositoryTestSuite) SetupSuite() {
	// Skip integration tests in short mode
	if testing.Short() {
		suite.T().Skip("Skipping database integration test in short mode")
	}
	
	// Set up test database
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := database.Initialize(testURL)
	require.NoError(suite.T(), err)
	
	suite.db = db
	suite.repo = NewPOIRepository(db)
}

func (suite *POIRepositoryTestSuite) SetupTest() {
	// Clean up tables before each test
	suite.db.Exec("DELETE FROM pois")
	suite.db.Exec("DELETE FROM sessions")
	suite.db.Exec("DELETE FROM maps")
}

func (suite *POIRepositoryTestSuite) TearDownSuite() {
	database.CloseConnection(suite.db)
}

func (suite *POIRepositoryTestSuite) createTestMap() *models.Map {
	testMap := &models.Map{
		ID:          "map-123",
		Name:        "Test Map",
		Description: "",
		CreatedBy:   "facilitator-456",
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err := suite.db.Create(testMap).Error
	suite.Require().NoError(err)
	return testMap
}

func (suite *POIRepositoryTestSuite) TestCreate() {
	// Setup
	testMap := suite.createTestMap()
	
	poi := &models.POI{
		MapID:           testMap.ID,
		Name:            "Test Meeting Room",
		Description:     "A place for team meetings",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}

	// Execute
	result, err := suite.repo.Create(context.Background(), poi)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.NotEmpty(result.ID)
	suite.Equal(poi.MapID, result.MapID)
	suite.Equal(poi.Name, result.Name)
	suite.Equal(poi.Description, result.Description)
	suite.Equal(poi.Position.Lat, result.Position.Lat)
	suite.Equal(poi.Position.Lng, result.Position.Lng)
	suite.Equal(poi.CreatedBy, result.CreatedBy)
	suite.Equal(poi.MaxParticipants, result.MaxParticipants)
	suite.False(result.CreatedAt.IsZero())
}

func (suite *POIRepositoryTestSuite) TestCreate_DuplicateLocation() {
	// Setup
	testMap := suite.createTestMap()
	
	position := models.LatLng{Lat: 40.7128, Lng: -74.0060}
	
	poi1 := &models.POI{
		MapID:           testMap.ID,
		Name:            "First POI",
		Position:        position,
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}
	
	poi2 := &models.POI{
		MapID:           testMap.ID,
		Name:            "Second POI",
		Position:        position,
		CreatedBy:       "user-456",
		MaxParticipants: 10,
	}

	// Execute
	_, err1 := suite.repo.Create(context.Background(), poi1)
	_, err2 := suite.repo.Create(context.Background(), poi2)

	// Assert - First creation should succeed, second should fail due to proximity
	suite.NoError(err1)
	suite.Error(err2)
	suite.Contains(err2.Error(), "too close to existing POI")
}

func (suite *POIRepositoryTestSuite) TestGetByID() {
	// Setup
	testMap := suite.createTestMap()
	
	poi := &models.POI{
		MapID:           testMap.ID,
		Name:            "Test POI",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}
	
	created, err := suite.repo.Create(context.Background(), poi)
	suite.Require().NoError(err)

	// Execute
	result, err := suite.repo.GetByID(context.Background(), created.ID)

	// Assert
	suite.NoError(err)
	suite.NotNil(result)
	suite.Equal(created.ID, result.ID)
	suite.Equal(created.Name, result.Name)
}

func (suite *POIRepositoryTestSuite) TestGetByID_NotFound() {
	// Execute
	result, err := suite.repo.GetByID(context.Background(), "non-existent-id")

	// Assert
	suite.Error(err)
	suite.Nil(result)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

func (suite *POIRepositoryTestSuite) TestGetByMapID() {
	// Setup
	testMap := suite.createTestMap()
	
	poi1 := &models.POI{
		MapID:           testMap.ID,
		Name:            "POI 1",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}
	
	poi2 := &models.POI{
		MapID:           testMap.ID,
		Name:            "POI 2",
		Position:        models.LatLng{Lat: 41.0000, Lng: -75.0000},
		CreatedBy:       "user-456",
		MaxParticipants: 15,
	}
	
	// Create POIs
	_, err1 := suite.repo.Create(context.Background(), poi1)
	_, err2 := suite.repo.Create(context.Background(), poi2)
	suite.Require().NoError(err1)
	suite.Require().NoError(err2)

	// Execute
	results, err := suite.repo.GetByMapID(context.Background(), testMap.ID)

	// Assert
	suite.NoError(err)
	suite.Len(results, 2)
	
	// Verify both POIs are returned
	names := make([]string, len(results))
	for i, poi := range results {
		names[i] = poi.Name
	}
	suite.Contains(names, "POI 1")
	suite.Contains(names, "POI 2")
}

func (suite *POIRepositoryTestSuite) TestGetInBounds() {
	// Setup
	testMap := suite.createTestMap()
	
	// POI inside bounds
	poiInside := &models.POI{
		MapID:           testMap.ID,
		Name:            "Inside POI",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}
	
	// POI outside bounds
	poiOutside := &models.POI{
		MapID:           testMap.ID,
		Name:            "Outside POI",
		Position:        models.LatLng{Lat: 50.0000, Lng: -80.0000},
		CreatedBy:       "user-456",
		MaxParticipants: 10,
	}
	
	_, err1 := suite.repo.Create(context.Background(), poiInside)
	_, err2 := suite.repo.Create(context.Background(), poiOutside)
	suite.Require().NoError(err1)
	suite.Require().NoError(err2)

	// Define bounds that include only the first POI
	bounds := models.Bounds{
		North: 41.0,
		South: 40.0,
		East:  -73.0,
		West:  -75.0,
	}

	// Execute
	results, err := suite.repo.GetInBounds(context.Background(), testMap.ID, bounds)

	// Assert
	suite.NoError(err)
	suite.Len(results, 1)
	suite.Equal("Inside POI", results[0].Name)
}

func (suite *POIRepositoryTestSuite) TestUpdate() {
	// Setup
	testMap := suite.createTestMap()
	
	poi := &models.POI{
		MapID:           testMap.ID,
		Name:            "Original Name",
		Description:     "Original Description",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}
	
	created, err := suite.repo.Create(context.Background(), poi)
	suite.Require().NoError(err)

	// Modify POI
	created.Name = "Updated Name"
	created.Description = "Updated Description"
	created.MaxParticipants = 15

	// Execute
	err = suite.repo.Update(context.Background(), created)

	// Assert
	suite.NoError(err)
	
	// Verify update
	updated, err := suite.repo.GetByID(context.Background(), created.ID)
	suite.NoError(err)
	suite.Equal("Updated Name", updated.Name)
	suite.Equal("Updated Description", updated.Description)
	suite.Equal(15, updated.MaxParticipants)
}

func (suite *POIRepositoryTestSuite) TestDelete() {
	// Setup
	testMap := suite.createTestMap()
	
	poi := &models.POI{
		MapID:           testMap.ID,
		Name:            "Test POI",
		Position:        models.LatLng{Lat: 40.7128, Lng: -74.0060},
		CreatedBy:       "user-123",
		MaxParticipants: 10,
	}
	
	created, err := suite.repo.Create(context.Background(), poi)
	suite.Require().NoError(err)

	// Execute
	err = suite.repo.Delete(context.Background(), created.ID)

	// Assert
	suite.NoError(err)
	
	// Verify deletion
	_, err = suite.repo.GetByID(context.Background(), created.ID)
	suite.Error(err)
	suite.Equal(gorm.ErrRecordNotFound, err)
}

func TestPOIRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(POIRepositoryTestSuite))
}