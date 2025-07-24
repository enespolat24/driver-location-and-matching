package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"errors"
	"the-driver-location-service/internal/adapter/middleware"
	"the-driver-location-service/internal/domain"
)

type mockDriverService struct {
	mock.Mock
}

func (m *mockDriverService) CreateDriver(req domain.CreateDriverRequest) (*domain.Driver, error) {
	args := m.Called(req)
	return args.Get(0).(*domain.Driver), args.Error(1)
}

func (m *mockDriverService) BatchCreateDrivers(req domain.BatchCreateRequest) ([]*domain.Driver, error) {
	args := m.Called(req)
	return args.Get(0).([]*domain.Driver), args.Error(1)
}

func (m *mockDriverService) SearchNearbyDrivers(req domain.SearchRequest) ([]*domain.DriverWithDistance, error) {
	args := m.Called(req)
	return args.Get(0).([]*domain.DriverWithDistance), args.Error(1)
}

func (m *mockDriverService) GetDriver(id string) (*domain.Driver, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Driver), args.Error(1)
}

func (m *mockDriverService) UpdateDriver(driver *domain.Driver) error {
	args := m.Called(driver)
	return args.Error(0)
}

func (m *mockDriverService) UpdateDriverLocation(id string, location domain.Point) error {
	args := m.Called(id, location)
	return args.Error(0)
}

func (m *mockDriverService) DeleteDriver(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func resetPrometheusRegistry() {
	prometheus.DefaultRegisterer = prometheus.NewRegistry()
}

// TestNewRouter tests router creation with valid dependencies
// Expected: Should create router instance with properly initialized echo server and handler
func TestNewRouter(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	assert.NotNil(t, router)
	assert.NotNil(t, router.echo)
	assert.NotNil(t, router.handler)
}

// TestRouter_HealthCheck tests the health check endpoint
// Expected: Should return 200 OK with health status information
func TestRouter_HealthCheck(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	err := router.handler.HealthCheck(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, "Service is healthy", response.Message)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, "healthy", data["status"])
	assert.Equal(t, "driver-location-service", data["service"])
}

// TestRouter_CreateDriver_Success tests successful driver creation endpoint
// Expected: Should return 201 Created with driver data when request is valid
func TestRouter_CreateDriver_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	e := echo.New()
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	mockService.On("BatchCreateDrivers", mock.Anything).Return([]*domain.Driver{drv}, nil)

	err := router.handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	assert.Contains(t, rec.Body.String(), "driver")
	mockService.AssertExpectations(t)
}

// TestRouter_CreateDriver_InvalidRequest tests driver creation with invalid request body
// Expected: Should return 400 Bad Request when request body is malformed
func TestRouter_CreateDriver_InvalidRequest(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"invalid": json}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	err := router.handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)
	assert.Equal(t, "invalid_request", response.Error)
}

// TestRouter_CreateDriver_ServiceError tests driver creation when service returns error
// Expected: Should return 500 Internal Server Error when service operation fails
func TestRouter_CreateDriver_ServiceError(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	e := echo.New()
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService.On("BatchCreateDrivers", mock.Anything).Return(([]*domain.Driver)(nil), errors.New("db error"))

	err := router.handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "db error")
	mockService.AssertExpectations(t)
}

// TestRouter_BatchCreateDrivers_Success tests successful batch driver creation endpoint
// Expected: Should return 201 Created with created drivers data when request is valid
func TestRouter_BatchCreateDrivers_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	e := echo.New()
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}, {"id":"d2","location":{"type":"Point","coordinates":[30,42]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	drv1 := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	drv2 := &domain.Driver{ID: "d2", Location: domain.NewPoint(30, 42)}
	mockService.On("BatchCreateDrivers", mock.Anything).Return([]*domain.Driver{drv1, drv2}, nil)

	err := router.handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	assert.Contains(t, rec.Body.String(), "d2")
	assert.Contains(t, rec.Body.String(), "drivers")
	mockService.AssertExpectations(t)
}

// TestRouter_SearchNearbyDrivers_Success tests successful nearby driver search endpoint
// Expected: Should return 200 OK with nearby drivers data when request is valid
func TestRouter_SearchNearbyDrivers_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"location":{"type":"Point","coordinates":[29.0,41.0]},"radius":1000,"limit":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	expectedDrivers := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "driver1"}, Distance: 100.0},
	}

	mockService.On("SearchNearbyDrivers", mock.AnythingOfType("domain.SearchRequest")).Return(expectedDrivers, nil)

	err := router.handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	data := response.Data.(map[string]interface{})
	assert.Equal(t, float64(1), data["count"])

	mockService.AssertExpectations(t)
}

