package domain

// geojson
type Location struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

type Rider struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Surname  string   `json:"surname"`
	Location Location `json:"location"`
}

// NewRider creates a new Rider with the given ID, name and location
// the important thing is we exctract user_id from jwt token
// in the case it is clear that we are using an existing token so
// we do not need to create an auth mechanism
// since we are in microservices architecture we have our seperate
// auth service and i presume this token is created by that service
// also we can name id as user_id alternatively but i assume rider is the user
func NewRider(ID, name, surname string, location Location) *Rider {
	return &Rider{
		ID:       ID,
		Name:     name,
		Surname:  surname,
		Location: location,
	}
}
