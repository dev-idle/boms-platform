// Package payment is how an order is paid: online through PayPal, or in cash
// at the counter when staff took the order.
package payment

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Currency is what every price is charged in: PayPal takes no VND.
const Currency = "USD"

// Provider is who takes the payment.
type Provider string

const (
	ProviderPayPal Provider = "paypal"
	// ProviderCash is cash due at the counter, taken as the order is handed over.
	ProviderCash Provider = "cash"
)

// Status is where a payment stands with its provider.
type Status string

const (
	// StatusCreated: the buyer was sent to approve and has not paid yet.
	StatusCreated Status = "created"
	// StatusPending: the provider holds the capture for review.
	StatusPending Status = "pending"
	// StatusCaptured: the money is taken.
	StatusCaptured Status = "captured"
	// StatusDenied: the provider refused a pending capture; the buyer may pay again.
	StatusDenied Status = "denied"
	// StatusRefunded: the money taken went back to the buyer.
	StatusRefunded Status = "refunded"
)

// Payment is an order's one payment.
type Payment struct {
	ID       uuid.UUID
	OrderID  uuid.UUID
	Provider Provider
	// ProviderOrderID and ApproveURL are PayPal's order and the page the buyer
	// approves the payment on; empty for cash.
	ProviderOrderID string
	ApproveURL      string
	Status          Status
	CaptureID       *string
	AmountCents     int64
	Currency        string
	CapturedAt      *time.Time
	// RefundRequestedAt is when the order's cancellation asked for the money
	// back; RefundedAt when the provider took the refund.
	RefundRequestedAt *time.Time
	RefundedAt        *time.Time
	CreatedAt         time.Time
}

// Capture is the provider's answer to taking the money.
type Capture struct {
	ID          string
	Status      Status
	AmountCents int64
	Currency    string
}

var (
	// ErrNotPayable refuses to take payment for an order that does not await one.
	ErrNotPayable = errors.New("order is not awaiting payment")
	// ErrNotCompleted: the buyer did not approve, or the provider declined the
	// funding they chose; approving again with another one may work.
	ErrNotCompleted = errors.New("payment not completed")
	// ErrAmountMismatch: the provider took an amount other than the order's.
	ErrAmountMismatch = errors.New("captured amount does not match the order")
	// ErrWebhookInvalid refuses a webhook delivery the provider did not sign.
	ErrWebhookInvalid = errors.New("webhook signature invalid")
	// ErrAlreadyRefunded: the provider already returned the whole capture, as
	// when the bakery refunded it from the provider's own dashboard.
	ErrAlreadyRefunded = errors.New("capture already refunded")
	// ErrUnderReview: the provider holds the capture for review, so whether
	// the money is taken is not known yet.
	ErrUnderReview = errors.New("payment under review")
)
