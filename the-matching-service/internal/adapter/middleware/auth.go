package middleware

import (
	"net/http"

	"the-matching-service/internal/adapter/config"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

func JWTAuthMiddleware(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Path() == "/health" {
				return next(c)
			}

			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "Missing Authorization header",
				})
			}

			if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
				tokenString = tokenString[7:]
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "Invalid or expired token",
				})
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "Invalid token claims",
				})
			}

			c.Set("user", claims)
			// i've added this to the context to check if the user is authenticated
			// we assume user is authenticated if the authenticated claim is true
			isAuth := false
			if v, ok := claims["authenticated"]; ok {
				if b, ok := v.(bool); ok && b {
					isAuth = true
				}
			}
			c.Set("is_authenticated", isAuth)

			if uid, ok := claims["user_id"].(string); ok {
				c.Set("user_id", uid)
			} else if sub, ok := claims["sub"].(string); ok {
				c.Set("user_id", sub)
			} else {
				return c.JSON(http.StatusUnauthorized, map[string]interface{}{
					"error":   "unauthorized",
					"message": "user_id or sub claim is required in JWT",
				})
			}

			return next(c)
		}
	}
}
