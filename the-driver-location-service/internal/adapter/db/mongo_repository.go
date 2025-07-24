package db

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"the-driver-location-service/config"
	"the-driver-location-service/internal/domain"
	"the-driver-location-service/internal/ports/secondary"
)

type MongoDriverRepository struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

var _ secondary.DriverRepository = (*MongoDriverRepository)(nil)

func NewMongoDriverRepository(cfg *config.Config) (*MongoDriverRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Database.ConnectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.Database.URI)
	clientOptions.SetMaxPoolSize(cfg.Database.MaxPoolSize)
	clientOptions.SetMinPoolSize(cfg.Database.MinPoolSize)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(cfg.Database.Database)
	collection := database.Collection("drivers")

	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "location", Value: "2dsphere"},
		},
		Options: options.Index().SetName("location_2dsphere"),
	}

	_, err = collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		return nil, fmt.Errorf("failed to create geospatial index: %w", err)
	}

	return &MongoDriverRepository{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

func (r *MongoDriverRepository) Create(driver *domain.Driver) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	now := time.Now()
	driver.CreatedAt = now
	driver.UpdatedAt = now

	if driver.ID == "" {
		driver.ID = primitive.NewObjectID().Hex()
	}

	_, err := r.collection.InsertOne(ctx, driver)
	if err != nil {
		return fmt.Errorf("failed to insert driver: %w", err)
	}

	return nil
}

func (r *MongoDriverRepository) BatchCreate(drivers []*domain.Driver) error {
	if len(drivers) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	documents := make([]interface{}, len(drivers))
	now := time.Now()

	for i, driver := range drivers {
		driver.CreatedAt = now
		driver.UpdatedAt = now

		if driver.ID == "" {
			driver.ID = primitive.NewObjectID().Hex()
		}

		documents[i] = driver
	}

	_, err := r.collection.InsertMany(ctx, documents)
	if err != nil {
		return fmt.Errorf("failed to batch insert drivers: %w", err)
	}

	return nil
}

// https://www.mongodb.com/docs/manual/reference/operator/query/near/
func (r *MongoDriverRepository) SearchNearby(location domain.Point, radiusMeters float64, limit int) ([]*domain.DriverWithDistance, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"location": bson.M{
			"$near": bson.M{
				"$geometry": bson.M{
					"type":        "Point",
					"coordinates": []float64{location.Longitude(), location.Latitude()},
				},
				"$maxDistance": radiusMeters,
			},
		},
	}

	opts := options.Find().SetLimit(int64(limit))

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to search nearby drivers: %w", err)
	}
	defer cursor.Close(ctx)

	var drivers []*domain.Driver
	if err := cursor.All(ctx, &drivers); err != nil {
		return nil, fmt.Errorf("failed to decode drivers: %w", err)
	}

	result := make([]*domain.DriverWithDistance, len(drivers))
	for i, driver := range drivers {
		distance := location.Distance(driver.Location)
		result[i] = &domain.DriverWithDistance{
			Driver:   *driver,
			Distance: distance,
		}
	}

	return result, nil
}

func (r *MongoDriverRepository) GetByID(id string) (*domain.Driver, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var driver domain.Driver
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&driver)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("driver not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get driver: %w", err)
	}

	return &driver, nil
}

func (r *MongoDriverRepository) Update(driver *domain.Driver) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	driver.UpdatedAt = time.Now()

	filter := bson.M{"_id": driver.ID}
	update := bson.M{"$set": driver}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update driver: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("driver not found: %s", driver.ID)
	}

	return nil
}

func (r *MongoDriverRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete driver: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("driver not found: %s", id)
	}

	return nil
}

func (r *MongoDriverRepository) IsEmpty() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return false, fmt.Errorf("failed to count documents: %w", err)
	}

	return count == 0, nil
}

func (r *MongoDriverRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return r.client.Disconnect(ctx)
}
