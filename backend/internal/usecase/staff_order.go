package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
	"github.com/boms/backend/internal/shared/utils"
)

type StaffOrderUsecase struct {
	users       port.UserRepository
	orders      port.OrderRepository
	tickets     port.TicketRepository
	payments    port.PaymentRepository
	store       port.StoreSettingsRepository
	cartUC      *CartUsecase
	pickupCodes domainorder.PickupCodes
	attempts    port.Quota
	attemptsMax port.QuotaLimit
	tx          port.TxManager
	events      port.EventOutbox
	transitions orderTransitions
	audit       *auditlogger.Service
	log         *zap.Logger
}

func NewStaffOrderUsecase(
	users port.UserRepository,
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
	pickupCodes domainorder.PickupCodes,
	attempts port.Quota,
	attemptsMax port.QuotaLimit,
) *StaffOrderUsecase {
	return &StaffOrderUsecase{
		users:       users,
		orders:      orders,
		tickets:     tickets,
		payments:    payments,
		store:       store,
		cartUC:      cartUC,
		pickupCodes: pickupCodes,
		attempts:    attempts,
		attemptsMax: attemptsMax,
		tx:          tx,
		events:      events,
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

	out, err := u.summaries(ctx, rows)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	return out, total, page, pageSize, nil
}

// Pickups pages a bakery day's pickups by time, for the counter to see what
// is due, what is ready to hand over and what is late.
func (u *StaffOrderUsecase) Pickups(
	ctx context.Context,
	date string,
	page, pageSize int32,
) (*dto.StaffPickupsResponse, int64, int32, int32, error) {
	page, pageSize = normalizeOrderListPage(page, pageSize)
	day, err := time.ParseInLocation(domainstore.DayLayout, strings.TrimSpace(date), domainstore.Location)
	if err != nil {
		return nil, 0, page, pageSize, apperrors.ErrValidation.WithDetail("date", "use YYYY-MM-DD")
	}
	settings, err := u.store.GetSettings(ctx)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	params := port.StaffListPickupsParams{
		From:   day,
		To:     day.AddDate(0, 0, 1),
		Limit:  pageSize,
		Offset: utils.PageOffset(page, pageSize),
	}
	rows, total, err := listWithTotal(ctx,
		func(ctx context.Context) ([]port.StaffOrderListRow, error) {
			return u.orders.StaffListPickups(ctx, params)
		},
		func(ctx context.Context) (int64, error) {
			return u.orders.StaffListPickupsCount(ctx, params.From, params.To)
		},
	)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	pickups, err := u.summaries(ctx, rows)
	if err != nil {
		return nil, 0, page, pageSize, err
	}
	return &dto.StaffPickupsResponse{
		SlotMinutes: int(settings.SlotLength / time.Minute),
		Pickups:     pickups,
	}, total, page, pageSize, nil
}

// summaries maps listed orders with how many items each holds.
func (u *StaffOrderUsecase) summaries(ctx context.Context, rows []port.StaffOrderListRow) ([]dto.StaffOrderSummaryResponse, error) {
	orderIDs := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		orderIDs = append(orderIDs, row.Order.ID)
	}
	itemCounts, err := u.orders.SumItemQuantitiesByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, err
	}
	out := make([]dto.StaffOrderSummaryResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, dto.StaffOrderSummaryResponse{
			ID:         row.Order.ID.String(),
			Code:       row.Order.Code,
			Status:     string(row.Order.Status),
			Channel:    string(row.Order.Channel),
			TotalCents: row.Order.TotalCents,
			ItemCount:  itemCounts[row.Order.ID],
			Customer:   toStaffOrderCustomer(&row),
			PickupAt:   row.Order.PickupAt,
			CreatedAt:  row.Order.CreatedAt,
		})
	}
	return out, nil
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
	if targetStatus != domainorder.StatusFulfilled && (req.PickupCode != "" || req.CashCollected) {
		return nil, apperrors.ErrValidation.WithDetail("status", "only a handover takes a pickup code or cash")
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
	cashDue := false
	if targetStatus == domainorder.StatusFulfilled {
		if cashDue, err = u.checkHandover(ctx, actorID, beforeRow.Order, req); err != nil {
			return nil, err
		}
	}

	actor := &port.OrderActor{ID: actorID, Role: actorRole}
	params := port.UpdateOrderStatusParams{
		OrderID:    orderID,
		FromStatus: beforeRow.Order.Status,
		ToStatus:   targetStatus,
	}
	var updated *domainorder.Order
	if cashDue {
		// The cash is taken with the order: both are recorded or neither.
		err = u.tx.WithTx(ctx, func(txCtx context.Context) error {
			if updated, err = u.transitions.applyInTx(txCtx, actor, params, reason); err != nil {
				return err
			}
			return u.payments.CollectCash(txCtx, orderID)
		})
	} else {
		updated, err = u.transitions.apply(ctx, actor, params, reason)
	}
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

// checkHandover checks how a ready order is handed over, and reports whether
// cash is due with it. An order placed online needs the code its customer
// gives; one staff took is paid at the counter, so the counter confirms it
// took the cash instead.
func (u *StaffOrderUsecase) checkHandover(
	ctx context.Context,
	actorID uuid.UUID,
	order domainorder.Order,
	req dto.PatchStaffOrderStatusRequest,
) (bool, error) {
	if !order.Channel.TakenByStaff() {
		if req.PickupCode == "" || req.CashCollected {
			return false, apperrors.ErrValidation.WithDetail("pickup_code", "the customer's pickup code hands the order over")
		}
		return false, u.checkPickupCode(ctx, actorID, order.ID, req.PickupCode)
	}
	if req.PickupCode != "" || !req.CashCollected {
		return false, apperrors.ErrValidation.WithDetail("cash_collected", "confirm the cash due was taken")
	}
	return order.TotalCents > 0, nil
}

// checkPickupCode checks the code the customer gave at handoff. Each order
// takes a few tries a day, so the counter cannot walk the four digits while an
// order waits; a correct code counts as a try too. Running out is logged with
// who tried, since only staff can try a code.
func (u *StaffOrderUsecase) checkPickupCode(ctx context.Context, actorID, orderID uuid.UUID, code string) error {
	allowed, err := u.attempts.Take(ctx, "pickup_code:"+orderID.String(), u.attemptsMax)
	if err != nil {
		return err
	}
	if !allowed {
		u.log.Warn("pickup_code_locked", zap.String("order_id", orderID.String()), zap.String("actor_id", actorID.String()))
		return domainorder.ErrPickupCodeLocked
	}
	if !u.pickupCodes.Matches(orderID, code) {
		return domainorder.ErrPickupCodeInvalid
	}
	return nil
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
	_, err = pickupPolicy(settings, closed).Validate(*order.PickupAt, now, needs.HeldOn(domainstore.DayOf(*order.PickupAt)))
	return err
}

// toStaffOrderCustomer is who the order is for: an account, or a guest with no
// id or email, whose name and phone the row carries as the customer's.
func toStaffOrderCustomer(row *port.StaffOrderListRow) dto.StaffOrderCustomerResponse {
	resp := dto.StaffOrderCustomerResponse{
		Email:       row.CustomerEmail,
		DisplayName: row.CustomerDisplayName,
		Phone:       row.CustomerPhone,
	}
	if row.Order.UserID != nil {
		id := row.Order.UserID.String()
		resp.UserID = &id
	}
	return resp
}

func toStaffOrderResponse(row *port.StaffOrderListRow, parts orderDetailParts) *dto.StaffOrderResponse {
	resp := &dto.StaffOrderResponse{
		ID:                   row.Order.ID.String(),
		Code:                 row.Order.Code,
		Status:               string(row.Order.Status),
		Channel:              string(row.Order.Channel),
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
