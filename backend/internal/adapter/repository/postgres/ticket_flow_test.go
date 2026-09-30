package postgres_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

// ticketFixture is the checkout fixture with the counter and the kitchen at work.
type ticketFixture struct {
	*checkoutFixture
	tickets *postgresadapter.TicketRepository
	staff   *usecase.StaffOrderUsecase
	counter *usecase.StaffTicketUsecase
	kitchen *usecase.BakerTicketUsecase
	baker   uuid.UUID
	clerk   uuid.UUID
}

func newTicketFixture(t *testing.T) *ticketFixture {
	t.Helper()
	f := newCheckoutFixture(t, 8)
	tickets := postgresadapter.NewTicketRepository(f.pool)
	return &ticketFixture{
		baker:           f.newWorker(t, domainuser.RoleBaker),
		clerk:           f.newWorker(t, domainuser.RoleStaff),
		checkoutFixture: f,
		tickets:         tickets,
		staff:           usecase.NewStaffOrderUsecase(f.orders, tickets, f.pool, f.outbox, nil, nil),
		counter:         usecase.NewStaffTicketUsecase(f.orders, tickets, f.pool, f.outbox, nil, nil),
		kitchen:         usecase.NewBakerTicketUsecase(f.orders, tickets, f.pool, f.outbox, nil, nil),
	}
}

// placeConfirmed checks out a cart of these lines tomorrow and has the counter
// accept the order.
func (f *ticketFixture) placeConfirmed(t *testing.T, products, combos []uuid.UUID, hour int) *dto.OrderResponse {
	t.Helper()
	order, err := f.orderUC.Checkout(context.Background(), f.newCustomer(t, products, combos), acceptingTerms(tomorrowAt(hour, 0)))
	require.NoError(t, err)
	_, err = f.staff.PatchStatus(context.Background(), f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), domainorder.StatusConfirmed)
	require.NoError(t, err)
	return order
}

// newWorker adds a member of the bakery's staff, who is recorded on every
// move of an order they cause.
func (f *checkoutFixture) newWorker(t *testing.T, role domainuser.Role) uuid.UUID {
	t.Helper()
	f.customer++
	worker, err := f.users.Create(context.Background(), port.CreateUserParams{
		Email:        fmt.Sprintf("worker-%d@example.com", f.customer),
		PasswordHash: testPasswordHashFixture,
		Role:         role,
	})
	require.NoError(t, err)
	return worker.ID
}

func (f *ticketFixture) ticketAt(t *testing.T, orderID string, station domaincategory.Station) domainorder.Ticket {
	t.Helper()
	tickets, err := f.tickets.ListByOrder(context.Background(), uuid.MustParse(orderID))
	require.NoError(t, err)
	for _, ticket := range tickets {
		if ticket.Station == station {
			return ticket
		}
	}
	t.Fatalf("order %s has no %s ticket", orderID, station)
	return domainorder.Ticket{}
}

func (f *ticketFixture) orderStatus(t *testing.T, orderID string) domainorder.Status {
	t.Helper()
	row, err := f.orders.StaffGetByID(context.Background(), uuid.MustParse(orderID))
	require.NoError(t, err)
	return row.Order.Status
}

// orderNotices lists what the outbox holds about an order, oldest first, as
// topic:status.
func (f *ticketFixture) orderNotices(t *testing.T, orderID string) []string {
	t.Helper()
	var events []domainevent.Event
	require.NoError(t, f.pool.WithTx(context.Background(), func(txCtx context.Context) error {
		var err error
		events, err = f.outbox.ClaimUnpublished(txCtx, 0, 1000)
		return err
	}))
	var notices []string
	for _, e := range events {
		if e.Data["order_id"] == orderID {
			notices = append(notices, string(e.Topic)+":"+e.Data["status"])
		}
	}
	return notices
}

