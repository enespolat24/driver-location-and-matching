package httpadapter

import (
	"the-matching-service/internal/adapter/config"
	"the-matching-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Router struct {
	echo    *echo.Echo
	handler *MatchHandler
}

func NewRouter(handler *MatchHandler, cfg *config.Config) *Router {
	e := echo.New()

	// Logger, Recover ve CORS middleware'leri
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())

	r := &Router{
		echo:    e,
		handler: handler,
	}

	r.setupRoutes(cfg)
	return r
}

func (r *Router) setupRoutes(cfg *config.Config) {
	r.echo.GET("/swagger/*", echoSwagger.WrapHandler)
	r.echo.GET("/health", r.handler.HealthCheck)

	// routes with authentication
	v1 := r.echo.Group("/api/v1", middleware.JWTAuthMiddleware(cfg))
	v1.POST("/match", r.handler.Match)
}

func (r *Router) Start(address string) error {
	return r.echo.Start(address)
}

func (r *Router) Shutdown() error {
	return r.echo.Close()
}

// GetEcho exposes the underlying Echo instance for middleware attachment
// solely for testing purposes
func (r *Router) GetEcho() *echo.Echo {
	return r.echo
}
