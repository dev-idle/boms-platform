package order

import (
	"time"

	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainstore "github.com/boms/backend/internal/domain/store"
	domainuser "github.com/boms/backend/internal/domain/user"
)

const (
	// TopicOrderCreated announces a new order.
	TopicOrderCreated domainevent.Topic = "order.created"
	// TopicOrderStatusChanged announces an order moving to another status.
	TopicOrderStatusChanged domainevent.Topic = "order.status_changed"
	// TopicSlotsChanged announces that the free pickup slots of a day changed.
	TopicSlotsChanged domainevent.Topic = "slots.changed"
	// TopicTicketChanged announces a ticket starting, finishing, moving station
	// or being cancelled with its order.
	TopicTicketChanged domainevent.Topic = "ticket.changed"
)

// TicketChangedEvent announces ticket to the order's customer, to the counter
// when staff see the order, and to the kitchen when bakers see it — a kitchen ticket shows where
// the order's other tickets stand, so the kitchen hears those too, as it hears
// the order's status changes. orderStatus is the status the kitchen knows the
// order by: for a cancellation, the status it left.
func TicketChangedEvent(customerID uuid.UUID, orderStatus Status, ticket Ticket) domainevent.Event {
	var roles []domainuser.Role
	if orderStatus.VisibleToStaff() {
		roles = append(roles, domainuser.RoleStaff)
	}
	if orderStatus.VisibleToBaker() {
		roles = append(roles, domainuser.RoleBaker)
	}
	return domainevent.New(TopicTicketChanged, domainevent.Audience{
		UserIDs: []uuid.UUID{customerID},
		Roles:   roles,
	}, map[string]string{
		"ticket_id": ticket.ID.String(),
		"order_id":  ticket.OrderID.String(),
		"station":   string(ticket.Station),
		"status":    string(ticket.Status),
	})
}

// SlotsChangedEvent tells every open checkout that a pickup slot on day (a
// bakery day, as domainstore.DayOf returns it) filled up or may have freed a
// place, so pickers showing that day refresh. It names the day only: what it
// reveals is what the slot list already shows.
func SlotsChangedEvent(day time.Time) domainevent.Event {
	return domainevent.New(TopicSlotsChanged, domainevent.Audience{Public: true}, map[string]string{
		"date": day.Format(domainstore.DayLayout),
	})
}

// CreatedEvent announces order to its customer only: it awaits payment, and
// the bakery hears about it when it is paid.
func CreatedEvent(order Order) domainevent.Event {
	return domainevent.New(TopicOrderCreated, domainevent.Audience{UserIDs: []uuid.UUID{order.UserID}}, map[string]string{
		"order_id": order.ID.String(),
		"status":   string(order.Status),
	})
}

// StatusChangedEvent announces order's move from the status it left. The
// counter hears about it unless the order was never paid, and the kitchen only
// when the order enters, leaves or moves within the statuses bakers work with;
// every other order is not theirs to see.
func StatusChangedEvent(from Status, order Order) domainevent.Event {
	var roles []domainuser.Role
	if from.VisibleToStaff() || order.Status.VisibleToStaff() {
		roles = append(roles, domainuser.RoleStaff)
	}
	if from.VisibleToBaker() || order.Status.VisibleToBaker() {
		roles = append(roles, domainuser.RoleBaker)
	}
	return domainevent.New(TopicOrderStatusChanged, domainevent.Audience{
		UserIDs: []uuid.UUID{order.UserID},
		Roles:   roles,
	}, map[string]string{
		"order_id": order.ID.String(),
		"from":     string(from),
		"status":   string(order.Status),
	})
}
