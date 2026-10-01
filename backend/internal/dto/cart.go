package dto

type AddCartItemRequest struct {
	ProductID *string `json:"product_id,omitempty" validate:"omitempty,uuid"`
	ComboID   *string `json:"combo_id,omitempty" validate:"omitempty,uuid"`
	Quantity  int32   `json:"quantity" validate:"required,min=1,max=99"`
	// Customization configures a customizable product; every other line has none.
	Customization *CartCustomizationRequest `json:"customization,omitempty"`
}

// CartCustomizationRequest is how a customer configures a customizable
// product: one option from each group it offers, a message for the cake, and a
// photo they uploaded for reference, with their word that they may share it.
type CartCustomizationRequest struct {
	OptionIDs                []string `json:"option_ids" validate:"max=3,dive,uuid"`
	Message                  string   `json:"message" validate:"max=60"`
	ReferenceImageURL        string   `json:"reference_image_url" validate:"omitempty,url,max=2048"`
	ReferenceRightsConfirmed bool     `json:"reference_rights_confirmed"`
}

type UpdateCartItemRequest struct {
	Quantity int32 `json:"quantity" validate:"required,min=1,max=99"`
}

type ApplyCartDiscountRequest struct {
	Code string `json:"code" validate:"required,min=3,max=64"`
}

type CartItemResponse struct {
	ID             string  `json:"id"`
	LineType       string  `json:"line_type"`
	ProductID      *string `json:"product_id,omitempty"`
	ComboID        *string `json:"combo_id,omitempty"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Quantity       int32   `json:"quantity"`
	UnitPriceCents int64   `json:"unit_price_cents"`
	LineTotalCents int64   `json:"line_total_cents"`
	IsAvailable    bool    `json:"is_available"`
	// Customization is how the line is configured, null for a plain line.
	Customization *CustomizationResponse `json:"customization"`
}

type CartDiscountResponse struct {
	Code          string `json:"code"`
	DiscountType  string `json:"discount_type"`
	Value         int64  `json:"value"`
	DiscountCents int64  `json:"discount_cents"`
}

// FulfillmentResponse is what items ask of the bakery — whether the kitchen
// makes any of them, and the longest notice one needs — so the pickup picker
// can offer the times the server will accept.
type FulfillmentResponse struct {
	HasKitchenItems bool `json:"has_kitchen_items"`
	LeadMinutes     int  `json:"lead_minutes"`
}

type CartResponse struct {
	ID            string                `json:"id"`
	Items         []CartItemResponse    `json:"items"`
	SubtotalCents int64                 `json:"subtotal_cents"`
	Discount      *CartDiscountResponse `json:"discount,omitempty"`
	DiscountCents int64                 `json:"discount_cents"`
	TotalCents    int64                 `json:"total_cents"`
	CheckoutReady bool                  `json:"checkout_ready"`
	Fulfillment   FulfillmentResponse   `json:"fulfillment"`
}
