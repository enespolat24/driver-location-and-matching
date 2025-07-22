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

// Ensure DriverApplicationService implements DriverService interface
var _ primary.DriverService = (*DriverApplicationService)(nil)

const (
	DriverCacheTTL = 1 * time.Minute // Individual driver cache
	NearbyCacheTTL = 1 * time.Minute // Nearby search cache
)

// NewDriverApplicationService creates a new driver application service
func NewDriverApplicationService(repo secondary.DriverRepository, cache secondary.DriverCache) *DriverApplicationService {
	return &DriverApplicationService{
		repo:      repo,
		cache:     cache,
		validator: validator.New(),
	}
}

// CreateDriver creates a new driver
func (s *DriverApplicationService) CreateDriver(req domain.CreateDriverRequest) (*domain.Driver, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	driver := &domain.Driver{
		Location: req.Location,
	}

	// Set custom ID if provided
	if req.ID != "" {
		driver.ID = strings.TrimSpace(req.ID)
	}

	if err := s.repo.Create(driver); err != nil {
		return nil, fmt.Errorf("failed to create driver: %w", err)
	}

	// Cache the newly created driver
	ctx := context.Background()
	if s.cache != nil {
		if err := s.cache.Set(ctx, driver.ID, driver, DriverCacheTTL); err != nil {
			// Log error but don't fail the operation
			fmt.Printf("Warning: failed to cache driver %s: %v\n", driver.ID, err)
		}

		// Invalidate nearby cache since a new driver is added
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

	// Convert request to domain models
	drivers := make([]*domain.Driver, len(req.Drivers))
	for i, driverReq := range req.Drivers {
		drivers[i] = &domain.Driver{
			Location: driverReq.Location,
		}

		// Set custom ID if provided
		if driverReq.ID != "" {
			drivers[i].ID = strings.TrimSpace(driverReq.ID)
		}
	}

	if err := s.repo.BatchCreate(drivers); err != nil {
		return nil, fmt.Errorf("failed to batch create drivers: %w", err)
	}

	// For batch operations, just invalidate nearby cache
	// Individual driver caching is not efficient for bulk operations
	ctx := context.Background()
	if s.cache != nil {
		// Invalidate nearby cache since new drivers are added
		if err := s.cache.InvalidateNearbyCache(ctx); err != nil {
			fmt.Printf("Warning: failed to invalidate nearby cache: %v\n", err)
		}
	}

	return drivers, nil
}

// SearchNearbyDrivers searches for drivers near a location with caching
func (s *DriverApplicationService) SearchNearbyDrivers(req domain.SearchRequest) ([]*domain.DriverWithDistance, error) {
	if err := s.validator.Struct(req); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Set default limit if not provided
	limit := req.Limit
	if limit <= 0 {
		limit = 10 // default limit
	}

	ctx := context.Background()

	// Try to get from cache first
	if s.cache != nil {
		cachedDrivers, err := s.cache.GetNearbyDrivers(ctx, req.Location.Latitude(), req.Location.Longitude(), req.Radius, limit)
		if err != nil {
			fmt.Printf("Warning: failed to get nearby drivers from cache: %v\n", err)
		} else if cachedDrivers != nil {
			return cachedDrivers, nil // Cache hit
		}
	}

	// Cache miss, get from repository
	drivers, err := s.repo.SearchNearby(req.Location, req.Radius, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search nearby drivers: %w", err)
	}

	// Cache the result
	if s.cache != nil {
		if err := s.cache.SetNearbyDrivers(ctx, req.Location.Latitude(), req.Location.Longitude(), req.Radius, limit, drivers, NearbyCacheTTL); err != nil {
			fmt.Printf("Warning: failed to cache nearby drivers: %v\n", err)
		}
	}

	return drivers, nil
}

// GetDriver retrieves a driver by ID with caching
func (s *DriverApplicationService) GetDriver(id string) (*domain.Driver, error) {
	if strings.TrimSpace(id) == "" {
		return nil, fmt.Errorf("driver ID is required")
	}

	ctx := context.Background()

	// Try to get from cache first
	if s.cache != nil {
		cachedDriver, err := s.cache.Get(ctx, id)
		if err != nil {
			fmt.Printf("Warning: failed to get driver from cache: %v\n", err)
		} else if cachedDriver != nil {
			return cachedDriver, nil // Cache hit
		}
	}

	// Cache miss, get from repository
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

// DeleteDriver deletes a driver with cache invalidation
func (s *DriverApplicationService) DeleteDriver(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("driver ID is required")
	}

	if err := s.repo.Delete(id); err != nil {
		return fmt.Errorf("failed to delete driver: %w", err)
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

// UpdateDriverLocation updates only the location of a driver
func (s *DriverApplicationService) UpdateDriverLocation(id string, location domain.Point) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("driver ID is required")
	}

	if err := s.validator.Struct(location); err != nil {
		return fmt.Errorf("invalid location: %w", err)
	}

	// Get existing driver
	driver, err := s.repo.GetByID(id)
	if err != nil {
		return fmt.Errorf("failed to get driver: %w", err)
	}

	// Update location
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
