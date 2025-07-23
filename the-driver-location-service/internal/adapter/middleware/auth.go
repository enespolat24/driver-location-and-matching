package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// AuthConfig holds authentication configuration
type AuthConfig struct {
	MatchingAPIKey string `json:"matching_api_key"`
	RequireAuth    bool   `json:"require_auth"`
}

// APIKeyAuthMiddleware creates middleware for API key authentication
func APIKeyAuthMiddleware(config AuthConfig) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			path := c.Request().URL.Path
			if path == "/health" || path == "/" || strings.HasPrefix(path, "/swagger/") {
				return next(c)
			}

			apiKey := c.Request().Header.Get("X-API-Key")
			if apiKey == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "API key is required",
				})
			}

			if strings.TrimSpace(apiKey) != strings.TrimSpace(config.MatchingAPIKey) {
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
