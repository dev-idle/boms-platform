package postgres

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainaccount "github.com/boms/backend/internal/domain/account"
	"github.com/boms/backend/internal/port"
)

// tokenBytes is the size of a link token: 256 random bits, beyond guessing.
const tokenBytes = 32

// UserTokenRepository keeps emailed link tokens as SHA-256 hashes, so neither
// the table nor a backup of it opens an account.
type UserTokenRepository struct {
	queries *sqlcgen.Queries
}

func NewUserTokenRepository(pool *Pool) *UserTokenRepository {
	return &UserTokenRepository{queries: pool.Queries()}
}

func (r *UserTokenRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Issue implements port.UserTokenRepository.
func (r *UserTokenRepository) Issue(ctx context.Context, userID uuid.UUID, purpose domainaccount.Purpose) (string, error) {
	raw := make([]byte, tokenBytes)
	if _, err := rand.Read(raw); err != nil {
		return "", fmt.Errorf("generate link token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	err := r.q(ctx).ReplaceUserToken(ctx, sqlcgen.ReplaceUserTokenParams{
		UserID:     userID,
		Purpose:    sqlcgen.UserTokenPurpose(purpose),
		TokenHash:  tokenHash(token),
		TtlSeconds: purpose.TTL().Seconds(),
	})
	if err != nil {
		return "", mapRepoError(err, "issue user token")
	}
	return token, nil
}

// Redeem implements port.UserTokenRepository.
func (r *UserTokenRepository) Redeem(ctx context.Context, token string, purpose domainaccount.Purpose) (uuid.UUID, error) {
	userID, err := r.q(ctx).RedeemUserToken(ctx, sqlcgen.RedeemUserTokenParams{
		TokenHash: tokenHash(token),
		Purpose:   sqlcgen.UserTokenPurpose(purpose),
	})
	if err != nil {
		return uuid.Nil, mapRepoError(err, "redeem user token")
	}
	return userID, nil
}

// DeleteForUser implements port.UserTokenRepository.
func (r *UserTokenRepository) DeleteForUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).DeleteUserTokens(ctx, userID); err != nil {
		return mapRepoError(err, "delete user tokens")
	}
	return nil
}

func tokenHash(token string) []byte {
	sum := sha256.Sum256([]byte(token))
	return sum[:]
}

var _ port.UserTokenRepository = (*UserTokenRepository)(nil)
