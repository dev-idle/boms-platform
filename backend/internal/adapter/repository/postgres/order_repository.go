package postgres

import (
	"context"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
)

type OrderRepository struct {
	queries *sqlcgen.Queries
}

func NewOrderRepository(pool *Pool) *OrderRepository {
	return &OrderRepository{queries: pool.Queries()}
}

func (r *OrderRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *OrderRepository) Create(ctx context.Context, params port.CreateOrderParams) (*domainorder.Order, error) {
	status, err := mapOrderStatusToSQL(params.Status)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).CreateOrder(ctx, sqlcgen.CreateOrderParams{
		UserID:               params.UserID,
		Status:               status,
		SubtotalCents:        params.SubtotalCents,
		DiscountCents:        params.DiscountCents,
		TotalCents:           params.TotalCents,
		DiscountCodeID:       optionalUUID(params.DiscountCodeID),
		DiscountCodeSnapshot: optionalString(params.DiscountCodeSnapshot),
	})
	if err != nil {
		return nil, mapRepoError(err, "create order")
	}
	return mapOrder(row), nil
}

func (r *OrderRepository) GetByIDForUser(ctx context.Context, userID, orderID uuid.UUID) (*domainorder.Order, error) {
	row, err := r.q(ctx).GetOrderByIDForUser(ctx, sqlcgen.GetOrderByIDForUserParams{
		ID:     orderID,
		UserID: userID,
	})
	if err != nil {
		return nil, mapRepoError(err, "get order")
	}
	return mapOrder(row), nil
}

func (r *OrderRepository) ListByUser(ctx context.Context, params port.ListOrdersParams) ([]domainorder.Order, error) {
	rows, err := r.q(ctx).ListOrdersByUser(ctx, sqlcgen.ListOrdersByUserParams{
		UserID: params.UserID,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "list orders")
	}
	out := make([]domainorder.Order, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapOrder(row))
	}
	return out, nil
}

func (r *OrderRepository) ListCountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	count, err := r.q(ctx).ListOrdersByUserCount(ctx, userID)
	if err != nil {
		return 0, mapRepoError(err, "list orders count")
	}
	return count, nil
}

func (r *OrderRepository) CreateItems(ctx context.Context, items []port.CreateOrderItemParams) error {
	for _, item := range items {
		lineType, err := mapCartLineTypeToSQL(domaincart.LineType(item.LineType))
		if err != nil {
			return err
		}
		if _, err := r.q(ctx).CreateOrderItem(ctx, sqlcgen.CreateOrderItemParams{
			OrderID:        item.OrderID,
			LineType:       lineType,
			ProductID:      optionalUUID(item.ProductID),
			ComboID:        optionalUUID(item.ComboID),
			Name:           item.Name,
			Slug:           item.Slug,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: item.LineTotalCents,
		}); err != nil {
			return mapRepoError(err, "create order item")
		}
	}
	return nil
}

func (r *OrderRepository) SumItemQuantitiesByOrderIDs(
	ctx context.Context,
	orderIDs []uuid.UUID,
) (map[uuid.UUID]int32, error) {
	if len(orderIDs) == 0 {
		return map[uuid.UUID]int32{}, nil
	}
	rows, err := r.q(ctx).SumOrderItemQuantitiesByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, mapRepoError(err, "sum order item quantities")
	}
	out := make(map[uuid.UUID]int32, len(rows))
	for _, row := range rows {
		out[row.OrderID] = utils.Int32FromInt64(row.ItemCount)
	}
	return out, nil
}

func (r *OrderRepository) ListItemsByOrderID(ctx context.Context, orderID uuid.UUID) ([]domainorder.Item, error) {
	rows, err := r.q(ctx).ListOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "list order items")
	}
	out := make([]domainorder.Item, 0, len(rows))
	for _, row := range rows {
		out = append(out, mapOrderItem(row))
	}
	return out, nil
}

func mapOrder(row sqlcgen.Order) *domainorder.Order {
	o := &domainorder.Order{
		ID:            row.ID,
		UserID:        row.UserID,
		Status:        mapOrderStatusFromSQL(row.Status),
		SubtotalCents: row.SubtotalCents,
		DiscountCents: row.DiscountCents,
		TotalCents:    row.TotalCents,
		CreatedAt:     row.CreatedAt,
		UpdatedAt:     row.UpdatedAt,
	}
	if row.DiscountCodeID.Valid {
		id := row.DiscountCodeID.UUID
		o.DiscountCodeID = &id
	}
	if row.DiscountCodeSnapshot.Valid {
		s := row.DiscountCodeSnapshot.String
		o.DiscountCodeSnapshot = &s
	}
	return o
}

func mapOrderItem(row sqlcgen.OrderItem) domainorder.Item {
	item := domainorder.Item{
		ID:             row.ID,
		OrderID:        row.OrderID,
		LineType:       mapCartLineTypeFromSQL(row.LineType),
		Name:           row.Name,
		Slug:           row.Slug,
		Quantity:       row.Quantity,
		UnitPriceCents: row.UnitPriceCents,
		LineTotalCents: row.LineTotalCents,
		CreatedAt:      row.CreatedAt,
	}
	if row.ProductID.Valid {
		id := row.ProductID.UUID
		item.ProductID = &id
	}
	if row.ComboID.Valid {
		id := row.ComboID.UUID
		item.ComboID = &id
	}
	return item
}

func mapOrderStatusToSQL(s domainorder.Status) (sqlcgen.OrderStatus, error) {
	switch s {
	case domainorder.StatusPending:
		return sqlcgen.OrderStatusPending, nil
	case domainorder.StatusConfirmed:
		return sqlcgen.OrderStatusConfirmed, nil
	case domainorder.StatusCancelled:
		return sqlcgen.OrderStatusCancelled, nil
	case domainorder.StatusFulfilled:
		return sqlcgen.OrderStatusFulfilled, nil
	default:
		return "", apperrors.Errorf("unsupported order status: %s", s)
	}
}

func mapOrderStatusFromSQL(s sqlcgen.OrderStatus) domainorder.Status {
	switch s {
	case sqlcgen.OrderStatusPending:
		return domainorder.StatusPending
	case sqlcgen.OrderStatusConfirmed:
		return domainorder.StatusConfirmed
	case sqlcgen.OrderStatusCancelled:
		return domainorder.StatusCancelled
	case sqlcgen.OrderStatusFulfilled:
		return domainorder.StatusFulfilled
	default:
		return domainorder.Status(s)
	}
}
