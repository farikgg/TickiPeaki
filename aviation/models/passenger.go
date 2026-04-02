package models

type Passenger struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	FullName    string `json:"full_name"`
	Email       string `json:"email" gorm:"uniqueIndex"`
	Phone       string `json:"phone"`
	PassportNum string `json:"passport_num" gorm:"uniqueIndex"`
}
