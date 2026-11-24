package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDollars(t *testing.T) {
	t.Parallel()
	for cents, want := range map[int64]string{
		0:         "$0.00",
		900:       "$9.00",
		123456:    "$1,234.56",
		100000000: "$1,000,000.00",
	} {
		assert.Equal(t, want, Dollars(cents))
	}
}
