package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"the-matching-service/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func generateJWT(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(secret))
	return t
}

// TestJWTAuthMiddleware_validToken tests successful authentication with valid JWT token
// Expected: Should authenticate successfully and set user_id and is_authenticated in context
func TestJWTAuthMiddleware_validToken(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}
	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error {
		userID := c.Get("user_id")
		isAuth := c.Get("is_authenticated")
		assert.Equal(t, "user-1", userID)
		assert.Equal(t, true, isAuth)
		return c.String(http.StatusOK, "ok")
	}

	middleware := JWTAuthMiddleware(cfg)(h)
	assert.NoError(t, middleware(c))
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestJWTAuthMiddleware_missingToken tests authentication failure when no token is provided
// Expected: Should return HTTP 401 Unauthorized with missing authorization header message
func TestJWTAuthMiddleware_missingToken(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	middleware := JWTAuthMiddleware(cfg)(h)
	_ = middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Missing Authorization header")
}

// TestJWTAuthMiddleware_invalidToken tests authentication failure with invalid JWT token
// Expected: Should return HTTP 401 Unauthorized with invalid token message
func TestJWTAuthMiddleware_invalidToken(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer invalidtoken")
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	middleware := JWTAuthMiddleware(cfg)(h)
	_ = middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired token")
}

// TestJWTAuthMiddleware_missingUserID tests authentication failure when token lacks user_id claim
// Expected: Should return HTTP 401 Unauthorized with missing user_id claim message
func TestJWTAuthMiddleware_missingUserID(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}
	claims := jwt.MapClaims{"authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error { return c.String(http.StatusOK, "ok") }
	middleware := JWTAuthMiddleware(cfg)(h)
	_ = middleware(c)
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "user_id or sub claim is required")
}

// TestJWTAuthMiddleware_authenticatedFalse tests authentication failure when authenticated claim is false
// Expected: Should return HTTP 401 Unauthorized when authenticated claim is false
func TestJWTAuthMiddleware_authenticatedFalse(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}
	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": false}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+token)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error {
		isAuth := c.Get("is_authenticated")
		assert.Equal(t, false, isAuth)
		return c.String(http.StatusOK, "ok")
	}

	middleware := JWTAuthMiddleware(cfg)(h)
	assert.NoError(t, middleware(c))
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestJWTAuthMiddleware_TokenWithoutBearerPrefix tests authentication with token that lacks "Bearer " prefix
// Expected: Should authenticate successfully even without "Bearer " prefix in Authorization header
func TestJWTAuthMiddleware_TokenWithoutBearerPrefix(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}
	claims := jwt.MapClaims{"user_id": "user-1", "authenticated": true}
	token := generateJWT(cfg.JWTSecret, claims)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(echo.HeaderAuthorization, token) // No "Bearer " prefix
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error {
		userID := c.Get("user_id")
		isAuth := c.Get("is_authenticated")
		assert.Equal(t, "user-1", userID)
		assert.Equal(t, true, isAuth)
		return c.String(http.StatusOK, "ok")
	}

	middleware := JWTAuthMiddleware(cfg)(h)
	assert.NoError(t, middleware(c))
	assert.Equal(t, http.StatusOK, w.Code)
}
