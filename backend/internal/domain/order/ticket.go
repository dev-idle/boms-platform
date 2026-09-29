package order

import (
	"slices"
	"time"

	"github.com/google/uuid"

	domaincategory "github.com/boms/backend/internal/domain/category"
)

// TicketStatus is where one station's part of an order stands.
type TicketStatus string

const (
	TicketQueued     TicketStatus = "queued"
	TicketInProgress TicketStatus = "in_progress"
	TicketReady      TicketStatus = "ready"
	TicketCancelled  TicketStatus = "cancelled"
)

// Valid reports whether s is a known ticket status.
func (s TicketStatus) Valid() bool {
	switch s {
	case TicketQueued, TicketInProgress, TicketReady, TicketCancelled:
		return true
	default:
		return false
	}
}

// CanAdvanceTicket reports whether a station may move its ticket from one
// status to the next: start it, then finish it. Cancelling follows the order.
func CanAdvanceTicket(from, to TicketStatus) bool {
	return (from == TicketQueued && to == TicketInProgress) || (from == TicketInProgress && to == TicketReady)
}

// TicketItem is one product a ticket makes, named as it was at checkout.
type TicketItem struct {
	OrderItemID uuid.UUID
	ProductID   uuid.UUID
	Name        string
	Quantity    int32
}

// Ticket is the part of an order one station makes.
type Ticket struct {
	ID        uuid.UUID
	OrderID   uuid.UUID
	Station   domaincategory.Station
	Status    TicketStatus
	Items     []TicketItem
	CreatedAt time.Time
	UpdatedAt time.Time
}

// TicketLine is one product an order line asks a station to make: a product
// line itself, or one product a combo line holds, times the line's quantity.
type TicketLine struct {
	OrderItemID uuid.UUID
	ProductID   uuid.UUID
	Name        string
	Station     domaincategory.Station
	Quantity    int32
}

// TicketDraft is a ticket about to be created: a station and what it makes.
type TicketDraft struct {
	Station domaincategory.Station
	Items   []TicketItem
}

// Decompose splits an order's lines into one ticket per station, kitchen first,
// each keeping its lines in the order they came. A combo whose products come
// from both stations lands on both tickets.
func Decompose(lines []TicketLine) []TicketDraft {
	drafts := make([]TicketDraft, 0, 2)
	for _, station := range []domaincategory.Station{domaincategory.StationKitchen, domaincategory.StationCounter} {
		var items []TicketItem
		for _, line := range lines {
			if line.Station == station {
				items = append(items, TicketItem{
					OrderItemID: line.OrderItemID,
					ProductID:   line.ProductID,
					Name:        line.Name,
					Quantity:    line.Quantity,
				})
			}
		}
		if len(items) > 0 {
			drafts = append(drafts, TicketDraft{Station: station, Items: items})
		}
	}
	return drafts
}

// ProducesTickets reports whether an order in this status is being made: its
// tickets move only once staff accepted it, and until it is ready.
func (s Status) ProducesTickets() bool {
	return s == StatusConfirmed || s == StatusInProduction
}

// DeriveStatus is the status an order being made reaches from its tickets:
// in production once any ticket started, ready once every ticket is. Cancelled
// tickets do not count; an order not being made keeps its status.
func DeriveStatus(current Status, tickets []TicketStatus) Status {
	if !current.ProducesTickets() {
		return current
	}
	active := slices.DeleteFunc(slices.Clone(tickets), func(s TicketStatus) bool { return s == TicketCancelled })
	if len(active) == 0 {
		return current
	}
	if !slices.ContainsFunc(active, func(s TicketStatus) bool { return s != TicketReady }) {
		return StatusReady
	}
	if slices.ContainsFunc(active, func(s TicketStatus) bool { return s == TicketInProgress || s == TicketReady }) {
		return StatusInProduction
	}
	return current
}
