package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// fixedSales answers every part of a report as told and records the periods
// asked for.
type fixedSales struct {
	asked   port.SalesRange
	periods []port.SalesPeriod
	err     error
}

func (f *fixedSales) SalesByPeriod(_ context.Context, params port.SalesRange) ([]port.SalesPeriod, error) {
	f.asked = params
	return f.periods, f.err
}

func (f *fixedSales) TopSoldItems(context.Context, time.Time, time.Time, int32) ([]port.SoldItem, error) {
	return nil, nil
}

func (f *fixedSales) SalesByCategory(context.Context, time.Time, time.Time) ([]port.CategorySales, error) {
	return nil, nil
}

func (f *fixedSales) SalesDiscounts(context.Context, time.Time, time.Time) (port.DiscountUse, error) {
	return port.DiscountUse{}, nil
}

func (f *fixedSales) SalesDiscountCodes(context.Context, time.Time, time.Time, int32) ([]port.DiscountCodeUse, error) {
	return nil, nil
}

func (f *fixedSales) ProductionTime(context.Context, time.Time, time.Time) (port.ProductionTime, error) {
	return port.ProductionTime{}, nil
}

func TestManagerSalesReportUsecase_Report(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	now := time.Date(2026, 10, 2, 12, 0, 0, 0, domainstore.Location)

	t.Run("reads_the_bakery_days_asked_for", func(t *testing.T) {
		t.Parallel()
		sales := &fixedSales{}

		out, err := usecase.NewManagerSalesReportUsecase(sales).Report(ctx, " 2026-09-01 ", "2026-10-02", "week", now)

		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 9, 1, 0, 0, 0, 0, domainstore.Location), sales.asked.From, "from midnight at the bakery")
		assert.Equal(t, time.Date(2026, 10, 3, 0, 0, 0, 0, domainstore.Location), sales.asked.To, "to the midnight after the last day")
		assert.Equal(t, "week", sales.asked.Grouping)
		assert.Equal(t, "2026-09-01", out.From)
		assert.Equal(t, "2026-10-02", out.To)
		assert.Nil(t, out.Totals.AverageOrderCents, "no order, no average")
		assert.Nil(t, out.Production.AverageMinutes)
	})

	t.Run("nets_each_period_and_rounds_the_average_to_the_cent", func(t *testing.T) {
		t.Parallel()
		sales := &fixedSales{periods: []port.SalesPeriod{
			{Start: time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC), Orders: 2, GrossCents: 301, RefundsCents: 0},
			{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), Orders: 1, GrossCents: 0, RefundsCents: 500},
		}}

		out, err := usecase.NewManagerSalesReportUsecase(sales).Report(ctx, "2026-09-01", "2026-09-02", "day", now)

		require.NoError(t, err)
		assert.Equal(t, int64(-500), out.Periods[1].NetCents, "a day can give back more than it took")
		assert.Equal(t, int64(-199), out.Totals.NetCents)
		require.NotNil(t, out.Totals.AverageOrderCents)
		assert.Equal(t, int64(100), *out.Totals.AverageOrderCents, "301 over 3 is 100.33")
	})

	t.Run("refuses_a_range_it_cannot_report", func(t *testing.T) {
		t.Parallel()
		uc := usecase.NewManagerSalesReportUsecase(&fixedSales{})
		for _, tc := range []struct{ from, to, group, field string }{
			{"2026/09/01", "2026-09-02", "day", "from"},
			{"2026-09-01", "", "day", "to"},
			{"2026-09-01", "2026-09-02", "year", "group"},
			{"2026-09-01", "2026-10-03", "day", "to"},
			{"2026-09-02", "2026-09-01", "day", "to"},
			{"2025-09-30", "2026-10-01", "month", "from"},
		} {
			_, err := uc.Report(ctx, tc.from, tc.to, tc.group, now)
			requireValidationField(t, err, tc.field)
		}
		_, err := uc.Report(ctx, "2025-10-01", "2026-10-01", "month", now)
		require.NoError(t, err, "366 days is the longest")
	})

	t.Run("a_failed_read_is_reported", func(t *testing.T) {
		t.Parallel()
		down := errors.New("database unavailable")

		_, err := usecase.NewManagerSalesReportUsecase(&fixedSales{err: down}).Report(ctx, "2026-09-01", "2026-09-02", "day", now)

		require.ErrorIs(t, err, down)
	})
}
