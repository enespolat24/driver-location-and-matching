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
// @Success 200 {object} domain.SuccessResponse "Success: data contains MatchResponse"
// @Failure 400 {object} domain.ErrorResponse "Bad Request - Validation error or invalid request"
// @Failure 401 {object} domain.ErrorResponse "Unauthorized - User not authenticated"
// @Failure 404 {object} domain.ErrorResponse "Not Found - No drivers found nearby"
// @Failure 500 {object} domain.ErrorResponse "Internal Server Error"
// @Security BearerAuth
// @Router /api/v1/match [post]
func (h *MatchHandler) Match(c echo.Context) error {
	isAuth, _ := c.Get("is_authenticated").(bool)
	if !isAuth {
		return c.JSON(http.StatusUnauthorized, domain.ErrorResponse{
			Success: false,
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
	}
	userID, _ := c.Get("user_id").(string)

	var req domain.MatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "invalid_request",
			Message: "Invalid request body",
		})
	}

	// Validate the request
	if err := domain.ValidateStruct(&req); err != nil {
		if validationErrors, ok := err.(*domain.ValidationErrors); ok {
			return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
				Success: false,
				Error:   "validation_error",
				Message: "Request validation failed",
				Details: validationErrors.Errors,
			})
		}
		return c.JSON(http.StatusBadRequest, domain.ErrorResponse{
			Success: false,
			Error:   "validation_error",
			Message: err.Error(),
		})
	}

	rider := req.CreateRider(userID)
	result, err := h.matchingService.MatchRiderToDriver(c.Request().Context(), *rider, req.Radius)
	if err != nil {
		if err.Error() == "no drivers found" {
			return c.JSON(http.StatusNotFound, domain.ErrorResponse{
				Success: false,
				Error:   "not_found",
				Message: "No drivers found nearby",
			})
		}
		return c.JSON(http.StatusInternalServerError, domain.ErrorResponse{
			Success: false,
			Error:   "internal_error",
			Message: err.Error(),
		})
	}

	response := domain.NewMatchResponse(result)
	return c.JSON(http.StatusOK, domain.SuccessResponse{
		Success: true,
		Data:    response,
		Message: "Matched successfully",
	})
}
