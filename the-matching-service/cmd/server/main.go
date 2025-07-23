package main

import (
	"log"
	"the-matching-service/internal/adapter/config"
	httpadapter "the-matching-service/internal/adapter/http"
	"the-matching-service/internal/application"
	"the-matching-service/internal/domain"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.LoadConfig()
	log.Println("cfg", cfg.JWTSecret)

	_ = domain.NewCustomValidator()
	log.Println("Custom validator initialized")

	client := httpadapter.NewDriverLocationClient(cfg.DriverLocationBaseURL, cfg.DriverLocationAPIKey)
	service := application.NewMatchingService(client)
	handler := httpadapter.NewMatchHandler(service)
	router := httpadapter.NewRouter(handler, cfg)

	log.Printf("Matching Service listening on %s", cfg.Port)
	if err := router.Start(cfg.Port); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
