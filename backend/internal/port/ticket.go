package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domaincategory "github.com/boms/backend/internal/domain/category"
	domainorder "github.com/boms/backend/internal/domain/order"
)

// UpdateTicketStatusParams moves a ticket at its own station from one status
// to the next; a ticket no longer in From, or at another station, is not moved.
type UpdateTicketStatusParams struct {
	TicketID uuid.UUID
	Station  domaincategory.Station
	From     domainorder.TicketStatus
	To       domainorder.TicketStatus
}

type StationTicketsParams struct {
	Station domaincategory.Station
	Status  *domainorder.TicketStatus
	Limit   int32
	Offset  int32
}

// StationTicket is a ticket in a station's queue with what the station needs
// to know of its order: no contact details, only the name to call.
type StationTicket struct {
	Ticket              domainorder.Ticket
	OrderCode           string
	OrderStatus         domainorder.Status
	PickupAt            *time.Time
	CustomerDisplayName *string
}

type TicketRepository interface {
	// ListLines is what an order's lines ask each station to make.
	ListLines(ctx context.Context, orderID uuid.UUID) ([]domainorder.TicketLine, error)
	// CreateForOrder creates an order's tickets and their items.
	CreateForOrder(ctx context.Context, orderID uuid.UUID, drafts []domainorder.TicketDraft) ([]domainorder.Ticket, error)
	// Get returns a ticket without its items; ErrNotFound when there is none.
	Get(ctx context.Context, id uuid.UUID) (domainorder.Ticket, error)
	// ListByOrder returns an order's tickets with their items.
	ListByOrder(ctx context.Context, orderID uuid.UUID) ([]domainorder.Ticket, error)
	// UpdateStatus returns ErrNotFound when the ticket was not in params.From at params.Station.
	UpdateStatus(ctx context.Context, params UpdateTicketStatusParams) (domainorder.Ticket, error)
	// Move returns ErrNotFound when the ticket is not queued at from, and
	// ErrConflict when the order already has a ticket at to.
	Move(ctx context.Context, id uuid.UUID, from, to domaincategory.Station) (domainorder.Ticket, error)
	// CancelForOrder cancels an order's tickets not cancelled yet and returns them.
	CancelForOrder(ctx context.Context, orderID uuid.UUID) ([]domainorder.Ticket, error)
	ListByStation(ctx context.Context, params StationTicketsParams) ([]StationTicket, error)
	CountByStation(ctx context.Context, station domaincategory.Station, status *domainorder.TicketStatus) (int64, error)
}
