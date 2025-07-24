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
	IsHealthy(ctx context.Context) bool
}
