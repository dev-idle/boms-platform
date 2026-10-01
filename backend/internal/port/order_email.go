package port

import (
	"context"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
)

// OrderEmailTask asks for one order notice to be emailed. EventID is the
// outbox event that called for it: the queue keeps one task per event, so an
// event delivered twice still queues one email.
type OrderEmailTask struct {
	EventID uuid.UUID
	OrderID uuid.UUID
	Notice  domainorder.Notice
}

// OrderEmailQueue holds order emails until a worker sends them. Enqueueing a
// task for an event that already has one succeeds without adding another.
type OrderEmailQueue interface {
	EnqueueOrderEmail(ctx context.Context, task OrderEmailTask) error
}

// OrderEmail is what an order notice tells its customer.
type OrderEmail struct {
	Notice       domainorder.Notice
	To           string
	CustomerName string
	Order        domainorder.Order
	Items        []domainorder.Item
	// Reason is why the bakery cancelled the order, empty when the customer
	// did; RefundCents what goes back to them. Both only for a cancellation.
	Reason      string
	RefundCents int64
	// PickupCode is what the customer gives at the counter, on the notices
	// sent once the bakery accepted the order.
	PickupCode string
}

// OrderEmailComposer writes the email for an order notice.
type OrderEmailComposer interface {
	ComposeOrderEmail(msg OrderEmail) (Email, error)
}
