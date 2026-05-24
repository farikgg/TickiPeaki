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

type generateRequest struct {
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

type generateResponse struct {
	TicketID uint   `json:"ticket_id"`
	Status   string `json:"status"`
}

type statusResponse struct {
	Status string  `json:"status"`
	URL    *string `json:"url"`
}

func NewPDFClient(baseURL string) *PDFClient {
	client := resty.New()

	client.OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
		log.Printf("[pdf-client] %s %s", req.Method, req.URL)
		return nil
	})

	client.OnAfterResponse(func(c *resty.Client, resp *resty.Response) error {
		log.Printf("[pdf-client] статус ответа: %d", resp.StatusCode())
		return nil
	})

	return &PDFClient{client: client, baseURL: baseURL}
}

func (p *PDFClient) RequestGeneration(ticket models.Ticket) error {
	payload := generateRequest{
		TicketID:       ticket.ID,
		FlightNumber:   ticket.Flight.FlightNumber,
		Origin:         ticket.Flight.Origin,
		Destination:    ticket.Flight.Destination,
		DepartureTime:  ticket.Flight.DepartureTime.Format(time.RFC3339),
		ArrivalTime:    ticket.Flight.ArrivalTime.Format(time.RFC3339),
		Carrier:        ticket.Flight.Carrier,
		PassengerName:  ticket.Passenger.FullName,
		PassengerEmail: ticket.Passenger.Email,
		SeatNumber:     ticket.Seat.SeatNumber,
		Class:          ticket.Seat.Class,
		Price:          ticket.Seat.Price,
	}

	var result generateResponse
	resp, err := p.client.R().
		SetBody(payload).
		SetResult(&result).
		Post(fmt.Sprintf("%s/api/v1/ticket_generate", p.baseURL))
	if err != nil {
		return err
	}

	if resp.StatusCode() != 202 {
		return fmt.Errorf("pdf-service вернул %d", resp.StatusCode())
	}

	return nil
}

func (p *PDFClient) PollStatus(ticketID uint) (*string, error) {
	maxAttempts := 15
	interval := 2 * time.Second

	for i := 0; i < maxAttempts; i++ {
		var result statusResponse
		resp, err := p.client.R().
			SetResult(&result).
			Get(fmt.Sprintf("%s/api/v1/ticket_status/%d", p.baseURL, ticketID))
		if err != nil {
			return nil, err
		}

		if resp.StatusCode() == 404 {
			return nil, fmt.Errorf("ticket %d не найден в pdf-service", ticketID)
		}

		if result.Status == "ready" && result.URL != nil {
			return result.URL, nil
		}

		if result.Status == "failed" {
			return nil, fmt.Errorf("pdf-service не смог сгенерировать билет %d", ticketID)
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("pdf-service не ответил за %d попыток", maxAttempts)
}
