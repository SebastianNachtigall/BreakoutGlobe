package models

import "fmt"

// Bounds represents a rectangular geographic area defined by north, south, east, and west coordinates
type Bounds struct {
	North float64 `json:"north"`
	South float64 `json:"south"`
	East  float64 `json:"east"`
	West  float64 `json:"west"`
}

// Validate checks if the bounds are valid
func (b Bounds) Validate() error {
	if b.North < -90 || b.North > 90 {
		return fmt.Errorf("north latitude must be between -90 and 90")
	}
	
	if b.South < -90 || b.South > 90 {
		return fmt.Errorf("south latitude must be between -90 and 90")
	}
	
	if b.East < -180 || b.East > 180 {
		return fmt.Errorf("east longitude must be between -180 and 180")
	}
	
	if b.West < -180 || b.West > 180 {
		return fmt.Errorf("west longitude must be between -180 and 180")
	}
	
	if b.North <= b.South {
		return fmt.Errorf("north latitude must be greater than south latitude")
	}
	
	// Handle longitude wrapping around 180/-180
	if b.West > b.East {
		// This is valid for bounds that cross the international date line
		// e.g., West=170, East=-170 (spans from 170 to -170 crossing 180/-180)
		return nil
	}
	
	return nil
}

// Contains checks if a given LatLng point is within these bounds
func (b Bounds) Contains(point LatLng) bool {
	if point.Lat < b.South || point.Lat > b.North {
		return false
	}
	
	// Handle longitude wrapping around 180/-180
	if b.West > b.East {
		// Bounds cross the international date line
		return point.Lng >= b.West || point.Lng <= b.East
	}
	
	return point.Lng >= b.West && point.Lng <= b.East
}

// Area calculates the approximate area of the bounds in square kilometers
func (b Bounds) Area() float64 {
	// Simple approximation - not accounting for Earth's curvature
	latDiff := b.North - b.South
	
	var lngDiff float64
	if b.West > b.East {
		// Bounds cross the international date line
		lngDiff = (180 - b.West) + (b.East - (-180))
	} else {
		lngDiff = b.East - b.West
	}
	
	// Convert to approximate kilometers (1 degree ≈ 111 km)
	return latDiff * lngDiff * 111 * 111
}