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
)

type mockPassengerRepo struct {
	findAll  func(page, limit int) ([]models.Passenger, int64, error)
	findByID func(id uint) (models.Passenger, error)
	create   func(p *models.Passenger) error
	update   func(p *models.Passenger) error
	delete   func(id uint) error
}

func (m *mockPassengerRepo) FindAll(page, limit int) ([]models.Passenger, int64, error) {
	return m.findAll(page, limit)
}
func (m *mockPassengerRepo) FindByID(id uint) (models.Passenger, error) {
	return m.findByID(id)
}
func (m *mockPassengerRepo) Create(p *models.Passenger) error { return m.create(p) }
func (m *mockPassengerRepo) Update(p *models.Passenger) error { return m.update(p) }
func (m *mockPassengerRepo) Delete(id uint) error             { return m.delete(id) }

func newPassengerRouter(repo repository.PassengerRepository) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewPassengerHandler(repo)
	r.GET("/passengers", h.GetAll)
	r.POST("/passengers", h.Create)
	r.PUT("/passengers/:id", h.Update)
	r.DELETE("/passengers/:id", h.Delete)
	return r
}

func TestPassengerList_Success(t *testing.T) {
	mock := &mockPassengerRepo{
		findAll: func(page, limit int) ([]models.Passenger, int64, error) {
			return []models.Passenger{
				{ID: 1, FullName: "Иван Иванов", Email: "ivan@example.com", Phone: "+77001234567", PassportNum: "N1111111"},
				{ID: 2, FullName: "Пётр Петров", Email: "petr@example.com", Phone: "+77001234568", PassportNum: "N2222222"},
				{ID: 3, FullName: "Анна Смирнова", Email: "anna@example.com", Phone: "+77001234569", PassportNum: "N3333333"},
			}, 3, nil
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/passengers?page=1&limit=10", nil)
	newPassengerRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 3)
	assert.Equal(t, float64(3), body["total"])
}

func TestPassengerCreate_Success(t *testing.T) {
	mock := &mockPassengerRepo{
		create: func(p *models.Passenger) error {
			p.ID = 42
			return nil
		},
	}

	passenger := models.Passenger{
		FullName:    "Иван Иванов",
		Email:       "ivan@example.com",
		Phone:       "+77001234567",
		PassportNum: "N1234567",
	}
	body, _ := json.Marshal(passenger)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/passengers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newPassengerRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, float64(42), resp["id"])
}

func TestPassengerCreate_ValidationError(t *testing.T) {
	mock := &mockPassengerRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/passengers", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	newPassengerRouter(mock).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
