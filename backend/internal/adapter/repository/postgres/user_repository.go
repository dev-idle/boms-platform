package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
)

// UserRepository implements port.UserRepository using sqlc-generated queries.
type UserRepository struct {
	queries *sqlcgen.Queries
}

// NewUserRepository returns a Postgres-backed user repository.
func NewUserRepository(pool *Pool) *UserRepository {
	return &UserRepository{queries: pool.Queries()}
}

func (r *UserRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// Create implements port.UserRepository.
func (r *UserRepository) Create(ctx context.Context, params port.CreateUserParams) (*domainuser.User, error) {
	role, err := toSQLRole(params.Role)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).CreateUser(ctx, sqlcgen.CreateUserParams{
		Email:              params.Email,
		PasswordHash:       params.PasswordHash,
		Role:               role,
		MustChangePassword: params.MustChangePassword,
	})
	if err != nil {
		return nil, mapRepoError(err, "create user")
	}
	return mapUserFields(
		row.ID, row.Email, row.PasswordHash, row.Role, row.EmailVerifiedAt, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	), nil
}

// AdminCreate implements port.UserRepository.
func (r *UserRepository) AdminCreate(ctx context.Context, params port.CreateUserParams) (*domainuser.User, error) {
	role, err := toSQLRole(params.Role)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).AdminCreate(ctx, sqlcgen.AdminCreateParams{
		Email:              params.Email,
		PasswordHash:       params.PasswordHash,
		Role:               role,
		MustChangePassword: params.MustChangePassword,
	})
	if err != nil {
		return nil, mapRepoError(err, "admin create user")
	}
	return mapUserFields(
		row.ID, row.Email, row.PasswordHash, row.Role, row.EmailVerifiedAt, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	), nil
}

// GetByEmail implements port.UserRepository.
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domainuser.User, error) {
	row, err := r.q(ctx).GetUserByEmail(ctx, email)
	if err != nil {
		return nil, mapRepoError(err, "get user by email")
	}
	return mapUserFields(
		row.ID, row.Email, row.PasswordHash, row.Role, row.EmailVerifiedAt, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	), nil
}

// GetByID implements port.UserRepository.
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	row, err := r.q(ctx).GetUserByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "get user by id")
	}
	return mapUserFields(
		row.ID, row.Email, row.PasswordHash, row.Role, row.EmailVerifiedAt, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	), nil
}

// GetByIDForUpdate implements port.UserRepository.
func (r *UserRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	row, err := r.q(ctx).GetUserByIDForUpdate(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "get user by id for update")
	}
	return mapUserFields(
		row.ID, row.Email, row.PasswordHash, row.Role, row.EmailVerifiedAt, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	), nil
}

