package models

import "time"

type Ticket struct {
	ID          uint      `json:"id"           gorm:"primaryKey"`
	FlightID    uint      `json:"flight_id"`
	Flight      Flight    `json:"flight"       gorm:"foreignKey:FlightID"`
	PassengerID uint      `json:"passenger_id"`
	Passenger   Passenger `json:"passenger"    gorm:"foreignKey:PassengerID"`
	SeatID      uint      `json:"seat_id"`
	Seat        Seat      `json:"seat"         gorm:"foreignKey:SeatID"`
	Status      string    `json:"status"`
	BookedAt    time.Time `json:"booked_at"`
	PDFURL      *string   `json:"pdf_url"      gorm:"default:null"`
}
