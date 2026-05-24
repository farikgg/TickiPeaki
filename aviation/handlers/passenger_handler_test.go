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
	findByID func(uint) (models.Passenger, error)
	create   func(*models.Passenger) error
	update   func(*models.Passenger) error
	delete   func(uint) error
}

func (m *mockPassengerRepo) FindAll(page, limit int) ([]models.Passenger, int64, error) {
	if m.findAll == nil {
		return nil, 0, nil
	}
	return m.findAll(page, limit)
}
func (m *mockPassengerRepo) FindByID(id uint) (models.Passenger, error) {
	if m.findByID == nil {
		return models.Passenger{}, nil
	}
	return m.findByID(id)
}
func (m *mockPassengerRepo) Create(p *models.Passenger) error {
	if m.create == nil {
		return nil
	}
	return m.create(p)
}
func (m *mockPassengerRepo) Update(p *models.Passenger) error {
	if m.update == nil {
		return nil
	}
	return m.update(p)
}
func (m *mockPassengerRepo) Delete(id uint) error {
	if m.delete == nil {
		return nil
	}
	return m.delete(id)
}

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
	repo := &mockPassengerRepo{
		findAll: func(page, limit int) ([]models.Passenger, int64, error) {
			return []models.Passenger{
				{ID: 1, FullName: "Рафаэль Ахметов", Email: "rafi@example.com", Phone: "+77011234567", PassportNum: "N12345678"},
				{ID: 2, FullName: "Айгерим Бекова", Email: "aigerim@example.com", Phone: "+77029876543", PassportNum: "N87654321"},
				{ID: 3, FullName: "Данияр Сейткали", Email: "daniyar@example.com", Phone: "+77031112233", PassportNum: "N11223344"},
			}, 3, nil
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/passengers?page=1&limit=10", nil)
	newPassengerRouter(repo).ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))

	data, ok := body["data"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, data, 3)
	assert.Equal(t, float64(3), body["total"])
}

func TestPassengerCreate_Success(t *testing.T) {
	repo := &mockPassengerRepo{
		create: func(p *models.Passenger) error {
			p.ID = 42
			return nil
		},
	}

	payload := models.Passenger{
		FullName:    "Рафаэль Ахметов",
		Email:       "rafi@example.com",
		Phone:       "+77011234567",
		PassportNum: "N12345678",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/passengers", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	newPassengerRouter(repo).ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]interface{}
	assert.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "rafi@example.com", resp["email"])
}

func TestPassengerCreate_ValidationError(t *testing.T) {
	repo := &mockPassengerRepo{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/passengers", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	newPassengerRouter(repo).ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
}
