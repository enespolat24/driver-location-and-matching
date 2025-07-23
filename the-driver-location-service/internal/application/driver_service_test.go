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
func (m *mockCache) GetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int) ([]*domain.DriverWithDistance, error) {
	args := m.Called(ctx, lat, lon, radius, limit)
	return args.Get(0).([]*domain.DriverWithDistance), args.Error(1)
}
func (m *mockCache) SetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int, drivers []*domain.DriverWithDistance, ttl time.Duration) error {
	args := m.Called(ctx, lat, lon, radius, limit, drivers, ttl)
	return args.Error(0)
}
func (m *mockCache) InvalidateNearbyCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockCache) IsHealthy(ctx context.Context) bool { return true }

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
	cache.On("InvalidateNearbyCache", mock.Anything).Return(nil)

	d, err := service.CreateDriver(req)
	assert.NoError(t, err)
	assert.Equal(t, req.ID, d.ID)
	assert.Equal(t, req.Location, d.Location)

	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestCreateDriver_InvalidRequest(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.CreateDriverRequest{ID: "", Location: domain.Point{}}
	d, err := service.CreateDriver(req)
	assert.Error(t, err)
	assert.Nil(t, d)
}

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

	repo.AssertExpectations(t)
}

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

func TestSearchNearbyDrivers_CacheHit(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	req := domain.SearchRequest{Location: domain.NewPoint(1, 2), Radius: 100, Limit: 5}
	drivers := []*domain.DriverWithDistance{{Driver: domain.Driver{ID: "d1"}, Distance: 10}}
	cache.On("GetNearbyDrivers", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(drivers, nil)
	result, err := service.SearchNearbyDrivers(req)
	assert.NoError(t, err)
	assert.Equal(t, drivers, result)
	cache.AssertExpectations(t)
}

func TestSearchNearbyDrivers_CacheMiss(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	req := domain.SearchRequest{Location: domain.NewPoint(1, 2), Radius: 100, Limit: 5}
	drivers := []*domain.DriverWithDistance{{Driver: domain.Driver{ID: "d2"}, Distance: 20}}
	cache.On("GetNearbyDrivers", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(([]*domain.DriverWithDistance)(nil), nil)
	repo.On("SearchNearby", req.Location, req.Radius, req.Limit).Return(drivers, nil)
	cache.On("SetNearbyDrivers", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, drivers, mock.Anything).Return(nil)
	result, err := service.SearchNearbyDrivers(req)
	assert.NoError(t, err)
	assert.Equal(t, drivers, result)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestUpdateDriverLocation_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}
	newLoc := domain.NewPoint(3, 4)
	repo.On("GetByID", "d1").Return(drv, nil)
	repo.On("Update", mock.Anything).Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(nil)
	cache.On("InvalidateNearbyCache", mock.Anything).Return(nil)
	err := service.UpdateDriverLocation("d1", newLoc)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestDeleteDriver_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	repo.On("Delete", "d1").Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(nil)
	cache.On("InvalidateNearbyCache", mock.Anything).Return(nil)
	err := service.DeleteDriver("d1")
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

func TestUpdateDriver_Success(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)
	drv := &domain.Driver{ID: "d1", Location: domain.NewPoint(1, 2)}
	repo.On("Update", drv).Return(nil)
	cache.On("Delete", mock.Anything, "d1").Return(nil)
	cache.On("InvalidateNearbyCache", mock.Anything).Return(nil)
	err := service.UpdateDriver(drv)
	assert.NoError(t, err)
	repo.AssertExpectations(t)
	cache.AssertExpectations(t)
}

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
	cache.On("InvalidateNearbyCache", mock.Anything).Return(nil).Maybe()

	result, err := service.BatchCreateDrivers(req)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	repo.AssertExpectations(t)
}

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
	repo.AssertExpectations(t)
}

func TestBatchCreateDrivers_InvalidRequest(t *testing.T) {
	repo := new(mockRepo)
	cache := new(mockCache)
	service := NewDriverApplicationService(repo, cache)

	req := domain.BatchCreateRequest{Drivers: nil}
	result, err := service.BatchCreateDrivers(req)
	assert.Error(t, err)
	assert.Nil(t, result)
}
