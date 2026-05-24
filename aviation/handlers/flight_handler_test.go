package handlers

import (
	"aviation/models"
	"aviation/repository"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockFlightRepo struct {
	findAll  func(repository.FlightFilter) ([]models.Flight, int64, error)
	findByID func(uint) (models.Flight, error)
	create   func(*models.Flight) error
	update   func(*models.Flight) error
	delete   func(uint) error
}

func (m *mockFlightRepo) FindAll(f repository.FlightFilter) ([]models.Flight, int64, error) {
	if m.findAll == nil {
		return nil, 0, nil
	}
	return m.findAll(f)
}
func (m *mockFlightRepo) FindByID(id uint) (models.Flight, error) {
	if m.findByID == nil {
		return models.Flight{}, nil
	}
	return m.findByID(id)
}
func (m *mockFlightRepo) Create(f *models.Flight) error {
	if m.create == nil {
		return nil
	}
	return m.create(f)
}
func (m *mockFlightRepo) Update(f *models.Flight) error {
	if m.update == nil {
		return nil
	}
	return m.update(f)
}
func (m *mockFlightRepo) Delete(id uint) error {
	if m.delete == nil {
		return nil
	}
	return m.delete(id)
}

type mockSeatRepo struct {
	findByFlight  func(uint) ([]models.Seat, error)
	findAvailable func(uint) ([]models.Seat, error)
	findByID      func(uint) (models.Seat, error)
	createBatch   func([]models.Seat) error
	bookSeat      func(uint) error
	releaseSeat   func(uint) error
}

func (m *mockSeatRepo) FindByFlight(id uint) ([]models.Seat, error) {
	if m.findByFlight == nil {
		return nil, nil
	}
	return m.findByFlight(id)
}
func (m *mockSeatRepo) FindAvailable(id uint) ([]models.Seat, error) {
	if m.findAvailable == nil {
		return nil, nil
	}
	return m.findAvailable(id)
}
func (m *mockSeatRepo) FindByID(id uint) (models.Seat, error) {
	if m.findByID == nil {
		return models.Seat{}, nil
	}
	return m.findByID(id)
}
func (m *mockSeatRepo) CreateBatch(seats []models.Seat) error {
	if m.createBatch == nil {
		return nil
	}
	return m.createBatch(seats)
}
func (m *mockSeatRepo) BookSeat(id uint) error {
	if m.bookSeat == nil {
		return nil
	}
	return m.bookSeat(id)
}
func (m *mockSeatRepo) ReleaseSeat(id uint) error {
	if m.releaseSeat == nil {
		return nil
	}
	return m.releaseSeat(id)
}

func newFlightRouter(flightRepo repository.FlightRepository, seatRepo repository.SeatRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewFlightHandler(flightRepo, seatRepo)
	r.GET("/flights", h.GetAll)
	r.GET("/flights/:id", h.GetByID)
	r.POST("/flights", h.Create)
	r.PUT("/flights/:id", h.Update)
	r.DELETE("/flights/:id", h.Delete)
	return r
}

func TestFlightList_Success(t *testing.T) {
	flightMock := &mockFlightRepo{
		findAll: func(f repository.FlightFilter) ([]models.Flight, int64, error) {
			return []models.Flight{
				{ID: 1, FlightNumber: "KC-101", Origin: "ALA", Destination: "NQZ", Carrier: "Air Astana"},
				{ID: 2, FlightNumber: "KC-202", Origin: "NQZ", Destination: "ALA", Carrier: "Air Astana"},
			}, 2, nil
		},
	}
	seatMock := &mockSeatRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flights?page=1&limit=10", nil)
	newFlightRouter(flightMock, seatMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)
	assert.Equal(t, float64(2), body["total"])
}

func TestFlightCreate_Success(t *testing.T) {
	flightMock := &mockFlightRepo{
		create: func(f *models.Flight) error {
			f.ID = 99
			return nil
		},
	}
	seatMock := &mockSeatRepo{}

	payload := map[string]string{
		"flight_number":  "KC-999",
		"origin":         "ALA",
		"destination":    "NQZ",
		"carrier":        "Air Astana",
		"departure_time": "2025-07-01T06:00:00Z",
		"arrival_time":   "2025-07-01T07:30:00Z",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flights", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newFlightRouter(flightMock, seatMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "KC-999", resp["flight_number"])
}

func TestFlightCreate_ValidationError(t *testing.T) {
	flightMock := &mockFlightRepo{}
	seatMock := &mockSeatRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/flights", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	newFlightRouter(flightMock, seatMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}

func TestFlightGetByID_NotFound(t *testing.T) {
	flightMock := &mockFlightRepo{
		findByID: func(id uint) (models.Flight, error) {
			return models.Flight{}, gorm.ErrRecordNotFound
		},
	}
	seatMock := &mockSeatRepo{
		findByFlight: func(id uint) ([]models.Seat, error) {
			return nil, nil
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/flights/999", nil)
	newFlightRouter(flightMock, seatMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
