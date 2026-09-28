package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// orderTransitions commits order status moves for the roles that make them.
type orderTransitions struct {
	tx     port.TxManager
	orders port.OrderRepository
	events port.EventOutbox
}

// apply moves an order to params.ToStatus and records the move in the order's
// history and the change notice in the same transaction, so both exist exactly
// when the move committed. A cancelled order frees its pickup slot, which open
// checkouts are told. A concurrent move that got there first leaves no row
// in FromStatus and is reported as an invalid transition.
func (t orderTransitions) apply(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	params port.UpdateOrderStatusParams,
) (*domainorder.Order, error) {
	var updated *domainorder.Order
	err := t.tx.WithTx(ctx, func(txCtx context.Context) error {
		order, err := t.orders.UpdateStatus(txCtx, params)
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return domainorder.ErrInvalidStatusTransition
			}
			return err
		}
		updated = order
		if err := t.orders.AddStatusEvent(txCtx, port.AddOrderStatusEventParams{
			OrderID:   order.ID,
			From:      &params.FromStatus,
			To:        order.Status,
			ActorID:   actorID,
			ActorRole: actorRole,
		}); err != nil {
			return err
		}
		if err := t.events.Add(txCtx, domainorder.StatusChangedEvent(params.FromStatus, *order)); err != nil {
			return err
		}
		if order.Status == domainorder.StatusCancelled && order.PickupAt != nil {
			return t.events.Add(txCtx, domainorder.SlotsChangedEvent(domainstore.DayOf(*order.PickupAt)))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
