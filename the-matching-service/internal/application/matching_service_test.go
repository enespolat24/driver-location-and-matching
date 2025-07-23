package application

import (
	"context"
	"errors"
	"testing"

	"the-matching-service/internal/domain"

	"github.com/stretchr/testify/assert"
)

type mockDriverLocationService struct {
	FindNearbyDriversFunc func(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error)
}

func (m *mockDriverLocationService) FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
	return m.FindNearbyDriversFunc(ctx, location, radius)
}

func TestMatchingService_MatchRiderToDriver_success(t *testing.T) {
	mockSvc := &mockDriverLocationService{
		FindNearbyDriversFunc: func(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
			return []domain.DriverDistancePair{
				{
					Driver:   domain.Driver{ID: "driver-1"},
					Distance: 123.45,
				},
			}, nil
		},
	}

	service := NewMatchingService(mockSvc)
	rider := domain.Rider{ID: "rider-1", Name: "Ali", Location: domain.Location{Type: "Point", Coordinates: [2]float64{28.9, 41.0}}}
	result, err := service.MatchRiderToDriver(context.Background(), rider, 500)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "rider-1", result.RiderID)
	assert.Equal(t, "driver-1", result.DriverID)
	assert.Equal(t, 123.45, result.Distance)
}

func TestMatchingService_MatchRiderToDriver_noDrivers(t *testing.T) {
	mockSvc := &mockDriverLocationService{
		FindNearbyDriversFunc: func(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
			return []domain.DriverDistancePair{}, nil
		},
	}

	service := NewMatchingService(mockSvc)
	rider := domain.Rider{ID: "rider-2", Name: "Veli", Location: domain.Location{Type: "Point", Coordinates: [2]float64{29.0, 41.1}}}
	result, err := service.MatchRiderToDriver(context.Background(), rider, 500)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "no drivers found", err.Error())
}

func TestMatchingService_MatchRiderToDriver_serviceError(t *testing.T) {
	mockSvc := &mockDriverLocationService{
		FindNearbyDriversFunc: func(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
			return nil, errors.New("external service error")
		},
	}

	service := NewMatchingService(mockSvc)
	rider := domain.Rider{ID: "rider-3", Name: "Ayşe", Location: domain.Location{Type: "Point", Coordinates: [2]float64{29.1, 41.2}}}
	result, err := service.MatchRiderToDriver(context.Background(), rider, 500)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "external service error", err.Error())
}
