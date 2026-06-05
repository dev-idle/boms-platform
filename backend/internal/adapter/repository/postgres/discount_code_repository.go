package postgres

import (
	"context"
	"database/sql"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaindiscount "github.com/boms/backend/internal/domain/discount"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/google/uuid"
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
		Code:          params.Code,
		DiscountType:  discountType,
		Value:         params.Value,
		MinOrderCents: optionalInt64(params.MinOrderCents),
		MaxUses:       optionalInt32(params.MaxUses),
		StartsAt:      params.StartsAt,
		EndsAt:        params.EndsAt,
		IsActive:      params.IsActive,
	})
	if err != nil {
		return nil, mapRepoError(err, "create discount code")
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
		ID:            params.ID,
		Code:          params.Code,
		DiscountType:  discountType,
		Value:         params.Value,
		MinOrderCents: optionalInt64(params.MinOrderCents),
		MaxUses:       optionalInt32(params.MaxUses),
		StartsAt:      params.StartsAt,
		EndsAt:        params.EndsAt,
		IsActive:      params.IsActive,
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
	c := &domaindiscount.Code{
		ID:           row.ID,
		Code:         row.Code,
		DiscountType: discountType,
		Value:        row.Value,
		UsedCount:    row.UsedCount,
		StartsAt:     row.StartsAt,
		EndsAt:       row.EndsAt,
		IsActive:     row.IsActive,
		CreatedAt:    row.CreatedAt,
		UpdatedAt:    row.UpdatedAt,
	}
	if row.MinOrderCents.Valid {
		c.MinOrderCents = &row.MinOrderCents.Int64
	}
	if row.MaxUses.Valid {
		c.MaxUses = &row.MaxUses.Int32
	}
	if row.DeletedAt.Valid {
		c.DeletedAt = &row.DeletedAt.Time
	}
	return c, nil
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

func optionalInt64(value *int64) sql.NullInt64 {
	if value == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: *value, Valid: true}
}

func optionalInt32(value *int32) sql.NullInt32 {
	if value == nil {
		return sql.NullInt32{}
	}
	return sql.NullInt32{Int32: *value, Valid: true}
}
