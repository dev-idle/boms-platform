package order

import "errors"

var (
	ErrNotFound                = errors.New("order not found")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
	ErrInvalidPickupAt         = errors.New("invalid pickup time")
	ErrPickupTooSoon           = errors.New("pickup time is sooner than the bakery can prepare the order")
	ErrPickupTooFar            = errors.New("pickup time is further ahead than the bakery takes orders")
	ErrPickupClosedDay         = errors.New("the bakery is closed on that day")
	ErrPickupOutsideHours      = errors.New("pickup time is outside opening hours")
	ErrPickupOffSlot           = errors.New("pickup time is not the start of a pickup slot")
	ErrPickupSlotFull          = errors.New("that pickup slot is full")
	ErrPickupDayLimit          = errors.New("the customer already has as many orders as one customer may book for that day")
)
