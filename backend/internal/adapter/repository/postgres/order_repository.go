package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domaincart "github.com/boms/backend/internal/domain/cart"
	domainorder "github.com/boms/backend/internal/domain/order"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
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
		DiscountCodeID:       params.DiscountCodeID,
		DiscountCodeSnapshot: params.DiscountCodeSnapshot,
		PickupAt:             params.PickupAt,
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

func (r *OrderRepository) StaffGetByID(ctx context.Context, orderID uuid.UUID) (*port.StaffOrderListRow, error) {
	row, err := r.q(ctx).StaffGetOrderByID(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "staff get order")
	}
	return mapStaffGetOrderByIDRow(row), nil
}

func (r *OrderRepository) StaffList(ctx context.Context, params port.StaffListOrdersParams) ([]port.StaffOrderListRow, error) {
	status, err := optionalOrderStatus(params.Status)
	if err != nil {
		return nil, err
	}
	rows, err := r.q(ctx).StaffListOrders(ctx, sqlcgen.StaffListOrdersParams{
		Status: status,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "staff list orders")
	}
	out := make([]port.StaffOrderListRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapStaffListOrdersRow(row))
	}
	return out, nil
}

func (r *OrderRepository) StaffListCount(ctx context.Context, status *domainorder.Status) (int64, error) {
	statusSQL, err := optionalOrderStatus(status)
	if err != nil {
		return 0, err
	}
	count, err := r.q(ctx).StaffListOrdersCount(ctx, statusSQL)
	if err != nil {
		return 0, mapRepoError(err, "staff list orders count")
	}
	return count, nil
}

func (r *OrderRepository) BakerListProduction(
	ctx context.Context,
	params port.BakerListOrdersParams,
) ([]port.StaffOrderListRow, error) {
	status, err := optionalOrderStatus(params.Status)
	if err != nil {
		return nil, err
	}
	rows, err := r.q(ctx).BakerListProductionOrders(ctx, sqlcgen.BakerListProductionOrdersParams{
		Status: status,
		Limit:  params.Limit,
		Offset: params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "baker list production orders")
	}
	out := make([]port.StaffOrderListRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapBakerListProductionOrdersRow(row))
	}
	return out, nil
}

func (r *OrderRepository) BakerListProductionCount(
	ctx context.Context,
	status *domainorder.Status,
) (int64, error) {
	statusSQL, err := optionalOrderStatus(status)
	if err != nil {
		return 0, err
	}
	count, err := r.q(ctx).BakerListProductionOrdersCount(ctx, statusSQL)
	if err != nil {
		return 0, mapRepoError(err, "baker list production orders count")
	}
	return count, nil
}

func (r *OrderRepository) UpdateStatus(
	ctx context.Context,
	params port.UpdateOrderStatusParams,
) (*domainorder.Order, error) {
	fromStatus, err := mapOrderStatusToSQL(params.FromStatus)
	if err != nil {
		return nil, err
	}
	toStatus, err := mapOrderStatusToSQL(params.ToStatus)
	if err != nil {
		return nil, err
	}
	row, err := r.q(ctx).UpdateOrderStatus(ctx, sqlcgen.UpdateOrderStatusParams{
		ID:         params.OrderID,
		FromStatus: fromStatus,
		ToStatus:   toStatus,
	})
	if err != nil {
		return nil, mapRepoError(err, "update order status")
	}
	return mapOrder(row), nil
}

// orderItemRecord is one element of the JSON array CreateOrderItems expands with
// jsonb_to_recordset; field names must match that query's column definition list.
type orderItemRecord struct {
	OrderID        uuid.UUID        `json:"order_id"`
	LineType       sqlcgen.LineType `json:"line_type"`
	ProductID      *uuid.UUID       `json:"product_id"`
	ComboID        *uuid.UUID       `json:"combo_id"`
	Configuration  json.RawMessage  `json:"configuration"`
	Name           string           `json:"name"`
	Slug           string           `json:"slug"`
	Quantity       int32            `json:"quantity"`
	UnitPriceCents int64            `json:"unit_price_cents"`
	LineTotalCents int64            `json:"line_total_cents"`
}

