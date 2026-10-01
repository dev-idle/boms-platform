package usecase

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// noShowBatch is how many missed pickups one run records; the next run takes
// the rest.
const noShowBatch = 50

// Cancel cancels the customer's order while it is not being made yet. A paid
// order is refunded in full and its discount use goes back to the code: both
// follow from the move. Before an unpaid order is cancelled the provider is
// asked about its payment, as before one expires: money the shop did not hear
// of is recorded, and so refunded, and a payment the provider is reviewing
// holds the cancellation until it decides.
func (u *OrderUsecase) Cancel(ctx context.Context, userID, orderID uuid.UUID) (*dto.OrderResponse, error) {
	order, err := u.orders.GetByIDForUser(ctx, userID, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainorder.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if order.Status == domainorder.StatusAwaitingPayment {
		taken, err := u.payment.settle(ctx, orderID)
		if err != nil {
			return nil, err
		}
		if taken == domainpayment.StatusPending {
			return nil, domainpayment.ErrUnderReview
		}
	}
	customer := &port.OrderActor{ID: userID, Role: domainuser.RoleCustomer}
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		// The status the order has now: a payment just settled confirms it.
		locked, err := u.orders.LockForUpdate(txCtx, orderID)
		if err != nil {
			return err
		}
		if !locked.Status.BeforeProduction() {
			return domainorder.ErrInvalidStatusTransition
		}
		_, err = u.transitions.applyInTx(txCtx, customer, port.UpdateOrderStatusParams{
			OrderID:    orderID,
			FromStatus: locked.Status,
			ToStatus:   domainorder.StatusCancelled,
		}, "")
		return err
	})
	if err != nil {
		return nil, err
	}
	return u.orderResponse(ctx, userID, orderID)
}

// Reschedule moves the pickup of the customer's order, while it is not being
// made yet, to another time the checkout rules allow. The order's type follows
// the new time; the old slot gets its place back and the new one takes it.
func (u *OrderUsecase) Reschedule(
	ctx context.Context,
	userID, orderID uuid.UUID,
	req dto.RescheduleOrderRequest,
) (*dto.OrderResponse, error) {
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return nil, err
	}
	policy := pickupPolicy(settings, closed)
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		order, err := u.orders.LockForUpdate(txCtx, orderID)
		if errors.Is(err, apperrors.ErrNotFound) || (err == nil && order.UserID != userID) {
			return domainorder.ErrNotFound
		}
		if err != nil {
			return err
		}
		if !order.Status.BeforeProduction() {
			return domainorder.ErrInvalidStatusTransition
		}
		if order.PickupAt != nil && order.PickupAt.Equal(req.PickupAt) {
			return nil
		}
		lines, err := u.orders.ListItemsByOrderID(txCtx, order.ID)
		if err != nil {
			return err
		}
		items, err := u.cartUC.orderFulfillment(txCtx, lines)
		if err != nil {
			return err
		}
		booking, err := u.bookPickup(txCtx, userID, items, req.PickupAt, now, policy, order.PickupAt)
		if err != nil {
			return err
		}
		moved, err := u.orders.Reschedule(txCtx, port.RescheduleOrderParams{
			OrderID: order.ID, PickupAt: req.PickupAt, Type: booking.orderType,
		})
		if err != nil {
			return err
		}
		if err := u.events.Add(txCtx, domainorder.RescheduledEvent(*moved)); err != nil {
			return err
		}
		// Open checkouts hear of the old slot's freed place, as after a
		// cancellation, and of the new slot only when this order fills it.
		var days []time.Time
		if order.PickupAt != nil {
			days = append(days, domainstore.DayOf(*order.PickupAt))
		}
		if newDay := domainstore.DayOf(req.PickupAt); booking.fillsSlot && !slices.ContainsFunc(days, newDay.Equal) {
			days = append(days, newDay)
		}
		for _, day := range days {
			if err := u.events.Add(txCtx, domainorder.SlotsChangedEvent(day)); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return u.orderResponse(ctx, userID, orderID)
}

// MarkNoShows records the orders ready and not collected by closing time on
// their pickup day as missed pickups, a batch at a time, and returns how many.
// One collected or cancelled meanwhile stays as it is. Each order stands
// alone: one that fails is reported with the rest while the others go on.
func (u *OrderUsecase) MarkNoShows(ctx context.Context) (int, error) {
	settings, err := u.store.GetSettings(ctx)
	if err != nil {
		return 0, err
	}
	ids, err := u.orders.ListMissedPickups(ctx, settings.MissedBefore(time.Now()), noShowBatch)
	if err != nil {
		return 0, err
	}
	marked := 0
	var failures []error
	for _, id := range ids {
		_, err := u.transitions.apply(ctx, nil, port.UpdateOrderStatusParams{
			OrderID: id, FromStatus: domainorder.StatusReady, ToStatus: domainorder.StatusNoShow,
		}, "")
		if errors.Is(err, domainorder.ErrInvalidStatusTransition) {
			continue
		}
		if err != nil {
			failures = append(failures, apperrors.Errorf("mark order %s missed: %w", id, err))
			continue
		}
		marked++
	}
	return marked, errors.Join(failures...)
}
