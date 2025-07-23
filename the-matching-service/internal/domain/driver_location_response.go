package domain

type DriverDistancePair struct {
	Driver   Driver  `json:"driver"`
	Distance float64 `json:"distance"`
}

type DriverLocationServiceResponse struct {
	Count   int                  `json:"count"`
	Drivers []DriverDistancePair `json:"drivers"`
}
