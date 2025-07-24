package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfig_Defaults tests configuration loading with default values when no environment variables are set
// Expected: Should load default configuration values for all settings
func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("DRIVER_LOCATION_BASE_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DRIVER_LOCATION_API_KEY")

	cfg := LoadConfig()
	assert.Equal(t, "http://localhost:8087", cfg.DriverLocationBaseURL)
	assert.Equal(t, ":8087", cfg.Port)
	assert.Equal(t, "changeme", cfg.JWTSecret)
	assert.Equal(t, "", cfg.DriverLocationAPIKey)
}

// TestLoadConfig_EnvOverride tests configuration loading with environment variable overrides
// Expected: Should load configuration values from environment variables when they are set
func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("DRIVER_LOCATION_BASE_URL", "http://test-url")
	os.Setenv("PORT", ":9999")
	os.Setenv("JWT_SECRET", "mysecret")
	os.Setenv("DRIVER_LOCATION_API_KEY", "apikey123")

	cfg := LoadConfig()
	assert.Equal(t, "http://test-url", cfg.DriverLocationBaseURL)
	assert.Equal(t, ":9999", cfg.Port)
	assert.Equal(t, "mysecret", cfg.JWTSecret)
	assert.Equal(t, "apikey123", cfg.DriverLocationAPIKey)
}
