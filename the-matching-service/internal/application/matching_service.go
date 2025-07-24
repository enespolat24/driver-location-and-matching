package application

import (
	"context"
	"errors"
	"math"

	"the-matching-service/internal/domain"
	"the-matching-service/internal/ports/secondary"
)

type MatchingService struct {
	DriverLocationService secondary.DriverLocationService
}

func NewMatchingService(driverLocationService secondary.DriverLocationService) *MatchingService {
	return &MatchingService{
		DriverLocationService: driverLocationService,
	}
}

// MatchRiderToDriver finds the nearest driver for the given rider and radius
func (s *MatchingService) MatchRiderToDriver(ctx context.Context, rider domain.Rider, radius float64) (*domain.MatchResult, error) {
	drivers, err := s.DriverLocationService.FindNearbyDrivers(ctx, rider.Location, radius)
	if err != nil {
		return nil, err
	}
	if len(drivers) == 0 {
		return nil, errors.New("no drivers found")
	}
	// En yakın sürücü ilk sırada varsayılıyor
	nearest := drivers[0]
	return &domain.MatchResult{
		RiderID:  rider.ID,
		DriverID: nearest.Driver.ID,
		Distance: math.Round(nearest.Distance*100) / 100,
	}, nil
}
