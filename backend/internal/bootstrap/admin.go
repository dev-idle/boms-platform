package bootstrap

import (
	"context"
	"errors"
	"strings"

	"github.com/boms/backend/internal/config"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

// EnsureDevAdmin creates configured development admin accounts at boot.
// Other roles are never auto-seeded — create them via register (customer) or admin UI (operational roles).
func EnsureDevAdmin(
	ctx context.Context,
	cfg *config.Config,
	users port.UserRepository,
	admins port.AdminProfileRepository,
	hasher port.PasswordHasher,
	tx port.TxManager,
) error {
	if cfg == nil || strings.ToLower(cfg.App.Env) != "development" {
		return nil
	}

	for _, seed := range cfg.Seed.DevAdmins() {
		if err := ensureDevAdminOne(ctx, users, admins, hasher, tx, seed); err != nil {
			return err
		}
	}
	return nil
}

func ensureDevAdminOne(
	ctx context.Context,
	users port.UserRepository,
	admins port.AdminProfileRepository,
	hasher port.PasswordHasher,
	tx port.TxManager,
	seed config.DevAdminSeed,
) error {
	email := utils.NormalizeEmail(seed.Email)
	password := strings.TrimSpace(seed.Password)
	if email == "" || password == "" {
		return nil
	}

	existing, err := users.GetByEmail(ctx, email)
	if err == nil && existing != nil {
		if existing.Role != domainuser.RoleAdmin {
			if err := users.UpdateRole(ctx, existing.ID, domainuser.RoleAdmin); err != nil {
				return err
			}
		}
		if _, err := admins.GetByUserID(ctx, existing.ID); err != nil {
			if !errors.Is(err, apperrors.ErrNotFound) {
				return err
			}
			_, err = admins.Create(ctx, port.UpsertAdminProfileParams{
				UserID:   existing.ID,
				FullName: seed.FullName,
				Phone:    normalizeSeedPhone(seed.Phone),
			})
			return err
		}
		return nil
	}
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}

	hash, err := hasher.Hash(password)
	if err != nil {
		return err
	}
	return tx.WithTx(ctx, func(txCtx context.Context) error {
		user, err := users.AdminCreate(txCtx, port.CreateUserParams{
			Email:              email,
			PasswordHash:       hash,
			Role:               domainuser.RoleAdmin,
			MustChangePassword: false,
		})
		if err != nil {
			return err
		}
		_, err = admins.Create(txCtx, port.UpsertAdminProfileParams{
			UserID:   user.ID,
			FullName: seed.FullName,
			Phone:    normalizeSeedPhone(seed.Phone),
		})
		return err
	})
}

func normalizeSeedPhone(v string) *string {
	v = strings.TrimSpace(v)
	if v == "" {
		return nil
	}
	return &v
}
