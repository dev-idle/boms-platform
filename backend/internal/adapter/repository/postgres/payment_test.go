package postgres_test

import (
	"context"
	"sync"
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
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// An order placed online is paid before the bakery sees it; a retried checkout
// places it once; a discount counts each customer's orders.
func TestPayment_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	staff := usecase.NewStaffOrderUsecase(f.orders, postgresadapter.NewTicketRepository(f.pool), f.pool, f.outbox, nil, nil,
		postgresadapter.NewPaymentRepository(f.pool), postgresadapter.NewDiscountCodeRepository(f.pool))
	discounts := postgresadapter.NewDiscountCodeRepository(f.pool)
	staffSees := func(orderID string) bool {
		list, _, _, _, err := staff.List(ctx, 1, 100, "")
		require.NoError(t, err)
		for _, order := range list {
			if order.ID == orderID {
				return true
			}
		}
		return false
	}
	newCode := func(code string, discountType domaindiscount.Type, value int64, perCustomer *int32) {
		_, err := discounts.Create(ctx, port.CreateDiscountCodeParams{
			Code: code, DiscountType: discountType, Value: value, MaxUsesPerCustomer: perCustomer, IsActive: true,
			StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
	}
	checkoutWith := func(customer uuid.UUID, code string, hour int) (*dto.OrderResponse, error) {
		_, err := f.cartUC.ApplyDiscount(ctx, customer, dto.ApplyCartDiscountRequest{Code: code})
		require.NoError(t, err)
		return f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(hour, 0)))
	}

	t.Run("the_bakery_sees_an_order_once_it_is_paid", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(12, 0)))
		require.NoError(t, err)
		assert.Equal(t, "awaiting_payment", order.Status)
		assert.Nil(t, order.Payment)
		assert.False(t, staffSees(order.ID), "an order not paid is not the bakery's to see")

		f.pay(t, customer, order.ID)

		paid, err := f.orderUC.Get(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err)
		assert.Equal(t, "confirmed", paid.Status)
		require.NotNil(t, paid.Payment)
		assert.Equal(t, "captured", paid.Payment.Status)
		assert.NotNil(t, paid.Payment.CapturedAt)
		assert.True(t, staffSees(order.ID))

		status, err := f.paymentUC.Capture(ctx, customer, uuid.MustParse(order.ID))
		require.NoError(t, err, "a reloaded return page answers the same")
		assert.Equal(t, domainpayment.StatusCaptured, status)
		_, err = f.paymentUC.Start(ctx, customer, uuid.MustParse(order.ID))
		require.ErrorIs(t, err, domainpayment.ErrNotPayable, "a paid order is not paid twice")
	})

	t.Run("a_retried_checkout_places_the_order_once", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		key := uuid.New()
		first, err := f.orderUC.Checkout(ctx, customer, key, acceptingTerms(tomorrowAt(13, 0)))
		require.NoError(t, err)

		retried, err := f.orderUC.Checkout(ctx, customer, key, acceptingTerms(tomorrowAt(13, 0)))

		require.NoError(t, err, "not an empty cart: the answer the first attempt lost")
		assert.Equal(t, first.ID, retried.ID)
	})

	t.Run("two_requests_with_one_key_place_one_order", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		key := uuid.New()

		var wg sync.WaitGroup
		ids := make([]string, 2)
		errs := make([]error, 2)
		start := make(chan struct{})
		for i := range ids {
			wg.Go(func() {
				<-start
				order, err := f.orderUC.Checkout(ctx, customer, key, acceptingTerms(tomorrowAt(13, 30)))
				errs[i] = err
				if err == nil {
					ids[i] = order.ID
				}
			})
		}
		close(start)
		wg.Wait()

		require.NoError(t, errs[0])
		require.NoError(t, errs[1], "the second waited on the cart and answers with the order")
		assert.Equal(t, ids[0], ids[1])
	})

	t.Run("a_discount_counts_each_customers_orders", func(t *testing.T) {
		once := int32(1)
		newCode("ONCEEACH", domaindiscount.TypePercent, 10, &once)
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		_, err := checkoutWith(customer, "ONCEEACH", 14)
		require.NoError(t, err)

		require.NoError(t, f.fillCart(customer, []uuid.UUID{f.pastry}, nil))
		_, err = checkoutWith(customer, "ONCEEACH", 15)
		require.ErrorIs(t, err, domaindiscount.ErrUsedUpByCustomer)

		_, err = checkoutWith(f.newCustomer(t, []uuid.UUID{f.pastry}, nil), "ONCEEACH", 15)
		require.NoError(t, err, "another customer has their own use")
	})

	t.Run("two_checkouts_race_for_a_codes_last_use", func(t *testing.T) {
		one := int32(1)
		_, err := discounts.Create(ctx, port.CreateDiscountCodeParams{
			Code: "LASTUSE", DiscountType: domaindiscount.TypePercent, Value: 10, MaxUses: &one, IsActive: true,
			StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
		})
		require.NoError(t, err)
		racers := make([]uuid.UUID, 2)
		for i := range racers {
			racers[i] = f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
			_, err := f.cartUC.ApplyDiscount(ctx, racers[i], dto.ApplyCartDiscountRequest{Code: "LASTUSE"})
			require.NoError(t, err)
		}

		var wg sync.WaitGroup
		errs := make([]error, len(racers))
		start := make(chan struct{})
		for i, customer := range racers {
			wg.Go(func() {
				<-start
				_, errs[i] = f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(10, 30)))
			})
		}
		close(start)
		wg.Wait()

		var placed, exhausted int
		for _, err := range errs {
			switch {
			case err == nil:
				placed++
			case assert.ErrorIs(t, err, domaindiscount.ErrExhausted):
				exhausted++
			}
		}
		assert.Equal(t, 1, placed, "the code has one use")
		assert.Equal(t, 1, exhausted)
		code, err := discounts.GetByCode(ctx, "LASTUSE")
		require.NoError(t, err)
		assert.Equal(t, int32(1), code.UsedCount)
	})

	t.Run("a_payment_refused_after_review_starts_over_with_a_new_paypal_order", func(t *testing.T) {
		payments := postgresadapter.NewPaymentRepository(f.pool)
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(17, 0)))
		require.NoError(t, err)
		params := port.CreatePaymentParams{
			OrderID: uuid.MustParse(order.ID), Provider: domainpayment.ProviderPayPal, ProviderOrderID: "REVIEWED-1",
			ApproveURL: "https://www.sandbox.paypal.com/checkoutnow?token=REVIEWED-1", AmountCents: order.TotalCents, Currency: domainpayment.Currency,
		}
		created, err := payments.Create(ctx, params)
		require.NoError(t, err)
		_, err = payments.Create(ctx, params)
		require.ErrorIs(t, err, apperrors.ErrConflict, "one payment per order")

		review := domainpayment.Capture{ID: "CAPTURE-R", Status: domainpayment.StatusPending, AmountCents: order.TotalCents, Currency: domainpayment.Currency}
		_, err = payments.RecordCapture(ctx, created.ID, review)
		require.NoError(t, err)
		review.Status = domainpayment.StatusDenied
		_, err = payments.RecordCapture(ctx, created.ID, review)
		require.NoError(t, err)
		review.Status = domainpayment.StatusCaptured
		_, err = payments.RecordCapture(ctx, created.ID, review)
		require.ErrorIs(t, err, apperrors.ErrNotFound, "a refused capture takes no later answer")

		restarted, err := payments.Restart(ctx, created.OrderID, "REVIEWED-2", "https://www.sandbox.paypal.com/checkoutnow?token=REVIEWED-2")
		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusCreated, restarted.Status)
		assert.Nil(t, restarted.CaptureID)
		found, err := payments.GetByProviderOrderID(ctx, domainpayment.ProviderPayPal, "REVIEWED-2")
		require.NoError(t, err)
		assert.Equal(t, created.ID, found.ID)
		_, err = payments.GetByProviderOrderID(ctx, domainpayment.ProviderPayPal, "REVIEWED-1")
		require.ErrorIs(t, err, apperrors.ErrNotFound, "a replaced PayPal order names no payment")
	})

	t.Run("an_order_with_nothing_to_pay_is_confirmed_at_once", func(t *testing.T) {
		newCode("FREEPASTRY", domaindiscount.TypeFixedCents, 300, nil)
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)

		order, err := checkoutWith(customer, "FREEPASTRY", 16)

		require.NoError(t, err)
		assert.Equal(t, int64(0), order.TotalCents)
		assert.Equal(t, "confirmed", order.Status)
		assert.Nil(t, order.Payment, "nothing to take")
		assert.True(t, staffSees(order.ID))
	})
}

