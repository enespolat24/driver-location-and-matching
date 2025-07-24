package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"the-matching-service/internal/domain"

	"github.com/sony/gobreaker"
)

type DriverLocationClient struct {
	baseURL    string
	httpClient *http.Client
	breaker    *gobreaker.CircuitBreaker
	apiKey     string
}

func NewDriverLocationClient(baseURL, apiKey string) *DriverLocationClient {
	cbSettings := gobreaker.Settings{
		Name:        "DriverLocationService",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     10 * time.Second,
	}
	return &DriverLocationClient{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		breaker:    gobreaker.NewCircuitBreaker(cbSettings),
		apiKey:     apiKey,
	}
}

func (c *DriverLocationClient) FindNearbyDrivers(ctx context.Context, location domain.Location, radius float64) ([]domain.DriverDistancePair, error) {
	requestBody := map[string]interface{}{
		"location": location,
		"radius":   radius,
	}
	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, err
	}

	var resp *http.Response
	result, err := c.breaker.Execute(func() (interface{}, error) {
		req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/v1/drivers/search", bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		if c.apiKey != "" {
			req.Header.Set("X-API-Key", c.apiKey)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}
		if resp.StatusCode != http.StatusOK {
			b, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(b))
		}
		return resp, nil
	})
	if err != nil {
		return nil, err
	}

	resp, ok := result.(*http.Response)
	if !ok || resp == nil {
		return nil, fmt.Errorf("invalid response type from circuit breaker")
	}
	defer resp.Body.Close()

	var serviceResp domain.DriverLocationServiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&serviceResp); err != nil {
		return nil, err
	}

	if !serviceResp.Success {
		return nil, fmt.Errorf("driver location service error: %s - %s", serviceResp.Error, serviceResp.Message)
	}

	data, ok := serviceResp.Data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid response data format from driver location service")
	}

	driversData, ok := data["drivers"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid drivers data format from driver location service")
	}

	var drivers []domain.DriverDistancePair
	driversBytes, err := json.Marshal(driversData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal drivers data: %w", err)
	}

	if err := json.Unmarshal(driversBytes, &drivers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal drivers: %w", err)
	}

	return drivers, nil
}
