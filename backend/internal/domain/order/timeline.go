package order

import (
	"time"

	domainuser "github.com/boms/backend/internal/domain/user"
)

// StatusEvent is one entry in an order's history: the status it entered, when,
// and the role of whoever moved it there.
type StatusEvent struct {
	To        Status
	ActorRole domainuser.Role
	At        time.Time
}
