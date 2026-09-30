package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpayment "github.com/boms/backend/internal/domain/payment"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// paymentOrders holds one order; a status move succeeds only from the status
// it is in, as the guarded UPDATE does. overdue puts it past its due time and
// the grace after it. Any other call panics.
type paymentOrders struct {
	port.OrderRepository
	order   domainorder.Order
	overdue bool
	history []domainorder.Status
	actors  []*port.OrderActor
}

func (f *paymentOrders) ListDueUnpaid(context.Context, time.Duration, int32) ([]uuid.UUID, error) {
	if f.order.Status != domainorder.StatusAwaitingPayment || !f.overdue {
		return nil, nil
	}
	return []uuid.UUID{f.order.ID}, nil
}

func (f *paymentOrders) Expire(context.Context, uuid.UUID, time.Duration) (*domainorder.Order, error) {
	if f.order.Status != domainorder.StatusAwaitingPayment || !f.overdue {
		return nil, apperrors.ErrNotFound
	}
	f.order.Status = domainorder.StatusExpired
	order := f.order
	return &order, nil
}

func (f *paymentOrders) GetByIDForUser(_ context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error) {
	if userID != f.order.UserID || orderID != f.order.ID {
		return nil, apperrors.ErrNotFound
	}
	order := f.order
	return &order, nil
}

func (f *paymentOrders) LockForUpdate(context.Context, uuid.UUID) (*domainorder.Order, error) {
	order := f.order
	return &order, nil
}

func (f *paymentOrders) UpdateStatus(_ context.Context, params port.UpdateOrderStatusParams) (*domainorder.Order, error) {
	if f.order.Status != params.FromStatus {
		return nil, apperrors.ErrNotFound
	}
	f.order.Status = params.ToStatus
	order := f.order
	return &order, nil
}

func (f *paymentOrders) AddStatusEvent(_ context.Context, params port.AddOrderStatusEventParams) error {
	f.history = append(f.history, params.To)
	f.actors = append(f.actors, params.Actor)
	return nil
}

// releasedCodes records the discount uses given back; any other call panics.
type releasedCodes struct {
	port.DiscountCodeRepository
	released []uuid.UUID
}

func (f *releasedCodes) ReleaseUse(_ context.Context, id uuid.UUID) error {
	f.released = append(f.released, id)
	return nil
}

// cancelledTicketsNone has no ticket to cancel; any other call panics.
type cancelledTicketsNone struct{ port.TicketRepository }

func (cancelledTicketsNone) CancelForOrder(context.Context, uuid.UUID) ([]domainorder.Ticket, error) {
	return nil, nil
}

// memoryPayments keeps at most one payment, with the repository's guards.
type memoryPayments struct {
	payment   *domainpayment.Payment
	createErr error
}

func (f *memoryPayments) Create(_ context.Context, params port.CreatePaymentParams) (*domainpayment.Payment, error) {
	if f.createErr != nil {
		return nil, f.createErr
	}
	f.payment = &domainpayment.Payment{
		ID: uuid.New(), OrderID: params.OrderID, Provider: params.Provider, ProviderOrderID: params.ProviderOrderID,
		ApproveURL: params.ApproveURL, Status: domainpayment.StatusCreated, AmountCents: params.AmountCents, Currency: params.Currency,
	}
	return f.payment, nil
}

func (f *memoryPayments) GetByOrderID(context.Context, uuid.UUID) (*domainpayment.Payment, error) {
	if f.payment == nil {
		return nil, apperrors.ErrNotFound
	}
	p := *f.payment
	return &p, nil
}

func (f *memoryPayments) GetByProviderOrderID(_ context.Context, _ domainpayment.Provider, id string) (*domainpayment.Payment, error) {
	if f.payment == nil || f.payment.ProviderOrderID != id {
		return nil, apperrors.ErrNotFound
	}
	p := *f.payment
	return &p, nil
}

func (f *memoryPayments) Restart(_ context.Context, _ uuid.UUID, providerOrderID, approveURL string) (*domainpayment.Payment, error) {
	if f.payment.Status != domainpayment.StatusDenied {
		return nil, apperrors.ErrNotFound
	}
	f.payment.ProviderOrderID, f.payment.ApproveURL, f.payment.Status = providerOrderID, approveURL, domainpayment.StatusCreated
	return f.payment, nil
}

