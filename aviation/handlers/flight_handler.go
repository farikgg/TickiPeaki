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

type FlightHandler struct {
	repo repository.FlightRepository
}

func NewFlightHandler(repo repository.FlightRepository) *FlightHandler {
	return &FlightHandler{repo: repo}
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
	var flight models.Flight
	if err := c.ShouldBindJSON(&flight); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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
	if f.AvailableSeats < 1 {
		return errors.New("available_seats должен быть >= 1")
	}
	if f.Price <= 0 {
		return errors.New("price должен быть > 0")
	}
	return nil
}
