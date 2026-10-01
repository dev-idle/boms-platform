package port

import (
	"context"

	"github.com/google/uuid"

	domainpayment "github.com/boms/backend/internal/domain/payment"
)

type CreatePaymentParams struct {
	OrderID         uuid.UUID
	Provider        domainpayment.Provider
	ProviderOrderID string
	ApproveURL      string
	AmountCents     int64
	Currency        string
}

// PaymentRepository keeps orders' payments; there is at most one per order.
type PaymentRepository interface {
	// Create records an order's payment. An order that already has one is
	// apperrors.ErrConflict.
	Create(ctx context.Context, params CreatePaymentParams) (*domainpayment.Payment, error)
	GetByOrderID(ctx context.Context, orderID uuid.UUID) (*domainpayment.Payment, error)
	GetByProviderOrderID(ctx context.Context, provider domainpayment.Provider, providerOrderID string) (*domainpayment.Payment, error)
	// Restart gives a denied payment a new provider order to approve. A payment
	// not denied is apperrors.ErrNotFound.
	Restart(ctx context.Context, orderID uuid.UUID, providerOrderID, approveURL string) (*domainpayment.Payment, error)
	// RecordCapture records the provider's answer to a capture on a payment
	// created or pending. Any other is apperrors.ErrNotFound: the answer was
	// recorded already, so a repeated answer changes nothing.
	RecordCapture(ctx context.Context, paymentID uuid.UUID, capture domainpayment.Capture) (*domainpayment.Payment, error)
	// RequestRefund asks for the money back on the order's payment, in the
	// transaction that cancels the order; a payment not taken, or asked for
	// already, stays as it is.
	RequestRefund(ctx context.Context, orderID uuid.UUID) error
	// ListRefundsDue returns up to limit payments whose refund is asked for and
	// not made yet, the oldest request first.
	ListRefundsDue(ctx context.Context, limit int32) ([]domainpayment.Payment, error)
	// RecordRefund records the provider's refund of a payment whose refund is
	// due; refundID is nil when the provider had refunded it already. Any other
	// is apperrors.ErrNotFound: the refund was recorded already.
	RecordRefund(ctx context.Context, paymentID uuid.UUID, refundID *string) (*domainpayment.Payment, error)
	// CreateCash records the cash due for an order staff took.
	CreateCash(ctx context.Context, orderID uuid.UUID, amountCents int64) error
	// CollectCash records that the counter took the cash due on the order.
	// An order with no cash due is apperrors.ErrNotFound.
	CollectCash(ctx context.Context, orderID uuid.UUID) error
}

// PaymentOrderRequest asks a provider to take an order's total; the buyer
// comes back to ReturnURL after approving and to CancelURL after giving up.
type PaymentOrderRequest struct {
	OrderID     uuid.UUID
	AmountCents int64
	Currency    string
	ReturnURL   string
	CancelURL   string
}

// PaymentWebhookEvent is a provider's signed notice. Capture is set only for
// the outcome of a capture: taken or refused after review.
type PaymentWebhookEvent struct {
	ID              string
	ProviderOrderID string
	Capture         *domainpayment.Capture
}

// PaymentGateway takes payments through the provider (adapter/paypal).
type PaymentGateway interface {
	// CreateOrder asks the provider to take the amount and returns its order
	// id and the page the buyer approves it on.
	CreateOrder(ctx context.Context, req PaymentOrderRequest) (providerOrderID, approveURL string, err error)
	// Capture takes the money the buyer approved. Asking again for the same
	// provider order returns the first answer. An order the buyer did not
	// approve, or whose funding was declined, is domainpayment.ErrNotCompleted.
	Capture(ctx context.Context, providerOrderID string) (domainpayment.Capture, error)
	// Lookup reads what the provider captured on its order: nil when nothing
	// was or the provider no longer has the order; a refused capture comes
	// back denied.
	Lookup(ctx context.Context, providerOrderID string) (*domainpayment.Capture, error)
	// VerifyWebhook checks that the provider signed a webhook delivery and
	// reads it; header reads the delivery's headers by name. An unsigned one
	// is domainpayment.ErrWebhookInvalid.
	VerifyWebhook(ctx context.Context, header func(name string) string, body []byte) (PaymentWebhookEvent, error)
	// Refund returns the whole of a capture to the buyer and the provider's id
	// for the refund. Asking again for the same capture returns the first
	// answer; a capture refunded already, as from the provider's own
	// dashboard, is domainpayment.ErrAlreadyRefunded.
	Refund(ctx context.Context, captureID string) (refundID string, err error)
}
