package domain

type MatchResult struct {
	RiderID  string  `json:"rider_id"`
	DriverID string  `json:"driver_id"`
	Distance float64 `json:"distance"` //meters
}
