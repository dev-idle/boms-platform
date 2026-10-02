package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainorder "github.com/boms/backend/internal/domain/order"
)

// AddOrderIncidentParams is an incident to record with an order. Note is what
// staff wrote reporting it; Actor is nil when the system records it.
type AddOrderIncidentParams struct {
	OrderID uuid.UUID
	Type    domainorder.IncidentType
	Note    *string
	Actor   *OrderActor
}

// ListOrderIncidentsParams pages the incidents recorded in [From, To); a nil
// Type does not filter.
type ListOrderIncidentsParams struct {
	From   time.Time
	To     time.Time
	Type   *domainorder.IncidentType
	Limit  int32
	Offset int32
}

// ManagerOrderIncident is an incident as the manager's log shows it: with its
// order's code and pickup time, the name of the staff member who recorded it,
// nil when the system did, and for a bakery cancellation the reason the
// customer was given, kept in the order's history.
type ManagerOrderIncident struct {
	Incident     domainorder.Incident
	OrderCode    string
	PickupAt     *time.Time
	ActorName    *string
	CancelReason *string
}

// IncidentTypeCount is how many incidents of one type were recorded.
type IncidentTypeCount struct {
	Type  domainorder.IncidentType
	Count int64
}

// OrderIncidentRepository keeps what went wrong with orders.
type OrderIncidentRepository interface {
	AddIncident(ctx context.Context, params AddOrderIncidentParams) (*domainorder.Incident, error)
	// FlagPaymentAnomaly flags the order's customer when their orders' payments
	// failed or expired threshold times within window, unless a flag within it
	// already covers them, and returns the flag; nil when it records none, as
	// for an order without a customer. It locks the customer's flags until the
	// transaction ends, so payments failing at once all count; call it only
	// inside one, after recording the failure.
	FlagPaymentAnomaly(ctx context.Context, orderID uuid.UUID, window time.Duration, threshold int32) (*domainorder.Incident, error)
	ListIncidents(ctx context.Context, params ListOrderIncidentsParams) ([]ManagerOrderIncident, error)
	CountIncidents(ctx context.Context, from, to time.Time, incidentType *domainorder.IncidentType) (int64, error)
	// CountIncidentsByType returns how many incidents of each type were
	// recorded in [from, to); a type with none is left out.
	CountIncidentsByType(ctx context.Context, from, to time.Time) ([]IncidentTypeCount, error)
	// ListIncidentsByOrderIDs reads many orders' incidents in one round trip,
	// keyed by order, oldest first.
	ListIncidentsByOrderIDs(ctx context.Context, orderIDs []uuid.UUID) (map[uuid.UUID][]domainorder.Incident, error)
	// EraseIncidentNotes clears what staff wrote reporting incidents with the
	// customer's orders: it is about them.
	EraseIncidentNotes(ctx context.Context, userID uuid.UUID) error
}
