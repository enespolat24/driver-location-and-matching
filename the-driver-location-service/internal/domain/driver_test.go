package domain

import (
	"math"
	"testing"
)

// TestNewPointAndAccessors tests the NewPoint constructor and its accessors.
// Expected: Should return correct longitude and latitude values.
func TestNewPointAndAccessors(t *testing.T) {
	p := NewPoint(29.0, 41.0)
	if p.Longitude() != 29.0 {
		t.Errorf("Longitude should be 29.0, got %v", p.Longitude())
	}
	if p.Latitude() != 41.0 {
		t.Errorf("Latitude should be 41.0, got %v", p.Latitude())
	}
}

// TestHaversineDistance_IstanbulAnkara tests the HaversineDistance between Istanbul and Ankara.
// Expected: Should be approximately 351 km (within 15 km tolerance).
func TestHaversineDistance_IstanbulAnkara(t *testing.T) {
	// Distance between Istanbul and Ankara is approximately 351 km
	// i measured it with geojson.io
	istLat, istLon := 41.0082, 28.9784
	ankLat, ankLon := 39.9334, 32.8597
	dist := HaversineDistance(istLat, istLon, ankLat, ankLon)
	if math.Abs(dist-351000) > 15000 {
		t.Errorf("Istanbul-Ankara distance is not within expected range: got %v", dist)
	}
}

// TestPointDistance_SamePoint tests the distance between the same point.
// Expected: Should be less than 1 meter.
func TestPointDistance_SamePoint(t *testing.T) {
	p1 := NewPoint(29.0, 41.0)
	p2 := NewPoint(29.0, 41.0)
	dist := p1.Distance(p2)
	if dist > 1 {
		t.Errorf("Distance between the same point should be less than 1 meter, got %v", dist)
	}
}

// TestPointDistance_ZeroLongitudeLatitude tests the distance between (0,0) and (0,0).
// Expected: Should be less than 1 meter.
func TestPointDistance_ZeroLongitudeLatitude(t *testing.T) {
	p1 := NewPoint(0, 0)
	p2 := NewPoint(0, 0)
	dist := p1.Distance(p2)
	if dist > 1 {
		t.Errorf("Distance between (0,0)-(0,0) should be less than 1 meter, got %v", dist)
	}
}

// TestHaversineDistance_AntipodalPoints tests the HaversineDistance between antipodal points.
// Expected: Should be about 20015 km (within 100 km tolerance).
func TestHaversineDistance_AntipodalPoints(t *testing.T) {
	// Antipodal points: max possible distance on Earth (~20015 km)
	// https://en.m.wikipedia.org/wiki/Antipodal_point
	lat1, lon1 := 0.0, 0.0
	lat2, lon2 := 0.0, 180.0
	dist := HaversineDistance(lat1, lon1, lat2, lon2)
	if math.Abs(dist-20015000) > 100000 {
		t.Errorf("Antipodal points should be about 20015 km apart, got %v", dist)
	}
}

// TestHaversineDistance_SmallDistance tests the HaversineDistance for a small difference in latitude.
// Expected: Should be about 111 meters (within 2 meters tolerance).
func TestHaversineDistance_SmallDistance(t *testing.T) {
	// Two points ~111 meters apart (1 arcsecond latitude difference)
	lat1, lon1 := 0.0, 0.0
	lat2, lon2 := 0.001, 0.0
	dist := HaversineDistance(lat1, lon1, lat2, lon2)
	if math.Abs(dist-111) > 2 {
		t.Errorf("1 arcsecond latitude should be about 111 meters, got %v", dist)
	}
}

// TestNewPoint_NegativeCoordinates tests NewPoint with negative coordinates.
// Expected: Should return correct negative longitude and latitude values.
func TestNewPoint_NegativeCoordinates(t *testing.T) {
	p := NewPoint(-74.0, -40.0)
	if p.Longitude() != -74.0 {
		t.Errorf("Longitude should be -74.0, got %v", p.Longitude())
	}
	if p.Latitude() != -40.0 {
		t.Errorf("Latitude should be -40.0, got %v", p.Latitude())
	}
}
