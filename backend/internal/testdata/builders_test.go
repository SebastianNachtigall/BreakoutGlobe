package testdata

import (
	"testing"

	"breakoutglobe/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPOIBuilder(t *testing.T) {
	t.Run("creates POI with defaults", func(t *testing.T) {
		poi := NewPOI().Build()

		assert.NotEmpty(t, poi.ID)
		assert.NotEmpty(t, poi.MapID)
		assert.Equal(t, "Test POI", poi.Name)
		assert.Equal(t, "Test Description", poi.Description)
		assert.Equal(t, 40.7128, poi.Position.Lat)
		assert.Equal(t, -74.0060, poi.Position.Lng)
		assert.NotEmpty(t, poi.CreatedBy)
		assert.Equal(t, 10, poi.MaxParticipants)
		assert.False(t, poi.CreatedAt.IsZero())
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		creatorID := GenerateUUID()
		mapID := GenerateUUID()
		
		poi := NewPOI().
			WithID("custom-id").
			WithName("Coffee Shop").
			WithDescription("Great coffee place").
			WithPosition(models.LatLng{Lat: 37.7749, Lng: -122.4194}).
			WithCreator(creatorID).
			WithMap(mapID).
			WithMaxParticipants(20).
			Build()

		assert.Equal(t, "custom-id", poi.ID)
		assert.Equal(t, "Coffee Shop", poi.Name)
		assert.Equal(t, "Great coffee place", poi.Description)
		assert.Equal(t, 37.7749, poi.Position.Lat)
		assert.Equal(t, -122.4194, poi.Position.Lng)
		assert.Equal(t, creatorID.String(), poi.CreatedBy)
		assert.Equal(t, mapID.String(), poi.MapID)
		assert.Equal(t, 20, poi.MaxParticipants)
	})

	t.Run("generates valid UUIDs by default", func(t *testing.T) {
		poi := NewPOI().Build()

		_, err := uuid.Parse(poi.ID)
		assert.NoError(t, err)
		
		_, err = uuid.Parse(poi.MapID)
		assert.NoError(t, err)
		
		_, err = uuid.Parse(poi.CreatedBy)
		assert.NoError(t, err)
	})

	t.Run("creates valid POI that passes validation", func(t *testing.T) {
		poi := NewPOI().Build()
		
		err := poi.Validate()
		assert.NoError(t, err)
	})
}

func TestSessionBuilder(t *testing.T) {
	t.Run("creates session with defaults", func(t *testing.T) {
		session := NewSession().Build()

		assert.NotEmpty(t, session.ID)
		assert.NotEmpty(t, session.UserID)
		assert.NotEmpty(t, session.MapID)
		assert.Equal(t, 40.7128, session.AvatarPos.Lat)
		assert.Equal(t, -74.0060, session.AvatarPos.Lng)
		assert.True(t, session.IsActive)
		assert.False(t, session.CreatedAt.IsZero())
		assert.False(t, session.LastActive.IsZero())
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		userID := GenerateUUID()
		mapID := GenerateUUID()
		position := models.LatLng{Lat: 37.7749, Lng: -122.4194}
		
		session := NewSession().
			WithID("custom-session-id").
			WithUser(userID).
			WithMap(mapID).
			WithPosition(position).
			WithActive(false).
			Build()

		assert.Equal(t, "custom-session-id", session.ID)
		assert.Equal(t, userID.String(), session.UserID)
		assert.Equal(t, mapID.String(), session.MapID)
		assert.Equal(t, position.Lat, session.AvatarPos.Lat)
		assert.Equal(t, position.Lng, session.AvatarPos.Lng)
		assert.False(t, session.IsActive)
	})

	t.Run("generates valid UUIDs by default", func(t *testing.T) {
		session := NewSession().Build()

		_, err := uuid.Parse(session.ID)
		assert.NoError(t, err)
		
		_, err = uuid.Parse(session.UserID)
		assert.NoError(t, err)
		
		_, err = uuid.Parse(session.MapID)
		assert.NoError(t, err)
	})

	t.Run("creates valid session that passes validation", func(t *testing.T) {
		session := NewSession().Build()
		
		err := session.Validate()
		assert.NoError(t, err)
	})
}

