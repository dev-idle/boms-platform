package order

import (
	"slices"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	domainuser "github.com/boms/backend/internal/domain/user"
)

func TestOrderEvents(t *testing.T) {
	t.Parallel()

	order := func(status Status) Order {
		return Order{ID: uuid.New(), UserID: uuid.New(), Status: status}
	}

	t.Run("a_new_order_is_told_to_its_customer_and_the_counter_only", func(t *testing.T) {
		t.Parallel()
		o := order(StatusPending)
		e := CreatedEvent(o)

		assert.Equal(t, TopicOrderCreated, e.Topic)
		assert.NotEqual(t, uuid.Nil, e.ID)
		assert.Equal(t, []uuid.UUID{o.UserID}, e.Audience.UserIDs)
		assert.Equal(t, []domainuser.Role{domainuser.RoleStaff}, e.Audience.Roles)
		assert.Equal(t, map[string]string{"order_id": o.ID.String(), "status": "pending"}, e.Data)
	})

	cases := []struct {
		name        string
		from, to    Status
		kitchenSees bool
	}{
		{name: "confirmed_enters_the_kitchen", from: StatusPending, to: StatusConfirmed, kitchenSees: true},
		{name: "moves_within_the_kitchen", from: StatusInProduction, to: StatusReady, kitchenSees: true},
		{name: "leaves_the_kitchen_when_picked_up", from: StatusReady, to: StatusFulfilled, kitchenSees: true},
		{name: "cancelled_before_the_kitchen_saw_it", from: StatusPending, to: StatusCancelled, kitchenSees: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			o := order(tc.to)
			e := StatusChangedEvent(tc.from, o)

			assert.Equal(t, TopicOrderStatusChanged, e.Topic)
			assert.Equal(t, []uuid.UUID{o.UserID}, e.Audience.UserIDs)
			assert.Contains(t, e.Audience.Roles, domainuser.RoleStaff)
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
