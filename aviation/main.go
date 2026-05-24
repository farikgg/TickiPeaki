package main

import (
	"aviation/clients"
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
		dsn = "host=database user=postgres password=postgres dbname=aviation port=5432 sslmode=disable"
	}

	pdfServiceURL := os.Getenv("PDF_SERVICE_URL")
	if pdfServiceURL == "" {
		pdfServiceURL = "http://localhost:8000"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "supersecret"
	}

	db, err := config.Connect(dsn)
	if err != nil {
		log.Fatal(err)
	}

	flightRepo := postgres.NewFlightRepo(db)
	seatRepo := postgres.NewSeatRepo(db)
	passengerRepo := postgres.NewPassengerRepo(db)
	ticketRepo := postgres.NewTicketRepo(db)
	userRepo := postgres.NewUserRepo(db)

	pdfClient := clients.NewPDFClient(pdfServiceURL)

	flightHandler := handlers.NewFlightHandler(flightRepo, seatRepo)
	seatHandler := handlers.NewSeatHandler(seatRepo, flightRepo)
	passengerHandler := handlers.NewPassengerHandler(passengerRepo)
	ticketHandler := handlers.NewTicketHandler(ticketRepo, flightRepo, seatRepo, userRepo, pdfClient)
	authHandler := handlers.NewAuthHandler(userRepo, passengerRepo, jwtSecret)

	r := gin.Default()
	r.Use(middleware.CORS())

	r.POST("/register", authHandler.Register)
	r.POST("/login", authHandler.Login)

	protected := r.Group("/")
	protected.Use(middleware.Auth(jwtSecret))
	{
		protected.GET("/me", authHandler.Me)
		protected.POST("/me/passenger", authHandler.CreatePassenger)

		protected.GET("/flights", flightHandler.GetAll)
		protected.GET("/flights/:id", flightHandler.GetByID)
		protected.POST("/flights", flightHandler.Create)
		protected.PUT("/flights/:id", flightHandler.Update)
		protected.DELETE("/flights/:id", flightHandler.Delete)

		protected.GET("/flights/:id/seats", seatHandler.ListByFlight)
		protected.GET("/seats/:id", seatHandler.GetByID)
		protected.POST("/seats", seatHandler.Create)
		protected.PUT("/seats/:id", seatHandler.Update)
		protected.DELETE("/seats/:id", seatHandler.Delete)

		protected.GET("/passengers", passengerHandler.GetAll)
		protected.POST("/passengers", passengerHandler.Create)
		protected.PUT("/passengers/:id", passengerHandler.Update)
		protected.DELETE("/passengers/:id", passengerHandler.Delete)

		protected.GET("/tickets", ticketHandler.GetAll)
		protected.POST("/tickets", ticketHandler.Create)
		protected.PUT("/tickets/:id", ticketHandler.Update)
		protected.POST("/tickets/:id/pay", ticketHandler.Pay)
		protected.DELETE("/tickets/:id", ticketHandler.Delete)
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
