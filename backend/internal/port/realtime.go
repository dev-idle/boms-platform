package port

import (
	"context"
	"time"

	domainrealtime "github.com/boms/backend/internal/domain/realtime"
)

// RealtimeTicketStore issues the single-use tickets a browser trades for a
// WebSocket connection.
type RealtimeTicketStore interface {
	// Issue stores t for ttl and returns the opaque token that redeems it.
	Issue(ctx context.Context, t domainrealtime.Ticket, ttl time.Duration) (string, error)
	// Consume redeems a token once; a second use or an expired token fails with
	// domainrealtime.ErrTicketInvalid.
	Consume(ctx context.Context, token string) (domainrealtime.Ticket, error)
}
