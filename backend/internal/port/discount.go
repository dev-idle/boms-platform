package port

import (
	"context"
	"time"

	domaindiscount "github.com/boms/backend/internal/domain/discount"
	"github.com/google/uuid"
)

type CreateDiscountCodeParams struct {
	Code          string
	DiscountType  domaindiscount.Type
	Value         int64
	MinOrderCents *int64
	MaxUses       *int32
	StartsAt      time.Time
	EndsAt        time.Time
	IsActive      bool
}

type UpdateDiscountCodeParams struct {
	ID            uuid.UUID
	Code          string
	DiscountType  domaindiscount.Type
	Value         int64
	MinOrderCents *int64
	MaxUses       *int32
	StartsAt      time.Time
	EndsAt        time.Time
	IsActive      bool
}

type ManagerListDiscountCodesParams struct {
	Search *string
	Limit  int32
	Offset int32
}

type DiscountCodeRepository interface {
	Create(ctx context.Context, params CreateDiscountCodeParams) (*domaindiscount.Code, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domaindiscount.Code, error)
	Update(ctx context.Context, params UpdateDiscountCodeParams) (*domaindiscount.Code, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	ManagerList(ctx context.Context, params ManagerListDiscountCodesParams) ([]domaindiscount.Code, error)
	ManagerListCount(ctx context.Context, search *string) (int64, error)
}
