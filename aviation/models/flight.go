package models

import "time"

type Flight struct {
	ID             uint      `json:"id" gorm:"primaryKey"`
	FlightNumber   string    `json:"flight_number" gorm:"uniqueIndex"`
	Origin         string    `json:"origin"`
	Destination    string    `json:"destination"`
	Carrier        string    `json:"carrier"`
	DepartureTime  time.Time `json:"departure_time"`
	ArrivalTime    time.Time `json:"arrival_time"`
	AvailableSeats int       `json:"available_seats"`
	Price          float64   `json:"price"`
}
