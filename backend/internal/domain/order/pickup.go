package order

import (
	"time"
)

const (
	pickupMinLeadTime = 2 * time.Hour
	pickupMaxAdvance  = 14 * 24 * time.Hour
	pickupOpenHour    = 8
	pickupCloseHour   = 18
)

var bakeryLocation = time.FixedZone("Asia/Ho_Chi_Minh", 7*60*60)

// ValidatePickupAt enforces bakery pickup window rules at checkout.
func ValidatePickupAt(at, now time.Time) error {
	if at.IsZero() {
		return ErrInvalidPickupAt
	}
	localAt := at.In(bakeryLocation)
	localNow := now.In(bakeryLocation)

	earliest := localNow.Add(pickupMinLeadTime)
	if localAt.Before(earliest) {
		return ErrInvalidPickupAt
	}
	latest := localNow.Add(pickupMaxAdvance)
	if localAt.After(latest) {
		return ErrInvalidPickupAt
	}

	hour := localAt.Hour()
	if hour < pickupOpenHour || hour >= pickupCloseHour {
		return ErrInvalidPickupAt
	}
	return nil
}
