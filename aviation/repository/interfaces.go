package repository

import "aviation/models"

type FlightRepository interface {
	FindAll(filter FlightFilter) ([]models.Flight, int64, error)
	FindByID(id uint) (models.Flight, error)
	Create(f *models.Flight) error
	Update(f *models.Flight) error
	Delete(id uint) error
}

type SeatRepository interface {
	FindByFlight(flightID uint) ([]models.Seat, error)
	FindAvailable(flightID uint) ([]models.Seat, error)
	FindByID(id uint) (models.Seat, error)
	CreateBatch(seats []models.Seat) error
	BookSeat(id uint) error
	ReleaseSeat(id uint) error
}

type PassengerRepository interface {
	FindAll(page, limit int) ([]models.Passenger, int64, error)
	FindByID(id uint) (models.Passenger, error)
	Create(p *models.Passenger) error
	Update(p *models.Passenger) error
	Delete(id uint) error
}

type TicketRepository interface {
	FindAll(filter TicketFilter) ([]models.Ticket, int64, error)
	FindByID(id uint) (models.Ticket, error)
	Create(t *models.Ticket) error
	Update(t *models.Ticket) error
	Delete(id uint) error
}

type UserRepository interface {
	FindByUsername(username string) (models.User, error)
	FindByID(id uint) (models.User, error)
	Create(u *models.User) error
	UpdatePassengerID(userID uint, passengerID uint) error
	FindWithPassenger(userID uint) (models.User, error)
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
	Page        int
	Limit       int
}
