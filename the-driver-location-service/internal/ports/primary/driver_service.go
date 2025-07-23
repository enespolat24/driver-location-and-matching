package primary

import "the-driver-location-service/internal/domain"

type DriverService interface {
	CreateDriver(req domain.CreateDriverRequest) (*domain.Driver, error)
	BatchCreateDrivers(req domain.BatchCreateRequest) ([]*domain.Driver, error)
	SearchNearbyDrivers(req domain.SearchRequest) ([]*domain.DriverWithDistance, error)
	GetDriver(id string) (*domain.Driver, error)
	UpdateDriver(driver *domain.Driver) error
	UpdateDriverLocation(id string, location domain.Point) error
	DeleteDriver(id string) error
}
