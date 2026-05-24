package postgres

import (
	"context"
	"database/sql"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type StaffProfileRepository struct {
	queries *sqlcgen.Queries
}

func NewStaffProfileRepository(pool *Pool) *StaffProfileRepository {
	return &StaffProfileRepository{queries: pool.Queries()}
}

func (r *StaffProfileRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *StaffProfileRepository) Create(ctx context.Context, params port.UpsertStaffProfileParams) (*domainprofile.Staff, error) {
	row, err := r.q(ctx).CreateStaffProfile(ctx, sqlcgen.CreateStaffProfileParams{
		UserID:       params.UserID,
		FullName:     params.FullName,
		Phone:        toNullString(params.Phone),
		EmployeeCode: params.EmployeeCode,
		HireDate:     params.HireDate,
		Shift:        params.Shift,
	})
	if err != nil {
		return nil, mapStaffError(err, "create staff profile")
	}
	return mapStaffProfile(row), nil
}

func (r *StaffProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Staff, error) {
	row, err := r.q(ctx).GetStaffProfileByUserID(ctx, userID)
	if err != nil {
		return nil, mapStaffError(err, "get staff profile")
	}
	return mapStaffProfile(row), nil
}

func (r *StaffProfileRepository) UpdateByUserID(ctx context.Context, params port.UpsertStaffProfileParams) (*domainprofile.Staff, error) {
	row, err := r.q(ctx).UpdateStaffProfileByUserID(ctx, sqlcgen.UpdateStaffProfileByUserIDParams{
		UserID:       params.UserID,
		FullName:     params.FullName,
		Phone:        toNullString(params.Phone),
		EmployeeCode: params.EmployeeCode,
		HireDate:     params.HireDate,
		Shift:        params.Shift,
	})
	if err != nil {
		return nil, mapStaffError(err, "update staff profile")
	}
	return mapStaffProfile(row), nil
}

func (r *StaffProfileRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	rows, err := r.q(ctx).DeleteStaffProfileByUserID(ctx, userID)
	if err != nil {
		return mapStaffError(err, "delete staff profile")
	}
	if rows == 0 {
		return mapStaffError(sql.ErrNoRows, "delete staff profile")
	}
	return nil
}

func mapStaffProfile(row sqlcgen.StaffProfile) *domainprofile.Staff {
	return &domainprofile.Staff{
		UserID:       row.UserID,
		FullName:     row.FullName,
		Phone:        nullStringPtr(row.Phone),
		EmployeeCode: row.EmployeeCode,
		HireDate:     row.HireDate,
		Shift:        row.Shift,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
}

func mapStaffError(err error, op string) error {
	return mapUserQueryError(err, op)
}

var _ port.StaffProfileRepository = (*StaffProfileRepository)(nil)
