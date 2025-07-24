package config

import (
	"os"
	"strings"
)

type Config struct {
	DriverLocationBaseURL string
	Port                  string
	JWTSecret             string
	DriverLocationAPIKey  string
}

func LoadConfig() *Config {
	baseURL := os.Getenv("DRIVER_LOCATION_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8086"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = ":8087"
	} else if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "changeme"
	}

	apiKey := os.Getenv("DRIVER_LOCATION_API_KEY")

	return &Config{
		DriverLocationBaseURL: baseURL,
		Port:                  port,
		JWTSecret:             jwtSecret,
		DriverLocationAPIKey:  apiKey,
	}
}
