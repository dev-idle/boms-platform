package dto

import "time"

// BakerOrderCustomerResponse names the order for the kitchen and nothing more:
// the baker never contacts the customer, so no email or phone reaches them.
type BakerOrderCustomerResponse struct {
	DisplayName *string `json:"display_name,omitempty"`
}

type BakerOrderSummaryResponse struct {
	ID         string                     `json:"id"`
	Code       string                     `json:"code"`
	Status     string                     `json:"status"`
	TotalCents int64                      `json:"total_cents"`
	ItemCount  int32                      `json:"item_count"`
	Customer   BakerOrderCustomerResponse `json:"customer"`
	PickupAt   *time.Time                 `json:"pickup_at,omitempty"`
	CreatedAt  time.Time                  `json:"created_at"`
}

type BakerOrderResponse struct {
	ID                   string                     `json:"id"`
	Code                 string                     `json:"code"`
	Status               string                     `json:"status"`
	SubtotalCents        int64                      `json:"subtotal_cents"`
	DiscountCents        int64                      `json:"discount_cents"`
	TotalCents           int64                      `json:"total_cents"`
	DiscountCodeSnapshot *string                    `json:"discount_code_snapshot,omitempty"`
	PickupAt             *time.Time                 `json:"pickup_at,omitempty"`
	Items                []OrderItemResponse        `json:"items"`
	Customer             BakerOrderCustomerResponse `json:"customer"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
}

type PatchBakerOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=in_production ready"`
}
