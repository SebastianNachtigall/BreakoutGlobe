package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPOI_Validate(t *testing.T) {
	tests := []struct {
		name    string
		poi     POI
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid POI",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Description:     "A great place to meet",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid POI without description",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty POI ID",
			poi: POI{
				ID:              "",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "POI ID is required",
		},
		{
			name: "empty map ID",
			poi: POI{
				ID:              "poi-123",
				MapID:           "",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "map ID is required",
		},
		{
			name: "empty name",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "POI name is required",
		},
		{
			name: "name too long",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "This is a very long name that exceeds the maximum allowed length for a POI name which should be limited to 255 characters to ensure database compatibility and reasonable display in the user interface and this string is definitely longer than that limit and needs to be even longer to exceed 255 characters so I'm adding more text here to make sure it's over the limit",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "POI name must be 255 characters or less",
		},
		{
			name: "invalid position",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 91.0, Lng: -74.0060}, // Invalid latitude
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name: "empty created by",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "",
				MaxParticipants: 10,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "created by is required",
		},
		{
			name: "invalid max participants - zero",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 0,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "max participants must be between 1 and 50",
		},
		{
			name: "invalid max participants - too high",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 51,
				CreatedAt:       time.Now(),
			},
			wantErr: true,
			errMsg:  "max participants must be between 1 and 50",
		},
		{
			name: "zero time created at",
			poi: POI{
				ID:              "poi-123",
				MapID:           "map-789",
				Name:            "Meeting Room A",
				Position:        LatLng{Lat: 40.7128, Lng: -74.0060},
				CreatedBy:       "user-456",
				MaxParticipants: 10,
				CreatedAt:       time.Time{},
			},
			wantErr: true,
			errMsg:  "created at is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.poi.Validate()

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

func TestNewPOI(t *testing.T) {
	mapID := "map-456"
	name := "Test Meeting Room"
	description := "A test room"
	position := LatLng{Lat: 40.7128, Lng: -74.0060}
	createdBy := "user-123"

	poi, err := NewPOI(mapID, name, description, position, createdBy)

	assert.NoError(t, err)
	assert.NotEmpty(t, poi.ID)
	assert.Equal(t, mapID, poi.MapID)
	assert.Equal(t, name, poi.Name)
	assert.Equal(t, description, poi.Description)
	assert.Equal(t, position, poi.Position)
	assert.Equal(t, createdBy, poi.CreatedBy)
	assert.Equal(t, 10, poi.MaxParticipants) // Default value
	assert.WithinDuration(t, time.Now(), poi.CreatedAt, time.Second)
}

func TestNewPOI_InvalidInput(t *testing.T) {
	tests := []struct {
		name        string
		mapID       string
		poiName     string
		description string
		position    LatLng
		createdBy   string
		wantErr     bool
		errMsg      string
	}{
		{
			name:        "empty map ID",
			mapID:       "",
			poiName:     "Test Room",
			description: "Test",
			position:    LatLng{Lat: 40.7128, Lng: -74.0060},
			createdBy:   "user-123",
			wantErr:     true,
			errMsg:      "map ID is required",
		},
		{
			name:        "empty name",
			mapID:       "map-456",
			poiName:     "",
			description: "Test",
			position:    LatLng{Lat: 40.7128, Lng: -74.0060},
			createdBy:   "user-123",
			wantErr:     true,
			errMsg:      "POI name is required",
		},
		{
			name:        "invalid position",
			mapID:       "map-456",
			poiName:     "Test Room",
			description: "Test",
			position:    LatLng{Lat: 91.0, Lng: -74.0060},
			createdBy:   "user-123",
			wantErr:     true,
			errMsg:      "latitude must be between -90 and 90",
		},
		{
			name:        "empty created by",
			mapID:       "map-456",
			poiName:     "Test Room",
			description: "Test",
			position:    LatLng{Lat: 40.7128, Lng: -74.0060},
			createdBy:   "",
			wantErr:     true,
			errMsg:      "created by is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			poi, err := NewPOI(tt.mapID, tt.poiName, tt.description, tt.position, tt.createdBy)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, poi)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, poi)
			}
		})
	}
}

