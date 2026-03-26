package repository

import "ticketing/models"

// FlightFilter — параметры фильтрации и пагинации рейсов.
type FlightFilter struct {
	Type        string
	Origin      string
	Destination string
	Page        int
	Limit       int
}

// TicketFilter — параметры фильтрации и пагинации билетов.
type TicketFilter struct {
	FlightID int
	Status   string
	Class    string
	Page     int
	Limit    int
}

// FlightRepository — операции с рейсами и управление местами.
type FlightRepository interface {
	FindAll(filter FlightFilter) ([]models.Flight, int)
	FindByID(id int) (models.Flight, bool)
	Create(f models.Flight) models.Flight
	Update(f models.Flight) (models.Flight, bool)
	Delete(id int) bool
	DecrementSeat(id int) bool
	IncrementSeat(id int) bool
}

// TicketRepository — CRUD для билетов.
type TicketRepository interface {
	FindAll(filter TicketFilter) ([]models.Ticket, int)
	FindByID(id int) (models.Ticket, bool)
	Create(t models.Ticket) models.Ticket
	Update(t models.Ticket) (models.Ticket, bool)
	Delete(id int) bool
}
