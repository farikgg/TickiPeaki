package postgres

import (
	"aviation/models"
	"errors"

	"gorm.io/gorm"
)

type PassengerRepo struct {
	db *gorm.DB
}

func NewPassengerRepo(db *gorm.DB) *PassengerRepo {
	return &PassengerRepo{db: db}
}

func (r *PassengerRepo) FindAll(page, limit int) ([]models.Passenger, int64, error) {
	var passengers []models.Passenger
	var total int64

	err := r.db.Model(&models.Passenger{}).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = r.db.Offset(offset).Limit(limit).Find(&passengers).Error
	if err != nil {
		return nil, 0, err
	}

	return passengers, total, nil
}

func (r *PassengerRepo) FindByID(id uint) (models.Passenger, error) {
	var passenger models.Passenger
	err := r.db.First(&passenger, id).Error
	return passenger, err
}

func (r *PassengerRepo) Create(p *models.Passenger) error {
	return r.db.Create(p).Error
}

func (r *PassengerRepo) Update(p *models.Passenger) error {
	return r.db.Save(p).Error
}

func (r *PassengerRepo) Delete(id uint) error {
	var count int64
	err := r.db.Model(&models.Ticket{}).Where("passenger_id = ?", id).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("у пассажира есть билеты")
	}
	return r.db.Delete(&models.Passenger{}, id).Error
}
