package config

import (
	"aviation/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	err = db.AutoMigrate(&models.Flight{}, &models.Passenger{}, &models.Ticket{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
