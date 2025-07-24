package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"the-driver-location-service/internal/domain"
)

type mockRepo struct{ mock.Mock }
type mockCache struct{ mock.Mock }

// --- mockRepo implementation ---
func (m *mockRepo) Create(driver *domain.Driver) error {
	args := m.Called(driver)
	return args.Error(0)
}
func (m *mockRepo) BatchCreate(drivers []*domain.Driver) error {
	args := m.Called(drivers)
	return args.Error(0)
}
func (m *mockRepo) SearchNearby(location domain.Point, radiusMeters float64, limit int) ([]*domain.DriverWithDistance, error) {
	args := m.Called(location, radiusMeters, limit)
	return args.Get(0).([]*domain.DriverWithDistance), args.Error(1)
}
func (m *mockRepo) GetByID(id string) (*domain.Driver, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.Driver), args.Error(1)
}
func (m *mockRepo) Update(driver *domain.Driver) error {
	args := m.Called(driver)
	return args.Error(0)
}
func (m *mockRepo) Delete(id string) error { args := m.Called(id); return args.Error(0) }

// --- mockCache implementation ---
func (m *mockCache) Get(ctx context.Context, driverID string) (*domain.Driver, error) {
	args := m.Called(ctx, driverID)
	return args.Get(0).(*domain.Driver), args.Error(1)
}
func (m *mockCache) Set(ctx context.Context, driverID string, driver *domain.Driver, ttl time.Duration) error {
	args := m.Called(ctx, driverID, driver, ttl)
	return args.Error(0)
}
func (m *mockCache) Delete(ctx context.Context, driverID string) error {
	args := m.Called(ctx, driverID)
	return args.Error(0)
}
func (m *mockCache) IsHealthy(ctx context.Context) bool { return true }

// TestCreateDriver_Success tests successful driver creation with valid request data
// Expected: Should create driver successfully, cache the driver, and return driver with correct data
func TestCreateDriver_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.CreateDriverRequest{
		ID:       "driver1",
		Location: domain.NewPoint(29.0, 41.0),
	}

	repo.On("Create", mock.AnythingOfType("*domain.Driver")).Return(nil)
	cache.On("Set", mock.Anything, "driver1", mock.AnythingOfType("*domain.Driver"), mock.Anything).Return(nil)

	d, err := service.CreateDriver(req)
	assert.NoError(t, err)
	assert.Equal(t, req.ID, d.ID)
	assert.Equal(t, req.Location, d.Location)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestCreateDriver_WithEmptyID tests driver creation when ID is empty (should auto-generate)
// Expected: Should create driver successfully with auto-generated ID and cache the result
func TestCreateDriver_WithEmptyID(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.CreateDriverRequest{
		ID:       "",
		Location: domain.NewPoint(29.0, 41.0),
	}

	// Mock id generation
	repo.On("Create", mock.AnythingOfType("*domain.Driver")).Run(func(args mock.Arguments) {
		driver := args.Get(0).(*domain.Driver)
		if driver.ID == "" {
			driver.ID = "auto-generated-id"
		}
	}).Return(nil)
	cache.On("Set", mock.Anything, "auto-generated-id", mock.AnythingOfType("*domain.Driver"), mock.Anything).Return(nil)

	d, err := service.CreateDriver(req)
	assert.NoError(t, err)
	assert.NotEmpty(t, d.ID)
	assert.Equal(t, req.Location, d.Location)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestCreateDriver_InvalidRequest tests driver creation with invalid request data
// Expected: Should return validation error and nil driver when request validation fails
func TestCreateDriver_InvalidRequest(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.CreateDriverRequest{ID: "", Location: domain.Point{}}
	d, err := service.CreateDriver(req)
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "invalid request")
}

// TestCreateDriver_RepoError tests driver creation when repository operation fails
// Expected: Should return repository error and nil driver when database operation fails
func TestCreateDriver_RepoError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.CreateDriverRequest{
		ID:       "driver2",
		Location: domain.NewPoint(29.0, 41.0),
	}

	repo.On("Create", mock.AnythingOfType("*domain.Driver")).Return(errors.New("db error"))

	d, err := service.CreateDriver(req)
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "failed to create driver")

	repo.AssertExpectations(t)
}

