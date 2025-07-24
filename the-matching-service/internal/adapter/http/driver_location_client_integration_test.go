package httpadapter

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"the-matching-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

// TestDriverLocationClient_FindNearbyDrivers_integration tests successful integration with driver location service
// Expected: Should successfully find nearby drivers and return driver list with distances
func TestDriverLocationClient_FindNearbyDrivers_integration(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/drivers/search", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-api-key", r.Header.Get("X-API-Key"))

		var req map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&req)
		assert.NotNil(t, req["location"])
		assert.NotNil(t, req["radius"])

		resp := domain.DriverLocationServiceResponse{
			Success: true,
			Data: map[string]interface{}{
				"count": 1,
				"drivers": []domain.DriverDistancePair{
					{
						Driver:   domain.Driver{ID: "driver-1"},
						Distance: 100.0,
					},
				},
			},
			Message: "Nearby drivers retrieved successfully",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	client := NewDriverLocationClient(ts.URL, "test-api-key")
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, "driver-1", result[0].Driver.ID)
	assert.Equal(t, 100.0, result[0].Distance)
}
