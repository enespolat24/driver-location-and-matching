package domain

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestNewRider tests the NewRider function.
// Expected: Should create a Rider with correct ID and location.
func TestNewRider(t *testing.T) {
	location := Location{
		Type:        "Point",
		Coordinates: [2]float64{28.9, 41.0},
	}

	rider := NewRider("user-123", location)

	assert.Equal(t, "user-123", rider.ID)
	assert.Equal(t, location, rider.Location)
}

// TestRider_JSONTags tests JSON marshaling of Rider.
// Expected: Should marshal without error and contain correct fields.
func TestRider_JSONTags(t *testing.T) {
	rider := &Rider{
		ID:       "user-123",
		Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
	}

	data, err := json.Marshal(rider)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "user-123")
}

// TestLocation_JSONTags tests JSON marshaling of Location.
// Expected: Should marshal without error and contain correct fields.
func TestLocation_JSONTags(t *testing.T) {
	location := &Location{
		Type:        "Point",
		Coordinates: [2]float64{28.9, 41.0},
	}

	data, err := json.Marshal(location)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "Point")
	assert.Contains(t, string(data), "28.9")
	assert.Contains(t, string(data), "41")
}

// TestLocation_UnmarshalJSON tests JSON unmarshaling of Location.
// Expected: Should unmarshal without error and match expected values.
func TestLocation_UnmarshalJSON(t *testing.T) {
	jsonData := `{"type":"Point","coordinates":[28.9,41.0]}`
	var location Location

	err := json.Unmarshal([]byte(jsonData), &location)
	assert.NoError(t, err)
	assert.Equal(t, "Point", location.Type)
	assert.Equal(t, [2]float64{28.9, 41.0}, location.Coordinates)
}

// TestDriver_JSONTags tests JSON marshaling of Driver.
// Expected: Should marshal without error and contain correct fields.
func TestDriver_JSONTags(t *testing.T) {
	driver := &Driver{
		ID:        "driver-123",
		Location:  Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
		CreatedAt: "2023-01-01T00:00:00Z",
		UpdatedAt: "2023-01-01T00:00:00Z",
	}

	data, err := json.Marshal(driver)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "driver-123")
	assert.Contains(t, string(data), "2023-01-01T00:00:00Z")
}

// TestDriverWithDistance_JSONTags tests JSON marshaling of DriverWithDistance.
// Expected: Should marshal without error and contain correct fields.
func TestDriverWithDistance_JSONTags(t *testing.T) {
	driverWithDistance := &DriverWithDistance{
		Driver:   Driver{ID: "driver-123"},
		Distance: 150.5,
	}

	data, err := json.Marshal(driverWithDistance)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "driver-123")
	assert.Contains(t, string(data), "150.5")
}

// TestMatchResult_JSONTags tests JSON marshaling of MatchResult.
// Expected: Should marshal without error and contain correct fields.
func TestMatchResult_JSONTags(t *testing.T) {
	matchResult := &MatchResult{
		RiderID:  "rider-123",
		DriverID: "driver-456",
		Distance: 200.5,
	}

	data, err := json.Marshal(matchResult)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "rider-123")
	assert.Contains(t, string(data), "driver-456")
	assert.Contains(t, string(data), "200.5")
}

// TestMatchResult_UnmarshalJSON tests JSON unmarshaling of MatchResult.
// Expected: Should unmarshal without error and match expected values.
func TestMatchResult_UnmarshalJSON(t *testing.T) {
	jsonData := `{"rider_id":"rider-123","driver_id":"driver-456","distance":200.5}`
	var matchResult MatchResult

	err := json.Unmarshal([]byte(jsonData), &matchResult)
	assert.NoError(t, err)
	assert.Equal(t, "rider-123", matchResult.RiderID)
	assert.Equal(t, "driver-456", matchResult.DriverID)
	assert.Equal(t, 200.5, matchResult.Distance)
}

// TestLocation_Validation tests basic longitude and latitude validation logic.
// Expected: Should correctly validate valid and invalid coordinates.
func TestLocation_Validation(t *testing.T) {
	tests := []struct {
		name        string
		location    Location
		shouldValid bool
	}{
		{
			name: "valid Point location",
			location: Location{
				Type:        "Point",
				Coordinates: [2]float64{28.9, 41.0},
			},
			shouldValid: true,
		},
		{
			name: "valid coordinates range",
			location: Location{
				Type:        "Point",
				Coordinates: [2]float64{-180.0, 90.0},
			},
			shouldValid: true,
		},
		{
			name: "invalid longitude",
			location: Location{
				Type:        "Point",
				Coordinates: [2]float64{181.0, 41.0},
			},
			shouldValid: false,
		},
		{
			name: "invalid latitude",
			location: Location{
				Type:        "Point",
				Coordinates: [2]float64{28.9, 91.0},
			},
			shouldValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Basic validation - longitude should be between -180 and 180
			// latitude should be between -90 and 90
			isValid := tt.location.Coordinates[0] >= -180 && tt.location.Coordinates[0] <= 180 &&
				tt.location.Coordinates[1] >= -90 && tt.location.Coordinates[1] <= 90

			assert.Equal(t, tt.shouldValid, isValid)
		})
	}
}

// TestRider_Equality tests equality logic for Rider struct.
// Expected: Should correctly compare Rider IDs and locations.
func TestRider_Equality(t *testing.T) {
	location1 := Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	location2 := Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}

	rider1 := NewRider("user-123", location1)
	rider2 := NewRider("user-123", location2)
	rider3 := NewRider("user-456", location1)

	assert.Equal(t, rider1.ID, rider2.ID)
	assert.Equal(t, rider1.Location, rider2.Location)

	assert.NotEqual(t, rider1.ID, rider3.ID)
}
