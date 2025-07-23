package http

import (
	"the-driver-location-service/internal/adapter/middleware"
	"the-driver-location-service/internal/ports/primary"

	"github.com/labstack/echo/v4"
)

// Router holds the HTTP router configuration
type Router struct {
	echo    *echo.Echo
	handler *DriverHandler
	config  middleware.AuthConfig
}

// NewRouter creates a new HTTP router
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
	r.echo.Use(middleware.LoggingMiddleware())
	r.echo.Use(middleware.RecoveryMiddleware())
	r.echo.Use(middleware.CORSMiddleware())

	auth := middleware.APIKeyAuthMiddleware(r.config)

	r.echo.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().URL.Path == "/health" || c.Request().URL.Path == "/" {
				return next(c)
			}
			return auth(next)(c)
		}
	})
}

func (r *Router) setupRoutes() {
	r.echo.GET("/health", r.handler.HealthCheck)

	// API v1 routes
	v1 := r.echo.Group("/api/v1")

	// Driver routes
	drivers := v1.Group("/drivers")
	{
		drivers.POST("", r.handler.CreateDriver)                       // Create single driver
		drivers.POST("/batch", r.handler.BatchCreateDrivers)           // Batch create drivers
		drivers.POST("/search", r.handler.SearchNearbyDrivers)         // Search nearby drivers
		drivers.GET("/:id", r.handler.GetDriver)                       // Get driver by ID
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
