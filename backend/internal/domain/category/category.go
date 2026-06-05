package category

import (
	"time"

	"github.com/google/uuid"
)

// Category is a catalog grouping for products.
type Category struct {
	ID        uuid.UUID
	Name      string
	Slug      string
	SortOrder int32
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}
