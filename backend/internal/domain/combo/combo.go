package combo

import (
	"time"

	"github.com/google/uuid"
)

// Combo is a time-bounded bundle of products sold at a fixed price.
type Combo struct {
	ID         uuid.UUID
	Name       string
	Slug       string
	PriceCents int64
	StartsAt   time.Time
	EndsAt     time.Time
	IsActive   bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time
}

// Item is a line in a combo referencing a catalog product.
type Item struct {
	ID          uuid.UUID
	ComboID     uuid.UUID
	ProductID   uuid.UUID
	Quantity    int32
	ProductName string
	ProductSlug string
	PriceCents  int64
}
