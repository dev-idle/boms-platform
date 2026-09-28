package usecase

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/shared/ctxmeta"
)

// Inside a transaction the two reads would share one connection, which pgx
// cannot use concurrently. The nil repository panics if either read runs.
func TestReadPickupRules_RefusesToRunInsideATransaction(t *testing.T) {
	t.Parallel()

	_, _, err := readPickupRules(ctxmeta.WithinTransaction(t.Context()), nil, time.Now())

	require.Error(t, err)
}
