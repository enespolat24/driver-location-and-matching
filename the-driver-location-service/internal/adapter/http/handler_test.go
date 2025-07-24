package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"the-driver-location-service/internal/domain"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDriverService struct{ mock.Mock }

func (m *MockDriverService) CreateDriver(req domain.CreateDriverRequest) (*domain.Driver, error) {
	args := m.Called(req)
	return args.Get(0).(*domain.Driver), args.Error(1)
}
func (m *MockDriverService) BatchCreateDrivers(req domain.BatchCreateRequest) ([]*domain.Driver, error) {
	args := m.Called(req)
	return args.Get(0).([]*domain.Driver), args.Error(1)
}
func (m *MockDriverService) SearchNearbyDrivers(req domain.SearchRequest) ([]*domain.DriverWithDistance, error) {
	args := m.Called(req)
	return args.Get(0).([]*domain.DriverWithDistance), args.Error(1)
}
func (m *MockDriverService) GetDriver(id string) (*domain.Driver, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Driver), args.Error(1)
}
func (m *MockDriverService) UpdateDriver(driver *domain.Driver) error {
	args := m.Called(driver)
	return args.Error(0)
}
func (m *MockDriverService) UpdateDriverLocation(id string, location domain.Point) error {
	args := m.Called(id, location)
	return args.Error(0)
}
func (m *MockDriverService) DeleteDriver(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

// TestCreateDrivers_SingleDriver_Success tests single driver creation.
// Expected: Should create a single driver and return correct response.
func TestCreateDrivers_SingleDriver_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	mockService.On("BatchCreateDrivers", mock.Anything).Return([]*domain.Driver{drv}, nil)

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	assert.Contains(t, rec.Body.String(), "driver")
	mockService.AssertExpectations(t)
}

// TestCreateDrivers_Batch_Success tests batch driver creation.
// Expected: Should create multiple drivers and return correct response.
func TestCreateDrivers_Batch_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}, {"id":"d2","location":{"type":"Point","coordinates":[30,42]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	drv1 := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	drv2 := &domain.Driver{ID: "d2", Location: domain.NewPoint(30, 42)}
	mockService.On("BatchCreateDrivers", mock.Anything).Return([]*domain.Driver{drv1, drv2}, nil)

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	assert.Contains(t, rec.Body.String(), "d2")
	assert.Contains(t, rec.Body.String(), "drivers")
	mockService.AssertExpectations(t)
}

// TestCreateDrivers_InvalidJSON tests invalid JSON in driver creation.
// Expected: Should return 400 Bad Request for invalid JSON.
func TestCreateDrivers_InvalidJSON(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader("not-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request body")
}

// TestCreateDrivers_ServiceError tests service error in driver creation.
// Expected: Should return 500 Internal Server Error on service failure.
func TestCreateDrivers_ServiceError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService.On("BatchCreateDrivers", mock.Anything).Return(([]*domain.Driver)(nil), errors.New("db error"))

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "db error")
	mockService.AssertExpectations(t)
}

