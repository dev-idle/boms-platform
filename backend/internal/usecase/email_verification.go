package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// EmailVerificationUsecase confirms that an account's address belongs to its
// holder: they follow a single-use link emailed to it. Orders, and the emails
// about them, need a confirmed address.
type EmailVerificationUsecase struct {
	tx     port.TxManager
	users  port.UserRepository
	tokens port.UserTokenRepository
	outbox port.EventOutbox
	audit  *auditlogger.Service
}

func NewEmailVerificationUsecase(
	tx port.TxManager,
	users port.UserRepository,
	tokens port.UserTokenRepository,
	outbox port.EventOutbox,
	audit *auditlogger.Service,
) *EmailVerificationUsecase {
	return &EmailVerificationUsecase{tx: tx, users: users, tokens: tokens, outbox: outbox, audit: audit}
}

// Verify confirms the address of the account a link was sent to. The link
// works once; one that is unknown, expired, used or of a closed account is
// domainaccount.ErrLinkInvalid. Following it needs no sign-in: it is often
// opened on another device.
func (u *EmailVerificationUsecase) Verify(ctx context.Context, token string) error {
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		userID, err := u.tokens.Redeem(txCtx, token, domainaccount.PurposeVerifyEmail)
		if errors.Is(err, apperrors.ErrNotFound) {
			return domainaccount.ErrLinkInvalid
		}
		if err != nil {
			return err
		}
		// Closed since the link was redeemed.
		if err := u.users.MarkEmailVerified(txCtx, userID); errors.Is(err, apperrors.ErrNotFound) {
			return domainaccount.ErrLinkInvalid
		} else if err != nil {
			return err
		}
		user, err := u.users.GetByID(txCtx, userID)
		if err != nil {
			return err
		}
		if err := u.audit.Log(txCtx, domainuser.AuditActionMeVerifiedEmail, userID, user.Role, &userID, "user",
			nil, map[string]any{"email_verified": true}); err != nil {
			return apperrors.Errorf("record email verification: %w", err)
		}
		return nil
	})
}

// Resend asks for a new confirmation email for the signed-in user; the new
// link replaces the last. An address already confirmed gets nothing and is
// domainuser.ErrEmailAlreadyVerified, so the page can say so.
func (u *EmailVerificationUsecase) Resend(ctx context.Context, userID uuid.UUID) error {
	user, err := u.users.GetByID(ctx, userID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return ErrMeNotFound
	}
	if err != nil {
		return err
	}
	if user.EmailVerified {
		return domainuser.ErrEmailAlreadyVerified
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		return u.outbox.Add(txCtx, domainaccount.VerificationRequestedEvent(userID))
	})
}
