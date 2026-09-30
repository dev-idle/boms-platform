package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

type DiscountCodeRepository struct {
	queries *sqlcgen.Queries
}

func NewDiscountCodeRepository(pool *Pool) *DiscountCodeRepository {
	return &DiscountCodeRepository{queries: pool.Queries()}
}

func (r *DiscountCodeRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *DiscountCodeRepository) Create(
	ctx context.Context,
	params port.CreateDiscountCodeParams,
) (*domaindiscount.Code, error) {
	discountType, err := mapDiscountTypeToSQL(params.DiscountType)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).CreateDiscountCode(ctx, sqlcgen.CreateDiscountCodeParams{
		Code:               params.Code,
		DiscountType:       discountType,
		Value:              params.Value,
		MinOrderCents:      params.MinOrderCents,
		MaxUses:            params.MaxUses,
		MaxUsesPerCustomer: params.MaxUsesPerCustomer,
		MaxDiscountCents:   params.MaxDiscountCents,
		StartsAt:           params.StartsAt,
		EndsAt:             params.EndsAt,
		IsActive:           params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "create discount code")
	}
	return mapDiscountCode(row)
}

func (r *DiscountCodeRepository) GetByCode(ctx context.Context, code string) (*domaindiscount.Code, error) {
	row, err := r.q(ctx).GetDiscountCodeByCode(ctx, code)
	if err != nil {
		return nil, mapRepoError(err, "get discount code by code")
	}
	return mapDiscountCode(row)
}

// ReleaseUse implements port.DiscountCodeRepository.
func (r *DiscountCodeRepository) ReleaseUse(ctx context.Context, id uuid.UUID) error {
	if _, err := r.q(ctx).ReleaseDiscountCodeUse(ctx, id); err != nil {
		return mapRepoError(err, "release discount code use")
	}
	return nil
}

func (r *DiscountCodeRepository) IncrementUsedCount(ctx context.Context, id uuid.UUID) (*domaindiscount.Code, error) {
	row, err := r.q(ctx).IncrementDiscountCodeUsedCount(ctx, id)
	if err != nil {
		mapped := mapRepoError(err, "increment discount used count")
		if errors.Is(mapped, apperrors.ErrNotFound) {
			return nil, domaindiscount.ErrExhausted
		}
		return nil, mapped
	}
	return mapDiscountCode(row)
}

func (r *DiscountCodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domaindiscount.Code, error) {
	row, err := r.q(ctx).GetDiscountCodeByID(ctx, id)
	if err != nil {
		return nil, mapRepoError(err, "get discount code")
	}
	return mapDiscountCode(row)
}

func (r *DiscountCodeRepository) Update(
	ctx context.Context,
	params port.UpdateDiscountCodeParams,
) (*domaindiscount.Code, error) {
	discountType, err := mapDiscountTypeToSQL(params.DiscountType)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).UpdateDiscountCode(ctx, sqlcgen.UpdateDiscountCodeParams{
		ID:                 params.ID,
		Code:               params.Code,
		DiscountType:       discountType,
		Value:              params.Value,
		MinOrderCents:      params.MinOrderCents,
		MaxUses:            params.MaxUses,
		MaxUsesPerCustomer: params.MaxUsesPerCustomer,
		MaxDiscountCents:   params.MaxDiscountCents,
		StartsAt:           params.StartsAt,
		EndsAt:             params.EndsAt,
		IsActive:           params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "update discount code")
	}
	return mapDiscountCode(row)
}

func (r *DiscountCodeRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	rows, err := r.q(ctx).SoftDeleteDiscountCode(ctx, id)
	if err != nil {
		return mapRepoError(err, "delete discount code")
	}
	if rows == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *DiscountCodeRepository) ManagerList(
	ctx context.Context,
	params port.ManagerListDiscountCodesParams,
) ([]domaindiscount.Code, error) {
	rows, err := r.q(ctx).ManagerListDiscountCodes(ctx, sqlcgen.ManagerListDiscountCodesParams{
		Limit:  params.Limit,
		Offset: params.Offset,
		Search: optionalSearch(params.Search),
	})
	if err != nil {
		return nil, mapRepoError(err, "manager list discount codes")
	}
	out := make([]domaindiscount.Code, 0, len(rows))
	for _, row := range rows {
		code, mapErr := mapDiscountCode(row)
		if mapErr != nil {
			return nil, mapErr
		}
		out = append(out, *code)
	}
	return out, nil
}

func (r *DiscountCodeRepository) ManagerListCount(ctx context.Context, search *string) (int64, error) {
	count, err := r.q(ctx).ManagerListDiscountCodesCount(ctx, optionalSearch(search))
	if err != nil {
		return 0, mapRepoError(err, "manager list discount codes count")
	}
	return count, nil
}

func mapDiscountCode(row sqlcgen.DiscountCode) (*domaindiscount.Code, error) {
	discountType, err := mapDiscountTypeFromSQL(row.DiscountType)
	if err != nil {
		return nil, err
	}
	return &domaindiscount.Code{
		ID:                 row.ID,
		Code:               row.Code,
		DiscountType:       discountType,
		Value:              row.Value,
		UsedCount:          row.UsedCount,
		StartsAt:           row.StartsAt,
		EndsAt:             row.EndsAt,
		IsActive:           row.IsActive,
		CreatedAt:          row.CreatedAt,
		UpdatedAt:          row.UpdatedAt,
		MinOrderCents:      row.MinOrderCents,
		MaxUses:            row.MaxUses,
		MaxUsesPerCustomer: row.MaxUsesPerCustomer,
		MaxDiscountCents:   row.MaxDiscountCents,
		DeletedAt:          row.DeletedAt,
	}, nil
}

func mapDiscountTypeToSQL(t domaindiscount.Type) (sqlcgen.DiscountType, error) {
	switch t {
	case domaindiscount.TypePercent:
		return sqlcgen.DiscountTypePercent, nil
	case domaindiscount.TypeFixedCents:
		return sqlcgen.DiscountTypeFixedCents, nil
	default:
		return "", apperrors.Errorf("unsupported discount type: %s", t)
	}
}

func mapDiscountTypeFromSQL(t sqlcgen.DiscountType) (domaindiscount.Type, error) {
	switch t {
	case sqlcgen.DiscountTypePercent:
		return domaindiscount.TypePercent, nil
	case sqlcgen.DiscountTypeFixedCents:
		return domaindiscount.TypeFixedCents, nil
	default:
		return "", apperrors.Errorf("unsupported discount type from db: %s", t)
	}
}
