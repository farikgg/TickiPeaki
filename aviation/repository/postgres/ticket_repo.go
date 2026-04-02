package postgres

import (
	"aviation/models"
	"aviation/repository"

	"gorm.io/gorm"
)

type TicketRepo struct {
	db *gorm.DB
}

func NewTicketRepo(db *gorm.DB) *TicketRepo {
	return &TicketRepo{db: db}
}

// preload чтобы не словить N+1
func (r *TicketRepo) FindAll(filter repository.TicketFilter) ([]models.Ticket, int64, error) {
	var tickets []models.Ticket
	var total int64

	q := r.db.Model(&models.Ticket{})

	if filter.FlightID != 0 {
		q = q.Where("flight_id = ?", filter.FlightID)
	}
	if filter.PassengerID != 0 {
		q = q.Where("passenger_id = ?", filter.PassengerID)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", filter.Status)
	}
	if filter.Class != "" {
		q = q.Where("class = ?", filter.Class)
	}

	err := q.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (filter.Page - 1) * filter.Limit
	err = q.Preload("Flight").Preload("Passenger").
		Offset(offset).Limit(filter.Limit).Find(&tickets).Error
	if err != nil {
		return nil, 0, err
	}

	return tickets, total, nil
}

func (r *TicketRepo) FindByID(id uint) (models.Ticket, error) {
	var ticket models.Ticket
	err := r.db.Preload("Flight").Preload("Passenger").First(&ticket, id).Error
	return ticket, err
}

func (r *TicketRepo) Create(t *models.Ticket) error {
	return r.db.Create(t).Error
}

func (r *TicketRepo) Update(t *models.Ticket) error {
	return r.db.Save(t).Error
}

func (r *TicketRepo) Delete(id uint) error {
	return r.db.Delete(&models.Ticket{}, id).Error
}
