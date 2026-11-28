package usecase

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	domaindiscount "github.com/boms/backend/internal/domain/discount"
	"github.com/boms/backend/internal/dto"
)

func TestCartUsecase_ApplyDiscountRefusesAMisspelledCodeAsUnknown(t *testing.T) {
	t.Parallel()
	// The check comes before any repository call, so no dependency is needed.
	_, err := (&CartUsecase{}).ApplyDiscount(t.Context(), uuid.New(), dto.ApplyCartDiscountRequest{Code: "BAD CODE!"})
	assert.ErrorIs(t, err, domaindiscount.ErrNotFound)
}
