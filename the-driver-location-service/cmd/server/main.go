package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"the-driver-location-service/config"
	_ "the-driver-location-service/docs"
	"the-driver-location-service/internal/adapter/cache"
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

	driverRepo, err := db.NewMongoDriverRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB repository: %v", err)
	}

	var driverCache secondary.DriverCache

	redisClient, err := cache.NewRedisClient(cfg.Redis)
	if err != nil {
		log.Fatalf("Warning: Failed to connect to Redis: %v", err)
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

	go func() {
		if err := runDataImport(); err != nil {
			log.Printf("Warning: Data import failed: %v", err)
			log.Println("Continuing without imported data...")
		}
	}()

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

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited gracefully")
}

func runDataImport() error {
	log.Println("Starting data import...")

	cmd := exec.Command("./importer")
	cmd.Dir = "/app" // Set working directory to app directory (containerized)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to run importer: %v", err)
	}

	log.Println("Data import completed successfully.")
	return nil
}
