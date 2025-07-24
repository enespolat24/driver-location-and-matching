package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-matching-service/config"
	"the-matching-service/internal/application"
	"the-matching-service/internal/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

type mockDriverLocationService struct{}

func (m *mockDriverLocationService) FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
	return []domain.DriverDistancePair{
		{
			Driver:   domain.Driver{ID: "driver-1"},
			Distance: 100,
		},
	}, nil
}

// TestRouter_HealthAndMatchEndpoints tests the /health and /api/v1/match endpoints.
// Expected: /health returns 200 OK and 'healthy', /api/v1/match without JWT returns 401 Unauthorized.
func TestRouter_HealthAndMatchEndpoints(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	mockService := &mockDriverLocationService{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)
	router := NewRouter(handler, cfg)
	e := router.GetEcho()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")

	matchReq := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{}`))
	matchReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	matchW := httptest.NewRecorder()
	e.ServeHTTP(matchW, matchReq)
	assert.Equal(t, http.StatusUnauthorized, matchW.Code)
}
