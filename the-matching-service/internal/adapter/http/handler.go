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

func (h *MatchHandler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":  "healthy",
		"service": "matching-service",
	})
}

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
