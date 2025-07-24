package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestAPIKeyAuthMiddleware_NoAPIKey tests authentication when no API key is provided
// Expected: Should return 401 Unauthorized with "API key is required" message
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

// TestAPIKeyAuthMiddleware_WrongAPIKey tests authentication with incorrect API key
// Expected: Should return 401 Unauthorized with "Invalid API key" message
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

// TestAPIKeyAuthMiddleware_CorrectAPIKey tests authentication with correct API key
// Expected: Should allow request to proceed and return 200 OK
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

// TestAPIKeyAuthMiddleware_WhitespaceHandling tests API key comparison with whitespace
// Expected: Should trim whitespace from both provided and expected API keys
func TestAPIKeyAuthMiddleware_WhitespaceHandling(t *testing.T) {
	e := echo.New()

	// Test with leading/trailing whitespace in provided API key
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req.Header.Set("X-API-Key", "  secret  ")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Test with whitespace in expected API key
	req2 := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req2.Header.Set("X-API-Key", "secret")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	mw2 := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "  secret  "})
	h2 := mw2(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err = h2(c2)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec2.Code)
}

// TestAPIKeyAuthMiddleware_EmptyAPIKey tests authentication with empty API key header
// Expected: Should return 401 Unauthorized when API key header is empty
func TestAPIKeyAuthMiddleware_EmptyAPIKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req.Header.Set("X-API-Key", "")
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

// TestAPIKeyAuthMiddleware_WhitespaceOnlyAPIKey tests authentication with whitespace-only API key
// Expected: Should return 401 Unauthorized when API key contains only whitespace
func TestAPIKeyAuthMiddleware_WhitespaceOnlyAPIKey(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req.Header.Set("X-API-Key", "   ")
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

// TestAPIKeyAuthMiddleware_CaseSensitive tests API key comparison is case sensitive
// Expected: Should return 401 Unauthorized when API key case doesn't match
func TestAPIKeyAuthMiddleware_CaseSensitive(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	req.Header.Set("X-API-Key", "SECRET")
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

// TestCORSMiddleware_RegularRequest tests CORS middleware with regular HTTP request
// Expected: Should set CORS headers and allow request to proceed
func TestCORSMiddleware_RegularRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/drivers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := CORSMiddleware()
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, X-API-Key", rec.Header().Get("Access-Control-Allow-Headers"))
}

// TestCORSMiddleware_OptionsRequest tests CORS middleware with OPTIONS request
// Expected: Should return 200 OK for preflight OPTIONS requests
func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/drivers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := CORSMiddleware()
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, X-API-Key", rec.Header().Get("Access-Control-Allow-Headers"))
}

// TestCORSMiddleware_PostRequest tests CORS middleware with POST request
// Expected: Should set CORS headers for POST requests
func TestCORSMiddleware_PostRequest(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/drivers", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	mw := CORSMiddleware()
	h := mw(func(c echo.Context) error {
		return c.String(http.StatusCreated, "created")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)
	assert.Equal(t, "*", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST, PUT, PATCH, DELETE, OPTIONS", rec.Header().Get("Access-Control-Allow-Methods"))
	assert.Equal(t, "Content-Type, X-API-Key", rec.Header().Get("Access-Control-Allow-Headers"))
}

// TestLoggingMiddleware tests logging middleware creation
// Expected: Should create logging middleware without error
func TestLoggingMiddleware(t *testing.T) {
	mw := LoggingMiddleware()
	assert.NotNil(t, mw)

	// Test that middleware can be applied
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestRecoveryMiddleware tests recovery middleware creation
// Expected: Should create recovery middleware without error
func TestRecoveryMiddleware(t *testing.T) {
	mw := RecoveryMiddleware()
	assert.NotNil(t, mw)

	// Test that middleware can be applied
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := mw(func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TestRecoveryMiddleware_PanicRecovery tests recovery middleware with panic
// Expected: Should recover from panic and return 500 Internal Server Error
func TestRecoveryMiddleware_PanicRecovery(t *testing.T) {
	mw := RecoveryMiddleware()
	assert.NotNil(t, mw)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := mw(func(c echo.Context) error {
		panic("test panic")
	})

	err := h(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// TestAuthConfig_Structure tests AuthConfig structure
// Expected: Should have correct field names and types
func TestAuthConfig_Structure(t *testing.T) {
	config := AuthConfig{
		MatchingAPIKey: "test-key",
		RequireAuth:    true,
	}

	assert.Equal(t, "test-key", config.MatchingAPIKey)
	assert.True(t, config.RequireAuth)
}

// TestAPIKeyAuthMiddleware_DifferentHTTPMethods tests authentication with different HTTP methods
// Expected: Should require authentication for all HTTP methods except excluded paths
func TestAPIKeyAuthMiddleware_DifferentHTTPMethods(t *testing.T) {
	e := echo.New()
	methods := []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete}

	for _, method := range methods {
		req := httptest.NewRequest(method, "/api/v1/drivers", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "Method %s should require authentication", method)
	}
}

// TestAPIKeyAuthMiddleware_ComplexPaths tests authentication with complex API paths
// Expected: Should require authentication for complex API paths
func TestAPIKeyAuthMiddleware_ComplexPaths(t *testing.T) {
	e := echo.New()
	paths := []string{
		"/api/v1/drivers/123",
		"/api/v1/drivers/123/location",
		"/api/v1/drivers/search",
		"/api/v1/drivers/batch",
	}

	for _, path := range paths {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		mw := APIKeyAuthMiddleware(AuthConfig{MatchingAPIKey: "secret"})
		h := mw(func(c echo.Context) error {
			return c.String(http.StatusOK, "ok")
		})

		err := h(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, rec.Code, "Path %s should require authentication", path)
	}
}
