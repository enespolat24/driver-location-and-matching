// TODO : make handler framework independent
package http

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"the-driver-location-service/internal/domain"
	"the-driver-location-service/internal/ports/primary"
)

type DriverHandler struct {
	driverService primary.DriverService
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func NewDriverHandler(driverService primary.DriverService) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
	}
}

func (h *DriverHandler) successResponse(c echo.Context, statusCode int, data interface{}, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: true,
		Data:    data,
		Message: message,
	})
}

func (h *DriverHandler) errorResponse(c echo.Context, statusCode int, errorType string, message string) error {
	return c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   errorType,
		Message: message,
	})
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse
// @Router /health [get]
func (h *DriverHandler) HealthCheck(c echo.Context) error {
	data := map[string]interface{}{
		"status":  "healthy",
		"service": "driver-location-service",
	}
	return h.successResponse(c, http.StatusOK, data, "Service is healthy")
}

// @Summary Create a driver
// @Description Create a new driver with location
// @Tags drivers
// @Accept json
// @Produce json
// @Param driver body domain.CreateDriverRequest true "Driver info"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers [post]
func (h *DriverHandler) CreateDriver(c echo.Context) error {
	var req domain.CreateDriverRequest
	if err := c.Bind(&req); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	driver, err := h.driverService.CreateDriver(req)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "internal_error", err.Error())
	}

	return h.successResponse(c, http.StatusCreated, driver, "Driver created successfully")
}

// @Summary Batch create drivers
// @Description Create multiple drivers in a single request
// @Tags drivers
// @Accept json
// @Produce json
// @Param batch body domain.BatchCreateRequest true "Batch driver info"
// @Success 201 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers/batch [post]
func (h *DriverHandler) BatchCreateDrivers(c echo.Context) error {
	var req domain.BatchCreateRequest
	if err := c.Bind(&req); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	drivers, err := h.driverService.BatchCreateDrivers(req)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "internal_error", err.Error())
	}

	data := map[string]interface{}{
		"drivers": drivers,
		"count":   len(drivers),
	}
	return h.successResponse(c, http.StatusCreated, data, "Drivers created successfully")
}

// @Summary Search nearby drivers
// @Description Find drivers near a given location
// @Tags drivers
// @Accept json
// @Produce json
// @Param search body domain.SearchRequest true "Search params"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers/search [post]
func (h *DriverHandler) SearchNearbyDrivers(c echo.Context) error {
	var req domain.SearchRequest
	if err := c.Bind(&req); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	drivers, err := h.driverService.SearchNearbyDrivers(req)
	if err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "internal_error", err.Error())
	}

	data := map[string]interface{}{
		"drivers": drivers,
		"count":   len(drivers),
	}
	return h.successResponse(c, http.StatusOK, data, "Nearby drivers retrieved successfully")
}

// @Summary Get driver by ID
// @Description Get a driver by its ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 404 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers/{id} [get]
func (h *DriverHandler) GetDriver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Driver ID is required")
	}

	driver, err := h.driverService.GetDriver(id)
	if err != nil {
		return h.errorResponse(c, http.StatusNotFound, "not_found", "Driver not found")
	}

	return h.successResponse(c, http.StatusOK, driver, "Driver retrieved successfully")
}

// @Summary Update driver by ID
// @Description Update a driver's information by ID
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param driver body domain.Driver true "Driver info"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Driver ID is required")
	}

	var driver domain.Driver
	if err := c.Bind(&driver); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	driver.ID = id

	if err := h.driverService.UpdateDriver(&driver); err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "internal_error", err.Error())
	}

	return h.successResponse(c, http.StatusOK, driver, "Driver updated successfully")
}

// @Summary Update driver location
// @Description Update a driver's location by ID
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param location body domain.Point true "New location"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers/{id}/location [patch]
func (h *DriverHandler) UpdateDriverLocation(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Driver ID is required")
	}

	var location domain.Point
	if err := c.Bind(&location); err != nil {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Invalid request body")
	}

	if err := h.driverService.UpdateDriverLocation(id, location); err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "internal_error", err.Error())
	}

	return h.successResponse(c, http.StatusOK, nil, "Driver location updated successfully")
}

// @Summary Delete driver by ID
// @Description Delete a driver by its ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security X-API-KEY
// @Router /api/v1/drivers/{id} [delete]
func (h *DriverHandler) DeleteDriver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return h.errorResponse(c, http.StatusBadRequest, "invalid_request", "Driver ID is required")
	}

	if err := h.driverService.DeleteDriver(id); err != nil {
		return h.errorResponse(c, http.StatusInternalServerError, "internal_error", err.Error())
	}

	return h.successResponse(c, http.StatusOK, nil, "Driver deleted successfully")
}
