package domain

type MatchRequest struct {
	Name     string   `json:"name" validate:"required,min=2,max=50"`
	Surname  string   `json:"surname" validate:"required,min=2,max=50"`
	Location Location `json:"location" validate:"required"`
	Radius   float64  `json:"radius" validate:"required,radius"`
}

func (r *MatchRequest) CreateRider(userID string) *Rider {
	return NewRider(userID, r.Name, r.Surname, r.Location)
}

type MatchResponse struct {
	Driver   Driver  `json:"driver"`
	Distance float64 `json:"distance"`
}

func NewMatchResponse(result *MatchResult) *MatchResponse {
	return &MatchResponse{
		Driver:   Driver{ID: result.DriverID},
		Distance: result.Distance,
	}
}

type DriverDistancePair struct {
	Driver   Driver  `json:"driver"`
	Distance float64 `json:"distance"`
}

type DriverLocationServiceResponse struct {
	Count   int                  `json:"count"`
	Drivers []DriverDistancePair `json:"drivers"`
}
