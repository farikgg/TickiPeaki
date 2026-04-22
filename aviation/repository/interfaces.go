package repository

import "aviation/models"

type FlightRepository interface {
	FindAll(filter FlightFilter) ([]models.Flight, int64, error)
	FindByID(id uint) (models.Flight, error)
	Create(f *models.Flight) error
	Update(f *models.Flight) error
	Delete(id uint) error
	DecrementSeat(id uint) error
	IncrementSeat(id uint) error
}

type PassengerRepository interface {
	FindAll(page, limit int) ([]models.Passenger, int64, error)
	FindByID(id uint) (models.Passenger, error)
	Create(p *models.Passenger) error
	Update(p *models.Passenger) error
	Delete(id uint) error
}

type UserRepository interface {
	FindByUsername(username string) (models.User, error)
	Create(u *models.User) error
}

type TicketRepository interface {
	FindAll(filter TicketFilter) ([]models.Ticket, int64, error)
	FindByID(id uint) (models.Ticket, error)
	Create(t *models.Ticket) error
	Update(t *models.Ticket) error
	Delete(id uint) error
}

type FavoriteRepository interface {
	FindAllByUser(userID uint, page, limit int) ([]models.Flight, int64, error)
	Add(userID, flightID uint) error
	Remove(userID, flightID uint) error
}

type FlightFilter struct {
	Origin      string
	Destination string
	Carrier     string
	Page        int
	Limit       int
}

type TicketFilter struct {
	FlightID    uint
	PassengerID uint
	Status      string
	Class       string
	Page        int
	Limit       int
}
