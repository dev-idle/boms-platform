package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// ticketFlow commits ticket moves for the stations that make them. Every move
// locks the ticket's order first — the row lock the order's own status moves
// take with their guarded UPDATE — so the order's status is derived from
// tickets nobody else is moving, and the last two tickets finishing at once
// make the order ready exactly once.
type ticketFlow struct {
	tx      port.TxManager
	orders  port.OrderRepository
	tickets port.TicketRepository
	events  port.EventOutbox
}

// ticketActor is who makes a ticket move.
type ticketActor struct {
	id   uuid.UUID
	role domainuser.Role
}

// ticketChange is a committed ticket move: the ticket as it was and is now,
// and its order after the move.
type ticketChange struct {
	before domainorder.Ticket
	after  domainorder.Ticket
	order  domainorder.Order
}

// lockTicket reads a ticket, locks its order, and reads the ticket again: the
// status seen after the lock is the one no other move can change until commit.
func (f ticketFlow) lockTicket(txCtx context.Context, ticketID uuid.UUID) (domainorder.Ticket, *domainorder.Order, error) {
	ticket, err := f.tickets.Get(txCtx, ticketID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNotFound) {
			return domainorder.Ticket{}, nil, domainorder.ErrTicketNotFound
		}
		return domainorder.Ticket{}, nil, err
	}
	order, err := f.orders.LockForUpdate(txCtx, ticket.OrderID)
	if err != nil {
		return domainorder.Ticket{}, nil, err
	}
	ticket, err = f.tickets.Get(txCtx, ticketID)
	if err != nil {
		return domainorder.Ticket{}, nil, err
	}
	return ticket, order, nil
}

// advance moves a ticket made at station to the next status and moves its
// order along when its tickets say so: in production at the first start,
// ready once every ticket is. A ticket of another station is not found there.
func (f ticketFlow) advance(
	ctx context.Context,
	actor ticketActor,
	ticketID uuid.UUID,
	station domaincategory.Station,
	to domainorder.TicketStatus,
) (ticketChange, error) {
	var change ticketChange
	err := f.tx.WithTx(ctx, func(txCtx context.Context) error {
		ticket, order, err := f.lockTicket(txCtx, ticketID)
		if err != nil {
			return err
		}
		if ticket.Station != station {
			return domainorder.ErrTicketNotFound
		}
		if !order.Status.ProducesTickets() {
			return domainorder.ErrTicketOrderNotActive
		}
		if !domainorder.CanAdvanceTicket(ticket.Status, to) {
			return domainorder.ErrInvalidTicketTransition
		}
		updated, err := f.tickets.UpdateStatus(txCtx, port.UpdateTicketStatusParams{
			TicketID: ticket.ID, Station: station, From: ticket.Status, To: to,
		})
		if err != nil {
			if errors.Is(err, apperrors.ErrNotFound) {
				return domainorder.ErrInvalidTicketTransition
			}
			return err
		}
		next, err := f.followTickets(txCtx, actor, *order)
		if err != nil {
			return err
		}
		change = ticketChange{before: ticket, after: updated, order: next}
		return f.events.Add(txCtx, domainorder.TicketChangedEvent(*order, order.Status, updated))
	})
	return change, err
}

// followTickets moves a locked order to the status its tickets now call for,
// recording the move as made by the actor whose ticket caused it; an order
// ready only after its pickup time is an incident.
func (f ticketFlow) followTickets(txCtx context.Context, actor ticketActor, order domainorder.Order) (domainorder.Order, error) {
	tickets, err := f.tickets.ListByOrder(txCtx, order.ID)
	if err != nil {
		return domainorder.Order{}, err
	}
	statuses := make([]domainorder.TicketStatus, 0, len(tickets))
	for _, ticket := range tickets {
		statuses = append(statuses, ticket.Status)
	}
	next := domainorder.DeriveStatus(order.Status, statuses)
	if next == order.Status {
		return order, nil
	}
	moved, err := f.orders.UpdateStatus(txCtx, port.UpdateOrderStatusParams{
		OrderID: order.ID, FromStatus: order.Status, ToStatus: next,
	})
	if err != nil {
		return domainorder.Order{}, err
	}
	if err := f.orders.AddStatusEvent(txCtx, port.AddOrderStatusEventParams{
		OrderID: order.ID, From: &order.Status, To: next, Actor: &port.OrderActor{ID: actor.id, Role: actor.role},
	}); err != nil {
		return domainorder.Order{}, err
	}
	if err := f.events.Add(txCtx, domainorder.StatusChangedEvent(order.Status, *moved)); err != nil {
		return domainorder.Order{}, err
	}
	if moved.ReadyLate(time.Now()) {
		if _, err := recordIncident(txCtx, f.orders, f.events, port.AddOrderIncidentParams{
			OrderID: moved.ID, Type: domainorder.IncidentReadyLate,
		}); err != nil {
			return domainorder.Order{}, err
		}
	}
	return *moved, nil
}

// move sends a ticket nobody has started to the other station, while its order
// is still to be made. An order has one ticket per station, so a station that
// already has one refuses it.
func (f ticketFlow) move(ctx context.Context, ticketID uuid.UUID, to domaincategory.Station) (ticketChange, error) {
	var change ticketChange
	err := f.tx.WithTx(ctx, func(txCtx context.Context) error {
		ticket, order, err := f.lockTicket(txCtx, ticketID)
		if err != nil {
			return err
		}
		if order.Status == domainorder.StatusCancelled || order.Status == domainorder.StatusFulfilled {
			return domainorder.ErrTicketOrderNotActive
		}
		if ticket.Status != domainorder.TicketQueued {
			return domainorder.ErrTicketNotMovable
		}
		if ticket.Station == to {
			return domainorder.ErrTicketStationTaken
		}
		moved, err := f.tickets.Move(txCtx, ticket.ID, ticket.Station, to)
		if err != nil {
			if errors.Is(err, apperrors.ErrConflict) {
				return domainorder.ErrTicketStationTaken
			}
			if errors.Is(err, apperrors.ErrNotFound) {
				return domainorder.ErrTicketNotMovable
			}
			return err
		}
		change = ticketChange{before: ticket, after: moved, order: *order}
		return f.events.Add(txCtx, domainorder.TicketChangedEvent(*order, order.Status, moved))
	})
	return change, err
}
