package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainsession "github.com/boms/backend/internal/domain/session"
	domainuser "github.com/boms/backend/internal/domain/user"
)

// PageCursor is where a keyset page stops: the next page holds rows older than
// this instant, or as old with a smaller id.
type PageCursor struct {
	At time.Time
	ID uuid.UUID
}

// SessionLister lists someone's live sign-in sessions, for their data export.
type SessionLister interface {
	ListForUser(ctx context.Context, userID string) ([]domainsession.SessionMeta, error)
}

// AccountActivity is one recorded change to an account.
type AccountActivity struct {
	ID         uuid.UUID
	ActorID    uuid.UUID
	ActorRole  domainuser.Role
	Action     domainuser.AuditAction
	BeforeJSON []byte
	AfterJSON  []byte
	// IP and UserAgent are the actor's; empty when none was recorded.
	IP        string
	UserAgent *string
	At        time.Time
}

// AccountActivityReader reads what was recorded about one account, newest
// first, a keyset page at a time.
type AccountActivityReader interface {
	ListForSubject(ctx context.Context, subjectID uuid.UUID, before *PageCursor, limit int32) ([]AccountActivity, error)
}
