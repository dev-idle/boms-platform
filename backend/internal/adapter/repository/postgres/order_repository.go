package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainpolicy "github.com/boms/backend/internal/domain/policy"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
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
	orderType, err := mapOrderTypeToSQL(params.Type)
	if err != nil {
		return nil, err
	}
	var guestName, guestPhone *string
	if params.Guest != nil {
		guestName, guestPhone = &params.Guest.Name, &params.Guest.Phone
	}
	row, err := r.q(ctx).CreateOrder(ctx, sqlcgen.CreateOrderParams{
		UserID:               params.UserID,
		Channel:              sqlcgen.OrderChannel(params.Channel),
		GuestName:            guestName,
		GuestPhone:           guestPhone,
		Code:                 params.Code,
		OrderType:            orderType,
		Status:               status,
		SubtotalCents:        params.SubtotalCents,
		DiscountCents:        params.DiscountCents,
		TotalCents:           params.TotalCents,
		DiscountCodeID:       params.DiscountCodeID,
		DiscountCodeSnapshot: params.DiscountCodeSnapshot,
		PickupAt:             params.PickupAt,
		TermsVersion:         params.TermsVersion,
		CheckoutKey:          params.CheckoutKey,
		PaymentHoldMinutes:   utils.Int32FromInt64(int64(params.PaymentHold / time.Minute)),
	})
	if err != nil {
		return nil, mapRepoError(err, "create order")
	}
	return mapOrder(row), nil
}

// GetByCheckoutKey implements port.OrderRepository.
func (r *OrderRepository) GetByCheckoutKey(ctx context.Context, userID, checkoutKey uuid.UUID) (*domainorder.Order, error) {
	row, err := r.q(ctx).GetOrderByCheckoutKey(ctx, sqlcgen.GetOrderByCheckoutKeyParams{
		UserID:      userID,
		CheckoutKey: &checkoutKey,
	})
	if err != nil {
		return nil, mapRepoError(err, "get order by checkout key")
	}
	return mapOrder(row), nil
}

// GetByStaffCheckoutKey implements port.OrderRepository.
func (r *OrderRepository) GetByStaffCheckoutKey(ctx context.Context, checkoutKey uuid.UUID) (*domainorder.Order, error) {
	row, err := r.q(ctx).GetStaffOrderByCheckoutKey(ctx, &checkoutKey)
	if err != nil {
		return nil, mapRepoError(err, "get staff order by checkout key")
	}
	return mapOrder(row), nil
}

// ListDueUnpaid implements port.OrderRepository.
func (r *OrderRepository) ListDueUnpaid(ctx context.Context, grace time.Duration, limit int32) ([]uuid.UUID, error) {
	ids, err := r.q(ctx).ListDueUnpaidOrders(ctx, sqlcgen.ListDueUnpaidOrdersParams{GraceSeconds: grace.Seconds(), MaxRows: limit})
	if err != nil {
		return nil, mapRepoError(err, "list due unpaid orders")
	}
	return ids, nil
}

// Expire implements port.OrderRepository.
func (r *OrderRepository) Expire(ctx context.Context, orderID uuid.UUID, grace time.Duration) (*domainorder.Order, error) {
	row, err := r.q(ctx).ExpireOrder(ctx, sqlcgen.ExpireOrderParams{ID: orderID, GraceSeconds: grace.Seconds()})
	if err != nil {
		return nil, mapRepoError(err, "expire order")
	}
	return mapOrder(row), nil
}

// Reschedule implements port.OrderRepository.
func (r *OrderRepository) Reschedule(ctx context.Context, params port.RescheduleOrderParams) (*domainorder.Order, error) {
	row, err := r.q(ctx).RescheduleOrder(ctx, sqlcgen.RescheduleOrderParams{
		ID:        params.OrderID,
		PickupAt:  &params.PickupAt,
		OrderType: sqlcgen.OrderType(params.Type),
	})
	if err != nil {
		return nil, mapRepoError(err, "reschedule order")
	}
	return mapOrder(row), nil
}

// ListMissedPickups implements port.OrderRepository.
func (r *OrderRepository) ListMissedPickups(ctx context.Context, missedBefore time.Time, limit int32) ([]uuid.UUID, error) {
	ids, err := r.q(ctx).ListMissedPickups(ctx, sqlcgen.ListMissedPickupsParams{MissedBefore: &missedBefore, MaxRows: limit})
	if err != nil {
		return nil, mapRepoError(err, "list missed pickups")
	}
	return ids, nil
}

