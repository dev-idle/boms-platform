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
	updated   *domainorder.Order
	discount  *uuid.UUID
	history   []port.AddOrderStatusEventParams
	incidents []port.AddOrderIncidentParams
	// flags makes every payment problem flag its customer.
	flags      bool
	err        error
	historyErr error
}

func (f *transitionOrders) UpdateStatus(_ context.Context, params port.UpdateOrderStatusParams) (*domainorder.Order, error) {
	if f.err != nil {
		return nil, f.err
	}
	pickupAt := time.Date(2026, 7, 10, 3, 0, 0, 0, time.UTC)
	customer := uuid.New()
	f.updated = &domainorder.Order{
		ID: params.OrderID, UserID: &customer, Status: params.ToStatus, PickupAt: &pickupAt, DiscountCodeID: f.discount,
	}
	return f.updated, nil
}

func (f *transitionOrders) AddStatusEvent(_ context.Context, params port.AddOrderStatusEventParams) error {
	if f.historyErr != nil {
		return f.historyErr
	}
	f.history = append(f.history, params)
	return nil
}

func (f *transitionOrders) AddIncident(_ context.Context, params port.AddOrderIncidentParams) (*domainorder.Incident, error) {
	f.incidents = append(f.incidents, params)
	return &domainorder.Incident{ID: uuid.New(), OrderID: params.OrderID, Type: params.Type, Note: params.Note}, nil
}