// TestCreateDrivers_ValidationError_EmptyArray tests empty array validation in driver creation.
// Expected: Should return 400 Bad Request for empty array.
func TestCreateDrivers_ValidationError_EmptyArray(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `[]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "At least one driver is required")
}

// TestCreateDrivers_ValidationError_InvalidDriver tests invalid driver data in creation.
// Expected: Should return 500 Internal Server Error for invalid driver data.
func TestCreateDrivers_ValidationError_InvalidDriver(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `[{"id":"","location":{"type":"InvalidType","coordinates":[181,91]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validationErr := errors.New("validation error: invalid location type")
	mockService.On("BatchCreateDrivers", mock.Anything).Return(([]*domain.Driver)(nil), validationErr)

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal_error")
	mockService.AssertExpectations(t)
}

// TestCreateDrivers_MissingContentType tests missing Content-Type header in driver creation.
// Expected: Should return 400 Bad Request for missing Content-Type.
func TestCreateDrivers_MissingContentType(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}]`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request body")
}

// TestSearchNearbyDrivers_Success tests successful nearby driver search.
// Expected: Should return drivers within the given radius.
func TestSearchNearbyDrivers_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"location":{"type":"Point","coordinates":[29,41]},"radius":1000,"limit":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	drivers := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "d1"}, Distance: 100},
		{Driver: domain.Driver{ID: "d2"}, Distance: 200},
	}
	mockService.On("SearchNearbyDrivers", mock.Anything).Return(drivers, nil)

	err := handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	assert.Contains(t, rec.Body.String(), "d2")
	mockService.AssertExpectations(t)
}

// TestGetDriver_Success tests successful driver retrieval by ID.
// Expected: Should return the driver with correct ID.
func TestGetDriver_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers/d1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	mockService.On("GetDriver", "d1").Return(drv, nil)

	err := handler.GetDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	mockService.AssertExpectations(t)
}

// TestGetDriver_NotFound tests retrieval of non-existent driver.
// Expected: Should return 404 Not Found for unknown driver.
func TestGetDriver_NotFound(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers/unknown", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("unknown")
	mockService.On("GetDriver", "unknown").Return((*domain.Driver)(nil), errors.New("not found"))

	err := handler.GetDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)
	assert.Contains(t, rec.Body.String(), "not found")
	mockService.AssertExpectations(t)
}

// TestUpdateDriver_Success tests successful driver update.
// Expected: Should update the driver and return correct response.
func TestUpdateDriver_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drivers/d1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")
	mockService.On("UpdateDriver", mock.MatchedBy(func(d *domain.Driver) bool {
		return d.ID == "d1" && d.Location.Longitude() == 29 && d.Location.Latitude() == 41
	})).Return(nil)

	err := handler.UpdateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	mockService.AssertExpectations(t)
}

// TestUpdateDriverLocation_Success tests successful driver location update.
// Expected: Should update the driver location and return correct response.
func TestUpdateDriverLocation_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"type":"Point","coordinates":[29,41]}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/drivers/d1/location", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")
	mockService.On("UpdateDriverLocation", "d1", domain.NewPoint(29, 41)).Return(nil)

	err := handler.UpdateDriverLocation(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "updated successfully")
	mockService.AssertExpectations(t)
}

// TestDeleteDriver_Success tests successful driver deletion.
// Expected: Should delete the driver and return correct response.
func TestDeleteDriver_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodDelete, "/api/v1/drivers/d1", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")
	mockService.On("DeleteDriver", "d1").Return(nil)

	err := handler.DeleteDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "deleted successfully")
	mockService.AssertExpectations(t)
}

// TestSearchNearbyDrivers_ValidationError tests validation error in search
// Expected: Should return 500 when search validation fails
func TestSearchNearbyDrivers_ValidationError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"location":{"type":"Point","coordinates":[181,91]},"radius":-100,"limit":-1}` // Invalid data
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validationErr := errors.New("validation error: invalid coordinates")
	mockService.On("SearchNearbyDrivers", mock.Anything).Return(([]*domain.DriverWithDistance)(nil), validationErr)

	err := handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal_error")
	mockService.AssertExpectations(t)
}

// TestUpdateDriverLocation_ValidationError tests location validation error
// Expected: Should return 500 when location validation fails
func TestUpdateDriverLocation_ValidationError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"type":"InvalidType","coordinates":[181,91]}` // Invalid location
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/drivers/d1/location", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")

	validationErr := errors.New("validation error: invalid location type")
	mockService.On("UpdateDriverLocation", "d1", mock.Anything).Return(validationErr)

	err := handler.UpdateDriverLocation(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal_error")
	mockService.AssertExpectations(t)
}

// TestSearchNearbyDrivers_InvalidJSON tests search with invalid JSON
// Expected: Should return 400 Bad Request when JSON is malformed
func TestSearchNearbyDrivers_InvalidJSON(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(`{"location":{"type":"Point","coordinates":[29,41invalid]}}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request body")
}

// TestSearchNearbyDrivers_EmptyResults tests search with no results
// Expected: Should return 200 OK with empty array when no drivers found
func TestSearchNearbyDrivers_EmptyResults(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"location":{"type":"Point","coordinates":[29,41]},"radius":1000,"limit":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService.On("SearchNearbyDrivers", mock.Anything).Return([]*domain.DriverWithDistance{}, nil)

	err := handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"count":0`)
	assert.Contains(t, rec.Body.String(), `"drivers":[]`)
	mockService.AssertExpectations(t)
}

// TestSearchNearbyDrivers_ServiceError tests search when service fails
// Expected: Should return 500 Internal Server Error when service returns error
func TestSearchNearbyDrivers_ServiceError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"location":{"type":"Point","coordinates":[29,41]},"radius":1000,"limit":2}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/search", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService.On("SearchNearbyDrivers", mock.Anything).Return(([]*domain.DriverWithDistance)(nil), errors.New("geo index error"))

	err := handler.SearchNearbyDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "geo index error")
	mockService.AssertExpectations(t)
}

// TestUpdateDriver_InvalidJSON tests driver update with invalid JSON
// Expected: Should return 400 Bad Request when JSON is malformed
func TestUpdateDriver_InvalidJSON(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drivers/d1", strings.NewReader(`{"id":"d1","location":invalid}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")

	err := handler.UpdateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request body")
}

// TestUpdateDriver_ServiceError tests driver update when service fails
// Expected: Should return 500 Internal Server Error when service returns error
func TestUpdateDriver_ServiceError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/drivers/d1", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")

	mockService.On("UpdateDriver", mock.Anything).Return(errors.New("update failed"))

	err := handler.UpdateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "update failed")
	mockService.AssertExpectations(t)
}

// TestUpdateDriverLocation_InvalidJSON tests location update with invalid JSON
// Expected: Should return 400 Bad Request when JSON is malformed
func TestUpdateDriverLocation_InvalidJSON(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/drivers/d1/location", strings.NewReader(`{"type":"Point","coordinates":[29,invalid]}`))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("d1")

	err := handler.UpdateDriverLocation(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request body")
}
