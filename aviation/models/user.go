package models

type User struct {
	ID          uint       `json:"id"           gorm:"primaryKey"`
	Username    string     `json:"username"     gorm:"uniqueIndex;not null"`
	Password    string     `json:"-"            gorm:"not null"`
	Role        string     `json:"role"         gorm:"default:'user'"`
	PassengerID *uint      `json:"passenger_id" gorm:"default:null"`
	Passenger   *Passenger `json:"passenger"    gorm:"foreignKey:PassengerID"`
}
