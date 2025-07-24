package http

import (
	"the-driver-location-service/internal/adapter/middleware"
	"the-driver-location-service/internal/ports/primary"

	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Router struct {
	echo    *echo.Echo
	handler *DriverHandler
	config  middleware.AuthConfig
}

func NewRouter(driverService primary.DriverService, authConfig middleware.AuthConfig) *Router {
	e := echo.New()
	handler := NewDriverHandler(driverService)

	router := &Router{
		echo:    e,
		handler: handler,
		config:  authConfig,
	}

	router.setupMiddleware()
	router.setupRoutes()

	return router
}

func (r *Router) setupMiddleware() {
	r.echo.Use(echomiddleware.Logger())
	r.echo.Use(echomiddleware.Recover())
	r.echo.Use(echomiddleware.CORS())
	r.echo.Use(echoprometheus.NewMiddleware("driver_location_service"))
}

func (r *Router) setupRoutes() {
	r.echo.GET("/health", r.handler.HealthCheck)
	r.echo.GET("/metrics", echoprometheus.NewHandler())
	r.echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// API v1 routes
	v1 := r.echo.Group("/api/v1")

	// Driver routes
	drivers := v1.Group("/drivers")
	drivers.Use(middleware.APIKeyAuthMiddleware(r.config))
	{
		drivers.POST("", r.handler.CreateDrivers)                      // Create driver(s) - supports both single and batch
		drivers.POST("/search", r.handler.SearchNearbyDrivers)         // Search nearby drivers
		drivers.GET("/:id", r.handler.GetDriver)                       // Get driver by ID
		drivers.PUT("/:id", r.handler.UpdateDriver)                    // Update driver by ID
		drivers.PATCH("/:id/location", r.handler.UpdateDriverLocation) // Update driver location
		drivers.DELETE("/:id", r.handler.DeleteDriver)                 // Delete driver
	}
}

func (r *Router) GetEcho() *echo.Echo {
	return r.echo
}

func (r *Router) Start(address string) error {
	return r.echo.Start(address)
}

func (r *Router) Shutdown() error {
	return r.echo.Close()
}
