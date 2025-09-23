package repository

import (
	"context"
	"fmt"
	"testing"
	"time"

	"breakoutglobe/internal/models"
	"breakoutglobe/internal/testdata"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPOIRepository_Integration demonstrates database integration testing
// These tests use real PostgreSQL database to validate repository behavior

func TestPOIRepository_Create_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	// Setup test database
	testDB := testdata.Setup(t)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Create POI using builder
	poi := testdata.NewDatabasePOI().
		WithName("Integration Test POI").
		WithMapID("map-integration").
		WithPosition(40.7128, -74.0060).
		WithCreator("user-integration").
		Build()

	// Test repository create
	createdPOI, err := repo.Create(ctx, poi)
	require.NoError(t, err)
	require.NotNil(t, createdPOI)

	// Verify POI was persisted correctly
	assert.NotEmpty(t, createdPOI.ID)
	assert.WithinDuration(t, time.Now(), createdPOI.CreatedAt, time.Second)

	// Verify in database
	var count int64
	err = testDB.DB.Model(&models.POI{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Retrieve and verify data integrity
	var savedPOI models.POI
	err = testDB.DB.First(&savedPOI).Error
	require.NoError(t, err)
	assert.Equal(t, createdPOI.Name, savedPOI.Name)
	assert.Equal(t, createdPOI.MapID, savedPOI.MapID)
	assert.Equal(t, createdPOI.Position.Lat, savedPOI.Position.Lat)
	assert.Equal(t, createdPOI.Position.Lng, savedPOI.Position.Lng)
	assert.Equal(t, createdPOI.CreatedBy, savedPOI.CreatedBy)
}

func TestPOIRepository_GetByID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data
	poi1 := testdata.NewDatabasePOI().WithName("POI 1").WithMapID("map-1").Build()
	poi2 := testdata.NewDatabasePOI().WithName("POI 2").WithMapID("map-1").Build()
	err := testDB.SeedFixtures(poi1, poi2)
	require.NoError(t, err)

	// Test successful retrieval
	foundPOI, err := repo.GetByID(ctx, poi1.ID)
	require.NoError(t, err)
	require.NotNil(t, foundPOI)
	assert.Equal(t, poi1.ID, foundPOI.ID)
	assert.Equal(t, "POI 1", foundPOI.Name)

	// Test not found scenario
	notFoundPOI, err := repo.GetByID(ctx, "non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, notFoundPOI)
}

func TestPOIRepository_GetByMapID_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data across multiple maps
	poi1 := testdata.NewDatabasePOI().WithName("Map1 POI 1").WithMapID("map-1").Build()
	poi2 := testdata.NewDatabasePOI().WithName("Map1 POI 2").WithMapID("map-1").Build()
	poi3 := testdata.NewDatabasePOI().WithName("Map2 POI 1").WithMapID("map-2").Build()
	err := testDB.SeedFixtures(poi1, poi2, poi3)
	require.NoError(t, err)

	// Test retrieval by map ID
	map1POIs, err := repo.GetByMapID(ctx, "map-1")
	require.NoError(t, err)
	assert.Len(t, map1POIs, 2)

	// Verify POIs belong to correct map
	for _, poi := range map1POIs {
		assert.Equal(t, "map-1", poi.MapID)
		assert.Contains(t, poi.Name, "Map1")
	}

	// Test empty map
	emptyMapPOIs, err := repo.GetByMapID(ctx, "empty-map")
	require.NoError(t, err)
	assert.Empty(t, emptyMapPOIs)
}

func TestPOIRepository_Update_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Seed initial POI
	originalPOI := testdata.NewDatabasePOI().
		WithName("Original Name").
		WithMapID("map-1").
		Build()
	err := testDB.SeedFixtures(originalPOI)
	require.NoError(t, err)

	// Update POI
	originalPOI.Name = "Updated Name"
	originalPOI.Description = "Updated Description"

	err = repo.Update(ctx, originalPOI)
	require.NoError(t, err)

	// Verify update in database
	retrievedPOI, err := repo.GetByID(ctx, originalPOI.ID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Name", retrievedPOI.Name)
	assert.Equal(t, "Updated Description", retrievedPOI.Description)

	// Verify created timestamp unchanged
	assert.Equal(t, originalPOI.CreatedAt, retrievedPOI.CreatedAt)
}

func TestPOIRepository_Delete_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data
	poi1 := testdata.NewDatabasePOI().WithName("POI to Delete").WithMapID("map-1").Build()
	poi2 := testdata.NewDatabasePOI().WithName("POI to Keep").WithMapID("map-1").Build()
	err := testDB.SeedFixtures(poi1, poi2)
	require.NoError(t, err)

	// Verify initial state
	var count int64
	err = testDB.DB.Model(&models.POI{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	// Delete POI
	err = repo.Delete(ctx, poi1.ID)
	require.NoError(t, err)

	// Verify deletion
	err = testDB.DB.Model(&models.POI{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)

	// Verify correct POI was deleted
	remainingPOI, err := repo.GetByID(ctx, poi2.ID)
	require.NoError(t, err)
	assert.Equal(t, "POI to Keep", remainingPOI.Name)

	// Verify deleted POI is not found
	deletedPOI, err := repo.GetByID(ctx, poi1.ID)
	assert.Error(t, err)
	assert.Nil(t, deletedPOI)
}

func TestPOIRepository_ConcurrentAccess_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping database integration test in short mode")
	}

	testDB := testdata.Setup(t)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Test concurrent POI creation
	const numGoroutines = 10
	done := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			poi := testdata.NewDatabasePOI().
				WithName(fmt.Sprintf("Concurrent POI %d", index)).
				WithMapID("concurrent-map").
				WithPosition(40.0+float64(index)*0.001, -74.0+float64(index)*0.001).
				Build()

			_, err := repo.Create(ctx, poi)
			done <- err
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		err := <-done
		assert.NoError(t, err)
	}

	// Verify all POIs were created
	var count int64
	err := testDB.DB.Model(&models.POI{}).Count(&count).Error
	require.NoError(t, err)
	assert.Equal(t, int64(numGoroutines), count)

	// Verify all POIs have unique IDs
	pois, err := repo.GetByMapID(ctx, "concurrent-map")
	require.NoError(t, err)
	assert.Len(t, pois, numGoroutines)

	ids := make(map[string]bool)
	for _, poi := range pois {
		assert.False(t, ids[poi.ID], "Duplicate POI ID found: %s", poi.ID)
		ids[poi.ID] = true
	}
}

