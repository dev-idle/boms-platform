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

type StaffListOrdersParams struct {
	Status *domainorder.Status
	Limit  int32
	Offset int32
}

type StaffOrderListRow struct {
	Order              domainorder.Order
	CustomerEmail      string
	CustomerDisplayName *string
}

type UpdateOrderStatusParams struct {
	OrderID    uuid.UUID
	FromStatus domainorder.Status
	ToStatus   domainorder.Status
}

type OrderRepository interface {
	Create(ctx context.Context, params CreateOrderParams) (*domainorder.Order, error)
	GetByIDForUser(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error)
	StaffGetByID(ctx context.Context, orderID uuid.UUID) (*StaffOrderListRow, error)
	ListByUser(ctx context.Context, params ListOrdersParams) ([]domainorder.Order, error)
	ListCountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	StaffList(ctx context.Context, params StaffListOrdersParams) ([]StaffOrderListRow, error)
	StaffListCount(ctx context.Context, status *domainorder.Status) (int64, error)
	UpdateStatus(ctx context.Context, params UpdateOrderStatusParams) (*domainorder.Order, error)
	CreateItems(ctx context.Context, items []CreateOrderItemParams) error
	ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domainorder.Item, error)
	SumItemQuantitiesByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error)
}
