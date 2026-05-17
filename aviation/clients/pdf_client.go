package clients

import (
	"aviation/models"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

type PDFClient struct {
	client  *resty.Client
	baseURL string
}

type ticketPayload struct {
	TicketID       uint    `json:"ticket_id"`
	FlightNumber   string  `json:"flight_number"`
	Origin         string  `json:"origin"`
	Destination    string  `json:"destination"`
	DepartureTime  string  `json:"departure_time"`
	ArrivalTime    string  `json:"arrival_time"`
	Carrier        string  `json:"carrier"`
	PassengerName  string  `json:"passenger_name"`
	PassengerEmail string  `json:"passenger_email"`
	SeatNumber     string  `json:"seat_number"`
	Class          string  `json:"class"`
	Price          float64 `json:"price"`
}

func NewPDFClient(baseURL string) *PDFClient {
	client := resty.New()

	// логируем каждый исходящий запрос
	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		log.Printf("[pdf-client] %s %s", req.Method, req.URL)
		return nil
	})

	// логируем каждый ответ
	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		log.Printf("[pdf-client] статус ответа: %d", resp.StatusCode())
		return nil
	})

	return &PDFClient{client: client, baseURL: baseURL}
}

func (p *PDFClient) GenerateTicket(ticket models.Ticket) error {
	payload := ticketPayload{
		TicketID:       ticket.ID,
		FlightNumber:   ticket.Flight.FlightNumber,
		Origin:         ticket.Flight.Origin,
		Destination:    ticket.Flight.Destination,
		DepartureTime:  ticket.Flight.DepartureTime.Format(time.RFC3339),
		ArrivalTime:    ticket.Flight.ArrivalTime.Format(time.RFC3339),
		Carrier:        ticket.Flight.Carrier,
		PassengerName:  ticket.Passenger.FullName,
		PassengerEmail: ticket.Passenger.Email,
		SeatNumber:     ticket.SeatNumber,
		Class:          ticket.Class,
		Price:          ticket.Price,
	}

	resp, err := p.client.R().
		SetBody(payload).
		Post(fmt.Sprintf("%s/api/v1/generate_ticket", p.baseURL))
	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("pdf-service вернул %d", resp.StatusCode())
	}

	return nil
}
