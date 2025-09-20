package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLatLng_Validate(t *testing.T) {
	tests := []struct {
		name    string
		lat     float64
		lng     float64
		wantErr bool
		errMsg  string
	}{
		{
			name:    "valid coordinates",
			lat:     40.7128,
			lng:     -74.0060,
			wantErr: false,
		},
		{
			name:    "valid coordinates at boundaries",
			lat:     90.0,
			lng:     180.0,
			wantErr: false,
		},
		{
			name:    "valid coordinates at negative boundaries",
			lat:     -90.0,
			lng:     -180.0,
			wantErr: false,
		},
		{
			name:    "invalid latitude too high",
			lat:     91.0,
			lng:     0.0,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name:    "invalid latitude too low",
			lat:     -91.0,
			lng:     0.0,
			wantErr: true,
			errMsg:  "latitude must be between -90 and 90",
		},
		{
			name:    "invalid longitude too high",
			lat:     0.0,
			lng:     181.0,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
		{
			name:    "invalid longitude too low",
			lat:     0.0,
			lng:     -181.0,
			wantErr: true,
			errMsg:  "longitude must be between -180 and 180",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			latlng := LatLng{
				Lat: tt.lat,
				Lng: tt.lng,
			}

			err := latlng.Validate()

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLatLng_String(t *testing.T) {
	latlng := LatLng{
		Lat: 40.7128,
		Lng: -74.0060,
	}

	result := latlng.String()
	expected := "40.7128,-74.0060"

	assert.Equal(t, expected, result)
}

func TestLatLng_DistanceTo(t *testing.T) {
	// New York City
	nyc := LatLng{Lat: 40.7128, Lng: -74.0060}
	// Los Angeles
	la := LatLng{Lat: 34.0522, Lng: -118.2437}

	distance := nyc.DistanceTo(la)

	// Distance between NYC and LA is approximately 3944 km
	// Allow for some variance in calculation
	assert.Greater(t, distance, 3900.0)
	assert.Less(t, distance, 4000.0)
}

func TestLatLng_DistanceTo_SamePoint(t *testing.T) {
	point := LatLng{Lat: 40.7128, Lng: -74.0060}
	
	distance := point.DistanceTo(point)
	
	assert.Equal(t, 0.0, distance)
}