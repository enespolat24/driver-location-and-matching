package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-matching-service/internal/adapter/config"
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

func TestRouter_HealthAndMatchEndpoints(t *testing.T) {
	cfg := &config.Config{JWTSecret: "testsecret"}
	mockService := &mockDriverLocationService{}
	matchingService := application.NewMatchingService(mockService)
	handler := NewMatchHandler(matchingService)
	router := NewRouter(handler, cfg)
	e := router.GetEcho()

	// Test /health endpoint (should not require JWT)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "healthy")

	// Test /api/v1/match endpoint without JWT (should return 401)
	matchReq := httptest.NewRequest(http.MethodPost, "/api/v1/match", strings.NewReader(`{}`))
	matchReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	matchW := httptest.NewRecorder()
	e.ServeHTTP(matchW, matchReq)
	assert.Equal(t, http.StatusUnauthorized, matchW.Code)
}
