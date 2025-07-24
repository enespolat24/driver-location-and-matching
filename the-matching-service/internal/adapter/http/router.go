package httpadapter

import (
	"the-matching-service/config"
	"the-matching-service/internal/adapter/middleware"

	"github.com/labstack/echo-contrib/echoprometheus"
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

	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.CORS())
	e.Use(echoprometheus.NewMiddleware("matching_service"))

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
	r.echo.GET("/metrics", echoprometheus.NewHandler())

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
