package usecase

import (
	"testing"

	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateComboBundleSize(t *testing.T) {
	t.Parallel()

	productID := uuid.New().String()

	t.Run("rejects one product with quantity 1", func(t *testing.T) {
		t.Parallel()
		err := validateComboBundleSize([]dto.ComboItemInput{
			{ProductID: productID, Quantity: 1},
		})
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, apperrors.ErrValidation.Code, ae.Code)
		assert.Equal(t, comboBundleMinDetail, ae.Details["items"])
	})

	t.Run("accepts two products", func(t *testing.T) {
		t.Parallel()
		err := validateComboBundleSize([]dto.ComboItemInput{
			{ProductID: productID, Quantity: 1},
			{ProductID: uuid.New().String(), Quantity: 1},
		})
		require.NoError(t, err)
	})

	t.Run("accepts one product with quantity 2", func(t *testing.T) {
		t.Parallel()
		err := validateComboBundleSize([]dto.ComboItemInput{
			{ProductID: productID, Quantity: 2},
		})
		require.NoError(t, err)
	})
}

func TestParseComboItemsBundleSize(t *testing.T) {
	t.Parallel()

	productID := uuid.New().String()

	t.Run("surfaces bundle rule through parseComboItems", func(t *testing.T) {
		t.Parallel()
		_, _, err := parseComboItems([]dto.ComboItemInput{
			{ProductID: productID, Quantity: 1},
		})
		require.Error(t, err)
		ae, ok := apperrors.AsAppError(err)
		require.True(t, ok)
		assert.Equal(t, comboBundleMinDetail, ae.Details["items"])
	})
}
