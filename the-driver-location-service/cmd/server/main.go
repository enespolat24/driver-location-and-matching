package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	_ "the-driver-location-service/docs"
	"the-driver-location-service/internal/adapter/cache"
	"the-driver-location-service/internal/adapter/config"
	"the-driver-location-service/internal/adapter/db"
	httpAdapter "the-driver-location-service/internal/adapter/http"
	"the-driver-location-service/internal/adapter/middleware"
	"the-driver-location-service/internal/application"
	"the-driver-location-service/internal/ports/primary"
	"the-driver-location-service/internal/ports/secondary"
)

// @title           Driver Location Service API
// @version         1.0
// @description     A service for finding nearby drivers
// @securityDefinitions.apikey X-API-KEY
// @in header
// @name X-API-KEY
// @description Type X-API-KEY followed by a space and API key.

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file not found or could not be loaded, environment variables will be read from the shell")
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

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

	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Continuing without cache...")
	} else {
		log.Println("Connected to Redis successfully")
		driverCache = cache.NewRedisDriverCache(redisClient)
		defer func() {
			if err := redisClient.Close(); err != nil {
				log.Printf("Error closing Redis connection: %v", err)
			}
		}()
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
