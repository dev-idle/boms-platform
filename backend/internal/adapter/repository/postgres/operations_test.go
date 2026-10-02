package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// The dashboard counts, at a moment of the bakery day the orders are due on:
// the orders still to make or hand over by status, each station's tickets,
// the day's pickups, the orders whose slot has ended uncollected, and how long
// the orders that became ready that day took.
func TestOperationsDashboard_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	f.setSettings(t, func(s *domainstore.Settings) { s.SlotLength = 30 * time.Minute })
	dashboard := usecase.NewManagerOperationsUsecase(
		postgresadapter.NewOperationsRepository(f.pool), postgresadapter.NewSalesReportRepository(f.pool),
	)
	ticketAt := func(t *testing.T, order *dto.OrderResponse, station domaincategory.Station) uuid.UUID {
		t.Helper()
		return f.ticketAt(t, order.ID, station).ID
	}

	// Due at 9:00, confirmed: its counter ticket waits.
	croissant := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 9)
	// Due at 9:30, a custom request waiting for the counter to accept it.
	customer := f.newCustomer(t, nil, nil)
	pendingAt := tomorrowAt(9, 30)
	_, err = f.orders.Create(ctx, port.CreateOrderParams{
		Code: "CH-261003-901", UserID: &customer, Channel: domainorder.ChannelOnline, Status: domainorder.StatusPending,
		Type: domainorder.TypePreOrder, SubtotalCents: 100, TotalCents: 100, PickupAt: &pendingAt,
	})
	require.NoError(t, err)
	// Due at 10:00, being made in the kitchen.
	cake := f.placeConfirmed(t, []uuid.UUID{f.cake}, nil, 10)
	_, err = f.kitchen.PatchStatus(ctx, f.baker, domainuser.RoleBaker, ticketAt(t, cake, domaincategory.StationKitchen), domainorder.TicketInProgress)
	require.NoError(t, err)
	// Due at 11:00, ready at both stations: made from 10:00 to 10:30 that day.
	combo := f.placeConfirmed(t, nil, []uuid.UUID{f.combo}, 11)
	for _, station := range []domaincategory.Station{domaincategory.StationKitchen, domaincategory.StationCounter} {
		ticket := ticketAt(t, combo, station)
		if station == domaincategory.StationKitchen {
			_, err = f.kitchen.PatchStatus(ctx, f.baker, domainuser.RoleBaker, ticket, domainorder.TicketInProgress)
			require.NoError(t, err)
			_, err = f.kitchen.PatchStatus(ctx, f.baker, domainuser.RoleBaker, ticket, domainorder.TicketReady)
		} else {
			_, err = f.counter.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, ticket, domainorder.TicketInProgress)
			require.NoError(t, err)
			_, err = f.counter.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, ticket, domainorder.TicketReady)
		}
		require.NoError(t, err)
	}
	_, err = raw.Exec(ctx, `
		UPDATE order_status_events SET created_at = CASE to_status WHEN 'in_production' THEN $2::timestamptz ELSE $3::timestamptz END
		WHERE order_id = $1 AND to_status IN ('in_production', 'ready')`, combo.ID, tomorrowAt(10, 0), tomorrowAt(10, 30))
	require.NoError(t, err)
	// Due at 12:00 and collected.
	collected := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 12)
	_, err = raw.Exec(ctx, "UPDATE orders SET status = 'fulfilled' WHERE id = $1", collected.ID)
	require.NoError(t, err)
	// Due at 13:00 and cancelled: not the bakery's to count.
	cancelled := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 13)
	_, err = f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, uuid.MustParse(cancelled.ID), dto.PatchStaffOrderStatusRequest{
		Status: "cancelled", Reason: "Out of butter",
	})
	require.NoError(t, err)
	// Due at 14:00 and not collected.
	missed := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 14)
	_, err = raw.Exec(ctx, "UPDATE orders SET status = 'no_show' WHERE id = $1", missed.ID)
	require.NoError(t, err)
	// Due at 15:00 and never paid: not the bakery's yet.
	_, err = f.orderUC.Checkout(ctx, f.newCustomer(t, []uuid.UUID{f.pastry}, nil), uuid.New(), acceptingTerms(tomorrowAt(15, 0)))
	require.NoError(t, err)
	// Due at 8:00, confirmed, of an account since closed: off the counter's and
	// the kitchen's lists, so off the dashboard.
	closed := f.placeConfirmed(t, []uuid.UUID{f.pastry}, nil, 8)
	_, err = raw.Exec(ctx, "UPDATE users SET deleted_at = now() WHERE id = (SELECT user_id FROM orders WHERE id = $1)", closed.ID)
	require.NoError(t, err)

	out, err := dashboard.Dashboard(ctx, tomorrowAt(11, 45))

	require.NoError(t, err)
	assert.Equal(t, dto.ActiveOrdersResponse{Pending: 1, Confirmed: 1, InProduction: 1, Ready: 1}, out.Orders)
	assert.Equal(t, []dto.StationLoadResponse{
		{Station: "kitchen", InProgress: 1, Ready: 1},
		{Station: "counter", Queued: 1, Ready: 1},
	}, out.Stations, "the collected and cancelled orders' tickets are done with")
	assert.Equal(t, dto.DayPickupsResponse{Due: 6, Collected: 1, ToCollect: 4, Missed: 1}, out.Pickups)
	assert.Equal(t, int64(4), out.Late.Total, "11:00 is late once its slot ends at 11:30")
	codes := make([]string, 0, len(out.Late.Orders))
	for _, order := range out.Late.Orders {
		codes = append(codes, order.Code)
	}
	assert.Equal(t, []string{croissant.Code, "CH-261003-901", cake.Code, combo.Code}, codes, "the longest waiting first")
	minutes := int32(30)
	assert.Equal(t, dto.ProductionTimeResponse{Orders: 1, AverageMinutes: &minutes}, out.Production)

	early, err := dashboard.Dashboard(ctx, tomorrowAt(9, 29))
	require.NoError(t, err)
	assert.Zero(t, early.Late.Total, "9:00 is not late before its slot ends at 9:30")
}
