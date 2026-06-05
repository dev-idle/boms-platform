package discount

import (
	"time"

	"github.com/google/uuid"
)

// Type is how a discount code reduces order total.
type Type string

const (
	TypePercent    Type = "percent"
	TypeFixedCents Type = "fixed_cents"
)

func (t Type) Valid() bool {
	return t == TypePercent || t == TypeFixedCents
}

// Code is a manager-defined promotion redeemable at checkout.
type Code struct {
	ID             uuid.UUID
	Code           string
	DiscountType   Type
	Value          int64
	MinOrderCents  *int64
	MaxUses        *int32
	UsedCount      int32
	StartsAt       time.Time
	EndsAt         time.Time
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
	DeletedAt      *time.Time
}
