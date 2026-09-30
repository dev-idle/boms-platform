package order

import (
	"encoding/json"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	"github.com/google/uuid"
)

// Status is the lifecycle state of a customer order.
type Status string

const (
	StatusPending      Status = "pending"
	StatusConfirmed    Status = "confirmed"
	StatusInProduction Status = "in_production"
	StatusReady        Status = "ready"
	StatusCancelled    Status = "cancelled"
	StatusFulfilled    Status = "fulfilled"
)

func (s Status) Valid() bool {
	switch s {
	case StatusPending, StatusConfirmed, StatusInProduction, StatusReady, StatusCancelled, StatusFulfilled:
		return true
	default:
		return false
	}
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
	// Terms is the policy version the customer accepted placing the order; orders
	// placed before acceptance was recorded have none.
	Terms     *domainpolicy.Acceptance
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Item is an immutable order line captured at checkout.
type Item struct {
	ID             uuid.UUID
	OrderID        uuid.UUID
	LineType       domaincart.LineType
	ProductID      *uuid.UUID
	ComboID        *uuid.UUID
	Configuration  json.RawMessage
	Name           string
	Slug           string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
	CreatedAt      time.Time
}
