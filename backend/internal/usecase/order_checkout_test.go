package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

var _ port.TxManager = (*countingTxManager)(nil)

// countingTxManager records whether checkout got as far as writing anything.
type countingTxManager struct{ calls int }

func (c *countingTxManager) WithTx(context.Context, func(context.Context) error) error {
	c.calls++
	return errors.New("checkout reached its transaction")
}

// unplacedOrders has no order for any checkout key; any other call panics.
type unplacedOrders struct{ port.OrderRepository }

func (unplacedOrders) GetByCheckoutKey(context.Context, uuid.UUID, uuid.UUID) (*domainorder.Order, error) {
	return nil, apperrors.ErrNotFound
}

// Checkout reads the pickup rules before it locks anything: when they cannot be
// read, it stops there. The rules themselves, which need the locked cart, are
// covered against Postgres in checkout_pickup_test.go.
func TestOrderUsecase_Checkout_StoreUnreachable(t *testing.T) {
	t.Parallel()
	store, tx := newMemoryStore(), &countingTxManager{}
	store.err = errors.New("database unavailable")
	uc := usecase.NewOrderUsecase(nil, unplacedOrders{}, nil, nil, nil, tx, &recordingOutbox{}, store, nil, nil, nil, domainorder.PickupCodes{})

	_, err := uc.Checkout(context.Background(), uuid.New(), uuid.New(), dto.CheckoutRequest{PickupAt: time.Now().Add(24 * time.Hour), TermsVersion: domainpolicy.TermsVersion})

	require.ErrorIs(t, err, store.err)
	assert.Zero(t, tx.calls)
}

// An order without the current policies accepted is refused before the rules
// are read or anything is locked.
func TestOrderUsecase_Checkout_WithoutTheCurrentTerms(t *testing.T) {
	t.Parallel()
	store, tx := newMemoryStore(), &countingTxManager{}
	store.err = errors.New("the rules must not be read")
	uc := usecase.NewOrderUsecase(nil, nil, nil, nil, nil, tx, &recordingOutbox{}, store, nil, nil, nil, domainorder.PickupCodes{})

	for _, version := range []string{"", "2025-01-01"} {
		_, err := uc.Checkout(context.Background(), uuid.New(), uuid.New(), dto.CheckoutRequest{
			PickupAt: time.Now().Add(24 * time.Hour), TermsVersion: version,
		})

		require.ErrorIs(t, err, domainpolicy.ErrTermsNotAccepted, "version %q", version)
	}
	assert.Zero(t, tx.calls)
}
