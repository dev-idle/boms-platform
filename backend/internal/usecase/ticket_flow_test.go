package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

// readyTickets lists an order's tickets, all ready; any other call panics.
type readyTickets struct{ port.TicketRepository }

func (readyTickets) ListByOrder(_ context.Context, orderID uuid.UUID) ([]domainorder.Ticket, error) {
	return []domainorder.Ticket{{ID: uuid.New(), OrderID: orderID, Status: domainorder.TicketReady}}, nil
}

// pickupOrders moves an order as transitionOrders does, to a pickup at pickupAt.
type pickupOrders struct {
	transitionOrders
	pickupAt time.Time
}

func (f *pickupOrders) UpdateStatus(ctx context.Context, params port.UpdateOrderStatusParams) (*domainorder.Order, error) {
	moved, err := f.transitionOrders.UpdateStatus(ctx, params)
	if err != nil {
		return nil, err
	}
	moved.PickupAt = &f.pickupAt
	return moved, nil
}

func TestTicketFlow_FollowTickets(t *testing.T) {
	t.Parallel()
	baker := ticketActor{id: uuid.New(), role: domainuser.RoleBaker}
	finish := func(t *testing.T, pickupAt time.Time) []port.AddOrderIncidentParams {
		t.Helper()
		orders := &pickupOrders{pickupAt: pickupAt}
		flow := ticketFlow{tx: inlineTx{}, orders: orders, tickets: readyTickets{}, events: &recordingOutbox{}}

		moved, err := flow.followTickets(context.Background(), baker, domainorder.Order{
			ID: uuid.New(), Status: domainorder.StatusInProduction,
		})

		require.NoError(t, err)
		assert.Equal(t, domainorder.StatusReady, moved.Status, "every ticket is ready")
		return orders.incidents
	}

	t.Run("an_order_ready_after_its_pickup_time_is_an_incident", func(t *testing.T) {
		t.Parallel()
		incidents := finish(t, time.Now().Add(-time.Minute))

		require.Len(t, incidents, 1)
		assert.Equal(t, domainorder.IncidentReadyLate, incidents[0].Type)
		assert.Nil(t, incidents[0].Actor, "the system records it, not the last baker")
	})

	t.Run("an_order_ready_in_time_is_none", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, finish(t, time.Now().Add(time.Hour)))
	})
}
