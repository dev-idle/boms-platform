package payment

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRefundRetryDelay(t *testing.T) {
	t.Parallel()
	for attempts, want := range map[int]time.Duration{
		1: time.Minute, 2: 2 * time.Minute, 5: 16 * time.Minute, 9: 256 * time.Minute, 10: 6 * time.Hour, 40: 6 * time.Hour,
	} {
		assert.Equal(t, want, RefundRetryDelay(attempts), "after %d refusals", attempts)
	}
}

func TestRefundError(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "paypal unavailable", RefundError(errors.New("paypal unavailable")))
	long := RefundError(errors.New(strings.Repeat("é", 600)))
	assert.Equal(t, 500, len([]rune(long)), "kept to what the column holds, whole characters")
}
