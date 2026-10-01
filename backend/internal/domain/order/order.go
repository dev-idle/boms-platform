package order

import (
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	"github.com/google/uuid"
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

// Order is a placed checkout snapshot with server-computed totals.
type Order struct {
	ID                   uuid.UUID
	Code                 string
	UserID               uuid.UUID
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
