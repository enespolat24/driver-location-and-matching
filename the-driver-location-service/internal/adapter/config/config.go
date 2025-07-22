package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	Auth     AuthConfig     `json:"auth"`
	App      AppConfig      `json:"app"`
}

type ServerConfig struct {
	Port         string        `json:"port"`
	Host         string        `json:"host"`
	ReadTimeout  time.Duration `json:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout"`
	IdleTimeout  time.Duration `json:"idle_timeout"`
}

type DatabaseConfig struct {
	URI            string        `json:"uri"`
	Database       string        `json:"database"`
	ConnectTimeout time.Duration `json:"connect_timeout"`
	MaxPoolSize    uint64        `json:"max_pool_size"`
	MinPoolSize    uint64        `json:"min_pool_size"`
}

type AuthConfig struct {
	MatchingAPIKey string `json:"matching_api_key"`
}

type AppConfig struct {
	Environment        string `json:"environment"`
	LogLevel           string `json:"log_level"`
	DefaultSearchLimit int    `json:"default_search_limit"`
	MaxSearchLimit     int    `json:"max_search_limit"`
	DefaultRadius      int    `json:"default_radius"` // in meters
	MaxRadius          int    `json:"max_radius"`     // in meters
}

type RedisConfig struct {
	Address    string        `json:"address"`
	Password   string        `json:"password"`
	DB         int           `json:"db"`
	MaxRetries int           `json:"max_retries"`
	PoolSize   int           `json:"pool_size"`
	Timeout    time.Duration `json:"timeout"`
	Enabled    bool          `json:"enabled"`
}

func LoadConfig() (*Config, error) {
	config := &Config{
		Server: ServerConfig{
			Port:         getEnv("PORT", "8080"),
			Host:         getEnv("HOST", "0.0.0.0"),
			ReadTimeout:  getDurationEnv("READ_TIMEOUT", 30*time.Second),
			WriteTimeout: getDurationEnv("WRITE_TIMEOUT", 30*time.Second),
			IdleTimeout:  getDurationEnv("IDLE_TIMEOUT", 120*time.Second),
		},
		Database: DatabaseConfig{
			URI:            getEnv("MONGO_URI", "mongodb://localhost:27017"),
			Database:       getEnv("MONGO_DATABASE", "driver_location"),
			ConnectTimeout: getDurationEnv("MONGO_CONNECT_TIMEOUT", 10*time.Second),
			MaxPoolSize:    getUint64Env("MONGO_MAX_POOL_SIZE", 100),
			MinPoolSize:    getUint64Env("MONGO_MIN_POOL_SIZE", 10),
		},
		Redis: RedisConfig{
			Address:    getEnv("REDIS_ADDRESS", "localhost:6379"),
			Password:   getEnv("REDIS_PASSWORD", ""),
			DB:         getIntEnv("REDIS_DB", 0),
			MaxRetries: getIntEnv("REDIS_MAX_RETRIES", 3),
			PoolSize:   getIntEnv("REDIS_POOL_SIZE", 10),
			Timeout:    getDurationEnv("REDIS_TIMEOUT", 5*time.Second),
			Enabled:    getBoolEnv("REDIS_ENABLED", true),
		},
		Auth: AuthConfig{
			MatchingAPIKey: getEnv("MATCHING_API_KEY", "default-matching-api-key"),
		},
		App: AppConfig{
			Environment:        getEnv("ENVIRONMENT", "development"),
			LogLevel:           getEnv("LOG_LEVEL", "info"),
			DefaultSearchLimit: getIntEnv("DEFAULT_SEARCH_LIMIT", 10),
			MaxSearchLimit:     getIntEnv("MAX_SEARCH_LIMIT", 100),
			DefaultRadius:      getIntEnv("DEFAULT_RADIUS", 2000), // 2km
			MaxRadius:          getIntEnv("MAX_RADIUS", 50000),    // 50km
		},
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}

	return config, nil
}

func (c *Config) Validate() error {
	if c.Database.URI == "" {
		return fmt.Errorf("database URI is required")
	}

	if c.Database.Database == "" {
		return fmt.Errorf("database name is required")
	}

	if c.Redis.Enabled && c.Redis.Address == "" {
		return fmt.Errorf("redis address is required when redis is enabled")
	}

	if c.Auth.MatchingAPIKey == "" {
		return fmt.Errorf("matching API key is required")
	}

	return nil
}

func (c *Config) IsDevelopment() bool {
	return c.App.Environment == "development"
}

func (c *Config) IsProduction() bool {
	return c.App.Environment == "production"
}

func (c *Config) GetAddress() string {
	return c.Server.Host + ":" + c.Server.Port
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getUint64Env(key string, defaultValue uint64) uint64 {
	if value := os.Getenv(key); value != "" {
		if uintValue, err := strconv.ParseUint(value, 10, 64); err == nil {
			return uintValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
