package usecase_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/usecase"
)

type fakeAccountComposer struct {
	got []port.AccountEmail
	err error
}

func (c *fakeAccountComposer) ComposeAccountEmail(msg port.AccountEmail) (port.Email, error) {
	c.got = append(c.got, msg)
	return port.Email{To: msg.To, Subject: string(msg.Purpose)}, c.err
}

type accountEmailFixture struct {
	users    *mockUserRepo
	tokens   *fakeTokens
	composer *fakeAccountComposer
	mailer   *fakeMailer
	user     *domainuser.User
}

func newAccountEmailFixture() *accountEmailFixture {
	user := &domainuser.User{ID: uuid.New(), Email: "mai@example.com", Role: domainuser.RoleCustomer}
	f := &accountEmailFixture{
		users:    new(mockUserRepo),
		tokens:   &fakeTokens{next: "fresh-token"},
		composer: &fakeAccountComposer{},
		mailer:   &fakeMailer{},
		user:     user,
	}
	f.users.On("GetByID", mock.Anything, user.ID).Return(user, nil).Maybe()
	return f
}

func (f *accountEmailFixture) send(purpose domainaccount.Purpose) error {
	uc := usecase.NewAccountEmailUsecase(f.users, f.tokens, f.composer, f.mailer, zap.NewNop())
	return uc.Send(context.Background(), port.AccountEmailTask{EventID: uuid.New(), UserID: f.user.ID, Purpose: purpose})
}

func TestAccountEmailUsecase_Send(t *testing.T) {
	t.Parallel()

	t.Run("makes_a_fresh_link_and_mails_it", func(t *testing.T) {
		t.Parallel()
		for _, purpose := range []domainaccount.Purpose{domainaccount.PurposeVerifyEmail, domainaccount.PurposeResetPassword} {
			f := newAccountEmailFixture()

			require.NoError(t, f.send(purpose))

			assert.Equal(t, []domainaccount.Purpose{purpose}, f.tokens.issued)
			require.Len(t, f.composer.got, 1)
			assert.Equal(t, port.AccountEmail{Purpose: purpose, To: "mai@example.com", Token: "fresh-token"}, f.composer.got[0])
			assert.Len(t, f.mailer.sent, 1)
		}
	})

	t.Run("a_confirmed_address_gets_no_confirmation", func(t *testing.T) {
		t.Parallel()
		f := newAccountEmailFixture()
		f.user.EmailVerified = true

		require.NoError(t, f.send(domainaccount.PurposeVerifyEmail))
		assert.Empty(t, f.tokens.issued)
		assert.Empty(t, f.mailer.sent)

		require.NoError(t, f.send(domainaccount.PurposeResetPassword), "but can still reset its password")
		assert.Len(t, f.mailer.sent, 1)
	})

	t.Run("a_closed_account_gets_nothing", func(t *testing.T) {
		t.Parallel()
		f := newAccountEmailFixture()
		f.users.ExpectedCalls = nil
		f.users.On("GetByID", mock.Anything, f.user.ID).Return(nil, apperrors.ErrNotFound)

		require.NoError(t, f.send(domainaccount.PurposeResetPassword))
		assert.Empty(t, f.tokens.issued)
		assert.Empty(t, f.mailer.sent)
	})

	t.Run("a_passing_failure_is_left_to_retry", func(t *testing.T) {
		t.Parallel()
		for _, set := range []func(*accountEmailFixture){
			func(f *accountEmailFixture) {
				f.users.ExpectedCalls = nil
				f.users.On("GetByID", mock.Anything, f.user.ID).Return(nil, apperrors.ErrInternal)
			},
			func(f *accountEmailFixture) { f.tokens.issueErr = apperrors.ErrInternal },
			func(f *accountEmailFixture) { f.mailer.err = errors.New("mail server unreachable") },
		} {
			f := newAccountEmailFixture()
			set(f)

			err := f.send(domainaccount.PurposeResetPassword)

			require.Error(t, err)
			assert.NotErrorIs(t, err, port.ErrEmailUndeliverable)
		}
	})

	t.Run("an_email_no_retry_can_send_is_final", func(t *testing.T) {
		t.Parallel()
		unwritable := newAccountEmailFixture()
		unwritable.composer.err = errors.New("template broke")
		require.ErrorIs(t, unwritable.send(domainaccount.PurposeVerifyEmail), port.ErrEmailUndeliverable)
		assert.Empty(t, unwritable.mailer.sent)

		refused := newAccountEmailFixture()
		refused.mailer.err = fmt.Errorf("%w: 550", port.ErrEmailUndeliverable)
		require.ErrorIs(t, refused.send(domainaccount.PurposeVerifyEmail), port.ErrEmailUndeliverable)
	})
}
