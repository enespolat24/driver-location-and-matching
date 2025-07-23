package secondary

import (
	"context"
	"the-matching-service/internal/domain"
)

type DriverLocationService interface {
	FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error)
}