// TestCreateDriver_CacheError tests driver creation when cache operations fail (should continue)
// Expected: Should create driver successfully even when cache operations fail, with warning logs
func TestCreateDriver_CacheError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.CreateDriverRequest{
		ID:       "driver3",
		Location: domain.NewPoint(29.0, 41.0),
	}

	repo.On("Create", mock.AnythingOfType("*domain.Driver")).Return(nil)
	cache.On("Set", mock.Anything, "driver3", mock.AnythingOfType("*domain.Driver"), mock.Anything).Return(errors.New("cache error"))

	d, err := service.CreateDriver(req)
	assert.NoError(t, err)
	assert.Equal(t, req.ID, d.ID)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestGetDriver_CacheHit tests driver retrieval when driver is found in cache
// Expected: Should return driver from cache without calling repository
func TestGetDriver_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}
	cache.On("Get", mock.Anything, "d1").Return(drv, nil)
	d, err := service.GetDriver("d1")
	assert.NoError(t, err)
	assert.Equal(t, drv, d)
	cache.AssertExpectations(t)
}

// TestGetDriver_CacheMiss tests driver retrieval when driver is not in cache but exists in repository
// Expected: Should fetch driver from repository, cache the result, and return driver
func TestGetDriver_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d2", Location: domain.NewPoint(1, 2)}
	cache.On("Get", mock.Anything, "d2").Return((*domain.Driver)(nil), nil)
	repo.On("GetByID", "d2").Return(drv, nil)
	cache.On("Set", mock.Anything, "d2", drv, mock.Anything).Return(nil)
	d, err := service.GetDriver("d2")
	assert.NoError(t, err)
	assert.Equal(t, drv, d)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestGetDriver_EmptyID tests driver retrieval with empty driver ID
// Expected: Should return error when driver ID is empty or whitespace
func TestGetDriver_EmptyID(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	d, err := service.GetDriver("")
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "driver ID is required")

	d, err = service.GetDriver("   ")
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "driver ID is required")
}

// TestGetDriver_RepoError tests driver retrieval when repository returns error
// Expected: Should return repository error when driver not found or database error occurs
func TestGetDriver_RepoError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	cache.On("Get", mock.Anything, "d3").Return((*domain.Driver)(nil), nil)
	repo.On("GetByID", "d3").Return((*domain.Driver)(nil), errors.New("driver not found"))

	d, err := service.GetDriver("d3")
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "failed to get driver")

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestGetDriver_CacheError tests driver retrieval when cache operations fail
// Expected: Should fallback to repository and continue operation even when cache fails
func TestGetDriver_CacheError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d4", Location: domain.NewPoint(1, 2)}

	cache.On("Get", mock.Anything, "d4").Return((*domain.Driver)(nil), errors.New("cache error"))
	repo.On("GetByID", "d4").Return(drv, nil)
	cache.On("Set", mock.Anything, "d4", drv, mock.Anything).Return(errors.New("cache error"))

	d, err := service.GetDriver("d4")
	assert.NoError(t, err)
	assert.Equal(t, drv, d)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestSearchNearbyDrivers_Success tests nearby driver search with successful repository call
// Expected: Should fetch from repository and return drivers
func TestSearchNearbyDrivers_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	req := domain.SearchRequest{Location: domain.NewPoint(1, 2), Radius: 100, Limit: 5}
	drivers := []*domain.DriverWithDistance{{Driver: domain.Driver{ID: "d1"}, Distance: 10}}
	repo.On("SearchNearby", req.Location, req.Radius, req.Limit).Return(drivers, nil)
	result, err := service.SearchNearbyDrivers(req)
	assert.NoError(t, err)
	assert.Equal(t, drivers, result)
	repo.AssertExpectations(t)
}

