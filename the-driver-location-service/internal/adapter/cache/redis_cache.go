package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"the-driver-location-service/config"
	"the-driver-location-service/internal/domain"
	"the-driver-location-service/internal/ports/secondary"
)

type RedisDriverCache struct {
	client *redis.Client
}

var _ secondary.DriverCache = (*RedisDriverCache)(nil)

func NewRedisDriverCache(client *redis.Client) *RedisDriverCache {
	return &RedisDriverCache{
		client: client,
	}
}

func NewRedisClient(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Address,
		Password:     cfg.Password,
		DB:           cfg.DB,
		MaxRetries:   cfg.MaxRetries,
		PoolSize:     cfg.PoolSize,
		DialTimeout:  cfg.Timeout,
		ReadTimeout:  cfg.Timeout,
		WriteTimeout: cfg.Timeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return client, nil
}

func (c *RedisDriverCache) Get(ctx context.Context, driverID string) (*domain.Driver, error) {
	key := c.generateDriverKey(driverID)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get driver from cache: %w", err)
	}

	var driver domain.Driver
	if err := json.Unmarshal([]byte(data), &driver); err != nil {
		return nil, fmt.Errorf("failed to unmarshal driver: %w", err)
	}

	return &driver, nil
}

func (c *RedisDriverCache) Set(ctx context.Context, driverID string, driver *domain.Driver, ttl time.Duration) error {
	key := c.generateDriverKey(driverID)

	data, err := json.Marshal(driver)
	if err != nil {
		return fmt.Errorf("failed to marshal driver: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set driver in cache: %w", err)
	}

	return nil
}

func (c *RedisDriverCache) Delete(ctx context.Context, driverID string) error {
	key := c.generateDriverKey(driverID)

	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete driver from cache: %w", err)
	}

	return nil
}

func (c *RedisDriverCache) GetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int) ([]*domain.DriverWithDistance, error) {
	key := c.generateNearbyKey(lat, lon, radius, limit)

	data, err := c.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get nearby drivers from cache: %w", err)
	}

	var drivers []*domain.DriverWithDistance
	if err := json.Unmarshal([]byte(data), &drivers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal nearby drivers: %w", err)
	}

	return drivers, nil
}

func (c *RedisDriverCache) SetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int, drivers []*domain.DriverWithDistance, ttl time.Duration) error {
	key := c.generateNearbyKey(lat, lon, radius, limit)

	data, err := json.Marshal(drivers)
	if err != nil {
		return fmt.Errorf("failed to marshal nearby drivers: %w", err)
	}

	err = c.client.Set(ctx, key, data, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to set nearby drivers in cache: %w", err)
	}

	return nil
}

func (c *RedisDriverCache) IsHealthy(ctx context.Context) bool {
	_, err := c.client.Ping(ctx).Result()
	return err == nil
}

func (c *RedisDriverCache) generateDriverKey(driverID string) string {
	return fmt.Sprintf("driver:%s", driverID)
}

func (c *RedisDriverCache) generateNearbyKey(lat, lon, radius float64, limit int) string {
	return fmt.Sprintf("nearby:%.6f:%.6f:%.0f:%d", lat, lon, radius, limit)
}
