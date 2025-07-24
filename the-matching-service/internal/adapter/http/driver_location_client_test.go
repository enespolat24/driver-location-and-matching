package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"the-matching-service/config"
	"the-matching-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

// TestDriverLocationClient_FindNearbyDrivers_error tests error handling when driver location service returns an error
// Expected: Should return error and nil result when service responds with HTTP 500
func TestDriverLocationClient_FindNearbyDrivers_error(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"success": false, "error": "internal_error", "message": "internal error"}`))
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	cfg := config.LoadConfig()
	client := NewDriverLocationClient(ts.URL, cfg.DriverLocationAPIKey)
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestDriverLocationClient_FindNearbyDrivers_invalidJSON tests handling of invalid JSON response from driver location service
// Expected: Should return error and nil result when service responds with malformed JSON
func TestDriverLocationClient_FindNearbyDrivers_invalidJSON(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not-a-json`))
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	cfg := config.LoadConfig()
	client := NewDriverLocationClient(ts.URL, cfg.DriverLocationAPIKey)
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestDriverLocationClient_FindNearbyDrivers_serviceError tests handling when service returns success=false
// Expected: Should return error and nil result when service returns success=false
func TestDriverLocationClient_FindNearbyDrivers_serviceError(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": false, "error": "validation_error", "message": "Invalid request"}`))
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	cfg := config.LoadConfig()
	client := NewDriverLocationClient(ts.URL, cfg.DriverLocationAPIKey)
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "validation_error")
}

// TestDriverLocationClient_FindNearbyDrivers_networkError tests network error handling when driver location service is unreachable
// Expected: Should return error and nil result when network connection fails
func TestDriverLocationClient_FindNearbyDrivers_networkError(t *testing.T) {
	cfg := config.LoadConfig()
	client := NewDriverLocationClient("http://127.0.0.1:0", cfg.DriverLocationAPIKey)
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestDriverLocationClient_FindNearbyDrivers_emptyList tests handling of empty driver list response from driver location service
// Expected: Should return empty slice and no error when service returns empty driver list
func TestDriverLocationClient_FindNearbyDrivers_emptyList(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true, "data": {"count": 0, "drivers": []}, "message": "Nearby drivers retrieved successfully"}`))
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	cfg := config.LoadConfig()
	client := NewDriverLocationClient(ts.URL, cfg.DriverLocationAPIKey)
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TestDriverLocationClient_FindNearbyDrivers_successWithDrivers tests successful response with drivers
// Expected: Should return drivers list when service returns successful response with drivers
func TestDriverLocationClient_FindNearbyDrivers_successWithDrivers(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		response := `{
			"success": true,
			"data": {
				"count": 1,
				"drivers": [
					{
						"driver": {
							"id": "driver-123",
							"location": {
								"type": "Point",
								"coordinates": [28.9, 41.0]
							}
						},
						"distance": 250.5
					}
				]
			},
			"message": "Nearby drivers retrieved successfully"
		}`
		w.Write([]byte(response))
	}

	ts := httptest.NewServer(http.HandlerFunc(mockHandler))
	defer ts.Close()

	cfg := config.LoadConfig()
	client := NewDriverLocationClient(ts.URL, cfg.DriverLocationAPIKey)
	location := domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}
	result, err := client.FindNearbyDrivers(context.Background(), location, 500)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "driver-123", result[0].Driver.ID)
	assert.Equal(t, 250.5, result[0].Distance)
}
