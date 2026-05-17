package handlers

import (
	"aviation/models"
	"aviation/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockFlightRepo struct {
	findAll       func(filter repository.FlightFilter) ([]models.Flight, int64, error)
	findByID      func(id uint) (models.Flight, error)
	create        func(f *models.Flight) error
	update        func(f *models.Flight) error
	delete        func(id uint) error
	decrementSeat func(id uint) error
	incrementSeat func(id uint) error
}

func (m *mockFlightRepo) FindAll(f repository.FlightFilter) ([]models.Flight, int64, error) {
	return m.findAll(f)
}
func (m *mockFlightRepo) FindByID(id uint) (models.Flight, error) { return m.findByID(id) }
func (m *mockFlightRepo) Create(f *models.Flight) error           { return m.create(f) }
func (m *mockFlightRepo) Update(f *models.Flight) error           { return m.update(f) }
func (m *mockFlightRepo) Delete(id uint) error                    { return m.delete(id) }
func (m *mockFlightRepo) DecrementSeat(id uint) error             { return m.decrementSeat(id) }
func (m *mockFlightRepo) IncrementSeat(id uint) error             { return m.incrementSeat(id) }

func newFlightRouter(repo repository.FlightRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewFlightHandler(repo)
	r.GET("/flights", h.GetAll)
	r.POST("/flights", h.Create)
	r.PUT("/flights/:id", h.Update)
	r.DELETE("/flights/:id", h.Delete)
	return r
}

func TestFlightList_Success(t *testing.T) {
	mock := &mockFlightRepo{
		findAll: func(f repository.FlightFilter) ([]models.Flight, int64, error) {
			return []models.Flight{
				{ID: 1, FlightNumber: "KC101", Origin: "ALA", Destination: "NQZ", Carrier: "Air Astana", AvailableSeats: 50, Price: 25000},
				{ID: 2, FlightNumber: "KC102", Origin: "NQZ", Destination: "ALA", Carrier: "Air Astana", AvailableSeats: 30, Price: 26000},
			}, 2, nil
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flights", nil)
	newFlightRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)
	assert.Equal(t, float64(2), body["total"])
}

func TestFlightCreate_Success(t *testing.T) {
	mock := &mockFlightRepo{
		create: func(f *models.Flight) error {
			f.ID = 1
			return nil
		},
	}

	flight := models.Flight{
		FlightNumber:   "KC101",
		Origin:         "ALA",
		Destination:    "NQZ",
		Carrier:        "Air Astana",
		DepartureTime:  time.Now().Add(time.Hour),
		ArrivalTime:    time.Now().Add(3 * time.Hour),
		AvailableSeats: 100,
		Price:          25000,
	}
	body, _ := json.Marshal(flight)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newFlightRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestFlightCreate_ValidationError(t *testing.T) {
	mock := &mockFlightRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flights", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	newFlightRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

func TestFlightGetByID_NotFound(t *testing.T) {
	mock := &mockFlightRepo{
		findByID: func(id uint) (models.Flight, error) {
			return models.Flight{}, gorm.ErrRecordNotFound
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/flights/999", nil)
	newFlightRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
