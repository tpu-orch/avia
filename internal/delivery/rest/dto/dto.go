package dto

type Airport struct {
	ID        int64   `json:"id"`
	CodeIata  string  `json:"codeIata"`
	City      string  `json:"city"`
	Longitude float64 `json:"longitude,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
}

type Flight struct {
	ID            int64   `json:"id"`
	DepartureDate string  `json:"departureDate"`
	ArrivalDate   string  `json:"arrivalDate"`
	From          Airport `json:"from"`
	To            Airport `json:"to"`
	IsCancelled   bool    `json:"isCancelled"`
}

type Ticket struct {
	ID                int64   `json:"id"`
	Price             float64 `json:"price"`
	Seat              string  `json:"seat"`
	Flight            Flight  `json:"flight"`
	ReservationStatus string  `json:"reservationStatus"`
}

type ReservationStatusResponse struct {
	TicketID  int64  `json:"ticketId"`
	Status    string `json:"status"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}
