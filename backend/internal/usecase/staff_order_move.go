package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// MovePickup moves an order the bakery still has to hand over to another
// pickup, at a customer's call or the bakery's: any time ahead within the
// bakery's days, hours and free slots, the notice its items need being the
// bakery's own call. The order keeps its type; the old slot gets its place
// back and the new one takes it.
func (u *StaffOrderUsecase) MovePickup(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	orderID uuid.UUID,
	req dto.RescheduleOrderRequest,
) (*dto.StaffOrderResponse, error) {
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return nil, err
	}
	policy := pickupPolicy(settings, closed)
	var before *time.Time
	moved := false
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		order, err := u.orders.LockForUpdate(txCtx, orderID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return domainorder.ErrNotFound
		}
		if err != nil {
			return err
		}
		if !order.Status.MovableByStaff() {
			return domainorder.ErrInvalidStatusTransition
		}
		before = order.PickupAt
		if order.PickupAt != nil && order.PickupAt.Equal(req.PickupAt) {
			return nil
		}
		if err := policy.Moves(req.PickupAt, now); err != nil {
			return err
		}
		fillsSlot, err := holdSlot(txCtx, u.orders, u.store, req.PickupAt, settings)
		if err != nil {
			return err
		}
		rescheduled, err := u.orders.StaffReschedule(txCtx, order.ID, req.PickupAt)
		if err != nil {
			return err
		}
		if err := u.events.Add(txCtx, domainorder.RescheduledEvent(*rescheduled)); err != nil {
			return err
		}
		moved = true
		return addSlotChanges(txCtx, u.events, order.PickupAt, req.PickupAt, fillsSlot)
	})
	if err != nil {
		return nil, err
	}
	if moved {
		recordAudit(u.log, u.audit, ctx, domainorder.AuditActionStaffMovedPickup, actorID, actorRole, &orderID, "order",
			map[string]any{"pickup_at": before}, map[string]any{"pickup_at": req.PickupAt})
	}
	return u.Get(ctx, orderID)
}
