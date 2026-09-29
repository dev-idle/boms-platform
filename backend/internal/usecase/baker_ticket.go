package usecase

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// BakerTicketUsecase is the kitchen's queue: the kitchen tickets of orders
// accepted and not yet collected, started and finished one ticket at a time.
type BakerTicketUsecase struct {
	orders  port.OrderRepository
	tickets port.TicketRepository
	flow    ticketFlow
	audit   *auditlogger.Service
	log     *zap.Logger
}

func NewBakerTicketUsecase(
	orders port.OrderRepository,
	tickets port.TicketRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *BakerTicketUsecase {
	return &BakerTicketUsecase{
		orders:  orders,
		tickets: tickets,
		flow:    ticketFlow{tx: tx, orders: orders, tickets: tickets, events: events},
		audit:   audit,
		log:     log,
	}
}

func (u *BakerTicketUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	statusFilter string,
) ([]dto.StationTicketResponse, int64, int32, int32, error) {
	return listStationTickets(ctx, u.tickets, domaincategory.StationKitchen, page, pageSize, statusFilter)
}

// Get returns a kitchen ticket of an order the kitchen may see, with where the
// order's other tickets stand.
func (u *BakerTicketUsecase) Get(ctx context.Context, ticketID uuid.UUID) (*dto.StationTicketDetailResponse, error) {
	ticket, err := u.tickets.Get(ctx, ticketID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, domainorder.ErrTicketNotFound
		}
		return nil, err
	}
	if ticket.Station != domaincategory.StationKitchen {
		return nil, domainorder.ErrTicketNotFound
	}
	row, err := u.orders.StaffGetByID(ctx, ticket.OrderID)
	if err != nil {
		return nil, err
	}
	if !row.Order.Status.VisibleToBaker() {
		return nil, domainorder.ErrTicketNotFound
	}
	orderTickets, err := u.tickets.ListByOrder(ctx, ticket.OrderID)
	if err != nil {
		return nil, err
	}
	for _, t := range orderTickets {
		if t.ID == ticket.ID {
			ticket.Items = t.Items
		}
	}
	return &dto.StationTicketDetailResponse{
		StationTicketResponse: toStationTicketResponse(port.StationTicket{
			Ticket:              ticket,
			OrderCode:           row.Order.Code,
			OrderStatus:         row.Order.Status,
			PickupAt:            row.Order.PickupAt,
			CustomerDisplayName: row.CustomerDisplayName,
		}),
		OrderTickets: mapTicketSummariesToDTO(orderTickets),
	}, nil
}

// PatchStatus starts or finishes a kitchen ticket.
func (u *BakerTicketUsecase) PatchStatus(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	ticketID uuid.UUID,
	to domainorder.TicketStatus,
) (*dto.TicketChangeResponse, error) {
	change, err := u.flow.advance(ctx, ticketActor{id: actorID, role: actorRole}, ticketID, domaincategory.StationKitchen, to)
	if err != nil {
		return nil, err
	}
	recordAudit(u.log, u.audit, ctx, domainorder.AuditActionBakerUpdatedTicketStatus, actorID, actorRole, &ticketID, "order_ticket",
		map[string]string{"status": string(change.before.Status)},
		map[string]string{"status": string(change.after.Status)},
	)
	return toTicketChangeResponse(change), nil
}
