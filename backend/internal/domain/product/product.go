package product

import (
	"time"

	"github.com/google/uuid"
)

// MaxLeadTime is the longest notice a product may need before pickup.
const MaxLeadTime = 7 * 24 * time.Hour

// Product is a sellable catalog item. LeadTime is the notice it needs before
// pickup, on top of the bakery's own.
type Product struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	Name        string
	Slug        string
	Description *string
	PriceCents  int64
	IsActive    bool
	LeadTime    time.Duration
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}
