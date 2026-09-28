package postgres

import (
	"context"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainprofile "github.com/boms/backend/internal/domain/profile"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
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
		DisplayName: params.DisplayName,
		Phone:       params.Phone,
	})
	if err != nil {
		return nil, mapRepoError(err, "create customer profile")
	}
	return mapCustomerProfile(row), nil
}

func (r *CustomerProfileRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainprofile.Customer, error) {
	row, err := r.q(ctx).GetCustomerProfileByUserID(ctx, userID)
	if err != nil {
		return nil, mapRepoError(err, "get customer profile")
	}
	return mapCustomerProfile(row), nil
}

func (r *CustomerProfileRepository) UpdateByUserID(ctx context.Context, params port.UpsertCustomerProfileParams) (*domainprofile.Customer, error) {
	row, err := r.q(ctx).UpdateCustomerProfileByUserID(ctx, sqlcgen.UpdateCustomerProfileByUserIDParams{
		UserID:      params.UserID,
		DisplayName: params.DisplayName,
		Phone:       params.Phone,
	})
	if err != nil {
		return nil, mapRepoError(err, "update customer profile")
	}
	return mapCustomerProfile(row), nil
}

func (r *CustomerProfileRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	rows, err := r.q(ctx).DeleteCustomerProfileByUserID(ctx, userID)
	if err != nil {
		return mapRepoError(err, "delete customer profile")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func mapCustomerProfile(row sqlcgen.CustomerProfile) *domainprofile.Customer {
	return &domainprofile.Customer{
		UserID:      row.UserID,
		DisplayName: row.DisplayName,
		Phone:       row.Phone,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

var _ port.CustomerProfileRepository = (*CustomerProfileRepository)(nil)
