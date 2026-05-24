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

type FlightHandler struct {
	repo     repository.FlightRepository
	seatRepo repository.SeatRepository
}

func NewFlightHandler(repo repository.FlightRepository, seatRepo repository.SeatRepository) *FlightHandler {
	return &FlightHandler{repo: repo, seatRepo: seatRepo}
}

type createFlightRequest struct {
	FlightNumber  string `json:"flight_number"  binding:"required"`
	Origin        string `json:"origin"         binding:"required"`
	Destination   string `json:"destination"    binding:"required"`
	Carrier       string `json:"carrier"        binding:"required"`
	DepartureTime string `json:"departure_time" binding:"required"`
	ArrivalTime   string `json:"arrival_time"   binding:"required"`
}

type flightDetailResponse struct {
	Flight     models.Flight `json:"flight"`
	Seats      []models.Seat `json:"seats"`
	Available  int           `json:"available_count"`
	TakenSeats []string      `json:"taken_seats"`
}

func (h *FlightHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный id"})
		return
	}

	flight, err := h.repo.FindByID(uint(id))
	if errors.Is(err, gorm.ErrRecordNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "рейс не найден"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	seats, err := h.seatRepo.FindByFlight(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	takenSeats := []string{}
	available := 0
	for _, s := range seats {
		if s.Status == "booked" {
			takenSeats = append(takenSeats, s.SeatNumber)
		}
		if s.Status == "available" {
			available++
		}
	}

	c.JSON(http.StatusOK, flightDetailResponse{
		Flight:     flight,
		Seats:      seats,
		Available:  available,
		TakenSeats: takenSeats,
	})
}

func (h *FlightHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	filter := repository.FlightFilter{
		Origin:      c.Query("origin"),
		Destination: c.Query("destination"),
		Carrier:     c.Query("carrier"),
		Page:        page,
		Limit:       limit,
	}

	flights, total, err := h.repo.FindAll(filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  flights,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *FlightHandler) Create(c *gin.Context) {
	var req createFlightRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	dep, err := time.Parse(time.RFC3339, req.DepartureTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "departure_time должен быть в формате RFC3339"})
		return
	}
	arr, err := time.Parse(time.RFC3339, req.ArrivalTime)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arrival_time должен быть в формате RFC3339"})
		return
	}

	flight := models.Flight{
		FlightNumber:  req.FlightNumber,
		Origin:        req.Origin,
		Destination:   req.Destination,
		Carrier:       req.Carrier,
		DepartureTime: dep,
		ArrivalTime:   arr,
	}

	if err := validateFlight(&flight); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(&flight); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, flight)
}

func (h *FlightHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	existing, err := h.repo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var flight models.Flight
	if err := c.ShouldBindJSON(&flight); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateFlight(&flight); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	flight.ID = existing.ID
	if err := h.repo.Update(&flight); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, flight)
}

func (h *FlightHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "некорректный id"})
		return
	}

	_, err = h.repo.FindByID(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "не найдено"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Delete(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func validateFlight(f *models.Flight) error {
	if f.FlightNumber == "" {
		return errors.New("flight_number обязателен")
	}
	if len(f.Origin) != 3 {
		return errors.New("origin должен быть IATA-кодом (3 символа)")
	}
	if len(f.Destination) != 3 {
		return errors.New("destination должен быть IATA-кодом (3 символа)")
	}
	if f.Origin == f.Destination {
		return errors.New("origin и destination не могут совпадать")
	}
	if f.Carrier == "" {
		return errors.New("carrier обязателен")
	}
	if f.DepartureTime.IsZero() {
		return errors.New("departure_time обязателен")
	}
	if f.ArrivalTime.IsZero() {
		return errors.New("arrival_time обязателен")
	}
	if !f.DepartureTime.Before(f.ArrivalTime) {
		return errors.New("departure_time должен быть раньше arrival_time")
	}
	return nil
}