// CreateItems inserts all order lines in a single statement.
func (r *OrderRepository) CreateItems(ctx context.Context, items []port.CreateOrderItemParams) error {
	if len(items) == 0 {
		return nil
	}
	records := make([]orderItemRecord, 0, len(items))
	for _, item := range items {
		lineType, err := mapLineTypeToSQL(item.LineType)
		if err != nil {
			return err
		}
		configuration := item.Configuration
		if len(configuration) == 0 {
			configuration = json.RawMessage(`{}`)
		}
		records = append(records, orderItemRecord{
			OrderID:        item.OrderID,
			LineType:       lineType,
			ProductID:      item.ProductID,
			ComboID:        item.ComboID,
			Configuration:  configuration,
			Name:           item.Name,
			Slug:           item.Slug,
			Quantity:       item.Quantity,
			UnitPriceCents: item.UnitPriceCents,
			LineTotalCents: item.LineTotalCents,
		})
	}
	payload, err := json.Marshal(records)
	if err != nil {
		return apperrors.Errorf("encode order items: %w", err)
	}
	inserted, err := r.q(ctx).CreateOrderItems(ctx, payload)
	if err != nil {
		return mapRepoError(err, "create order items")
	}
	if inserted != int64(len(records)) {
		return apperrors.Errorf("create order items: inserted %d of %d", inserted, len(records))
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
	return &domainorder.Order{
		ID:                   row.ID,
		UserID:               row.UserID,
		Status:               mapOrderStatusFromSQL(row.Status),
		SubtotalCents:        row.SubtotalCents,
		DiscountCents:        row.DiscountCents,
		TotalCents:           row.TotalCents,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		DiscountCodeID:       row.DiscountCodeID,
		DiscountCodeSnapshot: row.DiscountCodeSnapshot,
		PickupAt:             row.PickupAt,
	}
}

func mapOrderItem(row sqlcgen.OrderItem) domainorder.Item {
	item := domainorder.Item{
		ID:             row.ID,
		OrderID:        row.OrderID,
		LineType:       mapLineTypeFromSQL(row.LineType),
		Configuration:  row.Configuration,
		Name:           row.Name,
		Slug:           row.Slug,
		Quantity:       row.Quantity,
		UnitPriceCents: row.UnitPriceCents,
		LineTotalCents: row.LineTotalCents,
		CreatedAt:      row.CreatedAt,
		ProductID:      row.ProductID,
		ComboID:        row.ComboID,
	}
	if len(item.Configuration) == 0 {
		item.Configuration = domaincart.EmptyConfiguration
	}
	return item
}

func mapOrderStatusToSQL(s domainorder.Status) (sqlcgen.OrderStatus, error) {
	switch s {
	case domainorder.StatusPending:
		return sqlcgen.OrderStatusPending, nil
	case domainorder.StatusConfirmed:
		return sqlcgen.OrderStatusConfirmed, nil
	case domainorder.StatusInProduction:
		return sqlcgen.OrderStatusInProduction, nil
	case domainorder.StatusReady:
		return sqlcgen.OrderStatusReady, nil
	case domainorder.StatusCancelled:
		return sqlcgen.OrderStatusCancelled, nil
	case domainorder.StatusFulfilled:
		return sqlcgen.OrderStatusFulfilled, nil
	default:
		return "", apperrors.Errorf("unsupported order status: %s", s)
	}
}

func mapStaffListOrdersRow(row sqlcgen.StaffListOrdersRow) *port.StaffOrderListRow {
	return mapStaffOrderJoined(
		row.ID,
		row.UserID,
		row.Status,
		row.SubtotalCents,
		row.DiscountCents,
		row.TotalCents,
		row.DiscountCodeID,
		row.DiscountCodeSnapshot,
		row.PickupAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.CustomerEmail,
		row.CustomerDisplayName,
		row.CustomerPhone,
	)
}

func mapStaffGetOrderByIDRow(row sqlcgen.StaffGetOrderByIDRow) *port.StaffOrderListRow {
	return mapStaffOrderJoined(
		row.ID,
		row.UserID,
		row.Status,
		row.SubtotalCents,
		row.DiscountCents,
		row.TotalCents,
		row.DiscountCodeID,
		row.DiscountCodeSnapshot,
		row.PickupAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.CustomerEmail,
		row.CustomerDisplayName,
		row.CustomerPhone,
	)
}

func mapBakerListProductionOrdersRow(row sqlcgen.BakerListProductionOrdersRow) *port.StaffOrderListRow {
	return mapStaffOrderJoined(
		row.ID,
		row.UserID,
		row.Status,
		row.SubtotalCents,
		row.DiscountCents,
		row.TotalCents,
		row.DiscountCodeID,
		row.DiscountCodeSnapshot,
		row.PickupAt,
		row.CreatedAt,
		row.UpdatedAt,
		row.CustomerEmail,
		row.CustomerDisplayName,
		nil,
	)
}

func mapStaffOrderJoined(
	id, userID uuid.UUID,
	status sqlcgen.OrderStatus,
	subtotalCents, discountCents, totalCents int64,
	discountCodeID *uuid.UUID,
	discountCodeSnapshot *string,
	pickupAt *time.Time,
	createdAt, updatedAt time.Time,
	customerEmail string,
	customerDisplayName *string,
	customerPhone *string,
) *port.StaffOrderListRow {
	return &port.StaffOrderListRow{
		Order: domainorder.Order{
			ID:                   id,
			UserID:               userID,
			Status:               mapOrderStatusFromSQL(status),
			SubtotalCents:        subtotalCents,
			DiscountCents:        discountCents,
			TotalCents:           totalCents,
			DiscountCodeID:       discountCodeID,
			DiscountCodeSnapshot: discountCodeSnapshot,
			PickupAt:             pickupAt,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
		},
		CustomerEmail:       customerEmail,
		CustomerDisplayName: customerDisplayName,
		CustomerPhone:       customerPhone,
	}
}

func optionalOrderStatus(status *domainorder.Status) (*sqlcgen.OrderStatus, error) {
	if status == nil {
		return nil, nil
	}
	if !status.Valid() {
		return nil, apperrors.ErrValidation.WithDetail("status", "invalid order status")
	}
	mapped, err := mapOrderStatusToSQL(*status)
	if err != nil {
		return nil, err
	}
	return &mapped, nil
}

func mapOrderStatusFromSQL(s sqlcgen.OrderStatus) domainorder.Status {
	switch s {
	case sqlcgen.OrderStatusPending:
		return domainorder.StatusPending
	case sqlcgen.OrderStatusConfirmed:
		return domainorder.StatusConfirmed
	case sqlcgen.OrderStatusInProduction:
		return domainorder.StatusInProduction
	case sqlcgen.OrderStatusReady:
		return domainorder.StatusReady
	case sqlcgen.OrderStatusCancelled:
		return domainorder.StatusCancelled
	case sqlcgen.OrderStatusFulfilled:
		return domainorder.StatusFulfilled
	default:
		return domainorder.Status(s)
	}
}
