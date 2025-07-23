package domain

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

// CustomValidator holds the validator instance with custom validations
type CustomValidator struct {
	validator *validator.Validate
}

// NewCustomValidator creates a new validator with custom validation functions
func NewCustomValidator() *CustomValidator {
	v := validator.New()

	// Register custom validation functions
	v.RegisterValidation("coordinates", validateCoordinates)
	v.RegisterValidation("radius", validateRadius)

	return &CustomValidator{validator: v}
}

// Validate validates a struct using the custom validator
func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

// validateCoordinates validates that coordinates are within valid ranges
// According to GeoJSON spec: longitude: -180 to 180, latitude: -90 to 90
// Coordinates format: [longitude, latitude] (MongoDB standard)
func validateCoordinates(fl validator.FieldLevel) bool {
	coordinates := fl.Field().Interface().([2]float64)

	longitude := coordinates[0]
	latitude := coordinates[1]

	// Check longitude range (-180 to 180)
	if longitude < -180 || longitude > 180 {
		return false
	}

	// Check latitude range (-90 to 90)
	if latitude < -90 || latitude > 90 {
		return false
	}

	return true
}

// validateRadius validates that radius is within reasonable bounds
func validateRadius(fl validator.FieldLevel) bool {
	radius := fl.Field().Float()

	// Radius should be positive and reasonable (0.1 to 50000 meters)
	return radius >= 0.1 && radius <= 50000
}

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors struct {
	Errors []ValidationError `json:"errors"`
}

// ValidateStruct validates a struct and returns formatted validation errors
func ValidateStruct(s interface{}) error {
	v := NewCustomValidator()

	err := v.Validate(s)
	if err != nil {
		var validationErrors ValidationErrors

		for _, err := range err.(validator.ValidationErrors) {
			validationError := ValidationError{
				Field:   err.Field(),
				Message: getValidationMessage(err),
			}
			validationErrors.Errors = append(validationErrors.Errors, validationError)
		}

		return &validationErrors
	}

	return nil
}

// getValidationMessage returns a human-readable validation message
func getValidationMessage(err validator.FieldError) string {
	switch err.Tag() {
	case "required":
		return fmt.Sprintf("%s is required", err.Field())
	case "min":
		return fmt.Sprintf("%s must be at least %s", err.Field(), err.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", err.Field(), err.Param())
	case "eq":
		return fmt.Sprintf("%s must be equal to %s", err.Field(), err.Param())
	case "len":
		return fmt.Sprintf("%s must have length %s", err.Field(), err.Param())
	case "coordinates":
		return fmt.Sprintf("%s coordinates are invalid (longitude: -180 to 180, latitude: -90 to 90)", err.Field())
	case "radius":
		return fmt.Sprintf("%s must be between 0.1 and 50000 meters", err.Field())
	default:
		return fmt.Sprintf("%s failed validation: %s", err.Field(), err.Tag())
	}
}

// Error implements the error interface for ValidationErrors
func (ve *ValidationErrors) Error() string {
	if len(ve.Errors) == 0 {
		return "no validation errors"
	}

	return fmt.Sprintf("validation failed: %d errors", len(ve.Errors))
}
