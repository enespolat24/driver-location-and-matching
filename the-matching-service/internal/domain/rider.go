package domain

// geojson
type Location struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

type Rider struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

// NewRider creates a new Rider with a generated UUID as ID
func NewRider(ID string, name string, location Location) *Rider {
	return &Rider{
		ID:       ID,
		Name:     name,
		Location: location,
	}
}