// Benchmark tests for performance validation
func BenchmarkPOIRepository_Create_Integration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database integration benchmark in short mode")
	}

	testDB := testdata.Setup(b)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		poi := testdata.NewDatabasePOI().
			WithName(fmt.Sprintf("Benchmark POI %d", i)).
			WithMapID("benchmark-map").
			WithPosition(40.0+float64(i)*0.0001, -74.0+float64(i)*0.0001).
			Build()

		if _, err := repo.Create(ctx, poi); err != nil {
			b.Fatalf("Failed to create POI: %v", err)
		}
	}
}

func BenchmarkPOIRepository_GetByMapID_Integration(b *testing.B) {
	if testing.Short() {
		b.Skip("Skipping database integration benchmark in short mode")
	}

	testDB := testdata.Setup(b)
	repo := NewPOIRepository(testDB.DB)
	ctx := context.Background()

	// Seed test data
	for i := 0; i < 100; i++ {
		poi := testdata.NewDatabasePOI().
			WithName(fmt.Sprintf("Benchmark POI %d", i)).
			WithMapID("benchmark-map").
			Build()
		err := testDB.SeedFixtures(poi)
		if err != nil {
			b.Fatalf("Failed to seed POI: %v", err)
		}
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pois, err := repo.GetByMapID(ctx, "benchmark-map")
		if err != nil {
			b.Fatalf("Failed to get POIs: %v", err)
		}
		if len(pois) != 100 {
			b.Fatalf("Expected 100 POIs, got %d", len(pois))
		}
	}
}