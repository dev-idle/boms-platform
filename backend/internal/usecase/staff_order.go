package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type StaffOrderUsecase struct {
	orders      port.OrderRepository
	tickets     port.TicketRepository
	payments    port.PaymentRepository
	store       port.StoreSettingsRepository
	cartUC      *CartUsecase
	transitions orderTransitions
	audit       *auditlogger.Service
	log         *zap.Logger
}

func NewStaffOrderUsecase(
	orders port.OrderRepository,
	tickets port.TicketRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
	payments port.PaymentRepository,
	discounts port.DiscountCodeRepository,
	store port.StoreSettingsRepository,
	cartUC *CartUsecase,
) *StaffOrderUsecase {
	return &StaffOrderUsecase{
		orders:   orders,
		tickets:  tickets,
		payments: payments,
		store:    store,
		cartUC:   cartUC,
		transitions: orderTransitions{
			tx: tx, orders: orders, tickets: tickets, discounts: discounts, payments: payments, events: events,
		},
		audit: audit,
		log:   log,
	}
}

func (u *StaffOrderUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	statusFilter string,
) ([]dto.StaffOrderSummaryResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)

	var status *domainorder.Status
	if trimmed := strings.TrimSpace(statusFilter); trimmed != "" {
		parsed := domainorder.Status(trimmed)
		if !parsed.Valid() {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("status", "invalid order status")
		}
		status = &parsed
	}

	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.StaffOrderListRow, error) {
			return u.orders.StaffList(ctx, port.StaffListOrdersParams{
				Status: status,
				Limit:  pageSize,
				Offset: utils.PageOffset(page, pageSize),
			})
		},
		func(ctx context.Context) (int64, error) {
			return u.orders.StaffListCount(ctx, status)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	orderIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		orderIDs = append(orderIDs, row.Order.ID)
	}
	itemCounts, err := u.orders.SumItemQuantitiesByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, 0, page, pageSize, err
	}

	out := make([]dto.StaffOrderSummaryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.StaffOrderSummaryResponse{
			ID:         row.Order.ID.String(),
			Code:       row.Order.Code,
			Status:     string(row.Order.Status),
			TotalCents: row.Order.TotalCents,
			ItemCount:  itemCounts[row.Order.ID],
			Customer:   toStaffOrderCustomer(&row),
			PickupAt:   row.Order.PickupAt,
			CreatedAt:  row.Order.CreatedAt,
		})
	}
	return out, total, page, pageSize, nil
}

func (u *StaffOrderUsecase) Get(ctx context.Context, orderID uuid.UUID) (*dto.StaffOrderResponse, error) {
	row, err := u.orders.StaffGetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrNotFound
		}
		return nil, err
	}
	if !row.Order.Status.VisibleToStaff() {
		return nil, domainorder.ErrNotFound
	}
	parts, err := readOrderDetail(ctx, u.orders, u.tickets, u.payments, row.Order.ID)
	if err != nil {
		return nil, err
	}
	if !domainorder.SeenByStaff(row.Order.Status, parts.timeline) {
		return nil, domainorder.ErrNotFound
	}
	return toStaffOrderResponse(row, parts), nil
}

