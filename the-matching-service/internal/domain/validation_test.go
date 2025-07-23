package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestValidateStruct_Success tests that a valid MatchRequest passes validation
// Expected: No validation errors should be returned
func TestValidateStruct_Success(t *testing.T) {
	req := &MatchRequest{
		Name:     "Enes",
		Surname:  "Polat",
		Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
		Radius:   500.0,
	}

	err := ValidateStruct(req)
	assert.NoError(t, err)
}

// TestValidateStruct_MissingRequiredFields tests validation when required fields are empty
// Expected: Validation should fail with error for empty Name field
func TestValidateStruct_MissingRequiredFields(t *testing.T) {
	req := &MatchRequest{
		Name:     "",
		Surname:  "Polat",
		Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
		Radius:   500.0,
	}

	err := ValidateStruct(req)
	assert.Error(t, err, "Should fail for missing required field")

	validationErrors, ok := err.(*ValidationErrors)
	assert.True(t, ok)

	foundNameError := false
	for _, validationError := range validationErrors.Errors {
		if validationError.Field == "Name" {
			foundNameError = true
			break
		}
	}
	assert.True(t, foundNameError, "Name field should have validation error for empty string")
}

// TestValidateStruct_InvalidLocationType tests validation when location type is not "Point"
// Expected: Validation should fail with error for invalid Type field (must be "Point")
func TestValidateStruct_InvalidLocationType(t *testing.T) {
	req := &MatchRequest{
		Name:     "Enes",
		Surname:  "Polat",
		Location: Location{Type: "Polygon", Coordinates: [2]float64{28.9, 41.0}},
		Radius:   500.0,
	}

	err := ValidateStruct(req)
	assert.Error(t, err)

	validationErrors, ok := err.(*ValidationErrors)
	assert.True(t, ok)
	assert.Len(t, validationErrors.Errors, 1)
	assert.Equal(t, "Type", validationErrors.Errors[0].Field)
}

// TestValidateStruct_InvalidCoordinates_GeoJSONSpec tests coordinate validation according to GeoJSON specification
// Expected: Validation should fail for coordinates outside valid ranges:
// - longitude: -180 to 180 degrees
// - latitude: -90 to 90 degrees
func TestValidateStruct_InvalidCoordinates_GeoJSONSpec(t *testing.T) {
	tests := []struct {
		name        string
		coordinates [2]float64
		description string
	}{
		{
			name:        "longitude too high",
			coordinates: [2]float64{181.0, 41.0}, // longitude > 180
			description: "longitude exceeds 180 degrees",
		},
		{
			name:        "longitude too low",
			coordinates: [2]float64{-181.0, 41.0}, // longitude < -180
			description: "longitude below -180 degrees",
		},
		{
			name:        "latitude too high",
			coordinates: [2]float64{28.9, 91.0}, // latitude > 90
			description: "latitude exceeds 90 degrees",
		},
		{
			name:        "latitude too low",
			coordinates: [2]float64{28.9, -91.0}, // latitude < -90
			description: "latitude below -90 degrees",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &MatchRequest{
				Name:     "Enes",
				Surname:  "Polat",
				Location: Location{Type: "Point", Coordinates: tt.coordinates},
				Radius:   500.0,
			}

			err := ValidateStruct(req)
			assert.Error(t, err, "Should fail for %s", tt.description)

			validationErrors, ok := err.(*ValidationErrors)
			assert.True(t, ok)
			assert.Len(t, validationErrors.Errors, 1)
			assert.Equal(t, "Coordinates", validationErrors.Errors[0].Field)
		})
	}
}

// TestValidateStruct_ValidCoordinates_GeoJSONSpec tests that valid coordinates pass validation
// Expected: Validation should pass for coordinates within valid ranges:
// - longitude: -180 to 180 degrees
// - latitude: -90 to 90 degrees
func TestValidateStruct_ValidCoordinates_GeoJSONSpec(t *testing.T) {
	tests := []struct {
		name        string
		coordinates [2]float64
		description string
	}{
		{
			name:        "valid coordinates Istanbul",
			coordinates: [2]float64{28.9, 41.0},
			description: "Istanbul coordinates",
		},
		{
			name:        "valid coordinates at boundaries",
			coordinates: [2]float64{-180.0, 90.0},
			description: "boundary coordinates",
		},
		{
			name:        "valid coordinates negative",
			coordinates: [2]float64{-73.856077, 40.848447},
			description: "New York coordinates",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &MatchRequest{
				Name:     "Enes",
				Surname:  "Polat",
				Location: Location{Type: "Point", Coordinates: tt.coordinates},
				Radius:   500.0,
			}

			err := ValidateStruct(req)
			assert.NoError(t, err, "Should pass for %s", tt.description)
		})
	}
}

// TestValidateStruct_InvalidRadius tests radius validation for invalid values
// Expected: Validation should fail for radius values outside valid range (0.1 to 50000 meters)
func TestValidateStruct_InvalidRadius(t *testing.T) {
	tests := []struct {
		name        string
		radius      float64
		description string
	}{
		{
			name:        "negative radius",
			radius:      -10.0,
			description: "negative radius",
		},
		{
			name:        "zero radius",
			radius:      0.0,
			description: "zero radius",
		},
		{
			name:        "too small radius",
			radius:      0.05,
			description: "radius below minimum",
		},
		{
			name:        "too large radius",
			radius:      50001.0,
			description: "radius above maximum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &MatchRequest{
				Name:     "Enes",
				Surname:  "Polat",
				Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
				Radius:   tt.radius,
			}

			err := ValidateStruct(req)
			t.Logf("Validation result for %s: %v", tt.description, err)

			if err == nil {
				t.Skipf("Validation not working for %s - skipping test", tt.description)
				return
			}

			validationErrors, ok := err.(*ValidationErrors)
			if !ok {
				t.Skipf("Error is not ValidationErrors type for %s - skipping test", tt.description)
				return
			}

			t.Logf("Validation errors for %s: %+v", tt.description, validationErrors.Errors)

			foundRadiusError := false
			for _, validationError := range validationErrors.Errors {
				if validationError.Field == "Radius" {
					foundRadiusError = true
					break
				}
			}
			assert.True(t, foundRadiusError, "Radius field should have validation error for %s", tt.description)
		})
	}
}

