package dto

import "time"

// StaffOrderCustomerResponse is who to call when an order needs the customer: the
// counter hands the order over and phones a customer who is late or unreachable.
// A guest's order has no account: no id or email, and the name and phone staff
// wrote down.
type StaffOrderCustomerResponse struct {
	UserID      *string `json:"user_id"`
	Email       *string `json:"email"`
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
	Channel    string                     `json:"channel"`
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
	Channel              string                            `json:"channel"`
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
	// PickupCode hands an order placed online over: the customer gives it.
	PickupCode string `json:"pickup_code" validate:"omitempty,len=4,number"`
	// CashCollected hands an order staff took over: the counter took its cash.
	CashCollected bool `json:"cash_collected"`
}

// CreateStaffOrderRequest is an order staff take at the counter or on the
// phone, for a customer's account (CustomerID) or for a guest — one of them.
type CreateStaffOrderRequest struct {
	Channel    string                  `json:"channel" validate:"required,oneof=counter phone"`
	CustomerID *string                 `json:"customer_id" validate:"omitempty,uuid"`
	Guest      *StaffOrderGuestRequest `json:"guest"`
	PickupAt   time.Time               `json:"pickup_at" validate:"required"`
	Items      []StaffOrderItemRequest `json:"items" validate:"required,min=1,max=50,dive"`
}

// StaffOrderQuoteRequest is the items of an order staff are about to take.
type StaffOrderQuoteRequest struct {
	Items []StaffOrderItemRequest `json:"items" validate:"required,min=1,max=50,dive"`
}

// StaffOrderQuoteResponse is what those items cost now and ask of the bakery.
type StaffOrderQuoteResponse struct {
	TotalCents  int64               `json:"total_cents"`
	Fulfillment FulfillmentResponse `json:"fulfillment"`
}

// StaffOrderGuestRequest is who collects a guest's order.
type StaffOrderGuestRequest struct {
	Name  string `json:"name" validate:"required,max=100"`
	Phone string `json:"phone" validate:"required,max=20"`
}

// StaffOrderItemRequest is a product or a combo, with how many.
type StaffOrderItemRequest struct {
	ProductID *string `json:"product_id,omitempty" validate:"omitempty,uuid"`
	ComboID   *string `json:"combo_id,omitempty" validate:"omitempty,uuid"`
	Quantity  int32   `json:"quantity" validate:"required,min=1,max=99"`
}

// StaffProductResponse is a product the counter can mark sold out for the day.
type StaffProductResponse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	CategoryName string `json:"category_name"`
	Station      string `json:"station"`
	SoldOutToday bool   `json:"sold_out_today"`
}

// PatchStaffProductSoldOutRequest marks a product sold out for today, or back.
type PatchStaffProductSoldOutRequest struct {
	SoldOut *bool `json:"sold_out" validate:"required"`
}

// StaffCustomerLookupRequest is the email of the customer account an order is
// taken for; it rides in the body, never in a URL that logs keep.
type StaffCustomerLookupRequest struct {
	Email string `json:"email" validate:"required,email,max=254"`
}

// StaffCustomerResponse is a customer account staff take an order for.
type StaffCustomerResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	DisplayName *string `json:"display_name"`
	Phone       *string `json:"phone"`
}
