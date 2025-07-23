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

func (h *DriverHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "driver-location-service",
	})
}

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
