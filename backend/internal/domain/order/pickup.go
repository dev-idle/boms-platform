package order

import (
	"slices"
	"time"

	domainstore "github.com/boms/backend/internal/domain/store"
)

// PickupPolicy is the pickup window the bakery's settings allow at checkout.
// Until orders are told apart as instant or pre-order, every order waits the
// pre-order lead time.
type PickupPolicy struct {
	Settings domainstore.Settings
	// ClosedDays are bakery-local days, as domainstore.DayOf returns them.
	ClosedDays []time.Time
}

// Validate reports why at cannot be a pickup time, checked at now.
func (p PickupPolicy) Validate(at, now time.Time) error {
	if at.IsZero() {
		return ErrInvalidPickupAt
	}
	if at.Before(now.Add(p.Settings.PreorderMinLead)) {
		return ErrPickupTooSoon
	}
	if at.After(now.Add(time.Duration(p.Settings.MaxAdvanceDays) * 24 * time.Hour)) {
		return ErrPickupTooFar
	}
	if slices.ContainsFunc(p.ClosedDays, domainstore.DayOf(at).Equal) {
		return ErrPickupClosedDay
	}
	local := at.In(domainstore.Location)
	sinceMidnight := time.Duration(local.Hour())*time.Hour + time.Duration(local.Minute())*time.Minute
	if sinceMidnight < p.Settings.OpensAt || sinceMidnight >= p.Settings.ClosesAt {
		return ErrPickupOutsideHours
	}
	return nil
}
