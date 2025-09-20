package database

import (
	"testing"

	"breakoutglobe/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunMigrations(t *testing.T) {
	// Set up test database
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CloseConnection(db)
	
	// Drop all tables first to ensure clean state
	err = DropAllTables(db)
	require.NoError(t, err)
	
	// Run migrations
	err = RunMigrations(db)
	assert.NoError(t, err)
	
	// Verify tables were created
	assert.True(t, db.Migrator().HasTable(&models.Map{}))
	assert.True(t, db.Migrator().HasTable(&models.Session{}))
	assert.True(t, db.Migrator().HasTable(&models.POI{}))
	
	// Verify indexes were created
	assert.True(t, db.Migrator().HasIndex(&models.Session{}, "idx_sessions_user_id"))
	assert.True(t, db.Migrator().HasIndex(&models.Session{}, "idx_sessions_map_id"))
	assert.True(t, db.Migrator().HasIndex(&models.Session{}, "idx_sessions_last_active"))
	
	assert.True(t, db.Migrator().HasIndex(&models.POI{}, "idx_pois_map_id"))
	assert.True(t, db.Migrator().HasIndex(&models.POI{}, "idx_pois_created_by"))
	assert.True(t, db.Migrator().HasIndex(&models.POI{}, "idx_pois_position"))
	
	assert.True(t, db.Migrator().HasIndex(&models.Map{}, "idx_maps_created_by"))
}

func TestRunMigrations_NilDB(t *testing.T) {
	err := RunMigrations(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestDropAllTables(t *testing.T) {
	// Set up test database
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CloseConnection(db)
	
	// First run migrations to create tables
	err = RunMigrations(db)
	require.NoError(t, err)
	
	// Verify tables exist
	assert.True(t, db.Migrator().HasTable(&models.Map{}))
	assert.True(t, db.Migrator().HasTable(&models.Session{}))
	assert.True(t, db.Migrator().HasTable(&models.POI{}))
	
	// Drop all tables
	err = DropAllTables(db)
	assert.NoError(t, err)
	
	// Verify tables were dropped
	assert.False(t, db.Migrator().HasTable(&models.Map{}))
	assert.False(t, db.Migrator().HasTable(&models.Session{}))
	assert.False(t, db.Migrator().HasTable(&models.POI{}))
}

func TestDropAllTables_NilDB(t *testing.T) {
	err := DropAllTables(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestCreateIndexes(t *testing.T) {
	// Set up test database
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CloseConnection(db)
	
	// Drop and recreate tables without indexes
	err = DropAllTables(db)
	require.NoError(t, err)
	
	// Create tables without indexes
	err = db.AutoMigrate(&models.Map{}, &models.Session{}, &models.POI{})
	require.NoError(t, err)
	
	// Create indexes
	err = CreateIndexes(db)
	assert.NoError(t, err)
	
	// Verify indexes were created
	assert.True(t, db.Migrator().HasIndex(&models.Session{}, "idx_sessions_user_id"))
	assert.True(t, db.Migrator().HasIndex(&models.Session{}, "idx_sessions_map_id"))
	assert.True(t, db.Migrator().HasIndex(&models.Session{}, "idx_sessions_last_active"))
	
	assert.True(t, db.Migrator().HasIndex(&models.POI{}, "idx_pois_map_id"))
	assert.True(t, db.Migrator().HasIndex(&models.POI{}, "idx_pois_created_by"))
	assert.True(t, db.Migrator().HasIndex(&models.POI{}, "idx_pois_position"))
	
	assert.True(t, db.Migrator().HasIndex(&models.Map{}, "idx_maps_created_by"))
}

func TestCreateIndexes_NilDB(t *testing.T) {
	err := CreateIndexes(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}

func TestRollbackMigrations(t *testing.T) {
	// Set up test database
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CloseConnection(db)
	
	// First run migrations
	err = RunMigrations(db)
	require.NoError(t, err)
	
	// Verify tables exist
	assert.True(t, db.Migrator().HasTable(&models.Map{}))
	assert.True(t, db.Migrator().HasTable(&models.Session{}))
	assert.True(t, db.Migrator().HasTable(&models.POI{}))
	
	// Rollback migrations
	err = RollbackMigrations(db)
	assert.NoError(t, err)
	
	// Verify tables were dropped
	assert.False(t, db.Migrator().HasTable(&models.Map{}))
	assert.False(t, db.Migrator().HasTable(&models.Session{}))
	assert.False(t, db.Migrator().HasTable(&models.POI{}))
}

func TestRollbackMigrations_NilDB(t *testing.T) {
	err := RollbackMigrations(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}