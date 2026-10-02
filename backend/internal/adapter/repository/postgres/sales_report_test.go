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
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// The sales report adds up, against figures worked out by hand: the payments
// taken and refunds made in the bakery days asked for, by period; and what the
// orders kept — paid in them and not cancelled — sold, by item and category,
// with the discounts they used and how long orders took to make.
func TestSalesReport_Integration(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	f := newTicketFixture(t)
	raw, err := pgxpool.New(ctx, f.connStr)
	require.NoError(t, err)
	t.Cleanup(raw.Close)
	report := usecase.NewManagerSalesReportUsecase(postgresadapter.NewSalesReportRepository(f.pool))
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, domainstore.Location)
	at := func(day, clock string) time.Time {
		t.Helper()
		instant, err := time.ParseInLocation("2006-01-02 15:04", day+" "+clock, domainstore.Location)
		require.NoError(t, err)
		return instant
	}
	// sold places and pays an order of the lines given, then moves its
	// payment to paidAt.
	sold := func(t *testing.T, products, combos []uuid.UUID, code string, hour int, paidAt time.Time) uuid.UUID {
		t.Helper()
		customer := f.newCustomer(t, products, combos)
		if code != "" {
			_, err := f.cartUC.ApplyDiscount(ctx, customer, dto.ApplyCartDiscountRequest{Code: code})
			require.NoError(t, err)
		}
		placed, err := f.orderUC.Checkout(ctx, customer, uuid.New(), acceptingTerms(tomorrowAt(hour, 0)))
		require.NoError(t, err)
		f.pay(t, customer, placed.ID)
		order := uuid.MustParse(placed.ID)
		_, err = raw.Exec(ctx, "UPDATE payments SET captured_at = $2 WHERE order_id = $1", order, paidAt)
		require.NoError(t, err)
		return order
	}
	made := func(t *testing.T, orderID uuid.UUID, started, ready time.Time) {
		t.Helper()
		_, err := raw.Exec(ctx, `
			INSERT INTO order_status_events (order_id, from_status, to_status, created_at)
			VALUES ($1, 'confirmed', 'in_production', $2), ($1, 'in_production', 'ready', $3)`, orderID, started, ready)
		require.NoError(t, err)
	}
	_, err = postgresadapter.NewDiscountCodeRepository(f.pool).Create(ctx, port.CreateDiscountCodeParams{
		Code: "TEN", DiscountType: domaindiscount.TypePercent, Value: 10, IsActive: true,
		StartsAt: time.Now().Add(-time.Hour), EndsAt: time.Now().Add(24 * time.Hour),
	})
	require.NoError(t, err)

	// A croissant ($3.00) at 3 a.m. on Tuesday 1 September at the bakery — still
	// 31 August in UTC — made in 40 minutes.
	croissant := sold(t, []uuid.UUID{f.pastry}, nil, "", 9, at("2026-09-01", "03:00"))
	made(t, croissant, at("2026-09-01", "08:00"), at("2026-09-01", "08:40"))
	// A cake ($45.00) with TEN: $4.50 off, $40.50 paid on 2 September, made in an hour.
	cake := sold(t, []uuid.UUID{f.cake}, nil, "TEN", 10, at("2026-09-02", "09:00"))
	made(t, cake, at("2026-09-02", "07:00"), at("2026-09-02", "08:00"))
	// A combo ($46.00) paid on 8 September, cancelled and refunded on the 9th.
	combo := sold(t, nil, []uuid.UUID{f.combo}, "", 11, at("2026-09-08", "12:00"))
	_, err = f.staff.PatchStatus(ctx, f.clerk, domainuser.RoleStaff, combo, dto.PatchStaffOrderStatusRequest{
		Status: "cancelled", Reason: "Oven broke down",
	})
	require.NoError(t, err)
	_, err = f.paymentUC.RefundDue(ctx)
	require.NoError(t, err)
	_, err = raw.Exec(ctx, "UPDATE payments SET refunded_at = $2 WHERE order_id = $1", combo, at("2026-09-09", "10:00"))
	require.NoError(t, err)
	// A combo kept, paid on 10 September.
	sold(t, nil, []uuid.UUID{f.combo}, "", 14, at("2026-09-10", "12:00"))
	// Croissants paid half an hour before the range and a minute before its end.
	sold(t, []uuid.UUID{f.pastry}, nil, "", 12, at("2026-08-31", "23:30"))
	sold(t, []uuid.UUID{f.pastry}, nil, "", 13, at("2026-09-14", "23:59"))

	t.Run("adds_up_the_range_by_week", func(t *testing.T) {
		out, err := report.Report(ctx, "2026-09-01", "2026-09-14", "week", now)

		require.NoError(t, err)
		assert.Equal(t, []dto.SalesPeriodResponse{
			{Start: "2026-09-01", Orders: 2, GrossCents: 4350, NetCents: 4350},
			{Start: "2026-09-07", Orders: 2, GrossCents: 9200, RefundsCents: 4600, NetCents: 4600},
			{Start: "2026-09-14", Orders: 1, GrossCents: 300, NetCents: 300},
		}, out.Periods, "weeks from Monday; the first starts on the range's first day")
		average := int64(2770)
		assert.Equal(t, dto.SalesTotalsResponse{
			Orders: 5, GrossCents: 13850, RefundsCents: 4600, NetCents: 9250, AverageOrderCents: &average,
		}, out.Totals, "$138.50 over five orders")
		assert.Equal(t, []dto.SoldItemResponse{
			{Kind: "combo", ID: f.combo.String(), Name: "Tea set", Quantity: 1, SalesCents: 4600},
			{Kind: "product", ID: f.cake.String(), Name: "Matcha cake", Quantity: 1, SalesCents: 4500},
			{Kind: "product", ID: f.pastry.String(), Name: "Croissant", Quantity: 2, SalesCents: 600},
		}, out.Items, "the cancelled combo sold nothing; lines before the discount")
		assert.Equal(t, []dto.CategorySalesResponse{
			{Quantity: 1, SalesCents: 4600},
			{CategoryID: ptr(t, out.Categories[1].CategoryID), Name: ptr(t, out.Categories[1].Name), Quantity: 1, SalesCents: 4500},
			{CategoryID: ptr(t, out.Categories[2].CategoryID), Name: ptr(t, out.Categories[2].Name), Quantity: 2, SalesCents: 600},
		}, out.Categories, "the combos as one row without a category")
		assert.Equal(t, []string{"Cakes", "Pastries"}, []string{*out.Categories[1].Name, *out.Categories[2].Name})
		assert.Equal(t, dto.SalesDiscountsResponse{
			Orders: 4, DiscountedOrders: 1, DiscountCents: 450,
			Codes: []dto.DiscountCodeUseResponse{{Code: "TEN", Orders: 1, DiscountCents: 450}},
		}, out.Discounts)
		minutes := int32(50)
		assert.Equal(t, dto.ProductionTimeResponse{Orders: 2, AverageMinutes: &minutes}, out.Production)
	})

	t.Run("a_day_starts_at_midnight_at_the_bakery", func(t *testing.T) {
		out, err := report.Report(ctx, "2026-08-31", "2026-09-01", "day", now)

		require.NoError(t, err)
		assert.Equal(t, []dto.SalesPeriodResponse{
			{Start: "2026-08-31", Orders: 1, GrossCents: 300, NetCents: 300},
			{Start: "2026-09-01", Orders: 1, GrossCents: 300, NetCents: 300},
		}, out.Periods, "3 a.m. on the 1st at the bakery is the 1st, though UTC is still on the 31st")
	})

	t.Run("a_refund_counts_on_the_day_the_money_went_back", func(t *testing.T) {
		out, err := report.Report(ctx, "2026-09-08", "2026-09-09", "day", now)

		require.NoError(t, err)
		assert.Equal(t, []dto.SalesPeriodResponse{
			{Start: "2026-09-08", Orders: 1, GrossCents: 4600, NetCents: 4600},
			{Start: "2026-09-09", RefundsCents: 4600, NetCents: -4600},
		}, out.Periods)
	})

	t.Run("a_month_holds_its_days_and_a_quiet_range_holds_none", func(t *testing.T) {
		out, err := report.Report(ctx, "2026-09-01", "2026-09-30", "month", now)
		require.NoError(t, err)
		require.Len(t, out.Periods, 1)
		assert.Equal(t, int64(5), out.Periods[0].Orders)

		quiet, err := report.Report(ctx, "2026-07-01", "2026-07-03", "day", now)
		require.NoError(t, err)
		assert.Len(t, quiet.Periods, 3, "every day listed")
		assert.Nil(t, quiet.Totals.AverageOrderCents)
		assert.Empty(t, quiet.Items)
		assert.Nil(t, quiet.Production.AverageMinutes)
	})
}

// ptr requires v to be set and returns it, for a value only the database knows.
func ptr[T any](t *testing.T, v *T) *T {
	t.Helper()
	require.NotNil(t, v)
	return v
}
