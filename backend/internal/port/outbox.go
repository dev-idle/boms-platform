package port

import (
	"context"
	"time"

	"github.com/google/uuid"

	domainevent "github.com/boms/backend/internal/domain/event"
)

// EventOutbox records an event in the transaction making the change it
// announces; outside a transaction it refuses.
type EventOutbox interface {
	Add(ctx context.Context, e domainevent.Event) error
}

// OutboxStore hands recorded events to the dispatcher and keeps their delivery
// state. Ages are measured on the store's own clock.
type OutboxStore interface {
	// ClaimUnpublished locks events still unpublished at least grace after they
	// were recorded, skipping rows another worker holds. It needs a transaction.
	ClaimUnpublished(ctx context.Context, grace time.Duration, limit int32) ([]domainevent.Event, error)
	MarkPublished(ctx context.Context, ids []uuid.UUID) error
	RecordFailure(ctx context.Context, ids []uuid.UUID, reason string) error
	// DeletePublished removes delivery records published longer ago than retention.
	DeletePublished(ctx context.Context, retention time.Duration) (int64, error)
}

// EventPublisher delivers events to their audiences.
type EventPublisher interface {
	Publish(ctx context.Context, events []domainevent.Event) error
}
