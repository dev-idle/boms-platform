package postgres_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// What goes wrong with an order is recorded as it happens or as staff report
// it, and managers read it a bakery week at a time; a customer whose payments
// keep failing or expiring is flagged once.
func TestOrderIncidents_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	managers := usecase.NewManagerIncidentUsecase(f.orders)
	_, err := postgresadapter.NewStaffProfileRepository(f.pool).Create(ctx, port.UpsertStaffProfileParams{
		UserID: f.clerk, FullName: "Lan Tran", EmployeeCode: "EMP-9301",
	})
	require.NoError(t, err)
	today := time.Now().In(domainstore.Location)
	week := today.AddDate(0, 0, -(int(today.Weekday())+6)%7).Format(domainstore.DayLayout)
	logged := func(t *testing.T, incidentType string) []dto.ManagerOrderIncidentResponse {
		t.Helper()
		out, _, _, _, err := managers.List(ctx, week, incidentType, 1, 100)
		require.NoError(t, err)
		return out
	}
	// unpaidOrders places n orders awaiting payment for one new customer.
	placed := 0
	unpaidOrders := func(t *testing.T, n int) []uuid.UUID {
		t.Helper()
		customer := f.newCustomer(t, nil, nil)
		pickup := tomorrowAt(15, 0)
		ids := make([]uuid.UUID, 0, n)
		for range n {
			placed++
			order, err := f.orders.Create(ctx, port.CreateOrderParams{
				Code: fmt.Sprintf("CH-261002-%03d", 500+placed), UserID: &customer, Channel: domainorder.ChannelOnline,
				Status: domainorder.StatusAwaitingPayment, Type: domainorder.TypePreOrder, SubtotalCents: 100, TotalCents: 100, PickupAt: &pickup,
			})
			require.NoError(t, err)
			ids = append(ids, order.ID)
		}
		return ids
	}
	// fail records a failed payment for the order as a declined capture does,
	// runs beforeCommit, and returns the flag it raised, if any.
	fail := func(t *testing.T, orderID uuid.UUID, beforeCommit func()) *domainorder.Incident {
		t.Helper()
		var flag *domainorder.Incident
		require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error {
			if _, err := f.orders.AddIncident(txCtx, port.AddOrderIncidentParams{OrderID: orderID, Type: domainorder.IncidentPaymentFailed}); err != nil {
				return err
			}
			var err error
			flag, err = f.orders.FlagPaymentAnomaly(txCtx, orderID, domainorder.PaymentAnomalyWindow, domainorder.PaymentAnomalyThreshold)
			beforeCommit()
			return err
		}))
		return flag
	}

	t.Run("staff_report_a_problem_with_an_order_they_see", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 10)

		out, err := f.staff.ReportIncident(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), dto.ReportOrderIncidentRequest{
			Type: "wrong_items", Note: "  Two croissants short.  ",
		})

		require.NoError(t, err)
		assert.Equal(t, "wrong_items", out.Type)
		assert.Equal(t, "manual", out.Source)
		require.NotNil(t, out.Note)
		assert.Equal(t, "Two croissants short.", *out.Note)
		reported := logged(t, "wrong_items")
		require.Len(t, reported, 1)
		assert.Equal(t, out.ID, reported[0].ID)
		assert.Equal(t, order.Code, reported[0].OrderCode)
		require.NotNil(t, reported[0].ActorName)
		assert.Equal(t, "Lan Tran", *reported[0].ActorName)
	})

	t.Run("an_order_staff_do_not_see_takes_no_report", func(t *testing.T) {
		unpaid := unpaidOrders(t, 1)[0]

		_, err := f.staff.ReportIncident(ctx, f.clerk, domainuser.RoleStaff, unpaid, dto.ReportOrderIncidentRequest{
			Type: "other", Note: "Looks odd",
		})

		require.ErrorIs(t, err, domainorder.ErrNotFound)
	})

	t.Run("an_erased_customers_order_takes_no_report", func(t *testing.T) {
		customer := f.newCustomer(t, []uuid.UUID{f.pastry}, nil)
		order, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(12, 0)))
		require.NoError(t, err)
		f.pay(t, customer, order.ID)
		require.NoError(t, f.pool.WithTx(ctx, func(txCtx context.Context) error { return f.users.Erase(txCtx, customer) }))

		_, err = f.staff.ReportIncident(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), dto.ReportOrderIncidentRequest{
			Type: "other", Note: "Mai said the box was crushed",
		})

		require.ErrorIs(t, err, domainorder.ErrNotFound, "nothing new is written about an erased customer")
	})

	t.Run("the_bakery_cancelling_an_accepted_order_is_recorded_and_so_is_its_refund", func(t *testing.T) {
		order := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 11)

		_, err := f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(order.ID), dto.PatchStaffOrderStatusRequest{
			Status: "cancelled", Reason: "Out of butter",
		})
		require.NoError(t, err)
		_, err = f.paymentUC.RefundDue(ctx)
		require.NoError(t, err)

		cancelled := logged(t, "bakery_cancelled")
		require.Len(t, cancelled, 1)
		assert.Equal(t, order.Code, cancelled[0].OrderCode)
		require.NotNil(t, cancelled[0].Note)
		assert.Equal(t, "Out of butter", *cancelled[0].Note, "the reason the customer was given")
		require.NotNil(t, cancelled[0].ActorName)
		assert.Equal(t, "Lan Tran", *cancelled[0].ActorName)
		refunded := logged(t, "refunded")
		require.Len(t, refunded, 1)
		assert.Equal(t, "auto", refunded[0].Source)
		assert.Nil(t, refunded[0].ActorName, "the system recorded it")
	})

	t.Run("a_customer_whose_payments_keep_failing_is_flagged_once", func(t *testing.T) {
		orders := unpaidOrders(t, 4)

		assert.Nil(t, fail(t, orders[0], func() {}))
		assert.Nil(t, fail(t, orders[1], func() {}))
		flag := fail(t, orders[2], func() {})
		require.NotNil(t, flag, "the third failure within a day flags them")
		assert.Equal(t, domainorder.IncidentPaymentAnomaly, flag.Type)
		assert.Equal(t, orders[2], flag.OrderID, "on the order that tipped it")
		assert.Nil(t, fail(t, orders[3], func() {}), "one flag covers the day")
	})

	t.Run("payments_failing_at_once_all_count", func(t *testing.T) {
		orders := unpaidOrders(t, 3)
		require.Nil(t, fail(t, orders[0], func() {}))

		// Each keeps its transaction open until the other has counted too, or
		// long enough to show the other waits on the lock until it commits:
		// without the lock, both would count two and neither would flag.
		var counted sync.WaitGroup
		counted.Add(2)
		bothCounted := make(chan struct{})
		go func() {
			counted.Wait()
			close(bothCounted)
		}()
		hold := func() {
			counted.Done()
			select {
			case <-bothCounted:
			case <-time.After(300 * time.Millisecond):
			}
		}
		flags := make([]*domainorder.Incident, 2)
		var done sync.WaitGroup
		for i, orderID := range orders[1:] {
			done.Go(func() { flags[i] = fail(t, orderID, hold) })
		}
		done.Wait()

		raised := 0
		for _, flag := range flags {
			if flag != nil {
				raised++
			}
		}
		assert.Equal(t, 1, raised)
	})

	t.Run("a_week_is_counted_by_type", func(t *testing.T) {
		summary, err := managers.Summary(ctx, week)
		require.NoError(t, err)
		counts := map[string]int64{}
		for _, count := range summary.Types {
			counts[count.Type] = count.Count
		}
		assert.Equal(t, map[string]int64{
			"wrong_items": 1, "bakery_cancelled": 1, "refunded": 1, "payment_failed": 7, "payment_anomaly": 2,
		}, counts)

		earlier, err := managers.Summary(ctx, today.AddDate(0, 0, -7-(int(today.Weekday())+6)%7).Format(domainstore.DayLayout))
		require.NoError(t, err)
		assert.Empty(t, earlier.Types, "last week holds none of them")
	})
}
