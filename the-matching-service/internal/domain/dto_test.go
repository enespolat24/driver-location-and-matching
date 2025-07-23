package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMatchRequest_CreateRider(t *testing.T) {
	req := &MatchRequest{
		Name:     "Enes",
		Surname:  "Polat",
		Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
		Radius:   500,
	}

	userID := "user-123"
	rider := req.CreateRider(userID)

	assert.Equal(t, userID, rider.ID)
	assert.Equal(t, req.Name, rider.Name)
	assert.Equal(t, req.Surname, rider.Surname)
	assert.Equal(t, req.Location, rider.Location)
}

func TestNewMatchResponse(t *testing.T) {
	result := &MatchResult{
		RiderID:  "rider-123",
		DriverID: "driver-456",
		Distance: 150.5,
	}

	response := NewMatchResponse(result)

	assert.Equal(t, result.DriverID, response.Driver.ID)
	assert.Equal(t, result.Distance, response.Distance)
}

func TestMatchRequest_JSONTags(t *testing.T) {
	req := &MatchRequest{
		Name:     "Test",
		Surname:  "User",
		Location: Location{Type: "Point", Coordinates: [2]float64{0, 0}},
		Radius:   100,
	}

	_, err := json.Marshal(req)
	assert.NoError(t, err)
}

func TestMatchResponse_JSONTags(t *testing.T) {
	response := &MatchResponse{
		Driver:   Driver{ID: "driver-123"},
		Distance: 200.0,
	}

	_, err := json.Marshal(response)
	assert.NoError(t, err)
}

func TestDriverDistancePair_JSONTags(t *testing.T) {
	pair := &DriverDistancePair{
		Driver:   Driver{ID: "driver-123"},
		Distance: 150.5,
	}

	_, err := json.Marshal(pair)
	assert.NoError(t, err)
}

func TestDriverLocationServiceResponse_JSONTags(t *testing.T) {
	response := &DriverLocationServiceResponse{
		Count: 2,
		Drivers: []DriverDistancePair{
			{Driver: Driver{ID: "driver-1"}, Distance: 100.0},
			{Driver: Driver{ID: "driver-2"}, Distance: 200.0},
		},
	}

	_, err := json.Marshal(response)
	assert.NoError(t, err)
}
