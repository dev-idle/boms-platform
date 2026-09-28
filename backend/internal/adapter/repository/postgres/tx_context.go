package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	"github.com/boms/backend/internal/shared/ctxmeta"
)

type txContextKey struct{}

func withTx(ctx context.Context, tx pgx.Tx) context.Context {
	// The marker travels with the handle so code above the adapter can see that
	// it is on one transaction connection without reaching for the handle.
	return context.WithValue(ctxmeta.WithinTransaction(ctx), txContextKey{}, tx)
}

func txFromContext(ctx context.Context) pgx.Tx {
	if ctx == nil {
		return nil
	}
	tx, _ := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx
}
