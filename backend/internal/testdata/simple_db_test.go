package testdata

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDatabaseFixtures_POI(t *testing.T) {
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

func TestDatabaseFixtures_Session(t *testing.T) {
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

func TestDefaultTestDBConfig(t *testing.T) {
	config := DefaultTestDBConfig()
	
	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, "5432", config.Port)
	assert.Equal(t, "postgres", config.User)
	assert.Equal(t, "postgres", config.Password)
	assert.Equal(t, "disable", config.SSLMode)
}