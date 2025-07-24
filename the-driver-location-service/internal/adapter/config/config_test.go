package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestLoadConfig_DefaultValues tests config loading with no environment variables set
// Expected: Should load config with default values for all fields
func TestLoadConfig_DefaultValues(t *testing.T) {
	// Clear all relevant environment variables
	clearConfigEnvVars()

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test server defaults
	assert.Equal(t, "8080", config.Server.Port)
	assert.Equal(t, "0.0.0.0", config.Server.Host)
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 120*time.Second, config.Server.IdleTimeout)

	// Test database defaults
	assert.Equal(t, "mongodb://localhost:27017", config.Database.URI)
	assert.Equal(t, "driver_location", config.Database.Database)
	assert.Equal(t, 10*time.Second, config.Database.ConnectTimeout)
	assert.Equal(t, uint64(100), config.Database.MaxPoolSize)
	assert.Equal(t, uint64(10), config.Database.MinPoolSize)

	// Test redis defaults
	assert.Equal(t, "localhost:6379", config.Redis.Address)
	assert.Equal(t, "", config.Redis.Password)
	assert.Equal(t, 0, config.Redis.DB)
	assert.Equal(t, 3, config.Redis.MaxRetries)
	assert.Equal(t, 10, config.Redis.PoolSize)
	assert.Equal(t, 5*time.Second, config.Redis.Timeout)
	assert.True(t, config.Redis.Enabled)

	// Test auth defaults
	assert.Equal(t, "default-matching-api-key", config.Auth.MatchingAPIKey)

	// Test app defaults
	assert.Equal(t, "production", config.App.Environment)
}

// TestLoadConfig_CustomValues tests config loading with custom environment variables
// Expected: Should load config with custom values from environment variables
func TestLoadConfig_CustomValues(t *testing.T) {
	// Set custom environment variables
	setConfigEnvVars(map[string]string{
		"PORT":                  "9090",
		"HOST":                  "127.0.0.1",
		"READ_TIMEOUT":          "60s",
		"WRITE_TIMEOUT":         "60s",
		"IDLE_TIMEOUT":          "300s",
		"MONGO_URI":             "mongodb://custom:27017",
		"MONGO_DATABASE":        "custom_db",
		"MONGO_CONNECT_TIMEOUT": "20s",
		"MONGO_MAX_POOL_SIZE":   "200",
		"MONGO_MIN_POOL_SIZE":   "20",
		"REDIS_ADDRESS":         "custom-redis:6380",
		"REDIS_PASSWORD":        "secret123",
		"REDIS_DB":              "1",
		"REDIS_MAX_RETRIES":     "5",
		"REDIS_POOL_SIZE":       "20",
		"REDIS_TIMEOUT":         "10s",
		"REDIS_ENABLED":         "true",
		"MATCHING_API_KEY":      "custom-api-key",
		"ENVIRONMENT":           "development",
	})

	defer clearConfigEnvVars()

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Test custom server values
	assert.Equal(t, "9090", config.Server.Port)
	assert.Equal(t, "127.0.0.1", config.Server.Host)
	assert.Equal(t, 60*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 60*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 300*time.Second, config.Server.IdleTimeout)

	// Test custom database values
	assert.Equal(t, "mongodb://custom:27017", config.Database.URI)
	assert.Equal(t, "custom_db", config.Database.Database)
	assert.Equal(t, 20*time.Second, config.Database.ConnectTimeout)
	assert.Equal(t, uint64(200), config.Database.MaxPoolSize)
	assert.Equal(t, uint64(20), config.Database.MinPoolSize)

	// Test custom redis values
	assert.Equal(t, "custom-redis:6380", config.Redis.Address)
	assert.Equal(t, "secret123", config.Redis.Password)
	assert.Equal(t, 1, config.Redis.DB)
	assert.Equal(t, 5, config.Redis.MaxRetries)
	assert.Equal(t, 20, config.Redis.PoolSize)
	assert.Equal(t, 10*time.Second, config.Redis.Timeout)
	assert.True(t, config.Redis.Enabled)

	// Test custom auth values
	assert.Equal(t, "custom-api-key", config.Auth.MatchingAPIKey)

	// Test custom app values
	assert.Equal(t, "development", config.App.Environment)
}

