package conversation

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

const (
	// TopicMessageCreated announces a new message on a conversation.
	TopicMessageCreated domainevent.Topic = "message.created"
	// TopicConversationChanged announces a conversation read, resolved or
	// opened again.
	TopicConversationChanged domainevent.Topic = "conversation.changed"
)

// MessageCreatedEvent tells the customer and the counter that a message
// arrived on the customer's conversation about orderID, or on their general
// one when orderID is nil.
func MessageCreatedEvent(customerID uuid.UUID, orderID *uuid.UUID) domainevent.Event {
	return domainevent.New(TopicMessageCreated, domainevent.Audience{
		UserIDs: []uuid.UUID{customerID},
		Roles:   []domainuser.Role{domainuser.RoleStaff},
	}, threadData(customerID, orderID))
}

// ReadByCustomerEvent tells the customer's other pages that they read the
// conversation; the counter sees no read receipts.
func ReadByCustomerEvent(customerID uuid.UUID, orderID *uuid.UUID) domainevent.Event {
	return domainevent.New(TopicConversationChanged, domainevent.Audience{
		UserIDs: []uuid.UUID{customerID},
	}, threadData(customerID, orderID))
}

// ChangedAtCounterEvent tells the counter that a conversation was read there,
// resolved or opened again: their inbox shows it, the customer does not.
func ChangedAtCounterEvent(customerID uuid.UUID, orderID *uuid.UUID) domainevent.Event {
	return domainevent.New(TopicConversationChanged, domainevent.Audience{
		Roles: []domainuser.Role{domainuser.RoleStaff},
	}, threadData(customerID, orderID))
}

// threadData names the conversation an event is about: the order for an
// order's conversation, the customer for their general one.
func threadData(customerID uuid.UUID, orderID *uuid.UUID) map[string]string {
	if orderID != nil {
		return map[string]string{"order_id": orderID.String()}
	}
	return map[string]string{"customer_id": customerID.String()}
}
