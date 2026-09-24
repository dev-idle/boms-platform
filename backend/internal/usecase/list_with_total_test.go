package usecase

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/boms/backend/internal/shared/ctxmeta"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListWithTotal(t *testing.T) {
	t.Parallel()

	t.Run("returns_both_results", func(t *testing.T) {
		t.Parallel()

		items, total, err := listWithTotal(t.Context(),
			func(context.Context) ([]string, error) { return []string{"a", "b"}, nil },
			func(context.Context) (int64, error) { return 7, nil },
		)

		require.NoError(t, err)
		assert.Equal(t, []string{"a", "b"}, items)
		assert.Equal(t, int64(7), total)
	})

	t.Run("runs_the_two_queries_at_the_same_time", func(t *testing.T) {
		t.Parallel()

		// Each side waits for the other to start: sequential execution would
		// deadlock until the context deadline, so passing proves concurrency.
		listStarted := make(chan struct{})
		countStarted := make(chan struct{})
		ctx, cancel := context.WithTimeout(t.Context(), 2*time.Second)
		defer cancel()

		_, total, err := listWithTotal(ctx,
			func(ctx context.Context) ([]string, error) {
				close(listStarted)
				select {
				case <-countStarted:
					return nil, nil
				case <-ctx.Done():
					return nil, ctx.Err()
				}
			},
			func(ctx context.Context) (int64, error) {
				close(countStarted)
				select {
				case <-listStarted:
					return 1, nil
				case <-ctx.Done():
					return 0, ctx.Err()
				}
			},
		)

		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
	})

	t.Run("returns_the_list_failure_and_cancels_the_count", func(t *testing.T) {
		t.Parallel()

		listErr := errors.New("list failed")
		var countCancelled atomic.Bool

		items, total, err := listWithTotal(t.Context(),
			func(context.Context) ([]string, error) { return nil, listErr },
			func(ctx context.Context) (int64, error) {
				<-ctx.Done()
				countCancelled.Store(true)
				return 0, ctx.Err()
			},
		)

		require.ErrorIs(t, err, listErr)
		assert.Nil(t, items)
		assert.Zero(t, total)
		assert.True(t, countCancelled.Load(), "the count should see the cancelled group context")
	})

	t.Run("returns_the_count_failure", func(t *testing.T) {
		t.Parallel()

		countErr := errors.New("count failed")

		items, total, err := listWithTotal(t.Context(),
			func(context.Context) ([]string, error) { return []string{"a"}, nil },
			func(context.Context) (int64, error) { return 0, countErr },
		)

		require.ErrorIs(t, err, countErr)
		assert.Nil(t, items, "a partial page is never returned beside an error")
		assert.Zero(t, total)
	})

	t.Run("refuses_to_run_inside_a_transaction", func(t *testing.T) {
		t.Parallel()

		var ran atomic.Bool
		_, _, err := listWithTotal(ctxmeta.WithinTransaction(t.Context()),
			func(context.Context) ([]string, error) { ran.Store(true); return nil, nil },
			func(context.Context) (int64, error) { ran.Store(true); return 0, nil },
		)

		require.Error(t, err)
		assert.False(t, ran.Load(), "neither query may run on a transaction connection")
	})

	t.Run("passes_a_cancelled_context_through", func(t *testing.T) {
		t.Parallel()

		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		_, _, err := listWithTotal(ctx,
			func(ctx context.Context) ([]string, error) { return nil, ctx.Err() },
			func(ctx context.Context) (int64, error) { return 0, ctx.Err() },
		)

		require.ErrorIs(t, err, context.Canceled)
	})
}
