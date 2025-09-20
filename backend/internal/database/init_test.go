package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitialize(t *testing.T) {
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := Initialize(testURL)
	
	require.NoError(t, err)
	require.NotNil(t, db)
	
	// Verify connection works
	err = HealthCheck(db)
	assert.NoError(t, err)
	
	// Verify migrations ran
	status, err := GetMigrationStatus(db)
	require.NoError(t, err)
	
	assert.True(t, status["maps"])
	assert.True(t, status["sessions"])
	assert.True(t, status["pois"])
	
	// Clean up
	CloseConnection(db)
}

func TestInitialize_InvalidURL(t *testing.T) {
	db, err := Initialize("invalid://url")
	
	assert.Error(t, err)
	assert.Nil(t, db)
}

func TestInitialize_EmptyURL(t *testing.T) {
	db, err := Initialize("")
	
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "database URL is required")
}

func TestInitializeWithoutMigrations(t *testing.T) {
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := InitializeWithoutMigrations(testURL)
	
	require.NoError(t, err)
	require.NotNil(t, db)
	
	// Verify connection works
	err = HealthCheck(db)
	assert.NoError(t, err)
	
	// Clean up
	CloseConnection(db)
}

func TestMustInitialize(t *testing.T) {
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	// Should not panic with valid URL
	assert.NotPanics(t, func() {
		db := MustInitialize(testURL)
		assert.NotNil(t, db)
		CloseConnection(db)
	})
}

func TestMustInitialize_Panic(t *testing.T) {
	// Should panic with invalid URL
	assert.Panics(t, func() {
		MustInitialize("invalid://url")
	})
}