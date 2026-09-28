package order

import (
	"slices"
	"time"

	domainstore "github.com/boms/backend/internal/domain/store"
)

// MaxOrdersPerCustomerPerDay is how many orders not cancelled one customer may
// hold for pickup on a bakery day. Slots are shared: without a cap, one account
// could fill every slot of a day with orders it never collects.
const MaxOrdersPerCustomerPerDay = 3

// PickupPolicy is the pickup window the bakery's settings allow at checkout.
type PickupPolicy struct {
	Settings domainstore.Settings
	// ClosedDays are bakery-local days, as domainstore.DayOf returns them.
	ClosedDays []time.Time
}

// Validate decides how an order with these items is prepared if collected at
// at, checked at now, or reports why at cannot be its pickup time. An order of
// counter items only, collected the day it is placed, is instant and waits the
// counter's preparation time; any other order is a pre-order and waits the
// pre-order notice. Either waits the longest notice one of its items needs.
// Slot capacity is not checked here: it needs the orders already booked.
func (p PickupPolicy) Validate(at, now time.Time, items Fulfillment) (Type, error) {
	if at.IsZero() {
		return "", ErrInvalidPickupAt
	}
	orderType, prep := TypePreOrder, p.Settings.PreorderMinLead
	if !items.Kitchen && domainstore.DayOf(at).Equal(domainstore.DayOf(now)) {
		orderType, prep = TypeInstant, p.Settings.InstantPrep
	}
	if at.Before(now.Add(max(prep, items.Lead))) {
		return "", ErrPickupTooSoon
	}
	if at.After(now.Add(time.Duration(p.Settings.MaxAdvanceDays) * 24 * time.Hour)) {
		return "", ErrPickupTooFar
	}
	if slices.ContainsFunc(p.ClosedDays, domainstore.DayOf(at).Equal) {
		return "", ErrPickupClosedDay
	}
	local := at.In(domainstore.Location)
	sinceMidnight := local.Sub(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, domainstore.Location))
	if sinceMidnight < p.Settings.OpensAt || sinceMidnight >= p.Settings.ClosesAt {
		return "", ErrPickupOutsideHours
	}
	if (sinceMidnight-p.Settings.OpensAt)%p.Settings.SlotLength != 0 {
		return "", ErrPickupOffSlot
	}
	return orderType, nil
}
