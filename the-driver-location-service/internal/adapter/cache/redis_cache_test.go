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

func TestRedisDriverCache_SetGetNearbyDrivers(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()
	drivers := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "d1"}, Distance: 100},
		{Driver: domain.Driver{ID: "d2"}, Distance: 200},
	}
	lat, lon, radius, limit := 41.0, 29.0, 1000.0, 2
	require.NoError(t, cache.SetNearbyDrivers(ctx, lat, lon, radius, limit, drivers, 2*time.Second))
	got, err := cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	require.NoError(t, err)
	assert.Len(t, got, 2)
	assert.Equal(t, "d1", got[0].Driver.ID)
	assert.Equal(t, "d2", got[1].Driver.ID)
}

// TestRedisDriverCache_GetNearbyDrivers_CacheMiss tests when key doesn't exist in cache
// Expected: Should return nil, nil when cache key is not found (cache miss)
func TestRedisDriverCache_GetNearbyDrivers_CacheMiss(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	lat, lon, radius, limit := 50.0, 30.0, 1500.0, 5
	got, err := cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	require.NoError(t, err)
	assert.Nil(t, got)
}

// TestRedisDriverCache_GetNearbyDrivers_CorruptData tests unmarshal error handling
// Expected: Should return error when cached data is corrupted/invalid JSON
func TestRedisDriverCache_GetNearbyDrivers_CorruptData(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	lat, lon, radius, limit := 45.0, 35.0, 2000.0, 3
	key := fmt.Sprintf("nearby:%.6f:%.6f:%.0f:%d", lat, lon, radius, limit)

	err := cache.client.Set(ctx, key, "invalid-json-data", time.Minute).Err()
	require.NoError(t, err)

	got, err := cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	assert.Error(t, err)
	assert.Nil(t, got)
	assert.Contains(t, err.Error(), "failed to unmarshal nearby drivers")
}

// TestRedisDriverCache_GetNearbyDrivers_EmptyArray tests empty driver array
// Expected: Should successfully handle empty driver arrays
func TestRedisDriverCache_GetNearbyDrivers_EmptyArray(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()
	emptyDrivers := []*domain.DriverWithDistance{}
	lat, lon, radius, limit := 42.0, 32.0, 1200.0, 4
	require.NoError(t, cache.SetNearbyDrivers(ctx, lat, lon, radius, limit, emptyDrivers, 2*time.Second))

	got, err := cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Len(t, got, 0)
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

// TestRedisDriverCache_GetNearbyDrivers_LargeDataSet tests with larger data sets
// Expected: Should handle large arrays of drivers without issues
func TestRedisDriverCache_GetNearbyDrivers_LargeDataSet(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	var drivers []*domain.DriverWithDistance
	for i := 0; i < 100; i++ {
		drivers = append(drivers, &domain.DriverWithDistance{
			Driver:   domain.Driver{ID: fmt.Sprintf("driver_%d", i)},
			Distance: float64(i * 10),
		})
	}

	lat, lon, radius, limit := 43.0, 33.0, 3000.0, 100
	require.NoError(t, cache.SetNearbyDrivers(ctx, lat, lon, radius, limit, drivers, 2*time.Second))

	got, err := cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	require.NoError(t, err)
	assert.Len(t, got, 100)
	assert.Equal(t, "driver_0", got[0].Driver.ID)
	assert.Equal(t, "driver_99", got[99].Driver.ID)
}

// TestRedisDriverCache_GetNearbyDrivers_TTLExpiration tests TTL expiration
// Expected: Should return nil when data has expired
func TestRedisDriverCache_GetNearbyDrivers_TTLExpiration(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	drivers := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "temp_driver"}, Distance: 50},
	}
	lat, lon, radius, limit := 44.0, 34.0, 500.0, 1

	// Set with very short TTL
	require.NoError(t, cache.SetNearbyDrivers(ctx, lat, lon, radius, limit, drivers, 100*time.Millisecond))

	// Should exist immediately
	got, err := cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	require.NoError(t, err)
	assert.Len(t, got, 1)

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired now
	got, err = cache.GetNearbyDrivers(ctx, lat, lon, radius, limit)
	require.NoError(t, err)
	assert.Nil(t, got) // Should return nil after expiration
}

// TestRedisDriverCache_GetNearbyDrivers_KeyGeneration tests different key generation scenarios
// Expected: Should generate different keys for different parameters
func TestRedisDriverCache_GetNearbyDrivers_KeyGeneration(t *testing.T) {
	cache, cleanup := setupRedisTestCache(t)
	defer cleanup()
	ctx := context.Background()

	drivers1 := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "set1_driver"}, Distance: 100},
	}
	drivers2 := []*domain.DriverWithDistance{
		{Driver: domain.Driver{ID: "set2_driver"}, Distance: 200},
	}

	// Set data with different parameters (should generate different keys)
	lat1, lon1, radius1, limit1 := 40.0, 30.0, 1000.0, 5
	lat2, lon2, radius2, limit2 := 40.0, 30.0, 2000.0, 5

	require.NoError(t, cache.SetNearbyDrivers(ctx, lat1, lon1, radius1, limit1, drivers1, time.Minute))
	require.NoError(t, cache.SetNearbyDrivers(ctx, lat2, lon2, radius2, limit2, drivers2, time.Minute))

	// Should get different results for different keys
	got1, err := cache.GetNearbyDrivers(ctx, lat1, lon1, radius1, limit1)
	require.NoError(t, err)
	assert.Equal(t, "set1_driver", got1[0].Driver.ID)

	got2, err := cache.GetNearbyDrivers(ctx, lat2, lon2, radius2, limit2)
	require.NoError(t, err)
	assert.Equal(t, "set2_driver", got2[0].Driver.ID)
}
