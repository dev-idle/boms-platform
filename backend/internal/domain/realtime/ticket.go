// Package realtime holds the rules for admitting a browser to the push-only
// WebSocket listener.
package realtime

import (
	"encoding/base64"
	"errors"

	"github.com/google/uuid"

	domainuser "github.com/boms/backend/internal/domain/user"
)

// TicketBytes is the randomness a ticket token carries: enough that nobody can
// guess one in the seconds it lives. A token is its unpadded base64url form.
const TicketBytes = 32

// ErrTicketInvalid means a ticket was never issued, has expired, or was used.
var ErrTicketInvalid = errors.New("realtime ticket invalid or already used")

// Ticket is what a single-use realtime ticket admits: one signed-in session.
// The listener decides from it which channels the socket hears, so a client can
// never pick its own.
type Ticket struct {
	UserID    uuid.UUID
	Role      domainuser.Role
	SessionID uuid.UUID
}

// WellFormedToken reports whether token has the shape of an issued ticket, so
// anything else is refused before it costs a lookup.
func WellFormedToken(token string) bool {
	if len(token) != base64.RawURLEncoding.EncodedLen(TicketBytes) {
		return false
	}
	_, err := base64.RawURLEncoding.DecodeString(token)
	return err == nil
}
