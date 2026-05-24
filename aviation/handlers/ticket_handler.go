package handlers

import (
	"aviation/clients"
	"aviation/models"
	"aviation/repository"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TicketHandler struct {
	ticketRepo repository.TicketRepository
	flightRepo repository.FlightRepository
	seatRepo   repository.SeatRepository
	userRepo   repository.UserRepository
	pdfClient  *clients.PDFClient
}

func NewTicketHandler(
	ticketRepo repository.TicketRepository,
	flightRepo repository.FlightRepository,
	seatRepo repository.SeatRepository,
	userRepo repository.UserRepository,
	pdfClient *clients.PDFClient,
) *TicketHandler {
	return &TicketHandler{
		ticketRepo: ticketRepo,
		flightRepo: flightRepo,
		seatRepo:   seatRepo,
		userRepo:   userRepo,
		pdfClient:  pdfClient,
	}
}

type createTicketRequest struct {
	FlightID uint `json:"flight_id" binding:"required"`
	SeatID   uint `json:"seat_id"   binding:"required"`
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
	userID, ok := userIDFromContext(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "невалидный токен"})
		return
	}

	user, err := h.userRepo.FindWithPassenger(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "пользователь не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if user.PassengerID == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "заполните профиль пассажира перед покупкой билета"})
		return
	}

	var req createTicketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	seat, err := h.seatRepo.FindByID(req.SeatID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "место не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if seat.FlightID != req.FlightID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "место не принадлежит указанному рейсу"})
		return
	}

	if err := h.seatRepo.BookSeat(req.SeatID); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ticket := models.Ticket{
		FlightID:    req.FlightID,
		PassengerID: *user.PassengerID,
		SeatID:      req.SeatID,
		Status:      "reserved",
		BookedAt:    time.Now(),
	}

	if err := h.ticketRepo.Create(&ticket); err != nil {
		// откат — освобождаем место
		_ = h.seatRepo.ReleaseSeat(req.SeatID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := h.ticketRepo.FindByID(ticket.ID)
	if err != nil {
		c.JSON(http.StatusCreated, ticket)
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *TicketHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	oldTicket, err := h.ticketRepo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var input struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	existing := oldTicket
	if input.Status != "" {
		if input.Status == "cancelled" && existing.Status != "cancelled" {
			if err := h.seatRepo.ReleaseSeat(existing.SeatID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		existing.Status = input.Status
	}

	// сбрасываем связанные объекты чтобы Save не пытался их пересоздать
	existing.Flight = models.Flight{}
	existing.Passenger = models.Passenger{}
	existing.Seat = models.Seat{}

	if err := h.ticketRepo.Update(&existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedTicket, err := h.ticketRepo.FindByID(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if input.Status == "paid" && oldTicket.Status != "paid" {
		clients.StartPDFWorker(
			h.pdfClient,
			updatedTicket,
			func(pdfURL string) {
				updatedTicket.PDFURL = &pdfURL
				if err := h.ticketRepo.Update(&updatedTicket); err != nil {
					log.Printf("не смогли сохранить pdf_url для билета %d: %v", updatedTicket.ID, err)
					return
				}
				log.Printf("PDF для билета %d сохранён: %s", updatedTicket.ID, pdfURL)
			},
			func(err error) {
				log.Printf("ошибка генерации PDF для билета %d: %v", updatedTicket.ID, err)
			},
		)
	}

	c.JSON(http.StatusOK, updatedTicket)
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
		if err := h.seatRepo.ReleaseSeat(existing.SeatID); err != nil {
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
