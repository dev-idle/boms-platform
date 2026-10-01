package order

import (
	"time"

	domainuser "github.com/boms/backend/internal/domain/user"
)

// StatusEvent is one entry in an order's history: the status it entered, when,
// and the role of whoever moved it there — empty when the system did. Reason
// is why the bakery cancelled the order, empty for any other move.
type StatusEvent struct {
	To        Status
	ActorRole domainuser.Role
	Reason    string
	At        time.Time
}
