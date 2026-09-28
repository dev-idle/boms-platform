package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// transitionOrders implements only what a status move uses; any other call panics.
type transitionOrders struct {
	port.OrderRepository
	updated *domainorder.Order
	err     error
}

func (f *transitionOrders) UpdateStatus(_ context.Context, params port.UpdateOrderStatusParams) (*domainorder.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	f.updated = &domainorder.Order{ID: params.OrderID, UserID: uuid.New(), Status: params.ToStatus}
	return f.updated, nil
}

type recordingOutbox struct {
	port.EventOutbox
	added []domainevent.Event
	err   error
}

func (f *recordingOutbox) Add(_ context.Context, e domainevent.Event) error {
	if f.err != nil {
		return f.err
	}
	f.added = append(f.added, e)
	return nil
}

// inlineTx runs fn directly: the fakes above hold no state a rollback would undo.
type inlineTx struct{}

func (inlineTx) WithTx(ctx context.Context, fn func(txCtx context.Context) error) error {
	return fn(ctx)
}

func TestOrderTransitions_Apply(t *testing.T) {
	t.Parallel()

	params := port.UpdateOrderStatusParams{
		OrderID:    uuid.New(),
		FromStatus: domainorder.StatusConfirmed,
		ToStatus:   domainorder.StatusInProduction,
	}

	t.Run("records_the_move_as_an_event_in_the_same_transaction", func(t *testing.T) {
		t.Parallel()
		orders, outbox := &transitionOrders{}, &recordingOutbox{}
		transitions := orderTransitions{tx: inlineTx{}, orders: orders, events: outbox}

		got, err := transitions.apply(context.Background(), params)
		require.NoError(t, err)
		assert.Equal(t, domainorder.StatusInProduction, got.Status)
		require.Len(t, outbox.added, 1)
		assert.Equal(t, domainorder.TopicOrderStatusChanged, outbox.added[0].Topic)
		assert.Equal(t, params.OrderID.String(), outbox.added[0].Data["order_id"])
		assert.Equal(t, string(domainorder.StatusInProduction), outbox.added[0].Data["status"])
	})

	t.Run("reports_a_move_someone_else_made_first_as_invalid", func(t *testing.T) {
		t.Parallel()
		outbox := &recordingOutbox{}
		transitions := orderTransitions{
			tx:     inlineTx{},
			orders: &transitionOrders{err: apperrors.ErrNotFound},
			events: outbox,
		}

		_, err := transitions.apply(context.Background(), params)
		assert.ErrorIs(t, err, domainorder.ErrInvalidStatusTransition)
		assert.Empty(t, outbox.added, "no event for a move that did not happen")
	})

	t.Run("fails_the_move_when_its_event_cannot_be_recorded", func(t *testing.T) {
		t.Parallel()
		errOutbox := errors.New("outbox unavailable")
		transitions := orderTransitions{
			tx:     inlineTx{},
			orders: &transitionOrders{},
			events: &recordingOutbox{err: errOutbox},
		}

		got, err := transitions.apply(context.Background(), params)
		assert.ErrorIs(t, err, errOutbox, "the transaction rolls back rather than commit a silent change")
		assert.Nil(t, got)
	})
}
