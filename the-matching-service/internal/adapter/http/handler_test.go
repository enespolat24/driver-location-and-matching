package httpadapter

import (
	"context"
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