func (f *memoryPayments) RecordCapture(_ context.Context, _ uuid.UUID, capture domainpayment.Capture) (*domainpayment.Payment, error) {
	if f.payment.Status != domainpayment.StatusCreated && f.payment.Status != domainpayment.StatusPending {
		return nil, apperrors.ErrNotFound
	}
	f.payment.Status, f.payment.CaptureID = capture.Status, &capture.ID
	return f.payment, nil
}

// fakeGateway answers as told and counts the provider orders it made and
// the captures it was asked for.
type fakeGateway struct {
	created    int
	captures   int
	capture    domainpayment.Capture
	captureErr error
	lookup     *domainpayment.Capture
	lookupErr  error
	event      port.PaymentWebhookEvent
	webhookErr error
}

func (g *fakeGateway) Lookup(context.Context, string) (*domainpayment.Capture, error) {
	return g.lookup, g.lookupErr
}

func (g *fakeGateway) CreateOrder(context.Context, port.PaymentOrderRequest) (string, string, error) {
	g.created++
	id := fmt.Sprintf("PAYPAL-%d", g.created)
	return id, "https://www.sandbox.paypal.com/checkoutnow?token=" + id, nil
}

func (g *fakeGateway) Capture(context.Context, string) (domainpayment.Capture, error) {
	g.captures++
	return g.capture, g.captureErr
}

func (g *fakeGateway) VerifyWebhook(context.Context, func(string) string, []byte) (port.PaymentWebhookEvent, error) {
	return g.event, g.webhookErr
}

type paymentFixture struct {
	orders   *paymentOrders
	codes    *releasedCodes
	payments *memoryPayments
	gateway  *fakeGateway
	outbox   *recordingOutbox
	customer uuid.UUID
	orderID  uuid.UUID
}

func newPaymentFixture() *paymentFixture {
	codeID := uuid.New()
	due := time.Now().Add(15 * time.Minute)
	order := domainorder.Order{
		ID: uuid.New(), UserID: uuid.New(), Status: domainorder.StatusAwaitingPayment, TotalCents: 1250, DiscountCodeID: &codeID,
		PaymentDueAt: &due,
	}
	return &paymentFixture{
		orders:   &paymentOrders{order: order},
		codes:    &releasedCodes{},
		payments: &memoryPayments{},
		gateway: &fakeGateway{capture: domainpayment.Capture{
			ID: "CAPTURE-1", Status: domainpayment.StatusCaptured, AmountCents: 1250, Currency: domainpayment.Currency,
		}},
		outbox:   &recordingOutbox{},
		customer: order.UserID,
		orderID:  order.ID,
	}
}

func (f *paymentFixture) usecase() *usecase.PaymentUsecase {
	return usecase.NewPaymentUsecase(passthroughTxManager{}, f.orders, f.codes, cancelledTicketsNone{}, f.payments, f.gateway, f.outbox,
		"https://shop.example", zap.NewNop())
}

// started gives the order a payment awaiting approval, or in status.
func (f *paymentFixture) started(status domainpayment.Status) {
	f.payments.payment = &domainpayment.Payment{
		ID: uuid.New(), OrderID: f.orderID, ProviderOrderID: "PAYPAL-0", ApproveURL: "https://www.sandbox.paypal.com/checkoutnow?token=PAYPAL-0",
		Status: status, AmountCents: 1250, Currency: domainpayment.Currency,
	}
}

