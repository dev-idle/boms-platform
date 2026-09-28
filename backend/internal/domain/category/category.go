package category

import (
	"time"

	"github.com/google/uuid"
)

// Station is where a category's products are made: the kitchen bakes to
// order, the counter sells what is ready.
type Station string

const (
	StationKitchen Station = "kitchen"
	StationCounter Station = "counter"
)

// Valid reports whether s is a known station.
func (s Station) Valid() bool {
	return s == StationKitchen || s == StationCounter
}

// Category is a catalog grouping for products.
type Category struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	SortOrder int32
	IsActive  bool
	Station   Station
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
