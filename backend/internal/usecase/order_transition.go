package usecase

import (
	"context"
	"errors"

	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// orderTransitions commits order status moves for the roles that make them.
type orderTransitions struct {
	tx     port.TxManager
	orders port.OrderRepository
	events port.EventOutbox
}

// apply moves an order to params.ToStatus and records the change notice in the
// same transaction, so the notice exists exactly when the move committed. A
// concurrent move that got there first leaves no row in FromStatus and is
// reported as an invalid transition.
func (t orderTransitions) apply(ctx context.Context, params port.UpdateOrderStatusParams) (*domainorder.Order, error) {
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
		return t.events.Add(txCtx, domainorder.StatusChangedEvent(params.FromStatus, *order))
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}
