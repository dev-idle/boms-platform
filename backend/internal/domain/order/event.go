package order

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

const (
	// TopicOrderCreated announces a new order.
	TopicOrderCreated domainevent.Topic = "order.created"
	// TopicOrderStatusChanged announces an order moving to another status.
	TopicOrderStatusChanged domainevent.Topic = "order.status_changed"
)

// CreatedEvent announces order to its customer and the counter. The kitchen is
// not told: a new order is not yet production work it may see.
func CreatedEvent(order Order) domainevent.Event {
	return orderEvent(TopicOrderCreated, order, domainevent.Audience{
		UserIDs: []uuid.UUID{order.UserID},
		Roles:   []domainuser.Role{domainuser.RoleStaff},
	})
}

// StatusChangedEvent announces order's move from the status it left. The
// kitchen hears about it only when the order enters, leaves or moves within the
// statuses bakers work with; every other order is not theirs to see.
func StatusChangedEvent(from Status, order Order) domainevent.Event {
	roles := []domainuser.Role{domainuser.RoleStaff}
	if from.VisibleToBaker() || order.Status.VisibleToBaker() {
		roles = append(roles, domainuser.RoleBaker)
	}
	return orderEvent(TopicOrderStatusChanged, order, domainevent.Audience{
		UserIDs: []uuid.UUID{order.UserID},
		Roles:   roles,
	})
}

func orderEvent(topic domainevent.Topic, order Order, audience domainevent.Audience) domainevent.Event {
	return domainevent.New(topic, audience, map[string]string{
		"order_id": order.ID.String(),
		"status":   string(order.Status),
	})
}
