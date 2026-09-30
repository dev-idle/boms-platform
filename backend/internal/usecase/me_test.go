package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

func TestMeUsecase_UpdateProfile(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	name := "Pat"

	newCustomer := func(audit *recordingAuditLogs) (*usecase.MeUsecase, *mockUserRepo, *mockCustomerProfileRepo, uuid.UUID) {
		users := new(mockUserRepo)
		customers := new(mockCustomerProfileRepo)
		userID := uuid.New()
		users.On("GetByID", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer}, nil)
		customers.On("GetByUserID", mock.Anything, userID).Return(&domainprofile.Customer{UserID: userID}, nil)
		uc := usecase.NewMeUsecase(users, customers, nil, nil, nil, passthroughTxManager{}, nil, auditlogger.NewService(audit), nil)
		return uc, users, customers, userID
	}

	t.Run("an_account_closed_meanwhile_is_not_written_to", func(t *testing.T) {
		t.Parallel()
		audit := &recordingAuditLogs{}
		uc, users, customers, userID := newCustomer(audit)
		users.On("GetByIDForShare", mock.Anything, userID).Return(nil, apperrors.ErrNotFound)

		_, _, err := uc.UpdateProfile(ctx, userID, dto.UpdateMeRequest{DisplayName: &name})

		require.ErrorIs(t, err, usecase.ErrMeNotFound)
		customers.AssertNotCalled(t, "UpdateByUserID", mock.Anything, mock.Anything)
		require.Empty(t, audit.actions)
	})

	t.Run("a_change_that_cannot_be_recorded_is_not_made", func(t *testing.T) {
		t.Parallel()
		audit := &recordingAuditLogs{failWith: apperrors.ErrInternal}
		uc, users, customers, userID := newCustomer(audit)
		users.On("GetByIDForShare", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer}, nil)
		customers.On("UpdateByUserID", mock.Anything, mock.Anything).Return(&domainprofile.Customer{UserID: userID, DisplayName: &name}, nil)

		_, _, err := uc.UpdateProfile(ctx, userID, dto.UpdateMeRequest{DisplayName: &name})

		require.ErrorIs(t, err, apperrors.ErrInternal, "the transaction rolls the change back")
	})
}
