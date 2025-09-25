package models

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestMap_Validate(t *testing.T) {
	tests := []struct {
		name    string
		mapData Map
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid map",
			mapData: Map{
				ID:          "map-123",
				Name:        "Workshop Map 1",
				Description: "A map for the morning workshop",
				CreatedBy:   "facilitator-456",
				IsActive:    true,
				CreatedAt:   time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid map without description",
			mapData: Map{
				ID:        "map-123",
				Name:      "Workshop Map 1",
				CreatedBy: "facilitator-456",
				IsActive:  true,
				CreatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty map ID",
			mapData: Map{
				ID:        "",
				Name:      "Workshop Map 1",
				CreatedBy: "facilitator-456",
				IsActive:  true,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "map ID is required",
		},
		{
			name: "empty name",
			mapData: Map{
				ID:        "map-123",
				Name:      "",
				CreatedBy: "facilitator-456",
				IsActive:  true,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "map name is required",
		},
		{
			name: "name too long",
			mapData: Map{
				ID:        "map-123",
				Name:      "This is a very long map name that exceeds the maximum allowed length for a map name which should be limited to 255 characters to ensure database compatibility and reasonable display in the user interface and this string is definitely longer than that limit and needs to be even longer to exceed 255 characters so I'm adding more text here to make sure it's over the limit",
				CreatedBy: "facilitator-456",
				IsActive:  true,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "map name must be 255 characters or less",
		},
		{
			name: "empty created by",
			mapData: Map{
				ID:        "map-123",
				Name:      "Workshop Map 1",
				CreatedBy: "",
				IsActive:  true,
				CreatedAt: time.Now(),
			},
			wantErr: true,
			errMsg:  "created by is required",
		},
		{
			name: "zero time created at",
			mapData: Map{
				ID:        "map-123",
				Name:      "Workshop Map 1",
				CreatedBy: "facilitator-456",
				IsActive:  true,
				CreatedAt: time.Time{},
			},
			wantErr: true,
			errMsg:  "created at is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.mapData.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewMap(t *testing.T) {
	name := "Test Workshop Map"
	description := "A test map for workshops"
	createdBy := "facilitator-123"

	mapData, err := NewMap(name, description, createdBy)

	assert.NoError(t, err)
	assert.NotEmpty(t, mapData.ID)
	assert.Equal(t, name, mapData.Name)
	assert.Equal(t, description, mapData.Description)
	assert.Equal(t, createdBy, mapData.CreatedBy)
	assert.True(t, mapData.IsActive)
	assert.WithinDuration(t, time.Now(), mapData.CreatedAt, time.Second)
}

func TestNewMap_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		mapName     string
		description string
		createdBy   string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "empty name",
			mapName:     "",
			description: "Test",
			createdBy:   "facilitator-123",
			wantErr:     true,
			errMsg:      "map name is required",
		},
		{
			name:        "empty created by",
			mapName:     "Test Map",
			description: "Test",
			createdBy:   "",
			wantErr:     true,
			errMsg:      "created by is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapData, err := NewMap(tt.mapName, tt.description, tt.createdBy)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, mapData)
				if err != nil {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, mapData)
			}
		})
	}
}

func TestMap_Deactivate(t *testing.T) {
	mapData := Map{
		IsActive: true,
	}

	mapData.Deactivate()

	assert.False(t, mapData.IsActive)
}

func TestMap_Activate(t *testing.T) {
	mapData := Map{
		IsActive: false,
	}

	mapData.Activate()

	assert.True(t, mapData.IsActive)
}

func TestMap_CanBeAccessedBy(t *testing.T) {
	tests := []struct {
		name     string
		mapData  Map
		userID   string
		expected bool
	}{
		{
			name: "active map can be accessed",
			mapData: Map{
				ID:       "map-123",
				IsActive: true,
			},
			userID:   "user-456",
			expected: true,
		},
		{
			name: "inactive map cannot be accessed",
			mapData: Map{
				ID:       "map-123",
				IsActive: false,
			},
			userID:   "user-456",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mapData.CanBeAccessedBy(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMap_CanBeModifiedBy(t *testing.T) {
	tests := []struct {
		name     string
		mapData  Map
		user     *User
		expected bool
	}{
		{
			name: "nil user cannot modify",
			mapData: Map{
				ID:        "map-123",
				CreatedBy: "creator-456",
			},
			user:     nil,
			expected: false,
		},
		{
			name: "creator can modify their map",
			mapData: Map{
				ID:        "map-123",
				CreatedBy: "creator-456",
			},
			user: &User{
				ID:   "creator-456",
				Role: UserRoleUser,
			},
			expected: true,
		},
		{
			name: "non-creator regular user cannot modify",
			mapData: Map{
				ID:        "map-123",
				CreatedBy: "creator-456",
			},
			user: &User{
				ID:   "other-user-789",
				Role: UserRoleUser,
			},
			expected: false,
		},
		{
			name: "admin can modify any map",
			mapData: Map{
				ID:        "map-123",
				CreatedBy: "creator-456",
			},
			user: &User{
				ID:   "admin-789",
				Role: UserRoleAdmin,
			},
			expected: true,
		},
		{
			name: "superadmin can modify any map",
			mapData: Map{
				ID:        "map-123",
				CreatedBy: "creator-456",
			},
			user: &User{
				ID:   "superadmin-789",
				Role: UserRoleSuperAdmin,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mapData.CanBeModifiedBy(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMap_IsOwnedBy(t *testing.T) {
	mapData := Map{
		ID:        "map-123",
		CreatedBy: "creator-456",
	}

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{
			name:     "creator owns the map",
			userID:   "creator-456",
			expected: true,
		},
		{
			name:     "other user does not own the map",
			userID:   "other-user-789",
			expected: false,
		},
		{
			name:     "empty user ID does not own the map",
			userID:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapData.IsOwnedBy(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests using NewMap() builder for multi-map support relationships
func TestNewMap_WithBuilder(t *testing.T) {
	t.Run("create map with builder pattern", func(t *testing.T) {
		creatorID := uuid.New().String()
		creator := &User{
			ID:          creatorID,
			DisplayName: "Map Creator",
			Email:       stringPtr("creator@example.com"),
			Role:        UserRoleAdmin,
			AccountType: AccountTypeFull,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		mapData, err := NewMap("Workshop Map", "A collaborative workshop space", creatorID)
		assert.NoError(t, err)

		assert.Equal(t, "Workshop Map", mapData.Name)
		assert.Equal(t, "A collaborative workshop space", mapData.Description)
		assert.Equal(t, creator.ID, mapData.CreatedBy)
		assert.True(t, mapData.IsActive)
		assert.NotEmpty(t, mapData.ID)
	})

	t.Run("map ownership validation", func(t *testing.T) {
		creatorID := uuid.New().String()
		creator := &User{
			ID:          creatorID,
			DisplayName: "Map Owner",
			Email:       stringPtr("owner@example.com"),
			Role:        UserRoleUser,
			AccountType: AccountTypeFull,
			IsActive:    true,
		}

		mapData, err := NewMap("Private Map", "A private workspace", creatorID)
		assert.NoError(t, err)

		// Test ownership
		assert.True(t, mapData.IsOwnedBy(creator.ID))
		assert.False(t, mapData.IsOwnedBy("other-user-id"))

		// Test modification permissions
		assert.True(t, mapData.CanBeModifiedBy(creator))
	})

	t.Run("map access control validation", func(t *testing.T) {
		creatorID := uuid.New().String()
		creator := &User{
			ID:          creatorID,
			DisplayName: "Map Creator",
			Role:        UserRoleUser,
			AccountType: AccountTypeFull,
			IsActive:    true,
		}

		regularUserID := uuid.New().String()
		regularUser := &User{
			ID:          regularUserID,
			DisplayName: "Regular User",
			Role:        UserRoleUser,
			AccountType: AccountTypeFull,
			IsActive:    true,
		}

		adminUserID := uuid.New().String()
		adminUser := &User{
			ID:          adminUserID,
			DisplayName: "Admin User",
			Role:        UserRoleAdmin,
			AccountType: AccountTypeFull,
			IsActive:    true,
		}

		mapData, err := NewMap("Access Control Test Map", "Testing access control", creatorID)
		assert.NoError(t, err)

		// Active map can be accessed by anyone
		assert.True(t, mapData.CanBeAccessedBy(creator.ID))
		assert.True(t, mapData.CanBeAccessedBy(regularUser.ID))
		assert.True(t, mapData.CanBeAccessedBy(adminUser.ID))

		// Test modification permissions
		assert.True(t, mapData.CanBeModifiedBy(creator))      // Creator can modify
		assert.False(t, mapData.CanBeModifiedBy(regularUser)) // Regular user cannot
		assert.True(t, mapData.CanBeModifiedBy(adminUser))    // Admin can modify

		// Deactivated map cannot be accessed
		mapData.Deactivate()
		assert.False(t, mapData.CanBeAccessedBy(regularUser.ID))
	})
}

// Tests for multi-map isolation scenarios
func TestMap_UserIsolation(t *testing.T) {
	t.Run("users should be isolated between different maps", func(t *testing.T) {
		creator1ID := uuid.New().String()
		creator2ID := uuid.New().String()

		map1, err := NewMap("Map 1", "First map", creator1ID)
		assert.NoError(t, err)

		map2, err := NewMap("Map 2", "Second map", creator2ID)
		assert.NoError(t, err)

		// Each creator owns their respective map
		assert.True(t, map1.IsOwnedBy(creator1ID))
		assert.False(t, map1.IsOwnedBy(creator2ID))
		assert.True(t, map2.IsOwnedBy(creator2ID))
		assert.False(t, map2.IsOwnedBy(creator1ID))

		// Maps should have different IDs
		assert.NotEqual(t, map1.ID, map2.ID)
	})

	t.Run("session and POI isolation between maps", func(t *testing.T) {
		userID := uuid.New().String()

		map1, err := NewMap("Map 1", "First map", userID)
		assert.NoError(t, err)

		map2, err := NewMap("Map 2", "Second map", userID)
		assert.NoError(t, err)

		// Create sessions for different maps
		position := LatLng{Lat: 40.7128, Lng: -74.0060}
		session1, err := NewSession(userID, map1.ID, position)
		assert.NoError(t, err)

		session2, err := NewSession(userID, map2.ID, position)
		assert.NoError(t, err)

		// Sessions should be associated with different maps
		assert.Equal(t, map1.ID, session1.MapID)
		assert.Equal(t, map2.ID, session2.MapID)
		assert.NotEqual(t, session1.MapID, session2.MapID)

		// Create POIs for different maps
		poi1, err := NewPOI(map1.ID, "POI on Map 1", "First POI", position, userID)
		assert.NoError(t, err)

		poi2, err := NewPOI(map2.ID, "POI on Map 2", "Second POI", position, userID)
		assert.NoError(t, err)

		// POIs should be associated with different maps
		assert.Equal(t, map1.ID, poi1.MapID)
		assert.Equal(t, map2.ID, poi2.MapID)
		assert.NotEqual(t, poi1.MapID, poi2.MapID)
	})
}