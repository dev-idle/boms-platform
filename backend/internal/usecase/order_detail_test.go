package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/ctxmeta"
)

// detailOrders answers the reads a detail page makes of the order; any other
// call panics. A read told to wait holds until its context is cancelled.
type detailOrders struct {
	port.OrderRepository
	items       []domainorder.Item
	timeline    []domainorder.StatusEvent
	itemsErr    error
	historyWait bool
}

func (f *detailOrders) ListItemsByOrderID(context.Context, uuid.UUID) ([]domainorder.Item, error) {
	if f.itemsErr != nil {
		return nil, f.itemsErr
	}
	return f.items, nil
}

func (f *detailOrders) ListStatusEvents(ctx context.Context, _ uuid.UUID) ([]domainorder.StatusEvent, error) {
	if f.historyWait {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return f.timeline, nil
}

// detailTickets answers the order's tickets; any other call panics.
type detailTickets struct {
	port.TicketRepository
	tickets []domainorder.Ticket
}

func (f *detailTickets) ListByOrder(context.Context, uuid.UUID) ([]domainorder.Ticket, error) {
	return f.tickets, nil
}

func TestReadOrderDetail(t *testing.T) {
	t.Parallel()

	t.Run("returns_the_lines_the_history_and_the_tickets", func(t *testing.T) {
		t.Parallel()
		orders := &detailOrders{
			items:    []domainorder.Item{{Name: "Matcha cake", Quantity: 1}},
			timeline: []domainorder.StatusEvent{{To: domainorder.StatusPending, ActorRole: domainuser.RoleCustomer, At: time.Now()}},
		}
		tickets := &detailTickets{tickets: []domainorder.Ticket{{Station: domaincategory.StationKitchen, Status: domainorder.TicketQueued}}}

		parts, err := readOrderDetail(t.Context(), orders, tickets, uuid.New())

		require.NoError(t, err)
		assert.Equal(t, orders.items, parts.items)
		assert.Equal(t, orders.timeline, parts.timeline)
		assert.Equal(t, tickets.tickets, parts.tickets)
	})

	t.Run("a_failed_read_fails_the_detail_and_stops_the_others", func(t *testing.T) {
		t.Parallel()
		errItems := errors.New("items unavailable")
		orders := &detailOrders{itemsErr: errItems, historyWait: true}

		parts, err := readOrderDetail(t.Context(), orders, &detailTickets{}, uuid.New())

		require.ErrorIs(t, err, errItems, "the waiting history read was cancelled, or this would hang")
		assert.Equal(t, orderDetailParts{}, parts, "no half detail beside an error")
	})

	t.Run("refuses_to_run_inside_a_transaction", func(t *testing.T) {
		t.Parallel()
		// The nil repositories panic if any read runs.
		_, err := readOrderDetail(ctxmeta.WithinTransaction(t.Context()), nil, nil, uuid.New())

		require.Error(t, err)
	})
}
