package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPOI_WithImage_Success(t *testing.T) {
	// Test that POI can be created with an image URL
	position := LatLng{Lat: 40.7128, Lng: -74.0060}
	imageURL := "https://example.com/images/coffee-shop.jpg"
	
	poi, err := NewPOI("map-123", "Coffee Shop", "Great place to meet", position, "user-123")
	assert.NoError(t, err)
	
	// Set image URL
	poi.ImageURL = imageURL
	
	// Validate POI with image
	err = poi.Validate()
	assert.NoError(t, err)
	assert.Equal(t, imageURL, poi.ImageURL)
}

func TestPOI_WithoutImage_Success(t *testing.T) {
	// Test that POI can be created without an image (optional field)
	position := LatLng{Lat: 40.7128, Lng: -74.0060}
	
	poi, err := NewPOI("map-123", "Coffee Shop", "Great place to meet", position, "user-123")
	assert.NoError(t, err)
	
	// Image URL should be empty by default
	assert.Empty(t, poi.ImageURL)
	
	// Validate POI without image
	err = poi.Validate()
	assert.NoError(t, err)
}

func TestPOI_WithInvalidImageURL_Validation(t *testing.T) {
	// Test validation of image URL length
	position := LatLng{Lat: 40.7128, Lng: -74.0060}
	
	poi, err := NewPOI("map-123", "Coffee Shop", "Great place to meet", position, "user-123")
	assert.NoError(t, err)
	
	// Set very long image URL (over 500 characters)
	longURL := "https://example.com/" + string(make([]byte, 500)) + "image.jpg"
	poi.ImageURL = longURL
	
	// Validation should fail
	err = poi.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "image URL must be 500 characters or less")
}