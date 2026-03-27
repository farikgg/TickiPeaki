package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"ticketing/models"
	"ticketing/repository"
)

type FlightHandler struct {
	repo repository.FlightRepository
}

func NewFlightHandler(repo repository.FlightRepository) *FlightHandler {
	return &FlightHandler{repo: repo}
}

type flightListResponse struct {
	Data  []models.Flight `json:"data"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Total int             `json:"total"`
}

func (h *FlightHandler) List(c *gin.Context) {
	page, limit := parsePagination(c)

	filter := repository.FlightFilter{
		Type:        strings.TrimSpace(c.Query("type")),
		Origin:      strings.TrimSpace(c.Query("origin")),
		Destination: strings.TrimSpace(c.Query("destination")),
		Page:        page,
		Limit:       limit,
	}

	data, total := h.repo.FindAll(filter)
	c.JSON(http.StatusOK, flightListResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

func (h *FlightHandler) Create(c *gin.Context) {
	var f models.Flight
	if err := c.ShouldBindJSON(&f); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON"})
		return
	}

	errs := validateFlight(f)
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	created := h.repo.Create(f)
	c.JSON(http.StatusCreated, created)
}

func (h *FlightHandler) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID рейса"})
		return
	}

	f, ok := h.repo.FindByID(id)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "не нашли рейс"})
		return
	}

	c.JSON(http.StatusOK, f)
}

func (h *FlightHandler) Update(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID рейса"})
		return
	}

	if _, ok := h.repo.FindByID(id); !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "не нашли рейс"})
		return
	}

	var input models.Flight
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный JSON"})
		return
	}

	errs := validateFlight(input)
	if len(errs) > 0 {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
		return
	}

	input.ID = id
	updated, _ := h.repo.Update(input)
	c.JSON(http.StatusOK, updated)
}

func (h *FlightHandler) Delete(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный ID рейса"})
		return
	}

	if !h.repo.Delete(id) {
		c.JSON(http.StatusNotFound, gin.H{"error": "не нашли рейс"})
		return
	}

	c.Status(http.StatusNoContent)
}

// валидация

func validateFlight(f models.Flight) map[string]string {
	errs := map[string]string{}

	if strings.TrimSpace(f.Origin) == "" {
		errs["origin"] = "обязательное поле"
	}
	if strings.TrimSpace(f.Destination) == "" {
		errs["destination"] = "обязательное поле"
	}
	if strings.TrimSpace(f.Carrier) == "" {
		errs["carrier"] = "обязательное поле"
	}

	t := strings.ToLower(strings.TrimSpace(f.Type))
	if t != "air" && t != "rail" {
		errs["type"] = `должен быть "air" или "rail"`
	}

	if f.AvailableSeats < 1 {
		errs["available_seats"] = "должно быть >= 1"
	}
	if f.Price <= 0 {
		errs["price"] = "должна быть > 0"
	}

	validateTimes(f.DepartureTime, f.ArrivalTime, errs)
	return errs
}

func validateTimes(departure, arrival string, errs map[string]string) {
	var dep, arr time.Time

	if strings.TrimSpace(departure) == "" {
		errs["departure_time"] = "обязательное поле"
	} else {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(departure))
		if err != nil {
			errs["departure_time"] = "нужен формат RFC3339"
		} else {
			dep = t
		}
	}

	if strings.TrimSpace(arrival) == "" {
		errs["arrival_time"] = "обязательное поле"
	} else {
		t, err := time.Parse(time.RFC3339, strings.TrimSpace(arrival))
		if err != nil {
			errs["arrival_time"] = "нужен формат RFC3339"
		} else {
			arr = t
		}
	}

	if _, ok := errs["departure_time"]; !ok {
		if _, ok := errs["arrival_time"]; !ok {
			if !dep.Before(arr) {
				errs["departure_time"] = "должен быть раньше arrival_time"
			}
		}
	}
}

// parsePagination читает page и limit из query-параметров.
func parsePagination(c *gin.Context) (page, limit int) {
	page = 1
	limit = 10
	if s := c.Query("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 1 {
			page = n
		}
	}
	if s := c.Query("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 1 {
			limit = n
		}
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