// TestLoadConfig_InvalidDurationValues tests config loading with invalid duration values
// Expected: Should fallback to default values when duration parsing fails
func TestLoadConfig_InvalidDurationValues(t *testing.T) {
	setConfigEnvVars(map[string]string{
		"READ_TIMEOUT":          "invalid-duration",
		"WRITE_TIMEOUT":         "not-a-time",
		"MONGO_CONNECT_TIMEOUT": "bad-duration",
		"REDIS_TIMEOUT":         "wrong-format",
	})

	defer clearConfigEnvVars()

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Should fallback to defaults
	assert.Equal(t, 30*time.Second, config.Server.ReadTimeout)
	assert.Equal(t, 30*time.Second, config.Server.WriteTimeout)
	assert.Equal(t, 10*time.Second, config.Database.ConnectTimeout)
	assert.Equal(t, 5*time.Second, config.Redis.Timeout)
}

// TestLoadConfig_InvalidNumericValues tests config loading with invalid numeric values
// Expected: Should fallback to default values when numeric parsing fails
func TestLoadConfig_InvalidNumericValues(t *testing.T) {
	setConfigEnvVars(map[string]string{
		"MONGO_MAX_POOL_SIZE": "not-a-number",
		"MONGO_MIN_POOL_SIZE": "invalid",
		"REDIS_DB":            "bad-number",
		"REDIS_MAX_RETRIES":   "wrong",
		"REDIS_POOL_SIZE":     "invalid-size",
	})

	defer clearConfigEnvVars()

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Should fallback to defaults
	assert.Equal(t, uint64(100), config.Database.MaxPoolSize)
	assert.Equal(t, uint64(10), config.Database.MinPoolSize)
	assert.Equal(t, 0, config.Redis.DB)
	assert.Equal(t, 3, config.Redis.MaxRetries)
	assert.Equal(t, 10, config.Redis.PoolSize)
}

// TestLoadConfig_InvalidBooleanValues tests config loading with invalid boolean values
// Expected: Should fallback to default values when boolean parsing fails
func TestLoadConfig_InvalidBooleanValues(t *testing.T) {
	setConfigEnvVars(map[string]string{
		"REDIS_ENABLED": "not-a-boolean",
	})

	defer clearConfigEnvVars()

	config, err := LoadConfig()
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Should fallback to default (true)
	assert.True(t, config.Redis.Enabled)
}

// TestConfig_Validate_Success tests config validation with valid values
// Expected: Should pass validation when all required fields are present
func TestConfig_Validate_Success(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
		},
		Redis: RedisConfig{
			Enabled: true,
			Address: "localhost:6379",
		},
		Auth: AuthConfig{
			MatchingAPIKey: "test-api-key",
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

// TestConfig_Validate_EmptyDatabaseURI tests config validation with empty database URI
// Expected: Should return error when database URI is empty
func TestConfig_Validate_EmptyDatabaseURI(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URI:      "",
			Database: "test_db",
		},
		Auth: AuthConfig{
			MatchingAPIKey: "test-api-key",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database URI is required")
}

// TestConfig_Validate_EmptyDatabaseName tests config validation with empty database name
// Expected: Should return error when database name is empty
func TestConfig_Validate_EmptyDatabaseName(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URI:      "mongodb://localhost:27017",
			Database: "",
		},
		Auth: AuthConfig{
			MatchingAPIKey: "test-api-key",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database name is required")
}

// TestConfig_Validate_RedisEnabledButNoAddress tests config validation when Redis is enabled but no address
// Expected: Should return error when Redis is enabled but address is empty
func TestConfig_Validate_RedisEnabledButNoAddress(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
		},
		Redis: RedisConfig{
			Enabled: true,
			Address: "",
		},
		Auth: AuthConfig{
			MatchingAPIKey: "test-api-key",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "redis address is required when redis is enabled")
}

