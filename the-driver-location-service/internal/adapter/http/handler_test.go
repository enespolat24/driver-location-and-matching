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
	return nil, nil
}
func (m *MockDriverService) SearchNearbyDrivers(req domain.SearchRequest) ([]*domain.DriverWithDistance, error) {
	return nil, nil
}
func (m *MockDriverService) GetDriver(id string) (*domain.Driver, error) {
	return nil, nil
}
func (m *MockDriverService) UpdateDriver(driver *domain.Driver) error {
	return nil
}
func (m *MockDriverService) UpdateDriverLocation(id string, location domain.Point) error {
	return nil
}
func (m *MockDriverService) DeleteDriver(id string) error {
	return nil
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
