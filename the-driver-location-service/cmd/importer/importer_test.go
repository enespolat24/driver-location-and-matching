package main

import (
	"reflect"
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
