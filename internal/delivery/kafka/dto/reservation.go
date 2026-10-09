package dto

type ReservationStatusChangedPayload struct {
	CompositeBookingID int64  `json:"compositeBookingId"`
	TicketID           int64  `json:"ticketId"`
	Status             string `json:"status"`
}
