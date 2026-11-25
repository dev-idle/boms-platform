package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
)

// A cancellation re-checks, under the order's lock, that the change still
// strands its pickup: one moved meanwhile is left as it is.
func TestStrands(t *testing.T) {
	t.Parallel()
	day := time.Date(2031, 2, 10, 0, 0, 0, 0, domainstore.Location)
	at := func(hour, minute int) time.Time {
		return day.Add(time.Duration(hour)*time.Hour + time.Duration(minute)*time.Minute)
	}

	closed := closedDayParams(day)
	assert.True(t, strands(closed, at(9, 0)))
	assert.False(t, strands(closed, at(9, 0).AddDate(0, 0, 1)), "moved to the next day")

	hours := hoursParams(9*time.Hour, 17*time.Hour, day)
	assert.True(t, strands(hours, at(8, 30)), "before the new opening")
	assert.True(t, strands(hours, at(17, 0)), "at the new closing")
	assert.False(t, strands(hours, at(9, 0)))
	assert.False(t, strands(hours, at(16, 30)))
	assert.False(t, strands(hours, day.Add(-time.Hour)), "before the change applies")
}

// movedOrders answers the lock with the order as it is now: its pickup moved.
type movedOrders struct {
	port.OrderRepository
	now domainorder.Order
}

func (o movedOrders) LockForUpdate(context.Context, uuid.UUID) (*domainorder.Order, error) {
	return &o.now, nil
}

func TestCancelStranded_LeavesAnOrderMovedSince(t *testing.T) {
	t.Parallel()
	day := time.Date(2031, 2, 10, 0, 0, 0, 0, domainstore.Location)
	moved := day.AddDate(0, 0, 1).Add(9 * time.Hour)
	listed := port.StrandedPickup{ID: uuid.New(), Code: "CH-310210-001", Status: domainorder.StatusConfirmed, PickupAt: day.Add(9 * time.Hour)}
	// No transitions or audit: either would fail the test if the order were cancelled.
	u := &AdminStoreSettingsUsecase{
		orders: movedOrders{now: domainorder.Order{ID: listed.ID, Status: domainorder.StatusConfirmed, PickupAt: &moved}},
		tx:     inlineTx{},
	}

	done, err := u.cancelStranded(context.Background(), port.OrderActor{ID: uuid.New(), Role: domainuser.RoleAdmin},
		listed, closedDayParams(day), "Power cut")

	require.NoError(t, err)
	assert.False(t, done, "its pickup is no longer on the closed day")
}
