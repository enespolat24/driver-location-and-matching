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
func (m *mockRepo) BatchCreate(drivers []*domain.Driver) error { return nil }
func (m *mockRepo) SearchNearby(location domain.Point, radiusMeters float64, limit int) ([]*domain.DriverWithDistance, error) {
	return nil, nil
}
func (m *mockRepo) GetByID(id string) (*domain.Driver, error) { return nil, nil }
func (m *mockRepo) Update(driver *domain.Driver) error        { return nil }
func (m *mockRepo) Delete(id string) error                    { return nil }

// --- mockCache implementation ---
func (m *mockCache) Set(ctx context.Context, driverID string, driver *domain.Driver, ttl time.Duration) error {
	args := m.Called(ctx, driverID, driver, ttl)
	return args.Error(0)
}
func (m *mockCache) InvalidateNearbyCache(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}
func (m *mockCache) Get(ctx context.Context, driverID string) (*domain.Driver, error) {
	return nil, nil
}
func (m *mockCache) Delete(ctx context.Context, driverID string) error { return nil }
func (m *mockCache) GetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int) ([]*domain.DriverWithDistance, error) {
	return nil, nil
}
func (m *mockCache) SetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int, drivers []*domain.DriverWithDistance, ttl time.Duration) error {
	return nil
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
