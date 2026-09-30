package postgres_test

import (
	"context"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// Concurrent checkouts on one bakery day each get their own number, and the
// numbers run 1..n with no gap: the counter row serializes them. The day comes
// from the transaction clock, in bakery time.
func TestOrderRepository_DayNumbers_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	const checkouts = 12
	pool := newIntegrationPool(t, checkouts+2)
	orders := postgresadapter.NewOrderRepository(pool)
	_, _, err := orders.NextDayNumber(ctx)
	require.Error(t, err, "a number taken outside the order's transaction would not be held")

	before := domainstore.DayOf(time.Now())
	var (
		mu      sync.Mutex
		numbers []int
		days    = map[time.Time]struct{}{}
		wg      sync.WaitGroup
		start   = make(chan struct{})
		errs    = make(chan error, checkouts)
	)
	for range checkouts {
		wg.Go(func() {
			<-start
			errs <- pool.WithTx(ctx, func(txCtx context.Context) error {
				day, number, err := orders.NextDayNumber(txCtx)
				if err != nil {
					return err
				}
				mu.Lock()
				numbers = append(numbers, number)
				days[day] = struct{}{}
				mu.Unlock()
				return nil
			})
		})
	}
	close(start)
	wg.Wait()
	close(errs)
	for err := range errs {
		require.NoError(t, err)
	}

	slices.Sort(numbers)
	want := make([]int, checkouts)
	for i := range want {
		want[i] = i + 1
	}
	assert.Equal(t, want, numbers)
	require.Len(t, days, 1, "one bakery day")
	for day := range days {
		assert.False(t, day.Before(before), "the bakery day the numbers were taken on")
		assert.False(t, day.After(domainstore.DayOf(time.Now())))
	}
}

func TestOrderRepository_HistoryAndFilters_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	users := postgresadapter.NewUserRepository(pool)
	orders := postgresadapter.NewOrderRepository(pool)

	customer, err := users.Create(ctx, port.CreateUserParams{
		Email: "history@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleCustomer,
	})
	require.NoError(t, err)
	staff, err := users.Create(ctx, port.CreateUserParams{
		Email: "history-staff@example.com", PasswordHash: testPasswordHashFixture, Role: domainuser.RoleStaff,
	})
	require.NoError(t, err)

	placed, err := orders.Create(ctx, port.CreateOrderParams{
		Code: "CH-260928-001", Type: domainorder.TypePreOrder, UserID: customer.ID, Status: domainorder.StatusPending, SubtotalCents: 4500, TotalCents: 4500,
	})
	require.NoError(t, err)
	assert.Equal(t, "CH-260928-001", placed.Code)

	t.Run("codes_are_unique", func(t *testing.T) {
		_, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260928-001", Type: domainorder.TypePreOrder, UserID: customer.ID, Status: domainorder.StatusPending, SubtotalCents: 1, TotalCents: 1,
		})
		require.ErrorIs(t, err, apperrors.ErrConflict)
	})

	t.Run("history_lists_every_move_in_order", func(t *testing.T) {
		require.NoError(t, orders.AddStatusEvent(ctx, port.AddOrderStatusEventParams{
			OrderID: placed.ID, To: domainorder.StatusPending, Actor: &port.OrderActor{ID: customer.ID, Role: domainuser.RoleCustomer},
		}))
		from := domainorder.StatusPending
		require.NoError(t, orders.AddStatusEvent(ctx, port.AddOrderStatusEventParams{
			OrderID: placed.ID, From: &from, To: domainorder.StatusConfirmed, Actor: &port.OrderActor{ID: staff.ID, Role: domainuser.RoleStaff},
		}))

		history, err := orders.ListStatusEvents(ctx, placed.ID)
		require.NoError(t, err)
		require.Len(t, history, 2)
		assert.Equal(t, domainorder.StatusPending, history[0].To)
		assert.Equal(t, domainuser.RoleCustomer, history[0].ActorRole)
		assert.Equal(t, domainorder.StatusConfirmed, history[1].To)
		assert.Equal(t, domainuser.RoleStaff, history[1].ActorRole)
		assert.False(t, history[1].At.Before(history[0].At))
	})

	t.Run("history_refuses_a_move_to_the_same_status", func(t *testing.T) {
		same := domainorder.StatusPending
		err := orders.AddStatusEvent(ctx, port.AddOrderStatusEventParams{
			OrderID: placed.ID, From: &same, To: domainorder.StatusPending, Actor: &port.OrderActor{ID: staff.ID, Role: domainuser.RoleStaff},
		})
		require.Error(t, err)
	})

	t.Run("history_filters_by_status_and_bakery_days", func(t *testing.T) {
		ready, err := orders.Create(ctx, port.CreateOrderParams{
			Code: "CH-260928-002", Type: domainorder.TypePreOrder, UserID: customer.ID, Status: domainorder.StatusReady, SubtotalCents: 3200, TotalCents: 3200,
		})
		require.NoError(t, err)

		list := func(filter port.OrderHistoryFilter) []string {
			t.Helper()
			found, err := orders.ListByUser(ctx, port.ListOrdersParams{UserID: customer.ID, Filter: filter, Limit: 10})
			require.NoError(t, err)
			total, err := orders.ListCountByUser(ctx, customer.ID, filter)
			require.NoError(t, err)
			codes := make([]string, 0, len(found))
			for _, o := range found {
				codes = append(codes, o.Code)
			}
			assert.Equal(t, int64(len(codes)), total, "the count uses the same filter as the page")
			return codes
		}
		status := domainorder.StatusReady
		// Filters take instants: the bakery midnight a day starts at. The day is
		// the one the database placed the orders on, not the test's clock.
		placedAt := placed.CreatedAt.In(domainstore.Location)
		today := time.Date(placedAt.Year(), placedAt.Month(), placedAt.Day(), 0, 0, 0, 0, domainstore.Location)
		tomorrow := today.AddDate(0, 0, 1)

		assert.Equal(t, []string{ready.Code, placed.Code}, list(port.OrderHistoryFilter{}))
		assert.Equal(t, []string{ready.Code}, list(port.OrderHistoryFilter{Status: &status}))
		assert.Len(t, list(port.OrderHistoryFilter{PlacedFrom: &today, PlacedBefore: &tomorrow}), 2)
		assert.Empty(t, list(port.OrderHistoryFilter{PlacedFrom: &tomorrow}))
		assert.Empty(t, list(port.OrderHistoryFilter{PlacedBefore: &today}))
	})
}
