package order

import (
	"regexp"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPickupCodes(t *testing.T) {
	t.Parallel()
	codes := NewPickupCodes("a-key-long-enough-for-a-real-bakery-0001")
	order := uuid.New()

	code := codes.Of(order)

	assert.Regexp(t, regexp.MustCompile(`^\d{4}$`), code)
	assert.Equal(t, code, codes.Of(order), "the same order always has the same code")
	assert.True(t, codes.Matches(order, code))
	assert.False(t, codes.Matches(order, "12345"))
	assert.False(t, codes.Matches(uuid.New(), code+"x"))

	differ := 0
	other := NewPickupCodes("another-key-long-enough-for-a-bakery-0002")
	for range 20 {
		id := uuid.New()
		if codes.Of(id) != other.Of(id) {
			differ++
		}
	}
	assert.Greater(t, differ, 15, "a code depends on the key")
}

func TestStatus_HasPickupCode(t *testing.T) {
	t.Parallel()
	for _, s := range []Status{StatusConfirmed, StatusInProduction, StatusReady} {
		assert.True(t, s.HasPickupCode(), s)
	}
	for _, s := range []Status{StatusAwaitingPayment, StatusPending, StatusFulfilled, StatusCancelled, StatusExpired, StatusNoShow} {
		assert.False(t, s.HasPickupCode(), s)
	}
}
