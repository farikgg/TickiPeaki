package postgres

import (
	"aviation/models"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) FindByUsername(username string) (models.User, error) {
	var user models.User
	err := r.db.Where("username = ?", username).First(&user).Error
	return user, err
}

func (r *UserRepo) FindByID(id uint) (models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	return user, err
}

func (r *UserRepo) Create(u *models.User) error {
	return r.db.Create(u).Error
}

func (r *UserRepo) UpdatePassengerID(userID uint, passengerID uint) error {
	return r.db.Model(&models.User{}).
		Where("id = ?", userID).
		Update("passenger_id", passengerID).Error
}

func (r *UserRepo) FindWithPassenger(userID uint) (models.User, error) {
	var user models.User
	err := r.db.Preload("Passenger").First(&user, userID).Error
	return user, err
}
