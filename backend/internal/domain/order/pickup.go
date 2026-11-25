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
	return p.validate(at, now, now, items)
}

// Accepts checks, as staff accept an order placed at placedAt, that its pickup
// at is still ahead of now and still follows the rules, the notice its items
// need counted from placedAt: the time the order waited for staff is the
// bakery's to make up, not the customer's. The booking window counts from now,
// which holds a pickup booked at placedAt and one moved since.
func (p PickupPolicy) Accepts(at, placedAt, now time.Time, items Fulfillment) error {
	if !at.After(now) {
		return ErrPickupTooSoon
	}
	_, err := p.validate(at, placedAt, now, items)
	return err
}

// validate is Validate with the notice counted from noticeFrom and the booking
// window from now.
func (p PickupPolicy) validate(at, noticeFrom, now time.Time, items Fulfillment) (Type, error) {
	if at.IsZero() {
		return "", ErrInvalidPickupAt
	}
	orderType, prep := TypePreOrder, p.Settings.PreorderMinLead
	if !items.Kitchen && domainstore.DayOf(at).Equal(domainstore.DayOf(noticeFrom)) {
		orderType, prep = TypeInstant, p.Settings.InstantPrep
	}
	if at.Before(noticeFrom.Add(max(prep, items.Lead))) {
		return "", ErrPickupTooSoon
	}
	if err := p.fits(at, now, items.SoldOutOn); err != nil {
		return "", err
	}
	return orderType, nil
}

// Moves checks a pickup staff move an order to: ahead of now and within the
// bakery's days, hours and slots. Whether the kitchen still has the notice it
// needs is the bakery's own call, as for an order already made.
func (p PickupPolicy) Moves(at, now time.Time) error {
	if at.IsZero() {
		return ErrInvalidPickupAt
	}
	if !at.After(now) {
		return ErrPickupTooSoon
	}
	return p.fits(at, now, nil)
}

// fits checks what every pickup time meets, whoever books it: within the
// booking window from now, on a day the bakery opens and an item has not sold
// out, in its hours and on a slot.
func (p PickupPolicy) fits(at, now time.Time, soldOutOn *time.Time) error {
	if at.After(now.Add(time.Duration(p.Settings.MaxAdvanceDays) * 24 * time.Hour)) {
		return ErrPickupTooFar
	}
	if slices.ContainsFunc(p.ClosedDays, domainstore.DayOf(at).Equal) {
		return ErrPickupClosedDay
	}
	if soldOutOn != nil && domainstore.DayOf(at).Equal(*soldOutOn) {
		return ErrPickupSoldOut
	}
	local := at.In(domainstore.Location)
	sinceMidnight := local.Sub(time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, domainstore.Location))
	if sinceMidnight < p.Settings.OpensAt || sinceMidnight >= p.Settings.ClosesAt {
		return ErrPickupOutsideHours
	}
	if (sinceMidnight-p.Settings.OpensAt)%p.Settings.SlotLength != 0 {
		return ErrPickupOffSlot
	}
	return nil
}
