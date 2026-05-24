package postgres

import (
	"aviation/models"
	"aviation/repository"

	"gorm.io/gorm"
)

type FlightRepo struct {
	db *gorm.DB
}

func NewFlightRepo(db *gorm.DB) *FlightRepo {
	return &FlightRepo{db: db}
}

func (r *FlightRepo) FindAll(filter repository.FlightFilter) ([]models.Flight, int64, error) {
	var flights []models.Flight
	var total int64

	q := r.db.Model(&models.Flight{})

	if filter.Origin != "" {
		q = q.Where("origin = ?", filter.Origin)
	}
	if filter.Destination != "" {
		q = q.Where("destination = ?", filter.Destination)
	}
	if filter.Carrier != "" {
		q = q.Where("carrier = ?", filter.Carrier)
	}

	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err = q.Offset(offset).Limit(filter.Limit).Find(&flights).Error
	if err != nil {
		return nil, 0, err
	}

	return flights, total, nil
}

func (r *FlightRepo) FindByID(id uint) (models.Flight, error) {
	var flight models.Flight
	err := r.db.First(&flight, id).Error
	return flight, err
}

func (r *FlightRepo) Create(f *models.Flight) error {
	return r.db.Create(f).Error
}

func (r *FlightRepo) Update(f *models.Flight) error {
	return r.db.Save(f).Error
}

func (r *FlightRepo) Delete(id uint) error {
	return r.db.Delete(&models.Flight{}, id).Error
}