// An order not paid in time expires and gives back its pickup slot and its
// discount use; one just past its due time takes no payment and waits out the
// grace a capture started in time needs.
func TestOrderExpiry_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newCheckoutFixture(t, 8)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	overdue := func(orderID, by string) {
		_, err := raw.Exec(ctx, "UPDATE orders SET payment_due_at = now() - $2::interval WHERE id = $1", orderID, by)
		require.NoError(t, err)
	}
	f.setSettings(t, func(s *domainstore.Settings) { s.SlotCapacity = 1 })
	one := int32(1)
	_, err = postgresadapter.NewDiscountCodeRepository(f.pool).Create(ctx, port.CreateDiscountCodeParams{
		Code: "LASTONE", DiscountType: domaindiscount.TypePercent, Value: 10, MaxUses: &one, IsActive: true,
		StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)
	withLastCode := func() uuid.UUID {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		_, err := f.cartUC.ApplyDiscount(ctx, customer, dto.ApplyCartDiscountRequest{Code: "LASTONE"})
		require.NoError(t, err)
		return customer
	}
	slot := tomorrowAt(11, 0)

	unpaid := withLastCode()
	order, err := f.orderUC.Checkout(ctx, unpaid, uuid.New(), acceptingTerms(slot))
	require.NoError(t, err)
	_, err = f.paymentUC.Start(ctx, unpaid, uuid.MustParse(order.ID))
	require.NoError(t, err)
	late := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
	justDue, err := f.orderUC.Checkout(ctx, late, uuid.New(), acceptingTerms(tomorrowAt(12, 0)))
	require.NoError(t, err)
	_, err = f.paymentUC.Start(ctx, late, uuid.MustParse(justDue.ID))
	require.NoError(t, err)
	overdue(order.ID, "3 minutes")
	overdue(justDue.ID, "1 second")
	_, err = f.paymentUC.Capture(ctx, late, uuid.MustParse(justDue.ID))
	require.ErrorIs(t, err, domainpayment.ErrNotPayable, "the time to pay is up")

	expired, err := f.paymentUC.ExpireOverdue(ctx)

	require.NoError(t, err)
	assert.Equal(t, 1, expired)
	got, err := f.orderUC.Get(ctx, unpaid, uuid.MustParse(order.ID))
	require.NoError(t, err)
	assert.Equal(t, "expired", got.Status)
	var bySystem bool
	require.NoError(t, raw.QueryRow(ctx,
		"SELECT actor_id IS NULL AND actor_role IS NULL FROM order_status_events WHERE order_id = $1 AND to_status = 'expired'",
		order.ID).Scan(&bySystem))
	assert.True(t, bySystem, "no person expired it")
	staff := usecase.NewStaffOrderUsecase(f.orders, postgresadapter.NewTicketRepository(f.pool), f.pool, f.outbox, nil, nil,
		postgresadapter.NewPaymentRepository(f.pool), postgresadapter.NewDiscountCodeRepository(f.pool))
	_, err = staff.Get(ctx, uuid.MustParse(order.ID))
	require.ErrorIs(t, err, domainorder.ErrNotFound, "an order never paid is not the bakery's to see")

	_, err = f.orderUC.Checkout(ctx, withLastCode(), uuid.New(), acceptingTerms(slot))
	require.NoError(t, err, "the slot and the discount use are free again")

	waiting, err := f.orderUC.Get(ctx, late, uuid.MustParse(justDue.ID))
	require.NoError(t, err)
	assert.Equal(t, "awaiting_payment", waiting.Status, "a capture started in time may still be finishing")
	again, err := f.paymentUC.ExpireOverdue(ctx)
	require.NoError(t, err)
	assert.Zero(t, again, "an expired order expires once")
}
