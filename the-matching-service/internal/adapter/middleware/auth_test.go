package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"the-matching-service/internal/adapter/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func generateJWT(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, _ := token.SignedString([]byte(secret))
	return t
}

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

func TestJWTAuthMiddleware_HealthBypass(t *testing.T) {
	e := echo.New()
	cfg := &config.Config{JWTSecret: "testsecret"}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	c := e.NewContext(req, w)

	h := func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	middleware := JWTAuthMiddleware(cfg)(h)
	err := middleware(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, w.Code)
}

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
