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

// OrderTimelineEntryResponse is a status the order entered and when.
type OrderTimelineEntryResponse struct {
	Status string    `json:"status"`
	At     time.Time `json:"at"`
}

type OrderResponse struct {
	ID                   string                       `json:"id"`
	Code                 string                       `json:"code"`
	Status               string                       `json:"status"`
	OrderType            string                       `json:"order_type"`
	SubtotalCents        int64                        `json:"subtotal_cents"`
	DiscountCents        int64                        `json:"discount_cents"`
	TotalCents           int64                        `json:"total_cents"`
	DiscountCodeSnapshot *string                      `json:"discount_code_snapshot,omitempty"`
	PickupAt             *time.Time                   `json:"pickup_at,omitempty"`
	Items                []OrderItemResponse          `json:"items"`
	Timeline             []OrderTimelineEntryResponse `json:"timeline"`
	Tickets              []TicketSummaryResponse      `json:"tickets"`
	Payment              *OrderPaymentResponse        `json:"payment"`
	// PaymentDueAt is when the order expires if it is still unpaid.
	PaymentDueAt *time.Time `json:"payment_due_at"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// OrderPaymentResponse is where an order's payment stands; null before the
// customer started paying.
type OrderPaymentResponse struct {
	Provider   string     `json:"provider"`
	Status     string     `json:"status"`
	CapturedAt *time.Time `json:"captured_at"`
}

// PaymentStartResponse is the provider page the customer approves the payment on.
type PaymentStartResponse struct {
	ApproveURL string `json:"approve_url"`
}

// PaymentCaptureResponse is where the payment stands after the capture:
// captured, or pending while the provider reviews it.
type PaymentCaptureResponse struct {
	Status string `json:"status"`
}

type OrderSummaryResponse struct {
	ID         string     `json:"id"`
	Code       string     `json:"code"`
	Status     string     `json:"status"`
	TotalCents int64      `json:"total_cents"`
	ItemCount  int32      `json:"item_count"`
	PickupAt   *time.Time `json:"pickup_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// OrderHistoryQuery is a customer's order-history filter as sent: a status and
// a range of bakery days (YYYY-MM-DD), each optional.
type OrderHistoryQuery struct {
	Status string
	From   string
	To     string
}

type CheckoutRequest struct {
	PickupAt time.Time `json:"pickup_at" validate:"required"`
	// TermsVersion is the version of the policies the customer accepted for this
	// order; the usecase refuses any but the current one.
	TermsVersion string `json:"terms_version" validate:"max=32"`
}
