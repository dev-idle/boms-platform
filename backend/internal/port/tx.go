package port

import "context"

// TxManager executes function bodies inside a database transaction.
type TxManager interface {
	WithTx(ctx context.Context, fn func(txCtx context.Context) error) error
}
