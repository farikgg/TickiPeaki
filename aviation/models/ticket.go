package models

import "time"

type Ticket struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	FlightID    uint      `json:"flight_id"`
	Flight      Flight    `json:"flight" gorm:"foreignKey:FlightID"`
	PassengerID uint      `json:"passenger_id"`
	Passenger   Passenger `json:"passenger" gorm:"foreignKey:PassengerID"`
	SeatNumber  string    `json:"seat_number"`
	Class       string    `json:"class"`
	Price       float64   `json:"price"`
	Status      string    `json:"status"`
	BookedAt    time.Time `json:"booked_at"`
}
