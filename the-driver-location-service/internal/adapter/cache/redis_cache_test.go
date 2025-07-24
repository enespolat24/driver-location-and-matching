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
