package usecase

import (
	"context"
	"errors"

	domainpromotion "github.com/boms/backend/internal/domain/promotion"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// UnsubscribeUsecase is a customer withdrawing their agreement to promotions
// from the link in a promotion email, without signing in.
type UnsubscribeUsecase struct {
	users     port.UserRepository
	customers port.CustomerProfileRepository
	tx        port.TxManager
	audit     *auditlogger.Service
	tokens    domainpromotion.UnsubscribeTokens
}

func NewUnsubscribeUsecase(
	users port.UserRepository,
	customers port.CustomerProfileRepository,
	tx port.TxManager,
	audit *auditlogger.Service,
	tokens domainpromotion.UnsubscribeTokens,
) *UnsubscribeUsecase {
	return &UnsubscribeUsecase{users: users, customers: customers, tx: tx, audit: audit, tokens: tokens}
}

// Unsubscribe withdraws the agreement of the customer token names, even while
// an administrator has their account disabled, so it holds if the account is
// enabled again. One that withdrew already, or whose account is erased, has
// nothing left to stop. The account is locked first, as an erasure locks it,
// so the two never interleave and the erasure scrubs this record too.
func (u *UnsubscribeUsecase) Unsubscribe(ctx context.Context, token string) error {
	userID, err := u.tokens.UserOf(token)
	if err != nil {
		return err
	}
	return u.tx.WithTx(ctx, func(txCtx context.Context) error {
		account, err := u.users.AdminGetByIDForUpdate(txCtx, userID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil
		}
		if err != nil {
			return err
		}
		if account.Erased() {
			return nil
		}
		withdrawn, err := u.customers.WithdrawMarketingConsent(txCtx, userID)
		if err != nil || !withdrawn {
			return err
		}
		if err := u.audit.Log(txCtx, domainpromotion.AuditActionMeUnsubscribed, userID, domainuser.RoleCustomer, &userID, "user_profile", nil, nil); err != nil {
			return apperrors.Errorf("record unsubscribe: %w", err)
		}
		return nil
	})
}
