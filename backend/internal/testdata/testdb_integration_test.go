package testdata

import (
	"os"
	"testing"

	"breakoutglobe/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestMain(m *testing.M) {
	// Skip integration tests if TEST_INTEGRATION is not set
	if os.Getenv("TEST_INTEGRATION") == "" {
		os.Exit(0)
	}
	
	// Run tests
	code := m.Run()
	os.Exit(code)
}

func TestDatabaseIntegration_Setup(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Setup creates isolated test database
	testDB := Setup(t)
	require.NotNil(t, testDB)
	require.NotNil(t, testDB.DB)
	
	// Verify database connection works
	err := testDB.DB.Exec("SELECT 1").Error
	assert.NoError(t, err)
	
	// Verify migrations ran successfully
	var tableCount int64
	err = testDB.DB.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public'").Scan(&tableCount).Error
	assert.NoError(t, err)
	assert.Greater(t, tableCount, int64(0), "Expected tables to be created by migrations")
}

func TestDatabaseIntegration_POIOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Test POI creation and retrieval
	originalPOI := NewDatabasePOI().
		WithName("Integration Test POI").
		WithMapID("integration-map").
		WithPosition(40.7128, -74.0060).
		Build()
	
	// Create POI
	err := testDB.DB.Create(originalPOI).Error
	require.NoError(t, err)
	
	// Retrieve POI by ID
	var retrievedPOI models.POI
	err = testDB.DB.Where("id = ?", originalPOI.ID).First(&retrievedPOI).Error
	assert.NoError(t, err)
	assert.Equal(t, originalPOI.Name, retrievedPOI.Name)
	assert.Equal(t, originalPOI.MapID, retrievedPOI.MapID)
	
	// Test spatial query (POIs within bounds)
	var poisInBounds []models.POI
	err = testDB.DB.Where("position->>'lat' BETWEEN ? AND ? AND position->>'lng' BETWEEN ? AND ?",
		"40.7000", "40.8000", "-74.1000", "-73.9000").Find(&poisInBounds).Error
	assert.NoError(t, err)
	assert.Len(t, poisInBounds, 1)
	assert.Equal(t, originalPOI.ID, poisInBounds[0].ID)
}

func TestDatabaseIntegration_SessionOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Test session creation and retrieval
	originalSession := NewDatabaseSession().
		WithUserID("integration-user").
		WithMapID("integration-map").
		WithPosition(40.7128, -74.0060).
		WithActive(true).
		Build()
	
	// Create session
	err := testDB.DB.Create(originalSession).Error
	require.NoError(t, err)
	
	// Retrieve session by ID
	var retrievedSession models.Session
	err = testDB.DB.Where("id = ?", originalSession.ID).First(&retrievedSession).Error
	assert.NoError(t, err)
	assert.Equal(t, originalSession.UserID, retrievedSession.UserID)
	assert.Equal(t, originalSession.MapID, retrievedSession.MapID)
	assert.True(t, retrievedSession.IsActive)
}

func TestDatabaseIntegration_Fixtures(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Create test fixtures
	poi1 := NewDatabasePOI().WithName("Coffee Shop").WithPosition(40.7128, -74.0060).Build()
	poi2 := NewDatabasePOI().WithName("Park Bench").WithPosition(40.7589, -73.9851).Build()
	session1 := NewDatabaseSession().WithUserID("user-123").Build()
	
	// Seed fixtures
	err := testDB.SeedFixtures(poi1, poi2, session1)
	assert.NoError(t, err)
	
	// Verify fixtures were loaded
	var poiCount int64
	err = testDB.DB.Model(&models.POI{}).Count(&poiCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(2), poiCount)
	
	var sessionCount int64
	err = testDB.DB.Model(&models.Session{}).Count(&sessionCount).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), sessionCount)
}

func TestDatabaseIntegration_Isolation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	// Create two test databases
	testDB1 := Setup(t)
	testDB2 := Setup(t)
	
	require.NotNil(t, testDB1)
	require.NotNil(t, testDB2)
	
	// Verify they have different database names
	assert.NotEqual(t, testDB1.dbName, testDB2.dbName)
	
	// Verify they are isolated - data in one doesn't affect the other
	poi1 := NewDatabasePOI().WithName("POI in DB1").Build()
	err := testDB1.SeedFixtures(poi1)
	require.NoError(t, err)
	
	// Check that POI exists in DB1
	var count1 int64
	err = testDB1.DB.Model(&models.POI{}).Count(&count1).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(1), count1)
	
	// Check that POI doesn't exist in DB2
	var count2 int64
	err = testDB2.DB.Model(&models.POI{}).Count(&count2).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count2)
}

func TestDatabaseIntegration_Transaction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Test transaction rollback on error
	err := testDB.Transaction(func(tx *gorm.DB) error {
		// Create a POI
		poi := NewDatabasePOI().Build()
		if err := tx.Create(poi).Error; err != nil {
			return err
		}
		
		// Simulate an error
		return assert.AnError
	})
	
	// Transaction should have failed
	assert.Error(t, err)
	
	// Verify no data was committed
	var count int64
	testDB.DB.Model(&models.POI{}).Count(&count)
	assert.Equal(t, int64(0), count)
	
	// Test successful transaction
	err = testDB.Transaction(func(tx *gorm.DB) error {
		poi1 := NewDatabasePOI().WithName("POI 1").Build()
		poi2 := NewDatabasePOI().WithName("POI 2").Build()
		
		if err := tx.Create(poi1).Error; err != nil {
			return err
		}
		if err := tx.Create(poi2).Error; err != nil {
			return err
		}
		
		return nil
	})
	
	// Transaction should succeed
	assert.NoError(t, err)
	
	// Verify data was committed
	testDB.DB.Model(&models.POI{}).Count(&count)
	assert.Equal(t, int64(2), count)
}