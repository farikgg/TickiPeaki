package handlers

import (
	"aviation/clients"
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
	findAll  func(filter repository.TicketFilter) ([]models.Ticket, int64, error)
	findByID func(id uint) (models.Ticket, error)
	create   func(t *models.Ticket) error
	update   func(t *models.Ticket) error
	delete   func(id uint) error
}

func (m *mockTicketRepo) FindAll(f repository.TicketFilter) ([]models.Ticket, int64, error) {
	return m.findAll(f)
}
func (m *mockTicketRepo) FindByID(id uint) (models.Ticket, error) { return m.findByID(id) }
func (m *mockTicketRepo) Create(t *models.Ticket) error           { return m.create(t) }
func (m *mockTicketRepo) Update(t *models.Ticket) error           { return m.update(t) }
func (m *mockTicketRepo) Delete(id uint) error                    { return m.delete(id) }

func newTicketRouter(ticketRepo repository.TicketRepository, flightRepo repository.FlightRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	pdfClient := clients.NewPDFClient("")
	h := NewTicketHandler(ticketRepo, flightRepo, pdfClient)
	r.GET("/tickets", h.GetAll)
	r.POST("/tickets", h.Create)
	r.PUT("/tickets/:id", h.Update)
	r.DELETE("/tickets/:id", h.Delete)
	return r
}

func TestTicketList_Success(t *testing.T) {
	ticketMock := &mockTicketRepo{
		findAll: func(f repository.TicketFilter) ([]models.Ticket, int64, error) {
			return []models.Ticket{
				{ID: 1, FlightID: 1, PassengerID: 1, SeatNumber: "12A", Class: "economy", Price: 25000, Status: "reserved"},
				{ID: 2, FlightID: 2, PassengerID: 2, SeatNumber: "5B", Class: "business", Price: 45000, Status: "paid"},
			}, 2, nil
		},
	}
	flightMock := &mockFlightRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/tickets?page=1&limit=10", nil)
	newTicketRouter(ticketMock, flightMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 2)
	assert.Equal(t, float64(2), body["total"])
}

func TestTicketCreate_NoSeats(t *testing.T) {
	ticketMock := &mockTicketRepo{}
	flightMock := &mockFlightRepo{
		findByID: func(id uint) (models.Flight, error) {
			return models.Flight{ID: 1, FlightNumber: "KC101", Origin: "ALA", Destination: "NQZ", AvailableSeats: 1, Price: 25000}, nil
		},
		decrementSeat: func(id uint) error {
			return errors.New("мест нет")
		},
	}

	ticket := models.Ticket{
		FlightID:    1,
		PassengerID: 1,
		SeatNumber:  "12A",
		Class:       "economy",
		Price:       25000,
	}
	body, _ := json.Marshal(ticket)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newTicketRouter(ticketMock, flightMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "мест нет", resp["error"])
}

func TestTicketDelete_NotFound(t *testing.T) {
	ticketMock := &mockTicketRepo{
		findByID: func(id uint) (models.Ticket, error) {
			return models.Ticket{}, gorm.ErrRecordNotFound
		},
	}
	flightMock := &mockFlightRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/tickets/999", nil)
	newTicketRouter(ticketMock, flightMock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
