package usecase

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// OrderEmailUsecase sends the email an order notice calls for. It runs in the
// worker from the email queue, possibly well after the event, so it reads the
// order as it is now and never sends one that contradicts it.
type OrderEmailUsecase struct {
	orders   port.OrderRepository
	composer port.OrderEmailComposer
	mailer   port.Mailer
	log      *zap.Logger
}

func NewOrderEmailUsecase(
	orders port.OrderRepository,
	composer port.OrderEmailComposer,
	mailer port.Mailer,
	log *zap.Logger,
) *OrderEmailUsecase {
	return &OrderEmailUsecase{orders: orders, composer: composer, mailer: mailer, log: log}
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
	items, err := u.orders.ListItemsByOrderID(ctx, task.OrderID)
	if err != nil {
		return err
	}
	msg := port.OrderEmail{
		Notice: task.Notice,
		To:     order.CustomerEmail,
		Order:  order.Order,
		Items:  items,
	}
	if order.CustomerDisplayName != nil {
		msg.CustomerName = *order.CustomerDisplayName
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
