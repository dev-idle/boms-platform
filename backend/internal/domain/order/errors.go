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
	ErrPickupSoldOut           = errors.New("an item in the order is sold out on that day")
	ErrInvalidGuestName        = errors.New("a guest's name needs 1 to 100 plain characters")
	ErrTicketNotFound          = errors.New("ticket not found")
	ErrInvalidTicketTransition = errors.New("invalid ticket status transition")
	ErrTicketOrderNotActive    = errors.New("the order is not being made")
	ErrTicketNotMovable        = errors.New("only a ticket nobody has started can move")
	ErrTicketStationTaken      = errors.New("the order already has a ticket at that station")
	ErrPickupDayLimit          = errors.New("the customer already has as many orders as one customer may book for that day")
	ErrInvalidCancelReason     = errors.New("a cancellation needs a reason of 1 to 200 plain characters")
	ErrPickupCodeInvalid       = errors.New("the pickup code does not match the order")
	ErrPickupCodeLocked        = errors.New("too many wrong pickup codes for the order")
)