func TestPaymentUsecase_Start(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("asks_paypal_once_and_returns_the_same_page", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()

		first, err := f.usecase().Start(ctx, f.customer, f.orderID)
		require.NoError(t, err)
		again, err := f.usecase().Start(ctx, f.customer, f.orderID)
		require.NoError(t, err)

		assert.Equal(t, first, again)
		assert.Equal(t, 1, f.gateway.created)
		assert.Equal(t, int64(1250), f.payments.payment.AmountCents, "the order's total, set by the shop")
	})

	t.Run("a_denied_payment_starts_over", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusDenied)

		page, err := f.usecase().Start(ctx, f.customer, f.orderID)

		require.NoError(t, err)
		assert.Contains(t, page, "PAYPAL-1")
		assert.Equal(t, domainpayment.StatusCreated, f.payments.payment.Status)
	})

	t.Run("a_request_that_lost_the_race_returns_the_winners_page", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		winner := &domainpayment.Payment{Status: domainpayment.StatusCreated, ApproveURL: "https://www.sandbox.paypal.com/checkoutnow?token=WINNER"}
		racing := &racingPayments{memoryPayments: &memoryPayments{createErr: apperrors.ErrConflict}, winner: winner}
		uc := usecase.NewPaymentUsecase(passthroughTxManager{}, f.orders, nil, nil, racing, f.gateway, f.outbox, "https://shop.example", zap.NewNop())

		page, err := uc.Start(ctx, f.customer, f.orderID)

		require.NoError(t, err)
		assert.Equal(t, winner.ApproveURL, page)
	})

	t.Run("an_order_paid_under_review_closed_or_late_is_not_payable", func(t *testing.T) {
		t.Parallel()
		for _, status := range []domainpayment.Status{domainpayment.StatusPending, domainpayment.StatusCaptured} {
			f := newPaymentFixture()
			f.started(status)

			_, err := f.usecase().Start(ctx, f.customer, f.orderID)

			require.ErrorIs(t, err, domainpayment.ErrNotPayable, status)
		}
		closed := newPaymentFixture()
		closed.orders.order.Status = domainorder.StatusCancelled
		_, err := closed.usecase().Start(ctx, closed.customer, closed.orderID)
		require.ErrorIs(t, err, domainpayment.ErrNotPayable)
		assert.Zero(t, closed.gateway.created)
		late := newPaymentFixture()
		due := time.Now().Add(-time.Second)
		late.orders.order.PaymentDueAt = &due
		_, err = late.usecase().Start(ctx, late.customer, late.orderID)
		require.ErrorIs(t, err, domainpayment.ErrNotPayable)
		assert.Zero(t, late.gateway.created)
	})

	t.Run("another_customers_order_is_not_found", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()

		_, err := f.usecase().Start(ctx, uuid.New(), f.orderID)

		require.ErrorIs(t, err, domainorder.ErrNotFound)
	})
}

// racingPayments finds no payment on the first read and the winner's after a
// refused write, as when two requests start paying the same order at once.
type racingPayments struct {
	*memoryPayments
	winner *domainpayment.Payment
	reads  int
}

func (r *racingPayments) GetByOrderID(context.Context, uuid.UUID) (*domainpayment.Payment, error) {
	r.reads++
	if r.reads == 1 {
		return nil, apperrors.ErrNotFound
	}
	return r.winner, nil
}

