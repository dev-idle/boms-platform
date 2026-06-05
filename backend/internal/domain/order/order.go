package order

import (
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	"github.com/google/uuid"
)

// Status is the lifecycle state of a customer order.
type Status string

const (
	StatusPending   Status = "pending"
	StatusConfirmed Status = "confirmed"
	StatusCancelled Status = "cancelled"
	StatusFulfilled Status = "fulfilled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusCancelled, StatusFulfilled:
		return true
	default:
		return false
	}
}

// Order is a placed checkout snapshot with server-computed totals.
type Order struct {
	ID                   uuid.UUID
	UserID               uuid.UUID
	Status               Status
	SubtotalCents        int64
	DiscountCents        int64
	TotalCents           int64
	DiscountCodeID       *uuid.UUID
	DiscountCodeSnapshot *string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// Item is an immutable order line captured at checkout.
type Item struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	LineType       domaincart.LineType
	ProductID      *uuid.UUID
	ComboID        *uuid.UUID
	Name           string
	Slug           string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}
