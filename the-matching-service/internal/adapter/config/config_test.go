package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("DRIVER_LOCATION_BASE_URL")
	os.Unsetenv("PORT")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("DRIVER_LOCATION_API_KEY")

	cfg := LoadConfig()
	assert.Equal(t, "http://localhost:8086", cfg.DriverLocationBaseURL)
	assert.Equal(t, ":8087", cfg.Port)
	assert.Equal(t, "changeme", cfg.JWTSecret)
	assert.Equal(t, "", cfg.DriverLocationAPIKey)
}

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
