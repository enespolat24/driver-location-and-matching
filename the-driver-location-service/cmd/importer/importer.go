package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	httpPackage "the-driver-location-service/internal/adapter/http"
	"the-driver-location-service/internal/domain"
)

const (
	CSV_FILE_PATH = "Coordinates.csv"
	BATCH_SIZE    = 100
	NUM_WORKERS   = 4
)

var (
	apiURL = "http://localhost:8087/api/v1/drivers/batch"
	apiKey = getenvOrDefault("MATCHING_API_KEY", "changeme")
)

type ImportResult struct {
	RequestedCount int
	CreatedCount   int
	ErrorCount     int
}

func getenvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	log.Println("Driver location importer started...")

	result, err := importDataConcurrent()
	if err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	log.Printf("Import completed. Requested: %d, Created: %d, Errors: %d",
		result.RequestedCount, result.CreatedCount, result.ErrorCount)
}

// i implemented worker pool pattern to import data concurrently
// because i was asked about it in the interview
func importDataConcurrent() (*ImportResult, error) {
	file, err := os.Open(CSV_FILE_PATH)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read CSV header: %v", err)
	}

	batchCh := make(chan []domain.CreateDriverRequest, NUM_WORKERS*2)
	resultCh := make(chan ImportResult, NUM_WORKERS*10)
	var wg sync.WaitGroup

	// Start workers
	for i := 0; i < NUM_WORKERS; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for batch := range batchCh {
				batchResult := processBatchHTTP(batch, workerID)
				resultCh <- batchResult
			}
		}(i)
	}

	// Result aggregator goroutine
	var totalRequested, totalCreated, totalErrors int64
	done := make(chan struct{})

	go func() {
		for result := range resultCh {
			atomic.AddInt64(&totalRequested, int64(result.RequestedCount))
			atomic.AddInt64(&totalCreated, int64(result.CreatedCount))
			atomic.AddInt64(&totalErrors, int64(result.ErrorCount))
		}
		close(done)
	}()

	// Read CSV and send batches
	var batch []domain.CreateDriverRequest
	recordCount := 0

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error reading CSV record at line %d: %v", recordCount+2, err) // +2 for header and 1-indexed
			continue
		}

		recordCount++
		driverReq, err := parseDriverLocation(record)
		if err != nil {
			log.Printf("Error parsing driver location from record %d %v: %v", recordCount, record, err)
			continue
		}

		batch = append(batch, driverReq)

		if len(batch) >= BATCH_SIZE {
			batchCh <- batch
			batch = nil
		}
	}

	// Send remaining batch
	if len(batch) > 0 {
		batchCh <- batch
	}

	// Close channels and wait
	close(batchCh)
	wg.Wait()
	close(resultCh)
	<-done

	result := &ImportResult{
		RequestedCount: int(totalRequested),
		CreatedCount:   int(totalCreated),
		ErrorCount:     int(totalErrors),
	}

	log.Printf("CSV processing completed. Total records read: %d", recordCount)
	return result, nil
}

func processBatchHTTP(batch []domain.CreateDriverRequest, workerID int) ImportResult {
	result := ImportResult{
		RequestedCount: len(batch),
		CreatedCount:   0,
		ErrorCount:     0,
	}

	batchReq := domain.BatchCreateRequest{Drivers: batch}
	body, err := json.Marshal(batchReq)
	if err != nil {
		log.Printf("Worker %d: JSON marshal error: %v", workerID, err)
		result.ErrorCount = len(batch)
		return result
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		log.Printf("Worker %d: HTTP request create error: %v", workerID, err)
		result.ErrorCount = len(batch)
		return result
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Worker %d: HTTP request error: %v", workerID, err)
		result.ErrorCount = len(batch)
		return result
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		var apiResp httpPackage.APIResponse
		if err := json.Unmarshal(responseBody.Bytes(), &apiResp); err == nil {
			log.Printf("Worker %d: API error (status %d): %s - %s", workerID, resp.StatusCode, apiResp.Error, apiResp.Message)
		} else {
			log.Printf("Worker %d: API error (status %d): %s", workerID, resp.StatusCode, responseBody.String())
		}
		result.ErrorCount = len(batch)
		return result
	}

	var apiResp httpPackage.APIResponse
	if err := json.Unmarshal(responseBody.Bytes(), &apiResp); err != nil {
		log.Printf("Worker %d: Failed to parse API response: %v", workerID, err)
		result.ErrorCount = len(batch)
		return result
	}

	if !apiResp.Success {
		log.Printf("Worker %d: API operation failed: %s - %s", workerID, apiResp.Error, apiResp.Message)
		result.ErrorCount = len(batch)
		return result
	}

	// Extract actual created count from API response
	var createdCount int
	if data, ok := apiResp.Data.(map[string]interface{}); ok {
		if count, ok := data["count"].(float64); ok {
			createdCount = int(count)
		}
	}

	result.CreatedCount = createdCount

	// Log any discrepancy
	if createdCount != len(batch) {
		log.Printf("Worker %d: Batch discrepancy - requested: %d, created: %d",
			workerID, len(batch), createdCount)
		result.ErrorCount = len(batch) - createdCount
	}

	log.Printf("Worker %d: Batch completed - requested: %d, created: %d",
		workerID, len(batch), createdCount)

	return result
}

func parseDriverLocation(record []string) (domain.CreateDriverRequest, error) {
	if len(record) < 2 {
		return domain.CreateDriverRequest{}, fmt.Errorf("invalid record format: expected at least 2 fields (latitude,longitude), got %d", len(record))
	}

	latitudeStr := record[0]
	longitudeStr := record[1]

	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		return domain.CreateDriverRequest{}, fmt.Errorf("invalid latitude '%s': %w", latitudeStr, err)
	}
	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		return domain.CreateDriverRequest{}, fmt.Errorf("invalid longitude '%s': %w", longitudeStr, err)
	}

	return domain.CreateDriverRequest{
		Location: domain.Point{
			Type:        "Point",
			Coordinates: []float64{longitude, latitude},
		},
	}, nil
}
