package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// A customer cancels or moves an order until the bakery starts making it; the
// bakery cancels with a reason the customer reads; a paid cancellation is
// refunded; an order not collected by closing time is recorded as missed.
func TestOrderChanges_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	setStatus := func(orderID string, status domainorder.Status, pickupIn string) {
		_, err := raw.Exec(ctx, "UPDATE orders SET status = $2, pickup_at = now() + $3::interval WHERE id = $1",
			orderID, string(status), pickupIn)
		require.NoError(t, err)
	}
	f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 1 })

	t.Run("a_paid_order_its_customer_cancels_is_refunded_and_frees_its_slot_and_code", func(t *testing.T) {
		one := int32(1)
		_, err := postgresadapter.NewDiscountCodeRepository(f.pool).Create(ctx, port.CreateDiscountCodeParams{
			Code: "ONEUSE", DiscountType: domaindiscount.TypePercent, Value: 10, MaxUses: &one, IsActive: true,
			StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
		withCode := func() uuid.UUID {
			customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
			_, err := f.cartUC.ApplyDiscount(ctx, customer, dto.ApplyCartDiscountRequest{Code: "ONEUSE"})
			require.NoError(t, err)
			return customer
		}
		slot := tomorrowAt(9, 0)
		customer := withCode()
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(slot))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)

		cancelled, err := f.orderUC.Cancel(ctx, customer, uuid.MustParse(order.ID))

		require.NoError(t, err)
		assert.Equal(t, "cancelled", cancelled.Status)
		require.NotNil(t, cancelled.Payment)
		assert.NotNil(t, cancelled.Payment.RefundRequestedAt)
		assert.Nil(t, cancelled.Fulfillment, "a closed order has no pickup to move")
		_, err = f.orderUC.Checkout(ctx, withCode(), uuid.New(), acceptingTerms(slot))
		require.NoError(t, err, "the slot and the code's one use are free again")

		refunded, err := f.paymentUC.RefundDue(ctx)
		require.NoError(t, err)
		assert.Equal(t, 1, refunded)
		after, err := f.orderUC.Get(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, string(domainpayment.StatusRefunded), after.Payment.Status)
		assert.NotNil(t, after.Payment.RefundedAt)
		again, err := f.paymentUC.RefundDue(ctx)
		require.NoError(t, err)
		assert.Zero(t, again, "a refund is made once")
	})

	t.Run("an_unpaid_order_its_customer_cancels_stays_theirs_alone", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(10, 0)))
		require.NoError(t, err)

		cancelled, err := f.orderUC.Cancel(ctx, customer, uuid.MustParse(order.ID))

		require.NoError(t, err)
		assert.Equal(t, "cancelled", cancelled.Status)
		_, err = f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.ErrorIs(t, err, domainorder.ErrNotFound, "never paid, never the bakery's")
		list, _, _, _, err := f.staff.List(ctx, 1, 100, "cancelled")
		require.NoError(t, err)
		for _, row := range list {
			assert.NotEqual(t, order.ID, row.ID)
		}
	})

	t.Run("money_paypal_took_unheard_is_refunded_when_its_order_is_cancelled", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(10, 30)))
		require.NoError(t, err)
		_, err = f.paymentUC.Start(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		f.paypal.mu.Lock()
		f.paypal.taken["PAYPAL-"+order.ID] = domainpayment.StatusCaptured
		f.paypal.mu.Unlock()

		cancelled, err := f.orderUC.Cancel(ctx, customer, uuid.MustParse(order.ID))

		require.NoError(t, err)
		assert.Equal(t, "cancelled", cancelled.Status)
		require.NotNil(t, cancelled.Payment)
		assert.NotNil(t, cancelled.Payment.RefundRequestedAt, "the money goes back")
		_, err = f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err, "paid, so the bakery sees it")
	})

	t.Run("a_payment_under_review_holds_the_cancellation", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(13, 0)))
		require.NoError(t, err)
		_, err = f.paymentUC.Start(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		f.paypal.mu.Lock()
		f.paypal.taken["PAYPAL-"+order.ID] = domainpayment.StatusPending
		f.paypal.mu.Unlock()

		_, err = f.orderUC.Cancel(ctx, customer, uuid.MustParse(order.ID))

		require.ErrorIs(t, err, domainpayment.ErrUnderReview)
		still, err := f.orderUC.Get(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, "awaiting_payment", still.Status)
	})

	t.Run("an_order_being_made_stays_as_it_is", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(9, 30)))
		require.NoError(t, err)
		setStatus(order.ID, domainorder.StatusInProduction, "1 day")

		_, err = f.orderUC.Cancel(ctx, customer, uuid.MustParse(order.ID))
		require.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition)
		_, err = f.orderUC.Reschedule(ctx, customer, uuid.MustParse(order.ID), dto.RescheduleOrderRequest{PickupAt: tomorrowAt(14, 0)})
		require.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition)
		stranger := f.newCustomer(t, nil, nil)
		_, err = f.orderUC.Cancel(ctx, stranger, uuid.MustParse(order.ID))
		require.ErrorIs(t, err, domainorder.ErrNotFound, "another customer's order")
		_, err = f.orderUC.Reschedule(ctx, stranger, uuid.MustParse(order.ID), dto.RescheduleOrderRequest{PickupAt: tomorrowAt(14, 0)})
		require.ErrorIs(t, err, domainorder.ErrNotFound, "another customer's order")
	})

	t.Run("the_bakery_cancels_with_a_reason_its_customer_reads", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 10)

		_, err := f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "cancelled", Reason: "  "})
		require.ErrorIs(t, err, domainorder.ErrInvalidCancelReason)
		_, err = f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "fulfilled", Reason: "Out of matcha today"})
		var appErr *apperrors.AppError
		require.ErrorAs(t, err, &appErr, "only a cancellation takes a reason")
		assert.Equal(t, apperrors.ErrValidation.Code, appErr.Code)
		assert.Contains(t, appErr.Details, "reason")
		_, err = f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID),
			dto.PatchStaffOrderStatusRequest{Status: "cancelled", Reason: "Out of matcha today"})
		require.NoError(t, err)

		staffView, err := f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		last := staffView.Timeline[len(staffView.Timeline)-1]
		require.NotNil(t, last.Reason)
		assert.Equal(t, "Out of matcha today", *last.Reason)
		assert.NotNil(t, staffView.Payment.RefundRequestedAt, "a paid order is refunded in full")
	})

	t.Run("a_customer_moves_the_pickup_within_the_checkout_rules", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(11, 0)))
		require.NoError(t, err)
		require.NotNil(t, order.Fulfillment, "the picker learns what the items need")
		taken, err := f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.pastry}, nil), uuid.New(), acceptingTerms(tomorrowAt(12, 0)))
		require.NoError(t, err)
		id := uuid.MustParse(order.ID)

		_, err = f.orderUC.Reschedule(ctx, customer, id, dto.RescheduleOrderRequest{PickupAt: *taken.PickupAt})
		require.ErrorIs(t, err, domainorder.ErrPickupSlotFull)
		_, err = f.orderUC.Reschedule(ctx, customer, id, dto.RescheduleOrderRequest{PickupAt: tomorrowAt(11, 10)})
		require.ErrorIs(t, err, domainorder.ErrPickupOffSlot)
		moved, err := f.orderUC.Reschedule(ctx, customer, id, dto.RescheduleOrderRequest{PickupAt: tomorrowAt(12, 30)})

		require.NoError(t, err)
		assert.True(t, moved.PickupAt.Equal(tomorrowAt(12, 30)))
		booked, err := f.orders.CountByPickupTime(ctx, tomorrowAt(11, 0), tomorrowAt(11, 1))
		require.NoError(t, err)
		assert.Empty(t, booked, "the old slot is free again")
		_, err = f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.pastry}, nil), uuid.New(), acceptingTerms(tomorrowAt(11, 0)))
		require.NoError(t, err)
	})

	t.Run("an_order_not_collected_by_closing_time_is_recorded_as_missed", func(t *testing.T) {
		missed := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 15)
		setStatus(missed.ID, domainorder.StatusReady, "-2 days")
		waiting := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 16)
		setStatus(waiting.ID, domainorder.StatusReady, "1 day")

		marked, err := f.orderUC.MarkNoShows(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, marked)
		staffView, err := f.staff.Get(ctx, uuid.MustParse(missed.ID))
		require.NoError(t, err)
		assert.Equal(t, "no_show", staffView.Status)
		assert.Nil(t, staffView.Timeline[len(staffView.Timeline)-1].ActorRole, "the system recorded it")
		assert.Nil(t, staffView.Payment.RefundRequestedAt, "a missed pickup is not refunded")
		other, err := f.staff.Get(ctx, uuid.MustParse(waiting.ID))
		require.NoError(t, err)
		assert.Equal(t, "ready", other.Status, "its pickup day has not closed")
		again, err := f.orderUC.MarkNoShows(ctx)
		require.NoError(t, err)
		assert.Zero(t, again)
	})
}
