package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// WithTx must hand its connection and row locks back however fn ends. The pool
// holds a single connection, so a leaked transaction would stall the next query.
func TestPoolWithTx_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 1)
	categories := postgresadapter.NewCategoryRepository(pool)

	createIn := func(txCtx context.Context, slug string) (uuid.UUID, error) {
		category, err := categories.Create(txCtx, port.CreateCategoryParams{Name: slug, Slug: slug, IsActive: true})
		if err != nil {
			return uuid.Nil, err
		}
		return category.ID, nil
	}

	t.Run("commits_when_fn_succeeds", func(t *testing.T) {
		var id uuid.UUID
		require.NoError(t, pool.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			id, err = createIn(txCtx, "committed")
			return err
		}))
		_, err := categories.GetByID(ctx, id)
		assert.NoError(t, err)
	})

	t.Run("rolls_back_and_returns_the_error_when_fn_fails", func(t *testing.T) {
		errStop := errors.New("stop")
		var id uuid.UUID
		err := pool.WithTx(ctx, func(txCtx context.Context) error {
			var err error
			if id, err = createIn(txCtx, "rolled-back"); err != nil {
				return err
			}
			return errStop
		})
		require.ErrorIs(t, err, errStop)
		assert.Equal(t, errStop.Error(), err.Error(), "a clean rollback adds nothing to the error")

		_, err = categories.GetByID(ctx, id)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})

	t.Run("releases_the_connection_when_fn_panics", func(t *testing.T) {
		var id uuid.UUID
		assert.Panics(t, func() {
			_ = pool.WithTx(ctx, func(txCtx context.Context) error {
				var err error
				if id, err = createIn(txCtx, "panicked"); err != nil {
					return err
				}
				panic("boom")
			})
		})

		queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		_, err := categories.GetByID(queryCtx, id)
		assert.ErrorIs(t, err, apperrors.ErrNotFound, "the panicked transaction must be rolled back and its connection reusable")
	})
}
