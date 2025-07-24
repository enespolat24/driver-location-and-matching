package domain

// Location represents a GeoJSON Point location
// @Description GeoJSON Point location with longitude and latitude coordinates
type Location struct {
	Type        string     `json:"type" validate:"required,eq=Point" example:"Point" description:"GeoJSON type, must be 'Point'"`
	Coordinates [2]float64 `json:"coordinates" validate:"required,len=2,coordinates" example:"28.9784,41.0082" description:"Array of [longitude, latitude] coordinates"`
}

type Rider struct {
	ID       string   `json:"id"`
	Location Location `json:"location" validate:"required"`
}

// NewRider creates a new Rider with the given ID, name and location
// the important thing is we exctract user_id from jwt token
// in the case it is clear that we are using an existing token so
// we do not need to create an auth mechanism
// since we are in microservices architecture we have our seperate
// auth service and i presume this token is created by that service
// also we can name id as user_id alternatively but i assume rider is the user
func NewRider(ID string, location Location) *Rider {
	return &Rider{
		ID:       ID,
		Location: location,
	}
}