func TestPOI_DistanceTo(t *testing.T) {
	// New York City POI
	nycPOI := POI{
		Position: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	// Los Angeles position
	laPos := LatLng{Lat: 34.0522, Lng: -118.2437}

	distance := nycPOI.DistanceTo(laPos)

	// Distance between NYC and LA is approximately 3944 km
	assert.Greater(t, distance, 3900.0)
	assert.Less(t, distance, 4000.0)
}

func TestPOI_IsWithinRadius(t *testing.T) {
	poi := POI{
		Position: LatLng{Lat: 40.7128, Lng: -74.0060},
	}

	tests := []struct {
		name     string
		center   LatLng
		radius   float64
		expected bool
	}{
		{
			name:     "within radius - same point",
			center:   LatLng{Lat: 40.7128, Lng: -74.0060},
			radius:   1.0,
			expected: true,
		},
		{
			name:     "within radius - close point",
			center:   LatLng{Lat: 40.7130, Lng: -74.0062}, // Very close
			radius:   1.0,
			expected: true,
		},
		{
			name:     "outside radius - far point",
			center:   LatLng{Lat: 34.0522, Lng: -118.2437}, // Los Angeles
			radius:   1000.0,                                // 1000 km radius
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := poi.IsWithinRadius(tt.center, tt.radius)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Tests for new relationship functionality
func TestPOI_IsOwnedBy(t *testing.T) {
	creatorID := "user-123"
	poi := POI{
		CreatedBy: creatorID,
	}

	tests := []struct {
		name     string
		userID   string
		expected bool
	}{
		{
			name:     "POI is owned by creator",
			userID:   creatorID,
			expected: true,
		},
		{
			name:     "POI is not owned by other user",
			userID:   "other-user-456",
			expected: false,
		},
		{
			name:     "empty user ID does not own POI",
			userID:   "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := poi.IsOwnedBy(tt.userID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPOI_BelongsToMap(t *testing.T) {
	mapID := "map-123"
	poi := POI{
		MapID: mapID,
	}

	tests := []struct {
		name      string
		testMapID string
		expected  bool
	}{
		{
			name:      "POI belongs to map",
			testMapID: mapID,
			expected:  true,
		},
		{
			name:      "POI does not belong to other map",
			testMapID: "other-map-456",
			expected:  false,
		},
		{
			name:      "empty map ID does not match",
			testMapID: "",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := poi.BelongsToMap(tt.testMapID)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPOI_CanBeModifiedBy(t *testing.T) {
	creatorID := "user-123"
	poi := POI{
		CreatedBy: creatorID,
	}

	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name:     "nil user cannot modify",
			user:     nil,
			expected: false,
		},
		{
			name: "creator can modify their POI",
			user: &User{
				ID:   creatorID,
				Role: UserRoleUser,
			},
			expected: true,
		},
		{
			name: "other regular user cannot modify",
			user: &User{
				ID:   "other-user-456",
				Role: UserRoleUser,
			},
			expected: false,
		},
		{
			name: "admin can modify any POI",
			user: &User{
				ID:   "admin-789",
				Role: UserRoleAdmin,
			},
			expected: true,
		},
		{
			name: "superadmin can modify any POI",
			user: &User{
				ID:   "superadmin-999",
				Role: UserRoleSuperAdmin,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := poi.CanBeModifiedBy(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPOI_CanBeAccessedBy(t *testing.T) {
	poi := POI{
		CreatedBy: "user-123",
	}

	tests := []struct {
		name     string
		user     *User
		expected bool
	}{
		{
			name:     "nil user cannot access",
			user:     nil,
			expected: false,
		},
		{
			name: "any authenticated user can access POI",
			user: &User{
				ID:   "any-user-456",
				Role: UserRoleUser,
			},
			expected: true,
		},
		{
			name: "admin can access POI",
			user: &User{
				ID:   "admin-789",
				Role: UserRoleAdmin,
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := poi.CanBeAccessedBy(tt.user)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPOI_Update(t *testing.T) {
	poi := POI{
		UpdatedAt: time.Now().Add(-10 * time.Minute),
	}

	oldUpdatedAt := poi.UpdatedAt

	poi.Update()

	assert.True(t, poi.UpdatedAt.After(oldUpdatedAt))
	assert.WithinDuration(t, time.Now(), poi.UpdatedAt, time.Second)
}