package models

type Seat struct {
	ID         uint    `json:"id"          gorm:"primaryKey"`
	FlightID   uint    `json:"flight_id"`
	SeatNumber string  `json:"seat_number"`
	Class      string  `json:"class"`
	Price      float64 `json:"price"`
	Status     string  `json:"status"`
}
