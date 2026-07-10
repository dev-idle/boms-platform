package usecase

import (
	"context"
	"errors"
	"strings"

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

type BakerOrderUsecase struct {
	orders port.OrderRepository
	audit  *auditlogger.Service
	log    *zap.Logger
}

func NewBakerOrderUsecase(
	orders port.OrderRepository,
	audit *auditlogger.Service,
	log *zap.Logger,
) *BakerOrderUsecase {
	return &BakerOrderUsecase{orders: orders, audit: audit, log: log}
}

func (u *BakerOrderUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	statusFilter string,
) ([]dto.BakerOrderSummaryResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)

	var status *domainorder.Status
	if trimmed := strings.TrimSpace(statusFilter); trimmed != "" {
		parsed := domainorder.Status(trimmed)
		if !parsed.Valid() {
			return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("status", "invalid order status")
		}
		status = &parsed
	}

	rows, err := u.orders.BakerListProduction(ctx, port.BakerListOrdersParams{
		Status: status,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.orders.BakerListProductionCount(ctx, status)
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

	out := make([]dto.BakerOrderSummaryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.BakerOrderSummaryResponse{
			ID:         row.Order.ID.String(),
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

func (u *BakerOrderUsecase) Get(ctx context.Context, orderID uuid.UUID) (*dto.BakerOrderResponse, error) {
	row, err := u.orders.StaffGetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrNotFound
		}
		return nil, err
	}
	if !isBakerVisibleStatus(row.Order.Status) {
		return nil, domainorder.ErrNotFound
	}
	items, err := u.orders.ListItemsByOrderID(ctx, row.Order.ID)
	if err != nil {
		return nil, err
	}
	return toBakerOrderResponse(row, items), nil
}

func (u *BakerOrderUsecase) PatchStatus(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	orderID uuid.UUID,
	targetStatus domainorder.Status,
) (*dto.BakerOrderResponse, error) {
	if !targetStatus.Valid() {
		return nil, apperrors.ErrValidation.WithDetail("status", "invalid order status")
	}

	beforeRow, err := u.orders.StaffGetByID(ctx, orderID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrNotFound
		}
		return nil, err
	}
	if !isBakerVisibleStatus(beforeRow.Order.Status) {
		return nil, domainorder.ErrNotFound
	}
	if !domainorder.CanBakerTransition(beforeRow.Order.Status, targetStatus) {
		return nil, domainorder.ErrInvalidStatusTransition
	}

	updated, err := u.orders.UpdateStatus(ctx, port.UpdateOrderStatusParams{
		OrderID:    orderID,
		FromStatus: beforeRow.Order.Status,
		ToStatus:   targetStatus,
	})
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrInvalidStatusTransition
		}
		return nil, err
	}

	afterRow := *beforeRow
	afterRow.Order = *updated
	recordAudit(u.log, u.audit, ctx, domainorder.AuditActionBakerUpdatedOrderStatus, actorID, actorRole, &orderID, "order",
		map[string]string{"status": string(beforeRow.Order.Status)},
		map[string]string{"status": string(updated.Status)},
	)

	items, err := u.orders.ListItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return toBakerOrderResponse(&afterRow, items), nil
}

func isBakerVisibleStatus(status domainorder.Status) bool {
	switch status {
	case domainorder.StatusConfirmed, domainorder.StatusInProduction, domainorder.StatusReady:
		return true
	default:
		return false
	}
}

func toBakerOrderResponse(row *port.StaffOrderListRow, items []domainorder.Item) *dto.BakerOrderResponse {
	return &dto.BakerOrderResponse{
		ID:                   row.Order.ID.String(),
		Status:               string(row.Order.Status),
		SubtotalCents:        row.Order.SubtotalCents,
		DiscountCents:        row.Order.DiscountCents,
		TotalCents:           row.Order.TotalCents,
		DiscountCodeSnapshot: row.Order.DiscountCodeSnapshot,
		PickupAt:             row.Order.PickupAt,
		Items:                mapOrderItemsToDTO(items),
		Customer:             toStaffOrderCustomer(row),
		CreatedAt:            row.Order.CreatedAt,
		UpdatedAt:            row.Order.UpdatedAt,
	}
}
