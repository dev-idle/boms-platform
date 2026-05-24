package postgres

import (
	"context"
	"database/sql"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type AdminProfileRepository struct {
	queries *sqlcgen.Queries
}

func NewAdminProfileRepository(pool *Pool) *AdminProfileRepository {
	return &AdminProfileRepository{queries: pool.Queries()}
}

func (r *AdminProfileRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *AdminProfileRepository) Create(ctx context.Context, params port.UpsertAdminProfileParams) (*domainprofile.Admin, error) {
	row, err := r.q(ctx).CreateAdminProfile(ctx, sqlcgen.CreateAdminProfileParams{
		UserID:   params.UserID,
		FullName: params.FullName,
		Phone:    toNullString(params.Phone),
	})
	if err != nil {
		return nil, mapUserQueryError(err, "create admin profile")
	}
	return mapAdminProfile(row), nil
}

func (r *AdminProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Admin, error) {
	row, err := r.q(ctx).GetAdminProfileByUserID(ctx, userID)
	if err != nil {
		return nil, mapUserQueryError(err, "get admin profile")
	}
	return mapAdminProfile(row), nil
}

func (r *AdminProfileRepository) UpdateByUserID(ctx context.Context, params port.UpsertAdminProfileParams) (*domainprofile.Admin, error) {
	row, err := r.q(ctx).UpdateAdminProfileByUserID(ctx, sqlcgen.UpdateAdminProfileByUserIDParams{
		UserID:   params.UserID,
		FullName: params.FullName,
		Phone:    toNullString(params.Phone),
	})
	if err != nil {
		return nil, mapUserQueryError(err, "update admin profile")
	}
	return mapAdminProfile(row), nil
}

func (r *AdminProfileRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	rows, err := r.q(ctx).DeleteAdminProfileByUserID(ctx, userID)
	if err != nil {
		return mapUserQueryError(err, "delete admin profile")
	}
	if rows == 0 {
		return mapUserQueryError(sql.ErrNoRows, "delete admin profile")
	}
	return nil
}

func mapAdminProfile(row sqlcgen.AdminProfile) *domainprofile.Admin {
	return &domainprofile.Admin{
		UserID:    row.UserID,
		FullName:  row.FullName,
		Phone:     nullStringPtr(row.Phone),
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}
}

var _ port.AdminProfileRepository = (*AdminProfileRepository)(nil)
