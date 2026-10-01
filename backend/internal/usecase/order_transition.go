package usecase

import (
	"context"
	"errors"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// orderTransitions commits order status moves for the roles that make them.
type orderTransitions struct {
	tx        port.TxManager
	orders    port.OrderRepository
	tickets   port.TicketRepository
	discounts port.DiscountCodeRepository
	payments  port.PaymentRepository
	events    port.EventOutbox
}

// apply moves an order to params.ToStatus and records the move in the order's
// history and the change notice in the same transaction, so both exist exactly
// when the move committed. An order that will not be made, cancelled or
// expired, cancels its tickets and frees its pickup slot; the stations and open
// checkouts are told. reason is why the bakery cancels, empty for any other
// move. A concurrent move that got there first leaves no row in FromStatus and
// is reported as an invalid transition.
func (t orderTransitions) apply(
	ctx context.Context,
	actor *port.OrderActor,
	params port.UpdateOrderStatusParams,
	reason string,
) (*domainorder.Order, error) {
	var updated *domainorder.Order
	err := t.tx.WithTx(ctx, func(txCtx context.Context) (err error) {
		updated, err = t.applyInTx(txCtx, actor, params, reason)
		return err
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// applyInTx is apply inside a transaction the caller runs, for a move that
// commits with other writes.
func (t orderTransitions) applyInTx(
	txCtx context.Context,
	actor *port.OrderActor,
	params port.UpdateOrderStatusParams,
	reason string,
) (*domainorder.Order, error) {
	order, err := t.orders.UpdateStatus(txCtx, params)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrInvalidStatusTransition
		}
		return nil, err
	}
	return order, t.recordInTx(txCtx, actor, params.FromStatus, *order, reason)
}

// recordInTx follows an order's move from one status to its current one, in the
// transaction that made it: its history, the change notice and, for an order
// that will not be made, its tickets, pickup slot, discount use and payment.
// actor is nil for the system.
func (t orderTransitions) recordInTx(
	txCtx context.Context,
	actor *port.OrderActor,
	from domainorder.Status,
	order domainorder.Order,
	reason string,
) error {
	if err := t.orders.AddStatusEvent(txCtx, port.AddOrderStatusEventParams{
		OrderID: order.ID,
		From:    &from,
		To:      order.Status,
		Actor:   actor,
		Reason:  reason,
	}); err != nil {
		return err
	}
	if err := t.events.Add(txCtx, domainorder.StatusChangedEvent(from, order)); err != nil {
		return err
	}
	if !order.Status.Dropped() {
		return nil
	}
	if err := t.cancelTickets(txCtx, order, from); err != nil {
		return err
	}
	// An order dropped before it was made gives its discount use back.
	if from.BeforeProduction() && order.DiscountCodeID != nil {
		if err := t.discounts.ReleaseUse(txCtx, *order.DiscountCodeID); err != nil {
			return err
		}
	}
	// A cancelled order is refunded in full; the worker makes the refund, so a
	// provider that is down never undoes the cancellation.
	if order.Status == domainorder.StatusCancelled {
		if err := t.payments.RequestRefund(txCtx, order.ID); err != nil {
			return err
		}
	}
	if order.PickupAt != nil {
		return t.events.Add(txCtx, domainorder.SlotsChangedEvent(domainstore.DayOf(*order.PickupAt)))
	}
	return nil
}

// cancelTickets cancels a dropped order's tickets inside the transaction that
// drops it, and tells each station that saw the order in its last status.
func (t orderTransitions) cancelTickets(txCtx context.Context, order domainorder.Order, from domainorder.Status) error {
	cancelled, err := t.tickets.CancelForOrder(txCtx, order.ID)
	if err != nil {
		return err
	}
	for _, ticket := range cancelled {
		if err := t.events.Add(txCtx, domainorder.TicketChangedEvent(order, from, ticket)); err != nil {
			return err
		}
	}
	return nil
}
