package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"ticketing/models"
	"ticketing/repository"
)

// TicketHandler обработчик HTTP-запросов для /tickets.
type TicketHandler struct {
	flights repository.FlightRepository
	tickets repository.TicketRepository
}

// NewTicketHandler создаёт обработчик билетов.
func NewTicketHandler(flights repository.FlightRepository, tickets repository.TicketRepository) *TicketHandler {
	return &TicketHandler{flights: flights, tickets: tickets}
}

// ticketListResponse ответ с пагинацией для GET /tickets.
type ticketListResponse struct {
	Data  []models.Ticket `json:"data"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Total int             `json:"total"`
}

// ticketInput — поля билета для создания и обновления.
type ticketInput struct {
	FlightID       int     `json:"flight_id"`
	PassengerName  string  `json:"passenger_name"`
	PassengerEmail string  `json:"passenger_email"`
	SeatNumber     string  `json:"seat_number"`
	Class          string  `json:"class"`
	Price          float64 `json:"price"`
	Status         string  `json:"status"`
}

// допустимые классы и статусы билетов
var validClasses = map[string]bool{"economy": true, "business": true, "first": true}
var validStatuses = map[string]bool{"reserved": true, "paid": true, "cancelled": true}

// List обрабатывает GET /tickets — список с фильтрами и пагинацией.
func (h *TicketHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := parsePagination(q)

	var flightID int
	if s := q.Get("flight_id"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			flightID = n
		}
	}

	filter := repository.TicketFilter{
		FlightID: flightID,
		Status:   strings.TrimSpace(q.Get("status")),
		Class:    strings.TrimSpace(q.Get("class")),
		Page:     page,
		Limit:    limit,
	}

	data, total := h.tickets.FindAll(filter)
	writeJSON(w, http.StatusOK, ticketListResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

// Create обрабатывает POST /tickets — бронируем место и создаём билет.
func (h *TicketHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input ticketInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	errs := validateTicketInput(input)

	// рейс должен существовать
	if input.FlightID > 0 {
		if _, ok := h.flights.FindByID(input.FlightID); !ok {
			errs["flight_id"] = "рейс не найден"
		}
	}

	if len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	// если мест нет — сразу отдаём 409
	if !h.flights.DecrementSeat(input.FlightID) {
		writeError(w, http.StatusConflict, "мест нет")
		return
	}

	t := models.Ticket{
		FlightID:       input.FlightID,
		PassengerName:  strings.TrimSpace(input.PassengerName),
		PassengerEmail: strings.TrimSpace(input.PassengerEmail),
		SeatNumber:     strings.TrimSpace(input.SeatNumber),
		Class:          strings.ToLower(strings.TrimSpace(input.Class)),
		Price:          input.Price,
		Status:         "reserved",
	}
	created := h.tickets.Create(t)
	writeJSON(w, http.StatusCreated, created)
}

// GetByID обрабатывает GET /tickets/{id}.
func (h *TicketHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/tickets/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный ID билета")
		return
	}

	ticket, ok := h.tickets.FindByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "билет не найден")
		return
	}

	writeJSON(w, http.StatusOK, ticket)
}

// Update обрабатывает PUT /tickets/{id} — при отмене возвращаем место обратно в рейс.
func (h *TicketHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/tickets/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный ID билета")
		return
	}

	existing, ok := h.tickets.FindByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "билет не найден")
		return
	}

	var input ticketInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	// при обновлении flight_id не меняется
	input.FlightID = existing.FlightID

	errs := validateTicketInput(input)
	if len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	newStatus := strings.ToLower(strings.TrimSpace(input.Status))

	// при отмене возвращаем место обратно
	if newStatus == "cancelled" && existing.Status != "cancelled" {
		h.flights.IncrementSeat(existing.FlightID)
	}

	existing.PassengerName = strings.TrimSpace(input.PassengerName)
	existing.PassengerEmail = strings.TrimSpace(input.PassengerEmail)
	existing.SeatNumber = strings.TrimSpace(input.SeatNumber)
	existing.Class = strings.ToLower(strings.TrimSpace(input.Class))
	existing.Price = input.Price
	existing.Status = newStatus

	updated, _ := h.tickets.Update(existing)
	writeJSON(w, http.StatusOK, updated)
}

// Delete обрабатывает DELETE /tickets/{id} — если билет не отменён, освобождаем место.
func (h *TicketHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/tickets/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный ID билета")
		return
	}

	ticket, ok := h.tickets.FindByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "билет не найден")
		return
	}

	// если билет не был отменён — возвращаем место
	if ticket.Status != "cancelled" {
		h.flights.IncrementSeat(ticket.FlightID)
	}

	h.tickets.Delete(ticket.ID)
	w.WriteHeader(http.StatusNoContent)
}

// ── валидация ─────────────────────────────────────────────────────────────────

// validateTicketInput проверяет поля билета, существование рейса проверяет хендлер.
func validateTicketInput(t ticketInput) map[string]string {
	errs := map[string]string{}

	if t.FlightID <= 0 {
		errs["flight_id"] = "обязательное поле"
	}
	if strings.TrimSpace(t.PassengerName) == "" {
		errs["passenger_name"] = "обязательное поле"
	}
	email := strings.TrimSpace(t.PassengerEmail)
	if email == "" {
		errs["passenger_email"] = "обязательное поле"
	} else if !strings.Contains(email, "@") {
		errs["passenger_email"] = "должен содержать @"
	}
	if strings.TrimSpace(t.SeatNumber) == "" {
		errs["seat_number"] = "обязательное поле"
	}
	class := strings.ToLower(strings.TrimSpace(t.Class))
	if class == "" {
		errs["class"] = "обязательное поле"
	} else if !validClasses[class] {
		errs["class"] = `должен быть "economy", "business" или "first"`
	}
	if t.Price <= 0 {
		errs["price"] = "должна быть > 0"
	}
	if t.Status != "" {
		if !validStatuses[strings.ToLower(strings.TrimSpace(t.Status))] {
			errs["status"] = `должен быть "reserved", "paid" или "cancelled"`
		}
	}

	return errs
}
