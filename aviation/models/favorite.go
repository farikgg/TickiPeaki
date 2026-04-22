package models

import "time"

type Favorite struct {
	UserID    uint      `json:"user_id" gorm:"primaryKey"`
	FlightID  uint      `json:"flight_id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
}
