package usecase_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainprofile "github.com/boms/backend/internal/domain/profile"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
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
		uc := usecase.NewMeUsecase(users, customers, nil, nil, nil, nil, passthroughTxManager{}, nil, auditlogger.NewService(audit), nil)
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

	t.Run("promotions_change_only_when_asked", func(t *testing.T) {
		t.Parallel()
		agreedAt := time.Now()
		cases := []struct {
			name   string
			agreed *time.Time
			req    dto.UpdateMeRequest
			want   *bool
		}{
			{name: "left_as_it_is_when_left_out", agreed: &agreedAt, req: dto.UpdateMeRequest{DisplayName: &name}},
			{name: "withdrawn", agreed: &agreedAt, req: dto.UpdateMeRequest{MarketingOptIn: new(false)}, want: new(false)},
			{name: "given", req: dto.UpdateMeRequest{MarketingOptIn: new(true)}, want: new(true)},
		}
		for _, tc := range cases {
			users := new(mockUserRepo)
			customers := new(mockCustomerProfileRepo)
			userID := uuid.New()
			users.On("GetByID", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer}, nil)
			users.On("GetByIDForShare", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer}, nil)
			customers.On("GetByUserID", mock.Anything, userID).Return(&domainprofile.Customer{UserID: userID, MarketingConsentAt: tc.agreed}, nil)
			customers.On("UpdateByUserID", mock.Anything, mock.MatchedBy(func(p port.UpsertCustomerProfileParams) bool {
				return (p.MarketingOptIn == nil) == (tc.want == nil) && (tc.want == nil || *p.MarketingOptIn == *tc.want)
			})).Return(&domainprofile.Customer{UserID: userID}, nil)
			uc := usecase.NewMeUsecase(users, customers, nil, nil, nil, nil, passthroughTxManager{}, nil, auditlogger.NewService(&recordingAuditLogs{}), nil)

			_, _, err := uc.UpdateProfile(ctx, userID, tc.req)

			require.NoError(t, err, tc.name)
			customers.AssertExpectations(t)
		}
	})
}

func TestMeUsecase_ChangePassword(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	newFixture := func() (*usecase.MeUsecase, *mockUserRepo, *fakeTokens, *mockSessionStore, uuid.UUID) {
		users := new(mockUserRepo)
		hasher := new(mockHasher)
		sessions := new(mockSessionStore)
		tokens := &fakeTokens{}
		userID := uuid.New()
		users.On("GetByID", mock.Anything, userID).Return(&domainuser.User{ID: userID, Role: domainuser.RoleCustomer, PasswordHash: "old-hash"}, nil)
		hasher.On("Verify", "old-hash", "Old-Password-1").Return(nil)
		hasher.On("Hash", "New-Password-2").Return("new-hash", nil)
		return usecase.NewMeUsecase(users, nil, nil, nil, tokens, sessions, passthroughTxManager{}, hasher, nil, nil), users, tokens, sessions, userID
	}

	t.Run("a_new_password_voids_the_emailed_links", func(t *testing.T) {
		t.Parallel()
		uc, users, tokens, sessions, userID := newFixture()
		users.On("UpdatePassword", mock.Anything, userID, "new-hash").Return(nil)
		users.On("ClearMustChangePassword", mock.Anything, userID).Return(nil)
		sessions.On("DeleteAllForUser", mock.Anything, userID.String()).Return(nil)

		require.NoError(t, uc.ChangePassword(ctx, userID, "Old-Password-1", "New-Password-2"))

		assert.Equal(t, []uuid.UUID{userID}, tokens.deleted, "a reset link sent before cannot undo the change")
		users.AssertExpectations(t)
		sessions.AssertExpectations(t)
	})

	t.Run("links_that_cannot_be_voided_stop_the_change", func(t *testing.T) {
		t.Parallel()
		uc, users, tokens, sessions, userID := newFixture()
		tokens.deleteErr = apperrors.ErrInternal

		require.ErrorIs(t, uc.ChangePassword(ctx, userID, "Old-Password-1", "New-Password-2"), apperrors.ErrInternal)

		users.AssertNotCalled(t, "UpdatePassword", mock.Anything, mock.Anything, mock.Anything)
		sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})
}
