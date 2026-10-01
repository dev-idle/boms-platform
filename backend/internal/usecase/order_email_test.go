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

// emailOrders is the order repository an order email reads; any other call panics.
type emailOrders struct {
	port.OrderRepository
	row      *port.StaffOrderListRow
	getErr   error
	items    []domainorder.Item
	itemsErr error
	history  []domainorder.StatusEvent
}

func (o *emailOrders) ListStatusEvents(context.Context, uuid.UUID) ([]domainorder.StatusEvent, error) {
	return o.history, nil
}

// emailPayments holds the order's payment, if it has one; any other call panics.
type emailPayments struct {
	port.PaymentRepository
	payment *domainpayment.Payment
}

func (p *emailPayments) GetByOrderID(context.Context, uuid.UUID) (*domainpayment.Payment, error) {
	if p.payment == nil {
		return nil, apperrors.ErrNotFound
	}
	return p.payment, nil
}

func (o *emailOrders) StaffGetByID(context.Context, uuid.UUID) (*port.StaffOrderListRow, error) {
	return o.row, o.getErr
}

func (o *emailOrders) ListItemsByOrderID(context.Context, uuid.UUID) ([]domainorder.Item, error) {
	return o.items, o.itemsErr
}

type fakeComposer struct {
	got []port.OrderEmail
	err error
}

func (c *fakeComposer) ComposeOrderEmail(msg port.OrderEmail) (port.Email, error) {
	c.got = append(c.got, msg)
	return port.Email{To: msg.To, Subject: "subject " + string(msg.Notice)}, c.err
}

type fakeMailer struct {
	sent []port.Email
	err  error
}

func (m *fakeMailer) Send(_ context.Context, e port.Email) error {
	if m.err != nil {
		return m.err
	}
	m.sent = append(m.sent, e)
	return nil
}

type emailFixture struct {
	orders   *emailOrders
	payments *emailPayments
	composer *fakeComposer
	mailer   *fakeMailer
	task     port.OrderEmailTask
}

func newEmailFixture(notice domainorder.Notice, status domainorder.Status) *emailFixture {
	name := "Mai"
	orderID := uuid.New()
	return &emailFixture{
		orders: &emailOrders{
			row: &port.StaffOrderListRow{
				Order:                 domainorder.Order{ID: orderID, Code: "CH-260930-007", Status: status},
				CustomerEmail:         "mai@example.com",
				CustomerEmailVerified: true,
				CustomerDisplayName:   &name,
			},
			items: []domainorder.Item{{Name: "Almond Croissant", Quantity: 2, LineTotalCents: 900}},
		},
		payments: &emailPayments{},
		composer: &fakeComposer{},
		mailer:   &fakeMailer{},
		task:     port.OrderEmailTask{EventID: uuid.New(), OrderID: orderID, Notice: notice},
	}
}

// emailCodes signs pickup codes as the worker does, with a test key.
var emailCodes = domainorder.NewPickupCodes("an-email-test-key-of-thirty-two-chars")

func (f *emailFixture) send() error {
	uc := usecase.NewOrderEmailUsecase(f.orders, f.payments, f.composer, f.mailer, emailCodes, zap.NewNop())
	return uc.Send(context.Background(), f.task)
}

