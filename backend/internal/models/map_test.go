package models

import (
	"testing"
	"time"

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