// PatchStatus moves an order at the counter. A cancellation states the
// reason the customer is shown, and refunds a paid order in full.
func (u *StaffOrderUsecase) PatchStatus(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	orderID uuid.UUID,
	req dto.PatchStaffOrderStatusRequest,
) (*dto.StaffOrderResponse, error) {
	targetStatus := domainorder.Status(req.Status)
	if !targetStatus.Valid() {
		return nil, apperrors.ErrValidation.WithDetail("status", "invalid order status")
	}
	reason := ""
	if targetStatus == domainorder.StatusCancelled {
		var err error
		if reason, err = domainorder.NewCancelReason(req.Reason); err != nil {
			return nil, err
		}
	} else if req.Reason != "" {
		return nil, apperrors.ErrValidation.WithDetail("reason", "only a cancellation takes a reason")
	}

	beforeRow, err := u.orders.StaffGetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrNotFound
		}
		return nil, err
	}
	if !domainorder.CanStaffTransition(beforeRow.Order.Status, targetStatus) {
		return nil, domainorder.ErrInvalidStatusTransition
	}
	if beforeRow.Order.Status == domainorder.StatusPending && targetStatus == domainorder.StatusConfirmed {
		if err := u.stillInTime(ctx, beforeRow.Order); err != nil {
			return nil, err
		}
	}

	updated, err := u.transitions.apply(ctx, &port.OrderActor{ID: actorID, Role: actorRole}, port.UpdateOrderStatusParams{
		OrderID:    orderID,
		FromStatus: beforeRow.Order.Status,
		ToStatus:   targetStatus,
	}, reason)
	if err != nil {
		return nil, err
	}

	afterRow := *beforeRow
	afterRow.Order = *updated
	after := map[string]string{"status": string(updated.Status)}
	if reason != "" {
		after["reason"] = reason
	}
	recordAudit(u.log, u.audit, ctx, domainorder.AuditActionStaffUpdatedOrderStatus, actorID, actorRole, &orderID, "order",
		map[string]string{"status": string(beforeRow.Order.Status)}, after)

	parts, err := readOrderDetail(ctx, u.orders, u.tickets, u.payments, orderID)
	if err != nil {
		return nil, err
	}
	return toStaffOrderResponse(&afterRow, parts), nil
}

// stillInTime checks, as staff accept an order they reviewed, that its pickup
// still follows the rules it was booked under — above all, that the bakery
// still has the notice its items need: a request left waiting too long is
// refused rather than made late, and staff reject it or the customer moves it.
func (u *StaffOrderUsecase) stillInTime(ctx context.Context, order domainorder.Order) error {
	if order.PickupAt == nil {
		return nil
	}
	now := time.Now()
	settings, closed, err := readPickupRules(ctx, u.store, now)
	if err != nil {
		return err
	}
	items, err := u.orders.ListItemsByOrderID(ctx, order.ID)
	if err != nil {
		return err
	}
	needs, err := u.cartUC.orderFulfillment(ctx, items)
	if err != nil {
		return err
	}
	_, err = pickupPolicy(settings, closed).Validate(*order.PickupAt, now, needs)
	return err
}

func toStaffOrderCustomer(row *port.StaffOrderListRow) dto.StaffOrderCustomerResponse {
	return dto.StaffOrderCustomerResponse{
		UserID:      row.Order.UserID.String(),
		Email:       row.CustomerEmail,
		DisplayName: row.CustomerDisplayName,
		Phone:       row.CustomerPhone,
	}
}

func toStaffOrderResponse(row *port.StaffOrderListRow, parts orderDetailParts) *dto.StaffOrderResponse {
	resp := &dto.StaffOrderResponse{
		ID:                   row.Order.ID.String(),
		Code:                 row.Order.Code,
		Status:               string(row.Order.Status),
		OrderType:            string(row.Order.Type),
		SubtotalCents:        row.Order.SubtotalCents,
		DiscountCents:        row.Order.DiscountCents,
		TotalCents:           row.Order.TotalCents,
		DiscountCodeSnapshot: row.Order.DiscountCodeSnapshot,
		PickupAt:             row.Order.PickupAt,
		Items:                mapOrderItemsToDTO(parts.items),
		Timeline:             mapStaffOrderTimelineToDTO(parts.timeline),
		Tickets:              mapOrderTicketsToDTO(parts.tickets),
		Customer:             toStaffOrderCustomer(row),
		Payment:              mapOrderPaymentToDTO(parts.payment),
		CreatedAt:            row.Order.CreatedAt,
		UpdatedAt:            row.Order.UpdatedAt,
	}
	return resp
}
