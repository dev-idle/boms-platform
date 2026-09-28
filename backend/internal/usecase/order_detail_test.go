package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/shared/ctxmeta"
)

// detailOrders answers the two reads a detail page makes; any other call panics.
// A read told to wait holds until its context is cancelled.
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

func TestOrderLinesAndTimeline(t *testing.T) {
	t.Parallel()

	t.Run("returns_the_lines_and_the_history", func(t *testing.T) {
		t.Parallel()
		orders := &detailOrders{
			items:    []domainorder.Item{{Name: "Matcha cake", Quantity: 1}},
			timeline: []domainorder.StatusEvent{{To: domainorder.StatusPending, ActorRole: domainuser.RoleCustomer, At: time.Now()}},
		}

		items, timeline, err := orderLinesAndTimeline(t.Context(), orders, uuid.New())

		require.NoError(t, err)
		assert.Equal(t, orders.items, items)
		assert.Equal(t, orders.timeline, timeline)
	})

	t.Run("a_failed_read_fails_the_detail_and_stops_the_other", func(t *testing.T) {
		t.Parallel()
		errItems := errors.New("items unavailable")
		orders := &detailOrders{itemsErr: errItems, historyWait: true}

		items, timeline, err := orderLinesAndTimeline(t.Context(), orders, uuid.New())

		require.ErrorIs(t, err, errItems, "the waiting history read was cancelled, or this would hang")
		assert.Nil(t, items, "no half detail beside an error")
		assert.Nil(t, timeline)
	})

	t.Run("refuses_to_run_inside_a_transaction", func(t *testing.T) {
		t.Parallel()
		// The nil repository panics if either read runs.
		_, _, err := orderLinesAndTimeline(ctxmeta.WithinTransaction(t.Context()), nil, uuid.New())

		require.Error(t, err)
	})
}
