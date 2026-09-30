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
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// transitionOrders implements only what a status move uses; any other call panics.
type transitionOrders struct {
	port.OrderRepository
	updated    *domainorder.Order
	history    []port.AddOrderStatusEventParams
	err        error
	historyErr error
}

func (f *transitionOrders) UpdateStatus(_ context.Context, params port.UpdateOrderStatusParams) (*domainorder.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	pickupAt := time.Date(2026, 7, 10, 3, 0, 0, 0, time.UTC)
	f.updated = &domainorder.Order{ID: params.OrderID, UserID: uuid.New(), Status: params.ToStatus, PickupAt: &pickupAt}
	return f.updated, nil
}

func (f *transitionOrders) AddStatusEvent(_ context.Context, params port.AddOrderStatusEventParams) error {
	if f.historyErr != nil {
		return f.historyErr
	}
	f.history = append(f.history, params)
	return nil
}

// cancelledTickets cancels the tickets of an order; any other call panics.
type cancelledTickets struct {
	port.TicketRepository
	cancelled []domainorder.Ticket
	orderID   uuid.UUID
}

func (f *cancelledTickets) CancelForOrder(_ context.Context, orderID uuid.UUID) ([]domainorder.Ticket, error) {
	f.orderID = orderID
	return f.cancelled, nil
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
	baker := uuid.New()

	t.Run("records_the_move_as_an_event_in_the_same_transaction", func(t *testing.T) {
		t.Parallel()
		orders, outbox := &transitionOrders{}, &recordingOutbox{}
		transitions := orderTransitions{tx: inlineTx{}, orders: orders, events: outbox}

		got, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params)
		require.NoError(t, err)
		assert.Equal(t, domainorder.StatusInProduction, got.Status)
		require.Len(t, outbox.added, 1)
		assert.Equal(t, domainorder.TopicOrderStatusChanged, outbox.added[0].Topic)
		assert.Equal(t, params.OrderID.String(), outbox.added[0].Data["order_id"])
		assert.Equal(t, string(domainorder.StatusInProduction), outbox.added[0].Data["status"])
		require.Len(t, orders.history, 1)
		assert.Equal(t, port.AddOrderStatusEventParams{
			OrderID: params.OrderID,
			From:    &params.FromStatus,
			To:      domainorder.StatusInProduction,
			Actor:   &port.OrderActor{ID: baker, Role: domainuser.RoleBaker},
		}, orders.history[0], "the history names the move and who made it")
	})

	t.Run("a_cancelled_order_cancels_its_tickets_and_frees_its_slot", func(t *testing.T) {
		t.Parallel()
		orderID := uuid.New()
		kitchenTicket := domainorder.Ticket{ID: uuid.New(), OrderID: orderID, Station: domaincategory.StationKitchen, Status: domainorder.TicketCancelled}
		orders, outbox := &transitionOrders{}, &recordingOutbox{}
		tickets := &cancelledTickets{cancelled: []domainorder.Ticket{kitchenTicket}}
		transitions := orderTransitions{tx: inlineTx{}, orders: orders, tickets: tickets, events: outbox}

		_, err := transitions.apply(context.Background(), &port.OrderActor{ID: uuid.New(), Role: domainuser.RoleStaff}, port.UpdateOrderStatusParams{
			OrderID: orderID, FromStatus: domainorder.StatusInProduction, ToStatus: domainorder.StatusCancelled,
		})

		require.NoError(t, err)
		assert.Equal(t, orderID, tickets.orderID, "the tickets of the cancelled order")
		require.Len(t, outbox.added, 3)
		assert.Equal(t, domainorder.TopicOrderStatusChanged, outbox.added[0].Topic)
		assert.Equal(t, domainorder.TopicTicketChanged, outbox.added[1].Topic)
		assert.Contains(t, outbox.added[1].Audience.Roles, domainuser.RoleBaker, "the kitchen hears its ticket is off")
		assert.Equal(t, domainorder.TopicSlotsChanged, outbox.added[2].Topic)
		assert.Equal(t, "2026-07-10", outbox.added[2].Data["date"], "the bakery day of the freed slot")
	})

	t.Run("reports_a_move_someone_else_made_first_as_invalid", func(t *testing.T) {
		t.Parallel()
		outbox := &recordingOutbox{}
		transitions := orderTransitions{
			tx:     inlineTx{},
			orders: &transitionOrders{err: apperrors.ErrNotFound},
			events: outbox,
		}

		_, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params)
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

		got, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params)
		assert.ErrorIs(t, err, errOutbox, "the transaction rolls back rather than commit a silent change")
		assert.Nil(t, got)
	})

	t.Run("fails_the_move_when_its_history_cannot_be_recorded", func(t *testing.T) {
		t.Parallel()
		errHistory := errors.New("history unavailable")
		outbox := &recordingOutbox{}
		transitions := orderTransitions{
			tx:     inlineTx{},
			orders: &transitionOrders{historyErr: errHistory},
			events: outbox,
		}

		got, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params)
		assert.ErrorIs(t, err, errHistory)
		assert.Nil(t, got)
		assert.Empty(t, outbox.added, "no notice for a move the transaction rolls back")
	})
}