// TestConfig_Validate_RedisDisabled tests config validation when Redis is disabled
// Expected: Should pass validation when Redis is disabled (address can be empty)
func TestConfig_Validate_RedisDisabled(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
		},
		Redis: RedisConfig{
			Enabled: false,
			Address: "",
		},
		Auth: AuthConfig{
			MatchingAPIKey: "test-api-key",
		},
	}

	err := config.Validate()
	assert.NoError(t, err)
}

// TestConfig_Validate_EmptyAPIKey tests config validation with empty API key
// Expected: Should return error when matching API key is empty
func TestConfig_Validate_EmptyAPIKey(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URI:      "mongodb://localhost:27017",
			Database: "test_db",
		},
		Auth: AuthConfig{
			MatchingAPIKey: "",
		},
	}

	err := config.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "matching API key is required")
}

// TestConfig_IsDevelopment tests environment detection for development
// Expected: Should return true when environment is set to development
func TestConfig_IsDevelopment(t *testing.T) {
	config := &Config{
		App: AppConfig{
			Environment: "development",
		},
	}

	assert.True(t, config.IsDevelopment())
	assert.False(t, config.IsProduction())
}

// TestConfig_IsProduction tests environment detection for production
// Expected: Should return true when environment is set to production
func TestConfig_IsProduction(t *testing.T) {
	config := &Config{
		App: AppConfig{
			Environment: "production",
		},
	}

	assert.True(t, config.IsProduction())
	assert.False(t, config.IsDevelopment())
}

// TestConfig_IsProduction_OtherEnvironment tests environment detection for other environments
// Expected: Should return false for both development and production when environment is different
func TestConfig_IsProduction_OtherEnvironment(t *testing.T) {
	config := &Config{
		App: AppConfig{
			Environment: "staging",
		},
	}

	assert.False(t, config.IsProduction())
	assert.False(t, config.IsDevelopment())
}

// TestConfig_GetAddress tests server address construction
// Expected: Should return properly formatted host:port address
func TestConfig_GetAddress(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "8080",
		},
	}

	address := config.GetAddress()
	assert.Equal(t, "localhost:8080", address)
}

// TestConfig_GetAddress_EmptyHost tests server address construction with empty host
// Expected: Should handle empty host gracefully
func TestConfig_GetAddress_EmptyHost(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "",
			Port: "8080",
		},
	}

	address := config.GetAddress()
	assert.Equal(t, ":8080", address)
}

// TestConfig_GetAddress_EmptyPort tests server address construction with empty port
// Expected: Should handle empty port gracefully
func TestConfig_GetAddress_EmptyPort(t *testing.T) {
	config := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: "",
		},
	}

	address := config.GetAddress()
	assert.Equal(t, "localhost:", address)
}

// Helper functions for test setup and cleanup

func clearConfigEnvVars() {
	envVars := []string{
		"PORT", "HOST", "READ_TIMEOUT", "WRITE_TIMEOUT", "IDLE_TIMEOUT",
		"MONGO_URI", "MONGO_DATABASE", "MONGO_CONNECT_TIMEOUT", "MONGO_MAX_POOL_SIZE", "MONGO_MIN_POOL_SIZE",
		"REDIS_ADDRESS", "REDIS_PASSWORD", "REDIS_DB", "REDIS_MAX_RETRIES", "REDIS_POOL_SIZE", "REDIS_TIMEOUT", "REDIS_ENABLED",
		"MATCHING_API_KEY", "ENVIRONMENT",
	}

	for _, envVar := range envVars {
		os.Unsetenv(envVar)
	}
}

func setConfigEnvVars(vars map[string]string) {
	for key, value := range vars {
		os.Setenv(key, value)
	}
}