// TestRouter_GetDriver_Success tests successful driver retrieval endpoint
// Expected: Should return 200 OK with driver data when driver exists
func TestRouter_GetDriver_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers/driver1", nil)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("driver1")

	expectedDriver := &domain.Driver{
		ID:       "driver1",
		Location: domain.NewPoint(29.0, 41.0),
	}

	mockService.On("GetDriver", "driver1").Return(expectedDriver, nil)

	err := router.handler.GetDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	driverData := response.Data.(map[string]interface{})
	assert.Equal(t, expectedDriver.ID, driverData["id"])

	mockService.AssertExpectations(t)
}

// TestRouter_GetDriver_NotFound tests driver retrieval when driver doesn't exist
// Expected: Should return 404 Not Found when driver is not found
func TestRouter_GetDriver_NotFound(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	mockService.On("GetDriver", "nonexistent").Return((*domain.Driver)(nil), assert.AnError)

	err := router.handler.GetDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)

	mockService.AssertExpectations(t)
}

// TestRouter_UpdateDriver_Success tests successful driver update endpoint
// Expected: Should return 200 OK when driver update is successful
func TestRouter_UpdateDriver_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"location":{"type":"Point","coordinates":[30.0,42.0]}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drivers/driver1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("driver1")

	mockService.On("UpdateDriver", mock.AnythingOfType("*domain.Driver")).Return(nil)

	err := router.handler.UpdateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

// TestRouter_UpdateDriverLocation_Success tests successful driver location update endpoint
// Expected: Should return 200 OK when location update is successful
func TestRouter_UpdateDriverLocation_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"type":"Point","coordinates":[30.0,42.0]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/drivers/driver1/location", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("driver1")

	mockService.On("UpdateDriverLocation", "driver1", mock.AnythingOfType("domain.Point")).Return(nil)

	err := router.handler.UpdateDriverLocation(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

// TestRouter_DeleteDriver_Success tests successful driver deletion endpoint
// Expected: Should return 200 OK when driver deletion is successful
func TestRouter_DeleteDriver_Success(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/drivers/driver1", nil)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("driver1")

	mockService.On("DeleteDriver", "driver1").Return(nil)

	err := router.handler.DeleteDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response.Success)

	mockService.AssertExpectations(t)
}

// TestRouter_DeleteDriver_NotFound tests driver deletion when driver doesn't exist
// Expected: Should return 500 Internal Server Error when deletion fails
func TestRouter_DeleteDriver_NotFound(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/drivers/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("nonexistent")

	mockService.On("DeleteDriver", "nonexistent").Return(assert.AnError)

	err := router.handler.DeleteDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response APIResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.False(t, response.Success)

	mockService.AssertExpectations(t)
}

func TestRouter_RoutesRegistration(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	routes := router.echo.Routes()
	expectedRoutes := []string{
		"GET /health",
		"POST /api/v1/drivers",
		"POST /api/v1/drivers/search",
		"GET /api/v1/drivers/:id",
		"PUT /api/v1/drivers/:id",
		"PATCH /api/v1/drivers/:id/location",
		"DELETE /api/v1/drivers/:id",
		"GET /metrics",
	}

	for _, expectedRoute := range expectedRoutes {
		found := false
		for _, route := range routes {
			if route.Method+" "+route.Path == expectedRoute {
				found = true
				break
			}
		}
		assert.True(t, found, "Route %s not found", expectedRoute)
	}
}

func TestRouter_GetEcho(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	echo := router.GetEcho()
	assert.NotNil(t, echo)
	assert.Equal(t, router.echo, echo)
}

func TestRouter_Shutdown(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	err := router.Shutdown()
	assert.NoError(t, err)
}

func TestRouter_Start(t *testing.T) {
	resetPrometheusRegistry()
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	go func() {
		router.Start(":0")
	}()

	router.Shutdown()
}
