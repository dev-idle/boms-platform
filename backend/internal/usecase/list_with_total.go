package usecase

import (
	"context"

	"github.com/boms/backend/internal/shared/ctxmeta"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"golang.org/x/sync/errgroup"
)

// listWithTotal runs a page query and its count at the same time.
//
// The two do not depend on each other, and against a managed Postgres the round
// trip — not the query — is what a list endpoint spends its time on: running
// them in sequence pays that trip twice for one answer. Both callbacks get the
// group context, so the first failure cancels the other.
//
// Read paths only: it refuses to run inside a transaction, because a transaction
// is bound to one connection and cannot serve two queries at once — the pair
// would either serialize or poison the caller's transaction on cancellation.
func listWithTotal[T any](
	ctx context.Context,
	list func(ctx context.Context) ([]T, error),
	count func(ctx context.Context) (int64, error),
) ([]T, int64, error) {
	if ctxmeta.InTransaction(ctx) {
		return nil, 0, apperrors.Errorf("list with total: cannot run inside a transaction")
	}

	var (
		items []T
		total int64
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		found, err := list(groupCtx)
		if err != nil {
			return err
		}
		items = found
		return nil
	})
	group.Go(func() error {
		found, err := count(groupCtx)
		if err != nil {
			return err
		}
		total = found
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, 0, err
	}

	return items, total, nil
}
