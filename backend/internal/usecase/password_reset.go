package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainaccount "github.com/boms/backend/internal/domain/account"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// PasswordResetUsecase lets a user who forgot their password choose a new one
// through a single-use link emailed to their address.
type PasswordResetUsecase struct {
	tx       port.TxManager
	users    port.UserRepository
	tokens   port.UserTokenRepository
	outbox   port.EventOutbox
	sessions port.SessionStore
	hasher   port.PasswordHasher
	quota    port.Quota
	// perAccount caps the links one account is sent, whoever asks: anyone can
	// ask for an address, so without it a stranger could fill a customer's
	// inbox from many places at once.
	perAccount port.QuotaLimit
	audit      *auditlogger.Service
	log        *zap.Logger
}

func NewPasswordResetUsecase(
	tx port.TxManager,
	users port.UserRepository,
	tokens port.UserTokenRepository,
	outbox port.EventOutbox,
	sessions port.SessionStore,
	hasher port.PasswordHasher,
	quota port.Quota,
	perAccount port.QuotaLimit,
	audit *auditlogger.Service,
	log *zap.Logger,
) *PasswordResetUsecase {
	return &PasswordResetUsecase{
		tx: tx, users: users, tokens: tokens, outbox: outbox, sessions: sessions,
		hasher: hasher, quota: quota, perAccount: perAccount, audit: audit, log: log,
	}
}

// Request emails a reset link to the account at email, if there is an open
// one and it has not used up its links for now. The caller
// answers the same whatever happened, so the request tells nobody whether an
// address has an account.
func (u *PasswordResetUsecase) Request(ctx context.Context, email string) error {
	user, err := u.users.GetByEmail(ctx, utils.NormalizeEmail(email))
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	// Taken before the event is written: a write that fails costs the account
	// one of its links, never sends one past the limit.
	allowed, err := u.quota.Take(ctx, "password_reset:"+user.ID.String(), u.perAccount)
	if err != nil || !allowed {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		return u.outbox.Add(txCtx, domainaccount.PasswordResetRequestedEvent(user.ID))
	})
}

// Confirm sets the password of the account a reset link was sent to and ends
// every session it has, since whoever held the old password may be signed in.
// The link works once; one that is unknown, expired, used or of a closed
// account is domainaccount.ErrLinkInvalid.
func (u *PasswordResetUsecase) Confirm(ctx context.Context, token, newPassword string) error {
	hash, err := u.hasher.Hash(newPassword)
	if err != nil {
		return apperrors.Errorf("hash password: %w", err)
	}
	var userID uuid.UUID
	err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
		redeemed, err := u.tokens.Redeem(txCtx, token, domainaccount.PurposeResetPassword)
		if errors.Is(err, apperrors.ErrNotFound) {
			return domainaccount.ErrLinkInvalid
		}
		if err != nil {
			return err
		}
		userID = redeemed
		// The new password voids the account's other links too; a pending
		// confirmation is done, as this confirms the address. Links go before
		// the account, the order redeeming one takes them.
		if err := u.tokens.DeleteForUser(txCtx, userID); err != nil {
			return err
		}
		// Closed since the link was redeemed.
		if err := u.users.ResetPassword(txCtx, userID, hash); errors.Is(err, apperrors.ErrNotFound) {
			return domainaccount.ErrLinkInvalid
		} else if err != nil {
			return err
		}
		user, err := u.users.GetByID(txCtx, userID)
		if err != nil {
			return err
		}
		if err := u.audit.Log(txCtx, domainuser.AuditActionMeResetPassword, userID, user.Role, &userID, "user",
			nil, map[string]any{"changed": true}); err != nil {
			return apperrors.Errorf("record password reset: %w", err)
		}
		// Sessions end before the new password takes effect: if they cannot,
		// nothing changes and the link still works for another try.
		return u.sessions.DeleteAllForUser(txCtx, userID.String())
	})
	if err != nil {
		return err
	}
	// A sign-in with the old password between that and the commit would
	// outlive it; this ends it too.
	if err := u.sessions.DeleteAllForUser(ctx, userID.String()); err != nil {
		u.log.Warn("password_reset_session_sweep_failed", zap.String("user_id", userID.String()), zap.Error(err))
	}
	return nil
}
