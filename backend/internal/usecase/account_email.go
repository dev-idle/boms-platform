package usecase

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"

	domainaccount "github.com/boms/backend/internal/domain/account"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// AccountEmailUsecase sends a confirmation or a password reset link. It runs
// in the worker and makes the link as it sends it, so the link exists only in
// the email and, hashed, in the database — never in an event or a queue. A
// retry makes a new link, which replaces the one before.
type AccountEmailUsecase struct {
	users    port.UserRepository
	tokens   port.UserTokenRepository
	composer port.AccountEmailComposer
	mailer   port.Mailer
	log      *zap.Logger
}

func NewAccountEmailUsecase(
	users port.UserRepository,
	tokens port.UserTokenRepository,
	composer port.AccountEmailComposer,
	mailer port.Mailer,
	log *zap.Logger,
) *AccountEmailUsecase {
	return &AccountEmailUsecase{users: users, tokens: tokens, composer: composer, mailer: mailer, log: log}
}

// Send emails task's link to the account's address. It skips an account that
// is closed, and a confirmation for an address already confirmed. Errors
// wrapping port.ErrEmailUndeliverable are final; any other may pass on a retry.
func (u *AccountEmailUsecase) Send(ctx context.Context, task port.AccountEmailTask) error {
	user, err := u.users.GetByID(ctx, task.UserID)
	if errors.Is(err, apperrors.ErrNotFound) {
		u.skip(task, "account_gone")
		return nil
	}
	if err != nil {
		return err
	}
	if task.Purpose == domainaccount.PurposeVerifyEmail && user.EmailVerified {
		u.skip(task, "already_verified")
		return nil
	}
	token, err := u.tokens.Issue(ctx, user.ID, task.Purpose)
	if err != nil {
		return err
	}
	email, err := u.composer.ComposeAccountEmail(port.AccountEmail{Purpose: task.Purpose, To: user.Email, Token: token})
	if err != nil {
		return fmt.Errorf("%w: compose %s email: %w", port.ErrEmailUndeliverable, task.Purpose, err)
	}
	if err := u.mailer.Send(ctx, email); err != nil {
		return err
	}
	u.log.Info("account_email_sent",
		zap.String("user_id", task.UserID.String()),
		zap.String("purpose", string(task.Purpose)),
	)
	return nil
}

func (u *AccountEmailUsecase) skip(task port.AccountEmailTask, reason string) {
	u.log.Info("account_email_skipped",
		zap.String("user_id", task.UserID.String()),
		zap.String("purpose", string(task.Purpose)),
		zap.String("reason", reason),
	)
}
