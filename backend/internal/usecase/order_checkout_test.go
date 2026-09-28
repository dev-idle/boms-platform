package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

var _ port.TxManager = (*countingTxManager)(nil)

// countingTxManager records whether checkout got as far as writing anything.
type countingTxManager struct{ calls int }

func (c *countingTxManager) WithTx(context.Context, func(context.Context) error) error {
	c.calls++
	return errors.New("checkout reached its transaction")
}

// A refused pickup time never reaches the cart or the order tables: the rules
// come from the store settings, not from constants.
func TestOrderUsecase_Checkout_PickupRules(t *testing.T) {
	t.Parallel()
	inThreeDays := domainstore.DayOf(time.Now()).AddDate(0, 0, 3)
	at := func(day time.Time, hour, minute int) time.Time {
		return time.Date(day.Year(), day.Month(), day.Day(), hour, minute, 0, 0, domainstore.Location)
	}

	cases := map[string]struct {
		pickupAt time.Time
		prepare  func(*memoryStore)
		want     error
	}{
		"closed_day": {
			pickupAt: at(inThreeDays, 12, 0),
			prepare: func(m *memoryStore) {
				_, _ = m.AddClosedDate(context.Background(), inThreeDays, "Inventory")
			},
			want: domainorder.ErrPickupClosedDay,
		},
		"outside_edited_hours": {
			pickupAt: at(inThreeDays, 8, 30),
			prepare:  func(m *memoryStore) { m.settings.OpensAt = 9 * time.Hour },
			want:     domainorder.ErrPickupOutsideHours,
		},
		"past_a_shorter_booking_window": {
			pickupAt: at(inThreeDays, 12, 0),
			prepare:  func(m *memoryStore) { m.settings.MaxAdvanceDays = 2 },
			want:     domainorder.ErrPickupTooFar,
		},
		"store_unreachable": {
			pickupAt: at(inThreeDays, 12, 0),
			prepare:  func(m *memoryStore) { m.err = errors.New("database unavailable") },
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			store, tx := newMemoryStore(), &countingTxManager{}
			tc.prepare(store)
			uc := usecase.NewOrderUsecase(nil, nil, nil, nil, tx, &recordingOutbox{}, store)

			_, err := uc.Checkout(context.Background(), uuid.New(), tc.pickupAt)

			want := tc.want
			if want == nil {
				want = store.err
			}
			require.ErrorIs(t, err, want)
			assert.Zero(t, tx.calls)
		})
	}
}