// UpdatePassword implements port.UserRepository.
func (r *UserRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	rows, err := r.q(ctx).UpdateUserPassword(ctx, sqlcgen.UpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return mapRepoError(err, "update password")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// UpdateRole implements port.UserRepository.
func (r *UserRepository) UpdateRole(ctx context.Context, id uuid.UUID, role domainuser.Role) error {
	sqlRole, err := toSQLRole(role)
	if err != nil {
		return err
	}
	rows, err := r.q(ctx).UpdateRole(ctx, sqlcgen.UpdateRoleParams{
		ID:   id,
		Role: sqlRole,
	})
	if err != nil {
		return mapRepoError(err, "update user role")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// SetMustChangePassword implements port.UserRepository.
func (r *UserRepository) SetMustChangePassword(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).SetMustChangePassword(ctx, id)
	if err != nil {
		return mapRepoError(err, "set must change password")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// ClearMustChangePassword implements port.UserRepository.
func (r *UserRepository) ClearMustChangePassword(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).ClearMustChangePassword(ctx, id)
	if err != nil {
		return mapRepoError(err, "clear must change password")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// SoftDelete implements port.UserRepository.
func (r *UserRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).SoftDelete(ctx, id)
	if err != nil {
		return mapRepoError(err, "soft delete user")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// AdminGetByID implements port.UserRepository.
func (r *UserRepository) AdminGetByID(ctx context.Context, id uuid.UUID) (*domainuser.User, error) {
	row, err := r.q(ctx).AdminGetByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "admin get user by id")
	}
	return mapUserFields(
		row.ID, row.Email, row.PasswordHash, row.Role, row.EmailVerifiedAt, row.MustChangePassword, row.CreatedAt, row.UpdatedAt, row.DeletedAt,
	), nil
}

// Restore implements port.UserRepository.
func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).AdminRestore(ctx, id)
	if err != nil {
		return mapRepoError(err, "restore user")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// AdminUpdatePassword implements port.UserRepository.
func (r *UserRepository) AdminUpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	rows, err := r.q(ctx).AdminUpdateUserPassword(ctx, sqlcgen.AdminUpdateUserPasswordParams{
		ID:           id,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return mapRepoError(err, "admin update user password")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

// AdminList implements port.UserRepository.
func (r *UserRepository) AdminList(ctx context.Context, params port.AdminListUsersParams) ([]port.AdminListUser, int64, error) {
	roleFilter, err := adminListRoleFilter(params.Role)
	if err != nil {
		return nil, 0, err
	}
	rows, err := r.q(ctx).AdminList(ctx, sqlcgen.AdminListParams{
		Search:     params.Search,
		RoleFilter: roleFilter,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, 0, mapRepoError(err, "admin list users")
	}
	total, err := r.q(ctx).AdminListCount(ctx, sqlcgen.AdminListCountParams{
		Search:     params.Search,
		RoleFilter: roleFilter,
	})
	if err != nil {
		return nil, 0, mapRepoError(err, "admin list count")
	}

	out := make([]port.AdminListUser, 0, len(rows))
	for _, row := range rows {
		item := port.AdminListUser{
			ID:                 row.ID,
			Email:              row.Email,
			Role:               fromSQLRole(row.Role),
			EmailVerified:      row.EmailVerifiedAt.Valid,
			MustChangePassword: row.MustChangePassword,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
			FullName:           stringPtrOrNil(row.FullName),
			Phone:              nullStringPtr(row.Phone),
			EmployeeCode:       nullStringPtr(row.EmployeeCode),
			DisplayName:        nullStringPtr(row.DisplayName),
		}
		if row.DeletedAt.Valid {
			t := row.DeletedAt.Time
			item.DeletedAt = &t
		}
		out = append(out, item)
	}

	return out, total, nil
}

func mapUserFields(
	id uuid.UUID,
	email, passwordHash string,
	role sqlcgen.UserRole,
	emailVerifiedAt sql.NullTime,
	mustChangePassword bool,
	createdAt, updatedAt time.Time,
	deletedAt sql.NullTime,
) *domainuser.User {
	user := &domainuser.User{
		ID:                 id,
		Email:              email,
		PasswordHash:       passwordHash,
		Role:               fromSQLRole(role),
		EmailVerified:      emailVerifiedAt.Valid,
		MustChangePassword: mustChangePassword,
		CreatedAt:          createdAt,
		UpdatedAt:          updatedAt,
	}
	if deletedAt.Valid {
		t := deletedAt.Time
		user.DeletedAt = &t
	}
	return user
}

func toSQLRole(role domainuser.Role) (sqlcgen.UserRole, error) {
	switch role {
	case domainuser.RoleCustomer:
		return sqlcgen.UserRoleCustomer, nil
	case domainuser.RoleStaff:
		return sqlcgen.UserRoleStaff, nil
	case domainuser.RoleBaker:
		return sqlcgen.UserRoleBaker, nil
	case domainuser.RoleManager:
		return sqlcgen.UserRoleManager, nil
	case domainuser.RoleAdmin:
		return sqlcgen.UserRoleAdmin, nil
	default:
		return "", fmt.Errorf("unsupported role for persistence: %q", role)
	}
}

func fromSQLRole(role sqlcgen.UserRole) domainuser.Role {
	return domainuser.Role(role)
}

func adminListRoleFilter(role *domainuser.Role) (string, error) {
	if role == nil {
		return "", nil
	}
	if _, err := toSQLRole(*role); err != nil {
		return "", apperrors.ErrValidation.WithDetail("role", "unsupported role")
	}
	return string(*role), nil
}

func nullStringPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func stringPtrOrNil(v string) *string {
	if v == "" {
		return nil
	}
	s := v
	return &s
}

var _ port.UserRepository = (*UserRepository)(nil)
