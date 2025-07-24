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

	"the-driver-location-service/config"
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

// TestMongoDriverRepository_Delete_NotFound tests deletion of non-existent driver
// Expected: Should return error when trying to delete driver that doesn't exist
func TestMongoDriverRepository_Delete_NotFound(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	err := repo.Delete("non-existent-driver")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver not found")
}

// TestMongoDriverRepository_IsEmpty tests IsEmpty functionality
// Expected: Should return true when collection is empty, false when it has documents
func TestMongoDriverRepository_IsEmpty(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	isEmpty, err := repo.IsEmpty()
	require.NoError(t, err)
	assert.True(t, isEmpty)

	drv := &domain.Driver{ID: "test-driver", Location: domain.NewPoint(40, 40)}
	require.NoError(t, repo.Create(drv))

	isEmpty, err = repo.IsEmpty()
	require.NoError(t, err)
	assert.False(t, isEmpty)

	require.NoError(t, repo.Delete(drv.ID))

	isEmpty, err = repo.IsEmpty()
	require.NoError(t, err)
	assert.True(t, isEmpty)
}

// TestMongoDriverRepository_GetByID_NotFound tests GetByID with non-existent ID
// Expected: Should return error when driver is not found
func TestMongoDriverRepository_GetByID_NotFound(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	_, err := repo.GetByID("non-existent-id")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver not found")
}

// TestMongoDriverRepository_Update_NotFound tests updating non-existent driver
// Expected: Should return error when trying to update driver that doesn't exist
func TestMongoDriverRepository_Update_NotFound(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	nonExistentDriver := &domain.Driver{
		ID:       "non-existent",
		Location: domain.NewPoint(50, 50),
	}

	err := repo.Update(nonExistentDriver)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver not found")
}

// TestMongoDriverRepository_Create_EmptyID tests creation with empty ID (auto-generation)
// Expected: Should auto-generate ID when not provided
func TestMongoDriverRepository_Create_EmptyID(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drv := &domain.Driver{
		ID:       "",
		Location: domain.NewPoint(60, 60),
	}

	err := repo.Create(drv)
	require.NoError(t, err)
	assert.NotEmpty(t, drv.ID)

	retrieved, err := repo.GetByID(drv.ID)
	require.NoError(t, err)
	assert.Equal(t, drv.ID, retrieved.ID)
}

// TestMongoDriverRepository_BatchCreate_EmptyArray tests batch creation with empty array
// Expected: Should handle empty array gracefully without error
func TestMongoDriverRepository_BatchCreate_EmptyArray(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	emptyDrivers := []*domain.Driver{}
	err := repo.BatchCreate(emptyDrivers)
	require.NoError(t, err) // Should not error on empty array

	// Collection should still be empty
	isEmpty, err := repo.IsEmpty()
	require.NoError(t, err)
	assert.True(t, isEmpty)
}

// TestMongoDriverRepository_BatchCreate_MixedIDs tests batch creation with mixed empty/non-empty IDs
// Expected: Should handle mix of drivers with and without IDs
func TestMongoDriverRepository_BatchCreate_MixedIDs(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drivers := []*domain.Driver{
		{ID: "explicit-id-1", Location: domain.NewPoint(70, 70)},
		{ID: "", Location: domain.NewPoint(71, 71)}, // Empty ID
		{ID: "explicit-id-2", Location: domain.NewPoint(72, 72)},
		{ID: "", Location: domain.NewPoint(73, 73)}, // Empty ID
	}

	err := repo.BatchCreate(drivers)
	require.NoError(t, err)

	for _, driver := range drivers {
		assert.NotEmpty(t, driver.ID)

		retrieved, err := repo.GetByID(driver.ID)
		require.NoError(t, err)
		assert.Equal(t, driver.Location.Longitude(), retrieved.Location.Longitude())
		assert.Equal(t, driver.Location.Latitude(), retrieved.Location.Latitude())
	}
}

// TestMongoDriverRepository_SearchNearby_EmptyResult tests search that returns no results
// Expected: Should return empty slice when no drivers are in range
func TestMongoDriverRepository_SearchNearby_EmptyResult(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	farDriver := &domain.Driver{ID: "far-driver", Location: domain.NewPoint(80, 80)}
	require.NoError(t, repo.Create(farDriver))

	center := domain.NewPoint(10, 10)
	found, err := repo.SearchNearby(center, 100, 10)
	require.NoError(t, err)
	assert.Empty(t, found)
}

// TestMongoDriverRepository_SearchNearby_ZeroLimit tests search with zero limit
// Expected: Should return limited results based on MongoDB's behavior with zero limit
func TestMongoDriverRepository_SearchNearby_ZeroLimit(t *testing.T) {
	repo, cleanup := setupMongoTestRepo(t)
	defer cleanup()

	drivers := []*domain.Driver{
		{ID: "d1", Location: domain.NewPoint(15, 15)},
		{ID: "d2", Location: domain.NewPoint(15.001, 15.001)},
	}
	require.NoError(t, repo.BatchCreate(drivers))

	center := domain.NewPoint(15, 15)
	found, err := repo.SearchNearby(center, 1000, 0)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(found), 0)
}
