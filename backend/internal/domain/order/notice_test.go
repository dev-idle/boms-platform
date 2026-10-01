package order

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
)

func TestNoticeFor(t *testing.T) {
	t.Parallel()

	order := func(status Status) Order {
		return Order{ID: uuid.New(), UserID: uuid.New(), Status: status}
	}

	cases := []struct {
		name   string
		event  func(Order) domainevent.Event
		status Status
		want   Notice
	}{
		{"a_paid_order_is_received", func(o Order) domainevent.Event { return StatusChangedEvent(StatusAwaitingPayment, o) }, StatusConfirmed, NoticePlaced},
		{"ready_to_collect", func(o Order) domainevent.Event { return StatusChangedEvent(StatusInProduction, o) }, StatusReady, NoticeReady},
		{"cancelled", func(o Order) domainevent.Event { return StatusChangedEvent(StatusConfirmed, o) }, StatusCancelled, NoticeCancelled},
		{"expired_unpaid", func(o Order) domainevent.Event { return StatusChangedEvent(StatusAwaitingPayment, o) }, StatusExpired, NoticeExpired},
		{"not_collected", func(o Order) domainevent.Event { return StatusChangedEvent(StatusReady, o) }, StatusNoShow, NoticeNoShow},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o := order(tc.status)

			orderID, notice, ok := NoticeFor(tc.event(o))

			require.True(t, ok)
			assert.Equal(t, tc.want, notice)
			assert.Equal(t, o.ID, orderID)
		})
	}

	t.Run("steps_inside_the_bakery_send_nothing", func(t *testing.T) {
		t.Parallel()
		for _, to := range []Status{StatusConfirmed, StatusInProduction, StatusFulfilled} {
			_, _, ok := NoticeFor(StatusChangedEvent(StatusPending, order(to)))
			assert.False(t, ok, to)
		}
	})

	t.Run("an_order_not_paid_yet_sends_nothing", func(t *testing.T) {
		t.Parallel()
		_, _, ok := NoticeFor(CreatedEvent(order(StatusAwaitingPayment)))
		assert.False(t, ok, "the customer hears once the payment is taken")
	})

	t.Run("other_topics_send_nothing", func(t *testing.T) {
		t.Parallel()
		ticket := Ticket{ID: uuid.New(), OrderID: uuid.New(), Status: TicketReady}
		_, _, ok := NoticeFor(TicketChangedEvent(uuid.New(), StatusReady, ticket))
		assert.False(t, ok)
	})

	t.Run("an_event_without_an_order_id_sends_nothing", func(t *testing.T) {
		t.Parallel()
		e := CreatedEvent(order(StatusPending))
		e.Data["order_id"] = "not-an-id"
		_, _, ok := NoticeFor(e)
		assert.False(t, ok)
	})
}

func TestNotice_StillApplies(t *testing.T) {
	t.Parallel()

	cases := []struct {
		notice  Notice
		current Status
		want    bool
	}{
		{NoticePlaced, StatusPending, true},
		{NoticePlaced, StatusFulfilled, true},
		{NoticePlaced, StatusCancelled, false},
		{NoticeReady, StatusReady, true},
		{NoticeReady, StatusFulfilled, false},
		{NoticeReady, StatusCancelled, false},
		{NoticeCancelled, StatusCancelled, true},
		{NoticeExpired, StatusExpired, true},
		{NoticeExpired, StatusConfirmed, false},
		{NoticeNoShow, StatusNoShow, true},
		{NoticeNoShow, StatusFulfilled, false},
		{Notice("shipped"), StatusReady, false},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, tc.notice.StillApplies(tc.current), "%s while %s", tc.notice, tc.current)
	}
	assert.True(t, NoticeReady.Valid())
	assert.False(t, Notice("shipped").Valid())
}
