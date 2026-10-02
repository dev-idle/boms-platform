package promotion

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

const (
	// TopicCreated announces a promotion a manager sent: the worker emails it.
	TopicCreated domainevent.Topic = "promotion.created"
	// TopicSent announces a promotion queued for every customer it goes to.
	TopicSent domainevent.Topic = "promotion.sent"
)

// CreatedEvent asks the worker to email the promotion, and tells managers'
// pages it is sending.
func CreatedEvent(promotionID uuid.UUID) domainevent.Event {
	return managersEvent(TopicCreated, promotionID)
}

// SentEvent tells managers' pages the promotion went out, and to how many.
func SentEvent(promotionID uuid.UUID) domainevent.Event {
	return managersEvent(TopicSent, promotionID)
}

func managersEvent(topic domainevent.Topic, promotionID uuid.UUID) domainevent.Event {
	return domainevent.New(topic, domainevent.Audience{
		Roles: []domainuser.Role{domainuser.RoleManager},
	}, map[string]string{"promotion_id": promotionID.String()})
}

// ToSend names the promotion an event asks the worker to email.
func ToSend(e domainevent.Event) (uuid.UUID, bool) {
	if e.Topic != TopicCreated {
		return uuid.Nil, false
	}
	promotionID, err := uuid.Parse(e.Data["promotion_id"])
	if err != nil {
		return uuid.Nil, false
	}
	return promotionID, true
}
