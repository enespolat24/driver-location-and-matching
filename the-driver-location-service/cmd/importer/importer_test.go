package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"
)

// TestParseDriverLocation_Success tests parsing a valid record.
// Expected: Should parse the record correctly and return no error.
func TestParseDriverLocation_Success(t *testing.T) {
	record := []string{"41.12345", "29.98765"}
	driver, err := parseDriverLocation(record)
	if err != nil {
		t.Fatalf("Unexpected error: %v (should not error for valid record)", err)
	}

	expected := CreateDriverRequest{
		Location: Point{
			Type:        "Point",
			Coordinates: []float64{29.98765, 41.12345},
		},
	}

	if !reflect.DeepEqual(driver, expected) {
		t.Errorf("Expected: %+v, Got: %+v (should parse valid record correctly)", expected, driver)
	}
}

// TestParseDriverLocation_InvalidLength tests parsing a record with missing fields.
// Expected: Should return error when record has missing fields.
func TestParseDriverLocation_InvalidLength(t *testing.T) {
	record := []string{"41.12345"} // missing field
	_, err := parseDriverLocation(record)
	if err == nil {
		t.Error("Expected error: should return error when record has missing fields, but got nil")
	}
}

// TestParseDriverLocation_InvalidLatitude tests parsing a record with invalid latitude.
// Expected: Should return error when latitude is not a float.
func TestParseDriverLocation_InvalidLatitude(t *testing.T) {
	record := []string{"not-a-float", "29.98765"}
	_, err := parseDriverLocation(record)
	if err == nil {
		t.Error("Expected error: should return error when latitude is not a float, but got nil")
	}
}

// TestParseDriverLocation_InvalidLongitude tests parsing a record with invalid longitude.
// Expected: Should return error when longitude is not a float.
func TestParseDriverLocation_InvalidLongitude(t *testing.T) {
	record := []string{"41.12345", "not-a-float"}
	_, err := parseDriverLocation(record)
	if err == nil {
		t.Error("Expected error: should return error when longitude is not a float, but got nil")
	}
}

// TestProcessBatchHTTP_Success tests processBatchHTTP with a successful API response.
// Expected: Should return nil error when API returns 201 Created with success response.
func TestProcessBatchHTTP_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success": true, "data": {"count": 1, "drivers": [{"id": "test-driver"}]}, "message": "Drivers created successfully"}`))
	}))
	defer ts.Close()

	oldURL := apiURL
	apiURL = ts.URL
	defer func() { apiURL = oldURL }()

	batch := []CreateDriverRequest{{Location: Point{Type: "Point", Coordinates: []float64{1, 2}}}}
	err := processBatchHTTP(batch)
	if err != nil {
		t.Errorf("Expected nil error, got %v", err)
	}
}

// TestProcessBatchHTTP_APIError tests processBatchHTTP with a non-201 response.
// Expected: Should return error when API returns non-201 status and error should contain response details.
func TestProcessBatchHTTP_APIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"success": false, "error": "validation_error", "message": "Invalid request body"}`))
	}))
	defer ts.Close()

	oldURL := apiURL
	apiURL = ts.URL
	defer func() { apiURL = oldURL }()

	batch := []CreateDriverRequest{{Location: Point{Type: "Point", Coordinates: []float64{1, 2}}}}
	err := processBatchHTTP(batch)
	if err == nil {
		t.Error("Expected error for non-201 response, got nil")
	} else if !strings.Contains(err.Error(), "validation_error") {
		t.Errorf("Expected error to contain 'validation_error', got %v", err)
	}
}

// TestProcessBatchHTTP_HTTPError tests processBatchHTTP with an unreachable server.
// Expected: Should return error when server is unreachable.
func TestProcessBatchHTTP_HTTPError(t *testing.T) {
	oldURL := apiURL
	apiURL = "http://127.0.0.1:0" // invalid port
	defer func() { apiURL = oldURL }()

	batch := []CreateDriverRequest{{Location: Point{Type: "Point", Coordinates: []float64{1, 2}}}}
	err := processBatchHTTP(batch)
	if err == nil {
		t.Error("Expected error for unreachable server, got nil")
	}
}

// TestProcessBatchHTTP_APIServiceError tests processBatchHTTP when API returns success=false
// Expected: Should return error when API response indicates operation failure
func TestProcessBatchHTTP_APIServiceError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"success": false, "error": "service_error", "message": "Database connection failed"}`))
	}))
	defer ts.Close()

	oldURL := apiURL
	apiURL = ts.URL
	defer func() { apiURL = oldURL }()

	batch := []CreateDriverRequest{{Location: Point{Type: "Point", Coordinates: []float64{1, 2}}}}
	err := processBatchHTTP(batch)
	if err == nil {
		t.Error("Expected error for success=false response, got nil")
	} else if !strings.Contains(err.Error(), "service_error") {
		t.Errorf("Expected error to contain 'service_error', got %v", err)
	}
}

// TestProcessBatchHTTP_InvalidResponseJSON tests processBatchHTTP with invalid JSON response
// Expected: Should return error when API returns invalid JSON
func TestProcessBatchHTTP_InvalidResponseJSON(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`invalid-json`))
	}))
	defer ts.Close()

	oldURL := apiURL
	apiURL = ts.URL
	defer func() { apiURL = oldURL }()

	batch := []CreateDriverRequest{{Location: Point{Type: "Point", Coordinates: []float64{1, 2}}}}
	err := processBatchHTTP(batch)
	if err == nil {
		t.Error("Expected error for invalid JSON response, got nil")
	}
}

// TestGetenvOrDefault tests environment variable fallback logic
// Expected: Should return environment value when set, default when not set
func TestGetenvOrDefault(t *testing.T) {
	// Test default value when env var not set
	result := getenvOrDefault("NON_EXISTENT_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default', got %s", result)
	}

	// Test environment value when set
	os.Setenv("TEST_VAR", "env_value")
	defer os.Unsetenv("TEST_VAR")
	result = getenvOrDefault("TEST_VAR", "default")
	if result != "env_value" {
		t.Errorf("Expected 'env_value', got %s", result)
	}

	// Test empty environment value falls back to default
	os.Setenv("EMPTY_VAR", "")
	defer os.Unsetenv("EMPTY_VAR")
	result = getenvOrDefault("EMPTY_VAR", "default")
	if result != "default" {
		t.Errorf("Expected 'default' for empty env var, got %s", result)
	}
}
