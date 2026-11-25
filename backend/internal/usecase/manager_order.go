package usecase

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// ManagerOrderUsecase lets a manager hand over an online order whose customer
// cannot give its pickup code, once they have checked who is collecting it.
type ManagerOrderUsecase struct {
	orders      port.OrderRepository
	transitions orderTransitions
	audit       *auditlogger.Service
}

func NewManagerOrderUsecase(
	orders port.OrderRepository,
	tickets port.TicketRepository,
	discounts port.DiscountCodeRepository,
	payments port.PaymentRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
) *ManagerOrderUsecase {
	return &ManagerOrderUsecase{
		orders: orders,
		transitions: orderTransitions{
			tx: tx, orders: orders, tickets: tickets, discounts: discounts, payments: payments, events: events,
		},
		audit: audit,
	}
}

// Lookup finds an order the bakery sees by its code.
func (u *ManagerOrderUsecase) Lookup(ctx context.Context, code string) (*dto.StaffOrderSummaryResponse, error) {
	id, err := u.orders.IDByCode(ctx, strings.ToUpper(strings.TrimSpace(code)))
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainorder.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	row, err := u.seen(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := toStaffOrderSummary(*row)
	return &resp, nil
}

// HandOver hands over a ready or missed online order without its pickup code,
// keeping in the audit log who did it and why.
func (u *ManagerOrderUsecase) HandOver(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	orderID uuid.UUID,
	req dto.ManagerHandoverRequest,
) (*dto.StaffOrderSummaryResponse, error) {
	reason, err := domainorder.NewHandoverReason(req.Reason)
	if err != nil {
		return nil, err
	}
	row, err := u.seen(ctx, orderID)
	if err != nil {
		return nil, err
	}
	from := row.Order.Status
	// An order staff took has no code: the counter hands it over with its cash.
	if row.Order.Channel != domainorder.ChannelOnline || (from != domainorder.StatusReady && from != domainorder.StatusNoShow) {
		return nil, domainorder.ErrInvalidStatusTransition
	}
	var updated *domainorder.Order
	err = u.transitions.tx.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		updated, err = u.transitions.applyInTx(txCtx, &port.OrderActor{ID: actorID, Role: actorRole}, port.UpdateOrderStatusParams{
			OrderID: orderID, FromStatus: from, ToStatus: domainorder.StatusFulfilled,
		}, "")
		if err != nil {
			return err
		}
		return u.audit.Log(txCtx, domainorder.AuditActionManagerHandedOverWithoutCode, actorID, actorRole, &orderID, "order",
			map[string]string{"status": string(from)}, map[string]string{"status": string(updated.Status), "reason": reason})
	})
	if err != nil {
		return nil, err
	}
	row.Order = *updated
	resp := toStaffOrderSummary(*row)
	return &resp, nil
}

// seen reads an order the bakery sees (domainorder.SeenByStaff); any other is
// not found.
func (u *ManagerOrderUsecase) seen(ctx context.Context, orderID uuid.UUID) (*port.StaffOrderListRow, error) {
	row, err := u.orders.StaffGetByID(ctx, orderID)
	if errors.Is(err, apperrors.ErrNotFound) {
		return nil, domainorder.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	history, err := u.orders.ListStatusEvents(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if !domainorder.SeenByStaff(row.Order.Status, history) {
		return nil, domainorder.ErrNotFound
	}
	return row, nil
}
