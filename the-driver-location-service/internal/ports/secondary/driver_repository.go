package secondary

import "the-driver-location-service/internal/domain"

// this is  the interface for driver data operations (Secondary Port)
type DriverRepository interface {
	Create(driver *domain.Driver) error
	BatchCreate(drivers []*domain.Driver) error
	SearchNearby(location domain.Point, radiusMeters float64, limit int) ([]*domain.DriverWithDistance, error)
	GetByID(id string) (*domain.Driver, error)
	Update(driver *domain.Driver) error
	Delete(id string) error
}
