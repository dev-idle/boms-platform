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

// StaffPickupsResponse is a page of a bakery day's pickups, by time, with the
// slot length the counter measures lateness by: a pickup is late once its slot
// has ended.
type StaffPickupsResponse struct {
	SlotMinutes int                         `json:"slot_minutes"`
	Pickups     []StaffOrderSummaryResponse `json:"pickups"`
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
	Status string `json:"status"`
	// ActorRole is null when the system made the move.
	ActorRole *string `json:"actor_role"`
	// Reason is why the bakery cancelled the order; null for any other move.
	Reason *string   `json:"reason"`
	At     time.Time `json:"at"`
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
	Payment              *OrderPaymentResponse             `json:"payment"`
	CreatedAt            time.Time                         `json:"created_at"`
	UpdatedAt            time.Time                         `json:"updated_at"`
}

type PatchStaffOrderStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=confirmed fulfilled cancelled"`
	// Reason is required to cancel, and shown to the customer.
	Reason string `json:"reason" validate:"max=200"`
	// PickupCode is required to hand the order over: the customer gives it.
	PickupCode string `json:"pickup_code" validate:"omitempty,len=4,number"`
}
