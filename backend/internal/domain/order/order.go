package order

import (
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainstore "github.com/boms/backend/internal/domain/store"
)

// Status is the lifecycle state of a customer order.
type Status string

const (
	// StatusAwaitingPayment is an order placed online and not paid yet: it
	// holds its pickup slot and discount, and the bakery does not see it.
	StatusAwaitingPayment Status = "awaiting_payment"
	StatusPending         Status = "pending"
	StatusConfirmed       Status = "confirmed"
	StatusInProduction    Status = "in_production"
	StatusReady           Status = "ready"
	StatusCancelled       Status = "cancelled"
	StatusFulfilled       Status = "fulfilled"
	// StatusExpired is an order that was not paid in time: it gave back its
	// pickup slot and discount.
	StatusExpired Status = "expired"
	// StatusNoShow is an order made and not collected by closing time on its
	// pickup day; it is not refunded.
	StatusNoShow Status = "no_show"
)

func (s Status) Valid() bool {
	switch s {
	case StatusAwaitingPayment, StatusPending, StatusConfirmed, StatusInProduction, StatusReady, StatusCancelled, StatusFulfilled, StatusExpired, StatusNoShow:
		return true
	default:
		return false
	}
}

// Dropped reports whether an order in this status will not be made: its
// tickets are cancelled and its pickup slot is free again.
func (s Status) Dropped() bool {
	return s == StatusCancelled || s == StatusExpired
}

// BeforeProduction reports whether an order in this status is not being made
// yet: its customer may still cancel it or move its pickup, and a cancellation
// gives its discount use back.
func (s Status) BeforeProduction() bool {
	return s == StatusAwaitingPayment || s == StatusPending || s == StatusConfirmed
}

// Channel is where an order was taken.
type Channel string

const (
	// ChannelOnline is an order its customer placed and paid online.
	ChannelOnline Channel = "online"
	// ChannelCounter and ChannelPhone are orders staff took for a customer or a
	// guest; they are paid in cash when collected.
	ChannelCounter Channel = "counter"
	ChannelPhone   Channel = "phone"
)

// TakenByStaff reports whether staff took the order, at the counter or on the phone.
func (c Channel) TakenByStaff() bool {
	return c == ChannelCounter || c == ChannelPhone
}

// Guest is who collects an order taken for someone without an account.
type Guest struct {
	Name  string
	Phone string
}

// maxGuestNameLength is the longest name staff write down for a guest.
const maxGuestNameLength = 100

// NewGuest checks who collects a guest's order: a name of 1 to 100 plain
// characters, trimmed, and phone in the form it is stored in.
func NewGuest(name, phone string) (Guest, error) {
	name = strings.TrimSpace(name)
	if name == "" || utf8.RuneCountInString(name) > maxGuestNameLength || !domainstore.PlainText(name) {
		return Guest{}, ErrInvalidGuestName
	}
	return Guest{Name: name, Phone: phone}, nil
}

// Order is a placed checkout snapshot with server-computed totals.
type Order struct {
	ID   uuid.UUID
	Code string
	// UserID is the customer's account; nil for a guest's order, which Guest names.
	UserID               *uuid.UUID
	Guest                *Guest
	Channel              Channel
	Status               Status
	Type                 Type
	SubtotalCents        int64
	DiscountCents        int64
	TotalCents           int64
	DiscountCodeID       *uuid.UUID
	DiscountCodeSnapshot *string
	PickupAt             *time.Time
	// PaymentDueAt is when an order placed online expires if still unpaid.
	PaymentDueAt *time.Time
	// Terms is the policy version the customer accepted placing the order; orders
	// placed before acceptance was recorded have none.
	Terms     *domainpolicy.Acceptance
	CreatedAt time.Time
	UpdatedAt time.Time
}

// HasPickupCode reports whether the customer collects the order with a pickup
// code now: one placed online, from when the bakery accepts it until it is
// collected. An order staff took is handed over as its cash is paid.
func (o Order) HasPickupCode() bool {
	return o.Channel == ChannelOnline && o.Status.HasPickupCode()
}

// Payable reports whether the order still takes its payment at now: it awaits
// one and its time to pay has not run out.
func (o Order) Payable(now time.Time) bool {
	return o.Status == StatusAwaitingPayment && o.PaymentDueAt != nil && now.Before(*o.PaymentDueAt)
}

// PagePath is where the storefront shows a customer their order
// (ROUTE.orderDetail in frontend/src/constants/routes.ts): emails link to it,
// and the payment provider sends the buyer back to it.
func PagePath(id uuid.UUID) string {
	return "/orders/" + id.String()
}

// Item is an immutable order line captured at checkout.
type Item struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	LineType  domaincart.LineType
	ProductID *uuid.UUID
	ComboID   *uuid.UUID
	// Customization is how the customer configured the item, nil for a plain one.
	Customization  *Customization
	Name           string
	Slug           string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}
