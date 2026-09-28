// Package event defines the change notices the system publishes after a
// transaction commits. An event is a hint that something changed — who should
// look again and at what — never the changed record itself: readers refetch
// through the API, where authorization is enforced. Each aggregate defines its
// own topics and constructors beside its other rules.
package event

import (
	"time"

	"github.com/google/uuid"

	domainuser "github.com/boms/backend/internal/domain/user"
)

// Topic names what changed. Values are part of the wire contract with clients.
type Topic string

// Audience lists who is told about an event: individual users, whole roles,
// or — for a notice that carries nothing private — everyone with a page open.
type Audience struct {
	UserIDs []uuid.UUID
	Roles   []domainuser.Role
	Public  bool
}

// Event is one change notice. Data carries identifiers and labels only.
// OccurredAt is set by the outbox from the database clock when it is recorded.
type Event struct {
	ID         uuid.UUID
	Topic      Topic
	Audience   Audience
	Data       map[string]string
	OccurredAt time.Time
}

// New returns an event with a fresh id.
func New(topic Topic, audience Audience, data map[string]string) Event {
	return Event{ID: uuid.New(), Topic: topic, Audience: audience, Data: data}
}
