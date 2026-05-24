package postgres

import (
	"context"
	"database/sql"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/boms/backend/internal/port"
	"github.com/google/uuid"
)

type CustomerProfileRepository struct {
	queries *sqlcgen.Queries
}

func NewCustomerProfileRepository(pool *Pool) *CustomerProfileRepository {
	return &CustomerProfileRepository{queries: pool.Queries()}
}

func (r *CustomerProfileRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *CustomerProfileRepository) Create(ctx context.Context, params port.UpsertCustomerProfileParams) (*domainprofile.Customer, error) {
	row, err := r.q(ctx).CreateCustomerProfile(ctx, sqlcgen.CreateCustomerProfileParams{
		UserID:      params.UserID,
		DisplayName: toNullString(params.DisplayName),
		Phone:       toNullString(params.Phone),
	})
	if err != nil {
		return nil, mapUserQueryError(err, "create customer profile")
	}
	return mapCustomerProfile(row), nil
}

func (r *CustomerProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Customer, error) {
	row, err := r.q(ctx).GetCustomerProfileByUserID(ctx, userID)
	if err != nil {
		return nil, mapUserQueryError(err, "get customer profile")
	}
	return mapCustomerProfile(row), nil
}

func (r *CustomerProfileRepository) UpdateByUserID(ctx context.Context, params port.UpsertCustomerProfileParams) (*domainprofile.Customer, error) {
	row, err := r.q(ctx).UpdateCustomerProfileByUserID(ctx, sqlcgen.UpdateCustomerProfileByUserIDParams{
		UserID:      params.UserID,
		DisplayName: toNullString(params.DisplayName),
		Phone:       toNullString(params.Phone),
	})
	if err != nil {
		return nil, mapUserQueryError(err, "update customer profile")
	}
	return mapCustomerProfile(row), nil
}

func (r *CustomerProfileRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	rows, err := r.q(ctx).DeleteCustomerProfileByUserID(ctx, userID)
	if err != nil {
		return mapUserQueryError(err, "delete customer profile")
	}
	if rows == 0 {
		return mapUserQueryError(sql.ErrNoRows, "delete customer profile")
	}
	return nil
}

func mapCustomerProfile(row sqlcgen.CustomerProfile) *domainprofile.Customer {
	return &domainprofile.Customer{
		UserID:      row.UserID,
		DisplayName: nullStringPtr(row.DisplayName),
		Phone:       nullStringPtr(row.Phone),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toNullString(v *string) sql.NullString {
	if v == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *v, Valid: true}
}

var _ port.CustomerProfileRepository = (*CustomerProfileRepository)(nil)
