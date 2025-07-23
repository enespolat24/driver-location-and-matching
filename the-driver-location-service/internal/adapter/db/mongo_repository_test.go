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
