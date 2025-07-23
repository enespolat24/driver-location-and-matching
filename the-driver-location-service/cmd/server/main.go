package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"the-driver-location-service/internal/adapter/cache"
	"the-driver-location-service/internal/adapter/config"
	"the-driver-location-service/internal/adapter/db"
	httpAdapter "the-driver-location-service/internal/adapter/http"
	"the-driver-location-service/internal/adapter/middleware"
	"the-driver-location-service/internal/application"
	"the-driver-location-service/internal/domain"
	"the-driver-location-service/internal/ports/primary"
	"the-driver-location-service/internal/ports/secondary"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found or could not be loaded, environment variables will be read from the shell")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}
	fmt.Println(cfg.Auth.MatchingAPIKey)

	if cfg.IsDevelopment() {
		log.Println("Starting Driver Location Service in development mode...")
	} else {
		log.Println("Starting Driver Location Service in production mode...")
	}

	driverRepo, err := db.NewMongoDriverRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB repository: %v", err)
	}

	var driverCache secondary.DriverCache
	if cfg.Redis.Enabled {
		redisClient, err := cache.NewRedisClient(cfg.Redis)
		if err != nil {
			log.Printf("Warning: Failed to connect to Redis: %v", err)
			log.Println("Continuing without cache...")
			driverCache = &NoOpCache{}
		} else {
			log.Println("Connected to Redis successfully")
			driverCache = cache.NewRedisDriverCache(redisClient)
			defer func() {
				if err := redisClient.Close(); err != nil {
					log.Printf("Error closing Redis connection: %v", err)
				}
			}()
		}
	} else {
		log.Println("Redis caching is disabled")
		driverCache = &NoOpCache{}
	}

	var driverService primary.DriverService = application.NewDriverApplicationService(driverRepo, driverCache)

	authConfig := middleware.AuthConfig{
		MatchingAPIKey: cfg.Auth.MatchingAPIKey,
	}

	router := httpAdapter.NewRouter(driverService, authConfig)

	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      router.GetEcho(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Starting server on %s", cfg.GetAddress())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// gracefully shutdown
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}

// NoOpCache is a no-operation cache implementation for fallback
type NoOpCache struct{}

func (c *NoOpCache) Get(ctx context.Context, driverID string) (*domain.Driver, error) {
	return nil, nil // Always cache miss
}

func (c *NoOpCache) Set(ctx context.Context, driverID string, driver *domain.Driver, ttl time.Duration) error {
	return nil // Do nothing
}

func (c *NoOpCache) Delete(ctx context.Context, driverID string) error {
	return nil // Do nothing
}

func (c *NoOpCache) GetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int) ([]*domain.DriverWithDistance, error) {
	return nil, nil // Always cache miss
}

func (c *NoOpCache) SetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int, drivers []*domain.DriverWithDistance, ttl time.Duration) error {
	return nil // Do nothing
}

func (c *NoOpCache) InvalidateNearbyCache(ctx context.Context) error {
	return nil // Do nothing
}

func (c *NoOpCache) IsHealthy(ctx context.Context) bool {
	return true // Always healthy since it's no-op
}
