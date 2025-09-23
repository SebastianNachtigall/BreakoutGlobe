package testdata

import (
	"os"
	"testing"

	"breakoutglobe/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)



func TestSetup_CreatesIsolatedDatabase(t *testing.T) {
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

func TestSetup_MultipleTestsGetIsolatedDatabases(t *testing.T) {
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

func TestRunMigrations_CreatesRequiredTables(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Verify POI table exists and has correct structure
	var exists bool
	err := testDB.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'pois')").Scan(&exists).Error
	assert.NoError(t, err)
	assert.True(t, exists, "POI table should exist")
	
	// Verify Session table exists
	err = testDB.DB.Raw("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'sessions')").Scan(&exists).Error
	assert.NoError(t, err)
	assert.True(t, exists, "Session table should exist")
	
	// Test that we can create records in the tables
	poi := NewDatabasePOI().Build()
	err = testDB.DB.Create(poi).Error
	assert.NoError(t, err)
	
	session := NewDatabaseSession().Build()
	err = testDB.DB.Create(session).Error
	assert.NoError(t, err)
}

func TestSeedFixtures_LoadsTestData(t *testing.T) {
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
	
	// Verify data integrity
	var retrievedPOI models.POI
	err = testDB.DB.Where("name = ?", "Coffee Shop").First(&retrievedPOI).Error
	assert.NoError(t, err)
	assert.Equal(t, poi1.Name, retrievedPOI.Name)
	assert.Equal(t, poi1.Position.Lat, retrievedPOI.Position.Lat)
	assert.Equal(t, poi1.Position.Lng, retrievedPOI.Position.Lng)
}

func TestClear_RemovesAllData(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Seed some data
	poi := NewDatabasePOI().Build()
	session := NewDatabaseSession().Build()
	err := testDB.SeedFixtures(poi, session)
	require.NoError(t, err)
	
	// Verify data exists
	var poiCount, sessionCount int64
	testDB.DB.Model(&models.POI{}).Count(&poiCount)
	testDB.DB.Model(&models.Session{}).Count(&sessionCount)
	assert.Equal(t, int64(1), poiCount)
	assert.Equal(t, int64(1), sessionCount)
	
	// Clear data
	err = testDB.Clear(&models.POI{}, &models.Session{})
	assert.NoError(t, err)
	
	// Verify data is cleared
	testDB.DB.Model(&models.POI{}).Count(&poiCount)
	testDB.DB.Model(&models.Session{}).Count(&sessionCount)
	assert.Equal(t, int64(0), poiCount)
	assert.Equal(t, int64(0), sessionCount)
}

func TestTransaction_RollsBackOnError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Attempt transaction that should fail
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
}

func TestTransaction_CommitsOnSuccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}
	
	testDB := Setup(t)
	require.NotNil(t, testDB)
	
	// Successful transaction
	err := testDB.Transaction(func(tx *gorm.DB) error {
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
	var count int64
	testDB.DB.Model(&models.POI{}).Count(&count)
	assert.Equal(t, int64(2), count)
}

func TestDatabasePOIFixture_BuildsValidPOI(t *testing.T) {
	poi := NewDatabasePOI().
		WithID("poi-123").
		WithName("Coffee Shop").
		WithMapID("map-456").
		WithPosition(40.7128, -74.0060).
		WithCreator("user-789").
		Build()
	
	assert.Equal(t, "poi-123", poi.ID)
	assert.Equal(t, "Coffee Shop", poi.Name)
	assert.Equal(t, "map-456", poi.MapID)
	assert.Equal(t, 40.7128, poi.Position.Lat)
	assert.Equal(t, -74.0060, poi.Position.Lng)
	assert.Equal(t, "user-789", poi.CreatedBy)
}

func TestDatabaseSessionFixture_BuildsValidSession(t *testing.T) {
	session := NewDatabaseSession().
		WithID("session-123").
		WithUserID("user-456").
		WithMapID("map-789").
		WithPosition(41.0, -75.0).
		WithActive(true).
		Build()
	
	assert.Equal(t, "session-123", session.ID)
	assert.Equal(t, "user-456", session.UserID)
	assert.Equal(t, "map-789", session.MapID)
	assert.Equal(t, 41.0, session.AvatarPos.Lat)
	assert.Equal(t, -75.0, session.AvatarPos.Lng)
	assert.True(t, session.IsActive)
}

func TestDefaultTestDBConfig_UsesEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_DB_HOST", "testhost")
	os.Setenv("TEST_DB_PORT", "5433")
	os.Setenv("TEST_DB_USER", "testuser")
	defer func() {
		os.Unsetenv("TEST_DB_HOST")
		os.Unsetenv("TEST_DB_PORT")
		os.Unsetenv("TEST_DB_USER")
	}()
	
	config := DefaultTestDBConfig()
	
	assert.Equal(t, "testhost", config.Host)
	assert.Equal(t, "5433", config.Port)
	assert.Equal(t, "testuser", config.User)
	assert.Equal(t, "postgres", config.Password) // Should use default
}

func TestSanitizeTestName_HandlesSpecialCharacters(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"TestName", "TestName"},
		{"Test/Name", "Test_Name"},
		{"Test Name", "Test_Name"},
		{"Test-Name", "Test_Name"},
		{"123Test", "t_123Test"},
		{"", ""},
	}
	
	for _, tt := range tests {
		result := sanitizeTestName(tt.input)
		assert.Equal(t, tt.expected, result, "Input: %s", tt.input)
	}
}

// Integration test for repository layer with real database
func TestRepositoryIntegration_POIOperations(t *testing.T) {
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
	
	// Test POI update
	retrievedPOI.Name = "Updated POI Name"
	err = testDB.DB.Save(&retrievedPOI).Error
	assert.NoError(t, err)
	
	// Verify update
	var updatedPOI models.POI
	err = testDB.DB.Where("id = ?", originalPOI.ID).First(&updatedPOI).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated POI Name", updatedPOI.Name)
	
	// Test POI deletion
	err = testDB.DB.Delete(&updatedPOI).Error
	assert.NoError(t, err)
	
	// Verify deletion
	var count int64
	err = testDB.DB.Model(&models.POI{}).Where("id = ?", originalPOI.ID).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

// Integration test for session operations
func TestRepositoryIntegration_SessionOperations(t *testing.T) {
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
	
	// Test session update (avatar position)
	retrievedSession.AvatarPos = models.LatLng{Lat: 41.0, Lng: -75.0}
	err = testDB.DB.Save(&retrievedSession).Error
	assert.NoError(t, err)
	
	// Verify update
	var updatedSession models.Session
	err = testDB.DB.Where("id = ?", originalSession.ID).First(&updatedSession).Error
	assert.NoError(t, err)
	assert.Equal(t, 41.0, updatedSession.AvatarPos.Lat)
	assert.Equal(t, -75.0, updatedSession.AvatarPos.Lng)
	
	// Test querying active sessions for a map
	var activeSessions []models.Session
	err = testDB.DB.Where("map_id = ? AND is_active = ?", "integration-map", true).Find(&activeSessions).Error
	assert.NoError(t, err)
	assert.Len(t, activeSessions, 1)
	assert.Equal(t, originalSession.ID, activeSessions[0].ID)
}