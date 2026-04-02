package handlers

import (
	"aviation/models"
	"aviation/repository"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PassengerHandler struct {
	repo repository.PassengerRepository
}

func NewPassengerHandler(repo repository.PassengerRepository) *PassengerHandler {
	return &PassengerHandler{repo: repo}
}

func (h *PassengerHandler) GetAll(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}

	passengers, total, err := h.repo.FindAll(page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  passengers,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *PassengerHandler) Create(c *gin.Context) {
	var passenger models.Passenger
	if err := c.ShouldBindJSON(&passenger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validatePassenger(&passenger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(&passenger); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "email или passport_num уже существует"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, passenger)
}

func (h *PassengerHandler) Update(c *gin.Context) {
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

	var passenger models.Passenger
	if err := c.ShouldBindJSON(&passenger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validatePassenger(&passenger); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passenger.ID = existing.ID
	if err := h.repo.Update(&passenger); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			c.JSON(http.StatusConflict, gin.H{"error": "email или passport_num уже существует"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, passenger)
}

func (h *PassengerHandler) Delete(c *gin.Context) {
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
		if err.Error() == "у пассажира есть билеты" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func validatePassenger(p *models.Passenger) error {
	if p.FullName == "" {
		return errors.New("full_name обязателен")
	}
	if p.Email == "" || !strings.Contains(p.Email, "@") {
		return errors.New("email обязателен и должен содержать @")
	}
	if p.Phone == "" {
		return errors.New("phone обязателен")
	}
	if p.PassportNum == "" {
		return errors.New("passport_num обязателен")
	}
	return nil
}
