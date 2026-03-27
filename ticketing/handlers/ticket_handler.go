package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ticketing/models"
	"ticketing/repository"
)

type TicketHandler struct {
	flights repository.FlightRepository
	tickets repository.TicketRepository
}

func NewTicketHandler(flights repository.FlightRepository, tickets repository.TicketRepository) *TicketHandler {
	return &TicketHandler{flights: flights, tickets: tickets}
}

type ticketListResponse struct {
	Data  []models.Ticket `json:"data"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Total int             `json:"total"`
}

type ticketInput struct {
	FlightID       int     `json:"flight_id"`
	PassengerName  string  `json:"passenger_name"`
	PassengerEmail string  `json:"passenger_email"`
	SeatNumber     string  `json:"seat_number"`
	Class          string  `json:"class"`
	Price          float64 `json:"price"`
	Status         string  `json:"status"`
}

var validClasses = map[string]bool{"economy": true, "business": true, "first": true}
var validStatuses = map[string]bool{"reserved": true, "paid": true, "cancelled": true}

func (h *TicketHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	var flightID int
	if s := c.Query("flight_id"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			flightID = n
		}
	}

	filter := repository.TicketFilter{
		FlightID: flightID,
		Status:   strings.TrimSpace(c.Query("status")),
		Class:    strings.TrimSpace(c.Query("class")),
		Page:     page,
		Limit:    limit,
	}

	data, total := h.tickets.FindAll(filter)
	c.JSON(http.StatusOK, ticketListResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *TicketHandler) Create(c *gin.Context) {
	var input ticketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON"})
		return
	}

	errs := validateTicketInput(input)

	if input.FlightID > 0 {
		if _, ok := h.flights.FindByID(input.FlightID); !ok {
			errs["flight_id"] = "рейс не найден"
		}
	}

	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	if !h.flights.DecrementSeat(input.FlightID) {
		c.JSON(http.StatusConflict, gin.H{"error": "мест нет"})
		return
	}

	t := models.Ticket{
		FlightID:       input.FlightID,
		PassengerName:  strings.TrimSpace(input.PassengerName),
		PassengerEmail: strings.TrimSpace(input.PassengerEmail),
		SeatNumber:     strings.TrimSpace(input.SeatNumber),
		Class:          strings.ToLower(strings.TrimSpace(input.Class)),
		Price:          input.Price,
		Status:         "reserved",
	}
	created := h.tickets.Create(t)
	c.JSON(http.StatusCreated, created)
}

func (h *TicketHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID билета"})
		return
	}

	ticket, ok := h.tickets.FindByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "билет не найден"})
		return
	}

	c.JSON(http.StatusOK, ticket)
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID билета"})
		return
	}

	existing, ok := h.tickets.FindByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "билет не найден"})
		return
	}

	var input ticketInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON"})
		return
	}

	input.FlightID = existing.FlightID

	errs := validateTicketInput(input)
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	newStatus := strings.ToLower(strings.TrimSpace(input.Status))

	if newStatus == "cancelled" && existing.Status != "cancelled" {
		h.flights.IncrementSeat(existing.FlightID)
	}

	existing.PassengerName = strings.TrimSpace(input.PassengerName)
	existing.PassengerEmail = strings.TrimSpace(input.PassengerEmail)
	existing.SeatNumber = strings.TrimSpace(input.SeatNumber)
	existing.Class = strings.ToLower(strings.TrimSpace(input.Class))
	existing.Price = input.Price
	existing.Status = newStatus

	updated, _ := h.tickets.Update(existing)
	c.JSON(http.StatusOK, updated)
}

func (h *TicketHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID билета"})
		return
	}

	ticket, ok := h.tickets.FindByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "билет не найден"})
		return
	}

	if ticket.Status != "cancelled" {
		h.flights.IncrementSeat(ticket.FlightID)
	}

	h.tickets.Delete(ticket.ID)
	c.Status(http.StatusNoContent)
}

// ── валидация ─────────────────────────────────────────────────────────────────

func validateTicketInput(t ticketInput) map[string]string {
	errs := map[string]string{}

	if t.FlightID <= 0 {
		errs["flight_id"] = "обязательное поле"
	}
	if strings.TrimSpace(t.PassengerName) == "" {
		errs["passenger_name"] = "обязательное поле"
	}
	email := strings.TrimSpace(t.PassengerEmail)
	if email == "" {
		errs["passenger_email"] = "обязательное поле"
	} else if !strings.Contains(email, "@") {
		errs["passenger_email"] = "должен содержать @"
	}
	if strings.TrimSpace(t.SeatNumber) == "" {
		errs["seat_number"] = "обязательное поле"
	}
	class := strings.ToLower(strings.TrimSpace(t.Class))
	if class == "" {
		errs["class"] = "обязательное поле"
	} else if !validClasses[class] {
		errs["class"] = `должен быть "economy", "business" или "first"`
	}
	if t.Price <= 0 {
		errs["price"] = "должна быть > 0"
	}
	if t.Status != "" {
		if !validStatuses[strings.ToLower(strings.TrimSpace(t.Status))] {
			errs["status"] = `должен быть "reserved", "paid" или "cancelled"`
		}
	}

	return errs
}