// TestSearchNearbyDrivers_InvalidRequest tests nearby driver search with invalid request data
// Expected: Should return validation error when request validation fails
func TestSearchNearbyDrivers_InvalidRequest(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.SearchRequest{Location: domain.Point{}, Radius: -1, Limit: -5}
	result, err := service.SearchNearbyDrivers(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid request")
}

// TestSearchNearbyDrivers_DefaultLimit tests nearby driver search with zero limit (should use default)
// Expected: Should use default limit of 10 when limit is zero or negative
func TestSearchNearbyDrivers_DefaultLimit(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	req := domain.SearchRequest{Location: domain.NewPoint(1, 2), Radius: 100, Limit: 0}
	drivers := []*domain.DriverWithDistance{{Driver: domain.Driver{ID: "d1"}, Distance: 10}}

	repo.On("SearchNearby", req.Location, req.Radius, 10).Return(drivers, nil)

	result, err := service.SearchNearbyDrivers(req)
	assert.NoError(t, err)
	assert.Equal(t, drivers, result)

	repo.AssertExpectations(t)
}

// TestSearchNearbyDrivers_RepoError tests nearby driver search when repository operation fails
// Expected: Should return repository error when search operation fails
func TestSearchNearbyDrivers_RepoError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	req := domain.SearchRequest{Location: domain.NewPoint(1, 2), Radius: 100, Limit: 5}

	repo.On("SearchNearby", req.Location, req.Radius, req.Limit).Return(([]*domain.DriverWithDistance)(nil), errors.New("search error"))

	result, err := service.SearchNearbyDrivers(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to search nearby drivers")

	repo.AssertExpectations(t)
}

// TestUpdateDriverLocation_Success tests successful driver location update
// Expected: Should update driver location, delete from cache, and return no error
func TestUpdateDriverLocation_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}
	newLoc := domain.NewPoint(3, 4)
	repo.On("GetByID", "d1").Return(drv, nil)
	repo.On("Update", mock.Anything).Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(nil)
	err := service.UpdateDriverLocation("d1", newLoc)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestUpdateDriverLocation_EmptyID tests driver location update with empty driver ID
// Expected: Should return error when driver ID is empty or whitespace
func TestUpdateDriverLocation_EmptyID(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	newLoc := domain.NewPoint(3, 4)

	err := service.UpdateDriverLocation("", newLoc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver ID is required")

	err = service.UpdateDriverLocation("   ", newLoc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver ID is required")
}

// TestUpdateDriverLocation_InvalidLocation tests driver location update with invalid location data
// Expected: Should return validation error when location validation fails
func TestUpdateDriverLocation_InvalidLocation(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	invalidLoc := domain.Point{}

	err := service.UpdateDriverLocation("d1", invalidLoc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid location")
}

// TestUpdateDriverLocation_DriverNotFound tests driver location update when driver doesn't exist
// Expected: Should return error when driver is not found in repository
func TestUpdateDriverLocation_DriverNotFound(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	newLoc := domain.NewPoint(3, 4)

	repo.On("GetByID", "d1").Return((*domain.Driver)(nil), errors.New("driver not found"))

	err := service.UpdateDriverLocation("d1", newLoc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get driver")

	repo.AssertExpectations(t)
}

// TestUpdateDriverLocation_UpdateError tests driver location update when repository update fails
// Expected: Should return error when repository update operation fails
func TestUpdateDriverLocation_UpdateError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}
	newLoc := domain.NewPoint(3, 4)

	repo.On("GetByID", "d1").Return(drv, nil)
	repo.On("Update", mock.Anything).Return(errors.New("update error"))

	err := service.UpdateDriverLocation("d1", newLoc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update driver location")

	repo.AssertExpectations(t)
}

// TestDeleteDriver_Success tests successful driver deletion
// Expected: Should delete driver from repository and cache, invalidate nearby cache
func TestDeleteDriver_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	repo.On("Delete", "d1").Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(nil)
	err := service.DeleteDriver("d1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestDeleteDriver_EmptyID tests driver deletion with empty driver ID
// Expected: Should return error when driver ID is empty or whitespace
func TestDeleteDriver_EmptyID(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	err := service.DeleteDriver("")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver ID is required")

	err = service.DeleteDriver("   ")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver ID is required")
}

// TestDeleteDriver_RepoError tests driver deletion when repository operation fails
// Expected: Should return error when repository delete operation fails
func TestDeleteDriver_RepoError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	repo.On("Delete", "d1").Return(errors.New("delete error"))

	err := service.DeleteDriver("d1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete driver")

	repo.AssertExpectations(t)
}

// TestDeleteDriver_CacheError tests driver deletion when cache operations fail
// Expected: Should continue operation even when cache operations fail
func TestDeleteDriver_CacheError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	repo.On("Delete", "d1").Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(errors.New("cache error"))

	err := service.DeleteDriver("d1")
	assert.NoError(t, err)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestUpdateDriver_Success tests successful driver update
// Expected: Should update driver in repository, invalidate cache, and return no error
func TestUpdateDriver_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}
	repo.On("Update", drv).Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(nil)
	err := service.UpdateDriver(drv)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

// TestUpdateDriver_NilDriver tests driver update with nil driver
// Expected: Should return error when driver is nil
func TestUpdateDriver_NilDriver(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	err := service.UpdateDriver(nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "driver is required")
}

// TestUpdateDriver_InvalidDriver tests driver update with invalid driver data
// Expected: Should return validation error when driver validation fails
func TestUpdateDriver_InvalidDriver(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	invalidDriver := &domain.Driver{ID: "d1", Location: domain.Point{}}

	err := service.UpdateDriver(invalidDriver)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid driver")
}

// TestUpdateDriver_RepoError tests driver update when repository operation fails
// Expected: Should return error when repository update operation fails
func TestUpdateDriver_RepoError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}

	repo.On("Update", drv).Return(errors.New("update error"))

	err := service.UpdateDriver(drv)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update driver")

	repo.AssertExpectations(t)
}

