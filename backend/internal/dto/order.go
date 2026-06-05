package dto

import "time"

type OrderItemResponse struct {
	ID             string  `json:"id"`
	LineType       string  `json:"line_type"`
	ProductID      *string `json:"product_id,omitempty"`
	ComboID        *string `json:"combo_id,omitempty"`
	Name           string  `json:"name"`
	Slug           string  `json:"slug"`
	Quantity       int32   `json:"quantity"`
	UnitPriceCents int64   `json:"unit_price_cents"`
	LineTotalCents int64   `json:"line_total_cents"`
}

type OrderResponse struct {
	ID                   string              `json:"id"`
	Status               string              `json:"status"`
	SubtotalCents        int64               `json:"subtotal_cents"`
	DiscountCents        int64               `json:"discount_cents"`
	TotalCents           int64               `json:"total_cents"`
	DiscountCodeSnapshot *string             `json:"discount_code_snapshot,omitempty"`
	Items                []OrderItemResponse `json:"items"`
	CreatedAt            time.Time           `json:"created_at"`
	UpdatedAt            time.Time           `json:"updated_at"`
}

type OrderSummaryResponse struct {
	ID            string    `json:"id"`
	Status        string    `json:"status"`
	TotalCents    int64     `json:"total_cents"`
	ItemCount     int32     `json:"item_count"`
	CreatedAt     time.Time `json:"created_at"`
}
