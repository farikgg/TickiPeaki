package main

import (
	"aviation/config"
	"aviation/handlers"
	"aviation/middleware"
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

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecret"
	}

	flightRepo := postgres.NewFlightRepo(db)
	passengerRepo := postgres.NewPassengerRepo(db)
	ticketRepo := postgres.NewTicketRepo(db)
	userRepo := postgres.NewUserRepo(db)

	flightHandler := handlers.NewFlightHandler(flightRepo)
	passengerHandler := handlers.NewPassengerHandler(passengerRepo)
	ticketHandler := handlers.NewTicketHandler(ticketRepo, flightRepo)
	authHandler := handlers.NewAuthHandler(userRepo)

	r := gin.Default()

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	protected := r.Group("/")
	protected.Use(middleware.Auth(jwtSecret))
	{
		protected.GET("/flights", flightHandler.GetAll)
		protected.POST("/flights", flightHandler.Create)
		protected.PUT("/flights/:id", flightHandler.Update)
		protected.DELETE("/flights/:id", flightHandler.Delete)

		protected.GET("/passengers", passengerHandler.GetAll)
		protected.POST("/passengers", passengerHandler.Create)
		protected.PUT("/passengers/:id", passengerHandler.Update)
		protected.DELETE("/passengers/:id", passengerHandler.Delete)

		protected.GET("/tickets", ticketHandler.GetAll)
		protected.POST("/tickets", ticketHandler.Create)
		protected.PUT("/tickets/:id", ticketHandler.Update)
		protected.DELETE("/tickets/:id", ticketHandler.Delete)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
