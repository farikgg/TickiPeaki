package postgres

import (
	"aviation/models"
	"errors"

	"gorm.io/gorm"
)

type SeatRepo struct {
	db *gorm.DB
}

func NewSeatRepo(db *gorm.DB) *SeatRepo {
	return &SeatRepo{db: db}
}

func (r *SeatRepo) FindByFlight(flightID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.Where("flight_id = ?", flightID).
		Order("seat_number").
		Find(&seats).Error
	return seats, err
}

func (r *SeatRepo) FindAvailable(flightID uint) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.Where("flight_id = ? AND status = ?", flightID, "available").
		Order("seat_number").
		Find(&seats).Error
	return seats, err
}

func (r *SeatRepo) FindByID(id uint) (models.Seat, error) {
	var seat models.Seat
	err := r.db.First(&seat, id).Error
	return seat, err
}

func (r *SeatRepo) CreateBatch(seats []models.Seat) error {
	if len(seats) == 0 {
		return nil
	}
	return r.db.Create(&seats).Error
}

// атомарный апдейт — иначе будет гонка
func (r *SeatRepo) BookSeat(id uint) error {
	result := r.db.Model(&models.Seat{}).
		Where("id = ? AND status = ?", id, "available").
		Update("status", "booked")
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("место уже занято или не существует")
	}
	return nil
}

func (r *SeatRepo) ReleaseSeat(id uint) error {
	return r.db.Model(&models.Seat{}).
		Where("id = ?", id).
		Update("status", "available").Error
}
