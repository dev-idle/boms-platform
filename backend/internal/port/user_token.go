package port

import (
	"context"

	"github.com/google/uuid"

	domainaccount "github.com/boms/backend/internal/domain/account"
)

// UserTokenRepository keeps the single-use links emailed to users, as hashes.
type UserTokenRepository interface {
	// Issue makes a new token for purpose, replacing the user's last one, and
	// returns it. Only its hash is stored; the token itself goes in the email.
	Issue(ctx context.Context, userID uuid.UUID, purpose domainaccount.Purpose) (string, error)
	// Redeem uses token up and returns its user. A token that is unknown,
	// expired, already used or of a closed account is apperrors.ErrNotFound.
	Redeem(ctx context.Context, token string, purpose domainaccount.Purpose) (uuid.UUID, error)
	// DeleteForUser removes every token the user holds.
	DeleteForUser(ctx context.Context, userID uuid.UUID) error
}
