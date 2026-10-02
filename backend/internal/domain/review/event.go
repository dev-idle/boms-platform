package review

import (
	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// TopicChanged announces a review written, published or hidden.
const TopicChanged domainevent.Topic = "review.changed"

// ChangedEvent tells managers that a product's reviews changed — one waits for
// them, or one was read — and the author's pages, which show where it stands.
// The storefront reads the published reviews as it shows them.
func ChangedEvent(productID, authorID uuid.UUID) domainevent.Event {
	return domainevent.New(TopicChanged, domainevent.Audience{
		UserIDs: []uuid.UUID{authorID},
		Roles:   []domainuser.Role{domainuser.RoleManager},
	}, map[string]string{"product_id": productID.String()})
}
