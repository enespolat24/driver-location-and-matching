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

func NewDriverHandler(driverService primary.DriverService) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
	}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *DriverHandler) HealthCheck(c echo.Context) error {
	// TODO: if i have more time i will also add a health check for the database and redis. unutma
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "driver-location-service",
	})
}

// @Summary Create a driver
// @Description Create a new driver with location
// @Tags drivers
// @Accept json
// @Produce json
// @Param driver body domain.CreateDriverRequest true "Driver info"
// @Success 201 {object} domain.Driver
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers [post]
func (h *DriverHandler) CreateDriver(c echo.Context) error {
	var req domain.CreateDriverRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	driver, err := h.driverService.CreateDriver(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, driver)
}

// @Summary Batch create drivers
// @Description Create multiple drivers in a single request
// @Tags drivers
// @Accept json
// @Produce json
// @Param batch body domain.BatchCreateRequest true "Batch driver info"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers/batch [post]
func (h *DriverHandler) BatchCreateDrivers(c echo.Context) error {
	var req domain.BatchCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	drivers, err := h.driverService.BatchCreateDrivers(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"drivers": drivers,
		"count":   len(drivers),
	})
}

// @Summary Search nearby drivers
// @Description Find drivers near a given location
// @Tags drivers
// @Accept json
// @Produce json
// @Param search body domain.SearchRequest true "Search params"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers/search [post]
func (h *DriverHandler) SearchNearbyDrivers(c echo.Context) error {
	var req domain.SearchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	drivers, err := h.driverService.SearchNearbyDrivers(req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"drivers": drivers,
		"count":   len(drivers),
	})
}

// @Summary Get driver by ID
// @Description Get a driver by its ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} domain.Driver
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers/{id} [get]
func (h *DriverHandler) GetDriver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Driver ID is required",
		})
	}

	driver, err := h.driverService.GetDriver(id)
	if err != nil {
		return c.JSON(http.StatusNotFound, map[string]interface{}{
			"error":   "not_found",
			"message": "Driver not found",
		})
	}

	return c.JSON(http.StatusOK, driver)
}

// @Summary Update driver by ID
// @Description Update a driver's information by ID
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param driver body domain.Driver true "Driver info"
// @Success 200 {object} domain.Driver
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Driver ID is required",
		})
	}

	var driver domain.Driver
	if err := c.Bind(&driver); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	driver.ID = id

	if err := h.driverService.UpdateDriver(&driver); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, driver)
}

// @Summary Update driver location
// @Description Update a driver's location by ID
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID"
// @Param location body domain.Point true "New location"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers/{id}/location [patch]
func (h *DriverHandler) UpdateDriverLocation(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Driver ID is required",
		})
	}

	var location domain.Point
	if err := c.Bind(&location); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	if err := h.driverService.UpdateDriverLocation(id, location); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Driver location updated successfully",
	})
}

// @Summary Delete driver by ID
// @Description Delete a driver by its ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security X-API-KEY
// @Router /api/v1/drivers/{id} [delete]
func (h *DriverHandler) DeleteDriver(c echo.Context) error {
	id := c.Param("id")
	if id == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Driver ID is required",
		})
	}

	if err := h.driverService.DeleteDriver(id); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": "Driver deleted successfully",
	})
}
