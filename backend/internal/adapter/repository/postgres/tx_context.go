package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"

	domainevent "github.com/boms/backend/internal/domain/event"
	"github.com/boms/backend/internal/shared/ctxmeta"
)

type txContextKey struct{}

// txScope is one transaction and the events recorded in it. The events are
// handed to the pool's commit hook only once the transaction has committed, so
// nothing is announced for a change that rolled back.
type txScope struct {
	tx     pgx.Tx
	events []domainevent.Event
}

func withTx(ctx context.Context, scope *txScope) context.Context {
	// The marker travels with the handle so code above the adapter can see that
	// it is on one transaction connection without reaching for the handle.
	return context.WithValue(ctxmeta.WithinTransaction(ctx), txContextKey{}, scope)
}

func scopeFromContext(ctx context.Context) *txScope {
	if ctx == nil {
		return nil
	}
	scope, _ := ctx.Value(txContextKey{}).(*txScope)
	return scope
}

func txFromContext(ctx context.Context) pgx.Tx {
	if scope := scopeFromContext(ctx); scope != nil {
		return scope.tx
	}
	return nil
}
