package domain

import (
	"time"
)

type Point struct {
	Type        string    `json:"type" bson:"type" validate:"required,eq=Point"`
	Coordinates []float64 `json:"coordinates" bson:"coordinates" validate:"required,len=2,dive"`
}
type Driver struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Location  Point     `json:"location" bson:"location" validate:"required"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type DriverWithDistance struct {
	Driver   Driver  `json:"driver"`
	Distance float64 `json:"distance"` // meter
}

type SearchRequest struct {
	Location Point   `json:"location" validate:"required"`
	Radius   float64 `json:"radius" validate:"required,gt=0"` // radius in meters
	Limit    int     `json:"limit,omitempty" validate:"omitempty,gte=0"`
}

type BatchCreateRequest struct {
	Drivers []CreateDriverRequest `json:"drivers" validate:"required,min=1,dive"`
}

type CreateDriverRequest struct {
	ID       string `json:"id,omitempty"`
	Location Point  `json:"location" validate:"required"`
}

func NewPoint(longitude, latitude float64) Point {
	return Point{
		Type:        "Point",
		Coordinates: []float64{longitude, latitude},
	}
}

func (p Point) Longitude() float64 {
	return p.Coordinates[0]
}

func (p Point) Latitude() float64 {
	return p.Coordinates[1]
}
