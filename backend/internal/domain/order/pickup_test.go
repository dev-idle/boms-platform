package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatePickupAt(t *testing.T) {
	now := time.Date(2026, 7, 10, 10, 0, 0, 0, bakeryLocation)

	t.Run("rejects zero time", func(t *testing.T) {
		assert.ErrorIs(t, ValidatePickupAt(time.Time{}, now), ErrInvalidPickupAt)
	})

	t.Run("rejects before min lead time", func(t *testing.T) {
		at := now.Add(30 * time.Minute)
		assert.ErrorIs(t, ValidatePickupAt(at, now), ErrInvalidPickupAt)
	})

	t.Run("rejects outside bakery hours", func(t *testing.T) {
		at := time.Date(2026, 7, 10, 19, 0, 0, 0, bakeryLocation)
		assert.ErrorIs(t, ValidatePickupAt(at, now), ErrInvalidPickupAt)
	})

	t.Run("rejects beyond max advance", func(t *testing.T) {
		at := now.Add(pickupMaxAdvance + time.Minute)
		assert.ErrorIs(t, ValidatePickupAt(at, now), ErrInvalidPickupAt)
	})

	t.Run("accepts exactly at max advance", func(t *testing.T) {
		at := now.Add(pickupMaxAdvance)
		require.NoError(t, ValidatePickupAt(at, now))
	})

	t.Run("rejects one minute before opening", func(t *testing.T) {
		at := time.Date(2026, 7, 11, 7, 59, 0, 0, bakeryLocation)
		assert.ErrorIs(t, ValidatePickupAt(at, now), ErrInvalidPickupAt)
	})

	t.Run("accepts at opening hour", func(t *testing.T) {
		at := time.Date(2026, 7, 11, 8, 0, 0, 0, bakeryLocation)
		require.NoError(t, ValidatePickupAt(at, now))
	})

	t.Run("accepts one minute before closing", func(t *testing.T) {
		at := time.Date(2026, 7, 10, 17, 59, 0, 0, bakeryLocation)
		require.NoError(t, ValidatePickupAt(at, now))
	})

	t.Run("rejects at closing hour", func(t *testing.T) {
		at := time.Date(2026, 7, 10, 18, 0, 0, 0, bakeryLocation)
		assert.ErrorIs(t, ValidatePickupAt(at, now), ErrInvalidPickupAt)
	})

	t.Run("evaluates hours in bakery timezone for UTC input", func(t *testing.T) {
		// 07:00 UTC = 14:00 in Asia/Ho_Chi_Minh — inside business hours.
		at := time.Date(2026, 7, 10, 7, 0, 0, 0, time.UTC)
		require.NoError(t, ValidatePickupAt(at, now))

		// 12:00 UTC = 19:00 local — after closing even though 12:00 looks valid.
		late := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
		assert.ErrorIs(t, ValidatePickupAt(late, now), ErrInvalidPickupAt)
	})

	t.Run("accepts valid slot", func(t *testing.T) {
		at := time.Date(2026, 7, 10, 14, 0, 0, 0, bakeryLocation)
		require.NoError(t, ValidatePickupAt(at, now))
	})
}
