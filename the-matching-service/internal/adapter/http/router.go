package httpadapter

import (
	"the-matching-service/internal/adapter/config"
	"the-matching-service/internal/adapter/middleware"

	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
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
	e.Use(middleware.JWTAuthMiddleware(cfg))

	r := &Router{
		echo:    e,
		handler: handler,
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {

	r.echo.GET("/health", r.handler.HealthCheck)

	v1 := r.echo.Group("/api/v1")
	v1.POST("/match", r.handler.Match)
}

func (r *Router) Start(address string) error {
	return r.echo.Start(address)
}

func (r *Router) Shutdown() error {
	return r.echo.Close()
}
