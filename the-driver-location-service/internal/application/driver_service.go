package application

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	"the-driver-location-service/internal/domain"
	"the-driver-location-service/internal/ports/primary"
	"the-driver-location-service/internal/ports/secondary"
)

type DriverApplicationService struct {
	repo      secondary.DriverRepository
	cache     secondary.DriverCache
	validator *validator.Validate
}

var _ primary.DriverService = (*DriverApplicationService)(nil)

const (
	DriverCacheTTL = 1 * time.Minute // Individual driver cache
	NearbyCacheTTL = 1 * time.Minute // Nearby search cache
)

func NewDriverApplicationService(repo secondary.DriverRepository, cache secondary.DriverCache) *DriverApplicationService {
	return &DriverApplicationService{
		repo:      repo,
		cache:     cache,
		validator: validator.New(),
	}
}

func (s *DriverApplicationService) CreateDriver(req domain.CreateDriverRequest) (*domain.Driver, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	driver := &domain.Driver{
		Location: req.Location,
	}

	if req.ID != "" {
		driver.ID = strings.TrimSpace(req.ID)
	}

	if err := s.repo.Create(driver); err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	ctx := context.Background()
	if s.cache != nil {
		if err := s.cache.Set(ctx, driver.ID, driver, DriverCacheTTL); err != nil {
			fmt.Printf("Warning: failed to cache driver %s: %v\n", driver.ID, err)
		}

		if err := s.cache.InvalidateNearbyCache(ctx); err != nil {
			fmt.Printf("Warning: failed to invalidate nearby cache: %v\n", err)
		}
	}

	return driver, nil
}

// BatchCreateDrivers creates multiple drivers in a batch
func (s *DriverApplicationService) BatchCreateDrivers(req domain.BatchCreateRequest) ([]*domain.Driver, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	drivers := make([]*domain.Driver, len(req.Drivers))
	for i, driverReq := range req.Drivers {
		drivers[i] = &domain.Driver{
			Location: driverReq.Location,
		}

		if driverReq.ID != "" {
			drivers[i].ID = strings.TrimSpace(driverReq.ID)
		}
	}

	if err := s.repo.BatchCreate(drivers); err != nil {
		return nil, fmt.Errorf("failed to batch create drivers: %w", err)
	}

	// For batch operations, just invalidate nearby cache
	ctx := context.Background()
	if s.cache != nil {
		// Invalidate nearby cache since new drivers are added
		if err := s.cache.InvalidateNearbyCache(ctx); err != nil {
			fmt.Printf("Warning: failed to invalidate nearby cache: %v\n", err)
		}
	}

	return drivers, nil
}

func (s *DriverApplicationService) SearchNearbyDrivers(req domain.SearchRequest) ([]*domain.DriverWithDistance, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 10
	}

	ctx := context.Background()

	if s.cache != nil {
		cachedDrivers, err := s.cache.GetNearbyDrivers(ctx, req.Location.Latitude(), req.Location.Longitude(), req.Radius, limit)
		if err != nil {
			fmt.Printf("Warning: failed to get nearby drivers from cache: %v\n", err)
		} else if cachedDrivers != nil {
			return cachedDrivers, nil // Cache hit
		}
	}

	drivers, err := s.repo.SearchNearby(req.Location, req.Radius, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search nearby drivers: %w", err)
	}

	if s.cache != nil {
		if err := s.cache.SetNearbyDrivers(ctx, req.Location.Latitude(), req.Location.Longitude(), req.Radius, limit, drivers, NearbyCacheTTL); err != nil {
			fmt.Printf("Warning: failed to cache nearby drivers: %v\n", err)
		}
	}

	return drivers, nil
}

func (s *DriverApplicationService) GetDriver(id string) (*domain.Driver, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("driver ID is required")
	}

	ctx := context.Background()

	if s.cache != nil {
		cachedDriver, err := s.cache.Get(ctx, id)
		if err != nil {
			fmt.Printf("Warning: failed to get driver from cache: %v\n", err)
		} else if cachedDriver != nil {
			return cachedDriver, nil
		}
	}

	driver, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.Set(ctx, id, driver, DriverCacheTTL); err != nil {
			fmt.Printf("Warning: failed to cache driver: %v\n", err)
		}
	}

	return driver, nil
}

func (s *DriverApplicationService) DeleteDriver(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("driver ID is required")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete driver: %w", err)
	}

	ctx := context.Background()
	if s.cache != nil {
		if err := s.cache.Delete(ctx, id); err != nil {
			fmt.Printf("Warning: failed to delete driver from cache: %v\n", err)
		}
		if err := s.cache.InvalidateNearbyCache(ctx); err != nil {
			fmt.Printf("Warning: failed to invalidate nearby cache: %v\n", err)
		}
	}

	return nil
}

func (s *DriverApplicationService) UpdateDriverLocation(id string, location domain.Point) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("driver ID is required")
	}

	if err := s.validator.Struct(location); err != nil {
		return fmt.Errorf("invalid location: %w", err)
	}

	driver, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}

	driver.Location = location
	driver.UpdatedAt = time.Now()

	if err := s.repo.Update(driver); err != nil {
		return fmt.Errorf("failed to update driver location: %w", err)
	}

	// Invalidate cache
	ctx := context.Background()
	if s.cache != nil {
		if err := s.cache.Delete(ctx, id); err != nil {
			fmt.Printf("Warning: failed to delete driver from cache: %v\n", err)
		}
		if err := s.cache.InvalidateNearbyCache(ctx); err != nil {
			fmt.Printf("Warning: failed to invalidate nearby cache: %v\n", err)
		}
	}

	return nil
}

func (s *DriverApplicationService) UpdateDriver(driver *domain.Driver) error {
	if driver == nil {
		return fmt.Errorf("driver is required")
	}

	if err := s.validator.Struct(driver); err != nil {
		return fmt.Errorf("invalid driver: %w", err)
	}

	if err := s.repo.Update(driver); err != nil {
		return fmt.Errorf("failed to update driver: %w", err)
	}

	// Invalidate cache
	ctx := context.Background()
	if s.cache != nil {
		if err := s.cache.Delete(ctx, driver.ID); err != nil {
			fmt.Printf("Warning: failed to delete driver from cache: %v\n", err)
		}
		if err := s.cache.InvalidateNearbyCache(ctx); err != nil {
			fmt.Printf("Warning: failed to invalidate nearby cache: %v\n", err)
		}
	}

	return nil
}
