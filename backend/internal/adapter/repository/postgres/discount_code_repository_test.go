package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	postgresadapter "github.com/boms/backend/internal/adapter/repository/postgres"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// A discount code mixes an enum, optional limits that must stay unset when left
// empty, and a usage counter that must stop at its cap under the atomic update.
func TestDiscountCodeRepository_Integration(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	pool := newIntegrationPool(t, 5)
	codes := postgresadapter.NewDiscountCodeRepository(pool)

	maxUses := int32(1)
	maxDiscount := int64(500)

	// newCode creates a live single-use percent code of its own for one subtest.
	newCode := func(t *testing.T, code string) *domaindiscount.Code {
		t.Helper()
		now := time.Now().UTC()
		created, err := codes.Create(ctx, port.CreateDiscountCodeParams{
			Code:             code,
			DiscountType:     domaindiscount.TypePercent,
			Value:            10,
			MaxUses:          &maxUses,
			MaxDiscountCents: &maxDiscount,
			StartsAt:         now.Add(-time.Hour),
			EndsAt:           now.Add(24 * time.Hour),
			IsActive:         true,
		})
		require.NoError(t, err)
		return created
	}

	t.Run("round_trips_the_type_and_the_optional_limits", func(t *testing.T) {
		newCode(t, "SPRING")
		got, err := codes.GetByCode(ctx, "SPRING")
		require.NoError(t, err)
		assert.Equal(t, domaindiscount.TypePercent, got.DiscountType)
		assert.Nil(t, got.MinOrderCents)
		require.NotNil(t, got.MaxUses)
		assert.Equal(t, maxUses, *got.MaxUses)
		require.NotNil(t, got.MaxDiscountCents)
		assert.Equal(t, maxDiscount, *got.MaxDiscountCents)
		assert.Nil(t, got.DeletedAt)
	})

	t.Run("stops_counting_at_the_usage_cap", func(t *testing.T) {
		code := newCode(t, "SUMMER")
		used, err := codes.IncrementUsedCount(ctx, code.ID)
		require.NoError(t, err)
		assert.Equal(t, int32(1), used.UsedCount)

		_, err = codes.IncrementUsedCount(ctx, code.ID)
		assert.ErrorIs(t, err, domaindiscount.ErrExhausted)
	})

	t.Run("hides_a_soft_deleted_code", func(t *testing.T) {
		code := newCode(t, "AUTUMN")
		require.NoError(t, codes.SoftDelete(ctx, code.ID))
		_, err := codes.GetByID(ctx, code.ID)
		assert.ErrorIs(t, err, apperrors.ErrNotFound)
	})
}
