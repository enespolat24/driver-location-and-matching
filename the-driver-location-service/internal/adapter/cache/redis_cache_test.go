package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"the-driver-location-service/internal/domain"
)

func setupRedisTestCache(t *testing.T) (*RedisDriverCache, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "redis:7",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp").WithStartupTimeout(20 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)
	addr := fmt.Sprintf("%s:%s", host, port.Port())

	client := redis.NewClient(&redis.Options{Addr: addr})
	require.NoError(t, client.Ping(ctx).Err())

	cache := NewRedisDriverCache(client)

	cleanup := func() {
		client.Close()
		container.Terminate(ctx)
	}
	return cache, cleanup
}

// TestRedisDriverCache_SetGetDelete tests basic cache operations (set, get, delete)
// Expected: Should successfully store, retrieve, and delete driver data from Redis cache
func TestRedisDriverCache_SetGetDelete(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(29, 41)}
	require.NoError(t, cache.Set(ctx, drv.ID, drv, 2*time.Second))
	got, err := cache.Get(ctx, drv.ID)
	require.NoError(t, err)
	assert.Equal(t, drv.ID, got.ID)
	assert.Equal(t, drv.Location.Longitude(), got.Location.Longitude())
	assert.Equal(t, drv.Location.Latitude(), got.Location.Latitude())
	require.NoError(t, cache.Delete(ctx, drv.ID))
	gone, err := cache.Get(ctx, drv.ID)
	require.NoError(t, err)
	assert.Nil(t, gone)
}

// TestRedisDriverCache_Get_CacheMiss tests cache retrieval when key doesn't exist
// Expected: Should return nil, nil when cache key is not found (cache miss)
func TestRedisDriverCache_Get_CacheMiss(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	got, err := cache.Get(ctx, "non-existent-driver")
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRedisDriverCache_Get_CorruptData tests error handling when cached data is corrupted
// Expected: Should return error when cached data is corrupted/invalid JSON
func TestRedisDriverCache_Get_CorruptData(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	key := "driver:corrupt-test"
	err := cache.client.Set(ctx, key, "invalid-json-data", time.Minute).Err()
	require.NoError(t, err)

	got, err := cache.Get(ctx, "corrupt-test")
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to unmarshal driver")
}

// TestRedisDriverCache_IsHealthy tests the health check functionality
// Expected: Should return true when Redis is connected and responsive
func TestRedisDriverCache_IsHealthy(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	isHealthy := cache.IsHealthy(ctx)
	assert.True(t, isHealthy, "Redis should be healthy when properly connected")
}

// TestRedisDriverCache_IsHealthy_Disconnected tests health check when Redis is not available
// Expected: Should return false when Redis is not reachable
func TestRedisDriverCache_IsHealthy_Disconnected(t *testing.T) {
	invalidClient := redis.NewClient(&redis.Options{
		Addr:     "invalid-host:9999",
		Password: "",
		DB:       0,
	})

	cache := NewRedisDriverCache(invalidClient)
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	isHealthy := cache.IsHealthy(ctx)
	assert.False(t, isHealthy, "Redis should not be healthy when connection is broken")
}

// TestRedisDriverCache_TTLExpiration tests TTL expiration functionality
// Expected: Should return nil when data has expired after TTL timeout
func TestRedisDriverCache_TTLExpiration(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	driver := &domain.Driver{ID: "temp_driver", Location: domain.NewPoint(29, 41)}

	require.NoError(t, cache.Set(ctx, driver.ID, driver, 100*time.Millisecond))

	got, err := cache.Get(ctx, driver.ID)
	require.NoError(t, err)
	assert.Equal(t, driver.ID, got.ID)

	time.Sleep(150 * time.Millisecond)

	got, err = cache.Get(ctx, driver.ID)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRedisDriverCache_MultipleDrivers tests cache operations with multiple drivers
// Expected: Should handle multiple drivers independently in cache storage and retrieval
func TestRedisDriverCache_MultipleDrivers(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	drivers := []*domain.Driver{
		{ID: "driver_1", Location: domain.NewPoint(29.1, 41.1)},
		{ID: "driver_2", Location: domain.NewPoint(29.2, 41.2)},
		{ID: "driver_3", Location: domain.NewPoint(29.3, 41.3)},
	}

	for _, driver := range drivers {
		require.NoError(t, cache.Set(ctx, driver.ID, driver, time.Minute))
	}

	for _, expected := range drivers {
		got, err := cache.Get(ctx, expected.ID)
		require.NoError(t, err)
		assert.Equal(t, expected.ID, got.ID)
		assert.Equal(t, expected.Location.Longitude(), got.Location.Longitude())
		assert.Equal(t, expected.Location.Latitude(), got.Location.Latitude())
	}

	for _, driver := range drivers {
		require.NoError(t, cache.Delete(ctx, driver.ID))
	}

	for _, driver := range drivers {
		got, err := cache.Get(ctx, driver.ID)
		require.NoError(t, err)
		assert.Nil(t, got)
	}
}
