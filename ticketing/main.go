package main

import (
	"log"

	"github.com/gin-gonic/gin"

	"ticketing/handlers"
	"ticketing/repository/memory"
)

func main() {
	store := memory.NewStore()

	flightHandler := handlers.NewFlightHandler(store.Flights())
	ticketHandler := handlers.NewTicketHandler(store.Flights(), store.Tickets())

	r := gin.Default()

	r.GET("/flights", flightHandler.List)
	r.POST("/flights", flightHandler.Create)
	r.GET("/flights/:id", flightHandler.GetByID)
	r.PUT("/flights/:id", flightHandler.Update)
	r.DELETE("/flights/:id", flightHandler.Delete)

	r.GET("/tickets", ticketHandler.List)
	r.POST("/tickets", ticketHandler.Create)
	r.GET("/tickets/:id", ticketHandler.GetByID)
	r.PUT("/tickets/:id", ticketHandler.Update)
	r.DELETE("/tickets/:id", ticketHandler.Delete)

	log.Println("Ticketing API запущен на :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("ошибка сервера: %v", err)
	}
}