func TestMapBuilder(t *testing.T) {
	t.Run("creates map with defaults", func(t *testing.T) {
		mapData := NewMap().Build()

		assert.NotEmpty(t, mapData.ID)
		assert.Equal(t, "Test Map", mapData.Name)
		assert.Equal(t, "Test map description", mapData.Description)
		assert.NotEmpty(t, mapData.CreatedBy)
		assert.True(t, mapData.IsActive)
		assert.False(t, mapData.CreatedAt.IsZero())
		assert.False(t, mapData.UpdatedAt.IsZero())
	})

	t.Run("allows customization with fluent API", func(t *testing.T) {
		creatorID := GenerateUUID()
		
		mapData := NewMap().
			WithID("custom-map-id").
			WithName("Adventure Map").
			WithDescription("Epic adventure awaits").
			WithCreator(creatorID).
			WithActive(false).
			Build()

		assert.Equal(t, "custom-map-id", mapData.ID)
		assert.Equal(t, "Adventure Map", mapData.Name)
		assert.Equal(t, "Epic adventure awaits", mapData.Description)
		assert.Equal(t, creatorID.String(), mapData.CreatedBy)
		assert.False(t, mapData.IsActive)
	})

	t.Run("generates valid UUIDs by default", func(t *testing.T) {
		mapData := NewMap().Build()

		_, err := uuid.Parse(mapData.ID)
		assert.NoError(t, err)
		
		_, err = uuid.Parse(mapData.CreatedBy)
		assert.NoError(t, err)
	})

	t.Run("creates valid map that passes validation", func(t *testing.T) {
		mapData := NewMap().Build()
		
		err := mapData.Validate()
		assert.NoError(t, err)
	})
}

func TestUUIDUtilities(t *testing.T) {
	t.Run("GenerateUUID creates valid UUID", func(t *testing.T) {
		id := GenerateUUID()
		
		assert.NotEqual(t, uuid.Nil, id)
		
		// Verify it's a valid UUID by parsing it
		parsed, err := uuid.Parse(id.String())
		require.NoError(t, err)
		assert.Equal(t, id, parsed)
	})

	t.Run("ParseUUID parses valid UUID string", func(t *testing.T) {
		original := GenerateUUID()
		
		parsed := ParseUUID(original.String())
		
		assert.Equal(t, original, parsed)
	})

	t.Run("ParseUUID panics on invalid UUID string", func(t *testing.T) {
		assert.Panics(t, func() {
			ParseUUID("invalid-uuid")
		})
	})

	t.Run("GenerateUUID creates unique UUIDs", func(t *testing.T) {
		id1 := GenerateUUID()
		id2 := GenerateUUID()
		
		assert.NotEqual(t, id1, id2)
	})
}

func TestBuilderRelationships(t *testing.T) {
	t.Run("can create related models with shared IDs", func(t *testing.T) {
		creatorID := GenerateUUID()
		mapID := GenerateUUID()
		
		// Create a map
		mapData := NewMap().
			WithID(mapID.String()).
			WithCreator(creatorID).
			Build()
		
		// Create a POI on that map by the same creator
		poi := NewPOI().
			WithMap(mapID).
			WithCreator(creatorID).
			Build()
		
		// Create a session for the creator on that map
		session := NewSession().
			WithUser(creatorID).
			WithMap(mapID).
			Build()
		
		// Verify relationships
		assert.Equal(t, mapData.ID, poi.MapID)
		assert.Equal(t, mapData.CreatedBy, poi.CreatedBy)
		assert.Equal(t, mapData.ID, session.MapID)
		assert.Equal(t, poi.CreatedBy, session.UserID)
	})
}