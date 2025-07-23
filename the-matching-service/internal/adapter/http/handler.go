package httpadapter

import (
	"net/http"

	"the-matching-service/internal/application"
	"the-matching-service/internal/domain"

	"github.com/labstack/echo/v4"
)

type MatchHandler struct {
	matchingService *application.MatchingService
}

func NewMatchHandler(matchingService *application.MatchingService) *MatchHandler {
	return &MatchHandler{matchingService: matchingService}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the service is healthy
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *MatchHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "matching-service",
	})
}

// Match godoc
// @Summary Match rider with nearby driver
// @Description Find the nearest driver for a rider based on location and radius
// @Tags matching
// @Accept json
// @Produce json
// @Param request body domain.MatchRequest true "Match request"
// @Success 200 {object} domain.MatchResponse
// @Failure 400 {object} map[string]interface{} "Bad Request - Validation error or invalid request"
// @Failure 401 {object} map[string]interface{} "Unauthorized - User not authenticated"
// @Failure 404 {object} map[string]interface{} "Not Found - No drivers found nearby"
// @Failure 500 {object} map[string]interface{} "Internal Server Error"
// @Security BearerAuth
// @Router /api/v1/match [post]
func (h *MatchHandler) Match(c echo.Context) error {
	isAuth, _ := c.Get("is_authenticated").(bool)
	if !isAuth {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}
	userID, _ := c.Get("user_id").(string)

	var req domain.MatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	// Validate the request
	if err := domain.ValidateStruct(&req); err != nil {
		if validationErrors, ok := err.(*domain.ValidationErrors); ok {
			return c.JSON(http.StatusBadRequest, map[string]interface{}{
				"error":   "validation_error",
				"message": "Request validation failed",
				"details": validationErrors.Errors,
			})
		}
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "validation_error",
			"message": err.Error(),
		})
	}

	rider := req.CreateRider(userID)
	result, err := h.matchingService.MatchRiderToDriver(c.Request().Context(), *rider, req.Radius)
	if err != nil {
		if err.Error() == "no drivers found" {
			return c.JSON(http.StatusNotFound, map[string]interface{}{
				"error":   "not_found",
				"message": "No drivers found nearby",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   "internal_error",
			"message": err.Error(),
		})
	}

	response := domain.NewMatchResponse(result)
	return c.JSON(http.StatusOK, response)
}
