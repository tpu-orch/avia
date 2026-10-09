package rest

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tpu-orch/avia/internal/delivery/rest/dto"
	"github.com/tpu-orch/avia/internal/domain/models"
	service "github.com/tpu-orch/avia/internal/domain/services/avia"
)

type TicketQueryService interface {
	GetCities(ctx context.Context, query string) ([]string, error)
	GetTickets(ctx context.Context, from, to string) ([]models.Ticket, error)
	GetReservationStatus(ctx context.Context, ticketID int64) (models.ReservationStatus, time.Time, error)
}

type Handler struct {
	ticketQueryService TicketQueryService
}

func NewHandler(ticketQueryService TicketQueryService) *Handler {
	return &Handler{
		ticketQueryService: ticketQueryService,
	}
}

func (h *Handler) RegisterRoutes(r *gin.Engine) {
	r.GET("/cities", h.SearchCities)
	r.GET("/tickets", h.SearchTickets)
	r.GET("/tickets/:ticketId/reservation-status", h.GetReservationStatus)
}

func (h *Handler) SearchCities(c *gin.Context) {
	query := c.Query("query")

	cities, err := h.ticketQueryService.GetCities(c.Request.Context(), query)
	if err != nil {
		if errors.Is(err, service.ErrInvalidParameter) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PARAMETER",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, cities)
}

func (h *Handler) SearchTickets(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	tickets, err := h.ticketQueryService.GetTickets(c.Request.Context(), from, to)
	if err != nil {
		if errors.Is(err, service.ErrInvalidParameter) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PARAMETER",
				Message: err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
		return
	}

	var respTickets []dto.Ticket
	for _, t := range tickets {
		respTickets = append(respTickets, dto.Ticket{
			ID:    t.ID,
			Price: t.Price,
			Seat:  t.Seat,
			Flight: dto.Flight{
				ID:            t.Flight.ID,
				DepartureDate: t.Flight.DepartureDate.Format("2006-01-02T15:04:05Z07:00"),
				ArrivalDate:   t.Flight.ArrivalDate.Format("2006-01-02T15:04:05Z07:00"),
				From: dto.Airport{
					ID:        t.Flight.From.ID,
					CodeIata:  t.Flight.From.CodeIata,
					City:      t.Flight.From.City,
					Longitude: t.Flight.From.Longitude,
					Latitude:  t.Flight.From.Latitude,
				},
				To: dto.Airport{
					ID:        t.Flight.To.ID,
					CodeIata:  t.Flight.To.CodeIata,
					City:      t.Flight.To.City,
					Longitude: t.Flight.To.Longitude,
					Latitude:  t.Flight.To.Latitude,
				},
				IsCancelled: t.Flight.IsCancelled,
			},
			ReservationStatus: string(t.ReservationStatus),
		})
	}

	if respTickets == nil {
		respTickets = []dto.Ticket{}
	}

	c.JSON(http.StatusOK, respTickets)
}

func (h *Handler) GetReservationStatus(c *gin.Context) {
	ticketIDStr := c.Param("ticketId")
	ticketID, err := strconv.ParseInt(ticketIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Code:    "INVALID_PARAMETER",
			Message: "Invalid ticket ID.",
		})
		return
	}

	status, updatedAt, err := h.ticketQueryService.GetReservationStatus(c.Request.Context(), ticketID)
	if err != nil {
		if errors.Is(err, service.ErrInvalidParameter) {
			c.JSON(http.StatusBadRequest, dto.ErrorResponse{
				Code:    "INVALID_PARAMETER",
				Message: err.Error(),
			})
			return
		}
		if errors.Is(err, service.ErrTicketNotFound) {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{
				Code:    "NOT_FOUND",
				Message: "Ticket not found.",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Code:    "INTERNAL_ERROR",
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, dto.ReservationStatusResponse{
		TicketID:  ticketID,
		Status:    string(status),
		UpdatedAt: updatedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
