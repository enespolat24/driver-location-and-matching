package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"the-driver-location-service/internal/adapter/middleware"
	"the-driver-location-service/internal/domain"
)

type mockDriverService struct{ mock.Mock }

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

// TestNewRouter tests router creation with valid dependencies
// Expected: Should create router with proper middleware and routes configured
func TestNewRouter(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}

	router := NewRouter(mockService, authConfig)
	assert.NotNil(t, router)
	assert.NotNil(t, router.echo)
	assert.NotNil(t, router.handler)
	assert.Equal(t, authConfig, router.config)
}

// TestRouter_HealthCheck tests the health check endpoint
// Expected: Should return 200 OK with health status information
func TestRouter_HealthCheck(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	err := router.handler.HealthCheck(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.Equal(t, "driver-location-service", response["service"])
}

// TestRouter_CreateDriver_Success tests successful driver creation endpoint
// Expected: Should return 201 Created with driver data when request is valid
func TestRouter_CreateDriver_Success(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"id":"driver1","location":{"type":"Point","coordinates":[29.0,41.0]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	expectedDriver := &domain.Driver{
		ID:       "driver1",
		Location: domain.NewPoint(29.0, 41.0),
	}

	mockService.On("CreateDriver", mock.AnythingOfType("domain.CreateDriverRequest")).Return(expectedDriver, nil)

	err := router.handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response domain.Driver
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedDriver.ID, response.ID)
	assert.Equal(t, expectedDriver.Location, response.Location)

	mockService.AssertExpectations(t)
}

// TestRouter_CreateDriver_InvalidRequest tests driver creation with invalid request body
// Expected: Should return 400 Bad Request when request body is malformed
func TestRouter_CreateDriver_InvalidRequest(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"invalid": "json"`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	err := router.handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", response["error"])
	assert.Equal(t, "Invalid request body", response["message"])
}

// TestRouter_CreateDriver_ServiceError tests driver creation when service returns error
// Expected: Should return 500 Internal Server Error when service fails
func TestRouter_CreateDriver_ServiceError(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"id":"driver1","location":{"type":"Point","coordinates":[29.0,41.0]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	mockService.On("CreateDriver", mock.AnythingOfType("domain.CreateDriverRequest")).Return((*domain.Driver)(nil), assert.AnError)

	err := router.handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "internal_error", response["error"])

	mockService.AssertExpectations(t)
}

// TestRouter_BatchCreateDrivers_Success tests successful batch driver creation endpoint
// Expected: Should return 201 Created with drivers array and count when request is valid
func TestRouter_BatchCreateDrivers_Success(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"drivers":[{"id":"driver1","location":{"type":"Point","coordinates":[29.0,41.0]}}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/batch", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	expectedDrivers := []*domain.Driver{
		{ID: "driver1", Location: domain.NewPoint(29.0, 41.0)},
	}

	mockService.On("BatchCreateDrivers", mock.AnythingOfType("domain.BatchCreateRequest")).Return(expectedDrivers, nil)

	err := router.handler.BatchCreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["count"])

	mockService.AssertExpectations(t)
}

// TestRouter_SearchNearbyDrivers_Success tests successful nearby driver search endpoint
// Expected: Should return 200 OK with nearby drivers when request is valid
func TestRouter_SearchNearbyDrivers_Success(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"location":{"type":"Point","coordinates":[29.0,41.0]},"radius":1000,"limit":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)

	expectedDrivers := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "driver1"}, Distance: 100},
	}

	mockService.On("SearchNearbyDrivers", mock.AnythingOfType("domain.SearchRequest")).Return(expectedDrivers, nil)

	err := router.handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, float64(1), response["count"])

	mockService.AssertExpectations(t)
}

// TestRouter_GetDriver_Success tests successful driver retrieval endpoint
// Expected: Should return 200 OK with driver data when driver exists
func TestRouter_GetDriver_Success(t *testing.T) {
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

	var response domain.Driver
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, expectedDriver.ID, response.ID)

	mockService.AssertExpectations(t)
}

// TestRouter_GetDriver_NotFound tests driver retrieval when driver doesn't exist
// Expected: Should return 404 Not Found when driver is not found
func TestRouter_GetDriver_NotFound(t *testing.T) {
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

	mockService.AssertExpectations(t)
}

// TestRouter_UpdateDriver_Success tests successful driver update endpoint
// Expected: Should return 200 OK when driver update is successful
func TestRouter_UpdateDriver_Success(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	reqBody := `{"id":"driver1","location":{"type":"Point","coordinates":[30.0,42.0]}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drivers/driver1", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := router.echo.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("driver1")

	expectedDriver := &domain.Driver{
		ID:       "driver1",
		Location: domain.NewPoint(30.0, 42.0),
	}

	mockService.On("UpdateDriver", expectedDriver).Return(nil)

	err := router.handler.UpdateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

// TestRouter_UpdateDriverLocation_Success tests successful driver location update endpoint
// Expected: Should return 200 OK when location update is successful
func TestRouter_UpdateDriverLocation_Success(t *testing.T) {
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

	expectedLocation := domain.NewPoint(30.0, 42.0)

	mockService.On("UpdateDriverLocation", "driver1", expectedLocation).Return(nil)

	err := router.handler.UpdateDriverLocation(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	mockService.AssertExpectations(t)
}

// TestRouter_DeleteDriver_Success tests successful driver deletion endpoint
// Expected: Should return 200 OK when driver deletion is successful
func TestRouter_DeleteDriver_Success(t *testing.T) {
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

	mockService.AssertExpectations(t)
}

// TestRouter_DeleteDriver_NotFound tests driver deletion when driver doesn't exist
// Expected: Should return 500 Internal Server Error when driver is not found
func TestRouter_DeleteDriver_NotFound(t *testing.T) {
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

	mockService.AssertExpectations(t)
}

// TestRouter_RoutesRegistration tests that all routes are properly registered
// Expected: Should have all expected routes registered with correct HTTP methods
func TestRouter_RoutesRegistration(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	// Test that routes are registered by checking the echo router
	routes := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodGet, "/swagger/*"},
		{http.MethodPost, "/api/v1/drivers"},
		{http.MethodPost, "/api/v1/drivers/batch"},
		{http.MethodPost, "/api/v1/drivers/search"},
		{http.MethodGet, "/api/v1/drivers/:id"},
		{http.MethodPut, "/api/v1/drivers/:id"},
		{http.MethodPatch, "/api/v1/drivers/:id/location"},
		{http.MethodDelete, "/api/v1/drivers/:id"},
	}

	// Get all registered routes from echo
	registeredRoutes := router.echo.Routes()

	// Create a map of registered routes for easy lookup
	routeMap := make(map[string]bool)
	for _, route := range registeredRoutes {
		routeMap[route.Method+" "+route.Path] = true
	}

	// Check that all expected routes are registered
	for _, expectedRoute := range routes {
		routeKey := expectedRoute.method + " " + expectedRoute.path
		assert.True(t, routeMap[routeKey], "Route %s %s should be registered", expectedRoute.method, expectedRoute.path)
	}
}

// TestRouter_GetEcho tests that GetEcho returns the echo instance
// Expected: Should return the echo instance used by the router
func TestRouter_GetEcho(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	echo := router.GetEcho()
	assert.NotNil(t, echo)
	assert.Equal(t, router.echo, echo)
}

// TestRouter_Shutdown tests router shutdown functionality
// Expected: Should close the echo server without error
func TestRouter_Shutdown(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	err := router.Shutdown()
	assert.NoError(t, err)
}

// TestRouter_Start tests router start functionality
// Expected: Should start the echo server and return error for invalid address
func TestRouter_Start(t *testing.T) {
	mockService := new(mockDriverService)
	authConfig := middleware.AuthConfig{MatchingAPIKey: "test-key"}
	router := NewRouter(mockService, authConfig)

	// Test with invalid address (should fail)
	err := router.Start("invalid-address")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "listen")
}
