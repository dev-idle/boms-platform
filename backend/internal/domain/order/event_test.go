package order

import (
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestOrderEvents(t *testing.T) {
	t.Parallel()

	order := func(status Status) Order {
		customer := uuid.New()
		return Order{ID: uuid.New(), UserID: &customer, Status: status}
	}

	t.Run("a_new_order_is_told_to_its_customer_only_until_it_is_paid", func(t *testing.T) {
		t.Parallel()
		o := order(StatusAwaitingPayment)
		e := CreatedEvent(o)

		assert.Equal(t, TopicOrderCreated, e.Topic)
		assert.NotEqual(t, uuid.Nil, e.ID)
		assert.Equal(t, []uuid.UUID{*o.UserID}, e.Audience.UserIDs)
		assert.Empty(t, e.Audience.Roles)
		assert.Equal(t, map[string]string{"order_id": o.ID.String(), "status": "awaiting_payment"}, e.Data)
	})

	cases := []struct {
		name                     string
		from, to                 Status
		counterSees, kitchenSees bool
	}{
		{name: "paid_enters_the_kitchen", from: StatusAwaitingPayment, to: StatusConfirmed, counterSees: true, kitchenSees: true},
		{name: "confirmed_enters_the_kitchen", from: StatusPending, to: StatusConfirmed, counterSees: true, kitchenSees: true},
		{name: "moves_within_the_kitchen", from: StatusInProduction, to: StatusReady, counterSees: true, kitchenSees: true},
		{name: "leaves_the_kitchen_when_picked_up", from: StatusReady, to: StatusFulfilled, counterSees: true, kitchenSees: true},
		{name: "cancelled_before_the_kitchen_saw_it", from: StatusPending, to: StatusCancelled, counterSees: true},
		{name: "expired_unpaid_stays_with_the_customer", from: StatusAwaitingPayment, to: StatusExpired},
		{name: "cancelled_unpaid_stays_with_the_customer", from: StatusAwaitingPayment, to: StatusCancelled},
		{name: "cancelled_in_the_kitchen_is_told_to_it", from: StatusConfirmed, to: StatusCancelled, counterSees: true, kitchenSees: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o := order(tc.to)
			e := StatusChangedEvent(tc.from, o)

			assert.Equal(t, TopicOrderStatusChanged, e.Topic)
			assert.Equal(t, []uuid.UUID{*o.UserID}, e.Audience.UserIDs)
			assert.Equal(t, tc.counterSees, slices.Contains(e.Audience.Roles, domainuser.RoleStaff))
			assert.Equal(t, tc.counterSees, slices.Contains(e.Audience.Roles, domainuser.RoleManager), "managers count what the counter sees")
			assert.Equal(t, tc.kitchenSees, slices.Contains(e.Audience.Roles, domainuser.RoleBaker))
			assert.Equal(t, string(tc.to), e.Data["status"])
		})
	}
}

func TestStatus_VisibleToBaker(t *testing.T) {
	t.Parallel()
	visible := map[Status]bool{
		StatusPending:      false,
		StatusConfirmed:    true,
		StatusInProduction: true,
		StatusReady:        true,
		StatusFulfilled:    false,
		StatusCancelled:    false,
	}
	for status, want := range visible {
		assert.Equal(t, want, status.VisibleToBaker(), status)
	}
}

func TestSlotsChangedEvent(t *testing.T) {
	t.Parallel()
	e := SlotsChangedEvent(time.Date(2026, 7, 10, 0, 0, 0, 0, time.UTC))

	assert.Equal(t, TopicSlotsChanged, e.Topic)
	assert.True(t, e.Audience.Public, "every open checkout may be showing that day")
	assert.Empty(t, e.Audience.UserIDs)
	assert.Empty(t, e.Audience.Roles)
	assert.Equal(t, map[string]string{"date": "2026-07-10"}, e.Data, "the day only: no order, no customer")
}

