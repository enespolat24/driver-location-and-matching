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

func TestCreateDriver_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	mockService.On("CreateDriver", mock.Anything).Return(drv, nil)

	err := handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	mockService.AssertExpectations(t)
}

func TestCreateDriver_InvalidJSON(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader("not-json"))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid request body")
}

func TestCreateDriver_ServiceError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mockService.On("CreateDriver", mock.Anything).Return((*domain.Driver)(nil), errors.New("db error"))

	err := handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "db error")
	mockService.AssertExpectations(t)
}

func TestBatchCreateDrivers_Success(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"drivers":[{"id":"d1","location":{"type":"Point","coordinates":[29,41]}}, {"id":"d2","location":{"type":"Point","coordinates":[30,42]}}]}`
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers/batch", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	drivers := []*domain.Driver{
		{ID: "d1", Location: domain.NewPoint(29, 41)},
		{ID: "d2", Location: domain.NewPoint(30, 42)},
	}
	mockService.On("BatchCreateDrivers", mock.Anything).Return(drivers, nil)

	err := handler.BatchCreateDrivers(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Contains(t, rec.Body.String(), "d1")
	assert.Contains(t, rec.Body.String(), "d2")
	mockService.AssertExpectations(t)
}

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

// TestCreateDriver_ValidationError tests validation error handling in CreateDriver
// Expected: Should return 500 when validation fails (since service validation errors are treated as internal errors)
func TestCreateDriver_ValidationError(t *testing.T) {
	mockService := new(MockDriverService)
	handler := NewDriverHandler(mockService)
	e := echo.New()
	body := `{"id":"","location":{"type":"InvalidType","coordinates":[181,91]}}` // Invalid data
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	validationErr := errors.New("validation error: invalid location type")
	mockService.On("CreateDriver", mock.Anything).Return((*domain.Driver)(nil), validationErr)

	err := handler.CreateDriver(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Body.String(), "internal_error")
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
