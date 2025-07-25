package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type AuthConfig struct {
	MatchingAPIKey string `json:"matching_api_key"`
	RequireAuth    bool   `json:"require_auth"`
}

// Instead of using API key authentication, I could have alternatively
// restricted access to the service at the network level.
func APIKeyAuthMiddleware(config AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "API key is required",
				})
			}

			expectedKey := strings.TrimSpace(config.MatchingAPIKey)
			if expectedKey == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "Server misconfiguration: API key is not set",
				})
			}

			if strings.TrimSpace(apiKey) != expectedKey {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "Invalid API key",
				})
			}

			return next(c)
		}
	}
}

func CORSMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Response().Header().Set("Access-Control-Allow-Origin", "*")
			c.Response().Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
			c.Response().Header().Set("Access-Control-Allow-Headers", "Content-Type, X-API-Key")

			if c.Request().Method == "OPTIONS" {
				return c.NoContent(http.StatusOK)
			}

			return next(c)
		}
	}
}

func LoggingMiddleware() echo.MiddlewareFunc {
	return middleware.Logger()
}

func RecoveryMiddleware() echo.MiddlewareFunc {
	return middleware.Recover()
}
