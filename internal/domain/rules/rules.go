package rules

import "github.com/tpu-orch/avia/internal/domain/models"

func IsValidStatus(status models.ReservationStatus) bool {
	switch status {
	case models.StatusAvailable,
		models.StatusReservationPending,
		models.StatusReserved,
		models.StatusCancellationPending,
		models.StatusCancelled:
		return true
	default:
		return false
	}
}

func IsValidTransition(current, target models.ReservationStatus) bool {
	switch current {
	case models.StatusAvailable:
		return target == models.StatusReservationPending || target == models.StatusReserved
	case models.StatusReservationPending:
		return target == models.StatusReserved || target == models.StatusAvailable
	case models.StatusReserved:
		return target == models.StatusCancellationPending || target == models.StatusCancelled
	case models.StatusCancellationPending:
		return target == models.StatusCancelled || target == models.StatusReserved
	case models.StatusCancelled:
		return false
	default:
		return false
	}
}
