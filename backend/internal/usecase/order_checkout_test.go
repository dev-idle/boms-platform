package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/usecase"
)

var _ port.TxManager = (*countingTxManager)(nil)

// countingTxManager records whether checkout got as far as writing anything.
type countingTxManager struct{ calls int }

func (c *countingTxManager) WithTx(context.Context, func(context.Context) error) error {
	c.calls++
	return errors.New("checkout reached its transaction")
}

// Checkout reads the pickup rules before it locks anything: when they cannot be
// read, it stops there. The rules themselves, which need the locked cart, are
// covered against Postgres in checkout_pickup_test.go.
func TestOrderUsecase_Checkout_StoreUnreachable(t *testing.T) {
	t.Parallel()
	store, tx := newMemoryStore(), &countingTxManager{}
	store.err = errors.New("database unavailable")
	uc := usecase.NewOrderUsecase(nil, nil, nil, nil, tx, &recordingOutbox{}, store)

	_, err := uc.Checkout(context.Background(), uuid.New(), time.Now().Add(24*time.Hour))

	require.ErrorIs(t, err, store.err)
	assert.Zero(t, tx.calls)
}
