package dto

import "time"

type StaffOrderCustomerResponse struct {
	UserID      string  `json:"user_id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name,omitempty"`
}

type StaffOrderSummaryResponse struct {
	ID            string                     `json:"id"`
	Status        string                     `json:"status"`
	TotalCents    int64                      `json:"total_cents"`
	ItemCount     int32                      `json:"item_count"`
	Customer      StaffOrderCustomerResponse `json:"customer"`
	CreatedAt     time.Time                  `json:"created_at"`
}

type StaffOrderResponse struct {
	ID                   string                     `json:"id"`
	Status               string                     `json:"status"`
	SubtotalCents        int64                      `json:"subtotal_cents"`
	DiscountCents        int64                      `json:"discount_cents"`
	TotalCents           int64                      `json:"total_cents"`
	DiscountCodeSnapshot *string                    `json:"discount_code_snapshot,omitempty"`
	Items                []OrderItemResponse        `json:"items"`
	Customer             StaffOrderCustomerResponse `json:"customer"`
	CreatedAt            time.Time                  `json:"created_at"`
	UpdatedAt            time.Time                  `json:"updated_at"`
}

type PatchStaffOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed cancelled fulfilled"`
}