func TestPaymentUsecase_Capture(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("takes_the_money_and_confirms_the_order_in_one_go", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)

		status, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusCaptured, status)
		assert.Equal(t, domainorder.StatusConfirmed, f.orders.order.Status)
		assert.Equal(t, []domainorder.Status{domainorder.StatusConfirmed}, f.orders.history)
		require.Len(t, f.outbox.events, 1)
		assert.Equal(t, "awaiting_payment", f.outbox.events[0].Data["from"], "the email that the order is received goes out")
	})

	t.Run("a_capture_under_review_leaves_the_order_waiting", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		f.gateway.capture.Status = domainpayment.StatusPending

		status, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusPending, status)
		assert.Equal(t, domainorder.StatusAwaitingPayment, f.orders.order.Status)
	})

	t.Run("asking_again_answers_the_same_without_paypal", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCaptured)
		f.gateway.captureErr = errors.New("paypal must not be asked")

		status, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusCaptured, status)
	})

	t.Run("a_declined_funding_lets_the_buyer_try_again", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		f.gateway.captureErr = domainpayment.ErrNotCompleted

		_, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.ErrorIs(t, err, domainpayment.ErrNotCompleted)
		assert.Equal(t, domainpayment.StatusCreated, f.payments.payment.Status)

		denied := newPaymentFixture()
		denied.started(domainpayment.StatusDenied)
		_, err = denied.usecase().Capture(ctx, denied.customer, denied.orderID)
		require.ErrorIs(t, err, domainpayment.ErrNotCompleted, "a denied payment starts over through Start")
	})

	t.Run("a_capture_paypal_refuses_lets_the_buyer_start_over", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		f.gateway.capture.Status = domainpayment.StatusDenied

		_, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.ErrorIs(t, err, domainpayment.ErrNotCompleted)
		assert.Equal(t, domainpayment.StatusDenied, f.payments.payment.Status)
		_, err = f.usecase().Start(ctx, f.customer, f.orderID)
		require.NoError(t, err)
		assert.Equal(t, 1, f.gateway.created, "a new PayPal order")
	})

	t.Run("a_different_amount_confirms_nothing", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		f.gateway.capture.AmountCents = 1

		_, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.ErrorIs(t, err, domainpayment.ErrAmountMismatch)
		assert.Equal(t, domainpayment.StatusCreated, f.payments.payment.Status)
		assert.Equal(t, domainorder.StatusAwaitingPayment, f.orders.order.Status)
	})

	t.Run("money_taken_for_an_order_closed_meanwhile_stays_on_record", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		orders := &closingOrders{paymentOrders: f.orders}
		uc := usecase.NewPaymentUsecase(passthroughTxManager{}, orders, nil, nil, f.payments, f.gateway, f.outbox, "https://shop.example", zap.NewNop())

		status, err := uc.Capture(ctx, f.customer, f.orderID)

		require.NoError(t, err)
		assert.Equal(t, domainpayment.StatusCaptured, status)
		assert.Equal(t, domainpayment.StatusCaptured, f.payments.payment.Status, "the money can be returned")
		assert.Equal(t, domainorder.StatusCancelled, f.orders.order.Status)
	})

	t.Run("an_order_that_expired_takes_no_payment", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		f.orders.order.Status = domainorder.StatusExpired

		_, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.ErrorIs(t, err, domainpayment.ErrNotPayable)
		assert.Zero(t, f.gateway.captures, "PayPal is not asked to take the money")
	})

	t.Run("an_order_past_its_time_to_pay_takes_no_payment", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		due := time.Now().Add(-time.Second)
		f.orders.order.PaymentDueAt = &due

		_, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.ErrorIs(t, err, domainpayment.ErrNotPayable)
		assert.Zero(t, f.gateway.captures, "asking again cannot keep the order")
	})

	t.Run("an_order_without_a_payment_is_not_payable_here", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()

		_, err := f.usecase().Capture(ctx, f.customer, f.orderID)

		require.ErrorIs(t, err, domainpayment.ErrNotPayable)
	})
}

// closingOrders cancels the order the moment the payment locks it, as a
// cancellation landing while the buyer pays would.
type closingOrders struct{ *paymentOrders }

func (c *closingOrders) LockForUpdate(ctx context.Context, id uuid.UUID) (*domainorder.Order, error) {
	c.order.Status = domainorder.StatusCancelled
	return c.paymentOrders.LockForUpdate(ctx, id)
}

func TestPaymentUsecase_HandleWebhook(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	notice := func(status domainpayment.Status) port.PaymentWebhookEvent {
		return port.PaymentWebhookEvent{ID: "WH-1", ProviderOrderID: "PAYPAL-0", Capture: &domainpayment.Capture{
			ID: "CAPTURE-1", Status: status, AmountCents: 1250, Currency: domainpayment.Currency,
		}}
	}

	t.Run("a_reviewed_capture_that_went_through_confirms_the_order", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusPending)
		f.gateway.event = notice(domainpayment.StatusCaptured)

		require.NoError(t, f.usecase().HandleWebhook(ctx, nil, nil))
		require.NoError(t, f.usecase().HandleWebhook(ctx, nil, nil), "a notice delivered again changes nothing")

		assert.Equal(t, domainpayment.StatusCaptured, f.payments.payment.Status)
		assert.Equal(t, []domainorder.Status{domainorder.StatusConfirmed}, f.orders.history)
	})

	t.Run("a_refused_capture_lets_the_customer_pay_again", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusPending)
		f.gateway.event = notice(domainpayment.StatusDenied)

		require.NoError(t, f.usecase().HandleWebhook(ctx, nil, nil))

		assert.Equal(t, domainpayment.StatusDenied, f.payments.payment.Status)
		assert.Equal(t, domainorder.StatusAwaitingPayment, f.orders.order.Status)
	})

	t.Run("an_unsigned_delivery_is_refused", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.gateway.webhookErr = domainpayment.ErrWebhookInvalid

		require.ErrorIs(t, f.usecase().HandleWebhook(ctx, nil, nil), domainpayment.ErrWebhookInvalid)
	})

	t.Run("notices_about_other_things_or_other_orders_are_acknowledged", func(t *testing.T) {
		t.Parallel()
		f := newPaymentFixture()
		f.started(domainpayment.StatusCreated)
		f.gateway.event = port.PaymentWebhookEvent{ID: "WH-2"}
		require.NoError(t, f.usecase().HandleWebhook(ctx, nil, nil))

		f.gateway.event = notice(domainpayment.StatusCaptured)
		f.gateway.event.ProviderOrderID = "PAYPAL-OTHER"
		require.NoError(t, f.usecase().HandleWebhook(ctx, nil, nil))

		assert.Equal(t, domainpayment.StatusCreated, f.payments.payment.Status)
	})
}

