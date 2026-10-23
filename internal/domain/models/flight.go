package models

import "time"

type Flight struct {
	ID            int64
	DepartureDate time.Time
	ArrivalDate   time.Time
	From          Airport
	To            Airport
	IsCancelled   bool
}
