package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserRepository implements port.UserRepository using sqlc-generated queries.
type UserRepository struct {
	queries *sqlcgen.Queries
}

// NewUserRepository returns a Postgres-backed user repository.
func NewUserRepository(pool *Pool) *UserRepository {
	return &UserRepository{queries: pool.Queries()}
}

// Create implements port.UserRepository.
func (r *UserRepository) Create(ctx context.Context, params port.CreateUserParams) (*domainuser.User, error) {
	role, err := toSQLRole(params.Role)
	if err != nil {
		return nil, err
	}
	row, err := r.queries.CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		Role:         role,
	})
	if err != nil {
		return nil, mapUserQueryError(err, "create user")
	}
	return mapUser(row), nil
}

// GetByEmail implements port.UserRepository.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapUserQueryError(err, "get user by email")
	}
	return mapUser(row), nil
}

// GetByID implements port.UserRepository.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	row, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return nil, mapUserQueryError(err, "get user by id")
	}
	return mapUser(row), nil
}

// UpdatePassword implements port.UserRepository.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	rows, err := r.queries.UpdateUserPassword(ctx, sqlcgen.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return mapUserQueryError(err, "update password")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// SoftDelete implements port.UserRepository.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.queries.SoftDeleteUser(ctx, id)
	if err != nil {
		return mapUserQueryError(err, "soft delete user")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func mapUserQueryError(err error, op string) error {
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return apperrors.ErrConflict
	}
	return fmt.Errorf("%s: %w", op, err)
}

func mapUser(row sqlcgen.User) *domainuser.User {
	user := &domainuser.User{
		ID:           row.ID,
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		Role:         fromSQLRole(row.Role),
		EmailVerified: row.EmailVerifiedAt.Valid,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.DeletedAt.Valid {
		t := row.DeletedAt.Time
		user.DeletedAt = &t
	}
	return user
}

func toSQLRole(role domainuser.Role) (sqlcgen.UserRole, error) {
	switch role {
	case domainuser.RoleCustomer:
		return sqlcgen.UserRoleCustomer, nil
	case domainuser.RoleAdmin:
		return sqlcgen.UserRoleAdmin, nil
	default:
		return "", fmt.Errorf("unsupported role for persistence: %q", role)
	}
}

func fromSQLRole(role sqlcgen.UserRole) domainuser.Role {
	return domainuser.Role(role)
}

var _ port.UserRepository = (*UserRepository)(nil)