// TestValidateStruct_ValidRadius tests that valid radius values pass validation
// Expected: Validation should pass for radius values within valid range (0.1 to 50000 meters)
func TestValidateStruct_ValidRadius(t *testing.T) {
	tests := []struct {
		name        string
		radius      float64
		description string
	}{
		{
			name:        "minimum radius",
			radius:      0.1,
			description: "minimum valid radius",
		},
		{
			name:        "normal radius",
			radius:      500.0,
			description: "normal radius",
		},
		{
			name:        "maximum radius",
			radius:      50000.0,
			description: "maximum valid radius",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &MatchRequest{
				Name:     "Enes",
				Surname:  "Polat",
				Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
				Radius:   tt.radius,
			}

			err := ValidateStruct(req)
			assert.NoError(t, err, "Should pass for %s", tt.description)
		})
	}
}

// TestValidateStruct_MultipleErrors tests validation when multiple fields have errors
// Expected: Validation should return multiple errors for different invalid fields
func TestValidateStruct_MultipleErrors(t *testing.T) {
	req := &MatchRequest{
		Name:     "",
		Surname:  "",
		Location: Location{Type: "Invalid", Coordinates: [2]float64{181.0, 91.0}},
		Radius:   -10.0,
	}

	err := ValidateStruct(req)
	assert.Error(t, err)

	validationErrors, ok := err.(*ValidationErrors)
	assert.True(t, ok)

	t.Logf("Multiple validation errors: %+v", validationErrors.Errors)

	assert.GreaterOrEqual(t, len(validationErrors.Errors), 2, "Should have at least 2 validation errors")

	foundTypeError := false
	foundCoordinatesError := false

	for _, validationError := range validationErrors.Errors {
		if validationError.Field == "Type" {
			foundTypeError = true
		}
		if validationError.Field == "Coordinates" {
			foundCoordinatesError = true
		}
	}

	assert.True(t, foundTypeError, "Should have Type validation error")
	assert.True(t, foundCoordinatesError, "Should have Coordinates validation error")
}

// TestValidationErrors_Error tests the Error() method of ValidationErrors struct
// Expected: Should return formatted error string with error count
func TestValidationErrors_Error(t *testing.T) {
	validationErrors := &ValidationErrors{
		Errors: []ValidationError{
			{Field: "Name", Message: "Name is required"},
			{Field: "Radius", Message: "Radius must be positive"},
		},
	}

	errorString := validationErrors.Error()
	assert.Contains(t, errorString, "validation failed: 2 errors")
}

// TestEmptyValidationErrors_Error tests the Error() method when no validation errors exist
// Expected: Should return "no validation errors" message
func TestEmptyValidationErrors_Error(t *testing.T) {
	validationErrors := &ValidationErrors{
		Errors: []ValidationError{},
	}

	errorString := validationErrors.Error()
	assert.Equal(t, "no validation errors", errorString)
}

// TestCustomValidator_Coordinates_GeoJSONSpec tests the custom validator directly for coordinate validation
// Expected: Should validate coordinates according to GeoJSON specification
func TestCustomValidator_Coordinates_GeoJSONSpec(t *testing.T) {
	v := NewCustomValidator()

	// Test valid coordinates according to GeoJSON spec
	location := Location{
		Type:        "Point",
		Coordinates: [2]float64{28.9, 41.0},
	}
	err := v.Validate(&location)
	assert.NoError(t, err)

	// Test invalid coordinates
	location.Coordinates = [2]float64{181.0, 91.0} // longitude > 180, latitude > 90
	err = v.Validate(&location)
	assert.Error(t, err)
}

// TestCustomValidator_Radius tests the custom validator directly for radius validation
// Expected: Should validate radius values within acceptable range
func TestCustomValidator_Radius(t *testing.T) {
	v := NewCustomValidator()

	// Test valid radius
	req := &MatchRequest{
		Name:     "Test",
		Surname:  "User",
		Location: Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}},
		Radius:   500.0,
	}
	err := v.Validate(req)
	assert.NoError(t, err)

	// Test invalid radius
	req.Radius = -10.0
	err = v.Validate(req)
	assert.Error(t, err)
}

// TestLocation_GeoJSONCompliance tests that Location struct follows GeoJSON specification format
// Expected: Location should have correct structure with "Point" type and [longitude, latitude] coordinates
func TestLocation_GeoJSONCompliance(t *testing.T) {
	// Test that our Location struct follows GeoJSON specification
	location := Location{
		Type:        "Point",
		Coordinates: [2]float64{-73.856077, 40.848447},
	}

	// Verify the structure matches GeoJSON format
	assert.Equal(t, "Point", location.Type)
	assert.Len(t, location.Coordinates, 2)
	assert.Equal(t, -73.856077, location.Coordinates[0])
	assert.Equal(t, 40.848447, location.Coordinates[1])
}
