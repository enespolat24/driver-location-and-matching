package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAuthMiddleware_NoAPIKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "API key is required")
}

func TestAPIKeyAuthMiddleware_WrongAPIKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req.Header.Set("X-API-Key", "wrong")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), "Invalid API key")
}

func TestAPIKeyAuthMiddleware_CorrectAPIKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req.Header.Set("X-API-Key", "secret")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "ok", rec.Body.String())
}

func TestAPIKeyAuthMiddleware_HealthEndpointSkipsAuth(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "healthy")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "healthy", rec.Body.String())
}
