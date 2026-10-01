package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// OrderEmailUsecase sends the email an order notice calls for. It runs in the
// worker from the email queue, possibly well after the event, so it reads the
// order as it is now and never sends one that contradicts it.
type OrderEmailUsecase struct {
	orders   port.OrderRepository
	payments port.PaymentRepository
	composer port.OrderEmailComposer
	mailer   port.Mailer
	codes    domainorder.PickupCodes
	log      *zap.Logger
}

func NewOrderEmailUsecase(
	orders port.OrderRepository,
	payments port.PaymentRepository,
	composer port.OrderEmailComposer,
	mailer port.Mailer,
	codes domainorder.PickupCodes,
	log *zap.Logger,
) *OrderEmailUsecase {
	return &OrderEmailUsecase{orders: orders, payments: payments, composer: composer, mailer: mailer, codes: codes, log: log}
}

// Send emails task's notice to the order's customer. It skips the email when
// the order or its customer's account is gone — the order query joins open
// accounts only — or the order has moved past what the notice says. Errors
// wrapping port.ErrEmailUndeliverable are final; any other may pass on a retry.
func (u *OrderEmailUsecase) Send(ctx context.Context, task port.OrderEmailTask) error {
	order, err := u.orders.StaffGetByID(ctx, task.OrderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		u.skip(task, "order_or_account_gone")
		return nil
	}
	if err != nil {
		return err
	}
	if !task.Notice.StillApplies(order.Order.Status) {
		u.skip(task, "order_moved_on")
		return nil
	}
	// A guest's order, taken at the counter or on the phone, has no address.
	if order.CustomerEmail == nil {
		u.skip(task, "guest_order")
		return nil
	}
	// Checkout needs a confirmed address; this keeps any order placed before
	// that rule from mailing an address nobody confirmed.
	if !order.CustomerEmailVerified {
		u.skip(task, "email_not_verified")
		return nil
	}
	items, err := u.orders.ListItemsByOrderID(ctx, task.OrderID)
	if err != nil {
		return err
	}
	msg := port.OrderEmail{
		Notice: task.Notice,
		To:     *order.CustomerEmail,
		Order:  order.Order,
		Items:  items,
	}
	if order.CustomerDisplayName != nil {
		msg.CustomerName = *order.CustomerDisplayName
	}
	if order.Order.HasPickupCode() {
		msg.PickupCode = u.codes.Of(order.Order.ID)
	}
	if task.Notice == domainorder.NoticeCancelled {
		if msg.Reason, msg.RefundCents, err = u.cancellation(ctx, task.OrderID); err != nil {
			return err
		}
	}
	email, err := u.composer.ComposeOrderEmail(msg)
	if err != nil {
		return fmt.Errorf("%w: compose %s email: %w", port.ErrEmailUndeliverable, task.Notice, err)
	}
	if err := u.mailer.Send(ctx, email); err != nil {
		return err
	}
	u.log.Info("order_email_sent",
		zap.String("order_id", task.OrderID.String()),
		zap.String("notice", string(task.Notice)),
	)
	return nil
}

func (u *OrderEmailUsecase) skip(task port.OrderEmailTask, reason string) {
	u.log.Info("order_email_skipped",
		zap.String("order_id", task.OrderID.String()),
		zap.String("notice", string(task.Notice)),
		zap.String("reason", reason),
	)
}

// cancellation reads why the bakery cancelled the order, empty when the
// customer did, and what goes back to them: nothing for an order never paid.
func (u *OrderEmailUsecase) cancellation(ctx context.Context, orderID uuid.UUID) (string, int64, error) {
	events, err := u.orders.ListStatusEvents(ctx, orderID)
	if err != nil {
		return "", 0, err
	}
	reason := ""
	for _, event := range events {
		if event.To == domainorder.StatusCancelled {
			reason = event.Reason
		}
	}
	payment, err := u.payments.GetByOrderID(ctx, orderID)
	if errors.Is(err, apperrors.ErrNotFound) || (err == nil && payment.RefundRequestedAt == nil) {
		return reason, 0, nil
	}
	if err != nil {
		return "", 0, err
	}
	return reason, payment.AmountCents, nil
}
