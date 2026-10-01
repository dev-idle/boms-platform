package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/boms/backend/internal/shared/ctxmeta"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

func TestDataExportUsecase_Export(t *testing.T) {
	t.Parallel()

	t.Run("refuses_to_run_inside_a_transaction", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		exports := usecase.NewDataExportUsecase(users, nil, nil, nil, nil, nil, nil, nil, nil)

		_, err := exports.Export(ctxmeta.WithinTransaction(context.Background()), uuid.New())

		require.Error(t, err)
		users.AssertNotCalled(t, "GetByID", mock.Anything, mock.Anything)
	})

	t.Run("an_account_that_is_gone_is_not_found", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		users.On("GetByID", mock.Anything, mock.Anything).Return(nil, apperrors.ErrNotFound)
		exports := usecase.NewDataExportUsecase(users, nil, nil, nil, nil, nil, nil, nil, nil)

		_, err := exports.Export(context.Background(), uuid.New())

		require.ErrorIs(t, err, usecase.ErrMeNotFound)
	})
}
