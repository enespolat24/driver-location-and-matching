package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestMatchRequest_CreateRider tests the CreateRider method of MatchRequest.
// Expected: Should create a Rider with correct ID and location.
func TestMatchRequest_CreateRider(t *testing.T) {
	req := &MatchRequest{
		Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
		Radius:   500,
	}

	userID := "user-123"
	rider := req.CreateRider(userID)

	assert.Equal(t, userID, rider.ID)
	assert.Equal(t, req.Location, rider.Location)
}

// TestNewMatchResponse tests the NewMatchResponse function.
// Expected: Should create a MatchResponse with correct driver, rider, and distance.
func TestNewMatchResponse(t *testing.T) {
	result := &MatchResult{
		RiderID:  "rider-123",
		DriverID: "driver-456",
		Distance: 150.5,
	}

	response := NewMatchResponse(result)

	assert.Equal(t, result.DriverID, response.Driver)
	assert.Equal(t, result.RiderID, response.Rider)
	assert.Equal(t, result.Distance, response.Distance)
}

// TestMatchRequest_JSONTags tests JSON marshaling of MatchRequest.
// Expected: Should marshal without error.
func TestMatchRequest_JSONTags(t *testing.T) {
	req := &MatchRequest{
		Location: Location{Type: "Point", Coordinates: [2]float64{0, 0}},
		Radius:   100,
	}

	_, err := json.Marshal(req)
	assert.NoError(t, err)
}

// TestMatchResponse_JSONTags tests JSON marshaling of MatchResponse.
// Expected: Should marshal without error.
func TestMatchResponse_JSONTags(t *testing.T) {
	response := &MatchResponse{
		Driver:   "driver-123",
		Rider:    "rider-456",
		Distance: 200.0,
	}

	_, err := json.Marshal(response)
	assert.NoError(t, err)
}

// TestDriverDistancePair_JSONTags tests JSON marshaling of DriverDistancePair.
// Expected: Should marshal without error.
func TestDriverDistancePair_JSONTags(t *testing.T) {
	pair := &DriverDistancePair{
		Driver:   Driver{ID: "driver-123"},
		Distance: 150.5,
	}

	_, err := json.Marshal(pair)
	assert.NoError(t, err)
}

// TestDriverLocationServiceResponse_JSONTags tests JSON marshaling of DriverLocationServiceResponse.
// Expected: Should marshal without error.
func TestDriverLocationServiceResponse_JSONTags(t *testing.T) {
	response := &DriverLocationServiceResponse{
		Success: true,
		Data: map[string]interface{}{
			"count": 2,
			"drivers": []DriverDistancePair{
				{Driver: Driver{ID: "driver-1"}, Distance: 100.0},
				{Driver: Driver{ID: "driver-2"}, Distance: 200.0},
			},
		},
		Message: "Nearby drivers retrieved successfully",
	}

	_, err := json.Marshal(response)
	assert.NoError(t, err)
}

// TestDriverSearchData_JSONTags tests JSON marshaling of DriverSearchData.
// Expected: Should marshal without error.
func TestDriverSearchData_JSONTags(t *testing.T) {
	data := &DriverSearchData{
		Count: 2,
		Drivers: []DriverDistancePair{
			{Driver: Driver{ID: "driver-1"}, Distance: 100.0},
			{Driver: Driver{ID: "driver-2"}, Distance: 200.0},
		},
	}

	_, err := json.Marshal(data)
	assert.NoError(t, err)
}
