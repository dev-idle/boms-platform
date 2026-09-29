package usecase

import (
	"context"

	"github.com/google/uuid"
	"go.uber.org/zap"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/dto"
	"github.com/boms/backend/internal/port"
	"github.com/boms/backend/internal/service/auditlogger"
)

// StaffTicketUsecase is the counter's prep queue — its own tickets, started
// and finished one at a time — and the move of a ticket nobody has started to
// the other station.
type StaffTicketUsecase struct {
	tickets port.TicketRepository
	flow    ticketFlow
	audit   *auditlogger.Service
	log     *zap.Logger
}

func NewStaffTicketUsecase(
	orders port.OrderRepository,
	tickets port.TicketRepository,
	tx port.TxManager,
	events port.EventOutbox,
	audit *auditlogger.Service,
	log *zap.Logger,
) *StaffTicketUsecase {
	return &StaffTicketUsecase{
		tickets: tickets,
		flow:    ticketFlow{tx: tx, orders: orders, tickets: tickets, events: events},
		audit:   audit,
		log:     log,
	}
}

func (u *StaffTicketUsecase) List(
	ctx context.Context,
	page, pageSize int32,
	statusFilter string,
) ([]dto.StationTicketResponse, int64, int32, int32, error) {
	return listStationTickets(ctx, u.tickets, domaincategory.StationCounter, page, pageSize, statusFilter)
}

// PatchStatus starts or finishes a counter ticket.
func (u *StaffTicketUsecase) PatchStatus(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	ticketID uuid.UUID,
	to domainorder.TicketStatus,
) (*dto.TicketChangeResponse, error) {
	change, err := u.flow.advance(ctx, ticketActor{id: actorID, role: actorRole}, ticketID, domaincategory.StationCounter, to)
	if err != nil {
		return nil, err
	}
	recordAudit(u.log, u.audit, ctx, domainorder.AuditActionStaffUpdatedTicketStatus, actorID, actorRole, &ticketID, "order_ticket",
		map[string]string{"status": string(change.before.Status)},
		map[string]string{"status": string(change.after.Status)},
	)
	return toTicketChangeResponse(change), nil
}

// Move sends a ticket nobody has started to the other station.
func (u *StaffTicketUsecase) Move(
	ctx context.Context,
	actorID uuid.UUID,
	actorRole domainuser.Role,
	ticketID uuid.UUID,
	to domaincategory.Station,
) (*dto.TicketChangeResponse, error) {
	change, err := u.flow.move(ctx, ticketID, to)
	if err != nil {
		return nil, err
	}
	recordAudit(u.log, u.audit, ctx, domainorder.AuditActionStaffMovedTicket, actorID, actorRole, &ticketID, "order_ticket",
		map[string]string{"station": string(change.before.Station)},
		map[string]string{"station": string(change.after.Station)},
	)
	return toTicketChangeResponse(change), nil
}
