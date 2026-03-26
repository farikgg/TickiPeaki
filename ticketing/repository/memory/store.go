package memory

import (
	"strings"
	"sync"

	"ticketing/models"
	"ticketing/repository"
)

// Store — общее хранилище рейсов и билетов в памяти.
type Store struct {
	mu           sync.RWMutex
	flights      map[int]models.Flight
	tickets      map[int]models.Ticket
	nextFlightID int
	nextTicketID int
}

// NewStore создаёт хранилище с тестовыми данными.
func NewStore() *Store {
	s := &Store{
		flights: make(map[int]models.Flight),
		tickets: make(map[int]models.Ticket),
	}
	s.seed()
	return s
}

// flightRepo — обёртка над Store для реализации FlightRepository.
type flightRepo struct{ s *Store }

// ticketRepo — обёртка над Store для реализации TicketRepository.
type ticketRepo struct{ s *Store }

// Flights возвращает реализацию FlightRepository.
func (s *Store) Flights() repository.FlightRepository {
	return &flightRepo{s}
}

// Tickets возвращает реализацию TicketRepository.
func (s *Store) Tickets() repository.TicketRepository {
	return &ticketRepo{s}
}

// seed заполняет хранилище начальными данными.
func (s *Store) seed() {
	// рейсы — микс авиа и жд по Центральной Азии
	for _, f := range []models.Flight{
		{
			Origin: "Almaty", Destination: "Astana", Type: "air",
			Carrier:        "Air Astana",
			DepartureTime:  "2027-06-15T08:00:00Z",
			ArrivalTime:    "2027-06-15T10:00:00Z",
			AvailableSeats: 148, Price: 25000.00,
		},
		{
			Origin: "Almaty", Destination: "Shymkent", Type: "rail",
			Carrier:        "KTZ Express",
			DepartureTime:  "2027-06-16T07:00:00Z",
			ArrivalTime:    "2027-06-16T18:00:00Z",
			AvailableSeats: 200, Price: 8500.00,
		},
		{
			Origin: "Astana", Destination: "Almaty", Type: "air",
			Carrier:        "FlyArystan",
			DepartureTime:  "2027-06-17T14:00:00Z",
			ArrivalTime:    "2027-06-17T16:00:00Z",
			AvailableSeats: 180, Price: 18000.00,
		},
	} {
		s.nextFlightID++
		f.ID = s.nextFlightID
		s.flights[f.ID] = f
	}

	// билеты — два на первый рейс, поэтому у него 148 мест
	for _, t := range []models.Ticket{
		{
			FlightID: 1, PassengerName: "Айдар Касымов",
			PassengerEmail: "aidar@example.com",
			SeatNumber:     "12A", Class: "economy",
			Price: 25000.00, Status: "reserved",
		},
		{
			FlightID: 1, PassengerName: "Дана Нурланова",
			PassengerEmail: "dana@example.com",
			SeatNumber:     "3B", Class: "business",
			Price: 50000.00, Status: "paid",
		},
	} {
		s.nextTicketID++
		t.ID = s.nextTicketID
		s.tickets[t.ID] = t
	}
}

// ── FlightRepository ──────────────────────────────────────────────────────────

// FindAll отдаёт рейсы по фильтру с пагинацией.
func (r *flightRepo) FindAll(filter repository.FlightFilter) ([]models.Flight, int) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()

	filtered := make([]models.Flight, 0, len(r.s.flights))
	for _, f := range r.s.flights {
		if filter.Type != "" && !strings.EqualFold(f.Type, filter.Type) {
			continue
		}
		if filter.Origin != "" && !strings.EqualFold(f.Origin, filter.Origin) {
			continue
		}
		if filter.Destination != "" && !strings.EqualFold(f.Destination, filter.Destination) {
			continue
		}
		filtered = append(filtered, f)
	}

	total := len(filtered)
	start, end := paginate(total, filter.Page, filter.Limit)
	return filtered[start:end], total
}

// FindByID ищет рейс по ID, возвращает false если не нашли.
func (r *flightRepo) FindByID(id int) (models.Flight, bool) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	f, ok := r.s.flights[id]
	return f, ok
}

// Create сохраняет рейс, ID присваиваем сами.
func (r *flightRepo) Create(f models.Flight) models.Flight {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.nextFlightID++
	f.ID = r.s.nextFlightID
	r.s.flights[f.ID] = f
	return f
}

// Update заменяет рейс целиком, возвращает false если не нашли.
func (r *flightRepo) Update(f models.Flight) (models.Flight, bool) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.flights[f.ID]; !ok {
		return models.Flight{}, false
	}
	r.s.flights[f.ID] = f
	return f, true
}

// Delete удаляет рейс, возвращает false если не нашли.
func (r *flightRepo) Delete(id int) bool {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.flights[id]; !ok {
		return false
	}
	delete(r.s.flights, id)
	return true
}

// DecrementSeat уменьшает свободные места на 1, false если мест нет.
func (r *flightRepo) DecrementSeat(id int) bool {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	f, ok := r.s.flights[id]
	if !ok || f.AvailableSeats <= 0 {
		return false
	}
	f.AvailableSeats--
	r.s.flights[id] = f
	return true
}

// IncrementSeat освобождает место обратно.
func (r *flightRepo) IncrementSeat(id int) bool {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	f, ok := r.s.flights[id]
	if !ok {
		return false
	}
	f.AvailableSeats++
	r.s.flights[id] = f
	return true
}

// ── TicketRepository ──────────────────────────────────────────────────────────

// FindAll отдаёт билеты по фильтру с пагинацией.
func (r *ticketRepo) FindAll(filter repository.TicketFilter) ([]models.Ticket, int) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()

	filtered := make([]models.Ticket, 0, len(r.s.tickets))
	for _, t := range r.s.tickets {
		if filter.FlightID > 0 && t.FlightID != filter.FlightID {
			continue
		}
		if filter.Status != "" && !strings.EqualFold(t.Status, filter.Status) {
			continue
		}
		if filter.Class != "" && !strings.EqualFold(t.Class, filter.Class) {
			continue
		}
		filtered = append(filtered, t)
	}

	total := len(filtered)
	start, end := paginate(total, filter.Page, filter.Limit)
	return filtered[start:end], total
}

// FindByID ищет билет по ID, возвращает false если не нашли.
func (r *ticketRepo) FindByID(id int) (models.Ticket, bool) {
	r.s.mu.RLock()
	defer r.s.mu.RUnlock()
	t, ok := r.s.tickets[id]
	return t, ok
}

// Create сохраняет билет, ID присваиваем сами.
func (r *ticketRepo) Create(t models.Ticket) models.Ticket {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	r.s.nextTicketID++
	t.ID = r.s.nextTicketID
	r.s.tickets[t.ID] = t
	return t
}

// Update заменяет билет целиком, возвращает false если не нашли.
func (r *ticketRepo) Update(t models.Ticket) (models.Ticket, bool) {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.tickets[t.ID]; !ok {
		return models.Ticket{}, false
	}
	r.s.tickets[t.ID] = t
	return t, true
}

// Delete удаляет билет, возвращает false если не нашли.
func (r *ticketRepo) Delete(id int) bool {
	r.s.mu.Lock()
	defer r.s.mu.Unlock()
	if _, ok := r.s.tickets[id]; !ok {
		return false
	}
	delete(r.s.tickets, id)
	return true
}

// ── helpers ───────────────────────────────────────────────────────────────────

// paginate вычисляет границы среза для страницы.
func paginate(total, page, limit int) (start, end int) {
	start = (page - 1) * limit
	end = start + limit
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	return start, end
}
