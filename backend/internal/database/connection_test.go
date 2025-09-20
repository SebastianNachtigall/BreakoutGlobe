package database

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewConnection(t *testing.T) {
	// Use test database URL
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	
	require.NoError(t, err)
	require.NotNil(t, db)
	
	// Test that we can ping the database
	sqlDB, err := db.DB()
	require.NoError(t, err)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err = sqlDB.PingContext(ctx)
	assert.NoError(t, err)
	
	// Clean up
	CloseConnection(db)
}

func TestNewConnection_InvalidURL(t *testing.T) {
	invalidURL := "invalid://database/url"
	
	db, err := NewConnection(invalidURL)
	
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "failed to connect to database")
}

func TestNewConnection_EmptyURL(t *testing.T) {
	db, err := NewConnection("")
	
	assert.Error(t, err)
	assert.Nil(t, db)
	assert.Contains(t, err.Error(), "database URL is required")
}

func TestCloseConnection(t *testing.T) {
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	
	// Should not panic or error
	err = CloseConnection(db)
	assert.NoError(t, err)
}

func TestCloseConnection_NilDB(t *testing.T) {
	// Should handle nil gracefully
	err := CloseConnection(nil)
	assert.NoError(t, err)
}

func TestConfigureConnection(t *testing.T) {
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CloseConnection(db)
	
	// Configure connection settings
	err = ConfigureConnection(db)
	assert.NoError(t, err)
	
	// Verify connection pool settings
	sqlDB, err := db.DB()
	require.NoError(t, err)
	
	stats := sqlDB.Stats()
	assert.Greater(t, stats.MaxOpenConnections, 0)
}

func TestHealthCheck(t *testing.T) {
	testURL := "postgres://postgres:postgres@localhost:5432/breakoutglobe_test?sslmode=disable"
	
	db, err := NewConnection(testURL)
	require.NoError(t, err)
	require.NotNil(t, db)
	defer CloseConnection(db)
	
	// Health check should pass
	err = HealthCheck(db)
	assert.NoError(t, err)
}

func TestHealthCheck_NilDB(t *testing.T) {
	err := HealthCheck(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection is nil")
}