func TestPaymentUsecase_ExpireOverdue(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	overdue := func() *paymentFixture {
		f := newPaymentFixture()
		f.orders.overdue = true
		return f
	}
	answer := func(status domainpayment.Status) *domainpayment.Capture {
		return &domainpayment.Capture{ID: "CAPTURE-1", Status: status, AmountCents: 1250, Currency: domainpayment.Currency}
	}

	t.Run("an_unpaid_order_expires_and_gives_back_its_discount", func(t *testing.T) {
		t.Parallel()
		f := overdue()

		expired, err := f.usecase().ExpireOverdue(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, expired)
		assert.Equal(t, domainorder.StatusExpired, f.orders.order.Status)
		assert.Equal(t, []domainorder.Status{domainorder.StatusExpired}, f.orders.history)
		assert.Equal(t, []*port.OrderActor{nil}, f.orders.actors, "the system expired it")
		assert.Equal(t, []uuid.UUID{*f.orders.order.DiscountCodeID}, f.codes.released)
		require.NotEmpty(t, f.outbox.events)
		assert.Equal(t, "expired", f.outbox.events[0].Data["status"], "the customer is told")
	})

	t.Run("a_buyer_who_paid_after_all_is_confirmed_instead", func(t *testing.T) {
		t.Parallel()
		f := overdue()
		f.started(domainpayment.StatusCreated)
		f.gateway.lookup = answer(domainpayment.StatusCaptured)

		expired, err := f.usecase().ExpireOverdue(ctx)

		require.NoError(t, err)
		assert.Zero(t, expired)
		assert.Equal(t, domainorder.StatusConfirmed, f.orders.order.Status)
		assert.Equal(t, domainpayment.StatusCaptured, f.payments.payment.Status)
		assert.Empty(t, f.codes.released)
	})

	t.Run("a_capture_paypal_still_reviews_keeps_the_order_waiting", func(t *testing.T) {
		t.Parallel()
		f := overdue()
		f.started(domainpayment.StatusCreated)
		f.gateway.lookup = answer(domainpayment.StatusPending)

		expired, err := f.usecase().ExpireOverdue(ctx)

		require.NoError(t, err)
		assert.Zero(t, expired)
		assert.Equal(t, domainorder.StatusAwaitingPayment, f.orders.order.Status)
		assert.Equal(t, domainpayment.StatusPending, f.payments.payment.Status)
	})

	t.Run("a_refused_capture_expires_the_order", func(t *testing.T) {
		t.Parallel()
		f := overdue()
		f.started(domainpayment.StatusPending)
		f.gateway.lookup = answer(domainpayment.StatusDenied)

		expired, err := f.usecase().ExpireOverdue(ctx)

		require.NoError(t, err)
		assert.Equal(t, 1, expired)
		assert.Equal(t, domainpayment.StatusDenied, f.payments.payment.Status)
		assert.Equal(t, domainorder.StatusExpired, f.orders.order.Status)
	})

	t.Run("a_failure_is_reported_and_expires_nothing", func(t *testing.T) {
		t.Parallel()
		f := overdue()
		f.started(domainpayment.StatusCreated)
		down := errors.New("paypal unavailable")
		f.gateway.lookupErr = down

		expired, err := f.usecase().ExpireOverdue(ctx)

		require.ErrorIs(t, err, down)
		assert.Zero(t, expired)
		assert.Equal(t, domainorder.StatusAwaitingPayment, f.orders.order.Status, "an order is not expired unasked")
	})
}
