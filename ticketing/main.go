package main

import (
	"log"
	"net/http"

	"ticketing/handlers"
	"ticketing/repository/memory"
)

func main() {
	store := memory.NewStore()

	flightHandler := handlers.NewFlightHandler(store.Flights())
	ticketHandler := handlers.NewTicketHandler(store.Flights(), store.Tickets())

	mux := http.NewServeMux()

	// рейсы
	mux.HandleFunc("/flights", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			flightHandler.List(w, r)
		case http.MethodPost:
			flightHandler.Create(w, r)
		default:
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/flights/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			flightHandler.GetByID(w, r)
		case http.MethodPut:
			flightHandler.Update(w, r)
		case http.MethodDelete:
			flightHandler.Delete(w, r)
		default:
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	// билеты
	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ticketHandler.List(w, r)
		case http.MethodPost:
			ticketHandler.Create(w, r)
		default:
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/tickets/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ticketHandler.GetByID(w, r)
		case http.MethodPut:
			ticketHandler.Update(w, r)
		case http.MethodDelete:
			ticketHandler.Delete(w, r)
		default:
			http.Error(w, "метод не поддерживается", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Ticketing API запущен на :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}
