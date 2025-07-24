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

func getenvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	log.Println("Driver location importer started...")

	if err := importDataConcurrent(); err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	log.Println("Driver location import completed successfully.")
}

// i implemented worker pool pattern to import data concurrently
// because i was asked about it in the interview
func importDataConcurrent() error {
	file, err := os.Open(CSV_FILE_PATH)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read CSV header: %v", err)
	}

	batchCh := make(chan []CreateDriverRequest)
	errCh := make(chan error, 100)
	var wg sync.WaitGroup

	for i := 0; i < NUM_WORKERS; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for batch := range batchCh {
				if err := processBatchHTTP(batch); err != nil {
					log.Printf("Worker %d: Error processing batch: %v", workerID, err)
					errCh <- err
				}
			}
		}(i)
	}

	var batch []CreateDriverRequest
	successCount := 0
	errorCount := 0

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error reading CSV record: %v", err)
			errorCount++
			continue
		}

		driverReq, err := parseDriverLocation(record)
		if err != nil {
			log.Printf("Error parsing driver location from record %v: %v", record, err)
			errorCount++
			continue
		}

		batch = append(batch, driverReq)

		if len(batch) >= BATCH_SIZE {
			batchCh <- batch
			successCount += len(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		batchCh <- batch
		successCount += len(batch)
	}

	close(batchCh)
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			errorCount++
		}
	}

	log.Printf("Import completed. Success: %d, Errors: %d", successCount, errorCount)
	return nil
}

type Point struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type CreateDriverRequest struct {
	ID       string `json:"id,omitempty"`
	Location Point  `json:"location"`
}

type BatchCreateRequest struct {
	Drivers []CreateDriverRequest `json:"drivers"`
}

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

func processBatchHTTP(batch []CreateDriverRequest) error {
	batchReq := BatchCreateRequest{Drivers: batch}
	body, err := json.Marshal(batchReq)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("http request create error: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("X-API-KEY", apiKey)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	var responseBody bytes.Buffer
	responseBody.ReadFrom(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		var apiResp APIResponse
		if err := json.Unmarshal(responseBody.Bytes(), &apiResp); err == nil {
			return fmt.Errorf("API error (status %d): %s - %s", resp.StatusCode, apiResp.Error, apiResp.Message)
		}
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, responseBody.String())
	}

	var apiResp APIResponse
	if err := json.Unmarshal(responseBody.Bytes(), &apiResp); err != nil {
		return fmt.Errorf("failed to parse API response: %w", err)
	}

	if !apiResp.Success {
		return fmt.Errorf("API operation failed: %s - %s", apiResp.Error, apiResp.Message)
	}

	var createdCount int
	if data, ok := apiResp.Data.(map[string]interface{}); ok {
		if count, ok := data["count"].(float64); ok {
			createdCount = int(count)
		}
	}

	log.Printf("Batch import success, requested: %d, created: %d", len(batch), createdCount)
	return nil
}

func parseDriverLocation(record []string) (CreateDriverRequest, error) {
	if len(record) < 2 {
		return CreateDriverRequest{}, fmt.Errorf("invalid record format: expected at least 2 fields (latitude,longitude), got %d", len(record))
	}

	latitudeStr := record[0]
	longitudeStr := record[1]

	latitude, err := strconv.ParseFloat(latitudeStr, 64)
	if err != nil {
		return CreateDriverRequest{}, fmt.Errorf("invalid latitude '%s': %w", latitudeStr, err)
	}
	longitude, err := strconv.ParseFloat(longitudeStr, 64)
	if err != nil {
		return CreateDriverRequest{}, fmt.Errorf("invalid longitude '%s': %w", longitudeStr, err)
	}

	return CreateDriverRequest{
		Location: Point{
			Type:        "Point",
			Coordinates: []float64{longitude, latitude},
		},
	}, nil
}
