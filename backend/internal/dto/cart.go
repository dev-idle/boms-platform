package dto

type AddCartItemRequest struct {
	ProductID *string `json:"product_id,omitempty" validate:"omitempty,uuid"`
	ComboID   *string `json:"combo_id,omitempty" validate:"omitempty,uuid"`
	Quantity  int32   `json:"quantity" validate:"required,min=1,max=99"`
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
}

type CartDiscountResponse struct {
	Code          string `json:"code"`
	DiscountType  string `json:"discount_type"`
	Value         int64  `json:"value"`
	DiscountCents int64  `json:"discount_cents"`
}

type CartResponse struct {
	ID             string                `json:"id"`
	Items          []CartItemResponse    `json:"items"`
	SubtotalCents  int64                 `json:"subtotal_cents"`
	Discount       *CartDiscountResponse `json:"discount,omitempty"`
	DiscountCents  int64                 `json:"discount_cents"`
	TotalCents     int64                 `json:"total_cents"`
	CheckoutReady  bool                  `json:"checkout_ready"`
}
