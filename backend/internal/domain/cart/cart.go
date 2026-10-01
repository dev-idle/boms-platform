package cart

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// EmptyConfiguration is the default JSON object for cart line configuration.
var EmptyConfiguration = json.RawMessage(`{}`)

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
	ID            uuid.UUID
	CartID        uuid.UUID
	LineType      LineType
	ProductID     *uuid.UUID
	ComboID       *uuid.UUID
	Quantity      int32
	Configuration json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// Customization is how a customer configured a customizable product on a cart
// line, stored as the line's configuration: the options chosen, a message for
// the cake, and a reference photo. The options are named by id, so the line is
// priced by what they cost now.
type Customization struct {
	OptionIDs         []uuid.UUID `json:"option_ids"`
	Message           string      `json:"message,omitempty"`
	ReferenceImageURL string      `json:"reference_image_url,omitempty"`
}

// ParseCustomization reads a line's configuration: nil for a plain line.
func ParseCustomization(configuration json.RawMessage) (*Customization, error) {
	if len(configuration) == 0 || string(configuration) == string(EmptyConfiguration) {
		return nil, nil
	}
	var c Customization
	if err := json.Unmarshal(configuration, &c); err != nil {
		return nil, fmt.Errorf("parse cart line customization: %w", err)
	}
	return &c, nil
}
