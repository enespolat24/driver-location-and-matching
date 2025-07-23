package domain

type DriverWithDistance struct {
	Driver   Driver  `json:"driver"`
	Distance float64 `json:"distance"`
}

type Driver struct {
	ID        string   `json:"id"`
	Location  Location `json:"location"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}
