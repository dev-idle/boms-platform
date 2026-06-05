package cart

import (
	"time"

	"github.com/google/uuid"
)

// LineType identifies whether a cart line references a product or combo.
type LineType string

const (
	LineTypeProduct LineType = "product"
	LineTypeCombo   LineType = "combo"
)

func (t LineType) Valid() bool {
	return t == LineTypeProduct || t == LineTypeCombo
}

// Cart is a per-customer basket persisted on the server.
type Cart struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	DiscountCodeID *uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// Item is a single cart line (product or combo reference + quantity).
type Item struct {
	ID        uuid.UUID
	CartID    uuid.UUID
	LineType  LineType
	ProductID *uuid.UUID
	ComboID   *uuid.UUID
	Quantity  int32
	CreatedAt time.Time
	UpdatedAt time.Time
}
