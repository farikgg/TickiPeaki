package models

// Flight — рейс с маршрутом, временами и свободными местами.
type Flight struct {
	ID             int     `json:"id"`
	Origin         string  `json:"origin"`
	Destination    string  `json:"destination"`
	Type           string  `json:"type"`           // "air" или "rail"
	Carrier        string  `json:"carrier"`
	DepartureTime  string  `json:"departure_time"` // RFC3339
	ArrivalTime    string  `json:"arrival_time"`   // RFC3339
	AvailableSeats int     `json:"available_seats"`
	Price          float64 `json:"price"`
}
