package product

import (
	"time"

	"github.com/google/uuid"
)

// Product is a sellable catalog item.
type Product struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	IsAvailable bool
	ImageURL    *string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
