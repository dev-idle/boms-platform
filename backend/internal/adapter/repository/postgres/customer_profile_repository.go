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
		UserID:         params.UserID,
		DisplayName:    params.DisplayName,
		Phone:          params.Phone,
		MarketingOptIn: params.MarketingOptIn != nil && *params.MarketingOptIn,
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
		UserID:         params.UserID,
		DisplayName:    params.DisplayName,
		Phone:          params.Phone,
		MarketingOptIn: params.MarketingOptIn,
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
		UserID:             row.UserID,
		DisplayName:        row.DisplayName,
		Phone:              row.Phone,
		MarketingConsentAt: row.MarketingConsentAt,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
	}
}

// WithdrawMarketingConsent implements port.CustomerProfileRepository.
func (r *CustomerProfileRepository) WithdrawMarketingConsent(ctx context.Context, userID uuid.UUID) (bool, error) {
	n, err := r.q(ctx).WithdrawMarketingConsent(ctx, userID)
	if err != nil {
		return false, mapRepoError(err, "withdraw marketing consent")
	}
	return n > 0, nil
}

var _ port.CustomerProfileRepository = (*CustomerProfileRepository)(nil)

// Erase implements port.CustomerProfileRepository.
func (r *CustomerProfileRepository) Erase(ctx context.Context, userID uuid.UUID) error {
	if err := r.q(ctx).EraseCustomerProfile(ctx, userID); err != nil {
		return mapRepoError(err, "erase customer profile")
	}
	return nil
}
