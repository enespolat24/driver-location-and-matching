package db

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"the-driver-location-service/internal/adapter/config"
	"the-driver-location-service/internal/domain"
)

func setupMongoTestRepo(t *testing.T) (*MongoDriverRepository, func()) {
	t.Helper()
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:6",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForListeningPort("27017/tcp").WithStartupTimeout(30 * time.Second),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "27017")
	require.NoError(t, err)
	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())

	testCfg := &config.Config{
		Database: config.DatabaseConfig{
			URI:            uri,
			Database:       "testdb",
			ConnectTimeout: 10 * time.Second,
			MaxPoolSize:    10,
			MinPoolSize:    1,
		},
	}

	repo, err := NewMongoDriverRepository(testCfg)
	require.NoError(t, err)

	cleanup := func() {
		repo.Close()
		container.Terminate(ctx)
	}

	return repo, cleanup
}

func TestMongoDriverRepository_CreateAndGetByID(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drv := &domain.Driver{
		ID:       "driver1",
		Location: domain.NewPoint(29.0, 41.0),
	}
	err := repo.Create(drv)
	require.NoError(t, err)

	got, err := repo.GetByID(drv.ID)
	require.NoError(t, err)
	assert.Equal(t, drv.ID, got.ID)
	assert.Equal(t, drv.Location.Longitude(), got.Location.Longitude())
	assert.Equal(t, drv.Location.Latitude(), got.Location.Latitude())
}

func TestMongoDriverRepository_Update(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drv := &domain.Driver{ID: "driver2", Location: domain.NewPoint(10, 10)}
	require.NoError(t, repo.Create(drv))

	drv.Location = domain.NewPoint(20, 20)
	require.NoError(t, repo.Update(drv))

	got, err := repo.GetByID(drv.ID)
	require.NoError(t, err)
	assert.Equal(t, 20.0, got.Location.Longitude())
	assert.Equal(t, 20.0, got.Location.Latitude())
}

func TestMongoDriverRepository_Delete(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drv := &domain.Driver{ID: "driver3", Location: domain.NewPoint(30, 30)}
	require.NoError(t, repo.Create(drv))
	require.NoError(t, repo.Delete(drv.ID))
	_, err := repo.GetByID(drv.ID)
	assert.Error(t, err)
}

func TestMongoDriverRepository_BatchCreate(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drivers := []*domain.Driver{
		{ID: "b1", Location: domain.NewPoint(1, 1)},
		{ID: "b2", Location: domain.NewPoint(2, 2)},
	}
	require.NoError(t, repo.BatchCreate(drivers))
	for _, d := range drivers {
		got, err := repo.GetByID(d.ID)
		require.NoError(t, err)
		assert.Equal(t, d.Location.Longitude(), got.Location.Longitude())
	}
}

func TestMongoDriverRepository_SearchNearby(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drivers := []*domain.Driver{
		{ID: "s1", Location: domain.NewPoint(10, 10)},
		{ID: "s2", Location: domain.NewPoint(10.001, 10.001)},
		{ID: "s3", Location: domain.NewPoint(20, 20)},
	}
	require.NoError(t, repo.BatchCreate(drivers))

	center := domain.NewPoint(10, 10)
	// 200m radius should find s1 and s2, but not s3
	found, err := repo.SearchNearby(center, 200, 10)
	require.NoError(t, err)
	ids := make([]string, 0, len(found))
	for _, d := range found {
		ids = append(ids, d.Driver.ID)
	}
	assert.Contains(t, ids, "s1")
	assert.Contains(t, ids, "s2")
	assert.NotContains(t, ids, "s3")
}
