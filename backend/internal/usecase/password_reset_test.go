package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

var resetsPerAccount = port.QuotaLimit{Max: 3, Window: time.Hour}

type fakeQuota struct {
	usedUp bool
	err    error
	keys   []string
	limits []port.QuotaLimit
}

func (q *fakeQuota) Take(_ context.Context, key string, limit port.QuotaLimit) (bool, error) {
	q.keys = append(q.keys, key)
	q.limits = append(q.limits, limit)
	return !q.usedUp, q.err
}

type resetFixture struct {
	users    *mockUserRepo
	tokens   *fakeTokens
	outbox   *recordingOutbox
	sessions *mockSessionStore
	hasher   *mockHasher
	quota    *fakeQuota
	audit    *recordingAuditLogs
	user     *domainuser.User
}

func newResetFixture() *resetFixture {
	user := &domainuser.User{ID: uuid.New(), Email: "mai@example.com", Role: domainuser.RoleCustomer}
	return &resetFixture{
		users:    new(mockUserRepo),
		tokens:   &fakeTokens{owner: user.ID},
		outbox:   &recordingOutbox{},
		sessions: new(mockSessionStore),
		hasher:   new(mockHasher),
		quota:    &fakeQuota{},
		audit:    &recordingAuditLogs{},
		user:     user,
	}
}

func (f *resetFixture) usecase() *usecase.PasswordResetUsecase {
	return usecase.NewPasswordResetUsecase(passthroughTxManager{}, f.users, f.tokens, f.outbox, f.sessions, f.hasher,
		f.quota, resetsPerAccount, auditlogger.NewService(f.audit), zap.NewNop())
}

func TestPasswordResetUsecase_Request(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("an_account_gets_a_link", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.users.On("GetByEmail", mock.Anything, "mai@example.com").Return(f.user, nil)

		require.NoError(t, f.usecase().Request(ctx, "  Mai@Example.com "))

		require.Len(t, f.outbox.events, 1)
		userID, purpose, ok := domainaccount.EmailFor(f.outbox.events[0])
		require.True(t, ok)
		assert.Equal(t, f.user.ID, userID)
		assert.Equal(t, domainaccount.PurposeResetPassword, purpose)
		assert.Equal(t, []string{"password_reset:" + f.user.ID.String()}, f.quota.keys)
		assert.Equal(t, []port.QuotaLimit{resetsPerAccount}, f.quota.limits)
	})

	t.Run("an_unknown_address_is_answered_the_same_and_sends_nothing", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.users.On("GetByEmail", mock.Anything, "nobody@example.com").Return(nil, apperrors.ErrNotFound)

		require.NoError(t, f.usecase().Request(ctx, "nobody@example.com"))

		assert.Empty(t, f.outbox.events)
	})

	t.Run("an_account_that_used_up_its_links_is_sent_nothing", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.quota.usedUp = true
		f.users.On("GetByEmail", mock.Anything, "mai@example.com").Return(f.user, nil)

		require.NoError(t, f.usecase().Request(ctx, "mai@example.com"))

		assert.Empty(t, f.outbox.events, "nobody can fill a customer's inbox")
	})

	t.Run("a_failure_is_reported", func(t *testing.T) {
		t.Parallel()
		down := errors.New("redis unavailable")
		f := newResetFixture()
		f.quota.err = down
		f.users.On("GetByEmail", mock.Anything, "mai@example.com").Return(f.user, nil)

		require.ErrorIs(t, f.usecase().Request(ctx, "mai@example.com"), down)
		assert.Empty(t, f.outbox.events)

		lookup := newResetFixture()
		lookup.users.On("GetByEmail", mock.Anything, "mai@example.com").Return(nil, apperrors.ErrInternal)
		require.ErrorIs(t, lookup.usecase().Request(ctx, "mai@example.com"), apperrors.ErrInternal)

		unqueued := newResetFixture()
		unqueued.outbox.err = apperrors.ErrInternal
		unqueued.users.On("GetByEmail", mock.Anything, "mai@example.com").Return(unqueued.user, nil)
		require.ErrorIs(t, unqueued.usecase().Request(ctx, "mai@example.com"), apperrors.ErrInternal)
	})
}

