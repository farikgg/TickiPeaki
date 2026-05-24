package handlers

import (
	"aviation/models"
	"aviation/repository"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SeatHandler struct {
	repo       repository.SeatRepository
	flightRepo repository.FlightRepository
}

func NewSeatHandler(repo repository.SeatRepository, flightRepo repository.FlightRepository) *SeatHandler {
	return &SeatHandler{repo: repo, flightRepo: flightRepo}
}

type createSeatRequest struct {
	FlightID   uint    `json:"flight_id"   binding:"required"`
	SeatNumber string  `json:"seat_number" binding:"required"`
	Class      string  `json:"class"       binding:"required"`
	Price      float64 `json:"price"       binding:"required"`
}

type updateSeatRequest struct {
	Price  float64 `json:"price"`
	Status string  `json:"status"`
}

func (h *SeatHandler) ListByFlight(c *gin.Context) {
	flightID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный id"})
		return
	}

	if _, err := h.flightRepo.FindByID(uint(flightID)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "рейс не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	seats, err := h.repo.FindByFlight(uint(flightID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  seats,
		"total": len(seats),
	})
}

func (h *SeatHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный id"})
		return
	}

	seat, err := h.repo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "место не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, seat)
}

func (h *SeatHandler) Create(c *gin.Context) {
	var req createSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if req.Class != "economy" && req.Class != "business" && req.Class != "first" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "class должен быть economy, business или first"})
		return
	}
	if req.Price <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "price должна быть больше 0"})
		return
	}

	if _, err := h.flightRepo.FindByID(req.FlightID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "рейс не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	existing, err := h.repo.FindByFlight(req.FlightID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	for _, s := range existing {
		if s.SeatNumber == req.SeatNumber {
			c.JSON(http.StatusConflict, gin.H{"error": "место уже существует"})
			return
		}
	}

	seat := models.Seat{
		FlightID:   req.FlightID,
		SeatNumber: req.SeatNumber,
		Class:      req.Class,
		Price:      req.Price,
		Status:     "available",
	}

	if err := h.repo.CreateBatch([]models.Seat{seat}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created, err := h.repo.FindByFlight(req.FlightID)
	if err != nil {
		c.JSON(http.StatusCreated, seat)
		return
	}
	for _, s := range created {
		if s.SeatNumber == req.SeatNumber {
			c.JSON(http.StatusCreated, s)
			return
		}
	}

	c.JSON(http.StatusCreated, seat)
}

func (h *SeatHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный id"})
		return
	}

	seat, err := h.repo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "место не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var req updateSeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
		return
	}

	if req.Price != 0 && req.Price <= 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "price должна быть больше 0"})
		return
	}
	if req.Status != "" && req.Status != "available" && req.Status != "booked" {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"error": "status должен быть available или booked"})
		return
	}

	if req.Price > 0 {
		seat.Price = req.Price
	}
	if req.Status != "" {
		seat.Status = req.Status
	}

	if err := h.repo.Update(&seat); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, seat)
}

func (h *SeatHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный id"})
		return
	}

	seat, err := h.repo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "место не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if seat.Status == "booked" {
		c.JSON(http.StatusConflict, gin.H{"error": "нельзя удалить забронированное место"})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}