func TestOrderEmailUsecase_Send(t *testing.T) {
	t.Parallel()

	t.Run("a_cancellation_names_the_bakerys_reason_and_the_refund", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticeCancelled, domainorder.StatusCancelled)
		f.orders.history = []domainorder.StatusEvent{
			{To: domainorder.StatusConfirmed},
			{To: domainorder.StatusCancelled, Reason: "The oven broke down"},
		}
		requested := time.Now()
		f.payments.payment = &domainpayment.Payment{AmountCents: 3940, RefundRequestedAt: &requested}

		require.NoError(t, f.send())

		require.Len(t, f.composer.got, 1)
		assert.Equal(t, "The oven broke down", f.composer.got[0].Reason)
		assert.Equal(t, int64(3940), f.composer.got[0].RefundCents)
	})

	t.Run("an_order_never_paid_is_cancelled_with_nothing_to_return", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticeCancelled, domainorder.StatusCancelled)
		f.orders.history = []domainorder.StatusEvent{{To: domainorder.StatusCancelled}}

		require.NoError(t, f.send())

		require.Len(t, f.composer.got, 1)
		assert.Empty(t, f.composer.got[0].Reason, "the customer cancelled it")
		assert.Zero(t, f.composer.got[0].RefundCents)
	})

	t.Run("mails_the_customer_the_notice", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticeReady, domainorder.StatusReady)

		require.NoError(t, f.send())

		require.Len(t, f.composer.got, 1)
		msg := f.composer.got[0]
		assert.Equal(t, domainorder.NoticeReady, msg.Notice)
		assert.Equal(t, "mai@example.com", msg.To)
		assert.Equal(t, "Mai", msg.CustomerName)
		assert.Equal(t, "CH-260930-007", msg.Order.Code)
		assert.Equal(t, emailCodes.Of(f.task.OrderID), msg.PickupCode, "the code to collect it")
		assert.Len(t, msg.Items, 1)
		require.Len(t, f.mailer.sent, 1)
		assert.Equal(t, "subject ready", f.mailer.sent[0].Subject)
	})

	t.Run("a_customer_without_a_name_gets_it_too", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticePlaced, domainorder.StatusPending)
		f.orders.row.CustomerDisplayName = nil

		require.NoError(t, f.send())

		assert.Empty(t, f.composer.got[0].CustomerName)
		assert.Empty(t, f.composer.got[0].PickupCode, "no code while staff review the order")
		assert.Len(t, f.mailer.sent, 1)
	})

	t.Run("skips_a_notice_the_order_has_moved_past", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticeReady, domainorder.StatusFulfilled)

		require.NoError(t, f.send())

		assert.Empty(t, f.mailer.sent, "the order was already collected")
	})

	t.Run("skips_an_address_nobody_confirmed", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticePlaced, domainorder.StatusPending)
		f.orders.row.CustomerEmailVerified = false

		require.NoError(t, f.send())

		assert.Empty(t, f.mailer.sent, "an order from before confirmation was required mails nobody")
	})

	t.Run("skips_a_customer_who_closed_their_account", func(t *testing.T) {
		t.Parallel()
		f := newEmailFixture(domainorder.NoticeCancelled, domainorder.StatusCancelled)
		f.orders.row, f.orders.getErr = nil, apperrors.ErrNotFound

		require.NoError(t, f.send())

		assert.Empty(t, f.mailer.sent)
	})

	t.Run("a_passing_failure_is_left_to_retry", func(t *testing.T) {
		t.Parallel()
		down := errors.New("mail server unreachable")
		for _, set := range []func(*emailFixture){
			func(f *emailFixture) { f.orders.getErr = apperrors.ErrInternal },
			func(f *emailFixture) { f.orders.itemsErr = apperrors.ErrInternal },
			func(f *emailFixture) { f.mailer.err = down },
		} {
			f := newEmailFixture(domainorder.NoticePlaced, domainorder.StatusPending)
			set(f)

			err := f.send()

			require.Error(t, err)
			assert.NotErrorIs(t, err, port.ErrEmailUndeliverable)
		}
	})

	t.Run("an_email_no_retry_can_send_is_final", func(t *testing.T) {
		t.Parallel()
		refused := newEmailFixture(domainorder.NoticePlaced, domainorder.StatusPending)
		refused.mailer.err = fmt.Errorf("%w: 550", port.ErrEmailUndeliverable)
		require.ErrorIs(t, refused.send(), port.ErrEmailUndeliverable)

		unwritable := newEmailFixture(domainorder.NoticePlaced, domainorder.StatusPending)
		unwritable.composer.err = errors.New("template broke")
		require.ErrorIs(t, unwritable.send(), port.ErrEmailUndeliverable)
		assert.Empty(t, unwritable.mailer.sent)
	})
}
