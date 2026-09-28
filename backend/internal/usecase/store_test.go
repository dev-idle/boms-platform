package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

func TestStoreUsecase_PickupRules(t *testing.T) {
	t.Parallel()
	store := newMemoryStore()
	today := domainstore.DayOf(time.Now())
	_, _ = store.AddClosedDate(context.Background(), today.AddDate(0, 0, 3), "Inventory")
	_, _ = store.AddClosedDate(context.Background(), today.AddDate(0, 0, 40), "Beyond the booking window")

	out, err := usecase.NewStoreUsecase(store, &slotOrders{}).PickupRules(context.Background())

	require.NoError(t, err)
	assert.Equal(t, "08:00", out.OpensAt)
	assert.Equal(t, "18:00", out.ClosesAt)
	assert.Equal(t, 30, out.SlotMinutes)
	assert.Equal(t, 120, out.PreorderMinLeadMinutes)
	assert.Equal(t, 20, out.InstantPrepMinutes)
	assert.Equal(t, 14, out.MaxAdvanceDays)
	assert.Equal(t, []dto.PublicClosedDateResponse{
		{Date: today.AddDate(0, 0, 3).Format(domainstore.DayLayout), Reason: "Inventory"},
	}, out.ClosedDates, "only closures a customer could pick are listed")
}

func TestStoreUsecase_PickupRules_StoreDown(t *testing.T) {
	t.Parallel()
	store := newMemoryStore()
	store.err = errors.New("database unavailable")

	_, err := usecase.NewStoreUsecase(store, &slotOrders{}).PickupRules(context.Background())

	require.ErrorIs(t, err, store.err)
}

// slotOrders answers how many orders are due at each pickup time; any other
// call panics.
type slotOrders struct {
	port.OrderRepository
	due  []port.PickupCount
	from time.Time
	to   time.Time
}

func (s *slotOrders) CountByPickupTime(_ context.Context, from, to time.Time) ([]port.PickupCount, error) {
	s.from, s.to = from, to
	return s.due, nil
}

func TestStoreUsecase_PickupSlots(t *testing.T) {
	t.Parallel()
	tomorrow := domainstore.DayOf(time.Now()).AddDate(0, 0, 1)
	at := func(hour, minute int) time.Time {
		return time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), hour, minute, 0, 0, domainstore.Location)
	}

	t.Run("lists_the_day_and_marks_full_slots", func(t *testing.T) {
		t.Parallel()
		store := newMemoryStore()
		store.settings.SlotCapacity = 2
		// 08:40 and 08:45 were booked on an older grid; they still fill the 08:30 slot.
		orders := &slotOrders{due: []port.PickupCount{{At: at(8, 30), Orders: 1}, {At: at(8, 45), Orders: 1}, {At: at(9, 0), Orders: 1}}}

		out, err := usecase.NewStoreUsecase(store, orders).PickupSlots(context.Background(), tomorrow.Format(domainstore.DayLayout))

		require.NoError(t, err)
		require.Len(t, out.Slots, 20, "08:00 to 17:30 every 30 minutes")
		assert.Equal(t, at(8, 0).UTC().Format(time.RFC3339), out.Slots[0].StartsAt)
		assert.False(t, out.Slots[0].Full)
		assert.True(t, out.Slots[1].Full, "08:30 holds as many orders as a slot takes")
		assert.False(t, out.Slots[2].Full, "09:00 has room")
		assert.True(t, orders.from.Equal(at(8, 0)))
		assert.True(t, orders.to.Equal(at(18, 0)), "counted up to closing")
	})

	t.Run("a_closed_day_has_no_slots", func(t *testing.T) {
		t.Parallel()
		store := newMemoryStore()
		_, _ = store.AddClosedDate(context.Background(), tomorrow, "Inventory")

		out, err := usecase.NewStoreUsecase(store, &slotOrders{}).PickupSlots(context.Background(), tomorrow.Format(domainstore.DayLayout))

		require.NoError(t, err)
		assert.Empty(t, out.Slots)
	})

	t.Run("refuses_a_day_it_cannot_read_or_book", func(t *testing.T) {
		t.Parallel()
		uc := usecase.NewStoreUsecase(newMemoryStore(), &slotOrders{})
		today := domainstore.DayOf(time.Now())
		for _, date := range []string{"tomorrow", "", today.AddDate(0, 0, -1).Format(domainstore.DayLayout), today.AddDate(0, 0, 15).Format(domainstore.DayLayout)} {
			_, err := uc.PickupSlots(context.Background(), date)
			var appErr *apperrors.AppError
			require.ErrorAs(t, err, &appErr, "date %q", date)
			assert.Contains(t, appErr.Details, "date")
		}
	})
}
