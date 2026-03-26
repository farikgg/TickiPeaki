package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ticketing/models"
	"ticketing/repository"
)

// FlightHandler обработчик HTTP-запросов для /flights.
type FlightHandler struct {
	repo repository.FlightRepository
}

// NewFlightHandler создаёт обработчик рейсов.
func NewFlightHandler(repo repository.FlightRepository) *FlightHandler {
	return &FlightHandler{repo: repo}
}

// flightListResponse ответ с пагинацией для GET /flights.
type flightListResponse struct {
	Data  []models.Flight `json:"data"`
	Page  int             `json:"page"`
	Limit int             `json:"limit"`
	Total int             `json:"total"`
}

// List обрабатывает GET /flights — рейсы с фильтрацией и пагинацией.
func (h *FlightHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	page, limit := parsePagination(q)

	filter := repository.FlightFilter{
		Type:        strings.TrimSpace(q.Get("type")),
		Origin:      strings.TrimSpace(q.Get("origin")),
		Destination: strings.TrimSpace(q.Get("destination")),
		Page:        page,
		Limit:       limit,
	}

	data, total := h.repo.FindAll(filter)
	writeJSON(w, http.StatusOK, flightListResponse{
		Data:  data,
		Page:  page,
		Limit: limit,
		Total: total,
	})
}

// Create обрабатывает POST /flights — создаём рейс.
func (h *FlightHandler) Create(w http.ResponseWriter, r *http.Request) {
	var f models.Flight
	if err := json.NewDecoder(r.Body).Decode(&f); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	errs := validateFlight(f)
	if len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	created := h.repo.Create(f)
	writeJSON(w, http.StatusCreated, created)
}

// GetByID обрабатывает GET /flights/{id}.
func (h *FlightHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/flights/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный ID рейса")
		return
	}

	f, ok := h.repo.FindByID(id)
	if !ok {
		writeError(w, http.StatusNotFound, "не нашли рейс")
		return
	}

	writeJSON(w, http.StatusOK, f)
}

// Update обрабатывает PUT /flights/{id} — обновляем все поля кроме id.
func (h *FlightHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/flights/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный ID рейса")
		return
	}

	if _, ok := h.repo.FindByID(id); !ok {
		writeError(w, http.StatusNotFound, "не нашли рейс")
		return
	}

	var input models.Flight
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeError(w, http.StatusBadRequest, "невалидный JSON")
		return
	}

	errs := validateFlight(input)
	if len(errs) > 0 {
		writeValidationErrors(w, errs)
		return
	}

	input.ID = id
	updated, _ := h.repo.Update(input)
	writeJSON(w, http.StatusOK, updated)
}

// Delete обрабатывает DELETE /flights/{id}.
func (h *FlightHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := extractID(r.URL.Path, "/flights/")
	if err != nil {
		writeError(w, http.StatusBadRequest, "невалидный ID рейса")
		return
	}

	if !h.repo.Delete(id) {
		writeError(w, http.StatusNotFound, "не нашли рейс")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ── валидация ─────────────────────────────────────────────────────────────────

// validateFlight проверяет все обязательные поля рейса.
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

// validateTimes проверяет формат RFC3339 и что вылет раньше прилёта.
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

	// проверяем порядок только если оба времени валидны
	if _, ok := errs["departure_time"]; !ok {
		if _, ok := errs["arrival_time"]; !ok {
			if !dep.Before(arr) {
				errs["departure_time"] = "должен быть раньше arrival_time"
			}
		}
	}
}

// ── общие хелперы ─────────────────────────────────────────────────────────────

// writeJSON пишет JSON-ответ с нужным статусом.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v) //nolint:errcheck
}

// writeError пишет ошибку в формате {"error": "..."}.
func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// writeValidationErrors пишет 422 с ошибками валидации по полям.
func writeValidationErrors(w http.ResponseWriter, errs map[string]string) {
	writeJSON(w, http.StatusUnprocessableEntity, map[string]map[string]string{"errors": errs})
}

// extractID вытаскивает числовой ID из URL-пути.
func extractID(path, prefix string) (int, error) {
	s := strings.TrimPrefix(path, prefix)
	if idx := strings.Index(s, "/"); idx != -1 {
		s = s[:idx]
	}
	return strconv.Atoi(s)
}

// parsePagination читает page и limit из query, подставляет дефолты.
func parsePagination(query interface{ Get(string) string }) (page, limit int) {
	page = 1
	limit = 10
	if s := query.Get("page"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 1 {
			page = n
		}
	}
	if s := query.Get("limit"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n >= 1 {
			limit = n
		}
	}
	if limit > 100 {
		limit = 100
	}
	return page, limit
}