func (f *transitionOrders) FlagPaymentAnomaly(_ context.Context, orderID uuid.UUID, _ time.Duration, _ int32) (*domainorder.Incident, error) {
	if !f.flags {
		return nil, nil
	}
	return &domainorder.Incident{ID: uuid.New(), OrderID: orderID, Type: domainorder.IncidentPaymentAnomaly}, nil
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

// refundRequests records the orders whose payment is asked back; any other
// call panics.
type refundRequests struct {
	port.PaymentRepository
	orders []uuid.UUID
}

func (f *refundRequests) RequestRefund(_ context.Context, orderID uuid.UUID) error {
	f.orders = append(f.orders, orderID)
	return nil
}

// codeReleases records the discount uses given back; any other call panics.
type codeReleases struct {
	port.DiscountCodeRepository
	released []uuid.UUID
}

func (f *codeReleases) ReleaseUse(_ context.Context, id uuid.UUID) error {
	f.released = append(f.released, id)
	return nil
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

		got, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params, "")
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

	t.Run("a_cancelled_order_cancels_its_tickets_frees_its_slot_and_is_refunded", func(t *testing.T) {
		t.Parallel()
		orderID := uuid.New()
		kitchenTicket := domainorder.Ticket{ID: uuid.New(), OrderID: orderID, Station: domaincategory.StationKitchen, Status: domainorder.TicketCancelled}
		code := uuid.New()
		orders, outbox := &transitionOrders{discount: &code}, &recordingOutbox{}
		tickets := &cancelledTickets{cancelled: []domainorder.Ticket{kitchenTicket}}
		refunds, codes := &refundRequests{}, &codeReleases{}
		transitions := orderTransitions{
			tx: inlineTx{}, orders: orders, tickets: tickets, discounts: codes, payments: refunds, events: outbox,
		}

		staff := &port.OrderActor{ID: uuid.New(), Role: domainuser.RoleStaff}
		_, err := transitions.apply(context.Background(), staff, port.UpdateOrderStatusParams{
			OrderID: orderID, FromStatus: domainorder.StatusInProduction, ToStatus: domainorder.StatusCancelled,
		}, "The oven broke down")

		require.NoError(t, err)
		assert.Equal(t, "The oven broke down", orders.history[0].Reason, "the customer is told why")
		assert.Equal(t, []uuid.UUID{orderID}, refunds.orders, "a paid order is refunded in full")
		assert.Empty(t, codes.released, "a code used on an order already being made is not given back")
		assert.Equal(t, orderID, tickets.orderID, "the tickets of the cancelled order")
		assert.Equal(t, []port.AddOrderIncidentParams{{
			OrderID: orderID, Type: domainorder.IncidentBakeryCancelled, Actor: staff,
		}}, orders.incidents, "the bakery cancelling an order it accepted is an incident; its reason stays in the history")
		require.Len(t, outbox.added, 4)
		assert.Equal(t, domainorder.TopicOrderStatusChanged, outbox.added[0].Topic)
		assert.Equal(t, domainorder.TopicIncidentRecorded, outbox.added[1].Topic)
		assert.Equal(t, []domainuser.Role{domainuser.RoleManager}, outbox.added[1].Audience.Roles)
		assert.Equal(t, domainorder.TopicTicketChanged, outbox.added[2].Topic)
		assert.Contains(t, outbox.added[2].Audience.Roles, domainuser.RoleBaker, "the kitchen hears its ticket is off")
		assert.Equal(t, domainorder.TopicSlotsChanged, outbox.added[3].Topic)
		assert.Equal(t, "2026-07-10", outbox.added[3].Data["date"], "the bakery day of the freed slot")
	})

	t.Run("an_order_dropped_before_it_was_made_gives_its_discount_use_back", func(t *testing.T) {
		t.Parallel()
		code := uuid.New()
		orders := &transitionOrders{discount: &code}
		refunds, codes := &refundRequests{}, &codeReleases{}
		transitions := orderTransitions{
			tx: inlineTx{}, orders: orders, tickets: &cancelledTickets{},
			discounts: codes, payments: refunds, events: &recordingOutbox{},
		}

		_, err := transitions.apply(context.Background(), &port.OrderActor{ID: uuid.New(), Role: domainuser.RoleCustomer}, port.UpdateOrderStatusParams{
			OrderID: uuid.New(), FromStatus: domainorder.StatusConfirmed, ToStatus: domainorder.StatusCancelled,
		}, "")

		require.NoError(t, err)
		assert.Equal(t, []uuid.UUID{code}, codes.released)
		assert.Len(t, refunds.orders, 1)
		assert.Empty(t, orders.incidents, "a customer changing their mind is no incident")
	})

	t.Run("a_missed_pickup_keeps_its_slot_payment_and_discount", func(t *testing.T) {
		t.Parallel()
		code := uuid.New()
		orders, outbox := &transitionOrders{discount: &code}, &recordingOutbox{}
		transitions := orderTransitions{tx: inlineTx{}, orders: orders, events: outbox}

		_, err := transitions.apply(context.Background(), nil, port.UpdateOrderStatusParams{
			OrderID: uuid.New(), FromStatus: domainorder.StatusReady, ToStatus: domainorder.StatusNoShow,
		}, "")

		require.NoError(t, err, "no ticket, discount or payment repository is touched")
		require.Len(t, outbox.added, 2)
		assert.Equal(t, "no_show", outbox.added[0].Data["status"])
		require.Len(t, orders.incidents, 1)
		assert.Equal(t, domainorder.IncidentNoShow, orders.incidents[0].Type)
		assert.Nil(t, orders.incidents[0].Actor, "the system records it")
	})

	t.Run("an_expiry_that_is_one_too_many_flags_the_customer", func(t *testing.T) {
		t.Parallel()
		orders, outbox := &transitionOrders{flags: true}, &recordingOutbox{}
		transitions := orderTransitions{tx: inlineTx{}, orders: orders, tickets: &cancelledTickets{}, events: outbox}

		_, err := transitions.apply(context.Background(), nil, port.UpdateOrderStatusParams{
			OrderID: uuid.New(), FromStatus: domainorder.StatusAwaitingPayment, ToStatus: domainorder.StatusExpired,
		}, "")

		require.NoError(t, err)
		require.Len(t, orders.incidents, 1)
		assert.Equal(t, domainorder.IncidentPaymentExpired, orders.incidents[0].Type)
		var recorded []string
		for _, event := range outbox.added {
			if event.Topic == domainorder.TopicIncidentRecorded {
				recorded = append(recorded, event.Data["type"])
			}
		}
		assert.Equal(t, []string{"payment_expired", "payment_anomaly"}, recorded, "managers hear of the expiry and of the flag")
	})

	t.Run("reports_a_move_someone_else_made_first_as_invalid", func(t *testing.T) {
		t.Parallel()
		outbox := &recordingOutbox{}
		transitions := orderTransitions{
			tx:     inlineTx{},
			orders: &transitionOrders{err: apperrors.ErrNotFound},
			events: outbox,
		}

		_, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params, "")
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

		got, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params, "")
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

		got, err := transitions.apply(context.Background(), &port.OrderActor{ID: baker, Role: domainuser.RoleBaker}, params, "")
		assert.ErrorIs(t, err, errHistory)
		assert.Nil(t, got)
		assert.Empty(t, outbox.added, "no notice for a move the transaction rolls back")
	})
}