func TestPasswordResetUsecase_Confirm(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	t.Run("sets_the_password_and_ends_every_session", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		f.users.On("ResetPassword", mock.Anything, f.user.ID, "new-hash").Return(nil)
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(f.user, nil)
		f.sessions.On("DeleteAllForUser", mock.Anything, f.user.ID.String()).Return(nil)

		require.NoError(t, f.usecase().Confirm(ctx, "reset-token", "New-Password-9"))

		assert.Equal(t, []string{"reset-token"}, f.tokens.redeemed)
		assert.Equal(t, []uuid.UUID{f.user.ID}, f.tokens.deleted, "the account's other links go with the old password")
		f.users.AssertExpectations(t)
		f.sessions.AssertNumberOfCalls(t, "DeleteAllForUser", 2)
		assert.Equal(t, []domainuser.AuditAction{domainuser.AuditActionMeResetPassword}, f.audit.actions)
	})

	t.Run("an_account_closed_since_the_link_was_redeemed_is_invalid", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		f.users.On("ResetPassword", mock.Anything, f.user.ID, "new-hash").Return(apperrors.ErrNotFound)

		require.ErrorIs(t, f.usecase().Confirm(ctx, "reset-token", "New-Password-9"), domainaccount.ErrLinkInvalid)
		assert.Empty(t, f.audit.actions)
		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("a_failure_is_reported_and_ends_nothing", func(t *testing.T) {
		t.Parallel()
		unvoided := newResetFixture()
		unvoided.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		unvoided.tokens.deleteErr = apperrors.ErrInternal
		require.ErrorIs(t, unvoided.usecase().Confirm(ctx, "reset-token", "New-Password-9"), apperrors.ErrInternal)
		unvoided.users.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything)

		unsaved := newResetFixture()
		unsaved.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		unsaved.users.On("ResetPassword", mock.Anything, unsaved.user.ID, "new-hash").Return(apperrors.ErrInternal)
		require.ErrorIs(t, unsaved.usecase().Confirm(ctx, "reset-token", "New-Password-9"), apperrors.ErrInternal)

		unread := newResetFixture()
		unread.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		unread.users.On("ResetPassword", mock.Anything, unread.user.ID, "new-hash").Return(nil)
		unread.users.On("GetByID", mock.Anything, unread.user.ID).Return(nil, apperrors.ErrInternal)
		require.ErrorIs(t, unread.usecase().Confirm(ctx, "reset-token", "New-Password-9"), apperrors.ErrInternal)

		unrecorded := newResetFixture()
		unrecorded.audit.failWith = apperrors.ErrInternal
		unrecorded.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		unrecorded.users.On("ResetPassword", mock.Anything, unrecorded.user.ID, "new-hash").Return(nil)
		unrecorded.users.On("GetByID", mock.Anything, unrecorded.user.ID).Return(unrecorded.user, nil)
		require.ErrorIs(t, unrecorded.usecase().Confirm(ctx, "reset-token", "New-Password-9"), apperrors.ErrInternal,
			"the reset commits with its record or not at all")

		for _, f := range []*resetFixture{unvoided, unsaved, unread, unrecorded} {
			f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
		}
	})

	t.Run("a_link_that_opens_nothing_is_invalid", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		f.tokens.redeemErr = apperrors.ErrNotFound

		require.ErrorIs(t, f.usecase().Confirm(ctx, "used", "New-Password-9"), domainaccount.ErrLinkInvalid)

		f.users.AssertNotCalled(t, "ResetPassword", mock.Anything, mock.Anything, mock.Anything)
		f.sessions.AssertNotCalled(t, "DeleteAllForUser", mock.Anything, mock.Anything)
	})

	t.Run("sessions_that_cannot_end_stop_the_reset", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		f.users.On("ResetPassword", mock.Anything, f.user.ID, "new-hash").Return(nil)
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(f.user, nil)
		down := errors.New("redis unavailable")
		f.sessions.On("DeleteAllForUser", mock.Anything, f.user.ID.String()).Return(down)

		require.ErrorIs(t, f.usecase().Confirm(ctx, "reset-token", "New-Password-9"), down,
			"the transaction rolls back and the link still works")
	})

	t.Run("a_late_sweep_that_fails_does_not_undo_the_reset", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.hasher.On("Hash", "New-Password-9").Return("new-hash", nil)
		f.users.On("ResetPassword", mock.Anything, f.user.ID, "new-hash").Return(nil)
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(f.user, nil)
		f.sessions.On("DeleteAllForUser", mock.Anything, f.user.ID.String()).Return(nil).Once()
		f.sessions.On("DeleteAllForUser", mock.Anything, f.user.ID.String()).Return(errors.New("redis blip")).Once()

		require.NoError(t, f.usecase().Confirm(ctx, "reset-token", "New-Password-9"))
	})

	t.Run("a_password_that_cannot_be_hashed_changes_nothing", func(t *testing.T) {
		t.Parallel()
		f := newResetFixture()
		f.hasher.On("Hash", "New-Password-9").Return("", errors.New("argon2 failed"))

		require.Error(t, f.usecase().Confirm(ctx, "reset-token", "New-Password-9"))
		assert.Empty(t, f.tokens.redeemed, "the link is not used up")
	})
}
