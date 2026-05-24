package handlers

import (
	"aviation/models"
	"aviation/repository"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type mockTicketRepo struct {
	findAll  func(repository.TicketFilter) ([]models.Ticket, int64, error)
	findByID func(uint) (models.Ticket, error)
	create   func(*models.Ticket) error
	update   func(*models.Ticket) error
	delete   func(uint) error
}

func (m *mockTicketRepo) FindAll(f repository.TicketFilter) ([]models.Ticket, int64, error) {
	if m.findAll == nil {
		return nil, 0, nil
	}
	return m.findAll(f)
}
func (m *mockTicketRepo) FindByID(id uint) (models.Ticket, error) {
	if m.findByID == nil {
		return models.Ticket{}, nil
	}
	return m.findByID(id)
}
func (m *mockTicketRepo) Create(t *models.Ticket) error {
	if m.create == nil {
		return nil
	}
	return m.create(t)
}
func (m *mockTicketRepo) Update(t *models.Ticket) error {
	if m.update == nil {
		return nil
	}
	return m.update(t)
}
func (m *mockTicketRepo) Delete(id uint) error {
	if m.delete == nil {
		return nil
	}
	return m.delete(id)
}

type mockUserRepo struct {
	findByUsername    func(string) (models.User, error)
	findByID          func(uint) (models.User, error)
	create            func(*models.User) error
	updatePassengerID func(uint, uint) error
	findWithPassenger func(uint) (models.User, error)
}

func (m *mockUserRepo) FindByUsername(u string) (models.User, error) {
	if m.findByUsername == nil {
		return models.User{}, nil
	}
	return m.findByUsername(u)
}
func (m *mockUserRepo) FindByID(id uint) (models.User, error) {
	if m.findByID == nil {
		return models.User{}, nil
	}
	return m.findByID(id)
}
func (m *mockUserRepo) Create(u *models.User) error {
	if m.create == nil {
		return nil
	}
	return m.create(u)
}
func (m *mockUserRepo) UpdatePassengerID(userID, passengerID uint) error {
	if m.updatePassengerID == nil {
		return nil
	}
	return m.updatePassengerID(userID, passengerID)
}
func (m *mockUserRepo) FindWithPassenger(id uint) (models.User, error) {
	if m.findWithPassenger == nil {
		return models.User{}, nil
	}
	return m.findWithPassenger(id)
}

func injectUserID(userID float64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", userID)
		c.Next()
	}
}

func newTicketRouter(
	ticketRepo repository.TicketRepository,
	flightRepo repository.FlightRepository,
	seatRepo repository.SeatRepository,
	userRepo repository.UserRepository,
) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(injectUserID(1))
	h := NewTicketHandler(ticketRepo, flightRepo, seatRepo, userRepo, nil)
	r.GET("/tickets", h.GetAll)
	r.POST("/tickets", h.Create)
	r.DELETE("/tickets/:id", h.Delete)
	return r
}

func TestTicketList_Success(t *testing.T) {
	ticketMock := &mockTicketRepo{
		findAll: func(f repository.TicketFilter) ([]models.Ticket, int64, error) {
			return []models.Ticket{
				{ID: 1, FlightID: 1, PassengerID: 1, SeatID: 5, Status: "reserved"},
				{ID: 2, FlightID: 2, PassengerID: 2, SeatID: 8, Status: "paid"},
			}, 2, nil
		},
	}
	flightMock := &mockFlightRepo{}
	seatMock := &mockSeatRepo{}
	userMock := &mockUserRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tickets?page=1&limit=10", nil)
	newTicketRouter(ticketMock, flightMock, seatMock, userMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)
	assert.Equal(t, float64(2), body["total"])
}

func TestTicketCreate_SeatAlreadyBooked(t *testing.T) {
	passengerID := uint(1)
	ticketMock := &mockTicketRepo{}
	flightMock := &mockFlightRepo{
		findByID: func(id uint) (models.Flight, error) {
			return models.Flight{ID: 1, FlightNumber: "KC-101", Origin: "ALA", Destination: "NQZ"}, nil
		},
	}
	seatMock := &mockSeatRepo{
		findByID: func(id uint) (models.Seat, error) {
			return models.Seat{ID: 6, FlightID: 1, SeatNumber: "6B", Class: "economy", Price: 25000, Status: "booked"}, nil
		},
		bookSeat: func(id uint) error {
			return errors.New("место уже занято или не существует")
		},
	}
	userMock := &mockUserRepo{
		findWithPassenger: func(id uint) (models.User, error) {
			return models.User{ID: 1, Username: "rafi", Role: "user", PassengerID: &passengerID}, nil
		},
	}

	payload := map[string]uint{"flight_id": 1, "seat_id": 6}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newTicketRouter(ticketMock, flightMock, seatMock, userMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "место уже занято или не существует", resp["error"])
}

func TestTicketDelete_NotFound(t *testing.T) {
	ticketMock := &mockTicketRepo{
		findByID: func(id uint) (models.Ticket, error) {
			return models.Ticket{}, gorm.ErrRecordNotFound
		},
	}
	flightMock := &mockFlightRepo{}
	seatMock := &mockSeatRepo{}
	userMock := &mockUserRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/tickets/999", nil)
	newTicketRouter(ticketMock, flightMock, seatMock, userMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
