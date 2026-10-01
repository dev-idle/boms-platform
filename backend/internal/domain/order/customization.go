package order

import (
	"encoding/json"
	"fmt"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainproduct "github.com/boms/backend/internal/domain/product"
)

// Customization is what a customer configured on an order item, stored as the
// item's configuration as it was priced at checkout: options by group and
// label, so a later catalog edit leaves the order as it was bought.
type Customization struct {
	Options           []ChosenOption `json:"options"`
	Message           string         `json:"message,omitempty"`
	ReferenceImageURL string         `json:"reference_image_url,omitempty"`
}

// ChosenOption is one option an item was bought with.
type ChosenOption struct {
	Group           domainproduct.OptionGroup `json:"group"`
	Label           string                    `json:"label"`
	PriceDeltaCents int64                     `json:"price_delta_cents"`
}

// ParseCustomization reads an order item's configuration: nil for an item its
// customer did not configure.
func ParseCustomization(configuration json.RawMessage) (*Customization, error) {
	if len(configuration) == 0 || string(configuration) == string(domaincart.EmptyConfiguration) {
		return nil, nil
	}
	var c Customization
	if err := json.Unmarshal(configuration, &c); err != nil {
		return nil, fmt.Errorf("parse order item customization: %w", err)
	}
	return &c, nil
}