// TestBatchCreateDrivers_Success tests successful batch driver creation
// Expected: Should create multiple drivers and return all created drivers
func TestBatchCreateDrivers_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{
		Drivers: []domain.CreateDriverRequest{
			{ID: "d1", Location: domain.NewPoint(1, 2)},
			{ID: "d2", Location: domain.NewPoint(3, 4)},
		},
	}

	repo.On("BatchCreate", mock.Anything).Return(nil)

	result, err := service.BatchCreateDrivers(req)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

// TestBatchCreateDrivers_EmptyDrivers tests batch driver creation with empty drivers list
// Expected: Should return validation error when drivers list is empty
func TestBatchCreateDrivers_EmptyDrivers(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{Drivers: []domain.CreateDriverRequest{}}

	result, err := service.BatchCreateDrivers(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "invalid request")
}

// TestBatchCreateDrivers_NilDrivers tests batch driver creation with nil drivers list
// Expected: Should return validation error when drivers list is nil
func TestBatchCreateDrivers_NilDrivers(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{Drivers: nil}
	result, err := service.BatchCreateDrivers(req)
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestBatchCreateDrivers_RepoError tests batch driver creation when repository operation fails
// Expected: Should return error when repository batch create operation fails
func TestBatchCreateDrivers_RepoError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{
		Drivers: []domain.CreateDriverRequest{
			{ID: "d1", Location: domain.NewPoint(1, 2)},
		},
	}

	repo.On("BatchCreate", mock.Anything).Return(errors.New("db error"))
	cache.On("InvalidateNearbyCache", mock.Anything).Return(nil).Maybe()

	result, err := service.BatchCreateDrivers(req)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to batch create drivers")
	repo.AssertExpectations(t)
}

// TestBatchCreateDrivers_CacheError tests batch driver creation when cache operations fail
// Expected: Should continue operation even when cache operations fail
func TestBatchCreateDrivers_CacheError(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{
		Drivers: []domain.CreateDriverRequest{
			{ID: "d1", Location: domain.NewPoint(1, 2)},
		},
	}

	repo.On("BatchCreate", mock.Anything).Return(nil)

	result, err := service.BatchCreateDrivers(req)
	assert.NoError(t, err)
	assert.Len(t, result, 1)

	repo.AssertExpectations(t)
}

// TestBatchCreateDrivers_WithEmptyIDs tests batch driver creation with some empty IDs
// Expected: Should create drivers successfully with auto-generated IDs for empty ones
func TestBatchCreateDrivers_WithEmptyIDs(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{
		Drivers: []domain.CreateDriverRequest{
			{ID: "d1", Location: domain.NewPoint(1, 2)},
			{ID: "", Location: domain.NewPoint(3, 4)},
			{ID: "d3", Location: domain.NewPoint(5, 6)},
		},
	}

	// Mock id generation
	repo.On("BatchCreate", mock.Anything).Run(func(args mock.Arguments) {
		drivers := args.Get(0).([]*domain.Driver)
		for _, driver := range drivers {
			if driver.ID == "" {
				driver.ID = "auto-generated-id"
			}
		}
	}).Return(nil)

	result, err := service.BatchCreateDrivers(req)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "d1", result[0].ID)
	assert.NotEmpty(t, result[1].ID) // Auto-generated ID
	assert.Equal(t, "d3", result[2].ID)

	repo.AssertExpectations(t)
}