func TestTickets_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	baker, clerk := f.baker, f.clerk

	t.Run("checkout_splits_the_order_by_station", func(t *testing.T) {
		cases := map[string]struct {
			products, combos []uuid.UUID
			want             map[domaincategory.Station][]string
		}{
			"kitchen_only": {products: []uuid.UUID{f.cake}, want: map[domaincategory.Station][]string{
				domaincategory.StationKitchen: {"Matcha cake x1"},
			}},
			"counter_only": {products: []uuid.UUID{f.pastry}, want: map[domaincategory.Station][]string{
				domaincategory.StationCounter: {"Croissant x1"},
			}},
			"mixed_and_a_combo_spanning_both": {products: []uuid.UUID{f.pastry}, combos: []uuid.UUID{f.combo}, want: map[domaincategory.Station][]string{
				domaincategory.StationKitchen: {"Matcha cake x1"},
				domaincategory.StationCounter: {"Croissant x1", "Croissant x1"},
			}},
		}
		hour := 8
		for name, tc := range cases {
			order, err := f.orderUC.Checkout(ctx, f.newCustomer(t, tc.products, tc.combos), acceptingTerms(tomorrowAt(hour, 0)))
			require.NoError(t, err, name)
			hour++
			tickets, err := f.tickets.ListByOrder(ctx, uuid.MustParse(order.ID))
			require.NoError(t, err)
			got := map[domaincategory.Station][]string{}
			for _, ticket := range tickets {
				assert.Equal(t, domainorder.TicketQueued, ticket.Status, name)
				for _, item := range ticket.Items {
					got[ticket.Station] = append(got[ticket.Station], fmt.Sprintf("%s x%d", item.Name, item.Quantity))
				}
			}
			assert.ElementsMatch(t, keys(tc.want), keys(got), name)
			for station, items := range tc.want {
				assert.ElementsMatch(t, items, got[station], "%s: %s", name, station)
			}
			assert.Len(t, order.Tickets, len(tc.want), "%s: the customer sees each station", name)
		}
	})

	t.Run("tickets_wait_for_the_counter_to_accept_the_order", func(t *testing.T) {
		order, err := f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.cake}, nil), acceptingTerms(tomorrowAt(12, 0)))
		require.NoError(t, err)
		ticket := f.ticketAt(t, order.ID, domaincategory.StationKitchen)

		_, err = f.kitchen.Get(ctx, ticket.ID)
		require.ErrorIs(t, err, domainorder.ErrTicketNotFound, "the kitchen does not see an order nobody accepted")
		_, err = f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, ticket.ID, domainorder.TicketInProgress)
		require.ErrorIs(t, err, domainorder.ErrTicketOrderNotActive)
	})

	t.Run("a_station_moves_only_its_own_tickets_one_step_at_a_time", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 13)
		ticket := f.ticketAt(t, order.ID, domaincategory.StationCounter)

		_, err := f.kitchen.Get(ctx, ticket.ID)
		require.ErrorIs(t, err, domainorder.ErrTicketNotFound, "the kitchen does not see counter tickets")
		_, err = f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, ticket.ID, domainorder.TicketInProgress)
		require.ErrorIs(t, err, domainorder.ErrTicketNotFound, "nor moves them")
		_, err = f.counter.PatchStatus(ctx, clerk, domainuser.RoleStaff, ticket.ID, domainorder.TicketReady)
		require.ErrorIs(t, err, domainorder.ErrInvalidTicketTransition, "a ticket starts before it is ready")

		unknown := uuid.New()
		_, err = f.kitchen.Get(ctx, unknown)
		require.ErrorIs(t, err, domainorder.ErrTicketNotFound)
		_, err = f.counter.PatchStatus(ctx, clerk, domainuser.RoleStaff, unknown, domainorder.TicketInProgress)
		require.ErrorIs(t, err, domainorder.ErrTicketNotFound)
		_, err = f.counter.Move(ctx, clerk, domainuser.RoleStaff, unknown, domaincategory.StationKitchen)
		require.ErrorIs(t, err, domainorder.ErrTicketNotFound)
	})

	t.Run("a_queue_filters_only_by_the_statuses_it_holds", func(t *testing.T) {
		for _, filter := range []string{"cancelled", "baking"} {
			_, _, _, _, err := f.kitchen.List(ctx, 1, 20, filter)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr, filter)
			assert.Equal(t, apperrors.ErrValidation.Code, appErr.Code, filter)
		}
	})

	t.Run("the_order_is_ready_when_its_last_ticket_is", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.cake, f.pastry}, nil, 14)
		cake := f.ticketAt(t, order.ID, domaincategory.StationKitchen)
		pastry := f.ticketAt(t, order.ID, domaincategory.StationCounter)

		change, err := f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, cake.ID, domainorder.TicketInProgress)
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusInProduction), change.OrderStatus, "the first start puts the order in production")
		_, err = f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, cake.ID, domainorder.TicketReady)
		require.NoError(t, err)
		assert.Equal(t, domainorder.StatusInProduction, f.orderStatus(t, order.ID), "the counter still has its part to do")
		seen, err := f.kitchen.Get(ctx, cake.ID)
		require.NoError(t, err)
		require.Len(t, seen.OrderTickets, 2, "the kitchen sees where the whole order stands")
		assert.Equal(t, []dto.TicketItemResponse{{Name: "Matcha cake", Quantity: 1}}, seen.Items)

		_, err = f.counter.PatchStatus(ctx, clerk, domainuser.RoleStaff, pastry.ID, domainorder.TicketInProgress)
		require.NoError(t, err)
		change, err = f.counter.PatchStatus(ctx, clerk, domainuser.RoleStaff, pastry.ID, domainorder.TicketReady)
		require.NoError(t, err)
		assert.Equal(t, string(domainorder.StatusReady), change.OrderStatus)

		detail, err := f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		var statuses []string
		for _, entry := range detail.Timeline {
			statuses = append(statuses, entry.Status+":"+entry.ActorRole)
		}
		assert.Equal(t, []string{"pending:customer", "confirmed:staff", "in_production:baker", "ready:staff"}, statuses,
			"each derived move is recorded as made by whoever caused it")
		assert.Equal(t, []string{
			"order.created:pending", "order.status_changed:confirmed",
			"order.status_changed:in_production", "ticket.changed:in_progress",
			"ticket.changed:ready",
			"ticket.changed:in_progress", "order.status_changed:ready", "ticket.changed:ready",
		}, f.orderNotices(t, order.ID), "every ticket move is announced with the order status it derives")

		_, err = f.staff.PatchStatus(ctx, clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), domainorder.StatusFulfilled)
		require.NoError(t, err)
		_, err = f.counter.Move(ctx, clerk, domainuser.RoleStaff, pastry.ID, domaincategory.StationKitchen)
		require.ErrorIs(t, err, domainorder.ErrTicketOrderNotActive, "a collected order has nothing left to move")
	})

	t.Run("the_last_two_tickets_finishing_at_once_make_the_order_ready_once", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.cake, f.pastry}, nil, 15)
		cake := f.ticketAt(t, order.ID, domaincategory.StationKitchen)
		pastry := f.ticketAt(t, order.ID, domaincategory.StationCounter)
		_, err := f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, cake.ID, domainorder.TicketInProgress)
		require.NoError(t, err)
		_, err = f.counter.PatchStatus(ctx, clerk, domainuser.RoleStaff, pastry.ID, domainorder.TicketInProgress)
		require.NoError(t, err)

		var wg sync.WaitGroup
		errs := make([]error, 2)
		start := make(chan struct{})
		wg.Go(func() {
			<-start
			_, errs[0] = f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, cake.ID, domainorder.TicketReady)
		})
		wg.Go(func() {
			<-start
			_, errs[1] = f.counter.PatchStatus(ctx, clerk, domainuser.RoleStaff, pastry.ID, domainorder.TicketReady)
		})
		close(start)
		wg.Wait()

		require.NoError(t, errs[0])
		require.NoError(t, errs[1])
		assert.Equal(t, domainorder.StatusReady, f.orderStatus(t, order.ID))
		detail, err := f.staff.Get(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		ready := 0
		for _, entry := range detail.Timeline {
			if entry.Status == string(domainorder.StatusReady) {
				ready++
			}
		}
		assert.Equal(t, 1, ready, "the order became ready exactly once")
	})

	t.Run("cancelling_mid_production_cancels_every_ticket", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.cake, f.pastry}, nil, 16)
		cake := f.ticketAt(t, order.ID, domaincategory.StationKitchen)
		_, err := f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, cake.ID, domainorder.TicketInProgress)
		require.NoError(t, err)

		_, err = f.staff.PatchStatus(ctx, clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), domainorder.StatusCancelled)

		require.NoError(t, err)
		tickets, err := f.tickets.ListByOrder(ctx, uuid.MustParse(order.ID))
		require.NoError(t, err)
		for _, ticket := range tickets {
			assert.Equal(t, domainorder.TicketCancelled, ticket.Status, ticket.Station)
		}
		_, err = f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, cake.ID, domainorder.TicketReady)
		require.ErrorIs(t, err, domainorder.ErrTicketOrderNotActive)
		pastry := f.ticketAt(t, order.ID, domaincategory.StationCounter)
		_, err = f.counter.Move(ctx, clerk, domainuser.RoleStaff, pastry.ID, domaincategory.StationKitchen)
		require.ErrorIs(t, err, domainorder.ErrTicketOrderNotActive, "a cancelled order has nothing left to move")
	})

	// Both paths lock the order row before its tickets, so they queue behind
	// each other instead of deadlocking, and the cancel always wins in the end.
	t.Run("a_cancel_racing_a_move_leaves_every_ticket_cancelled", func(t *testing.T) {
		for range 5 {
			order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 11)
			pastry := f.ticketAt(t, order.ID, domaincategory.StationCounter)

			var wg sync.WaitGroup
			var cancelErr, moveErr error
			start := make(chan struct{})
			wg.Go(func() {
				<-start
				_, cancelErr = f.staff.PatchStatus(ctx, clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), domainorder.StatusCancelled)
			})
			wg.Go(func() {
				<-start
				_, moveErr = f.counter.Move(ctx, clerk, domainuser.RoleStaff, pastry.ID, domaincategory.StationKitchen)
			})
			close(start)
			wg.Wait()

			require.NoError(t, cancelErr)
			if moveErr != nil {
				require.ErrorIs(t, moveErr, domainorder.ErrTicketOrderNotActive, "a move that came second finds the order cancelled")
			}
			tickets, err := f.tickets.ListByOrder(ctx, uuid.MustParse(order.ID))
			require.NoError(t, err)
			require.Len(t, tickets, 1)
			assert.Equal(t, domainorder.TicketCancelled, tickets[0].Status, "wherever the ticket ended up")
		}
	})

	t.Run("staff_move_a_ticket_nobody_started_to_a_free_station", func(t *testing.T) {
		single := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 17)
		pastry := f.ticketAt(t, single.ID, domaincategory.StationCounter)

		_, err := f.counter.Move(ctx, clerk, domainuser.RoleStaff, pastry.ID, domaincategory.StationCounter)
		require.ErrorIs(t, err, domainorder.ErrTicketStationTaken, "a ticket is already at its own station")
		change, err := f.counter.Move(ctx, clerk, domainuser.RoleStaff, pastry.ID, domaincategory.StationKitchen)
		require.NoError(t, err)
		assert.Equal(t, string(domaincategory.StationKitchen), change.Station)
		queue, _, _, _, err := f.kitchen.List(ctx, 1, 100, "")
		require.NoError(t, err)
		assert.True(t, containsTicket(queue, pastry.ID), "the kitchen now makes it")

		_, err = f.kitchen.PatchStatus(ctx, baker, domainuser.RoleBaker, pastry.ID, domainorder.TicketInProgress)
		require.NoError(t, err)
		_, err = f.counter.Move(ctx, clerk, domainuser.RoleStaff, pastry.ID, domaincategory.StationCounter)
		require.ErrorIs(t, err, domainorder.ErrTicketNotMovable, "a started ticket stays where it is")

		mixed := f.placeConfirmed(t, []uuid.UUID{f.cake, f.pastry}, nil, 17)
		mixedPastry := f.ticketAt(t, mixed.ID, domaincategory.StationCounter)
		_, err = f.counter.Move(ctx, clerk, domainuser.RoleStaff, mixedPastry.ID, domaincategory.StationKitchen)
		require.ErrorIs(t, err, domainorder.ErrTicketStationTaken, "an order has one ticket per station")
	})
}

func containsTicket(queue []dto.StationTicketResponse, id uuid.UUID) bool {
	for _, ticket := range queue {
		if ticket.ID == id.String() {
			return true
		}
	}
	return false
}

func keys(m map[domaincategory.Station][]string) []domaincategory.Station {
	out := make([]domaincategory.Station, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
