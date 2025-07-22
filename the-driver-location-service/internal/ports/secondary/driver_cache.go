package secondary

import (
	"context"
	"time"

	"the-driver-location-service/internal/domain"
)

type DriverCache interface {
	Get(ctx context.Context, driverID string) (*domain.Driver, error)
	Set(ctx context.Context, driverID string, driver *domain.Driver, ttl time.Duration) error
	Delete(ctx context.Context, driverID string) error
	GetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int) ([]*domain.DriverWithDistance, error)
	SetNearbyDrivers(ctx context.Context, lat, lon, radius float64, limit int, drivers []*domain.DriverWithDistance, ttl time.Duration) error
	InvalidateNearbyCache(ctx context.Context) error
	IsHealthy(ctx context.Context) bool
}
