package conversation

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

const (
	// TopicMessageCreated announces a new message on an order's conversation.
	TopicMessageCreated domainevent.Topic = "message.created"
	// TopicConversationChanged announces a conversation read, resolved or
	// opened again.
	TopicConversationChanged domainevent.Topic = "conversation.changed"
)

// MessageCreatedEvent tells the order's customer and the counter that a
// message arrived on the order.
func MessageCreatedEvent(orderID, customerID uuid.UUID) domainevent.Event {
	return domainevent.New(TopicMessageCreated, domainevent.Audience{
		UserIDs: []uuid.UUID{customerID},
		Roles:   []domainuser.Role{domainuser.RoleStaff},
	}, map[string]string{"order_id": orderID.String()})
}

// ReadByCustomerEvent tells the customer's other pages that they read the
// order's messages; the counter sees no read receipts.
func ReadByCustomerEvent(orderID, customerID uuid.UUID) domainevent.Event {
	return domainevent.New(TopicConversationChanged, domainevent.Audience{
		UserIDs: []uuid.UUID{customerID},
	}, map[string]string{"order_id": orderID.String()})
}

// ChangedAtCounterEvent tells the counter that a conversation was read there,
// resolved or opened again: their inbox shows it, the customer does not.
func ChangedAtCounterEvent(orderID uuid.UUID) domainevent.Event {
	return domainevent.New(TopicConversationChanged, domainevent.Audience{
		Roles: []domainuser.Role{domainuser.RoleStaff},
	}, map[string]string{"order_id": orderID.String()})
}
