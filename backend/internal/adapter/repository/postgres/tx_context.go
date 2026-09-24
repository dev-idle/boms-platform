package postgres

import (
	"context"
	"database/sql"

	"github.com/boms/backend/internal/shared/ctxmeta"
)

type txContextKey struct{}

func withTx(ctx context.Context, tx *sql.Tx) context.Context {
	// The marker travels with the handle so code above the adapter can see that
	// it is on one transaction connection without reaching for the handle.
	return context.WithValue(ctxmeta.WithinTransaction(ctx), txContextKey{}, tx)
}

func txFromContext(ctx context.Context) *sql.Tx {
	if ctx == nil {
		return nil
	}
	tx, _ := ctx.Value(txContextKey{}).(*sql.Tx)
	return tx
}
