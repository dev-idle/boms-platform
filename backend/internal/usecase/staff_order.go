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

type StaffOrderUsecase struct {
	orders port.OrderRepository
	audit  *auditlogger.Service
	log    *zap.Logger
}

func NewStaffOrderUsecase(
	orders port.OrderRepository,
	audit *auditlogger.Service,
	log *zap.Logger,
) *StaffOrderUsecase {
	return &StaffOrderUsecase{orders: orders, audit: audit, log: log}
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

	rows, err := u.orders.StaffList(ctx, port.StaffListOrdersParams{
		Status: status,
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	})
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	total, err := u.orders.StaffListCount(ctx, status)
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
			Status:     string(row.Order.Status),
			TotalCents: row.Order.TotalCents,
			ItemCount:  itemCounts[row.Order.ID],
			Customer:   toStaffOrderCustomer(&row),
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
	items, err := u.orders.ListItemsByOrderID(ctx, row.Order.ID)
	if err != nil {
		return nil, err
	}
	return toStaffOrderResponse(row, items), nil
}

func (u *StaffOrderUsecase) PatchStatus(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	orderID uuid.UUID,
	targetStatus domainorder.Status,
) (*dto.StaffOrderResponse, error) {
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
	if !domainorder.CanTransition(beforeRow.Order.Status, targetStatus) {
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
	recordAudit(u.log, u.audit, ctx, domainorder.AuditActionStaffUpdatedOrderStatus, actorID, actorRole, &orderID, "order",
		map[string]string{"status": string(beforeRow.Order.Status)},
		map[string]string{"status": string(updated.Status)},
	)

	items, err := u.orders.ListItemsByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	return toStaffOrderResponse(&afterRow, items), nil
}

func toStaffOrderCustomer(row *port.StaffOrderListRow) dto.StaffOrderCustomerResponse {
	return dto.StaffOrderCustomerResponse{
		UserID:      row.Order.UserID.String(),
		Email:       row.CustomerEmail,
		DisplayName: row.CustomerDisplayName,
	}
}

func toStaffOrderResponse(row *port.StaffOrderListRow, items []domainorder.Item) *dto.StaffOrderResponse {
	resp := &dto.StaffOrderResponse{
		ID:                   row.Order.ID.String(),
		Status:               string(row.Order.Status),
		SubtotalCents:        row.Order.SubtotalCents,
		DiscountCents:        row.Order.DiscountCents,
		TotalCents:           row.Order.TotalCents,
		DiscountCodeSnapshot: row.Order.DiscountCodeSnapshot,
		Items:                mapOrderItemsToDTO(items),
		Customer:             toStaffOrderCustomer(row),
		CreatedAt:            row.Order.CreatedAt,
		UpdatedAt:            row.Order.UpdatedAt,
	}
	return resp
}
