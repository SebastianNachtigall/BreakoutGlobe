package models

import (
	"fmt"
	"math"
)

// LatLng represents a geographic coordinate with latitude and longitude
type LatLng struct {
	Lat float64 `json:"lat" gorm:"column:lat" validate:"required,min=-90,max=90"`
	Lng float64 `json:"lng" gorm:"column:lng" validate:"required,min=-180,max=180"`
}

// Validate checks if the LatLng coordinates are within valid bounds
func (ll LatLng) Validate() error {
	if ll.Lat < -90 || ll.Lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90")
	}
	if ll.Lng < -180 || ll.Lng > 180 {
		return fmt.Errorf("longitude must be between -180 and 180")
	}
	return nil
}

// String returns a string representation of the coordinates
func (ll LatLng) String() string {
	return fmt.Sprintf("%.4f,%.4f", ll.Lat, ll.Lng)
}

// DistanceTo calculates the distance in kilometers between two LatLng points
// using the Haversine formula
func (ll LatLng) DistanceTo(other LatLng) float64 {
	if ll.Lat == other.Lat && ll.Lng == other.Lng {
		return 0.0
	}

	const earthRadius = 6371.0 // Earth's radius in kilometers

	// Convert degrees to radians
	lat1Rad := ll.Lat * math.Pi / 180
	lng1Rad := ll.Lng * math.Pi / 180
	lat2Rad := other.Lat * math.Pi / 180
	lng2Rad := other.Lng * math.Pi / 180

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlng := lng2Rad - lng1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*
			math.Sin(dlng/2)*math.Sin(dlng/2)

	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}