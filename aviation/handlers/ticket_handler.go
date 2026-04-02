package handlers

import (
	"aviation/models"
	"aviation/repository"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketHandler struct {
	ticketRepo repository.TicketRepository
	flightRepo repository.FlightRepository
}

func NewTicketHandler(tr repository.TicketRepository, fr repository.FlightRepository) *TicketHandler {
	return &TicketHandler{ticketRepo: tr, flightRepo: fr}
}

func (h *TicketHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	flightID, _ := strconv.ParseUint(c.Query("flight_id"), 10, 64)
	passengerID, _ := strconv.ParseUint(c.Query("passenger_id"), 10, 64)

	filter := repository.TicketFilter{
		FlightID:    uint(flightID),
		PassengerID: uint(passengerID),
		Status:      c.Query("status"),
		Class:       c.Query("class"),
		Page:        page,
		Limit:       limit,
	}

	tickets, total, err := h.ticketRepo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  tickets,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *TicketHandler) Create(c *gin.Context) {
	var ticket models.Ticket
	if err := c.ShouldBindJSON(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateTicket(&ticket); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := h.flightRepo.FindByID(ticket.FlightID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "рейс не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.flightRepo.DecrementSeat(ticket.FlightID); err != nil {
		if err.Error() == "мест нет" {
			c.JSON(http.StatusConflict, gin.H{"error": "мест нет"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ticket.Status = "reserved"
	ticket.BookedAt = time.Now()

	if err := h.ticketRepo.Create(&ticket); err != nil {
		// откат места если не удалось создать билет
		_ = h.flightRepo.IncrementSeat(ticket.FlightID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, ticket)
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	existing, err := h.ticketRepo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		SeatNumber string  `json:"seat_number"`
		Class      string  `json:"class"`
		Price      float64 `json:"price"`
		Status     string  `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.SeatNumber != "" {
		existing.SeatNumber = input.SeatNumber
	}
	if input.Class != "" {
		if input.Class != "economy" && input.Class != "business" && input.Class != "first" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "class должен быть economy, business или first"})
			return
		}
		existing.Class = input.Class
	}
	if input.Price > 0 {
		existing.Price = input.Price
	}
	if input.Status != "" {
		if input.Status == "cancelled" && existing.Status != "cancelled" {
			if err := h.flightRepo.IncrementSeat(existing.FlightID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		existing.Status = input.Status
	}

	// сбрасываем связанные объекты чтобы Save не пытался их пересоздать
	existing.Flight = models.Flight{}
	existing.Passenger = models.Passenger{}

	if err := h.ticketRepo.Update(&existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, existing)
}

func (h *TicketHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	existing, err := h.ticketRepo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if existing.Status != "cancelled" {
		if err := h.flightRepo.IncrementSeat(existing.FlightID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	if err := h.ticketRepo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func validateTicket(t *models.Ticket) error {
	if t.FlightID == 0 {
		return errors.New("flight_id обязателен")
	}
	if t.PassengerID == 0 {
		return errors.New("passenger_id обязателен")
	}
	if t.SeatNumber == "" {
		return errors.New("seat_number обязателен")
	}
	if t.Class != "economy" && t.Class != "business" && t.Class != "first" {
		return errors.New("class должен быть economy, business или first")
	}
	if t.Price <= 0 {
		return errors.New("price должен быть > 0")
	}
	return nil
}
