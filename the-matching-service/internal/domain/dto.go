package domain

// MatchRequest represents a request to match a rider with a nearby driver
// @Description Request to find a nearby driver for a rider
type MatchRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=50" example:"John" description:"Rider's first name"`
	Surname  string   `json:"surname" validate:"required,min=2,max=50" example:"Doe" description:"Rider's last name"`
	Location Location `json:"location" validate:"required" description:"Rider's current location in GeoJSON format"`
	Radius   float64  `json:"radius" validate:"required,radius" example:"500" description:"Search radius in meters"`
}

func (r *MatchRequest) CreateRider(userID string) *Rider {
	return NewRider(userID, r.Name, r.Surname, r.Location)
}

// MatchResponse represents the response when a driver is successfully matched
// @Description Response containing matched driver information
type MatchResponse struct {
	Driver   string  `json:"driver" example:"driver-123" description:"Matched driver ID"`
	Rider    string  `json:"rider" example:"rider-456" description:"Rider ID"`
	Distance float64 `json:"distance" example:"250.5" description:"Distance between rider and driver in meters"`
}

func NewMatchResponse(result *MatchResult) *MatchResponse {
	return &MatchResponse{
		Driver:   result.DriverID,
		Rider:    result.RiderID,
		Distance: result.Distance,
	}
}

type DriverDistancePair struct {
	Driver   Driver  `json:"driver"`
	Distance float64 `json:"distance"`
}

type DriverLocationServiceResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
	Message string      `json:"message,omitempty"`
}

type DriverSearchData struct {
	Count   int                  `json:"count"`
	Drivers []DriverDistancePair `json:"drivers"`
}
