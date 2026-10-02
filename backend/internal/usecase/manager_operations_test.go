package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

// dayOperations answers the dashboard's parts and records the day asked for.
type dayOperations struct {
	from, to time.Time
	lateAt   time.Time
	err      error
}

func (f *dayOperations) ActiveOrders(context.Context) (port.ActiveOrders, error) {
	return port.ActiveOrders{Ready: 2}, f.err
}

func (f *dayOperations) StationLoads(context.Context) ([]port.StationLoad, error) {
	return nil, nil
}

func (f *dayOperations) DayPickups(_ context.Context, from, to time.Time) (port.DayPickups, error) {
	f.from, f.to = from, to
	return port.DayPickups{}, nil
}

func (f *dayOperations) LateOrders(_ context.Context, now time.Time, _ int32) ([]port.LateOrder, int64, error) {
	f.lateAt = now
	return []port.LateOrder{{Code: "CH-261003-001", Status: domainorder.StatusReady, PickupAt: now.Add(-time.Hour)}}, 3, nil
}

func TestManagerOperationsUsecase_Dashboard(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("reads_the_bakery_day_now_falls_in", func(t *testing.T) {
		t.Parallel()
		operations := &dayOperations{}
		// 3 a.m. on 3 October at the bakery: still 2 October in UTC.
		now := time.Date(2026, 10, 2, 20, 0, 0, 0, time.UTC)

		out, err := usecase.NewManagerOperationsUsecase(operations, &fixedSales{}).Dashboard(ctx, now)

		require.NoError(t, err)
		assert.Equal(t, time.Date(2026, 10, 3, 0, 0, 0, 0, domainstore.Location), operations.from)
		assert.Equal(t, time.Date(2026, 10, 4, 0, 0, 0, 0, domainstore.Location), operations.to)
		assert.Equal(t, now, operations.lateAt)
		assert.Equal(t, int64(2), out.Orders.Ready)
		assert.Equal(t, int64(3), out.Late.Total, "how many are late, beyond the ones listed")
		require.Len(t, out.Late.Orders, 1)
		assert.Equal(t, "ready", out.Late.Orders[0].Status)
		assert.Empty(t, out.Stations)
		assert.Nil(t, out.Production.AverageMinutes, "no order became ready today")
	})

	t.Run("a_failed_read_is_reported", func(t *testing.T) {
		t.Parallel()
		down := errors.New("database unavailable")

		_, err := usecase.NewManagerOperationsUsecase(&dayOperations{err: down}, &fixedSales{}).Dashboard(ctx, time.Now())

		require.ErrorIs(t, err, down)
	})
}
