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

type MatchRequest struct {
	Name     string          `json:"name"`
	Location domain.Location `json:"location"`
	Radius   float64         `json:"radius"`
}

func (h *MatchHandler) Match(c echo.Context) error {
	user := c.Get("user")
	if user == nil {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "unauthorized",
			"message": "Missing or invalid token",
		})
	}
	claims, ok := user.(map[string]interface{})
	if !ok || claims["authenticated"] != true {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error":   "unauthorized",
			"message": "User not authenticated",
		})
	}

	var req MatchRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error":   "invalid_request",
			"message": "Invalid request body",
		})
	}

	rider := domain.NewRider(req.Name, req.Location)
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

	return c.JSON(http.StatusOK, result)
}