func TestTicketChangedEvent(t *testing.T) {
	t.Parallel()
	customer := uuid.New()
	ticket := Ticket{ID: uuid.New(), OrderID: uuid.New(), Station: domaincategory.StationCounter, Status: TicketReady}

	order := Order{ID: ticket.OrderID, UserID: &customer}
	counter := TicketChangedEvent(order, StatusInProduction, ticket)
	assert.Equal(t, TopicTicketChanged, counter.Topic)
	assert.Equal(t, []uuid.UUID{customer}, counter.Audience.UserIDs)
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff, domainuser.RoleManager, domainuser.RoleBaker}, counter.Audience.Roles,
		"the kitchen sees where the counter's part of an order it works on stands")
	assert.Equal(t, map[string]string{
		"ticket_id": ticket.ID.String(), "order_id": ticket.OrderID.String(), "station": "counter", "status": "ready",
	}, counter.Data)

	unseen := TicketChangedEvent(order, StatusPending, Ticket{Station: domaincategory.StationKitchen})
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff, domainuser.RoleManager}, unseen.Audience.Roles,
		"the kitchen hears nothing of an order it may not see yet")

	unpaid := TicketChangedEvent(order, StatusAwaitingPayment, Ticket{Station: domaincategory.StationCounter})
	assert.Empty(t, unpaid.Audience.Roles, "an order never paid is the customer's alone")
}

func TestRescheduledAndRefundedEvents(t *testing.T) {
	t.Parallel()
	customer := uuid.New()
	o := Order{ID: uuid.New(), UserID: &customer, Status: StatusConfirmed}

	moved := RescheduledEvent(o)
	assert.Equal(t, TopicOrderRescheduled, moved.Topic)
	assert.Equal(t, []uuid.UUID{*o.UserID}, moved.Audience.UserIDs)
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff, domainuser.RoleManager, domainuser.RoleBaker}, moved.Audience.Roles,
		"the kitchen bakes to the pickup time")

	unpaid := RescheduledEvent(Order{ID: uuid.New(), UserID: o.UserID, Status: StatusAwaitingPayment})
	assert.Empty(t, unpaid.Audience.Roles, "an order not paid yet is the customer's alone")

	o.Status = StatusCancelled
	refunded := RefundedEvent(o)
	assert.Equal(t, TopicOrderRefunded, refunded.Topic)
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff, domainuser.RoleManager}, refunded.Audience.Roles)
	assert.Equal(t, map[string]string{"order_id": o.ID.String()}, refunded.Data)
}

func TestSeenByStaff(t *testing.T) {
	t.Parallel()
	accepted := []StatusEvent{{To: StatusAwaitingPayment}, {To: StatusConfirmed}, {To: StatusCancelled}}
	unpaid := []StatusEvent{{To: StatusAwaitingPayment}, {To: StatusCancelled}}
	assert.True(t, SeenByStaff(StatusCancelled, accepted), "cancelled after it was paid")
	assert.False(t, SeenByStaff(StatusCancelled, unpaid), "cancelled before it was paid")
	assert.True(t, SeenByStaff(StatusReady, nil))
	assert.False(t, SeenByStaff(StatusAwaitingPayment, nil))
}

func TestStatus_VisibleToStaff(t *testing.T) {
	t.Parallel()
	for _, status := range []Status{StatusAwaitingPayment, StatusExpired} {
		assert.False(t, status.VisibleToStaff(), status)
	}
	for _, status := range []Status{StatusPending, StatusConfirmed, StatusInProduction, StatusReady, StatusFulfilled, StatusCancelled} {
		assert.True(t, status.VisibleToStaff(), status)
	}
}

// An order staff took starts confirmed, so the counter and the kitchen hear of
// it at once; a guest's has no account to tell.
func TestCreatedEvent_TakenByStaff(t *testing.T) {
	t.Parallel()
	guest := Order{ID: uuid.New(), Guest: &Guest{Name: "Lan", Phone: "+84901234567"}, Channel: ChannelCounter, Status: StatusConfirmed}

	e := CreatedEvent(guest)

	assert.Empty(t, e.Audience.UserIDs)
	assert.Equal(t, []domainuser.Role{domainuser.RoleStaff, domainuser.RoleManager, domainuser.RoleBaker}, e.Audience.Roles)
}
