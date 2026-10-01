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
	// Customization is how the customer configured the item, null for a plain one.
	Customization *CustomizationResponse `json:"customization"`
}

// CustomizationResponse is how a customer configured a custom item: the
// options chosen with what each added, a message for the cake and a reference
// photo (empty when none).
type CustomizationResponse struct {
	Options           []ChosenOptionResponse `json:"options"`
	Message           string                 `json:"message"`
	ReferenceImageURL string                 `json:"reference_image_url"`
}

// ChosenOptionResponse is one option a custom item has.
type ChosenOptionResponse struct {
	Group           string `json:"group"`
	Label           string `json:"label"`
	PriceDeltaCents int64  `json:"price_delta_cents"`
}

// OrderTimelineEntryResponse is a status the order entered and when.
type OrderTimelineEntryResponse struct {
	Status string `json:"status"`
	// Reason is why the bakery cancelled the order; null for any other move.
	Reason *string   `json:"reason"`
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
	// Fulfillment is what the items ask of the bakery, for choosing another
	// pickup; null once the order is being made.
	Fulfillment *FulfillmentResponse `json:"fulfillment"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
}

// OrderPaymentResponse is where an order's payment stands; null before the
// customer started paying.
type OrderPaymentResponse struct {
	Provider   string     `json:"provider"`
	Status     string     `json:"status"`
	CapturedAt *time.Time `json:"captured_at"`
	// RefundRequestedAt is set once the order is cancelled; RefundedAt once the
	// money went back.
	RefundRequestedAt *time.Time `json:"refund_requested_at"`
	RefundedAt        *time.Time `json:"refunded_at"`
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

// RescheduleOrderRequest moves an order's pickup to another slot.
type RescheduleOrderRequest struct {
	PickupAt time.Time `json:"pickup_at" validate:"required"`
}

type CheckoutRequest struct {
	PickupAt time.Time `json:"pickup_at" validate:"required"`
	// TermsVersion is the version of the policies the customer accepted for this
	// order; the usecase refuses any but the current one.
	TermsVersion string `json:"terms_version" validate:"max=32"`
}
