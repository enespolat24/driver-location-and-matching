package httpadapter

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-matching-service/internal/adapter/config"
	"the-matching-service/internal/adapter/middleware"
	"the-matching-service/internal/application"
	"the-matching-service/internal/domain"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type mockDriverLocationServiceForHandler struct{}

func (m *mockDriverLocationServiceForHandler) FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
	return []domain.DriverDistancePair{
		{
			Driver:   domain.Driver{ID: "driver-1"},
			Distance: 100,
		},
	}, nil
}

type mockDriverLocationServiceForHandlerNoDrivers struct{}

func (m *mockDriverLocationServiceForHandlerNoDrivers) FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
	return []domain.DriverDistancePair{}, nil
}

type mockDriverLocationServiceForHandlerError struct{}

func (m *mockDriverLocationServiceForHandlerError) FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
	return nil, errors.New("database connection failed")
}

func generateJWT(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(secret))
	return t
}

// TestMatchHandler_Success tests successful rider matching with valid request and authentication
// Expected: HTTP 200 OK with driver match response containing driver ID
func TestMatchHandler_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	mockService := &mockDriverLocationServiceForHandler{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{
		"name": "Enes",
		"surname": "Polat",
		"location": {"type": "Point", "coordinates": [28.9, 41.0]},
		"radius": 500
	}`))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "driver-1")
	assert.Contains(t, w.Body.String(), "user-1")
	assert.Contains(t, w.Body.String(), "distance")
}

// TestMatchHandler_ValidationError tests validation error handling with invalid request data
// Expected: HTTP 400 Bad Request with validation error message for invalid fields
func TestMatchHandler_ValidationError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	mockService := &mockDriverLocationServiceForHandler{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	// Test with invalid request (missing required fields)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{
		"name": "",
		"surname": "P",
		"location": {"type": "Invalid", "coordinates": [181.0, 91.0]},
		"radius": -10
	}`))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "validation_error")
	assert.Contains(t, w.Body.String(), "Request validation failed")
}

// TestMatchHandler_Unauthorized tests unauthorized access without authentication
// Expected: HTTP 401 Unauthorized when user is not authenticated
func TestMatchHandler_Unauthorized(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	mockService := &mockDriverLocationServiceForHandler{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	// Request without authentication token
	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{
		"name": "Enes",
		"surname": "Polat",
		"location": {"type": "Point", "coordinates": [28.9, 41.0]},
		"radius": 500
	}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "unauthorized")
}

// TestMatchHandler_InvalidRequestBody tests handling of invalid request body
// Expected: HTTP 400 Bad Request when request body cannot be parsed
func TestMatchHandler_InvalidRequestBody(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	mockService := &mockDriverLocationServiceForHandler{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	// Invalid JSON request body
	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{
		"name": "Enes",
		"surname": "Polat",
		"location": {"type": "Point", "coordinates": [28.9, 41.0]},
		"radius": 500,
	`)) // Missing closing brace
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "invalid_request")
}

// TestMatchHandler_InternalServerError tests handling of internal server errors
// Expected: HTTP 500 Internal Server Error when matching service returns unexpected error
func TestMatchHandler_InternalServerError(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}

	// Mock service that returns error
	mockService := &mockDriverLocationServiceForHandlerError{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{
		"name": "Enes",
		"surname": "Polat",
		"location": {"type": "Point", "coordinates": [28.9, 41.0]},
		"radius": 500
	}`))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "internal_error")
}

// TestMatchHandler_NoDriversFound tests the 404 response when no drivers are found nearby
// Expected: HTTP 404 Not Found with "No drivers found nearby" message when no drivers match criteria
func TestMatchHandler_NoDriversFound(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}

	// Mock service that returns no drivers
	mockService := &mockDriverLocationServiceForHandlerNoDrivers{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{
		"name": "Enes",
		"surname": "Polat",
		"location": {"type": "Point", "coordinates": [28.9, 41.0]},
		"radius": 500
	}`))
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	w := httptest.NewRecorder()

	e.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "not_found")
	assert.Contains(t, w.Body.String(), "No drivers found nearby")
}

// TestMatchHandler_GeoJSONPointSearch_NoDriversFound tests 404 response for GeoJSON point search with no matching drivers
// Expected: HTTP 404 Not Found when searching with valid GeoJSON Point coordinates but no drivers match criteria
func TestMatchHandler_GeoJSONPointSearch_NoDriversFound(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}

	mockService := &mockDriverLocationServiceForHandlerNoDrivers{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	testCases := []struct {
		name        string
		coordinates [2]float64
		radius      float64
		description string
	}{
		{
			name:        "Istanbul coordinates",
			coordinates: [2]float64{28.9, 41.0},
			radius:      500,
			description: "Istanbul area search",
		},
		{
			name:        "New York coordinates",
			coordinates: [2]float64{-73.856077, 40.848447},
			radius:      1000,
			description: "New York area search",
		},
		{
			name:        "London coordinates",
			coordinates: [2]float64{-0.1276, 51.5074},
			radius:      2000,
			description: "London area search",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := fmt.Sprintf(`{
				"name": "Test",
				"surname": "User",
				"location": {"type": "Point", "coordinates": [%f, %f]},
				"radius": %f
			}`, tc.coordinates[0], tc.coordinates[1], tc.radius)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(reqBody))
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			e.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotFound, w.Code, "Should return 404 for %s", tc.description)
			assert.Contains(t, w.Body.String(), "not_found", "Should contain not_found error for %s", tc.description)
			assert.Contains(t, w.Body.String(), "No drivers found nearby", "Should contain correct message for %s", tc.description)
		})
	}
}

// TestMatchHandler_GeoJSONPointSearch_Success tests successful driver matching with GeoJSON Point coordinates
// Expected: HTTP 200 OK with driver match when searching with valid GeoJSON Point coordinates and drivers are found
func TestMatchHandler_GeoJSONPointSearch_Success(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}

	// Mock service that returns drivers for GeoJSON point search
	mockService := &mockDriverLocationServiceForHandler{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)

	e := echo.New()
	e.Use(middleware.JWTAuthMiddleware(cfg))
	e.POST("/api/v1/match", handler.Match)

	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	testCases := []struct {
		name        string
		coordinates [2]float64
		radius      float64
		description string
	}{
		{
			name:        "Istanbul coordinates with drivers",
			coordinates: [2]float64{28.9, 41.0},
			radius:      500,
			description: "Istanbul area search with available drivers",
		},
		{
			name:        "New York coordinates with drivers",
			coordinates: [2]float64{-73.856077, 40.848447},
			radius:      1000,
			description: "New York area search with available drivers",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			reqBody := fmt.Sprintf(`{
				"name": "Test",
				"surname": "User",
				"location": {"type": "Point", "coordinates": [%f, %f]},
				"radius": %f
			}`, tc.coordinates[0], tc.coordinates[1], tc.radius)

			req := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(reqBody))
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			w := httptest.NewRecorder()

			e.ServeHTTP(w, req)

			assert.Equal(t, http.StatusOK, w.Code, "Should return 200 for %s", tc.description)
			assert.Contains(t, w.Body.String(), "driver-1", "Should contain driver ID for %s", tc.description)
			assert.Contains(t, w.Body.String(), "user-1", "Should contain rider ID for %s", tc.description)
			assert.Contains(t, w.Body.String(), "distance", "Should contain distance field for %s", tc.description)
		})
	}
}
