package models

// Ticket — билет на рейс, данные пассажира хранятся прямо тут.
type Ticket struct {
	ID             int     `json:"id"`
	FlightID       int     `json:"flight_id"`
	PassengerName  string  `json:"passenger_name"`
	PassengerEmail string  `json:"passenger_email"`
	SeatNumber     string  `json:"seat_number"`
	Class          string  `json:"class"`  // "economy", "business", "first"
	Price          float64 `json:"price"`
	Status         string  `json:"status"` // "reserved", "paid", "cancelled"
}
