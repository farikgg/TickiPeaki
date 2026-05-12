package config

import (
	"aviation/models"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error

	// даём БД время подняться
	for i := range 5 {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("БД не готова, попытка %d/5...", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(
		&models.Flight{},
		&models.Passenger{},
		&models.Ticket{},
	); err != nil {
		return nil, err
	}

	return db, nil
}
