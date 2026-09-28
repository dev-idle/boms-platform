package order

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCode(t *testing.T) {
	t.Parallel()
	day := time.Date(2026, 9, 28, 0, 0, 0, 0, time.UTC)

	assert.Equal(t, "CH-260928-001", Code(day, 1))
	assert.Equal(t, "CH-260928-042", Code(day, 42))
	assert.Equal(t, "CH-260928-1000", Code(day, 1000), "a busy day grows the number instead of wrapping it")
}
