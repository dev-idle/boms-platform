package postgres

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/boms/backend/internal/adapter/repository/postgres/sqlcgen"
	domainevent "github.com/boms/backend/internal/domain/event"
	domainuser "github.com/boms/backend/internal/domain/user"
	"github.com/boms/backend/internal/port"
	apperrors "github.com/boms/backend/internal/shared/errors"
)

// maxLastErrorLength keeps a delivery failure readable without letting a
// verbose driver error bloat the row.
const maxLastErrorLength = 500

type OutboxRepository struct {
	queries *sqlcgen.Queries
}

func NewOutboxRepository(pool *Pool) *OutboxRepository {
	return &OutboxRepository{queries: pool.Queries()}
}

func (r *OutboxRepository) q(ctx context.Context) *sqlcgen.Queries {
	if tx := txFromContext(ctx); tx != nil {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

// audienceRecord is the stored shape of an event's audience.
type audienceRecord struct {
	UserIDs []uuid.UUID `json:"user_ids,omitempty"`
	Roles   []string    `json:"roles,omitempty"`
}

// Add records e in the caller's transaction, stamped with the database clock,
// and queues it for the pool's commit hook. Outside a transaction the event
// could outlive a change that rolled back, so it is refused.
func (r *OutboxRepository) Add(ctx context.Context, e domainevent.Event) error {
	scope := scopeFromContext(ctx)
	if scope == nil {
		return apperrors.Errorf("add outbox event: requires a transaction")
	}
	roles := make([]string, 0, len(e.Audience.Roles))
	for _, role := range e.Audience.Roles {
		roles = append(roles, string(role))
	}
	audience, err := json.Marshal(audienceRecord{UserIDs: e.Audience.UserIDs, Roles: roles})
	if err != nil {
		return apperrors.Errorf("encode outbox audience: %w", err)
	}
	data, err := json.Marshal(e.Data)
	if err != nil {
		return apperrors.Errorf("encode outbox data: %w", err)
	}
	recordedAt, err := r.q(ctx).InsertOutboxEvent(ctx, sqlcgen.InsertOutboxEventParams{
		ID:       e.ID,
		Topic:    string(e.Topic),
		Audience: audience,
		Data:     data,
	})
	if err != nil {
		return mapRepoError(err, "add outbox event")
	}
	e.OccurredAt = recordedAt
	scope.events = append(scope.events, e)
	return nil
}

func (r *OutboxRepository) ClaimUnpublished(ctx context.Context, grace time.Duration, limit int32) ([]domainevent.Event, error) {
	if txFromContext(ctx) == nil {
		return nil, apperrors.Errorf("claim outbox events: requires a transaction")
	}
	rows, err := r.q(ctx).ClaimUnpublishedOutboxEvents(ctx, sqlcgen.ClaimUnpublishedOutboxEventsParams{
		GraceSeconds: grace.Seconds(),
		Limit:        limit,
	})
	if err != nil {
		return nil, mapRepoError(err, "claim outbox events")
	}
	events := make([]domainevent.Event, 0, len(rows))
	for _, row := range rows {
		e, err := mapOutboxEvent(row)
		if err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, nil
}

func (r *OutboxRepository) MarkPublished(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	if err := r.q(ctx).MarkOutboxEventsPublished(ctx, ids); err != nil {
		return mapRepoError(err, "mark outbox events published")
	}
	return nil
}

func (r *OutboxRepository) RecordFailure(ctx context.Context, ids []uuid.UUID, reason string) error {
	if len(ids) == 0 {
		return nil
	}
	if len(reason) > maxLastErrorLength {
		// Cutting mid-rune would leave invalid UTF-8, which Postgres rejects.
		reason = strings.ToValidUTF8(reason[:maxLastErrorLength], "")
	}
	if err := r.q(ctx).RecordOutboxPublishFailure(ctx, sqlcgen.RecordOutboxPublishFailureParams{
		LastError: reason,
		Ids:       ids,
	}); err != nil {
		return mapRepoError(err, "record outbox publish failure")
	}
	return nil
}

func (r *OutboxRepository) DeletePublished(ctx context.Context, retention time.Duration) (int64, error) {
	deleted, err := r.q(ctx).DeletePublishedOutboxEvents(ctx, retention.Seconds())
	if err != nil {
		return 0, mapRepoError(err, "delete published outbox events")
	}
	return deleted, nil
}

func mapOutboxEvent(row sqlcgen.ClaimUnpublishedOutboxEventsRow) (domainevent.Event, error) {
	var audience audienceRecord
	if err := json.Unmarshal(row.Audience, &audience); err != nil {
		return domainevent.Event{}, apperrors.Errorf("decode outbox audience %s: %w", row.ID, err)
	}
	var data map[string]string
	if err := json.Unmarshal(row.Data, &data); err != nil {
		return domainevent.Event{}, apperrors.Errorf("decode outbox data %s: %w", row.ID, err)
	}
	roles := make([]domainuser.Role, 0, len(audience.Roles))
	for _, role := range audience.Roles {
		roles = append(roles, domainuser.Role(role))
	}
	return domainevent.Event{
		ID:         row.ID,
		Topic:      domainevent.Topic(row.Topic),
		Audience:   domainevent.Audience{UserIDs: audience.UserIDs, Roles: roles},
		Data:       data,
		OccurredAt: row.CreatedAt,
	}, nil
}

var (
	_ port.EventOutbox = (*OutboxRepository)(nil)
	_ port.OutboxStore = (*OutboxRepository)(nil)
)
