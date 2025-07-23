package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"the-driver-location-service/internal/adapter/config"
	"the-driver-location-service/internal/adapter/db"
	"the-driver-location-service/internal/application"
	"the-driver-location-service/internal/domain"
)

const (
	CSV_FILE_PATH = "Coordinates.csv"
	BATCH_SIZE    = 100
)

func main() {
	log.Println("Driver location importer started...")

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	driverRepo, err := db.NewMongoDriverRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB repository: %v", err)
	}

	isEmpty, err := driverRepo.IsEmpty()
	if err != nil {
		log.Printf("Failed to check if collection is empty: %v", err)
	} else if !isEmpty {
		log.Println("Collection is not empty. Data already imported. Exiting...")
		return
	}

	driverService := application.NewDriverApplicationService(driverRepo, nil)

	if err := importData(driverService); err != nil {
		log.Fatalf("Import failed: %v", err)
	}

	log.Println("Driver location import completed successfully.")
}

func importData(service *application.DriverApplicationService) error {
	file, err := os.Open(CSV_FILE_PATH)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	if _, err := reader.Read(); err != nil {
		return fmt.Errorf("failed to read CSV header: %v", err)
	}

	var batch []domain.CreateDriverRequest
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
			if err := processBatch(service, batch); err != nil {
				log.Printf("Error processing batch: %v", err)
				errorCount += len(batch)
			} else {
				successCount += len(batch)
			}
			batch = nil
		}
	}

	if len(batch) > 0 {
		if err := processBatch(service, batch); err != nil {
			log.Printf("Error processing final batch: %v", err)
			errorCount += len(batch)
		} else {
			successCount += len(batch)
		}
	}

	log.Printf("Import completed. Success: %d, Errors: %d", successCount, errorCount)
	return nil
}

func processBatch(service *application.DriverApplicationService, batch []domain.CreateDriverRequest) error {
	batchReq := domain.BatchCreateRequest{Drivers: batch}

	_, err := service.BatchCreateDrivers(batchReq)
	if err != nil {
		return fmt.Errorf("batch create failed: %w", err)
	}

	log.Printf("Successfully processed batch of %d drivers", len(batch))
	return nil
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
		Location: domain.NewPoint(longitude, latitude),
	}, nil
}
