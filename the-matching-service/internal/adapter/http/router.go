package httpadapter

import (
	"github.com/labstack/echo/v4"
)

type Router struct {
	echo    *echo.Echo
	handler *MatchHandler
}

func NewRouter(handler *MatchHandler) *Router {
	e := echo.New()

	r := &Router{
		echo:    e,
		handler: handler,
	}

	r.setupRoutes()
	return r
}

func (r *Router) setupRoutes() {
	v1 := r.echo.Group("/api/v1")
	v1.POST("/match", r.handler.Match)
}

func (r *Router) Start(address string) error {
	return r.echo.Start(address)
}

func (r *Router) Shutdown() error {
	return r.echo.Close()
}
