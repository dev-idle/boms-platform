package usecase_test

import (
	"context"
	"errors"
	"testing"

	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAdminUserUsecase_CreateOperationalUser(t *testing.T) {
	t.Parallel()

	t.Run("duplicate_email_reports_email_exists", func(t *testing.T) {
		t.Parallel()
		users := new(mockUserRepo)
		hasher := new(mockHasher)
		uc := usecase.NewAdminUserUsecase(users, nil, nil, nil, nil, passthroughTxManager{}, hasher, nil, nil, nil)

		hasher.On("Hash", mock.Anything).Return("temp-hash", nil).Once()
		users.On("AdminCreate", mock.Anything, mock.Anything).Return(nil, apperrors.ErrConflict).Once()

		_, err := uc.CreateOperationalUser(context.Background(), uuid.New(), domainuser.RoleAdmin, dto.CreateOperationalUserRequest{
			Email:    "taken@example.com",
			Role:     string(domainuser.RoleManager),
			FullName: "Taken Name",
		})

		require.Error(t, err)
		assert.True(t, errors.Is(err, usecase.ErrEmailExists), "got %v", err)
		assert.False(t, errors.Is(err, domainuser.ErrEmployeeCodeExists))
		users.AssertExpectations(t)
	})
}
