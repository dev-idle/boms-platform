package port

import (
	"context"

	domainsession "github.com/boms/backend/internal/domain/session"
)

// SessionStore persists server-side sessions (TTL from config).
type SessionStore interface {
	Create(ctx context.Context, userID, sessionID string, meta domainsession.SessionMeta) error
	Get(ctx context.Context, userID, sessionID string) (domainsession.SessionMeta, error)
	Delete(ctx context.Context, userID, sessionID string) error
	DeleteAllForUser(ctx context.Context, userID string) error
}