// CountCustomerDiscountUses implements port.OrderRepository.
func (r *OrderRepository) CountCustomerDiscountUses(ctx context.Context, userID, discountCodeID uuid.UUID) (int, error) {
	count, err := r.q(ctx).CountCustomerDiscountUses(ctx, sqlcgen.CountCustomerDiscountUsesParams{
		UserID:         userID,
		DiscountCodeID: &discountCodeID,
	})
	if err != nil {
		return 0, mapRepoError(err, "count customer discount uses")
	}
	return int(count), nil
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

func (r *OrderRepository) LockForUpdate(ctx context.Context, orderID uuid.UUID) (*domainorder.Order, error) {
	if txFromContext(ctx) == nil {
		return nil, apperrors.Errorf("lock order: requires a transaction")
	}
	row, err := r.q(ctx).LockOrder(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "lock order")
	}
	return mapOrder(row), nil
}

func (r *OrderRepository) ListByUser(ctx context.Context, params port.ListOrdersParams) ([]domainorder.Order, error) {
	status, err := optionalOrderStatus(params.Filter.Status)
	if err != nil {
		return nil, err
	}
	rows, err := r.q(ctx).ListOrdersByUser(ctx, sqlcgen.ListOrdersByUserParams{
		UserID:       params.UserID,
		Status:       status,
		PlacedFrom:   params.Filter.PlacedFrom,
		PlacedBefore: params.Filter.PlacedBefore,
		Limit:        params.Limit,
		Offset:       params.Offset,
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

func (r *OrderRepository) ListCountByUser(
	ctx context.Context,
	userID uuid.UUID,
	filter port.OrderHistoryFilter,
) (int64, error) {
	status, err := optionalOrderStatus(filter.Status)
	if err != nil {
		return 0, err
	}
	count, err := r.q(ctx).ListOrdersByUserCount(ctx, sqlcgen.ListOrdersByUserCountParams{
		UserID:       userID,
		Status:       status,
		PlacedFrom:   filter.PlacedFrom,
		PlacedBefore: filter.PlacedBefore,
	})
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

func (r *OrderRepository) StaffListPickups(ctx context.Context, params port.StaffListPickupsParams) ([]port.StaffOrderListRow, error) {
	rows, err := r.q(ctx).StaffListPickups(ctx, sqlcgen.StaffListPickupsParams{
		PickupFrom: params.From,
		PickupTo:   params.To,
		Limit:      params.Limit,
		Offset:     params.Offset,
	})
	if err != nil {
		return nil, mapRepoError(err, "staff list pickups")
	}
	out := make([]port.StaffOrderListRow, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapStaffListOrdersRow(sqlcgen.StaffListOrdersRow(row)))
	}
	return out, nil
}

func (r *OrderRepository) StaffListPickupsCount(ctx context.Context, from, to time.Time) (int64, error) {
	count, err := r.q(ctx).StaffListPickupsCount(ctx, sqlcgen.StaffListPickupsCountParams{PickupFrom: from, PickupTo: to})
	if err != nil {
		return 0, mapRepoError(err, "staff list pickups count")
	}
	return count, nil
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
		configuration := json.RawMessage(`{}`)
		if item.Customization != nil {
			if configuration, err = json.Marshal(item.Customization); err != nil {
				return apperrors.Errorf("encode order item customization: %w", err)
			}
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
		item, err := mapOrderItem(row)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}

func (r *OrderRepository) HasCustomItems(ctx context.Context, orderID uuid.UUID) (bool, error) {
	has, err := r.q(ctx).OrderHasCustomItems(ctx, orderID)
	if err != nil {
		return false, mapRepoError(err, "order has custom items")
	}
	return has, nil
}

func (r *OrderRepository) HasOpen(ctx context.Context, userID uuid.UUID) (bool, error) {
	open, err := r.q(ctx).HasOpenOrdersForUser(ctx, userID)
	if err != nil {
		return false, mapRepoError(err, "has open orders for user")
	}
	return open, nil
}

func (r *OrderRepository) ListByUserBefore(
	ctx context.Context,
	userID uuid.UUID,
	before *port.PageCursor,
	limit int32,
) ([]domainorder.Order, error) {
	params := sqlcgen.ListOrdersByUserBeforeParams{UserID: userID, Limit: limit}
	if before != nil {
		params.BeforeAt, params.BeforeID = &before.At, &before.ID
	}
	rows, err := r.q(ctx).ListOrdersByUserBefore(ctx, params)
	if err != nil {
		return nil, mapRepoError(err, "list orders by user before")
	}
	out := make([]domainorder.Order, 0, len(rows))
	for _, row := range rows {
		out = append(out, *mapOrder(row))
	}
	return out, nil
}

func (r *OrderRepository) ListItemsByOrderIDs(
	ctx context.Context,
	orderIDs []uuid.UUID,
) (map[uuid.UUID][]domainorder.Item, error) {
	rows, err := r.q(ctx).ListOrderItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, mapRepoError(err, "list order items by order ids")
	}
	out := make(map[uuid.UUID][]domainorder.Item, len(orderIDs))
	for _, row := range rows {
		item, err := mapOrderItem(row)
		if err != nil {
			return nil, err
		}
		out[row.OrderID] = append(out[row.OrderID], item)
	}
	return out, nil
}

func (r *OrderRepository) ListStatusEventsByOrderIDs(
	ctx context.Context,
	orderIDs []uuid.UUID,
) (map[uuid.UUID][]domainorder.StatusEvent, error) {
	rows, err := r.q(ctx).ListOrderStatusEventsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, mapRepoError(err, "list order status events by order ids")
	}
	out := make(map[uuid.UUID][]domainorder.StatusEvent, len(orderIDs))
	for _, row := range rows {
		out[row.OrderID] = append(out[row.OrderID], domainorder.StatusEvent{
			To:        mapOrderStatusFromSQL(row.ToStatus),
			ActorRole: actorRoleFromSQL(row.ActorRole),
			Reason:    reasonFromSQL(row.Reason),
			At:        row.CreatedAt,
		})
	}
	return out, nil
}

func (r *OrderRepository) NextDayNumber(ctx context.Context) (time.Time, int, error) {
	if txFromContext(ctx) == nil {
		return time.Time{}, 0, apperrors.Errorf("next order day number: requires a transaction")
	}
	row, err := r.q(ctx).NextOrderDayNumber(ctx, domainstore.Location.String())
	if err != nil {
		return time.Time{}, 0, mapRepoError(err, "next order day number")
	}
	return row.Day, int(row.LastNumber), nil
}

// pickupSlotLockNamespace keeps slot locks apart from any other advisory lock.
const pickupSlotLockNamespace = 1101

func (r *OrderRepository) HoldPickupSlot(ctx context.Context, startsAt time.Time, length time.Duration) (int, error) {
	if txFromContext(ctx) == nil {
		return 0, apperrors.Errorf("hold pickup slot: requires a transaction")
	}
	if err := r.q(ctx).LockPickupSlot(ctx, sqlcgen.LockPickupSlotParams{
		Namespace: pickupSlotLockNamespace,
		StartsAt:  startsAt,
	}); err != nil {
		return 0, mapRepoError(err, "lock pickup slot")
	}
	count, err := r.q(ctx).CountOrdersInSlot(ctx, sqlcgen.CountOrdersInSlotParams{
		FromAt: startsAt,
		ToAt:   startsAt.Add(length),
	})
	if err != nil {
		return 0, mapRepoError(err, "count orders in slot")
	}
	return int(count), nil
}

// customerBookingLockNamespace keeps customer locks apart from any other advisory lock.
const customerBookingLockNamespace = 1102

func (r *OrderRepository) HoldCustomerDay(ctx context.Context, userID uuid.UUID, from, to time.Time) (int, error) {
	if txFromContext(ctx) == nil {
		return 0, apperrors.Errorf("hold customer day: requires a transaction")
	}
	if err := r.q(ctx).LockCustomerBookings(ctx, sqlcgen.LockCustomerBookingsParams{
		Namespace: customerBookingLockNamespace,
		UserID:    userID,
	}); err != nil {
		return 0, mapRepoError(err, "lock customer bookings")
	}
	count, err := r.q(ctx).CountCustomerOrdersBetween(ctx, sqlcgen.CountCustomerOrdersBetweenParams{
		UserID: userID,
		FromAt: from,
		ToAt:   to,
	})
	if err != nil {
		return 0, mapRepoError(err, "count customer orders between")
	}
	return int(count), nil
}

func (r *OrderRepository) CountByPickupTime(ctx context.Context, from, to time.Time) ([]port.PickupCount, error) {
	rows, err := r.q(ctx).CountOrdersByPickupTime(ctx, sqlcgen.CountOrdersByPickupTimeParams{FromAt: from, ToAt: to})
	if err != nil {
		return nil, mapRepoError(err, "count orders by pickup time")
	}
	out := make([]port.PickupCount, 0, len(rows))
	for _, row := range rows {
		out = append(out, port.PickupCount{At: row.PickupAt, Orders: int(row.Count)})
	}
	return out, nil
}

func (r *OrderRepository) AddStatusEvent(ctx context.Context, params port.AddOrderStatusEventParams) error {
	var from *sqlcgen.OrderStatus
	if params.From != nil {
		mapped, err := mapOrderStatusToSQL(*params.From)
		if err != nil {
			return err
		}
		from = &mapped
	}
	to, err := mapOrderStatusToSQL(params.To)
	if err != nil {
		return err
	}
	var actorID *uuid.UUID
	var actorRole *sqlcgen.UserRole
	if params.Actor != nil {
		role, err := toSQLRole(params.Actor.Role)
		if err != nil {
			return err
		}
		actorID, actorRole = &params.Actor.ID, &role
	}
	var reason *string
	if params.Reason != "" {
		reason = &params.Reason
	}
	err = r.q(ctx).CreateOrderStatusEvent(ctx, sqlcgen.CreateOrderStatusEventParams{
		OrderID:    params.OrderID,
		FromStatus: from,
		ToStatus:   to,
		ActorID:    actorID,
		ActorRole:  actorRole,
		Reason:     reason,
	})
	if err != nil {
		return mapRepoError(err, "create order status event")
	}
	return nil
}

func (r *OrderRepository) ListStatusEvents(ctx context.Context, orderID uuid.UUID) ([]domainorder.StatusEvent, error) {
	rows, err := r.q(ctx).ListOrderStatusEvents(ctx, orderID)
	if err != nil {
		return nil, mapRepoError(err, "list order status events")
	}
	out := make([]domainorder.StatusEvent, 0, len(rows))
	for _, row := range rows {
		out = append(out, domainorder.StatusEvent{
			To:        mapOrderStatusFromSQL(row.ToStatus),
			ActorRole: actorRoleFromSQL(row.ActorRole),
			Reason:    reasonFromSQL(row.Reason),
			At:        row.CreatedAt,
		})
	}
	return out, nil
}

func reasonFromSQL(reason *string) string {
	if reason == nil {
		return ""
	}
	return *reason
}

// actorRoleFromSQL is the role of whoever moved an order; empty for the system.
func actorRoleFromSQL(role *sqlcgen.UserRole) domainuser.Role {
	if role == nil {
		return ""
	}
	return fromSQLRole(*role)
}

func mapOrder(row sqlcgen.Order) *domainorder.Order {
	return &domainorder.Order{
		ID:                   row.ID,
		Code:                 row.Code,
		Type:                 domainorder.Type(row.OrderType),
		UserID:               row.UserID,
		Guest:                guestFromSQL(row.GuestName, row.GuestPhone),
		Channel:              domainorder.Channel(row.Channel),
		Status:               mapOrderStatusFromSQL(row.Status),
		SubtotalCents:        row.SubtotalCents,
		DiscountCents:        row.DiscountCents,
		TotalCents:           row.TotalCents,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		DiscountCodeID:       row.DiscountCodeID,
		DiscountCodeSnapshot: row.DiscountCodeSnapshot,
		PickupAt:             row.PickupAt,
		PaymentDueAt:         row.PaymentDueAt,
		Terms:                mapTermsAcceptance(row.TermsVersion, row.TermsAcceptedAt),
	}
}

// guestFromSQL reads the guest a row names; the table's check keeps the name
// and the phone both set, for a guest's order only, or both empty.
func guestFromSQL(name, phone *string) *domainorder.Guest {
	if name == nil || phone == nil {
		return nil
	}
	return &domainorder.Guest{Name: *name, Phone: *phone}
}

// mapTermsAcceptance reads the version and instant a row recorded together;
// the table's check keeps them both set or both empty.
func mapTermsAcceptance(version *string, acceptedAt *time.Time) *domainpolicy.Acceptance {
	if version == nil || acceptedAt == nil {
		return nil
	}
	return &domainpolicy.Acceptance{Version: *version, AcceptedAt: *acceptedAt}
}

func mapOrderItem(row sqlcgen.OrderItem) (domainorder.Item, error) {
	customization, err := domainorder.ParseCustomization(row.Configuration)
	if err != nil {
		return domainorder.Item{}, apperrors.Errorf("order item %s: %w", row.ID, err)
	}
	return domainorder.Item{
		ID:             row.ID,
		OrderID:        row.OrderID,
		LineType:       mapLineTypeFromSQL(row.LineType),
		Customization:  customization,
		Name:           row.Name,
		Slug:           row.Slug,
		Quantity:       row.Quantity,
		UnitPriceCents: row.UnitPriceCents,
		LineTotalCents: row.LineTotalCents,
		CreatedAt:      row.CreatedAt,
		ProductID:      row.ProductID,
		ComboID:        row.ComboID,
	}, nil
}

func mapOrderTypeToSQL(t domainorder.Type) (sqlcgen.OrderType, error) {
	switch t {
	case domainorder.TypeInstant:
		return sqlcgen.OrderTypeInstant, nil
	case domainorder.TypePreOrder:
		return sqlcgen.OrderTypePreOrder, nil
	default:
		return "", apperrors.Errorf("unsupported order type: %s", t)
	}
}

func mapOrderStatusToSQL(s domainorder.Status) (sqlcgen.OrderStatus, error) {
	switch s {
	case domainorder.StatusAwaitingPayment:
		return sqlcgen.OrderStatusAwaitingPayment, nil
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
	case domainorder.StatusExpired:
		return sqlcgen.OrderStatusExpired, nil
	case domainorder.StatusNoShow:
		return sqlcgen.OrderStatusNoShow, nil
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
		row.Code,
		row.OrderType,
		row.Channel,
		row.CustomerEmail,
		row.CustomerEmailVerified,
		row.CustomerDisplayName,
		// The list never carries phones: staff open an order to call its customer.
		nil,
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
		row.Code,
		row.OrderType,
		row.Channel,
		row.CustomerEmail,
		row.CustomerEmailVerified,
		row.CustomerDisplayName,
		row.CustomerPhone,
	)
}

func mapStaffOrderJoined(
	id uuid.UUID,
	userID *uuid.UUID,
	status sqlcgen.OrderStatus,
	subtotalCents, discountCents, totalCents int64,
	discountCodeID *uuid.UUID,
	discountCodeSnapshot *string,
	pickupAt *time.Time,
	createdAt, updatedAt time.Time,
	code string,
	orderType sqlcgen.OrderType,
	channel sqlcgen.OrderChannel,
	customerEmail *string,
	customerEmailVerified bool,
	customerDisplayName *string,
	customerPhone *string,
) *port.StaffOrderListRow {
	return &port.StaffOrderListRow{
		Order: domainorder.Order{
			ID:                   id,
			UserID:               userID,
			Channel:              domainorder.Channel(channel),
			Status:               mapOrderStatusFromSQL(status),
			SubtotalCents:        subtotalCents,
			DiscountCents:        discountCents,
			TotalCents:           totalCents,
			DiscountCodeID:       discountCodeID,
			DiscountCodeSnapshot: discountCodeSnapshot,
			PickupAt:             pickupAt,
			CreatedAt:            createdAt,
			UpdatedAt:            updatedAt,
			Code:                 code,
			Type:                 domainorder.Type(orderType),
		},
		CustomerEmail:         customerEmail,
		CustomerEmailVerified: customerEmailVerified,
		CustomerDisplayName:   customerDisplayName,
		CustomerPhone:         customerPhone,
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
	case sqlcgen.OrderStatusAwaitingPayment:
		return domainorder.StatusAwaitingPayment
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
	case sqlcgen.OrderStatusExpired:
		return domainorder.StatusExpired
	case sqlcgen.OrderStatusNoShow:
		return domainorder.StatusNoShow
	case sqlcgen.OrderStatusFulfilled:
		return domainorder.StatusFulfilled
	default:
		return domainorder.Status(s)
	}
}
