package main

import (
	"aviation/config"
	"aviation/handlers"
	"aviation/repository/postgres"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=aviation port=5432 sslmode=disable"
	}

	db, err := config.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}

	flightRepo := postgres.NewFlightRepo(db)
	passengerRepo := postgres.NewPassengerRepo(db)
	ticketRepo := postgres.NewTicketRepo(db)

	flightHandler := handlers.NewFlightHandler(flightRepo)
	passengerHandler := handlers.NewPassengerHandler(passengerRepo)
	ticketHandler := handlers.NewTicketHandler(ticketRepo, flightRepo)

	r := gin.Default()

	r.GET("/flights", flightHandler.GetAll)
	r.POST("/flights", flightHandler.Create)
	r.PUT("/flights/:id", flightHandler.Update)
	r.DELETE("/flights/:id", flightHandler.Delete)

	r.GET("/passengers", passengerHandler.GetAll)
	r.POST("/passengers", passengerHandler.Create)
	r.PUT("/passengers/:id", passengerHandler.Update)
	r.DELETE("/passengers/:id", passengerHandler.Delete)

	r.GET("/tickets", ticketHandler.GetAll)
	r.POST("/tickets", ticketHandler.Create)
	r.PUT("/tickets/:id", ticketHandler.Update)
	r.DELETE("/tickets/:id", ticketHandler.Delete)

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
