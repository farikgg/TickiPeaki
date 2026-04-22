package postgres

import (
	"errors"
	"strings"
	"time"

	"aviation/models"

	"gorm.io/gorm"
)

type FavoriteRepo struct {
	db *gorm.DB
}

func NewFavoriteRepo(db *gorm.DB) *FavoriteRepo {
	return &FavoriteRepo{db: db}
}

func (r *FavoriteRepo) FindAllByUser(userID uint, page, limit int) ([]models.Flight, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	var flights []models.Flight
	var total int64

	base := r.db.Model(&models.Flight{}).
		Joins("JOIN favorites ON flights.id = favorites.flight_id").
		Where("favorites.user_id = ?", userID)

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := base.Offset(offset).Limit(limit).Find(&flights).Error; err != nil {
		return nil, 0, err
	}

	return flights, total, nil
}

func (r *FavoriteRepo) Add(userID, flightID uint) error {
	fav := models.Favorite{
		UserID:    userID,
		FlightID:  flightID,
		CreatedAt: time.Now(),
	}
	if err := r.db.Create(&fav).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") ||
			strings.Contains(strings.ToLower(err.Error()), "unique") {
			return errors.New("уже в избранном")
		}
		return err
	}
	return nil
}

func (r *FavoriteRepo) Remove(userID, flightID uint) error {
	res := r.db.Where("user_id = ? AND flight_id = ?", userID, flightID).Delete(&models.Favorite{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
