package usecase_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

type verificationFixture struct {
	users  *mockUserRepo
	tokens *fakeTokens
	outbox *recordingOutbox
	audit  *recordingAuditLogs
	user   *domainuser.User
}

func newVerificationFixture() *verificationFixture {
	user := &domainuser.User{ID: uuid.New(), Role: domainuser.RoleCustomer}
	return &verificationFixture{
		users:  new(mockUserRepo),
		tokens: &fakeTokens{owner: user.ID},
		outbox: &recordingOutbox{},
		audit:  &recordingAuditLogs{},
		user:   user,
	}
}

func (f *verificationFixture) usecase() *usecase.EmailVerificationUsecase {
	return usecase.NewEmailVerificationUsecase(passthroughTxManager{}, f.users, f.tokens, f.outbox, auditlogger.NewService(f.audit))
}

func TestEmailVerificationUsecase_Verify(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("confirms_the_address_the_link_was_sent_to", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.users.On("MarkEmailVerified", mock.Anything, f.user.ID).Return(nil)
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(f.user, nil)

		require.NoError(t, f.usecase().Verify(ctx, "link-token"))

		assert.Equal(t, []string{"link-token"}, f.tokens.redeemed)
		f.users.AssertExpectations(t)
		assert.Equal(t, []domainuser.AuditAction{domainuser.AuditActionMeVerifiedEmail}, f.audit.actions)
	})

	t.Run("a_link_that_opens_nothing_is_invalid", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.tokens.redeemErr = apperrors.ErrNotFound

		require.ErrorIs(t, f.usecase().Verify(ctx, "used-or-expired"), domainaccount.ErrLinkInvalid)

		f.users.AssertNotCalled(t, "MarkEmailVerified", mock.Anything, mock.Anything)
		assert.Empty(t, f.audit.actions)
	})

	t.Run("an_account_closed_since_the_link_was_redeemed_is_invalid", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.users.On("MarkEmailVerified", mock.Anything, f.user.ID).Return(apperrors.ErrNotFound)

		require.ErrorIs(t, f.usecase().Verify(ctx, "link-token"), domainaccount.ErrLinkInvalid)
		assert.Empty(t, f.audit.actions)
	})

	t.Run("a_failure_is_reported_not_disguised", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.tokens.redeemErr = apperrors.ErrInternal
		require.ErrorIs(t, f.usecase().Verify(ctx, "t"), apperrors.ErrInternal)

		unmarked := newVerificationFixture()
		unmarked.users.On("MarkEmailVerified", mock.Anything, unmarked.user.ID).Return(apperrors.ErrInternal)
		require.ErrorIs(t, unmarked.usecase().Verify(ctx, "t"), apperrors.ErrInternal)

		unread := newVerificationFixture()
		unread.users.On("MarkEmailVerified", mock.Anything, unread.user.ID).Return(nil)
		unread.users.On("GetByID", mock.Anything, unread.user.ID).Return(nil, apperrors.ErrInternal)
		require.ErrorIs(t, unread.usecase().Verify(ctx, "t"), apperrors.ErrInternal)
		assert.Empty(t, unread.audit.actions)

		unrecorded := newVerificationFixture()
		unrecorded.audit.failWith = apperrors.ErrInternal
		unrecorded.users.On("MarkEmailVerified", mock.Anything, unrecorded.user.ID).Return(nil)
		unrecorded.users.On("GetByID", mock.Anything, unrecorded.user.ID).Return(unrecorded.user, nil)
		require.Error(t, unrecorded.usecase().Verify(ctx, "t"), "the confirmation commits with its record or not at all")
	})
}

func TestEmailVerificationUsecase_Resend(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("asks_for_a_new_link", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(f.user, nil)

		require.NoError(t, f.usecase().Resend(ctx, f.user.ID))

		require.Len(t, f.outbox.events, 1)
		userID, purpose, ok := domainaccount.EmailFor(f.outbox.events[0])
		require.True(t, ok)
		assert.Equal(t, f.user.ID, userID)
		assert.Equal(t, domainaccount.PurposeVerifyEmail, purpose)
	})

	t.Run("a_confirmed_address_gets_nothing", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.user.EmailVerified = true
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(f.user, nil)

		require.ErrorIs(t, f.usecase().Resend(ctx, f.user.ID), domainuser.ErrEmailAlreadyVerified)

		assert.Empty(t, f.outbox.events)
	})

	t.Run("a_closed_account_is_not_found", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(nil, apperrors.ErrNotFound)

		require.ErrorIs(t, f.usecase().Resend(ctx, f.user.ID), usecase.ErrMeNotFound)
		assert.Empty(t, f.outbox.events)
	})

	t.Run("a_failure_is_reported_not_disguised", func(t *testing.T) {
		t.Parallel()
		f := newVerificationFixture()
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(nil, apperrors.ErrInternal)
		require.ErrorIs(t, f.usecase().Resend(ctx, f.user.ID), apperrors.ErrInternal)

		unqueued := newVerificationFixture()
		unqueued.outbox.err = apperrors.ErrInternal
		unqueued.users.On("GetByID", mock.Anything, unqueued.user.ID).Return(unqueued.user, nil)
		require.ErrorIs(t, unqueued.usecase().Resend(ctx, unqueued.user.ID), apperrors.ErrInternal)
	})
}
