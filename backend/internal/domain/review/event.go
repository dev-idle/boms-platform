package review

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// TopicChanged announces a review written, hidden or published again.
const TopicChanged domainevent.Topic = "review.changed"

// ChangedEvent tells managers that a product's reviews changed — one was
// written, hidden or published again — and the author's pages, which show
// whether theirs is hidden. The storefront reads the published reviews as it
// shows them.
func ChangedEvent(productID, authorID uuid.UUID) domainevent.Event {
	return domainevent.New(TopicChanged, domainevent.Audience{
		UserIDs: []uuid.UUID{authorID},
		Roles:   []domainuser.Role{domainuser.RoleManager},
	}, map[string]string{"product_id": productID.String()})
}
