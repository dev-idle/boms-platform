package port

import (
	"context"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/google/uuid"
)

type CreateOrderParams struct {
	UserID               uuid.UUID
	Status               domainorder.Status
	SubtotalCents        int64
	DiscountCents        int64
	TotalCents           int64
	DiscountCodeID       *uuid.UUID
	DiscountCodeSnapshot *string
}

type CreateOrderItemParams struct {
	OrderID        uuid.UUID
	LineType       domaincart.LineType
	ProductID      *uuid.UUID
	ComboID        *uuid.UUID
	Name           string
	Slug           string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
}

type ListOrdersParams struct {
	UserID uuid.UUID
	Limit  int32
	Offset int32
}

type OrderRepository interface {
	Create(ctx context.Context, params CreateOrderParams) (*domainorder.Order, error)
	GetByIDForUser(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error)
	ListByUser(ctx context.Context, params ListOrdersParams) ([]domainorder.Order, error)
	ListCountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CreateItems(ctx context.Context, items []CreateOrderItemParams) error
	ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domainorder.Item, error)
	SumItemQuantitiesByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error)
}
