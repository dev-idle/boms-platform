package dto

import "time"

// StaffOrderCustomerResponse is who to call when an order needs the customer: the
// counter hands the order over and phones a customer who is late or unreachable.
type StaffOrderCustomerResponse struct {
	UserID      string  `json:"user_id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name,omitempty"`
	Phone       *string `json:"phone,omitempty"`
}

type StaffOrderSummaryResponse struct {
	ID         string                     `json:"id"`
	Code       string                     `json:"code"`
	Status     string                     `json:"status"`
	TotalCents int64                      `json:"total_cents"`
	ItemCount  int32                      `json:"item_count"`
	Customer   StaffOrderCustomerResponse `json:"customer"`
	PickupAt   *time.Time                 `json:"pickup_at,omitempty"`
	CreatedAt  time.Time                  `json:"created_at"`
}

// StaffOrderTimelineEntryResponse is a status the order entered, when, and the
// role of whoever moved it there — enough to answer a customer at the counter.
type StaffOrderTimelineEntryResponse struct {
	Status    string    `json:"status"`
	ActorRole string    `json:"actor_role"`
	At        time.Time `json:"at"`
}

type StaffOrderResponse struct {
	ID                   string                            `json:"id"`
	Code                 string                            `json:"code"`
	Status               string                            `json:"status"`
	OrderType            string                            `json:"order_type"`
	SubtotalCents        int64                             `json:"subtotal_cents"`
	DiscountCents        int64                             `json:"discount_cents"`
	TotalCents           int64                             `json:"total_cents"`
	DiscountCodeSnapshot *string                           `json:"discount_code_snapshot,omitempty"`
	PickupAt             *time.Time                        `json:"pickup_at,omitempty"`
	Items                []OrderItemResponse               `json:"items"`
	Timeline             []StaffOrderTimelineEntryResponse `json:"timeline"`
	Tickets              []OrderTicketResponse             `json:"tickets"`
	Customer             StaffOrderCustomerResponse        `json:"customer"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}

type PatchStaffOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed fulfilled cancelled"`
}
