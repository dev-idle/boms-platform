package port

import (
	"context"
	"encoding/json"
	"time"

	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/google/uuid"
)

type CreateOrderParams struct {
	UserID               uuid.UUID
	Code                 string
	Status               domainorder.Status
	SubtotalCents        int64
	DiscountCents        int64
	TotalCents           int64
	DiscountCodeID       *uuid.UUID
	DiscountCodeSnapshot *string
	PickupAt             *time.Time
}

type CreateOrderItemParams struct {
	OrderID        uuid.UUID
	LineType       domaincart.LineType
	ProductID      *uuid.UUID
	ComboID        *uuid.UUID
	Configuration  json.RawMessage
	Name           string
	Slug           string
	Quantity       int32
	UnitPriceCents int64
	LineTotalCents int64
}

// OrderHistoryFilter narrows a customer's order history. Nil fields do not
// filter; PlacedBefore is exclusive.
type OrderHistoryFilter struct {
	Status       *domainorder.Status
	PlacedFrom   *time.Time
	PlacedBefore *time.Time
}

type ListOrdersParams struct {
	UserID uuid.UUID
	Filter OrderHistoryFilter
	Limit  int32
	Offset int32
}

type StaffListOrdersParams struct {
	Status *domainorder.Status
	Limit  int32
	Offset int32
}

type StaffOrderListRow struct {
	Order               domainorder.Order
	CustomerEmail       string
	CustomerDisplayName *string
	CustomerPhone       *string
}

type BakerListOrdersParams struct {
	Status *domainorder.Status
	Limit  int32
	Offset int32
}

// AddOrderStatusEventParams records an order entering To. From is nil when the
// order is being placed.
type AddOrderStatusEventParams struct {
	OrderID   uuid.UUID
	From      *domainorder.Status
	To        domainorder.Status
	ActorID   uuid.UUID
	ActorRole domainuser.Role
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
	ListCountByUser(ctx context.Context, userID uuid.UUID, filter OrderHistoryFilter) (int64, error)
	StaffList(ctx context.Context, params StaffListOrdersParams) ([]StaffOrderListRow, error)
	StaffListCount(ctx context.Context, status *domainorder.Status) (int64, error)
	BakerListProduction(ctx context.Context, params BakerListOrdersParams) ([]StaffOrderListRow, error)
	BakerListProductionCount(ctx context.Context, status *domainorder.Status) (int64, error)
	UpdateStatus(ctx context.Context, params UpdateOrderStatusParams) (*domainorder.Order, error)
	CreateItems(ctx context.Context, items []CreateOrderItemParams) error
	ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domainorder.Item, error)
	SumItemQuantitiesByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID]int32, error)
	// NextDayNumber hands out the next order number of the bakery day the
	// transaction runs on, and that day. It must run in the transaction that
	// creates the order: the number is held until then.
	NextDayNumber(ctx context.Context) (day time.Time, number int, err error)
	AddStatusEvent(ctx context.Context, params AddOrderStatusEventParams) error
	ListStatusEvents(ctx context.Context, orderID uuid.UUID) ([]domainorder.StatusEvent, error)
}
